package main

import (
	"fmt"
	"net/http"

	"github.com/Barbod-Biometrics/Backend/gateway/bootstrap"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/application/service"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/entity"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/enum"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/logger"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/infrastructure/communication/sms"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/infrastructure/database/redis"
	infraJWT "github.com/Barbod-Biometrics/Backend/gateway/internal/infrastructure/jwt"
	Logger "github.com/Barbod-Biometrics/Backend/gateway/internal/infrastructure/logger"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/infrastructure/repository/postgres"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/presentation/controller/v1/user"
	v1 "github.com/Barbod-Biometrics/Backend/gateway/internal/presentation/routes/http/v1"
	"github.com/gin-gonic/gin"
	driverPostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

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

	// -- Postgres --
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		cfg.Env.Postgres.Host, cfg.Env.Postgres.User, cfg.Env.Postgres.Password,
		cfg.Env.Postgres.DBName, cfg.Env.Postgres.Port)

	db, err := gorm.Open(driverPostgres.Open(dsn), &gorm.Config{})
	if err != nil {
		appLogger.Fatal("Failed to connect to Postgres", logger.Field{Key: "error", Value: err})
	}

	// --- Auto Migrate ---
	if err := db.AutoMigrate(&entity.User{}); err != nil {
		appLogger.Fatal("Failed to migrate database", logger.Field{Key: "error", Value: err})
	}

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

	// 5. Initialize Application Layer (Services)
	jwtService := service.NewJWTService(cfg, jwtKeyManager)
	otpService := service.NewOTPService(cacheRepo, cfg)

	// 6. Initialize Usecases
	authUsecase := service.NewAuthUsecase(userRepo, otpService, smsService, jwtService)

	// 7. Initialize Controllers
	authController := user.NewAuthController(authUsecase)

	// 8. Setup Router
	// Initialize the V1 Router
	v1Router := v1.NewRouter(authController)

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
