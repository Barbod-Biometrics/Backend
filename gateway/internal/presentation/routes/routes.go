package routes

import (
	"github.com/Barbod-Biometrics/Backend/gateway/internal/presentation/handler"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "github.com/Barbod-Biometrics/Backend/gateway/docs"
)

// Setup registers all routes for the application
func Setup(router *gin.Engine, profileHandler *handler.ProfileHandler) {

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API Version 1 Group
	v1 := router.Group("/api/v1")
	{
		// Profile Routes
		profiles := v1.Group("/profiles")
		{
			// Step 1: Create Draft
			profiles.POST("/", profileHandler.CreateDraft)

			// Step 2/3: Update Draft (Partial Save)
			profiles.PATCH("/:id", profileHandler.UpdateDraft)

			// Step 4: Get Profile (To view/edit draft)
			profiles.GET("/:id", profileHandler.GetProfile)

			// Step 5: Upload Document URL
			profiles.POST("/:id/documents", profileHandler.SaveDocument)

			// Step 6: Final Submit
			profiles.POST("/:id/submit", profileHandler.Submit)

			profiles.POST("/upload-url", profileHandler.GetUploadUrl)
		}
	}
}
