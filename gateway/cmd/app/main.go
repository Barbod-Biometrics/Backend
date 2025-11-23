package main

import (
	"log"

	"github.com/Barbod-Biometrics/Backend/gateway/bootstrap"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/application/service"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/entity"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/infrastructure/repository/postgres"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/presentation/handler"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/presentation/routes"
	"github.com/Barbod-Biometrics/Backend/gateway/pkg/database"
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

	db := database.NewPostgresDatabase(cfg.Env)

	err := db.AutoMigrate(
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
	profileService := service.NewProfileService(profileRepo)
	profileHandler := handler.NewProfileHandler(profileService)

	ginEngine := gin.New()
	ginEngine.Use(gin.Logger())
	ginEngine.Use(gin.Recovery())

	routes.Setup(ginEngine, profileHandler)

	if err := ginEngine.Run(":" + cfg.Env.Server.Port); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}

}
