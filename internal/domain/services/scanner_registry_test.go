package services

import (
	"testing"

	"github.com/tungnt127/scb-core-v2/internal/domain/entities"
	"github.com/tungnt127/scb-core-v2/internal/domain/valueobjects"
)

func TestDefaultRegistryResolveAliases(t *testing.T) {
	registry := DefaultRegistry()

	tests := []struct {
		alias   string
		scanner valueobjects.ScannerType
	}{
		{alias: "FS_CODE", scanner: valueobjects.ScannerSemgrep},
		{alias: "FS_SECRET_COMMIT", scanner: valueobjects.ScannerGitleaks},
		{alias: "FS_IAC", scanner: valueobjects.ScannerKICS},
		{alias: "FS_IMAGE", scanner: valueobjects.ScannerTrivyImage},
		{alias: "DAST_NUCLEI", scanner: valueobjects.ScannerNuclei},
		{alias: "nuclei", scanner: valueobjects.ScannerNuclei},
	}

	for _, tc := range tests {
		strategy, err := registry.Resolve(tc.alias)
		if err != nil {
			t.Fatalf("resolve %s: %v", tc.alias, err)
		}
		if strategy.Scanner() != tc.scanner {
			t.Fatalf("alias %s mapped to %s", tc.alias, strategy.Scanner())
		}
	}
}

func TestNucleiNormalizationMovesProjectURIToTargetURL(t *testing.T) {
	registry := DefaultRegistry()
	strategy, err := registry.Resolve("NUCLEI")
	if err != nil {
		t.Fatal(err)
	}

	request, err := strategy.Normalize(entities.ScanRequest{
		ScannerAlias: "NUCLEI",
		Target: entities.ScanTarget{
			RepositoryURL: "https://example.com",
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	if request.Target.TargetURL != "https://example.com" {
		t.Fatalf("expected target URL to be normalized, got %q", request.Target.TargetURL)
	}
}
