package status

import (
	"testing"

	"github.com/tungnt127/scb-core-v2/internal/domain/entities"
	"github.com/tungnt127/scb-core-v2/internal/domain/valueobjects"
)

func TestStatusEngineSuppressesDuplicatesAndBackwardsStages(t *testing.T) {
	engine := NewEngine()
	state := StatusState{}

	first := engine.Advance(state, entities.ScanStatusEvent{Stage: valueobjects.StageInitializing})
	if !first.Publish {
		t.Fatal("expected first stage to publish")
	}

	second := engine.Advance(first.State, entities.ScanStatusEvent{Stage: valueobjects.StageInitializing})
	if second.Publish {
		t.Fatal("expected duplicate stage to be suppressed")
	}

	third := engine.Advance(first.State, entities.ScanStatusEvent{Stage: valueobjects.StageCloning})
	if !third.Publish {
		t.Fatal("expected forward stage to publish")
	}

	backwards := engine.Advance(third.State, entities.ScanStatusEvent{Stage: valueobjects.StageInitializing})
	if backwards.Publish {
		t.Fatal("expected backwards stage to be suppressed")
	}
}

func TestStatusEngineTerminalStopsFurtherEvents(t *testing.T) {
	engine := NewEngine()

	state := engine.Advance(StatusState{}, entities.ScanStatusEvent{
		Stage:    valueobjects.StageResults,
		Terminal: true,
	}).State

	decision := engine.Advance(state, entities.ScanStatusEvent{Stage: valueobjects.StageResults})
	if decision.Publish {
		t.Fatal("expected terminal state to stop future events")
	}
}
