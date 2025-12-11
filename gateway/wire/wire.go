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
	ProvideAPIControllerLogger,
	ProvideGenericLogger,
)

// Infrastructure Set
var InfrastructureSet = wire.NewSet(
	ProvidePostgresDatabase,
	ProvideRedisClient,
	ProvideMinioClient,
	ProvideJWTKeyManager,
	ProvideSMSService,

	// Bind Structs to Interfaces for Infrastructure
	wire.Bind(new(domainJWT.KeyManager), new(*jwt.JWTKeyManager)),
	wire.Bind(new(communication.SMSService), new(*sms.SMSService)),
)

// 4. Repository Set
var RepositorySet = wire.NewSet(
	postgres.NewUserRepository,
	postgres.NewGormUnitOfWork,

	// Bindings for Struct-returning repositories
	postgres.NewProfileRepository,
	wire.Bind(new(repository.ProfileRepository), new(*postgres.ProfileRepository)),

	postgres.NewTransactionRepository,
	wire.Bind(new(repository.TransactionRepository), new(*postgres.TransactionRepository)),

	postgres.NewApiKeyRepository,
	wire.Bind(new(repository.APIKeyRepository), new(*postgres.ApiKeyRepository)),

	redis.NewCacheRepository,
	wire.Bind(new(repository.CacheRepository), new(*redis.CacheRepository)),
)

// 5. Service Set
var ServiceSet = wire.NewSet(
	// These constructors return Interfaces directly -> NO BIND NEEDED
	service.NewProfileService,
	service.NewWalletService,
	service.NewAdminProfileService,

	// These constructors return Structs -> BIND REQUIRED
	service.NewJWTService,
	service.NewOTPService,
	service.NewAuthService,
	service.NewAPIKeyService,

	// Binds for the struct-returning services
	wire.Bind(new(usecase.TokenUsecase), new(*service.JWTService)),
	wire.Bind(new(usecase.OTPUsecase), new(*service.OTPService)),
	wire.Bind(new(usecase.AuthUsecase), new(*service.AuthService)),
	wire.Bind(new(usecase.APIKeyUsecase), new(*service.APIKeyService)),
)

// 6. Controller Set
var ControllerSet = wire.NewSet(
	ProvideAuthController,
	ProvideProfileHandler,
	ProvideAPIKeyController,
	ProvideAdminProfileHandler,
	ProvideWalletHandler,
)

// 7. Router Set
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
		RouterSet,
		NewApplication,
	)
	return &Application{}, nil
}
