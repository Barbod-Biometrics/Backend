package wire

import (
	"context"

	"github.com/Barbod-Biometrics/Backend/gateway/bootstrap"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/application/usecase"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/enum"
	domainJWT "github.com/Barbod-Biometrics/Backend/gateway/internal/domain/jwt"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/logger"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/repository"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/infrastructure/communication/sms"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/infrastructure/database/redis"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/infrastructure/jwt"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/infrastructure/localization"
	Logger "github.com/Barbod-Biometrics/Backend/gateway/internal/infrastructure/logger"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/infrastructure/telemetry"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/presentation/controller/v1/admin"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/presentation/controller/v1/apikey"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/presentation/controller/v1/profile"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/presentation/controller/v1/transaction"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/presentation/controller/v1/user"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/presentation/controller/v1/wallet"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/presentation/middleware"
	v1 "github.com/Barbod-Biometrics/Backend/gateway/internal/presentation/routes/http/v1"
	"github.com/Barbod-Biometrics/Backend/gateway/pkg/database"
	"github.com/Barbod-Biometrics/Backend/gateway/pkg/storage"
	"gorm.io/gorm"
)

type AppLogger logger.Logger

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

func ProvideGenericLogger(l AppLogger) logger.Logger {
	return logger.Logger(l)
}

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

func ProvideTelemetry(cfg *bootstrap.Config) (*telemetry.Telemetry, error) {
	ctx := context.Background()
	return telemetry.InitTelemetry(ctx, &cfg.Env.Telemetry)
}

func ProvideJWTKeyManager() *jwt.JWTKeyManager {
	return jwt.NewJWTKeyManager()
}

func ProvideSMSService(cfg *bootstrap.Config) *sms.SMSService {
	return sms.NewSMSService(cfg.Env.SMSGateway.APIKey, cfg.Env.OTP.BackdoorCode)
}

func ProvideTranslator() *localization.TranslationService {
    return localization.GetService()
}

func ProvideLocalizationTranslator(trans *localization.TranslationService) *middleware.LocalizationMiddleware {
    return middleware.NewLocalization(trans)
}

func ProvideRecovery(constants *bootstrap.Constants) *middleware.RecoveryMiddleware {
	return middleware.NewRecovery(constants)
}

func ProvideUserController(authUsecase usecase.AuthUsecase, userUsecase usecase.UserUsecase) *user.GeneralUserController {
	return user.NewUserController(authUsecase, userUsecase)
}

func ProvideProfileHandler(profileUsecase usecase.ProfileUsecase) *profile.ProfileHandler {
	return profile.NewProfileHandler(profileUsecase)
}

func ProvideAPIKeyController(
	apikeKeyUsecase usecase.APIKeyUsecase,
	l logger.Logger,
) *apikey.ApiKeyHandler {
	return apikey.NewApiKeyHandler(apikeKeyUsecase, logger.Logger(l))
}

func ProvideAdminProfileHandler(adminProfileUsecase usecase.AdminProfileUsecase) *admin.AdminProfileHandler {
	return admin.NewAdminProfileHandler(adminProfileUsecase)
}

func ProvideWalletHandler(walletUsecase usecase.WalletUsecase, l logger.Logger) *wallet.WalletHandler {
	return wallet.NewWalletHandler(walletUsecase, logger.Logger(l))
}

func ProvideTransactionHandler(transactionUsecase usecase.TransactionUsecase, profileRepository repository.ProfileRepository, l logger.Logger) *transaction.TransactionHandler {
	return transaction.NewTransactionHandler(transactionUsecase, profileRepository, logger.Logger(l))
}

func ProvideRouter(
	authController *user.GeneralUserController,
	profileHandler *profile.ProfileHandler,
	apiKeyController *apikey.ApiKeyHandler,
	adminProfileHandler *admin.AdminProfileHandler,
	walletHandler *wallet.WalletHandler,
	transactionHandler *transaction.TransactionHandler,
	jwtKeyManager domainJWT.KeyManager,
	cfg *bootstrap.Config,
) *v1.Route {
	return v1.NewRouter(
		authController,
		profileHandler,
		apiKeyController,
		adminProfileHandler,
		walletHandler,
		transactionHandler,
		jwtKeyManager,
		cfg.Env.Telemetry.ServiceName,
		cfg.Constants,
	)
}

type Application struct {
	Config      *bootstrap.Config
	Logger      AppLogger
	DB          *gorm.DB
	RedisClient *redis.RedisClient
	Route       *v1.Route
	MinioClient *storage.MinioClient
	Telemetry   *telemetry.Telemetry
}

func NewApplication(
	cfg *bootstrap.Config,
	appLogger AppLogger,
	db *gorm.DB,
	redisClient *redis.RedisClient,
	route *v1.Route,
	minioClient *storage.MinioClient,
	telemetryClient *telemetry.Telemetry,
) *Application {
	return &Application{
		Config:      cfg,
		Logger:      appLogger,
		DB:          db,
		RedisClient: redisClient,
		Route:       route,
		MinioClient: minioClient,
		Telemetry:   telemetryClient,
	}
}

func (app *Application) Close() error {
	if app.Logger != nil {
		logger.Logger(app.Logger).Close()
	}
	if app.RedisClient != nil {
		app.RedisClient.Close()
	}
	if app.Telemetry != nil {
		ctx := context.Background()
		if err := app.Telemetry.Shutdown(ctx); err != nil {
			return err
		}
	}
	return nil
}
