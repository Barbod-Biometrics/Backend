package postgres

import (
	"context"

	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/entity"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/repository"
	"gorm.io/gorm"
)

type ApiKeyRepository struct {
	db *gorm.DB
}

var _ repository.APIKeyRepository = (*ApiKeyRepository)(nil)

func NewApiKeyRepository(db *gorm.DB) *ApiKeyRepository {
	return &ApiKeyRepository{db: db}
}

func (r *ApiKeyRepository) Create(ctx context.Context, apiKey *entity.APIKey) error {
	db := r.getDB(ctx)
	return db.WithContext(ctx).Create(apiKey).Error
}

func (r *ApiKeyRepository) GetByPrefix(ctx context.Context, prefix string) (*entity.APIKey, error) {
	return nil, nil
}

func (r *ApiKeyRepository) GetActiveByProfileID(ctx context.Context, profileID uint64) (*entity.APIKey, error) {
	return nil, nil
}

func (r *ApiKeyRepository) Revoke(ctx context.Context, keyID string) error {
	return nil
}

func (r *ApiKeyRepository) getDB(ctx context.Context) *gorm.DB {
	if tx, ok := ctx.Value("db_tx").(*gorm.DB); ok {
		return tx
	}
	return r.db
}
