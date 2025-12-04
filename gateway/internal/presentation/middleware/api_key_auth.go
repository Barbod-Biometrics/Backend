package middleware

import (
	"net/http"

	"github.com/Barbod-Biometrics/Backend/gateway/internal/application/usecase"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/logger"
	"github.com/gin-gonic/gin"
)

func APIKeyAuth(apiKeyUsecase usecase.APIKeyUsecase, log logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		rawKey := c.GetHeader("X-API-KEY")
		if rawKey == "" {
			log.Warn("request blocked: missing api key",
				logger.Field{Key: "ip", Value: c.ClientIP()},
				logger.Field{Key: "path", Value: c.Request.URL.Path},
			)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing API Key"})
			c.Abort()
			return
		}

		profileID, err := apiKeyUsecase.Authenticate(c.Request.Context(), rawKey)
		if err != nil {
			log.Warn("request blocked: authentication failed",
				logger.Field{Key: "ip", Value: c.ClientIP()},
				logger.Field{Key: "path", Value: c.Request.URL.Path},
				logger.Field{Key: "method", Value: c.Request.Method},
			)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or revoked API Key"})
			c.Abort()
			return
		}

		c.Set("profile_id", profileID)
		c.Next()
	}
}
