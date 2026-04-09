package services

import (
	"context"

	"github.com/tungnt127/scb-core-v2/internal/domain/entities"
	"github.com/tungnt127/scb-core-v2/internal/usecase/ports"
)

type SubmitScanService struct {
	submitter ports.ScanSubmitter
}

func NewSubmitScanService(submitter ports.ScanSubmitter) *SubmitScanService {
	return &SubmitScanService{submitter: submitter}
}

func (s *SubmitScanService) Execute(ctx context.Context, manifest entities.ScanManifestSpec) error {
	return s.submitter.SubmitScan(ctx, manifest)
}
