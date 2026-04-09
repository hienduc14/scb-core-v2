package services

import (
	"context"
	"fmt"

	"github.com/tungnt127/scb-core-v2/internal/domain/entities"
	domainservices "github.com/tungnt127/scb-core-v2/internal/domain/services"
	"github.com/tungnt127/scb-core-v2/internal/usecase/ports"
)

type VerifyTargetService struct {
	repositoryVerifier ports.RepositoryVerifier
	imageVerifier      ports.ImageVerifier
	urlVerifier        ports.URLTargetVerifier
	registry           *domainservices.Registry
}

func NewVerifyTargetService(repositoryVerifier ports.RepositoryVerifier, imageVerifier ports.ImageVerifier, urlVerifier ports.URLTargetVerifier, registry *domainservices.Registry) *VerifyTargetService {
	return &VerifyTargetService{
		repositoryVerifier: repositoryVerifier,
		imageVerifier:      imageVerifier,
		urlVerifier:        urlVerifier,
		registry:           registry,
	}
}

func (s *VerifyTargetService) Execute(ctx context.Context, request entities.ScanRequest) (entities.VerificationResult, error) {
	strategy, err := s.registry.Resolve(request.ScannerAlias)
	if err != nil {
		return entities.VerificationResult{}, err
	}

	switch strategy.VerificationMode() {
	case domainservices.VerifyRepository:
		return s.repositoryVerifier.VerifyRepository(ctx, request)
	case domainservices.VerifyImage:
		return s.imageVerifier.VerifyImage(ctx, request)
	case domainservices.VerifyURL:
		return s.urlVerifier.VerifyURL(ctx, request)
	default:
		return entities.VerificationResult{}, fmt.Errorf("unsupported verification mode %q", strategy.VerificationMode())
	}
}
