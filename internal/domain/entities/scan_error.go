package entities

import "github.com/tungnt127/scb-core-v2/internal/domain/valueobjects"

type ScanError struct {
	Code       valueobjects.ErrorCode
	Stage      valueobjects.ScanStage
	Message    string
	Cause      string
	Context    map[string]string
	LogExcerpt string
}
