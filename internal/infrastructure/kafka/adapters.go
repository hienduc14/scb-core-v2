package kafka

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"time"

	kafkalib "github.com/segmentio/kafka-go"
	"github.com/tungnt127/scb-core-v2/internal/domain/entities"
	"github.com/tungnt127/scb-core-v2/internal/interfaceadapters/mappers"
	"github.com/tungnt127/scb-core-v2/internal/interfaceadapters/presenters"
	"github.com/tungnt127/scb-core-v2/internal/usecase/ports"
)

type localResolver struct{}

func (localResolver) LookupHost(ctx context.Context, host string) ([]string, error) {
	if host == "kafka" {
		return []string{"127.0.0.1"}, nil
	}
	return net.DefaultResolver.LookupHost(ctx, host)
}

func (localResolver) LookupBrokerIPAddr(ctx context.Context, b kafkalib.Broker) ([]net.IPAddr, error) {
	if b.Host == "kafka" {
		return []net.IPAddr{{IP: net.ParseIP("127.0.0.1")}}, nil
	}
	var ipAddrs []net.IPAddr
	ips, err := net.DefaultResolver.LookupIPAddr(ctx, b.Host)
	if err != nil {
		return nil, err
	}
	for _, ip := range ips {
		ipAddrs = append(ipAddrs, net.IPAddr{IP: ip.IP})
	}
	return ipAddrs, nil
}

type Consumer struct {
	reader         *kafkalib.Reader
	logger         ports.Logger
	workerPoolSize int
	minQueueBuffer int
	localTestMode  bool
}

func NewConsumer(address, topic, groupID string, logger ports.Logger, localTestMode bool, workerPoolSize, minQueueBuffer int) *Consumer {
	if workerPoolSize <= 0 {
		workerPoolSize = 10
	}
	if minQueueBuffer <= 0 {
		minQueueBuffer = 5
	}
	config := kafkalib.ReaderConfig{
		Brokers:     []string{address},
		Topic:       topic,
		GroupID:     groupID,
		MinBytes:    10e3, // 10KB
		MaxBytes:    10e6, // 10MB
		StartOffset: kafkalib.FirstOffset,
	}
	if localTestMode {
		config.Dialer = &kafkalib.Dialer{Resolver: localResolver{}}
	}
	return &Consumer{
		reader:         kafkalib.NewReader(config),
		logger:         logger,
		workerPoolSize: workerPoolSize,
		minQueueBuffer: minQueueBuffer,
		localTestMode:  localTestMode,
	}
}

// Start chạy một worker pool với workerPoolSize goroutine để xử lý message song song.
// Vòng lặp poll chính fetch message rồi đẩy vào channel (blocking khi channel đầy — backpressure).
// Mỗi worker goroutine gọi handler rồi commit sau khi xử lý xong (Option A: commit-after-process).
func (c *Consumer) Start(ctx context.Context, handler func(context.Context, entities.ScanRequest) error) error {
	defer c.reader.Close()

	if c.logger != nil {
		c.logger.Info("kafka consumer started with worker pool", map[string]any{
			"topic":          c.reader.Config().Topic,
			"workerPoolSize": c.workerPoolSize,
			"minQueueBuffer": c.minQueueBuffer,
		})
	}

	// Channel có buffer = workerPoolSize, đóng vai trò là FINDING_PROCESS_QUEUE
	msgCh := make(chan kafkalib.Message, c.workerPoolSize)

	// Spawn N worker goroutine (tương đương init_finding_consumer_start trong Python)
	for i := 0; i < c.workerPoolSize; i++ {
		workerID := i
		go func() {
			if c.logger != nil {
				c.logger.Info("worker started", map[string]any{"workerID": workerID})
			}
			for m := range msgCh {
				c.processMessage(ctx, m, handler)
			}
		}()
	}

	// Vòng lặp poll chính: chỉ fetch + enqueue, không xử lý trực tiếp
	for {
		m, err := c.reader.FetchMessage(ctx)
		if err != nil {
			if err == context.Canceled {
				// Tắt gracefully: đóng channel, để các worker drain nốt
				close(msgCh)
				return err
			}
			if c.logger != nil {
				c.logger.Error("failed to fetch message", map[string]any{"error": err.Error()})
			}
			continue
		}

		if c.logger != nil {
			c.logger.Info("message received from kafka", map[string]any{
				"partition": m.Partition,
				"offset":    m.Offset,
			})
		}

		// Backpressure: gửi vào channel, sẽ block nếu tất cả worker đang bận
		select {
		case msgCh <- m:
		case <-ctx.Done():
			close(msgCh)
			return ctx.Err()
		}
	}
}

// processMessage là logic xử lý 1 message trong worker goroutine.
// Tương đương hàm on_handle_msg trong Python: decode → validate → gọi handler → commit.
func (c *Consumer) processMessage(ctx context.Context, m kafkalib.Message, handler func(context.Context, entities.ScanRequest) error) {
	// --- Bước 1: Kiểm tra payload có phải là JSON object hợp lệ không ---
	// Lỗi phổ biến: producer gửi JSON pretty-printed bị split từng dòng
	// (ví dụ: dùng kafka-console-producer paste multi-line JSON).
	// Mỗi dòng trở thành 1 message riêng → không parse được.
	trimmed := bytes.TrimSpace(m.Value)
	if len(trimmed) == 0 {
		c.reader.CommitMessages(ctx, m)
		return
	}
	if trimmed[0] != '{' {
		if c.logger != nil {
			c.logger.Error("skipping non-JSON-object message — likely a line fragment from pretty-printed JSON producer. Use compact JSON (no newlines) when producing.", map[string]any{
				"partition": m.Partition,
				"offset":    m.Offset,
				"payload":   string(trimmed),
			})
		}
		c.reader.CommitMessages(ctx, m)
		return
	}

	// --- Bước 2: Decode JSON → ScanRequest ---
	req, err := mappers.DecodeScanRequest(trimmed)
	if err != nil {
		if c.logger != nil {
			c.logger.Error("failed to decode scan request", map[string]any{
				"error":     err.Error(),
				"partition": m.Partition,
				"offset":    m.Offset,
				"payload":   string(trimmed),
				"hint":      "ensure the Kafka producer sends compact single-line JSON",
			})
		}
		// Ghi debug file ngay cả khi parse lỗi
		if c.localTestMode {
			c.writeDebugEntry(m.Partition, m.Offset, trimmed, nil, err)
		}
		c.reader.CommitMessages(ctx, m)
		return
	}

	// --- Bước 3: Ghi debug file (chỉ khi LOCAL_TEST_MODE=true) ---
	if c.localTestMode {
		c.writeDebugEntry(m.Partition, m.Offset, trimmed, &req, nil)
	}

	// --- Bước 4: Gọi handler xử lý business logic ---
	err = handler(ctx, req)
	if err != nil && c.logger != nil {
		c.logger.Error("handler error", map[string]any{
			"error": err.Error(),
			"uuid":  req.RequestUUID,
		})
	}

	// Commit sau khi xử lý xong (Option A: at-least-once, đảm bảo không mất message)
	c.reader.CommitMessages(ctx, m)
}

// writeDebugEntry ghi thông tin 1 message vào file scb-debug-messages.jsonl.
// Mỗi dòng là 1 JSON entry độc lập, dễ đọc bằng jq, grep, v.v.
// Chỉ được gọi khi localTestMode = true.
func (c *Consumer) writeDebugEntry(partition int, offset int64, raw []byte, parsed *entities.ScanRequest, parseErr error) {
	fmt.Printf("\n[DEBUG-KAFKA] Dang thu ghi message (offset: %d) vao file debug...\n", offset)
	
	exePath, err := os.Executable()
	if err != nil {
		fmt.Printf("[DEBUG-KAFKA] ERROR: Khong lay duoc duong dan exe: %v\n", err)
		return
	}
	debugFile := filepath.Join(filepath.Dir(exePath), "scb-debug-messages.jsonl")

	entry := map[string]any{
		"time":      time.Now().Format(time.RFC3339Nano),
		"partition": partition,
		"offset":    offset,
		"raw":       string(raw),
	}
	if parseErr != nil {
		entry["parseError"] = parseErr.Error()
	}
	if parsed != nil {
		entry["parsed"] = map[string]any{
			"requestUUID": parsed.RequestUUID,
			"scanner":     parsed.ScannerAlias,
			"projectType": parsed.ProjectType,
			"projectUri":  parsed.Target.RepositoryURL,
			"branch":      parsed.Branch,
			"commitId":    parsed.BaseCommit,
			"refId":       parsed.RefID,
			"requestTime": parsed.RequestTime,
			"authToken":   "[REDACTED]",
		}
	}

	data, parseErrJSON := json.Marshal(entry)
	if parseErrJSON != nil {
		fmt.Printf("[DEBUG-KAFKA] ERROR: Loi khi marshal JSON: %v\n", parseErrJSON)
		if c.logger != nil {
			c.logger.Error("debug: failed to marshal debug entry", map[string]any{"error": parseErrJSON.Error()})
		}
		return
	}

	fmt.Printf("[DEBUG-KAFKA] Mo file (hoac tao moi neu chua co): %s\n", debugFile)
	f, err := os.OpenFile(debugFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("[DEBUG-KAFKA] ERROR: Loi khi mo file debug: %v\n", err)
		if c.logger != nil {
			c.logger.Error("debug: failed to open debug file", map[string]any{"error": err.Error(), "file": debugFile})
		}
		return
	}
	defer f.Close()

	_, err = fmt.Fprintf(f, "%s\n", data)
	if err != nil {
		fmt.Printf("[DEBUG-KAFKA] ERROR: Loi khi ghi vao file: %v\n", err)
	} else {
		fmt.Printf("[DEBUG-KAFKA] Thanh cong! Da ghi message vao %s\n", debugFile)
	}

	if c.logger != nil {
		c.logger.Info("debug: message written to debug file", map[string]any{
			"file":   debugFile,
			"offset": offset,
			"uuid":   func() string {
				if parsed != nil {
					return parsed.RequestUUID
				}
				return "(parse failed)"
			}(),
		})
	}
}


type StatusPublisher struct {
	presenter *presenters.StatusPresenter
	logger    ports.Logger
	writer    *kafkalib.Writer
}

func NewStatusPublisher(address, topic string, logger ports.Logger, localTestMode bool) *StatusPublisher {
	writer := &kafkalib.Writer{
		Addr:     kafkalib.TCP(address),
		Topic:    topic,
		Balancer: &kafkalib.LeastBytes{},
	}
	if localTestMode {
		writer.Transport = &kafkalib.Transport{Resolver: localResolver{}}
	}
	return &StatusPublisher{
		presenter: presenters.NewStatusPresenter(),
		logger:    logger,
		writer:    writer,
	}
}

func (p *StatusPublisher) PublishStatusEvent(ctx context.Context, event entities.ScanStatusEvent) error {
	payload := p.presenter.Present(event)
	data, err := json.Marshal(payload)
	if err != nil {
		if p.logger != nil {
			p.logger.Error("failed to marshal status event", map[string]any{"error": err.Error()})
		}
		return err
	}

	err = p.writer.WriteMessages(ctx, kafkalib.Message{
		Key:   []byte(event.RequestUUID),
		Value: data,
	})
	if err != nil {
		if p.logger != nil {
			p.logger.Error("failed to publish status event", map[string]any{"error": err.Error()})
		}
		return err
	}
	if p.logger != nil {
		p.logger.Info("published status event", map[string]any{"uuid": event.RequestUUID, "stage": event.Stage})
	}
	return nil
}

func (p *StatusPublisher) PublishErrorEvent(ctx context.Context, event entities.ScanStatusEvent) error {
	payload := p.presenter.Present(event)
	data, err := json.Marshal(payload)
	if err != nil {
		if p.logger != nil {
			p.logger.Error("failed to marshal error event", map[string]any{"error": err.Error()})
		}
		return err
	}

	err = p.writer.WriteMessages(ctx, kafkalib.Message{
		Key:   []byte(event.RequestUUID),
		Value: data,
	})
	if err != nil {
		if p.logger != nil {
			p.logger.Error("failed to publish error event", map[string]any{"error": err.Error()})
		}
		return err
	}
	if p.logger != nil {
		p.logger.Error("published error event", map[string]any{"uuid": event.RequestUUID, "stage": event.Stage})
	}
	return nil
}

type FindingsConsumer struct {
	reader *kafkalib.Reader
}

func NewFindingsConsumer(address, topic, groupID string, localTestMode bool) *FindingsConsumer {
	config := kafkalib.ReaderConfig{
		Brokers: []string{address},
		Topic:   topic,
		GroupID: groupID,
	}
	if localTestMode {
		config.Dialer = &kafkalib.Dialer{Resolver: localResolver{}}
	}
	return &FindingsConsumer{reader: kafkalib.NewReader(config)}
}

func (c *FindingsConsumer) Start(ctx context.Context, handler func(context.Context, ports.WatchSignal) error) error {
	defer c.reader.Close()
	<-ctx.Done()
	return ctx.Err()
}
