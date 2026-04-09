package scheduler

import (
	"context"
	"time"

	"github.com/tungnt127/scb-core-v2/internal/usecase/contracts"
)

type CleanupScheduler struct {
	interval time.Duration
	useCase  contracts.CleanupScansUseCase
}

func NewCleanupScheduler(interval time.Duration, useCase contracts.CleanupScansUseCase) *CleanupScheduler {
	return &CleanupScheduler{
		interval: interval,
		useCase:  useCase,
	}
}

func (s *CleanupScheduler) Start(ctx context.Context) error {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := s.useCase.Execute(ctx); err != nil {
				return err
			}
		}
	}
}
