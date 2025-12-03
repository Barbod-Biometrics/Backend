package v1

import (
	"net/http"

	apikey "github.com/Barbod-Biometrics/Backend/gateway/internal/presentation/controller/v1/apiKey"
	_ "github.com/Barbod-Biometrics/Backend/gateway/docs"
	domainJWT "github.com/Barbod-Biometrics/Backend/gateway/internal/domain/jwt"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/presentation/controller/v1/profile"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/presentation/controller/v1/user"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/presentation/middleware"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type Route struct {
	authController    *user.GeneralUserController
	profileController *profile.ProfileHandler
	apiKeyController  *apikey.ApiKeyHandler
	jwtKeyManager     domainJWT.KeyManager
	serviceName       string
}

func NewRouter(
	authController *user.GeneralUserController,
	profileHandler *profile.ProfileHandler,
	apiKeyHandler *apikey.ApiKeyHandler,
	jwtKeyManager domainJWT.KeyManager,
	serviceName string,
) *Route {
	return &Route{
		authController:    authController,
		profileController: profileHandler,
		apiKeyController:  apiKeyHandler,
		jwtKeyManager:     jwtKeyManager,
		serviceName:       serviceName,
	}
}

func (r *Route) RegisterRoutes() http.Handler {
	router := gin.New()

	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(middleware.OpenTelemetryMiddleware(r.serviceName))

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	v1 := router.Group("/api/v1")
	{
		auth := v1.Group("/auth")
		{
			auth.POST("/request-otp", r.authController.RequestOTPHandler)
			auth.POST("/verify-otp", r.authController.VerifyOTPHandler)
		}
		profiles := v1.Group("/profiles")
		profiles.Use(middleware.JWTMiddleware(r.jwtKeyManager))
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
			api_key.POST("/:profile_id/regenerate", r.apiKeyController.RegenerateKey)
		}
	}

	return router
}
