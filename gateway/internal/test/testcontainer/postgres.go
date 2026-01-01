package testcontainer

import (
	"context"
	"fmt"
	"time"

	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/entity"
	"github.com/docker/go-connections/nat"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	gormPostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"

	_ "github.com/lib/pq" // required for wait.ForSQL
)

type PostgresContainer struct {
	Container *postgres.PostgresContainer
	DB        *gorm.DB
	DSN       string
}

func SetupPostgres(ctx context.Context) (*PostgresContainer, error) {
	pgContainer, err := postgres.Run(ctx,
		"postgres:latest",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("user"),
		postgres.WithPassword("password"),
		testcontainers.WithWaitStrategy(
			wait.ForSQL("5432/tcp", "postgres", func(host string, port nat.Port) string {
				return fmt.Sprintf("host=%s port=%s user=user password=password dbname=testdb sslmode=disable", host, port.Port())
			}).
				WithStartupTimeout(30*time.Second),
		),
	)
	if err != nil {
		return nil, err
	}

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		return nil, err
	}

	db, err := gorm.Open(gormPostgres.Open(connStr), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// We migrate tables one by one. If one fails, we know exactly which one.

	// 1. User (Independent - Must be first)
	if err := db.AutoMigrate(&entity.User{}); err != nil {
		return nil, fmt.Errorf("failed to migrate User: %w", err)
	}

	// 2. Profile (Depends on User)
	if err := db.AutoMigrate(&entity.Profile{}); err != nil {
		return nil, fmt.Errorf("failed to migrate Profile: %w", err)
	}

	// 3. Profile Details (Depend on Profile)
	if err := db.AutoMigrate(&entity.ProfilePersonDetails{}); err != nil {
		return nil, fmt.Errorf("failed to migrate PersonDetails: %w", err)
	}
	if err := db.AutoMigrate(&entity.ProfileBusinessDetails{}); err != nil {
		return nil, fmt.Errorf("failed to migrate BusinessDetails: %w", err)
	}

	// 4. API Key (Depends on Profile)
	// Now that 'profiles' table definitely exists, this will succeed.
	if err := db.AutoMigrate(&entity.APIKey{}); err != nil {
		return nil, fmt.Errorf("failed to migrate APIKey: %w", err)
	}

	if err := db.AutoMigrate(&entity.Ticket{}); err != nil {
		return nil, fmt.Errorf("failed to migrate Ticket: %w", err)
	}

	// 5. Others
	if err := db.AutoMigrate(&entity.Transaction{}); err != nil {
		return nil, fmt.Errorf("failed to migrate Transaction: %w", err)
	}

	// AutoMigrate entities
	// err = db.AutoMigrate(
	// 	&entity.User{},                   // 1. User (Independent)
	// 	&entity.Profile{},                // 2. Profile (Depends on User)
	// 	&entity.ProfilePersonDetails{},   // 3. Details (Depends on Profile)
	// 	&entity.ProfileBusinessDetails{}, // 4. Details (Depends on Profile)
	// 	&entity.APIKey{},                 // 5. APIKey (Depends on Profile)
	// 	&entity.Transaction{},            // 6. Others...
	// )
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to migrate: %w", err)
	// }

	return &PostgresContainer{
		Container: pgContainer,
		DB:        db,
		DSN:       connStr,
	}, nil
}

func (c *PostgresContainer) Terminate(ctx context.Context) error {
	return c.Container.Terminate(ctx)
}
