package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/entity"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/enum"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/repository"
	repoPostgres "github.com/Barbod-Biometrics/Backend/gateway/internal/infrastructure/repository/postgres"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/infrastructure/test/mocks"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/test/testcontainer"
	"github.com/stretchr/testify/assert"
)

func TestTransactionRepository_Create(t *testing.T) {
	ctx := context.Background()
	pg, err := testcontainer.SetupPostgres(ctx)
	if err != nil {
		t.Fatalf("failed to setup postgres container: %v", err)
	}
	defer pg.Terminate(ctx)

	mockLogger := mocks.NewMockAppLogger(t)
	repo := repoPostgres.NewTransactionRepository(pg.DB, mockLogger)

	tx := &entity.Transaction{
		ProfileID:       100,
		TransactionType: enum.TransactionTypeDeposit,
		Amount:          50000,
		Notes:           "Initial deposit",
		CreatedAt:       time.Now(),
	}

	err = repo.Create(ctx, tx)
	assert.NoError(t, err)
	assert.NotZero(t, tx.TransactionID)

	// Verify persistence
	var savedTx entity.Transaction
	err = pg.DB.First(&savedTx, tx.TransactionID).Error
	assert.NoError(t, err)
	assert.Equal(t, tx.Amount, savedTx.Amount)
}

func TestTransactionRepository_GetByID(t *testing.T) {
	ctx := context.Background()
	pg, err := testcontainer.SetupPostgres(ctx)
	if err != nil {
		t.Fatalf("failed to setup postgres container: %v", err)
	}
	defer pg.Terminate(ctx)

	mockLogger := mocks.NewMockAppLogger(t)
	repo := repoPostgres.NewTransactionRepository(pg.DB, mockLogger)

	// Seed
	tx := entity.Transaction{
		ProfileID:       101,
		TransactionType: enum.TransactionTypeWithdrawal,
		Amount:          -20000,
		CreatedAt:       time.Now(),
	}
	err = pg.DB.Create(&tx).Error
	assert.NoError(t, err)

	// Found
	{
		foundTx, err := repo.GetByID(ctx, tx.TransactionID)
		assert.NoError(t, err)
		assert.NotNil(t, foundTx)
		assert.Equal(t, tx.Amount, foundTx.Amount)
	}

	// Not Found - GetByID returns error on GORM not found based on implementation analysis?
	// Let's check implementation:
	// func (r *TransactionRepository) GetByID(...) { ... err := db...First(...).Error; if err != nil { return nil, err } ... }
	// So it returns error.
	{
		foundTx, err := repo.GetByID(ctx, 999999)
		assert.Error(t, err)
		assert.Nil(t, foundTx)
	}
}

func TestTransactionRepository_GetByProfileID(t *testing.T) {
	ctx := context.Background()
	pg, err := testcontainer.SetupPostgres(ctx)
	if err != nil {
		t.Fatalf("failed to setup postgres container: %v", err)
	}
	defer pg.Terminate(ctx)

	mockLogger := mocks.NewMockAppLogger(t)
	repo := repoPostgres.NewTransactionRepository(pg.DB, mockLogger)
	profileID := uint64(200)

	// Seed multiple
	for i := 0; i < 15; i++ {
		tx := entity.Transaction{
			ProfileID:       profileID,
			TransactionType: enum.TransactionTypeDeposit,
			Amount:          1000,
			CreatedAt:       time.Now().Add(time.Duration(i) * time.Minute),
		}
		if i%2 == 0 {
			tx.ProfileID = 201 // Different profile
		}
		pg.DB.Create(&tx)
	}

	// Pagination
	// Profile 200 should have roughly 7 or 8 txs.
	// Total 15 iterations:
	// i=0: pid=201
	// i=1: pid=200
	// ...
	// Evens are 201 (0, 2, ... 14) -> 8 items
	// Odds are 200 (1, 3, ... 13) -> 7 items

	pagination := repository.TransactionPagination{
		Page:     1,
		PageSize: 5,
	}

	result, err := repo.GetByProfileID(ctx, profileID, pagination)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, int64(7), result.TotalCount)
	assert.Len(t, result.Transactions, 5) // Page size constraint
	assert.Equal(t, 2, result.TotalPages) // 7 items, page size 5 -> 2 pages
}

func TestTransactionRepository_GetWalletSummary(t *testing.T) {
	ctx := context.Background()
	pg, err := testcontainer.SetupPostgres(ctx)
	if err != nil {
		t.Fatalf("failed to setup postgres container: %v", err)
	}
	defer pg.Terminate(ctx)

	mockLogger := mocks.NewMockAppLogger(t)
	repo := repoPostgres.NewTransactionRepository(pg.DB, mockLogger)
	profileID := uint64(300)

	// Seed
	// Deposit 1000
	pg.DB.Create(&entity.Transaction{
		ProfileID:       profileID,
		TransactionType: enum.TransactionTypeDeposit,
		Amount:          1000,
		CreatedAt:       time.Now(),
	})
	// Deposit 500
	pg.DB.Create(&entity.Transaction{
		ProfileID:       profileID,
		TransactionType: enum.TransactionTypeDeposit,
		Amount:          500,
		CreatedAt:       time.Now(),
	})
	// Withdrawal -200
	pg.DB.Create(&entity.Transaction{
		ProfileID:       profileID,
		TransactionType: enum.TransactionTypeWithdrawal,
		Amount:          -200,
		CreatedAt:       time.Now(),
	})

	// Diff profile
	pg.DB.Create(&entity.Transaction{
		ProfileID:       301,
		TransactionType: enum.TransactionTypeDeposit,
		Amount:          9999,
		CreatedAt:       time.Now(),
	})

	summary, err := repo.GetWalletSummary(ctx, profileID)
	assert.NoError(t, err)
	assert.Equal(t, int64(1500), summary.TotalDeposits)
	assert.Equal(t, int64(200), summary.TotalWithdrawals) // Absolute value
	assert.Equal(t, int64(3), summary.TransactionCount)
}
