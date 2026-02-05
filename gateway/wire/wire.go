//go:build wireinject
// +build wireinject

package wire

import (
	"github.com/Barbod-Biometrics/Backend/gateway/internal/application/service"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/application/usecase"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/communication"
	domainJWT "github.com/Barbod-Biometrics/Backend/gateway/internal/domain/jwt"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/repository"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/infrastructure/communication/sms"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/infrastructure/database/redis"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/infrastructure/jwt"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/infrastructure/repository/postgres"
	"github.com/google/wire"
)

// Config Set
var ConfigSet = wire.NewSet(
	ProvideConfig,
	ProvideLoggerConfig,
)

// Logger Set
var LoggerSet = wire.NewSet(
	ProvideAppLogger,
	ProvideGenericLogger,
)

// Infrastructure Set
var InfrastructureSet = wire.NewSet(
	ProvidePostgresDatabase,
	ProvideRedisClient,
	ProvideMinioClient,
	ProvideTelemetry,
	ProvideJWTKeyManager,
	ProvideSMSService,
	ProvideEmailService,
	ProvideTranslator,
	ProvideLocalizationTranslator,
	ProvideRecovery,
	ProvideFaceVerificationClient,
	ProvideOCRClient,
	ProvideRecaptchaVerifier,

	wire.Bind(new(domainJWT.KeyManager), new(*jwt.JWTKeyManager)),
	wire.Bind(new(communication.SMSService), new(*sms.SMSService)),
)

// Repository Set
var RepositorySet = wire.NewSet(
	postgres.NewUserRepository,
	postgres.NewGormUnitOfWork,

	postgres.NewProfileRepository,
	wire.Bind(new(repository.ProfileRepository), new(*postgres.ProfileRepository)),

	postgres.NewTransactionRepository,
	wire.Bind(new(repository.TransactionRepository), new(*postgres.TransactionRepository)),

	postgres.NewApiKeyRepository,
	wire.Bind(new(repository.APIKeyRepository), new(*postgres.ApiKeyRepository)),

	postgres.NewFaceVerificationRepository,
	wire.Bind(new(repository.FaceVerificationRepository), new(*postgres.FaceVerificationRepository)),

	postgres.NewOCRRepository,
	wire.Bind(new(repository.OCRRepository), new(*postgres.OCRRepository)),

	ProvideContactSalesRepository,

	postgres.NewSessionRepository,

	postgres.NewServiceRepository,

	postgres.NewTicketRepository,
	wire.Bind(new(repository.TicketRepository), new(*postgres.TicketRepository)),

	postgres.NewWorkflowConfigRepository,

	redis.NewCacheRepository,
	wire.Bind(new(repository.CacheRepository), new(*redis.CacheRepository)),
)

// Service Set
var ServiceSet = wire.NewSet(
	service.NewProfileService,
	service.NewWalletService,
	service.NewAdminProfileService,

	service.NewUserService,
	service.NewJWTService,
	service.NewOTPService,
	service.NewAuthService,
	service.NewAPIKeyService,
	service.NewTransactionService,
	service.NewTrialService,
	service.NewFaceVerificationService,
	service.NewOCRService,
	ProvideContactSalesUsecase,
	service.NewTicketService,

	wire.Bind(new(usecase.UserUsecase), new(*service.UserService)),
	wire.Bind(new(usecase.TokenUsecase), new(*service.JWTService)),
	wire.Bind(new(usecase.OTPUsecase), new(*service.OTPService)),
	wire.Bind(new(usecase.AuthUsecase), new(*service.AuthService)),
	wire.Bind(new(usecase.APIKeyUsecase), new(*service.APIKeyService)),
	wire.Bind(new(usecase.TransactionUsecase), new(*service.TransactionService)),
	wire.Bind(new(usecase.FaceVerificationUsecase), new(*service.FaceVerificationService)),
	wire.Bind(new(usecase.OCRUsecase), new(*service.OCRService)),
	wire.Bind(new(usecase.TicketUsecase), new(*service.TicketService)),
)

// Controller Set
var ControllerSet = wire.NewSet(
	ProvideUserController,
	ProvideProfileHandler,
	ProvideAPIKeyController,
	ProvideAdminProfileHandler,
	ProvideWalletHandler,
	ProvideTransactionHandler,
	ProvideFaceVerificationHandler,
	ProvideOCRHandler,
	ProvideContactSalesHandler,
	ProvideTicketHandler,
	ProvideWorkflowConfigHandler,
)

// Session Set
var SessionSet = wire.NewSet(
	ProvideSessionStore,
	ProvideSessionManager,
	ProvideSessionWorkflow,
	ProvideSessionHandler,
)

// Router Set
var RouterSet = wire.NewSet(
	ProvideRouter,
)

// MAIN INJECTOR
func InitializeApplication() (*Application, error) {
	wire.Build(
		ConfigSet,
		LoggerSet,
		InfrastructureSet,
		RepositorySet,
		ServiceSet,
		ControllerSet,
		SessionSet,
		RouterSet,
		NewApplication,
	)
	return &Application{}, nil
}
