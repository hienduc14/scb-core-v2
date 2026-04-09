package entities

type VerificationResult struct {
	Accepted bool
	Metadata map[string]string
	Reason   string
}
