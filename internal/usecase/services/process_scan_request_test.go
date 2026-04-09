package services

import (
	"context"
	"errors"
	"testing"

	"github.com/tungnt127/scb-core-v2/internal/domain/entities"
	domainservices "github.com/tungnt127/scb-core-v2/internal/domain/services"
	"github.com/tungnt127/scb-core-v2/internal/domain/valueobjects"
)

func TestProcessScanRequestHappyPath(t *testing.T) {
	statusPublisher := &statusPublisherMock{}
	errorPublisher := &errorPublisherMock{}
	decryptor := &decryptAuthMock{
		out: entities.ScanRequest{
			AuthToken: "plain",
		},
	}
	verifier := &verifyTargetMock{
		result: entities.VerificationResult{Accepted: true},
	}
	builder := &buildManifestMock{
		manifest: entities.ScanManifestSpec{
			RequestUUID: "req-1",
			Scanner:     valueobjects.ScannerSemgrep,
		},
	}
	submitter := &submitScanMock{}
	service := NewProcessScanRequestService(
		domainservices.DefaultRegistry(),
		statusPublisher,
		errorPublisher,
		decryptor,
		verifier,
		builder,
		submitter,
		nil,  // podChecker
		"",   // k8sNamespace
		nil,
	)

	err := service.Execute(context.Background(), entities.ScanRequest{
		RequestUUID:  "req-1",
		ScannerAlias: "FS_CODE",
		ProjectType:  valueobjects.ProjectTypeGitHub,
		Target: entities.ScanTarget{
			RepositoryURL: "https://github.com/acme/repo",
		},
		AuthToken: "cipher",
	})
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if len(statusPublisher.events) != 2 {
		t.Fatalf("expected 2 status events, got %d", len(statusPublisher.events))
	}
	if submitter.called != 1 {
		t.Fatalf("expected submit to be called once, got %d", submitter.called)
	}
	if len(errorPublisher.events) != 0 {
		t.Fatalf("expected no error events, got %d", len(errorPublisher.events))
	}
}

func TestProcessScanRequestPublishesFailure(t *testing.T) {
	service := NewProcessScanRequestService(
		domainservices.DefaultRegistry(),
		&statusPublisherMock{},
		&errorPublisherMock{},
		&decryptAuthMock{},
		&verifyTargetMock{err: errors.New("boom")},
		&buildManifestMock{},
		&submitScanMock{},
		nil,  // podChecker
		"",   // k8sNamespace
		nil,
	)

	err := service.Execute(context.Background(), entities.ScanRequest{
		RequestUUID:  "req-1",
		ScannerAlias: "FS_IMAGE",
		Target: entities.ScanTarget{
			ImageRef: "alpine:latest",
		},
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

type statusPublisherMock struct {
	events []entities.ScanStatusEvent
}

func (m *statusPublisherMock) Execute(_ context.Context, event entities.ScanStatusEvent) error {
	m.events = append(m.events, event)
	return nil
}

type errorPublisherMock struct {
	events []entities.ScanStatusEvent
}

func (m *errorPublisherMock) Execute(_ context.Context, event entities.ScanStatusEvent) error {
	m.events = append(m.events, event)
	return nil
}

type decryptAuthMock struct {
	out entities.ScanRequest
	err error
}

func (m *decryptAuthMock) Execute(_ context.Context, request entities.ScanRequest) (entities.ScanRequest, error) {
	if m.err != nil {
		return request, m.err
	}
	if m.out.AuthToken != "" {
		request.AuthToken = m.out.AuthToken
	}
	return request, nil
}

type verifyTargetMock struct {
	result entities.VerificationResult
	err    error
}

func (m *verifyTargetMock) Execute(_ context.Context, _ entities.ScanRequest) (entities.VerificationResult, error) {
	return m.result, m.err
}

type buildManifestMock struct {
	manifest entities.ScanManifestSpec
	err      error
}

func (m *buildManifestMock) Execute(_ context.Context, _ entities.ScanRequest) (entities.ScanManifestSpec, error) {
	return m.manifest, m.err
}

type submitScanMock struct {
	called int
	err    error
}

func (m *submitScanMock) Execute(_ context.Context, _ entities.ScanManifestSpec) error {
	m.called++
	return m.err
}
