package status

import (
	"github.com/tungnt127/scb-core-v2/internal/domain/entities"
	"github.com/tungnt127/scb-core-v2/internal/domain/valueobjects"
	"github.com/tungnt127/scb-core-v2/internal/usecase/ports"
)

type Evaluation struct {
	StatusEvent *entities.ScanStatusEvent
	ErrorEvent  *entities.ScanStatusEvent
}

type Evaluator struct {
	engine *Engine
}

func NewEvaluator(engine *Engine) *Evaluator {
	return &Evaluator{engine: engine}
}

func (e *Evaluator) Evaluate(state StatusState, signal ports.WatchSignal) (Evaluation, StatusState) {
	if signal.Type == ports.WatchSignalScanFailure && signal.Error != nil {
		decision := e.engine.Fail(state, "", signal.RequestUUID, *signal.Error)
		if !decision.Publish {
			return Evaluation{}, state
		}
		return Evaluation{ErrorEvent: &decision.Event}, decision.State
	}

	stage := valueobjects.ScanStage(signal.Stage)
	if signal.Type == ports.WatchSignalFindings {
		stage = valueobjects.StageResults
	}

	decision := e.engine.Advance(state, entities.ScanStatusEvent{
		RequestUUID: signal.RequestUUID,
		Stage:       stage,
		Status:      "running",
		Message:     signal.Message,
		Terminal:    stage == valueobjects.StageResults,
		Metadata:    signal.Metadata,
	})
	if !decision.Publish {
		return Evaluation{}, state
	}
	if decision.Event.Terminal {
		decision.Event.Status = "completed"
	}
	return Evaluation{StatusEvent: &decision.Event}, decision.State
}
