package v1

import (
	"net/http"

	domainJWT "github.com/Barbod-Biometrics/Backend/gateway/internal/domain/jwt"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/presentation/controller/v1/admin"
	apikey "github.com/Barbod-Biometrics/Backend/gateway/internal/presentation/controller/v1/apikey"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/presentation/controller/v1/profile"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/presentation/controller/v1/user"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/presentation/controller/v1/wallet"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/presentation/middleware"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type Route struct {
	userController         *user.GeneralUserController
	profileController      *profile.ProfileHandler
	apiKeyController       *apikey.ApiKeyHandler
	adminProfileController *admin.AdminProfileHandler
	walletController       *wallet.WalletHandler
	jwtKeyManager          domainJWT.KeyManager
	serviceName            string
}

func NewRouter(
	userController *user.GeneralUserController,
	profileHandler *profile.ProfileHandler,
	apiKeyHandler *apikey.ApiKeyHandler,
	adminProfileHandler *admin.AdminProfileHandler,
	walletHandler *wallet.WalletHandler,
	jwtKeyManager domainJWT.KeyManager,
	serviceName string,
) *Route {
	return &Route{
		userController:         userController,
		profileController:      profileHandler,
		apiKeyController:       apiKeyHandler,
		adminProfileController: adminProfileHandler,
		walletController:       walletHandler,
		jwtKeyManager:          jwtKeyManager,
		serviceName:            serviceName,
	}
}

func (r *Route) RegisterRoutes() http.Handler {
	router := gin.New()

	router.Use(middleware.NewCorsMiddleware().CORS())
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(middleware.OpenTelemetryMiddleware(r.serviceName))

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	v1 := router.Group("/api/v1")
	{
		auth := v1.Group("/auth")
		{
			auth.POST("/request-otp", r.userController.RequestOTPHandler)
			auth.POST("/verify-otp", r.userController.VerifyOTPHandler)
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
			profiles.GET("/", r.profileController.GetUserProfiles)

			// Wallet routes
			profiles.GET("/:id/wallet/summary", r.walletController.GetWalletSummary)
			profiles.GET("/:id/wallet/transactions", r.walletController.GetTransactions)
			profiles.POST("/:id/wallet/deposit", r.walletController.Deposit)
		}
		api_key := v1.Group("/api-key")
		{
			api_key.POST("/:profile_id/regenerate", r.apiKeyController.RegenerateKey)
		}

		// Admin routes - require JWT + admin role
		adminGroup := v1.Group("/admin")
		adminGroup.Use(middleware.JWTMiddleware(r.jwtKeyManager))
		adminGroup.Use(middleware.AdminMiddleware())
		{
			adminProfiles := adminGroup.Group("/profiles")
			{
				adminProfiles.GET("", r.adminProfileController.ListProfiles)
				adminProfiles.GET("/:id", r.adminProfileController.GetProfileDetail)
				adminProfiles.POST("/:id/approve", r.adminProfileController.ApproveProfile)
				adminProfiles.POST("/:id/reject", r.adminProfileController.RejectProfile)
			}
		}

		user := v1.Group("/user")
		user.Use(middleware.JWTMiddleware(r.jwtKeyManager))
		{
			user.GET("/info", r.userController.GetUserProfileHandler)
			user.POST("/update-info", r.userController.UpdateProfileHandler)
		}
	}

	return router
}
