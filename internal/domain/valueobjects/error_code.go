package valueobjects

type ErrorCode string

const (
	ErrInvalidRequest     ErrorCode = "INVALID_REQUEST"
	ErrUnsupportedScanner ErrorCode = "UNSUPPORTED_SCANNER"
	ErrAuthDecrypt        ErrorCode = "AUTH_DECRYPT_FAILED"
	ErrTargetVerification ErrorCode = "TARGET_VERIFICATION_FAILED"
	ErrManifestBuild      ErrorCode = "MANIFEST_BUILD_FAILED"
	ErrScanSubmission     ErrorCode = "SCAN_SUBMISSION_FAILED"
	ErrStatusPublish      ErrorCode = "STATUS_PUBLISH_FAILED"
	ErrScanObservation    ErrorCode = "SCAN_OBSERVATION_FAILED"
	ErrCleanup            ErrorCode = "CLEANUP_FAILED"
	ErrInfrastructure     ErrorCode = "INFRASTRUCTURE_ERROR"
	ErrRepositoryLookup   ErrorCode = "REPOSITORY_LOOKUP_FAILED"
	ErrImageLookup        ErrorCode = "IMAGE_LOOKUP_FAILED"
	ErrURLVerification    ErrorCode = "URL_VERIFICATION_FAILED"
	ErrCloneExecution     ErrorCode = "CLONE_FAILED"
	ErrScanExecution      ErrorCode = "SCAN_FAILED"
	ErrParseExecution     ErrorCode = "PARSE_FAILED"
	ErrHookExecution      ErrorCode = "HOOK_FAILED"
)

func (c ErrorCode) IsTerminal() bool {
	return c != ErrStatusPublish
}
