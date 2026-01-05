package apikey

import (
	"errors"
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

// GenerateKey handles API Key generation
// @Summary Generate API Key
// @Security BearerAuth
// @Description Generate a new API key for the given profile ID.
// @Tags API-Keys
// @Accept json
// @Produce json
// @Param profile_id path int true "Profile ID"
// @Success 200 {object} business.NewAPIKeyResponse
// @Failure 400 {object} map[string]string
// @Failure 409 {object} map[string]string "Key Already Exists (Use Regenerate)"
// @Failure 500 {object} map[string]string
// @Router /api-key/{profile_id}/generate [POST]
func (h *ApiKeyHandler) GenerateKey(c *gin.Context) {
	h.logger.Info("received generate api key request", logger.Field{Key: "ip", Value: c.ClientIP()})

	profileIDStr := c.Param("profile_id")
	ProfileID, err := strconv.ParseUint(profileIDStr, 10, 64)
	if err != nil {
		h.logger.Warn("generate request failed: invalid profile_id", logger.Field{Key: "input_id", Value: profileIDStr})
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid profile ID"})
		return
	}

	req := business.GenerateAPIKeyRequest{
		ProfileID: ProfileID,
	}

	rawKey, err := h.apiKeyUsecase.GenerateKey(c.Request.Context(), req.ProfileID)
	if err != nil {
		if errors.Is(err, errors.New("active key exists for this profile id")) {
			h.logger.Warn("generate request blocked: active key exists", logger.Field{Key: "profile_id", Value: req.ProfileID})
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Active API key already exists for this profile ID",
				"message": "Use /regenerate to replace your key.",
			})
			return
		}
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}

	resp := business.NewAPIKeyResponse{
		ProfileID: req.ProfileID,
		APIKey:    rawKey,
	}

	h.logger.Info("api key generated successfully", logger.Field{Key: "profile_id", Value: req.ProfileID})

	c.JSON(http.StatusOK, resp)
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
func (h *ApiKeyHandler) RegenerateKey(c *gin.Context) {

	h.logger.Info("received regenerate api key request", logger.Field{Key: "ip", Value: c.ClientIP()})

	profileIDStr := c.Param("profile_id")
	ProfileID, err := strconv.ParseUint(profileIDStr, 10, 64)
	if err != nil {
		h.logger.Warn("regenerate request failed: invalid profile_id", logger.Field{Key: "input_id", Value: profileIDStr})
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid profile ID"})
		return
	}

	req := business.GenerateAPIKeyRequest{
		ProfileID: ProfileID,
	}

	rawKey, err := h.apiKeyUsecase.RegenerateKey(c.Request.Context(), req.ProfileID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp := business.NewAPIKeyResponse{
		ProfileID: req.ProfileID,
		APIKey:    rawKey,
	}

	h.logger.Info("api key regenerated successfully", logger.Field{Key: "profile_id", Value: req.ProfileID})

	c.JSON(http.StatusOK, resp)
}
