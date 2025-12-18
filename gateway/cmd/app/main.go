package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	docs "github.com/Barbod-Biometrics/Backend/gateway/docs"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/entity"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/logger"
	postgresRepo "github.com/Barbod-Biometrics/Backend/gateway/internal/infrastructure/repository/postgres"

	appWire "github.com/Barbod-Biometrics/Backend/gateway/wire"
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
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name X-API-KEY
// @description Provide the API key in the `X-API-KEY` header for requests to protected endpoints

func main() {

	gin.DisableConsoleColor()

	app, err := appWire.InitializeApplication()
	if err != nil {
		fmt.Printf("FATAL: Failed to initialize application: %v\n", err)
		return
	}

	defer app.Close()

	cfg := app.Config

	app.Logger.Info("Application initialized successfully", logger.Field{Key: "logfile", Value: cfg.Env.Logger.LogFile})

	// Swagger setup
	setSwaggerHost(cfg.Env.Server.Host, cfg.Env.Server.Port)

	// Database Migration
	err = app.DB.AutoMigrate(
		&entity.User{},
		&entity.Profile{},
		&entity.ProfileBusinessDetails{},
		&entity.ProfilePersonDetails{},
		&entity.AuthorizedSignatory{},
		&entity.APIKey{},
		&entity.Transaction{},
		&postgresRepo.FaceVerificationModel{},
	)
	if err != nil {
		app.Logger.Fatal("Failed to migrate database", logger.Field{Key: "error", Value: err})
	}

	// Seed Admin User
	seedAdminUser(app.DB, cfg.Env.Admin.PhoneNumber, app.Logger)

	// Server Start
	handler := app.Route.RegisterRoutes()

	app.Logger.Info("Starting the gateway application",
		logger.Field{Key: "port", Value: cfg.Env.Server.Port},
		logger.Field{Key: "database", Value: fmt.Sprintf("%s@%s:%s", cfg.Env.Postgres.User, cfg.Env.Postgres.Host, cfg.Env.Postgres.Port)},
		logger.Field{Key: "redis", Value: fmt.Sprintf("%s:%s", cfg.Env.PrimaryRedis.Address, cfg.Env.PrimaryRedis.Port)},
		logger.Field{Key: "minio", Value: fmt.Sprintf("%s:%s", cfg.Env.Minio.Host, cfg.Env.Minio.Port)},
	)

	if err := http.ListenAndServe(":"+cfg.Env.Server.Port, handler); err != nil {
		app.Logger.Error("Error starting server", logger.Field{Key: "error", Value: err})
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
