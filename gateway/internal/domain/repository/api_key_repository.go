package repository

import (
	"context"

	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/entity"
)

type APIKeyRepository interface {
	Create(ctx context.Context, apiKey *entity.APIKey) error
	GetByPrefix(ctx context.Context, prefix string) (*entity.APIKey, error)
	GetActiveByProfileID(ctx context.Context, profileID uint64) (*entity.APIKey, error)
	Revoke(ctx context.Context, keyID string) error
}
