package httpclients

import (
	"context"
	"strings"

	"github.com/tungnt127/scb-core-v2/internal/domain/entities"
)

type RepositoryVerifier struct{}
type ImageVerifier struct{}
type URLVerifier struct{}

func NewRepositoryVerifier() *RepositoryVerifier {
	return &RepositoryVerifier{}
}

func NewImageVerifier() *ImageVerifier {
	return &ImageVerifier{}
}

func NewURLVerifier() *URLVerifier {
	return &URLVerifier{}
}

func (v *RepositoryVerifier) VerifyRepository(_ context.Context, request entities.ScanRequest) (entities.VerificationResult, error) {
	// TODO: Integrate GitHub/GitLab clients, commit root detection, and auth-aware access checks.
	return entities.VerificationResult{
		Accepted: strings.TrimSpace(request.Target.RepositoryURL) != "",
		Metadata: map[string]string{"kind": "repository"},
		Reason:   "repository URL is empty",
	}, nil
}

func (v *ImageVerifier) VerifyImage(_ context.Context, request entities.ScanRequest) (entities.VerificationResult, error) {
	// TODO: Integrate registry manifest verification with insecure TLS retry support.
	return entities.VerificationResult{
		Accepted: strings.TrimSpace(request.Target.ImageRef) != "",
		Metadata: map[string]string{"kind": "image"},
		Reason:   "image reference is empty",
	}, nil
}

func (v *URLVerifier) VerifyURL(_ context.Context, request entities.ScanRequest) (entities.VerificationResult, error) {
	// TODO: Integrate nuclei target/OpenAPI validation and S3 document accessibility checks.
	return entities.VerificationResult{
		Accepted: strings.TrimSpace(request.Target.TargetURL) != "",
		Metadata: map[string]string{"kind": "url"},
		Reason:   "target URL is empty",
	}, nil
}
