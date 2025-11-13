package main

import (
	"fmt"

	"github.com/Barbod-Biometrics/Backend/gateway/pkg/config"
	"github.com/gin-gonic/gin"
)

func main() {

	gin.DisableConsoleColor()

	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		return
	}

	ginEngine := gin.New()

	fmt.Printf("Starting the gateway application on port %s...\n", cfg.ServerPort)
	fmt.Printf("Database: %s@%s:%s\n", cfg.PostgresUser, cfg.PostgresHost, cfg.PostgresPort)
	fmt.Printf("Redis: %s:%s\n", cfg.RedisHost, cfg.RedisPort)
	fmt.Printf("MinIO: %s:%s\n", cfg.MinioHost, cfg.MinioPort)

	if err := ginEngine.Run(":" + cfg.ServerPort); err != nil {
		fmt.Printf("Error starting server: %v\n", err)
	}

}
