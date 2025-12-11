package wire

import (
	"github.com/Barbod-Biometrics/Backend/gateway/bootstrap"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/application/service"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/application/usecase"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/enum"
	domainJWT "github.com/Barbod-Biometrics/Backend/gateway/internal/domain/jwt"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/logger"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/infrastructure/communication/sms"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/infrastructure/database/redis"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/infrastructure/jwt"
	Logger "github.com/Barbod-Biometrics/Backend/gateway/internal/infrastructure/logger"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/presentation/controller/v1/admin"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/presentation/controller/v1/apikey"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/presentation/controller/v1/profile"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/presentation/controller/v1/user"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/presentation/controller/v1/wallet"
	v1 "github.com/Barbod-Biometrics/Backend/gateway/internal/presentation/routes/http/v1"
	"github.com/Barbod-Biometrics/Backend/gateway/pkg/database"
	"github.com/Barbod-Biometrics/Backend/gateway/pkg/storage"
	"gorm.io/gorm"
)

// --- Custom Types for Distinct Dependencies ---
type AppLogger logger.Logger
type ControllerLogger logger.Logger

// --- Configuration & Logger Providers ---

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

func ProvideAppLogger(loggerCfg *Logger.LoggerConfig) (AppLogger, error) {
	log, err := Logger.NewModuleLogger("app", loggerCfg)
	return AppLogger(log), err
}

func ProvideAPIControllerLogger(loggerCfg *Logger.LoggerConfig) (ControllerLogger, error) {
	log, err := Logger.NewModuleLogger("apikey_controller", loggerCfg)
	return ControllerLogger(log), err
}

// ProvideGenericLogger converts AppLogger back to the interface for Services that need it
func ProvideGenericLogger(l AppLogger) logger.Logger {
	return logger.Logger(l)
}

// --- Infrastructure Providers ---

func ProvidePostgresDatabase(cfg *bootstrap.Config) *gorm.DB {
	return database.NewPostgresDatabase(cfg.Env)
}

func ProvideRedisClient(cfg *bootstrap.Config) (*redis.RedisClient, error) {
	return redis.NewRedisClient(
		cfg.Env.PrimaryRedis.Address,
		cfg.Env.PrimaryRedis.Port,
		cfg.Env.PrimaryRedis.Password,
		cfg.Env.PrimaryRedis.DB,
		cfg.Env.PrimaryRedis.PoolSize,
	)
}

func ProvideMinioClient(cfg *bootstrap.Config) (*storage.MinioClient, error) {
	return storage.NewMinioClient(*cfg.Env)
}

func ProvideJWTKeyManager() *jwt.JWTKeyManager {
	return jwt.NewJWTKeyManager()
}

func ProvideSMSService(cfg *bootstrap.Config) *sms.SMSService {
	return sms.NewSMSService(cfg.Env.SMSGateway.APIKey, cfg.Env.OTP.BackdoorCode)
}

// --- Controller Providers ---

func ProvideAuthController(authUsecase usecase.AuthUsecase) *user.UserHandler {
	// Note: We cast usecase to *service.AuthService because your controller expects the concrete struct currently.
	// Ideally controller should accept interface, but this works for now.
	return user.NewAuthController(authUsecase.(*service.AuthService))
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

// --- Router & App Providers ---

func ProvideRouter(
	authController *user.UserHandler,
	profileHandler *profile.ProfileHandler,
	apiKeyController *apikey.ApiKeyHandler,
	adminProfileHandler *admin.AdminProfileHandler,
	walletHandler *wallet.WalletHandler,
	jwtKeyManager domainJWT.KeyManager,
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
