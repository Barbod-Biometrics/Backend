package workflow_config

import (
	"net/http"
	"strconv"

	dto "github.com/Barbod-Biometrics/Backend/gateway/internal/application/dto/workflow_config"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/application/service"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/application/usecase"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/exception"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/logger"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/repository"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/presentation/middleware"
	"github.com/gin-gonic/gin"
)

type WorkflowConfigHandler struct {
	uc     usecase.WorkflowConfigUsecase
	logger logger.Logger
}

func NewWorkflowConfigHandler(repo repository.WorkflowConfigRepository, l logger.Logger) *WorkflowConfigHandler {
	svc := service.NewWorkflowConfigService(repo, l)
	return &WorkflowConfigHandler{uc: svc, logger: l}
}

// SaveConfig godoc
// @Summary Create a new workflow config
// @Tags Workflow-Config
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param config body SaveConfigRequest true "Workflow config"
// @Success 201 {object} ConfigResponse
// @Router /workflow/config [post]
func (h *WorkflowConfigHandler) SaveConfig(c *gin.Context) {
	pid, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		panic(exception.NewUnauthorizedError("", err))
	}
	var req dto.SaveConfigRequest
	if err := c.BindJSON(&req); err != nil {
		panic(exception.ErrInvalidRequest)
	}

	cfg, err := h.uc.Save(c.Request.Context(), pid, req)
	if err != nil {
		panic(exception.ErrFailedToSave)
	}

	c.JSON(http.StatusCreated, dto.ConfigResponse{
		ID:               cfg.ID,
		ProfileID:        cfg.ProfileID,
		Name:             cfg.Name,
		Instruction:      cfg.Instruction,
		LivenessSentence: cfg.LivenessSentence,
		CreatedAt:        cfg.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:        cfg.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}

// ListConfigs godoc
// @Summary List all workflow configs for profile
// @Tags Workflow-Config
// @Security BearerAuth
// @Produce json
// @Success 200 {array} ConfigResponse
// @Router /workflow/config [get]
func (h *WorkflowConfigHandler) ListConfigs(c *gin.Context) {
	pid, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		panic(exception.NewUnauthorizedError("", err))
	}
	cfgs, err := h.uc.ListByProfile(c.Request.Context(), pid)
	if err != nil {
		panic(exception.ErrFailedToList)
	}

	var responses []dto.ConfigResponse
	for _, cfg := range cfgs {
		responses = append(responses, dto.ConfigResponse{
			ID:               cfg.ID,
			ProfileID:        cfg.ProfileID,
			Name:             cfg.Name,
			Instruction:      cfg.Instruction,
			LivenessSentence: cfg.LivenessSentence,
			CreatedAt:        cfg.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:        cfg.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	c.JSON(http.StatusOK, responses)
}

// GetConfig godoc
// @Summary Get a specific workflow config
// @Tags Workflow-Config
// @Security BearerAuth
// @Produce json
// @Param id path int true "Config ID"
// @Success 200 {object} ConfigResponse
// @Router /workflow/config/{id} [get]
func (h *WorkflowConfigHandler) GetConfig(c *gin.Context) {
	pid, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		panic(exception.NewUnauthorizedError("", err))
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		panic(exception.ErrInvalidID)
	}

	cfg, err := h.uc.GetByID(c.Request.Context(), id)
	if err != nil {
		panic(exception.NewNotFoundError("workflow_config"))
	}

	// Ensure the config belongs to the requesting user
	if cfg.ProfileID != pid {
		panic(exception.NewNoPropertyAccessForbiddenError("workflow_config"))
	}

	c.JSON(http.StatusOK, dto.ConfigResponse{
		ID:               cfg.ID,
		ProfileID:        cfg.ProfileID,
		Name:             cfg.Name,
		Instruction:      cfg.Instruction,
		LivenessSentence: cfg.LivenessSentence,
		CreatedAt:        cfg.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:        cfg.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}

// UpdateConfig godoc
// @Summary Update a workflow config
// @Tags Workflow-Config
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "Config ID"
// @Param config body UpdateConfigRequest true "Workflow config"
// @Success 200 {object} ConfigResponse
// @Router /workflow/config/{id} [put]
func (h *WorkflowConfigHandler) UpdateConfig(c *gin.Context) {
	pid, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		panic(exception.NewUnauthorizedError("", err))
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		panic(exception.ErrInvalidID)
	}

	var req dto.UpdateConfigRequest
	if err := c.BindJSON(&req); err != nil {
		panic(exception.ErrInvalidRequest)
	}

	cfg, err := h.uc.Update(c.Request.Context(), pid, id, req)
	if err != nil {
		if nf, ok := err.(exception.NotFoundError); ok {
			panic(nf)
		}
		if forbErr, ok := err.(exception.ForbiddenError); ok {
			panic(forbErr)
		}
		if appErr, ok := err.(*exception.AppError); ok {
			panic(appErr)
		}
		panic(exception.ErrFailedToUpdate)
	}

	c.JSON(http.StatusOK, dto.ConfigResponse{
		ID:               cfg.ID,
		ProfileID:        cfg.ProfileID,
		Name:             cfg.Name,
		Instruction:      cfg.Instruction,
		LivenessSentence: cfg.LivenessSentence,
		CreatedAt:        cfg.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:        cfg.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}

// DeleteConfig godoc
// @Summary Delete a workflow config
// @Tags Workflow-Config
// @Security BearerAuth
// @Produce json
// @Param id path int true "Config ID"
// @Success 200 {object} map[string]interface{}
// @Router /workflow/config/{id} [delete]
func (h *WorkflowConfigHandler) DeleteConfig(c *gin.Context) {
	pid, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		panic(exception.NewUnauthorizedError("", err))
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		panic(exception.ErrInvalidID)
	}

	if err := h.uc.Delete(c.Request.Context(), pid, id); err != nil {
		if nf, ok := err.(exception.NotFoundError); ok {
			panic(nf)
		}
		if forbErr, ok := err.(exception.ForbiddenError); ok {
			panic(forbErr)
		}
		if appErr, ok := err.(*exception.AppError); ok {
			panic(appErr)
		}
		panic(exception.ErrFailedToDelete)
	}

	c.JSON(http.StatusOK, gin.H{"message": "config deleted successfully"})
}
