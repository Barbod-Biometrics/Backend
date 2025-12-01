package postgres

import (
	"context"
	"errors"
	"strconv"

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
	var apiKey entity.APIKey
	db := r.getDB(ctx)

	err := db.WithContext(ctx).Where("key_prefix = ?", prefix).First(&apiKey).Error
	if err != nil {
		return nil, err
	}
	return &apiKey, nil
}

func (r *ApiKeyRepository) GetActiveByProfileID(ctx context.Context, profileID uint64) (*entity.APIKey, error) {
	var apiKey entity.APIKey
	db := r.getDB(ctx)

	err := db.WithContext(ctx).Where("profile_id = ? AND is_active = ?", profileID, true).First(&apiKey).Error
	if err != nil {
		return nil, err
	}
	return &apiKey, nil
}

func (r *ApiKeyRepository) Revoke(ctx context.Context, keyID string) error {
	db := r.getDB(ctx)

	id, err := strconv.ParseUint(keyID, 10, 64)
	if err != nil {
		return errors.New("invalid key id format")
	}

	result := db.WithContext(ctx).Model(&entity.APIKey{}).Where("ke_id = ?", id).Update("is_active", false)
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("key not found or already revoked")
	}

	return nil
}

func (r *ApiKeyRepository) getDB(ctx context.Context) *gorm.DB {
	if tx, ok := ctx.Value("db_tx").(*gorm.DB); ok {
		return tx
	}
	return r.db
}
