package status

import (
	"github.com/tungnt127/scb-core-v2/internal/domain/entities"
	"github.com/tungnt127/scb-core-v2/internal/domain/valueobjects"
)

type StatusState struct {
	Current  valueobjects.ScanStage
	Terminal bool
}

type StatusDecision struct {
	Publish bool
	Event   entities.ScanStatusEvent
	State   StatusState
}

type Engine struct{}

func NewEngine() *Engine {
	return &Engine{}
}

func (e *Engine) Advance(state StatusState, candidate entities.ScanStatusEvent) StatusDecision {
	if state.Terminal {
		return StatusDecision{Publish: false, State: state}
	}
	if !candidate.Stage.IsValid() {
		return StatusDecision{Publish: false, State: state}
	}
	if state.Current != "" && candidate.Stage.Rank() <= state.Current.Rank() {
		return StatusDecision{Publish: false, State: state}
	}

	nextState := StatusState{
		Current:  candidate.Stage,
		Terminal: candidate.Terminal,
	}
	return StatusDecision{
		Publish: true,
		Event:   candidate,
		State:   nextState,
	}
}

func (e *Engine) Fail(state StatusState, scanner valueobjects.ScannerType, requestUUID string, scanErr entities.ScanError) StatusDecision {
	if state.Terminal {
		return StatusDecision{Publish: false, State: state}
	}

	stage := scanErr.Stage
	if !stage.IsValid() {
		stage = state.Current
		if !stage.IsValid() {
			stage = valueobjects.StageInitializing
		}
	}

	event := entities.ScanStatusEvent{
		RequestUUID: requestUUID,
		Scanner:     scanner,
		Stage:       stage,
		Status:      "failed",
		Message:     scanErr.Message,
		Terminal:    true,
		Error:       &scanErr,
	}
	return StatusDecision{
		Publish: true,
		Event:   event,
		State: StatusState{
			Current:  stage,
			Terminal: true,
		},
	}
}
