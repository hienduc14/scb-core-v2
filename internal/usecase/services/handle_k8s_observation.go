package services

import (
	"context"
	"sync"

	"github.com/tungnt127/scb-core-v2/internal/status"
	"github.com/tungnt127/scb-core-v2/internal/usecase/contracts"
	"github.com/tungnt127/scb-core-v2/internal/usecase/ports"
)

type HandleK8sObservationService struct {
	watcher         ports.ScanWatcher
	findings        ports.FindingsSignalSource
	statusPublisher contracts.PublishStatusEventUseCase
	errorPublisher  contracts.PublishErrorEventUseCase
	evaluator       *status.Evaluator
	mu              sync.Mutex
	states          map[string]status.StatusState
}

func NewHandleK8sObservationService(
	watcher ports.ScanWatcher,
	findings ports.FindingsSignalSource,
	statusPublisher contracts.PublishStatusEventUseCase,
	errorPublisher contracts.PublishErrorEventUseCase,
	evaluator *status.Evaluator,
) *HandleK8sObservationService {
	return &HandleK8sObservationService{
		watcher:         watcher,
		findings:        findings,
		statusPublisher: statusPublisher,
		errorPublisher:  errorPublisher,
		evaluator:       evaluator,
		states:          make(map[string]status.StatusState),
	}
}

func (s *HandleK8sObservationService) Start(ctx context.Context) error {
	if s.watcher != nil {
		go func() { _ = s.watcher.Start(ctx, s.handleSignal) }()
	}
	if s.findings != nil {
		go func() { _ = s.findings.Start(ctx, s.handleSignal) }()
	}
	<-ctx.Done()
	return ctx.Err()
}

func (s *HandleK8sObservationService) handleSignal(ctx context.Context, signal ports.WatchSignal) error {
	s.mu.Lock()
	state := s.states[signal.RequestUUID]
	s.mu.Unlock()

	evaluation, nextState := s.evaluator.Evaluate(state, signal)
	if evaluation.StatusEvent != nil {
		s.storeState(signal.RequestUUID, nextState)
		return s.statusPublisher.Execute(ctx, *evaluation.StatusEvent)
	}
	if evaluation.ErrorEvent != nil {
		s.storeState(signal.RequestUUID, nextState)
		return s.errorPublisher.Execute(ctx, *evaluation.ErrorEvent)
	}
	return nil
}

func (s *HandleK8sObservationService) storeState(requestUUID string, state status.StatusState) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.states[requestUUID] = state
}
