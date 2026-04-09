package k8s

import (
	"context"
	"time"
)

type CleanupRepository struct{}

func NewCleanupRepository() *CleanupRepository {
	return &CleanupRepository{}
}

func (r *CleanupRepository) DeleteCompletedScans(_ context.Context, _ time.Duration) (int, error) {
	// TODO: List Scan CRs, remove finalizers safely, and delete done resources.
	return 0, nil
}

func (r *CleanupRepository) DeleteFailedScans(_ context.Context, _ time.Duration) (int, error) {
	// TODO: Delete stale failed/error scans using retention policy.
	return 0, nil
}
