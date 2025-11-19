package main

import (
	"fmt"

	"github.com/Barbod-Biometrics/Backend/gateway/bootstrap"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/enum"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/logger"
	Logger "github.com/Barbod-Biometrics/Backend/gateway/internal/infrastructure/logger"
	"github.com/gin-gonic/gin"
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

	ginEngine := gin.New()

	appLogger.Info("Starting the gateway application",
		logger.Field{Key: "port", Value: cfg.Env.Server.Port},
		logger.Field{Key: "database", Value: fmt.Sprintf("%s@%s:%s", cfg.Env.Postgres.User, cfg.Env.Postgres.Host, cfg.Env.Postgres.Port)},
		logger.Field{Key: "redis", Value: fmt.Sprintf("%s:%s", cfg.Env.PrimaryRedis.Address, cfg.Env.PrimaryRedis.Port)},
		logger.Field{Key: "minio", Value: fmt.Sprintf("%s:%s", cfg.Env.Minio.Host, cfg.Env.Minio.Port)},
	)

	if err := ginEngine.Run(":" + cfg.Env.Server.Port); err != nil {
		appLogger.Error("Error starting server", logger.Field{Key: "error", Value: err})
	}

}
