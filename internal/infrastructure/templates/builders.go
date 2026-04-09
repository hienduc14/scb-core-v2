package templates

import (
	"net/url"
	"strings"

	"github.com/tungnt127/scb-core-v2/internal/domain/entities"
)

// =======================
// STRUCT
// =======================

// =======================
// NAMING
// =======================

// map alias -> scanner name chuẩn
func normalizeScannerName(alias string) string {
	alias = strings.ToUpper(strings.TrimSpace(alias))

	switch alias {
	case "FS_SECRET", "FS_SECRET_COMMIT":
		return "gitleaks"
	case "FS_CODE", "FS_CODE_COMMIT":
		return "semgrep"
	case "IMAGE":
		return "trivy"
	case "IAC":
		return "kics"
	default:
		return strings.ToLower(alias)
	}
}

func buildScanName(scannerAlias string, scanID string) string {
	scanner := normalizeScannerName(scannerAlias)
	return scanner + "-scan-" + scanID
}

// =======================
// GITLEAKS
// =======================

func buildGitleaksTemplateInput(request entities.ScanRequest, input map[string]any, namespace string) ScanTemplateInput {
	data := buildCommonTemplateInput(request, input, namespace)
	data.AuthToken = url.QueryEscape(data.AuthToken)

	alias := strings.ToUpper(strings.TrimSpace(request.ScannerAlias))

	if alias == "FS_SECRET_COMMIT" {
		// Python: base_commit -1 + remove --no-git
		data.IsCommitScan = true
		data.EffectiveCommit = data.BaseCommit + " -1"
	} else {
		// Default
		data.IsCommitScan = false
		data.EffectiveCommit = data.Branch
	}

	return data
}

// =======================
// SEMGREP
// =======================

func buildSemgrepTemplateInput(request entities.ScanRequest, input map[string]any, namespace string) ScanTemplateInput {
	data := buildCommonTemplateInput(request, input, namespace)
	data.AuthToken = url.QueryEscape(data.AuthToken)

	alias := strings.ToUpper(strings.TrimSpace(request.ScannerAlias))

	if alias == "FS_CODE" {
		// Python: remove baseline commit
		data.IsCommitScan = false
		data.EffectiveCommit = ""
	} else if alias == "FS_CODE_COMMIT" {
		// Python: base_commit^
		data.IsCommitScan = true
		data.EffectiveCommit = data.BaseCommit + "^"
	}

	return data
}

// =======================
// KICS
// =======================

func buildKicsTemplateInput(request entities.ScanRequest, input map[string]any, namespace string) ScanTemplateInput {
	data := buildCommonTemplateInput(request, input, namespace)
	data.AuthToken = url.QueryEscape(data.AuthToken)

	data.IsCommitScan = false
	data.EffectiveCommit = data.Branch

	return data
}

// =======================
// TRIVY
// =======================

func buildTrivyTemplateInput(request entities.ScanRequest, input map[string]any, namespace string) ScanTemplateInput {
	data := buildCommonTemplateInput(request, input, namespace)

	// username:password
	if data.AuthToken != "" {
		parts := strings.SplitN(data.AuthToken, ":", 2)
		if len(parts) == 2 {
			data.TrivyUsername = parts[0]
			data.TrivyPassword = parts[1]
		}
	}

	// --insecure
	if isInsecure, ok := input["isInsecure"].(bool); ok {
		data.IsInsecure = isInsecure
	}

	return data
}

// =======================
// COMMON
// =======================

func buildCommonTemplateInput(request entities.ScanRequest, input map[string]any, namespace string) ScanTemplateInput {
	authToken, _ := input["authToken"].(string)
	platformTag, _ := input["platformTag"].(string)

	if platformTag == "" {
		platformTag = "linux/amd64"
	}

	repoURL := request.Target.RepositoryURL
	repoURLWithToken := repoURL

	if strings.HasPrefix(repoURL, "https://") {
		repoURLWithToken = "https://oauth2:$(GIT_TOKEN)@" + repoURL[len("https://"):]
	}

	scanID := request.RequestUUID
	scanName := buildScanName(request.ScannerAlias, scanID)

	return ScanTemplateInput{
		ScanID:           scanID,
		ScanName:         scanName,
		Namespace:        namespace,
		RepoURL:          repoURL,
		RepoURLWithToken: repoURLWithToken,
		AuthToken:        authToken,
		Branch:           request.Branch,
		BaseCommit:       request.BaseCommit,
		ScannerAlias:     request.ScannerAlias,
		ImageRef:         request.Target.ImageRef,
		PlatformTag:      platformTag,
		IsCommitScan:     false,
		IsInsecure:       false,
	}
}