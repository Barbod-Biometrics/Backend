package service

import (
	"context"

	"github.com/Barbod-Biometrics/Backend/gateway/internal/application/dto/workflow_config"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/entity"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/exception"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/logger"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/repository"
)

type WorkflowConfigService interface {
	Save(ctx context.Context, profileID uint64, req workflow_config.SaveConfigRequest) (*entity.WorkflowConfig, error)
	ListByProfile(ctx context.Context, profileID uint64) ([]*entity.WorkflowConfig, error)
	GetByID(ctx context.Context, id uint64) (*entity.WorkflowConfig, error)
	Update(ctx context.Context, profileID, id uint64, req workflow_config.UpdateConfigRequest) (*entity.WorkflowConfig, error)
	Delete(ctx context.Context, profileID, id uint64) error
}

type workflowConfigService struct {
	repo repository.WorkflowConfigRepository
	l    logger.Logger
}

func NewWorkflowConfigService(repo repository.WorkflowConfigRepository, l logger.Logger) WorkflowConfigService {
	return &workflowConfigService{repo: repo, l: l}
}

func (s *workflowConfigService) Save(ctx context.Context, profileID uint64, req workflow_config.SaveConfigRequest) (*entity.WorkflowConfig, error) {
	cfg := &entity.WorkflowConfig{
		ProfileID:        profileID,
		Name:             req.Name,
		Instruction:      req.Instruction,
		LivenessSentence: req.LivenessSentence,
	}
	if err := s.repo.Save(ctx, cfg); err != nil {
		s.l.Error("Failed to save workflow config", logger.Field{Key: "error", Value: err})
		return nil, err
	}
	return cfg, nil
}

func (s *workflowConfigService) ListByProfile(ctx context.Context, profileID uint64) ([]*entity.WorkflowConfig, error) {
	return s.repo.ListByProfileID(ctx, profileID)
}

func (s *workflowConfigService) GetByID(ctx context.Context, id uint64) (*entity.WorkflowConfig, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *workflowConfigService) Update(ctx context.Context, profileID, id uint64, req workflow_config.UpdateConfigRequest) (*entity.WorkflowConfig, error) {
	cfg, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if cfg.ProfileID != profileID {
		return nil, exception.NewNoPropertyAccessForbiddenError("workflow_config")
	}
	if req.Name != "" {
		cfg.Name = req.Name
	}
	cfg.Instruction = req.Instruction
	cfg.LivenessSentence = req.LivenessSentence
	if err := s.repo.Update(ctx, cfg); err != nil {
		s.l.Error("Failed to update workflow config", logger.Field{Key: "error", Value: err})
		return nil, err
	}
	return cfg, nil
}

func (s *workflowConfigService) Delete(ctx context.Context, profileID, id uint64) error {
	cfg, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if cfg.ProfileID != profileID {
		return exception.NewNoPropertyAccessForbiddenError("workflow_config")
	}
	return s.repo.Delete(ctx, id)
}