package services

import (
	"context"

	"github.com/tungnt127/scb-core-v2/internal/domain/entities"
	domainservices "github.com/tungnt127/scb-core-v2/internal/domain/services"
	"github.com/tungnt127/scb-core-v2/internal/usecase/ports"
)

type BuildScanManifestService struct {
	renderer ports.ManifestRenderer
	registry *domainservices.Registry
}

func NewBuildScanManifestService(renderer ports.ManifestRenderer, registry *domainservices.Registry) *BuildScanManifestService {
	return &BuildScanManifestService{renderer: renderer, registry: registry}
}

func (s *BuildScanManifestService) Execute(ctx context.Context, request entities.ScanRequest) (entities.ScanManifestSpec, error) {
	strategy, err := s.registry.Resolve(request.ScannerAlias)
	if err != nil {
		return entities.ScanManifestSpec{}, err
	}
	return s.renderer.RenderManifest(ctx, request, strategy.BuildManifestInput(request))
}
