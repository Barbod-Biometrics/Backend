package usecase

import (
	"context"

	"github.com/Barbod-Biometrics/Backend/gateway/internal/application/dto/workflow_config"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/entity"
)

type WorkflowConfigUsecase interface {
	Save(ctx context.Context, profileID uint64, req workflow_config.SaveConfigRequest) (*entity.WorkflowConfig, error)
	ListByProfile(ctx context.Context, profileID uint64) ([]*entity.WorkflowConfig, error)
	GetByID(ctx context.Context, id uint64) (*entity.WorkflowConfig, error)
	Update(ctx context.Context, profileID, id uint64, req workflow_config.UpdateConfigRequest) (*entity.WorkflowConfig, error)
	Delete(ctx context.Context, profileID, id uint64) error
}
