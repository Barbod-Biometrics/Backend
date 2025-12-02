package middleware

import (
	"net/http"

	"github.com/Barbod-Biometrics/Backend/gateway/internal/application/usecase"
	"github.com/gin-gonic/gin"
)

func APIKeyAuth(apiKeyUsecase usecase.APIKeyUsecase) gin.HandlerFunc {
	return func(c *gin.Context) {
		rawKey := c.GetHeader("X-API-KEY")
		if rawKey == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing API Key"})
			c.Abort()
			return
		}

		profileID, err := apiKeyUsecase.Authenticate(c.Request.Context(), rawKey)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or revoked API Key"})
			c.Abort()
			return
		}

		c.Set("profile_id", profileID)
		c.Next()
	}
}
