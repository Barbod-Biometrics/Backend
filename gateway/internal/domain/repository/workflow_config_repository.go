package repository

import (
	"context"

	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/entity"
)

type WorkflowConfigRepository interface {
	Save(ctx context.Context, cfg *entity.WorkflowConfig) error
	GetByID(ctx context.Context, id uint64) (*entity.WorkflowConfig, error)
	ListByProfileID(ctx context.Context, profileID uint64) ([]*entity.WorkflowConfig, error)
	Update(ctx context.Context, cfg *entity.WorkflowConfig) error
	Delete(ctx context.Context, id uint64) error
}
