package apikey

import (
	"net/http"
	"strconv"

	"github.com/Barbod-Biometrics/Backend/gateway/internal/application/dto/business"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/application/usecase"
	"github.com/gin-gonic/gin"
)

type ApiKeyHandler struct {
	apiKeyUsecase usecase.APIKeyUsecase
}

func NewApiKeyHandler(apiKeyUsecase usecase.APIKeyUsecase) *ApiKeyHandler {
	return &ApiKeyHandler{
		apiKeyUsecase: apiKeyUsecase,
	}
}

func (c *ApiKeyHandler) RegenerateKey(ctx *gin.Context) {
	profileIDStr := ctx.Params("profile_id")
	profileID, err := strconv.ParseUint(profileIDStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid profile ID"})
		return
	}

	req := business.ReGenerateAPIKeyResquest{
		profileID: profileID,
	}

	rawKey, err := c.apiKeyUsecase.RegenerateKey(ctx.Request.Context(), req.profileID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp := business.NewAPIKeyResponse{
		profileID: req.profileID,
		APIKey:    rawKey,
	}

	ctx.JSON(http.StatusOK, resp)
}
