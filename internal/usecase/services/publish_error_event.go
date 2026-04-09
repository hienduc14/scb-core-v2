package services

import (
	"context"

	"github.com/tungnt127/scb-core-v2/internal/domain/entities"
	"github.com/tungnt127/scb-core-v2/internal/usecase/ports"
)

type PublishErrorEventService struct {
	publisher ports.ErrorEventPublisher
}

func NewPublishErrorEventService(publisher ports.ErrorEventPublisher) *PublishErrorEventService {
	return &PublishErrorEventService{publisher: publisher}
}

func (s *PublishErrorEventService) Execute(ctx context.Context, event entities.ScanStatusEvent) error {
	return s.publisher.PublishErrorEvent(ctx, event)
}
