package services

import (
	"context"
	"fmt"

	"github.com/tungnt127/scb-core-v2/internal/domain/entities"
	domainservices "github.com/tungnt127/scb-core-v2/internal/domain/services"
	"github.com/tungnt127/scb-core-v2/internal/domain/valueobjects"
	"github.com/tungnt127/scb-core-v2/internal/usecase/contracts"
	"github.com/tungnt127/scb-core-v2/internal/usecase/ports"
)

type ProcessScanRequestService struct {
	registry          *domainservices.Registry
	statusPublisher   contracts.PublishStatusEventUseCase
	errorPublisher    contracts.PublishErrorEventUseCase
	decryptAuthToken  contracts.DecryptAuthTokenUseCase
	verifyTarget      contracts.VerifyTargetUseCase
	buildScanManifest contracts.BuildScanManifestUseCase
	submitScan        contracts.SubmitScanUseCase
	podChecker        ports.PodChecker
	k8sNamespace      string
	logger            ports.Logger
}

func NewProcessScanRequestService(
	registry *domainservices.Registry,
	statusPublisher contracts.PublishStatusEventUseCase,
	errorPublisher contracts.PublishErrorEventUseCase,
	decryptAuthToken contracts.DecryptAuthTokenUseCase,
	verifyTarget contracts.VerifyTargetUseCase,
	buildScanManifest contracts.BuildScanManifestUseCase,
	submitScan contracts.SubmitScanUseCase,
	podChecker ports.PodChecker,
	k8sNamespace string,
	logger ports.Logger,
) *ProcessScanRequestService {
	return &ProcessScanRequestService{
		registry:          registry,
		statusPublisher:   statusPublisher,
		errorPublisher:    errorPublisher,
		decryptAuthToken:  decryptAuthToken,
		verifyTarget:      verifyTarget,
		buildScanManifest: buildScanManifest,
		submitScan:        submitScan,
		podChecker:        podChecker,
		k8sNamespace:      k8sNamespace,
		logger:            logger,
	}
}

func (s *ProcessScanRequestService) Execute(ctx context.Context, request entities.ScanRequest) error {
	strategy, err := s.registry.Resolve(request.ScannerAlias)
	if err != nil {
		return s.publishFailure(ctx, request, entities.ScanError{
			Code:    valueobjects.ErrUnsupportedScanner,
			Stage:   valueobjects.StageInitializing,
			Message: err.Error(),
		})
	}

	request, err = strategy.Normalize(request)
	if err != nil {
		return s.publishFailure(ctx, request, entities.ScanError{
			Code:    valueobjects.ErrInvalidRequest,
			Stage:   valueobjects.StageInitializing,
			Message: err.Error(),
		})
	}

	if err := request.Validate(); err != nil {
		return s.publishFailure(ctx, request, entities.ScanError{
			Code:    valueobjects.ErrInvalidRequest,
			Stage:   valueobjects.StageInitializing,
			Message: err.Error(),
		})
	}

	// Kiểm tra Pod Deduplication: nếu đã có pod đang chạy cho requestUUID này thì skip
	if s.podChecker != nil {
		existingPods, checkErr := s.podChecker.ExistingScanPods(ctx, s.k8sNamespace, request.RequestUUID)
		if checkErr != nil && s.logger != nil {
			s.logger.Error("failed to check existing pods", map[string]any{
				"error":       checkErr.Error(),
				"requestUUID": request.RequestUUID,
			})
		}
		if len(existingPods) > 0 {
			if s.logger != nil {
				s.logger.Info("duplicate scan request detected, skipping", map[string]any{
					"requestUUID":   request.RequestUUID,
					"existingPods":  existingPods,
				})
			}
			return s.publishFailure(ctx, request, entities.ScanError{
				Code:    valueobjects.ErrScanSubmission,
				Stage:   valueobjects.StageInitializing,
				Message: "scan with this ID is already being processed",
			})
		}
	}

	if err := s.statusPublisher.Execute(ctx, entities.ScanStatusEvent{
		RequestUUID: request.RequestUUID,
		Scanner:     request.Scanner,
		Stage:       valueobjects.StageInitializing,
		Status:      "accepted",
		Message:     "scan request accepted",
	}); err != nil {
		return fmt.Errorf("publish accepted status: %w", err)
	}

	if strategy.RequiresAuth(request) {
		request, err = s.decryptAuthToken.Execute(ctx, request)
		if err != nil {
			return s.publishFailure(ctx, request, entities.ScanError{
				Code:    valueobjects.ErrAuthDecrypt,
				Stage:   valueobjects.StageInitializing,
				Message: "failed to decrypt auth token",
				Cause:   err.Error(),
			})
		}
	}

	verification, err := s.verifyTarget.Execute(ctx, request)
	if err != nil {
		return s.publishFailure(ctx, request, entities.ScanError{
			Code:    mapVerificationError(strategy.VerificationMode()),
			Stage:   valueobjects.StageInitializing,
			Message: "target verification failed",
			Cause:   err.Error(),
		})
	}
	if !verification.Accepted {
		return s.publishFailure(ctx, request, entities.ScanError{
			Code:    mapVerificationError(strategy.VerificationMode()),
			Stage:   valueobjects.StageInitializing,
			Message: verification.Reason,
			Context: verification.Metadata,
		})
	}

	manifest, err := s.buildScanManifest.Execute(ctx, request)
	if err != nil {
		return s.publishFailure(ctx, request, entities.ScanError{
			Code:    valueobjects.ErrManifestBuild,
			Stage:   valueobjects.StageInitializing,
			Message: "failed to build scan manifest",
			Cause:   err.Error(),
		})
	}

	if s.submitScan != nil {
		if err := s.submitScan.Execute(ctx, manifest); err != nil {
			return s.publishFailure(ctx, request, entities.ScanError{
				Code:    valueobjects.ErrScanSubmission,
				Stage:   valueobjects.StageInitializing,
				Message: "failed to submit scan",
				Cause:   err.Error(),
			})
		}
	} else if s.logger != nil {
		s.logger.Info("submitScan is disabled (local test mode), skipping K8s submission", map[string]any{
			"requestUUID": request.RequestUUID,
			"scanner":     request.Scanner,
		})
	}

	if s.logger != nil {
		s.logger.Info("scan request submitted", map[string]any{
			"requestUUID": request.RequestUUID,
			"scanner":     request.Scanner,
		})
	}

	return s.statusPublisher.Execute(ctx, entities.ScanStatusEvent{
		RequestUUID: request.RequestUUID,
		Scanner:     request.Scanner,
		Stage:       valueobjects.StageCloning,
		Status:      "running",
		Message:     "scan submitted to kubernetes",
	})
}

func (s *ProcessScanRequestService) publishFailure(ctx context.Context, request entities.ScanRequest, scanErr entities.ScanError) error {
	event := entities.ScanStatusEvent{
		RequestUUID: request.RequestUUID,
		Scanner:     request.Scanner,
		Stage:       scanErr.Stage,
		Status:      "failed",
		Message:     scanErr.Message,
		Terminal:    true,
		Error:       &scanErr,
	}

	publishErr := s.errorPublisher.Execute(ctx, event)
	if publishErr != nil {
		return fmt.Errorf("%s: publish error event: %w", scanErr.Message, publishErr)
	}
	return fmt.Errorf("%s", scanErr.Message)
}

func mapVerificationError(mode domainservices.VerificationMode) valueobjects.ErrorCode {
	switch mode {
	case domainservices.VerifyRepository:
		return valueobjects.ErrRepositoryLookup
	case domainservices.VerifyImage:
		return valueobjects.ErrImageLookup
	case domainservices.VerifyURL:
		return valueobjects.ErrURLVerification
	default:
		return valueobjects.ErrTargetVerification
	}
}
