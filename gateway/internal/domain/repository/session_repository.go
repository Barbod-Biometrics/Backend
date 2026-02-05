package repository

import (
	"context"

	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/entity"
)

type SessionRepository interface {
	Save(ctx context.Context, s *entity.Session) error
	GetByID(ctx context.Context, id string) (*entity.Session, error)
	Update(ctx context.Context, s *entity.Session) error
	Delete(ctx context.Context, id string) error
	ListByProfileID(ctx context.Context, profileID uint64) ([]*entity.Session, error)
}
