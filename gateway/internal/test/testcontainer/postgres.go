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

	// AutoMigrate entities
	// We include entities that appeared in the tests
	err = db.AutoMigrate(
		&entity.Profile{},
		&entity.ProfilePersonDetails{},
		&entity.ProfileBusinessDetails{},
		&entity.APIKey{},
		&entity.User{},
		&entity.Transaction{},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to migrate: %w", err)
	}

	return &PostgresContainer{
		Container: pgContainer,
		DB:        db,
		DSN:       connStr,
	}, nil
}

func (c *PostgresContainer) Terminate(ctx context.Context) error {
	return c.Container.Terminate(ctx)
}
