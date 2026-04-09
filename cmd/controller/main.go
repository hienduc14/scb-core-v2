package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/tungnt127/scb-core-v2/internal/domain/entities"
	domainservices "github.com/tungnt127/scb-core-v2/internal/domain/services"
	"github.com/tungnt127/scb-core-v2/internal/infrastructure/config"
	"github.com/tungnt127/scb-core-v2/internal/infrastructure/crypto"
	"github.com/tungnt127/scb-core-v2/internal/infrastructure/httpclients"
	"github.com/tungnt127/scb-core-v2/internal/infrastructure/k8s"
	"github.com/tungnt127/scb-core-v2/internal/infrastructure/kafka"
	"github.com/tungnt127/scb-core-v2/internal/infrastructure/logging"
	"github.com/tungnt127/scb-core-v2/internal/infrastructure/scheduler"
	"github.com/tungnt127/scb-core-v2/internal/infrastructure/templates"
	"github.com/tungnt127/scb-core-v2/internal/interfaceadapters/controllers"
	"github.com/tungnt127/scb-core-v2/internal/status"
	usecaseservices "github.com/tungnt127/scb-core-v2/internal/usecase/services"
)

func main() {
	// Load .env file nếu tồn tại (bỏ qua nếu không có — production dùng env system)
	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}

	logger := logging.New()
	registry := domainservices.DefaultRegistry()
	statusEngine := status.NewEngine()
	statusEvaluator := status.NewEvaluator(statusEngine)

	statusPublisher := kafka.NewStatusPublisher(cfg.Kafka.Address, cfg.Kafka.StatusTopic, logger, cfg.Kafka.LocalTestMode)
	errorPublisher := kafka.NewStatusPublisher(cfg.Kafka.Address, cfg.Kafka.StatusTopic, logger, cfg.Kafka.LocalTestMode)
	decryptor := crypto.NewAESDecryptor(cfg.Encryption.Key)
	repositoryVerifier := httpclients.NewRepositoryVerifier()
	imageVerifier := httpclients.NewImageVerifier()
	urlVerifier := httpclients.NewURLVerifier()
	renderer := templates.NewSecureCodeBoxRenderer(cfg.Kubernetes.ScannerNamespace, cfg.Kafka.LocalTestMode)
	submitter, err := k8s.NewScanSubmitter(logger)
	if err != nil {
		panic(fmt.Errorf("init k8s scan submitter: %w", err))
	}
	watcher := k8s.NewEventDrivenWatcher()
	findings := kafka.NewFindingsConsumer(cfg.Kafka.Address, cfg.Kafka.FindingsTopic, cfg.Kafka.RequestConsumerID+"-findings", cfg.Kafka.LocalTestMode)
	cleanupRepository := k8s.NewCleanupRepository()
	cleanupPolicy, err := buildCleanupPolicy(cfg.Cleanup)
	if err != nil {
		panic(err)
	}

	decryptAuth := usecaseservices.NewDecryptAuthTokenService(decryptor)
	verifyTarget := usecaseservices.NewVerifyTargetService(repositoryVerifier, imageVerifier, urlVerifier, registry)
	buildManifest := usecaseservices.NewBuildScanManifestService(renderer, registry)
	submitScan := usecaseservices.NewSubmitScanService(submitter)
	publishStatus := usecaseservices.NewPublishStatusEventService(statusPublisher)
	publishError := usecaseservices.NewPublishErrorEventService(errorPublisher)
	processScan := usecaseservices.NewProcessScanRequestService(
		registry,
		publishStatus,
		publishError,
		decryptAuth,
		verifyTarget,
		buildManifest,
		submitScan,
		submitter,                          // podChecker (ScanSubmitter implements PodChecker)
		cfg.Kubernetes.ScannerNamespace,    // k8sNamespace
		logger,
	)
	handleObservation := usecaseservices.NewHandleK8sObservationService(watcher, findings, publishStatus, publishError, statusEvaluator)
	cleanupScans := usecaseservices.NewCleanupScansService(cleanupRepository, cleanupPolicy, logger)
	cleanupScheduler := scheduler.NewCleanupScheduler(cleanupPolicy.CheckInterval, cleanupScans)

	controller := controllers.NewKafkaScanRequestController(processScan)
	consumer := kafka.NewConsumer(cfg.Kafka.Address, cfg.Kafka.RequestTopic, cfg.Kafka.RequestConsumerID, logger, cfg.Kafka.LocalTestMode, cfg.Kafka.WorkerPoolSize, cfg.Kafka.MinQueueBuffer)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 3)

	go func() {
		errCh <- consumer.Start(ctx, controller.Handle)
	}()
	go func() {
		errCh <- handleObservation.Start(ctx)
	}()
	go func() {
		// handleObservation isn't enough, we also need findings logic if it uses Start
		// actually handleObservation.Start(ctx) probably uses it inside.
		errCh <- cleanupScheduler.Start(ctx)
	}()

	select {
	case err := <-errCh:
		if err != nil && !errors.Is(err, context.Canceled) {
			logger.Error("controller stopped", map[string]any{"error": err.Error()})
			os.Exit(1)
		}
	case <-ctx.Done():
	}
}

func buildCleanupPolicy(cfg config.CleanupConfig) (entities.CleanupPolicy, error) {
	completedMinutes, err := strconv.Atoi(cfg.CompletedRetentionMinutes)
	if err != nil {
		return entities.CleanupPolicy{}, fmt.Errorf("parse MINUTES: %w", err)
	}
	failedDays, err := strconv.Atoi(cfg.FailedRetentionDays)
	if err != nil {
		return entities.CleanupPolicy{}, fmt.Errorf("parse ERROR_SCAN_RETENTION_DAYS: %w", err)
	}
	checkMinutes, err := strconv.Atoi(cfg.CheckIntervalMinutes)
	if err != nil {
		return entities.CleanupPolicy{}, fmt.Errorf("parse ERROR_SCAN_CHECK_MINUTES: %w", err)
	}

	return entities.CleanupPolicy{
		CompletedRetention: time.Duration(completedMinutes) * time.Minute,
		FailedRetention:    time.Duration(failedDays) * 24 * time.Hour,
		CheckInterval:      time.Duration(checkMinutes) * time.Minute,
	}, nil
}
