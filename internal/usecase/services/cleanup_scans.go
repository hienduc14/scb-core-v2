package services

import (
	"context"

	"github.com/tungnt127/scb-core-v2/internal/domain/entities"
	"github.com/tungnt127/scb-core-v2/internal/usecase/ports"
)

type CleanupScansService struct {
	repository ports.ScanCleanupRepository
	policy     entities.CleanupPolicy
	logger     ports.Logger
}

func NewCleanupScansService(repository ports.ScanCleanupRepository, policy entities.CleanupPolicy, logger ports.Logger) *CleanupScansService {
	return &CleanupScansService{repository: repository, policy: policy, logger: logger}
}

func (s *CleanupScansService) Execute(ctx context.Context) error {
	completed, err := s.repository.DeleteCompletedScans(ctx, s.policy.CompletedRetention)
	if err != nil {
		return err
	}
	failed, err := s.repository.DeleteFailedScans(ctx, s.policy.FailedRetention)
	if err != nil {
		return err
	}
	if s.logger != nil {
		s.logger.Info("cleanup cycle completed", map[string]any{
			"completedDeleted": completed,
			"failedDeleted":    failed,
		})
	}
	return nil
}
