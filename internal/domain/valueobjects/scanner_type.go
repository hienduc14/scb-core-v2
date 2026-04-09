package valueobjects

type ScannerType string

const (
	ScannerSemgrep    ScannerType = "semgrep"
	ScannerGitleaks   ScannerType = "gitleaks"
	ScannerKICS       ScannerType = "kics"
	ScannerTrivyImage ScannerType = "trivy-image"
	ScannerNuclei     ScannerType = "nuclei"
)

func (s ScannerType) IsRepositoryScanner() bool {
	switch s {
	case ScannerSemgrep, ScannerGitleaks, ScannerKICS:
		return true
	default:
		return false
	}
}

func (s ScannerType) IsImageScanner() bool {
	return s == ScannerTrivyImage
}

func (s ScannerType) IsURLScanner() bool {
	return s == ScannerNuclei
}
