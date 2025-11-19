package main

import (
	"fmt"

	"github.com/Barbod-Biometrics/Backend/gateway/bootstrap"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/logger"
	Logger "github.com/Barbod-Biometrics/Backend/gateway/internal/infrastructure/logger"
	"github.com/gin-gonic/gin"
)

func main() {

	gin.DisableConsoleColor()

	cfg := bootstrap.Run()

	loggerCfg := &Logger.LoggerConfig{
		LogLevel:      cfg.LogLevel,
		ConsoleOutput: cfg.ConsoleOutput,
		LogFile:       cfg.LogFile,
	}
	appLogger, err := Logger.NewLogger(loggerCfg)
	if err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		return
	}
	defer appLogger.Close()

	ginEngine := gin.New()

	appLogger.Info("Starting the gateway application",
		logger.Field{Key: "port", Value: cfg.ServerPort},
		logger.Field{Key: "database", Value: fmt.Sprintf("%s@%s:%s", cfg.PostgresUser, cfg.PostgresHost, cfg.PostgresPort)},
		logger.Field{Key: "redis", Value: fmt.Sprintf("%s:%s", cfg.RedisHost, cfg.RedisPort)},
		logger.Field{Key: "minio", Value: fmt.Sprintf("%s:%s", cfg.MinioHost, cfg.MinioPort)},
	)

	if err := ginEngine.Run(":" + cfg.ServerPort); err != nil {
		appLogger.Error("Error starting server", logger.Field{Key: "error", Value: err})
	}

}
