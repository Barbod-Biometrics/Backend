package main

import (
	"fmt"

	"github.com/Barbod-Biometrics/Backend/gateway/bootstrap"
	"github.com/gin-gonic/gin"
)

func main() {

	gin.DisableConsoleColor()

	cfg := bootstrap.Run()

	ginEngine := gin.New()

	fmt.Printf("Starting the gateway application on port %s...\n", cfg.Env.Server.Port)
	fmt.Printf("Database: %s@%s:%s\n", cfg.Env.Postgres.User, cfg.Env.Postgres.Host, cfg.Env.Postgres.Port)
	fmt.Printf("Redis: %s:%s\n", cfg.Env.PrimaryRedis.Address, cfg.Env.PrimaryRedis.Port)
	fmt.Printf("MinIO: %s:%s\n", cfg.Env.Minio.Host, cfg.Env.Minio.Port)

	if err := ginEngine.Run(":" + cfg.Env.Server.Port); err != nil {
		fmt.Printf("Error starting server: %v\n", err)
	}

}
