package templates

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/tungnt127/scb-core-v2/internal/domain/entities"
	"github.com/tungnt127/scb-core-v2/internal/domain/valueobjects"
)

// templateFS embeds tất cả file .tmpl vào binary khi compile.
// Không cần đọc từ disk lúc runtime → phù hợp khi deploy trong container.
//
//go:embed configs/*.tmpl
var templateFS embed.FS

// ScanTemplateInput là data struct được bind vào text/template.
// Ánh xạ 1-1 với các placeholder trong Python cũ (string.replace).
type ScanTemplateInput struct {
	// Chung cho tất cả scanner
	ScanID           string // UUID gốc
	ScanName         string // requestUUID — dùng làm tên Scan CR
	Namespace        string
	RepoURL          string // URL gốc (https://gitlab...)
	RepoURLWithToken string // https://oauth2:$(GIT_TOKEN)@gitlab... — embed vào git clone
	AuthToken        string // đã decrypt (plain text)
	Branch           string
	BaseCommit       string
	EffectiveCommit  string // Commit thực tế dùng để scan (chỉ apply cho một số scanner)
	IsCommitScan     bool // true nếu BaseCommit != "" (FS_*_COMMIT)
	ScannerAlias     string

	// Trivy-specific
	ImageRef      string // registry/image:tag
	TrivyUsername string
	TrivyPassword string
	PlatformTag   string // linux/amd64 (default)
	IsInsecure    bool
}

type SecureCodeBoxRenderer struct {
	namespace     string
	localTestMode bool // nếu true: ghi YAML đã render ra folder debug-manifests/
}

func NewSecureCodeBoxRenderer(namespace string, localTestMode bool) *SecureCodeBoxRenderer {
	return &SecureCodeBoxRenderer{namespace: namespace, localTestMode: localTestMode}
}

// RenderManifest chọn file .tmpl phù hợp theo scanner, bind ScanTemplateInput,
// render vào bytes.Buffer (in-memory), và trả về ScanManifestSpec với RenderedYAML.
// TUYỆT ĐỐI KHÔNG ghi file ra /tmp hay dùng os.Exec.
func (r *SecureCodeBoxRenderer) RenderManifest(_ context.Context, request entities.ScanRequest, input map[string]any) (entities.ScanManifestSpec, error) {
	// 1. Chọn file template theo scanner type
	tmplFile, err := resolveTemplateName(request.Scanner)
	if err != nil {
		return entities.ScanManifestSpec{}, err
	}

	// 2. Build data struct từ request + input map (chia case theo từng Scanner)
	var data ScanTemplateInput
	switch request.Scanner {
	case valueobjects.ScannerGitleaks:
		data = buildGitleaksTemplateInput(request, input, r.namespace)
	case valueobjects.ScannerSemgrep:
		data = buildSemgrepTemplateInput(request, input, r.namespace)
	case valueobjects.ScannerKICS:
		data = buildKicsTemplateInput(request, input, r.namespace)
	case valueobjects.ScannerTrivyImage:
		data = buildTrivyTemplateInput(request, input, r.namespace)
	default:
		data = buildCommonTemplateInput(request, input, r.namespace)
	}

	// 3. Parse template từ embed.FS
	tmpl, err := template.New(tmplFile).Funcs(templateFuncMap()).ParseFS(templateFS, "configs/"+tmplFile)
	if err != nil {
		return entities.ScanManifestSpec{}, fmt.Errorf("parse template %q: %w", tmplFile, err)
	}

	// 4. Execute template → bytes.Buffer (in-memory, không ghi file)
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return entities.ScanManifestSpec{}, fmt.Errorf("render template %q: %w", tmplFile, err)
	}

	// 5. [DEBUG] Nếu LOCAL_TEST_MODE=true: dump YAML ra folder debug-manifests/
	// Dùng để kiểm tra nội dung manifest trước khi gắn lên K8s.
	// AuthToken bị redact (***) để không lộ secret trong file debug.
	// if r.localTestMode {
	r.writeDebugManifest(request.RequestUUID, string(request.Scanner), buf.Bytes(), data.AuthToken)
	// }

	return entities.ScanManifestSpec{
		RequestUUID:  request.RequestUUID,
		Scanner:      request.Scanner,
		Namespace:    r.namespace,
		RenderedYAML: buf.Bytes(),
	}, nil
}

// resolveTemplateName ánh xạ scanner type → tên file .tmpl tương ứng.
func resolveTemplateName(scanner valueobjects.ScannerType) (string, error) {
	switch scanner {
	case valueobjects.ScannerGitleaks:
		return "gitleaks-scan.tmpl", nil
	case valueobjects.ScannerSemgrep:
		return "semgrep-scan.tmpl", nil
	case valueobjects.ScannerKICS:
		return "kics-scan.tmpl", nil
	case valueobjects.ScannerTrivyImage:
		return "trivy-image-scan.tmpl", nil
	default:
		return "", fmt.Errorf("no template defined for scanner %q", scanner)
	}
}


// templateFuncMap cung cấp các function bổ sung dùng trong template.
// Ví dụ: {{not .IsCommitScan}} thay vì phải dùng {{if not}}
func templateFuncMap() template.FuncMap {
	return template.FuncMap{
		"not": func(v bool) bool { return !v },
	}
}

// writeDebugManifest ghi YAML đã render ra file trong folder debug-manifests/
// (cùng thư mục với file binary controller.exe).
// Tên file: debug-manifests/<requestUUID>-<scanner>.yaml
// AuthToken được redact bằng *** để tránh lộ secret trong log/file debug.
func (r *SecureCodeBoxRenderer) writeDebugManifest(requestUUID, scanner string, yamlBytes []byte, authToken string) {
	// Dùng đường dẫn tuyệt đối cạnh binary — không phụ thuộc working directory
	exePath, err := os.Executable()
	if err != nil {
		fmt.Printf("[debug] cannot get executable path: %v\n", err)
		return
	}
	debugDir := filepath.Join(filepath.Dir(exePath), "debug-manifests")

	if err := os.MkdirAll(debugDir, 0o755); err != nil {
		fmt.Printf("[debug] cannot create debug dir: %v\n", err)
		return
	}

	// Redact authToken trước khi ghi — không để token thật trong file debug
	content := string(yamlBytes)
	if authToken != "" {
		content = strings.ReplaceAll(content, authToken, "***REDACTED***")
	}

	filename := filepath.Join(debugDir, fmt.Sprintf("%s-%s.yaml", requestUUID, scanner))
	if err := os.WriteFile(filename, []byte(content), 0o644); err != nil {
		fmt.Printf("[debug] cannot write debug manifest: %v\n", err)
		return
	}
	fmt.Printf("[debug] manifest written to %s\n", filename)
}

