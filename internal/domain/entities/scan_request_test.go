package entities

import (
	"testing"

	"github.com/tungnt127/scb-core-v2/internal/domain/valueobjects"
)

func TestScanRequestValidateRepositoryScanner(t *testing.T) {
	request := ScanRequest{
		RequestUUID:  "req-1",
		ScannerAlias: "FS_CODE",
		Scanner:      valueobjects.ScannerSemgrep,
		ProjectType:  valueobjects.ProjectTypeGitHub,
		Target: ScanTarget{
			RepositoryURL: "https://github.com/acme/repo",
		},
	}

	if err := request.Validate(); err != nil {
		t.Fatalf("validate: %v", err)
	}
}

func TestScanRequestValidateFailsWithoutTarget(t *testing.T) {
	request := ScanRequest{
		RequestUUID:  "req-1",
		ScannerAlias: "FS_IMAGE",
		Scanner:      valueobjects.ScannerTrivyImage,
	}

	if err := request.Validate(); err == nil {
		t.Fatal("expected validation error")
	}
}
