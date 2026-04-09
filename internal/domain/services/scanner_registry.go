package services

import (
	"fmt"
	"strings"

	"github.com/tungnt127/scb-core-v2/internal/domain/entities"
	"github.com/tungnt127/scb-core-v2/internal/domain/valueobjects"
)

type VerificationMode string

const (
	VerifyRepository VerificationMode = "repository"
	VerifyImage      VerificationMode = "image"
	VerifyURL        VerificationMode = "url"
)

type ScannerStrategy interface {
	Scanner() valueobjects.ScannerType
	Aliases() []string
	VerificationMode() VerificationMode
	RequiresAuth(request entities.ScanRequest) bool
	Normalize(request entities.ScanRequest) (entities.ScanRequest, error)
	BuildManifestInput(request entities.ScanRequest) map[string]any
}

type Registry struct {
	byAlias map[string]ScannerStrategy
}

func NewRegistry(strategies ...ScannerStrategy) *Registry {
	registry := &Registry{byAlias: make(map[string]ScannerStrategy)}
	for _, strategy := range strategies {
		for _, alias := range strategy.Aliases() {
			registry.byAlias[strings.ToLower(alias)] = strategy
		}
	}
	return registry
}

func DefaultRegistry() *Registry {
	return NewRegistry(
		newRepositoryStrategy(valueobjects.ScannerSemgrep, []string{"FS_CODE", "FS_CODE_COMMIT", "semgrep"}, true),
		newRepositoryStrategy(valueobjects.ScannerGitleaks, []string{"FS_SECRET", "FS_SECRET_COMMIT", "gitleaks"}, true),
		newRepositoryStrategy(valueobjects.ScannerKICS, []string{"FS_IAC", "FS_IAC_COMMIT", "kics"}, false),
		newImageStrategy(),
		newNucleiStrategy(),
	)
}

func (r *Registry) Resolve(alias string) (ScannerStrategy, error) {
	strategy, ok := r.byAlias[strings.ToLower(strings.TrimSpace(alias))]
	if !ok {
		return nil, fmt.Errorf("unknown scanner alias %q", alias)
	}
	return strategy, nil
}

type baseStrategy struct {
	scanner          valueobjects.ScannerType
	aliases          []string
	verificationMode VerificationMode
	requiresAuth     bool
	commitAware      bool
}

func (s baseStrategy) Scanner() valueobjects.ScannerType {
	return s.scanner
}

func (s baseStrategy) Aliases() []string {
	return s.aliases
}

func (s baseStrategy) VerificationMode() VerificationMode {
	return s.verificationMode
}

func (s baseStrategy) RequiresAuth(_ entities.ScanRequest) bool {
	return s.requiresAuth
}

func (s baseStrategy) Normalize(request entities.ScanRequest) (entities.ScanRequest, error) {
	request.Scanner = s.scanner
	return request, nil
}

func (s baseStrategy) BuildManifestInput(request entities.ScanRequest) map[string]any {
	return map[string]any{
		"requestUUID":  request.RequestUUID,
		"scanner":      request.Scanner,
		"scannerAlias": request.ScannerAlias,   // e.g. FS_SECRET vs FS_SECRET_COMMIT
		"target":       request.Target.Primary(),
		"branch":       request.Branch,
		"baseCommit":   request.BaseCommit,
		"refId":        request.RefID,
		"commitAware":  s.commitAware,
		"isCommitScan": request.IsCommitScan(), // true nếu BaseCommit != ""
		"authToken":    request.AuthToken,       // đã decrypt, cần cho template
		"platformTag":  request.PlatformTag,     // trivy arch tag
	}
}

type nucleiStrategy struct {
	baseStrategy
}

func newRepositoryStrategy(scanner valueobjects.ScannerType, aliases []string, commitAware bool) ScannerStrategy {
	return baseStrategy{
		scanner:          scanner,
		aliases:          aliases,
		verificationMode: VerifyRepository,
		requiresAuth:     true,
		commitAware:      commitAware,
	}
}

func newImageStrategy() ScannerStrategy {
	return baseStrategy{
		scanner:          valueobjects.ScannerTrivyImage,
		aliases:          []string{"FS_IMAGE", "trivy-image"},
		verificationMode: VerifyImage,
		requiresAuth:     true,
	}
}

func newNucleiStrategy() ScannerStrategy {
	return nucleiStrategy{
		baseStrategy: baseStrategy{
			scanner:          valueobjects.ScannerNuclei,
			aliases:          []string{"FS_DAST", "DAST_WEB", "DAST_NUCLEI", "NUCLEI", "nuclei"},
			verificationMode: VerifyURL,
		},
	}
}

func (s nucleiStrategy) RequiresAuth(request entities.ScanRequest) bool {
	return strings.TrimSpace(request.AuthToken) != ""
}

func (s nucleiStrategy) Normalize(request entities.ScanRequest) (entities.ScanRequest, error) {
	request.Scanner = s.scanner
	if request.Target.TargetURL == "" {
		request.Target.TargetURL = request.Target.RepositoryURL
		request.Target.RepositoryURL = ""
	}
	return request, nil
}

func (s nucleiStrategy) BuildManifestInput(request entities.ScanRequest) map[string]any {
	input := s.baseStrategy.BuildManifestInput(request)
	input["target"] = request.Target.TargetURL
	input["requestParams"] = request.FlattenedNucleiVars()
	input["nucleiConfig"] = request.NucleiConfig
	return input
}
