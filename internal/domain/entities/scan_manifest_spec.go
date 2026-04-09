package entities

import "github.com/tungnt127/scb-core-v2/internal/domain/valueobjects"

type ScanManifestSpec struct {
	RequestUUID  string
	Scanner      valueobjects.ScannerType
	Namespace    string
	Labels       map[string]string
	Annotations  map[string]string
	Spec         map[string]any
	RenderedYAML []byte // output của text/template, input của dynamic K8s client
}
