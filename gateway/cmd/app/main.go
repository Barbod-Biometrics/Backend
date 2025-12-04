package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/Barbod-Biometrics/Backend/gateway/bootstrap"
	docs "github.com/Barbod-Biometrics/Backend/gateway/docs"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/application/service"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/entity"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/enum"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/logger"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/infrastructure/communication/sms"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/infrastructure/database/redis"
	infraJWT "github.com/Barbod-Biometrics/Backend/gateway/internal/infrastructure/jwt"
	Logger "github.com/Barbod-Biometrics/Backend/gateway/internal/infrastructure/logger"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/infrastructure/repository/postgres"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/infrastructure/telemetry"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/presentation/controller/v1/admin"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/presentation/controller/v1/apikey"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/presentation/controller/v1/profile"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/presentation/controller/v1/user"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/presentation/controller/v1/wallet"
	v1 "github.com/Barbod-Biometrics/Backend/gateway/internal/presentation/routes/http/v1"
	"github.com/Barbod-Biometrics/Backend/gateway/pkg/database"
	"github.com/Barbod-Biometrics/Backend/gateway/pkg/storage"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// @title Barbod Biometrics Gateway API
// @version 1.0
// @description This is the Gateway service API documentation for Barbod Biometrics.
// @BasePath /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer {your JWT token}" to authorize requests (without quotes)
func main() {

	gin.DisableConsoleColor()
	cfg := bootstrap.Run()

	loggerCfg := &Logger.LoggerConfig{
		LogLevel:      string(enum.LogLevelInfo),
		ConsoleOutput: cfg.Env.Logger.ConsoleOutput,
		LogFile:       cfg.Env.Logger.LogFile,
	}

	fmt.Printf("logfile: %s\n", cfg.Env.Logger.LogFile)

	appLogger, err := Logger.NewModuleLogger("app", loggerCfg)
	if err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		return
	}
	defer appLogger.Close()

	ctx := context.Background()
	otelTelemetry, err := telemetry.InitTelemetry(ctx, &cfg.Env.Telemetry)
	setSwaggerHost(cfg.Env.Server.Host, cfg.Env.Server.Port)
	if err != nil {
		appLogger.Error("Failed to initialize OpenTelemetry", logger.Field{Key: "error", Value: err})
	}
	defer otelTelemetry.Shutdown(ctx)

	db := database.NewPostgresDatabase(cfg.Env)

	err = db.AutoMigrate(
		&entity.User{},
		&entity.Profile{},
		&entity.ProfileBusinessDetails{},
		&entity.ProfilePersonDetails{},
		&entity.AuthorizedSignatory{},
		&entity.APIKey{},
		&entity.Transaction{},
	)

	if err != nil {
		appLogger.Fatal("Failed to migrate database", logger.Field{Key: "error", Value: err})
	}

	profileRepo := postgres.NewProfileRepository(db)
	transactionRepo := postgres.NewTransactionRepository(db)
	unitOfWork := postgres.NewGormUnitOfWork(db)
	minioStorage, err := storage.NewMinioClient(*cfg.Env)
	if err != nil {
		appLogger.Fatal("Failed to initialize Minio client", logger.Field{Key: "error", Value: err})
	}
	profileService := service.NewProfileService(profileRepo, minioStorage)
	profileHandler := profile.NewProfileHandler(profileService)
	walletService := service.NewWalletService(profileRepo, transactionRepo, unitOfWork)
	walletHandler := wallet.NewWalletHandler(walletService)

	// --- Redis ---
	redisClient, err := redis.NewRedisClient(
		cfg.Env.PrimaryRedis.Address,
		cfg.Env.PrimaryRedis.Port,
		cfg.Env.PrimaryRedis.Password,
		cfg.Env.PrimaryRedis.DB,
		cfg.Env.PrimaryRedis.PoolSize,
	)
	if err != nil {
		appLogger.Fatal("Failed to connect to Redis", logger.Field{Key: "error", Value: err})
	}
	defer redisClient.Close()

	// 4. Initialize Infrastructure Layer
	userRepo := postgres.NewUserRepository(db)
	cacheRepo := redis.NewCacheRepository(redisClient)
	jwtKeyManager := infraJWT.NewJWTKeyManager()
	smsService := sms.NewSMSService(cfg.Env.SMSGateway.APIKey, cfg.Env.OTP.BackdoorCode)
	apiKeyRepo := postgres.NewApiKeyRepository(db)

	apiControllerLogger, _ := Logger.NewModuleLogger("apikey_controller", loggerCfg)

	// 5. Initialize Application Layer (Services)
	jwtService := service.NewJWTService(cfg, jwtKeyManager)
	otpService := service.NewOTPService(cacheRepo, cfg)
	adminProfileService := service.NewAdminProfileService(profileRepo)
	apiKeyService := service.NewAPIKeyService(apiKeyRepo, appLogger) // <--- Init Service

	// 6. Initialize Usecases
	authUsecase := service.NewAuthUsecase(userRepo, otpService, smsService, jwtService)

	// 7. Initialize Controllers
	authController := user.NewAuthController(authUsecase)
	adminProfileHandler := admin.NewAdminProfileHandler(adminProfileService)
	apiKeyController := apikey.NewApiKeyHandler(apiKeyService, apiControllerLogger)

	// 8. Seed admin user from environment variable
	seedAdminUser(db, cfg.Env.Admin.PhoneNumber, appLogger)

	// 9. Setup Router
	v1Router := v1.NewRouter(
		authController,
		profileHandler,
		apiKeyController,
		adminProfileHandler,
    walletHandler,
		jwtKeyManager,
		cfg.Env.Telemetry.ServiceName,
	)

	// Get the handler (which is a Gin Engine)
	handler := v1Router.RegisterRoutes()

	appLogger.Info("Starting the gateway application",
		logger.Field{Key: "port", Value: cfg.Env.Server.Port},
		logger.Field{Key: "database", Value: fmt.Sprintf("%s@%s:%s", cfg.Env.Postgres.User, cfg.Env.Postgres.Host, cfg.Env.Postgres.Port)},
		logger.Field{Key: "redis", Value: fmt.Sprintf("%s:%s", cfg.Env.PrimaryRedis.Address, cfg.Env.PrimaryRedis.Port)},
		logger.Field{Key: "minio", Value: fmt.Sprintf("%s:%s", cfg.Env.Minio.Host, cfg.Env.Minio.Port)},
	)

	if err := http.ListenAndServe(":"+cfg.Env.Server.Port, handler); err != nil {
		appLogger.Error("Error starting server", logger.Field{Key: "error", Value: err})
	}
}

// seedAdminUser creates or updates the admin user based on ADMIN_PHONE_NUMBER env variable
func seedAdminUser(db *gorm.DB, adminPhone string, appLogger logger.Logger) {
	if adminPhone == "" {
		appLogger.Warn("ADMIN_PHONE_NUMBER not set, skipping admin seeding")
		return
	}

	ctx := context.Background()
	var existingUser entity.User

	err := db.WithContext(ctx).Where("phone_number = ?", adminPhone).First(&existingUser).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Create new admin user
			newAdmin := &entity.User{
				PhoneNumber: adminPhone,
				IsAdmin:     true,
			}
			if err := db.WithContext(ctx).Create(newAdmin).Error; err != nil {
				appLogger.Error("Failed to create admin user", logger.Field{Key: "error", Value: err})
				return
			}
			appLogger.Info("Admin user created", logger.Field{Key: "phone", Value: adminPhone})
			return
		}
		appLogger.Error("Failed to check for existing admin user", logger.Field{Key: "error", Value: err})
		return
	}

	// User exists, ensure they are admin
	if !existingUser.IsAdmin {
		existingUser.IsAdmin = true
		if err := db.WithContext(ctx).Save(&existingUser).Error; err != nil {
			appLogger.Error("Failed to update user to admin", logger.Field{Key: "error", Value: err})
			return
		}
		appLogger.Info("Existing user promoted to admin", logger.Field{Key: "phone", Value: adminPhone})
	} else {
		appLogger.Info("Admin user already exists", logger.Field{Key: "phone", Value: adminPhone})
	}
}

func setSwaggerHost(host string, port string) {
	if host == "" {
		docs.SwaggerInfo.Host = "localhost:" + port
	}

	if host != "" && port != "" {
		docs.SwaggerInfo.Host = host + ":" + port
	}
}
