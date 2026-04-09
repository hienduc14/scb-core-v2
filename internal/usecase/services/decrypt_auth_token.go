package services

import (
	"context"
	"strings"

	"github.com/tungnt127/scb-core-v2/internal/domain/entities"
	"github.com/tungnt127/scb-core-v2/internal/usecase/ports"
)

type DecryptAuthTokenService struct {
	decryptor ports.AuthDecryptor
}

func NewDecryptAuthTokenService(decryptor ports.AuthDecryptor) *DecryptAuthTokenService {
	return &DecryptAuthTokenService{decryptor: decryptor}
}

func (s *DecryptAuthTokenService) Execute(ctx context.Context, request entities.ScanRequest) (entities.ScanRequest, error) {
	if s.decryptor == nil || strings.TrimSpace(request.AuthToken) == "" {
		return request, nil
	}

	plain, err := s.decryptor.Decrypt(ctx, request.AuthToken)
	if err != nil {
		return request, err
	}
	request.AuthToken = plain
	return request, nil
}
