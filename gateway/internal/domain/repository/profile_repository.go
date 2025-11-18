package repository

import (
	"context"

	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/entity"
)

type ProfileRepository interface {
	Create(ctx context.Context, profile *entity.Profile) error
	Update(ctx context.Context, profile *entity.Profile) error
	GetByID(ctx context.Context, profileID uint64) (*entity.Profile, error)
	GetByUserID(ctx context.Context, userID uint64) ([]*entity.Profile, error)
}
