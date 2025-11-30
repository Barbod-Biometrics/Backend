package service

import (
	"context"

	"github.com/Barbod-Biometrics/Backend/gateway/internal/application/usecase"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/repository"
)

type APIKeyService struct {
	apiKeyRepo repository.APIKeyRepository
}

var _ usecase.APIKeyService = (*APIKeyService)(nil)

func NewAPIKeyService(apiKeyRepo repository.APIKeyRepository) *APIKeyService {
	return &APIKeyService{
		apiKeyRepo: apiKeyRepo,
	}
}

func (s *APIKeyService) GenerateKey(ctx context.Context, profileID uint64) (string, error) {
	return "", nil
}

func (s *APIKeyService) RegenerateKey(ctx context.Context, profileID uint64) (string, error) {
	return "", nil
}

func (s *APIKeyService) Authenticate(ctx context.Context, rawKey string) (uint64, error) {
	return 0, nil
}
