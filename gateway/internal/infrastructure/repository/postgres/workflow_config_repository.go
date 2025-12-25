package postgres

import (
	"context"

	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/entity"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/repository"
	"gorm.io/gorm"
)

type WorkflowConfigRepository struct {
	db *gorm.DB
}

func NewWorkflowConfigRepository(db *gorm.DB) repository.WorkflowConfigRepository {
	return &WorkflowConfigRepository{db: db}
}

func (r *WorkflowConfigRepository) Save(ctx context.Context, cfg *entity.WorkflowConfig) error {
	return r.db.WithContext(ctx).Create(cfg).Error
}

func (r *WorkflowConfigRepository) GetByID(ctx context.Context, id uint64) (*entity.WorkflowConfig, error) {
	var cfg entity.WorkflowConfig
	if err := r.db.WithContext(ctx).First(&cfg, id).Error; err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (r *WorkflowConfigRepository) ListByProfileID(ctx context.Context, profileID uint64) ([]*entity.WorkflowConfig, error) {
	var cfgs []*entity.WorkflowConfig
	if err := r.db.WithContext(ctx).Where("profile_id = ?", profileID).Find(&cfgs).Error; err != nil {
		return nil, err
	}
	return cfgs, nil
}

func (r *WorkflowConfigRepository) Update(ctx context.Context, cfg *entity.WorkflowConfig) error {
	return r.db.WithContext(ctx).Save(cfg).Error
}

func (r *WorkflowConfigRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&entity.WorkflowConfig{}, id).Error
}
