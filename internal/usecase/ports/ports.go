package ports

import (
	"context"
	"time"

	"github.com/tungnt127/scb-core-v2/internal/domain/entities"
)

type ScanRequestConsumer interface {
	Start(ctx context.Context, handler func(context.Context, entities.ScanRequest) error) error
}

type StatusEventPublisher interface {
	PublishStatusEvent(ctx context.Context, event entities.ScanStatusEvent) error
}

type ErrorEventPublisher interface {
	PublishErrorEvent(ctx context.Context, event entities.ScanStatusEvent) error
}

type AuthDecryptor interface {
	Decrypt(ctx context.Context, cipherText string) (string, error)
}

type RepositoryVerifier interface {
	VerifyRepository(ctx context.Context, request entities.ScanRequest) (entities.VerificationResult, error)
}

type ImageVerifier interface {
	VerifyImage(ctx context.Context, request entities.ScanRequest) (entities.VerificationResult, error)
}

type URLTargetVerifier interface {
	VerifyURL(ctx context.Context, request entities.ScanRequest) (entities.VerificationResult, error)
}

type ManifestRenderer interface {
	RenderManifest(ctx context.Context, request entities.ScanRequest, input map[string]any) (entities.ScanManifestSpec, error)
}

type ScanSubmitter interface {
	SubmitScan(ctx context.Context, manifest entities.ScanManifestSpec) error
}

// PodChecker kiểm tra xem đã có pod/job nào đang chạy cho requestUUID này chưa
// để tránh xử lý trùng lặp cùng một scan request.
type PodChecker interface {
	ExistingScanPods(ctx context.Context, namespace, requestUUID string) ([]string, error)
}

type WatchSignalType string

const (
	WatchSignalScanStage   WatchSignalType = "stage"
	WatchSignalScanFailure WatchSignalType = "failure"
	WatchSignalFindings    WatchSignalType = "findings"
)

type WatchSignal struct {
	RequestUUID string
	Type        WatchSignalType
	Stage       string
	Message     string
	Error       *entities.ScanError
	Metadata    map[string]string
}

type ScanWatcher interface {
	Start(ctx context.Context, handler func(context.Context, WatchSignal) error) error
}

type FindingsSignalSource interface {
	Start(ctx context.Context, handler func(context.Context, WatchSignal) error) error
}

type ScanCleanupRepository interface {
	DeleteCompletedScans(ctx context.Context, olderThan time.Duration) (int, error)
	DeleteFailedScans(ctx context.Context, olderThan time.Duration) (int, error)
}

type Clock interface {
	Now() time.Time
}

type Logger interface {
	Info(msg string, fields map[string]any)
	Error(msg string, fields map[string]any)
}
