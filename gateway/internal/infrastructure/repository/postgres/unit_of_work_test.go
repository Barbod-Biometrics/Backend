package postgres_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/entity"
	repoPostgres "github.com/Barbod-Biometrics/Backend/gateway/internal/infrastructure/repository/postgres"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/test/testcontainer"
	"github.com/stretchr/testify/assert"
)

func TestGormUnitOfWork_Do_Commit(t *testing.T) {
	ctx := context.Background()
	pg, err := testcontainer.SetupPostgres(ctx)
	if err != nil {
		t.Fatalf("failed to setup postgres container: %v", err)
	}
	defer pg.Terminate(ctx)

	uow := repoPostgres.NewGormUnitOfWork(pg.DB)
	userRepo := repoPostgres.NewUserRepository(pg.DB)

	user := &entity.User{
		PhoneNumber: "09181234567",
		CreatedAt:   time.Now(),
	}

	// Successful transaction
	err = uow.Do(ctx, func(txCtx context.Context) error {
		return userRepo.CreateUser(txCtx, user)
	})
	assert.NoError(t, err)

	// Verify it is persisted
	var count int64
	pg.DB.Model(&entity.User{}).Where("phone_number = ?", "09181234567").Count(&count)
	assert.Equal(t, int64(1), count)
}

func TestGormUnitOfWork_Do_Rollback(t *testing.T) {
	ctx := context.Background()
	pg, err := testcontainer.SetupPostgres(ctx)
	if err != nil {
		t.Fatalf("failed to setup postgres container: %v", err)
	}
	defer pg.Terminate(ctx)

	uow := repoPostgres.NewGormUnitOfWork(pg.DB)
	userRepo := repoPostgres.NewUserRepository(pg.DB)

	user := &entity.User{
		PhoneNumber: "09187654321",
		CreatedAt:   time.Now(),
	}

	// Failed transaction
	expectedErr := errors.New("boom")
	err = uow.Do(ctx, func(txCtx context.Context) error {
		if err := userRepo.CreateUser(txCtx, user); err != nil {
			return err
		}
		return expectedErr // Force rollback
	})
	assert.ErrorIs(t, err, expectedErr)

	// Verify rollback (not persisted)
	var count int64
	pg.DB.Model(&entity.User{}).Where("phone_number = ?", "09187654321").Count(&count)
	assert.Equal(t, int64(0), count)
}
