package postgres

import (
	"context"
	"fmt"
	"testing"

	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/entity"
	"github.com/Barbod-Biometrics/Backend/gateway/pkg/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	cfg, err := config.LoadConfig()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.PostgresUser,
		cfg.PostgresPass,
		cfg.PostgresHost,
		cfg.PostgresPort,
		cfg.PostgresName,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect to database: %v", err)
	}

	err = db.AutoMigrate(&entity.User{})
	if err != nil {
		t.Fatalf("failed to migrate database: %v", err)
	}

	db.Exec("TRUNCATE TABLE users RESTART IDENTITY CASCADE")

	return db
}

func TestUserRepository_CreateAndGet(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	phone := "09124206969"
	email := "info@barbod.com"
	newUser := &entity.User{
		PhoneNumber: phone,
		Email:       &email,
	}

	t.Run("Create User", func(t *testing.T) {
		err := repo.CreateUser(ctx, newUser)
		if err != nil {
			t.Fatalf("failed to create user: %v", err)
		}
	})

	t.Run("Get User By Phone", func(t *testing.T) {
		user, err := repo.GetByPhoneNumber(ctx, phone)
		if err != nil {
			t.Fatalf("failed to get user by phone: %v", err)
		}
		if user.PhoneNumber != newUser.PhoneNumber || (user.Email == nil || *user.Email != *newUser.Email) {
			t.Fatalf("retrieved user by phone does not match created user")
		}
	})

	t.Run("Get User By Email", func(t *testing.T) {
		user, err := repo.GetByEmail(ctx, *newUser.Email)
		if err != nil {
			t.Fatalf("failed to get user by email: %v", err)
		}
		if user.PhoneNumber != newUser.PhoneNumber || (user.Email == nil || *user.Email != *newUser.Email) {
			t.Fatalf("retrieved user by email does not match created user")
		}
	})
}
