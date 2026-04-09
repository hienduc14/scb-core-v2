package k8s

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/dynamic"
	sigsyaml "sigs.k8s.io/yaml"

	"github.com/tungnt127/scb-core-v2/internal/domain/entities"
	"github.com/tungnt127/scb-core-v2/internal/usecase/ports"
)

// scanGVR định nghĩa GroupVersionResource của SecureCodeBox Scan CRD.
// kubectl api-resources --api-group=execution.securecodebox.io
// → NAME: scans, APIVERSION: execution.securecodebox.io/v1, KIND: Scan
var scanGVR = schema.GroupVersionResource{
	Group:    "execution.securecodebox.io",
	Version:  "v1",
	Resource: "scans",
}

// ScanSubmitter implement cả ports.ScanSubmitter và ports.PodChecker.
type ScanSubmitter struct {
	dynClient dynamic.Interface
	logger    ports.Logger
}

// NewScanSubmitter tạo ScanSubmitter với dynamic K8s client.
// Tự động chọn in-cluster hoặc local kubeconfig.
func NewScanSubmitter(logger ports.Logger) (*ScanSubmitter, error) {
	dynClient, err := buildDynamicClient()
	if err != nil {
		return nil, fmt.Errorf("init scan submitter: %w", err)
	}
	return &ScanSubmitter{
		dynClient: dynClient,
		logger:    logger,
	}, nil
}

// SubmitScan apply Scan CR lên K8s thông qua dynamic client.
//
// Flow:
//  1. manifest.RenderedYAML (output của text/template) → yaml.Unmarshal → map[string]interface{}
//  2. Wrap thành unstructured.Unstructured (Kubernetes "JSON động")
//  3. dynamicClient.Resource(scanGVR).Namespace(...).Create(ctx, obj)
//  4. Nếu AlreadyExists → Update (apply semantics, tương đương kubectl apply)
//
// SecureCodeBox Operator sẽ "chộp" Scan CR này và tạo ra các Job/Pod tương ứng.
func (s *ScanSubmitter) SubmitScan(ctx context.Context, manifest entities.ScanManifestSpec) error {
	if len(manifest.RenderedYAML) == 0 {
		return fmt.Errorf("RenderedYAML is empty — renderer did not produce output")
	}

	// Bước 1: YAML → JSON trung gian → map[string]interface{}
	// sigs.k8s.io/yaml xử lý YAML-1.1 → JSON → Go map (khác encoding/json thuần)
	var rawObj map[string]interface{}
	if err := sigsyaml.Unmarshal(manifest.RenderedYAML, &rawObj); err != nil {
		return fmt.Errorf("unmarshal rendered yaml: %w", err)
	}

	// Bước 2: Wrap vào Unstructured — k8s "generic object" không cần Struct cụ thể
	obj := &unstructured.Unstructured{Object: rawObj}

	if s.logger != nil {
		s.logger.Info("submitting scan CR to kubernetes", map[string]any{
			"name":      obj.GetName(),
			"namespace": manifest.Namespace,
			"scanner":   string(manifest.Scanner),
		})
	}

	// Bước 3: Gọi K8s API — Create Scan CR
	resource := s.dynClient.Resource(scanGVR).Namespace(manifest.Namespace)
	_, err := resource.Create(ctx, obj, metav1.CreateOptions{})
	if err == nil {
		if s.logger != nil {
			s.logger.Info("scan CR created successfully", map[string]any{
				"name":      obj.GetName(),
				"namespace": manifest.Namespace,
			})
		}
		return nil
	}

	// Bước 4: Nếu đã tồn tại → Update (kubectl apply semantics)
	if k8serrors.IsAlreadyExists(err) {
		
		if s.logger != nil {
			s.logger.Info("scan CR already exists, updating", map[string]any{
				"name":      obj.GetName(),
				"namespace": manifest.Namespace,
			})
		}
		// Lấy resourceVersion hiện tại để Update hợp lệ
		existing, getErr := resource.Get(ctx, obj.GetName(), metav1.GetOptions{})
		if getErr != nil {
			return fmt.Errorf("get existing scan CR for update: %w", getErr)
		}
		obj.SetResourceVersion(existing.GetResourceVersion())
		_, err = resource.Update(ctx, obj, metav1.UpdateOptions{})
		if err != nil {
			return fmt.Errorf("update scan CR %q: %w", obj.GetName(), err)
		}
		if s.logger != nil {
			s.logger.Info("scan CR updated successfully", map[string]any{
				"name": obj.GetName(),
			})
		}
		return nil
	}

	return fmt.Errorf("create scan CR %q: %w", obj.GetName(), err)
}

// ExistingScanPods trả về danh sách tên pod đang chạy có tên chứa requestUUID.
// TODO: Implement bằng K8s dynamic client để list pods theo label hoặc prefix tên.
func (s *ScanSubmitter) ExistingScanPods(_ context.Context, namespace, requestUUID string) ([]string, error) {
	if s.logger != nil {
		s.logger.Info("pod dedup check (placeholder)", map[string]any{
			"namespace":   namespace,
			"requestUUID": requestUUID,
		})
	}
	// Placeholder: luôn trả về empty → không bao giờ block request
	return nil, nil
}
