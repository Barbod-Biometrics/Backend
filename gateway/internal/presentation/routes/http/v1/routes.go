package v1

import (
	"net/http"

	"github.com/Barbod-Biometrics/Backend/gateway/bootstrap"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/application/usecase"
	domainJWT "github.com/Barbod-Biometrics/Backend/gateway/internal/domain/jwt"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/logger"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/infrastructure/localization"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/presentation/controller/v1/admin"
	apikey "github.com/Barbod-Biometrics/Backend/gateway/internal/presentation/controller/v1/apikey"
	faceVerificationController "github.com/Barbod-Biometrics/Backend/gateway/internal/presentation/controller/v1/face_verification"
	ocrController "github.com/Barbod-Biometrics/Backend/gateway/internal/presentation/controller/v1/ocr"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/presentation/controller/v1/profile"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/presentation/controller/v1/session"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/presentation/controller/v1/support"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/presentation/controller/v1/transaction"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/presentation/controller/v1/user"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/presentation/controller/v1/wallet"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/presentation/controller/v1/workflow_config"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/presentation/middleware"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type Route struct {
	userController             *user.GeneralUserController
	profileController          *profile.ProfileHandler
	apiKeyController           *apikey.ApiKeyHandler
	adminProfileController     *admin.AdminProfileHandler
	sessionController          *session.SessionHandler
	workflowConfigController   *workflow_config.WorkflowConfigHandler
	faceVerificationController *faceVerificationController.FaceVerificationHandler
	ocrController              *ocrController.OCRHandler
	walletController           *wallet.WalletHandler
	transactionController      *transaction.TransactionHandler
	ticketController           *support.TicketHandler
	jwtKeyManager              domainJWT.KeyManager
	serviceName                string
	constants                  *bootstrap.Constants
	apiKeyUsecase              usecase.APIKeyUsecase
	log                        logger.Logger
}

func NewRouter(
	userController *user.GeneralUserController,
	profileHandler *profile.ProfileHandler,
	apiKeyHandler *apikey.ApiKeyHandler,
	adminProfileHandler *admin.AdminProfileHandler,
	sessionHandler *session.SessionHandler,
	workflowConfigHandler *workflow_config.WorkflowConfigHandler,
	faceVerificationHandler *faceVerificationController.FaceVerificationHandler,
	ocrHandler *ocrController.OCRHandler,
	walletHandler *wallet.WalletHandler,
	transactionHandler *transaction.TransactionHandler,
	ticketHandler *support.TicketHandler,
	apiKeyUsecase usecase.APIKeyUsecase,
	appLogger logger.Logger,
	jwtKeyManager domainJWT.KeyManager,
	serviceName string,
	constants *bootstrap.Constants,
) *Route {
	return &Route{
		userController:             userController,
		profileController:          profileHandler,
		apiKeyController:           apiKeyHandler,
		adminProfileController:     adminProfileHandler,
		sessionController:          sessionHandler,
		workflowConfigController:   workflowConfigHandler,
		walletController:           walletHandler,
		transactionController:      transactionHandler,
		ticketController:           ticketHandler,
		faceVerificationController: faceVerificationHandler,
		ocrController:              ocrHandler,
		jwtKeyManager:              jwtKeyManager,
		serviceName:                serviceName,
		constants:                  constants,
		apiKeyUsecase:              apiKeyUsecase,
		log:                        appLogger,
	}
}

func (r *Route) RegisterRoutes() http.Handler {
	router := gin.New()

	translatorService := localization.GetService()

	router.Use(middleware.NewCorsMiddleware().CORS())
	router.Use(middleware.NewLocalization(translatorService).Localization())
	router.Use(middleware.NewRecovery(r.constants).Recovery)
	router.Use(gin.Logger())
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

		billing := v1.Group("/billing")
		billing.Use(middleware.JWTMiddleware(r.jwtKeyManager))
		{
			// GET /api/v1/billing/summary?profile_id=123
			billing.GET("/summary", r.transactionController.GetUsageSummary)
		}

		api_key := v1.Group("/api-key")
		{
			api_key.POST("/:profile_id/generate", r.apiKeyController.GenerateKey)
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

			// Admin ticket routes
			adminTickets := adminGroup.Group("/tickets")
			{
				adminTickets.GET("", r.ticketController.GetAllTickets)
				adminTickets.GET("/:ticketId", r.ticketController.GetTicketDetail)
				adminTickets.GET("/:ticketId/file", r.ticketController.GetTicketFile)
			}
		}

		user := v1.Group("/user")
		user.Use(middleware.JWTMiddleware(r.jwtKeyManager))
		{
			user.GET("/info", r.userController.GetUserProfileHandler)
			user.POST("/update-info", r.userController.UpdateProfileHandler)
		}

		// Support ticket routes - require JWT
		supportGroup := v1.Group("/support")
		supportGroup.Use(middleware.JWTMiddleware(r.jwtKeyManager))
		{
			supportGroup.GET("/tickets", r.ticketController.GetUserTickets)
			supportGroup.POST("/tickets", r.ticketController.CreateTicket)
			supportGroup.GET("/tickets/upload-url", r.ticketController.GetUploadURL)
		}

		// Ticket message routes - require JWT (shared between user and admin)
		tickets := v1.Group("/tickets")
		tickets.Use(middleware.JWTMiddleware(r.jwtKeyManager))
		{
			tickets.GET("/:ticketId/messages", r.ticketController.GetTicketMessages)
			tickets.POST("/:ticketId/messages", r.ticketController.SendMessage)
			tickets.POST("/:ticketId/close", r.ticketController.CloseTicket)
		}

		faceVerification := v1.Group("/face-verification")
		faceVerification.Use(middleware.APIKeyAuth(r.apiKeyUsecase, r.log))
		{
			faceVerification.POST("/verify", r.faceVerificationController.VerifyFace)
			faceVerification.POST("/crop", r.faceVerificationController.CropImage)
			faceVerification.GET("/health", r.faceVerificationController.HealthCheck)

		}

		modelReports := v1.Group("/report")
		modelReports.Use((middleware.JWTMiddleware(r.jwtKeyManager)))
		{
			modelReports.GET("/face-verification", r.faceVerificationController.GetReport)
		}

		workflow := v1.Group("/workflow")
		workflow.Use(middleware.APIKeyAuth(r.apiKeyUsecase, r.log))
		{
			workflow.POST("/", r.sessionController.Start)
			workflow.POST("/:id/upload", r.sessionController.Upload)
			workflow.GET("/:id", r.sessionController.Get)
			workflow.POST("/:id/cancel", r.sessionController.Cancel)
			workflow.POST("/:id/accept-ocr", r.sessionController.AcceptOCR)
		}

		workflowConfig := v1.Group("/workflow/config")
		workflowConfig.Use(middleware.JWTMiddleware(r.jwtKeyManager))
		{
			workflowConfig.POST("/", r.workflowConfigController.SaveConfig)
			workflowConfig.GET("/", r.workflowConfigController.ListConfigs)
			workflowConfig.GET("/:id", r.workflowConfigController.GetConfig)
			workflowConfig.PUT("/:id", r.workflowConfigController.UpdateConfig)
			workflowConfig.DELETE("/:id", r.workflowConfigController.DeleteConfig)
		}

		ocr := v1.Group("/ocr")
		ocr.Use(middleware.APIKeyAuth(r.apiKeyUsecase, r.log))
		{
			ocr.POST("/extract", r.ocrController.ExtractText)
			ocr.GET("/health", r.ocrController.HealthCheck)
		}

		demo := v1.Group("/demo")
		{
			demo.POST("/face-verification/verify", r.faceVerificationController.DemoVerify)
			demo.POST("/ocr/extract", r.ocrController.DemoExtract)
		}

	}

	return router
}
