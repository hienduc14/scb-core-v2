package entities

import "github.com/tungnt127/scb-core-v2/internal/domain/valueobjects"

type ScanStatusEvent struct {
	RequestUUID string
	Scanner     valueobjects.ScannerType
	Stage       valueobjects.ScanStage
	Status      string
	Message     string
	Terminal    bool
	Error       *ScanError
	Metadata    map[string]string
}
