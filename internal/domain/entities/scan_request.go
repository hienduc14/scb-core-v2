package entities

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/tungnt127/scb-core-v2/internal/domain/valueobjects"
)

type ScanRequest struct {
	RequestUUID   string
	ScannerAlias  string
	Scanner       valueobjects.ScannerType
	ProjectType   valueobjects.ProjectType
	Target        ScanTarget
	AuthToken     string
	Branch        string
	BaseCommit    string
	RefID         string
	RequestTime   string
	PlatformTag   string
	NucleiConfig  map[string]string
	RequestParams map[string]string
	Metadata      map[string]string
}

func (r ScanRequest) Validate() error {
	if strings.TrimSpace(r.RequestUUID) == "" {
		return errors.New("requestUUID is required")
	}
	if strings.TrimSpace(r.ScannerAlias) == "" {
		return errors.New("scanner is required")
	}
	switch {
	case r.Scanner.IsRepositoryScanner():
		if strings.TrimSpace(r.Target.RepositoryURL) == "" {
			return errors.New("projectUri is required for repository scanners")
		}
		if !r.ProjectType.IsSupported() {
			return fmt.Errorf("unsupported projectType %q", r.ProjectType)
		}
	case r.Scanner.IsImageScanner():
		if strings.TrimSpace(r.Target.ImageRef) == "" {
			return errors.New("projectUri is required for image scanners")
		}
	case r.Scanner.IsURLScanner():
		if strings.TrimSpace(r.Target.TargetURL) == "" {
			return errors.New("targetUrl is required for nuclei")
		}
	default:
		return fmt.Errorf("scanner %q is not mapped", r.Scanner)
	}
	return nil
}

func (r ScanRequest) IsCommitScan() bool {
	return strings.TrimSpace(r.BaseCommit) != ""
}

func (r ScanRequest) FlattenedNucleiVars() []string {
	if len(r.RequestParams) == 0 {
		return nil
	}

	keys := make([]string, 0, len(r.RequestParams))
	for key := range r.RequestParams {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	args := make([]string, 0, len(keys))
	for _, key := range keys {
		args = append(args, fmt.Sprintf("-var %s=%s", key, r.RequestParams[key]))
	}
	return args
}
