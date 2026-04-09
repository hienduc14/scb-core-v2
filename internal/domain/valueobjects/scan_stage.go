package valueobjects

type ScanStage string

const (
	StageInitializing ScanStage = "initializing"
	StageCloning      ScanStage = "cloning"
	StageScanning     ScanStage = "scanning"
	StageParsing      ScanStage = "parsing"
	StageHooking      ScanStage = "hooking"
	StageResults      ScanStage = "results"
)

var orderedStages = []ScanStage{
	StageInitializing,
	StageCloning,
	StageScanning,
	StageParsing,
	StageHooking,
	StageResults,
}

func OrderedStages() []ScanStage {
	out := make([]ScanStage, len(orderedStages))
	copy(out, orderedStages)
	return out
}

func (s ScanStage) IsValid() bool {
	for _, stage := range orderedStages {
		if s == stage {
			return true
		}
	}
	return false
}

func (s ScanStage) Rank() int {
	for i, stage := range orderedStages {
		if s == stage {
			return i
		}
	}
	return -1
}

func (s ScanStage) IsForwardFrom(previous ScanStage) bool {
	return s.Rank() > previous.Rank()
}
