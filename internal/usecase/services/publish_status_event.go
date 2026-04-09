package services

import (
	"context"

	"github.com/tungnt127/scb-core-v2/internal/domain/entities"
	"github.com/tungnt127/scb-core-v2/internal/usecase/ports"
)

type PublishStatusEventService struct {
	publisher ports.StatusEventPublisher
}

func NewPublishStatusEventService(publisher ports.StatusEventPublisher) *PublishStatusEventService {
	return &PublishStatusEventService{publisher: publisher}
}

func (s *PublishStatusEventService) Execute(ctx context.Context, event entities.ScanStatusEvent) error {
	return s.publisher.PublishStatusEvent(ctx, event)
}
