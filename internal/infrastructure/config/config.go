package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	Kafka      KafkaConfig
	Kubernetes KubernetesConfig
	Runtime    RuntimeConfig
	Cleanup    CleanupConfig
	Encryption EncryptionConfig
}

type KafkaConfig struct {
	Address           string
	Username          string
	Password          string
	RequestTopic      string
	StatusTopic       string
	RequestConsumerID string
	FindingsTopic     string
	LocalTestMode     bool
	WorkerPoolSize    int // số goroutine worker xử lý message song song (env: FINDING_PROCESS_MAX_THREAD)
	MinQueueBuffer    int // ngưỡng buffer tối thiểu trước khi poll thêm (env: FINDING_PROCESS_MIN_QUEUE)
}

type KubernetesConfig struct {
	ScannerNamespace  string
	OperatorNamespace string
}

type RuntimeConfig struct {
	WorkerNode string
}

type EncryptionConfig struct {
	Key string
}

type CleanupConfig struct {
	CompletedRetentionMinutes string
	FailedRetentionDays       string
	CheckIntervalMinutes      string
}

func readEnv(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func readRequiredEnv(key string) (string, error) {
	value := os.Getenv(key)
	if value == "" {
		return "", fmt.Errorf("missing required env %s", key)
	}
	return value, nil
}

func readEnvInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	n, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return n
}

func Load() (Config, error) {
	kafkaAddr := readEnv("KAFKA_ADDR", "localhost:9092")
	requestTopic := readEnv("KAFKA_SCAN_REQUEST_TOPIC", "scan-task")
	statusTopic := readEnv("KAFKA_SCAN_STATUS", "scan-status")
	requestConsumerID := readEnv("KAFKA_SCB_REQUEST_CONSUMER", "scb-core-consumer")
	findingsTopic := readEnv("KAFKA_FINDINGS_TOPIC", "scan-findings")

	// Kubernetes config might still be required or mocked later.
	scannerNamespace := readEnv("SECURECODEBOX_NAMESPACE_SCANNER", "scb-scanner")
	operatorNamespace := readEnv("SECURECODEBOX_NAMESPACE_OPERATOR", "scb-system")

	completedRetentionMinutes := readEnv("MINUTES", "60")
	failedRetentionDays := readEnv("ERROR_SCAN_RETENTION_DAYS", "7")
	checkIntervalMinutes := readEnv("ERROR_SCAN_CHECK_MINUTES", "5")

	workerPoolSize := readEnvInt("FINDING_PROCESS_MAX_THREAD", 10)
	minQueueBuffer := readEnvInt("FINDING_PROCESS_MIN_QUEUE", 5)

	return Config{
		Kafka: KafkaConfig{
			Address:           kafkaAddr,
			Username:          os.Getenv("KAFKA_USER"),
			Password:          os.Getenv("KAFKA_PASSWORD"),
			RequestTopic:      requestTopic,
			StatusTopic:       statusTopic,
			RequestConsumerID: requestConsumerID,
			FindingsTopic:     findingsTopic,
			LocalTestMode:     os.Getenv("LOCAL_TEST_MODE") == "true",
			WorkerPoolSize:    workerPoolSize,
			MinQueueBuffer:    minQueueBuffer,
		},
		Kubernetes: KubernetesConfig{
			ScannerNamespace:  scannerNamespace,
			OperatorNamespace: operatorNamespace,
		},
		Runtime: RuntimeConfig{
			WorkerNode: os.Getenv("WORKER_NODE"),
		},
		Encryption: EncryptionConfig{
			Key: os.Getenv("ENCRYPTION-KEY"),
		},
		Cleanup: CleanupConfig{
			CompletedRetentionMinutes: completedRetentionMinutes,
			FailedRetentionDays:       failedRetentionDays,
			CheckIntervalMinutes:      checkIntervalMinutes,
		},
	}, nil
}
