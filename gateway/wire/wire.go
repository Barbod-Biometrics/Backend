package wire

import (
	"github.com/Barbod-Biometrics/Backend/gateway/bootstrap"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/application/service"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/application/usecase"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/enum"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/logger"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/repository"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/infrastructure/communication/sms"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/infrastructure/database/redis"
	infraJWT "github.com/Barbod-Biometrics/Backend/gateway/internal/infrastructure/jwt"
	Logger "github.com/Barbod-Biometrics/Backend/gateway/internal/infrastructure/logger"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/infrastructure/repository/postgres"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/presentation/controller/v1/admin"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/presentation/controller/v1/apikey"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/presentation/controller/v1/profile"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/presentation/controller/v1/user"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/presentation/controller/v1/wallet"
	v1 "github.com/Barbod-Biometrics/Backend/gateway/internal/presentation/routes/http/v1"
	"github.com/Barbod-Biometrics/Backend/gateway/pkg/database"
	"github.com/Barbod-Biometrics/Backend/gateway/pkg/storage"
	"github.com/google/wire"
	"gorm.io/gorm"
)

// provide configuration dependencies
var ConfigSet = wire.NewSet(
	ProvideConfig,
	ProvideLoggerConfig,
)

// provides logger dependencies
var LoggerSet = wire.NewSet(
	ProvideAppLogger,
	ProvideAPIControllerLogger,
)

// provides infrastructure layer dependencies
var InfrastructureSet = wire.NewSet(
	ProvidePostgresDatabase,
	ProvideRedisClient,
	ProvideMinioClient,
	ProvideJWTKeyManager,
	ProvideSMSService,
)

// Repo set
var RepositorySet = wire.NewSet(
	postgres.NewUserRepository,
	postgres.NewProfileRepository,
	postgres.NewTransactionRepository,
	redis.NewCacheRepository,
	postgres.NewGormUnitOfWork,

	wire.Bind(new(repository.UserRepository), new(*postgres.UserRepository)),
	wire.Bind(new(repository.ProfileRepository), new(*postgres.ProfileRepository)),
	wire.Bind(new(repository.TransactionRepository), new(*postgres.TransactionRepository)),
	wire.Bind(new(repository.APIKeyRepository), new(*postgres.ApiKeyRepository)),
	wire.Bind(new(repository.CacheRepository), new(*redis.CacheRepository)),
	wire.Bind(new(repository.UnitOfWork), new(*postgres.GormUnitOfWork)),
)

// service set provides service/usecase
var ServiceSet = wire.NewSet(
	service.NewJWTService,
	service.NewOTPService,
	service.NewAuthService,
	service.NewProfileService,
	service.NewAPIKeyService,
	service.NewWalletService,
	service.NewAdminProfileService,

	wire.Bind(new(usecase.TokenUsecase), new(*service.JWTService)),
	wire.Bind(new(usecase.OTPUsecase), new(*service.OTPService)),
	wire.Bind(new(usecase.AuthUsecase), new(*service.AuthService)),
	wire.Bind(new(usecase.ProfileUsecase), new(*service.ProfileService)),
	wire.Bind(new(usecase.APIKeyUsecase), new(*service.APIKeyService)),
	wire.Bind(new(usecase.WalletUsecase), new(*service.WalletService)),
	wire.Bind(new(usecase.AdminProfileUsecase), new(*service.AdminProfileService)),
)

// providing presentation layer
var ControllerSet = wire.NewSet(
	ProvideAuthController,
	ProvideProfileHandler,
	ProvideAPIKeyController,
	ProvideAdminProfileHandler,
	ProvideWalletHandler,
)

// provides the router
var RouterSet = wire.NewSet(
	ProvideRouter,
)

func ProvideConfig() *bootstrap.Config {
	return bootstrap.Run()
}

func ProvideLoggerConfig(cfg *bootstrap.Config) *Logger.LoggerConfig {
	return &Logger.LoggerConfig{
		LogLevel:      string(enum.LogLevelInfo),
		ConsoleOutput: cfg.Env.Logger.ConsoleOutput,
		LogFile:       cfg.Env.Logger.LogFile,
	}
}

type AppLogger logger.Logger

type ControllerLogger logger.Logger

func ProvideAppLogger(loggerCfg *Logger.LoggerConfig) (AppLogger, error) {
	log, err := Logger.NewModuleLogger("app", loggerCfg)
	return AppLogger(log), err
}

func ProvideAPIControllerLogger(loggerCfg *Logger.LoggerConfig) (ControllerLogger, error) {
	log, err := Logger.NewModuleLogger("apikey_controller", loggerCfg)
	return ControllerLogger(log), err
}

func ProvidePostgresDatabase(cfg *bootstrap.Config, appLogger AppLogger) *gorm.DB {
	return database.NewPostgresDatabase(cfg.Env)
}

func ProvideRedisClient(cfg *bootstrap.Config, appLogger AppLogger) (*redis.RedisClient, error) {
	return redis.NewRedisClient(
		cfg.Env.PrimaryRedis.Address,
		cfg.Env.PrimaryRedis.Port,
		cfg.Env.PrimaryRedis.Password,
		cfg.Env.PrimaryRedis.DB,
		cfg.Env.PrimaryRedis.PoolSize,
	)
}

func ProvideMinioClient(cfg *bootstrap.Config, appLogger AppLogger) (*storage.MinioClient, error) {
	return storage.NewMinioClient(*cfg.Env)
}

// check this before moving further
func ProvideJWTKeyManager() *infraJWT.JWTKeyManager {
	return &infraJWT.JWTKeyManager{}
}
func ProvideSMSService(cfg *bootstrap.Config) *sms.SMSService {
	return sms.NewSMSService(cfg.Env.SMSGateway.APIKey, cfg.Env.OTP.BackdoorCode)
}

func ProvideAuthController(authUsecase *service.AuthService) *user.UserHandler {
	return user.NewAuthController(authUsecase)
}

func ProvideProfileHandler(profileUsecase usecase.ProfileUsecase) *profile.ProfileHandler {
	return profile.NewProfileHandler(profileUsecase)
}

func ProvideAPIKeyController(
	apikeKeyUsecase usecase.APIKeyUsecase,
	controllerLogger ControllerLogger,
) *apikey.ApiKeyHandler {
	return apikey.NewApiKeyHandler(apikeKeyUsecase, logger.Logger(controllerLogger))
}

func ProvideAdminProfileHandler(adminProfileUsecase usecase.AdminProfileUsecase) *admin.AdminProfileHandler {
	return admin.NewAdminProfileHandler(adminProfileUsecase)
}

func ProvideWalletHandler(walletUsecase usecase.WalletUsecase) *wallet.WalletHandler {
	return wallet.NewWalletHandler(walletUsecase)
}

func ProvideRouter(
	authController *user.UserHandler,
	profileHandler *profile.ProfileHandler,
	apiKeyController *apikey.ApiKeyHandler,
	adminProfileHandler *admin.AdminProfileHandler,
	walletHandler *wallet.WalletHandler,
	jwtKeyManager *infraJWT.JWTKeyManager,
	cfg *bootstrap.Config,
) *v1.Route {
	return v1.NewRouter(
		authController,
		profileHandler,
		apiKeyController,
		adminProfileHandler,
		walletHandler,
		jwtKeyManager,
		cfg.Env.Telemetry.ServiceName,
	)
}

type Application struct {
	Config      *bootstrap.Config
	Logger      AppLogger
	DB          *gorm.DB
	RedisClient *redis.RedisClient
	Route       *v1.Route
	MinioClient *storage.MinioClient
}

func NewApplication(
	cfg *bootstrap.Config,
	appLogger AppLogger,
	db *gorm.DB,
	redisClient *redis.RedisClient,
	route *v1.Route,
	minioClient *storage.MinioClient,
) *Application {
	return &Application{
		Config:      cfg,
		Logger:      appLogger,
		DB:          db,
		RedisClient: redisClient,
		Route:       route,
		MinioClient: minioClient,
	}
}
func (app *Application) Close() error {
	if app.Logger != nil {
		logger.Logger(app.Logger).Close()
	}
	if app.RedisClient != nil {
		app.RedisClient.Close()
	}
	return nil
}

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
