package v1

import (
	"net/http"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	apikey "github.com/Barbod-Biometrics/Backend/gateway/internal/presentation/controller/v1/apiKey"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/presentation/controller/v1/profile"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/presentation/controller/v1/user"
	"github.com/gin-gonic/gin"
)

type Route struct {
	authController    *user.GeneralUserController
	profileController *profile.ProfileHandler
	apiKeyController  *apikey.ApiKeyHandler
}

func NewRouter(
	authController *user.GeneralUserController,
	profileHandler *profile.ProfileHandler,
	apiKeyHandler *apikey.ApiKeyHandler,
) *Route {
	return &Route{
		authController:    authController,
		profileController: profileHandler,
		apiKeyController:  apiKeyHandler,
	}
}

func (r *Route) RegisterRoutes() http.Handler {
	router := gin.New()

	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	v1 := router.Group("/api/v1")
	{
		auth := v1.Group("/auth")
		{
			auth.POST("/request-otp", r.authController.RequestOTPHandler)
			auth.POST("/verify-otp", r.authController.VerifyOTPHandler)
		}
		profiles := v1.Group("/profiles")
		{
			profiles.POST("/", r.profileController.CreateDraft)
			profiles.PATCH("/:id", r.profileController.UpdateDraft)
			profiles.GET("/:id", r.profileController.GetProfile)
			profiles.POST("/:id/documents", r.profileController.SaveDocument)
			profiles.POST("/:id/submit", r.profileController.Submit)
			profiles.POST("/upload-url", r.profileController.GetUploadUrl)
		}
		api_key := v1.Group("/api-key")
		{
			api_key.POST("/regenerate", r.apiKeyController.RegenerateKey)
		}
	}

	return router
}
