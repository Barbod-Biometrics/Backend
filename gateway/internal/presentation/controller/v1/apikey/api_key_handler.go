package apikey

import (
	"net/http"
	"strconv"

	"github.com/Barbod-Biometrics/Backend/gateway/internal/application/dto/business"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/application/usecase"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/logger"
	"github.com/gin-gonic/gin"
)

type ApiKeyHandler struct {
	apiKeyUsecase usecase.APIKeyUsecase
	logger        logger.Logger
}

func NewApiKeyHandler(apiKeyUsecase usecase.APIKeyUsecase, logger logger.Logger) *ApiKeyHandler {
	return &ApiKeyHandler{
		apiKeyUsecase: apiKeyUsecase,
		logger:        logger,
	}
}

// Regenerate handles replacing a lost API Key
// @Summary Regenerate API Key
// @Security BearerAuth
// @Description Revokes the existing active key and generate a new one.
// @Tags API-Keys
// @Accept json
// @Produce json
// @Param profile_id path int true "Profile ID"
// @Success 200 {object} business.NewAPIKeyResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api-key/{profile_id}/regenerate [POST]
func (c *ApiKeyHandler) RegenerateKey(ctx *gin.Context) {

	c.logger.Info("received regenerate api key request", logger.Field{Key: "ip", Value: ctx.ClientIP()})

	profileIDStr := ctx.Param("profile_id")
	ProfileID, err := strconv.ParseUint(profileIDStr, 10, 64)
	if err != nil {
		c.logger.Warn("regenerate request failed: invalid profile_id", logger.Field{Key: "input_id", Value: profileIDStr})
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid profile ID"})
		return
	}

	req := business.ReGenerateAPIKeyResquest{
		ProfileID: ProfileID,
	}

	rawKey, err := c.apiKeyUsecase.RegenerateKey(ctx.Request.Context(), req.ProfileID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp := business.NewAPIKeyResponse{
		ProfileID: req.ProfileID,
		APIKey:    rawKey,
	}

	c.logger.Info("api key regenerated successfully", logger.Field{Key: "profile_id", Value: req.ProfileID})

	ctx.JSON(http.StatusOK, resp)
}
