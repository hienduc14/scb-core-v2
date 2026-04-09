package controllers

import (
	"context"

	"github.com/tungnt127/scb-core-v2/internal/domain/entities"
	"github.com/tungnt127/scb-core-v2/internal/usecase/contracts"
)

type KafkaScanRequestController struct {
	process contracts.ProcessScanRequestUseCase
}

func NewKafkaScanRequestController(process contracts.ProcessScanRequestUseCase) *KafkaScanRequestController {
	return &KafkaScanRequestController{process: process}
}

func (c *KafkaScanRequestController) Handle(ctx context.Context, request entities.ScanRequest) error {
	return c.process.Execute(ctx, request)
}
