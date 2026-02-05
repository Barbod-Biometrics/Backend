package postgres

import (
	"context"

	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/entity"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/repository"
	"gorm.io/gorm"
)

type SessionRepository struct {
	db *gorm.DB
}

func NewSessionRepository(db *gorm.DB) repository.SessionRepository {
	return &SessionRepository{db: db}
}

func (r *SessionRepository) Save(ctx context.Context, s *entity.Session) error {
	return r.db.WithContext(ctx).Create(s).Error
}

func (r *SessionRepository) GetByID(ctx context.Context, id string) (*entity.Session, error) {
	var s entity.Session
	if err := r.db.WithContext(ctx).First(&s, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *SessionRepository) Update(ctx context.Context, s *entity.Session) error {
	return r.db.WithContext(ctx).Save(s).Error
}

func (r *SessionRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&entity.Session{}, "id = ?", id).Error
}

func (r *SessionRepository) ListByProfileID(ctx context.Context, profileID uint64) ([]*entity.Session, error) {
	var list []*entity.Session
	if err := r.db.WithContext(ctx).Where("profile_id = ?", profileID).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}
