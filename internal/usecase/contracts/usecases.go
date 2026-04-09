package contracts

import (
	"context"

	"github.com/tungnt127/scb-core-v2/internal/domain/entities"
)

type ProcessScanRequestUseCase interface {
	Execute(ctx context.Context, request entities.ScanRequest) error
}

type VerifyTargetUseCase interface {
	Execute(ctx context.Context, request entities.ScanRequest) (entities.VerificationResult, error)
}

type DecryptAuthTokenUseCase interface {
	Execute(ctx context.Context, request entities.ScanRequest) (entities.ScanRequest, error)
}

type BuildScanManifestUseCase interface {
	Execute(ctx context.Context, request entities.ScanRequest) (entities.ScanManifestSpec, error)
}

type SubmitScanUseCase interface {
	Execute(ctx context.Context, manifest entities.ScanManifestSpec) error
}

type HandleK8sObservationUseCase interface {
	Start(ctx context.Context) error
}

type PublishStatusEventUseCase interface {
	Execute(ctx context.Context, event entities.ScanStatusEvent) error
}

type PublishErrorEventUseCase interface {
	Execute(ctx context.Context, event entities.ScanStatusEvent) error
}

type CleanupScansUseCase interface {
	Execute(ctx context.Context) error
}
