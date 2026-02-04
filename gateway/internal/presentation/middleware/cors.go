package middleware

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type CORSMiddleware struct{}

func NewCorsMiddleware() *CORSMiddleware {
	return &CORSMiddleware{}
}

func (cm *CORSMiddleware) CORS() gin.HandlerFunc {
	corsConfig := cors.Config{
		AllowOrigins: []string{
			"*",
		},
		AllowMethods:  []string{"POST", "GET", "OPTIONS", "PUT", "DELETE" , "PATCH"},
		AllowHeaders:  []string{"Origin", "Content-Type", "Authorization", "X-API-Key", "ngrok-skip-browser-warning"},
		ExposeHeaders: []string{"Content-Length"},
		MaxAge: 12 * time.Hour,
	}

	return cors.New(corsConfig)
}
