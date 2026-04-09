package k8s

import (
	"context"

	"github.com/tungnt127/scb-core-v2/internal/usecase/ports"
)

type EventDrivenWatcher struct{}

func NewEventDrivenWatcher() *EventDrivenWatcher {
	return &EventDrivenWatcher{}
}

func (w *EventDrivenWatcher) Start(ctx context.Context, handler func(context.Context, ports.WatchSignal) error) error {
	// TODO: Implement informer/watch-based Scan CR, Pod, and Job observers.
	// TODO: Do not introduce polling loops; keep Kubernetes event collection and stage evaluation separate.
	<-ctx.Done()
	_ = handler
	return ctx.Err()
}
