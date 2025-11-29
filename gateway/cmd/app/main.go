package main

import (
	"fmt"
	"log"

	"github.com/Barbod-Biometrics/Backend/gateway/bootstrap"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/application/service"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/entity"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/enum"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/logger"
	Logger "github.com/Barbod-Biometrics/Backend/gateway/internal/infrastructure/logger"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/infrastructure/repository/postgres"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/presentation/controller/v1/profile"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/presentation/routes"
	"github.com/Barbod-Biometrics/Backend/gateway/pkg/database"
	"github.com/Barbod-Biometrics/Backend/gateway/pkg/storage"
	"github.com/gin-gonic/gin"
)

// @title Barbod Biometrics Gateway API
// @version 1.0
// @description This is the Gateway service API documentation for Barbod Biometrics.
// @host localhost:8080
// @BasePath /api/v1
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

	db := database.NewPostgresDatabase(cfg.Env)

	err = db.AutoMigrate(
		&entity.User{},
		&entity.Profile{},
		&entity.ProfileBusinessDetails{},
		&entity.ProfilePersonDetails{},
		&entity.AuthorizedSignatory{},
	)

	if err != nil {
		log.Fatalf("Database migration failed: %v", err)
	}

	profileRepo := postgres.NewProfileRepository(db)
	minioStorage, err := storage.NewMinioClient(*cfg.Env)
	if err != nil {
		log.Fatalf("Failed to create MinIO client: %v", err)
	}
	profileService := service.NewProfileService(profileRepo, minioStorage)
	profileHandler := profile.NewProfileHandler(profileService)

	ginEngine := gin.New()

	appLogger.Info("Starting the gateway application",
		logger.Field{Key: "port", Value: cfg.Env.Server.Port},
		logger.Field{Key: "database", Value: fmt.Sprintf("%s@%s:%s", cfg.Env.Postgres.User, cfg.Env.Postgres.Host, cfg.Env.Postgres.Port)},
		logger.Field{Key: "redis", Value: fmt.Sprintf("%s:%s", cfg.Env.PrimaryRedis.Address, cfg.Env.PrimaryRedis.Port)},
		logger.Field{Key: "minio", Value: fmt.Sprintf("%s:%s", cfg.Env.Minio.Host, cfg.Env.Minio.Port)},
	)

	ginEngine.Use(gin.Logger())
	ginEngine.Use(gin.Recovery())

	routes.Setup(ginEngine, profileHandler)

	if err := ginEngine.Run(":" + cfg.Env.Server.Port); err != nil {
		appLogger.Error("Error starting server", logger.Field{Key: "error", Value: err})
	}
}
