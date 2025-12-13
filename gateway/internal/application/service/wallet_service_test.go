package service

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/Barbod-Biometrics/Backend/gateway/internal/application/dto/wallet"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/entity"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/enum"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/repository"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/infrastructure/test/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetWalletSummary_Success(t *testing.T) {
	ctx := context.Background()
	mockProfile := mocks.NewMockProfileRepository(t)
	mockTxRepo := mocks.NewMockTransactionRepository(t)
	mockUow := mocks.NewMockUnitOfWork(t)

	svc := NewWalletService(mockProfile, mockTxRepo, mockUow)

	profile := &entity.Profile{ProfileID: 1, UserID: 42, Balance: 1000}
	mockProfile.On("GetByID", mock.Anything, uint64(1)).Return(profile, nil)

	ws := &repository.WalletSummary{Balance: 1000, TotalDeposits: 1500, TotalWithdrawals: 500, TransactionCount: 3}
	mockTxRepo.On("GetWalletSummary", mock.Anything, uint64(1)).Return(ws, nil)

	resp, err := svc.GetWalletSummary(ctx, 42, 1)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, uint64(1000), resp.Data.Balance)
	assert.Equal(t, int64(1500), resp.Data.TotalDeposits)
	assert.Equal(t, int64(500), resp.Data.TotalWithdrawals)
	assert.Equal(t, int64(3), resp.Data.TransactionCount)
	// LastUpdated is formatted daily; just ensure parseable
	_, parseErr := time.Parse("2006-01-02", resp.Data.LastUpdated)
	assert.NoError(t, parseErr)

	mockProfile.AssertExpectations(t)
	mockTxRepo.AssertExpectations(t)
}

func TestGetTransactions_Success(t *testing.T) {
	ctx := context.Background()
	mockProfile := mocks.NewMockProfileRepository(t)
	mockTxRepo := mocks.NewMockTransactionRepository(t)
	mockUow := mocks.NewMockUnitOfWork(t)

	svc := NewWalletService(mockProfile, mockTxRepo, mockUow)

	profile := &entity.Profile{ProfileID: 1, UserID: 100, Balance: 200}
	mockProfile.On("GetByID", mock.Anything, uint64(1)).Return(profile, nil)

	t1 := &entity.Transaction{TransactionID: 11, ProfileID: 1, TransactionType: enum.TransactionTypeDeposit, Amount: 100, Notes: "topup", CreatedAt: time.Now()}
	t2 := &entity.Transaction{TransactionID: 12, ProfileID: 1, TransactionType: enum.TransactionTypeWithdrawal, Amount: -50, Notes: "withdraw", CreatedAt: time.Now()}

	paginated := &repository.TransactionPaginatedResult{
		Transactions: []*entity.Transaction{t1, t2},
		TotalCount:   2,
		Page:         1,
		PageSize:     10,
		TotalPages:   1,
	}

	mockTxRepo.On("GetByProfileID", mock.Anything, uint64(1), repository.TransactionPagination{Page: 1, PageSize: 10}).Return(paginated, nil)

	resp, err := svc.GetTransactions(ctx, 100, 1, 1, 10)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, 2, len(resp.Data.Transactions))
	assert.Equal(t, "11", resp.Data.Transactions[0].ID)
	assert.Equal(t, "deposit", resp.Data.Transactions[0].Type)
	assert.Equal(t, "topup", resp.Data.Transactions[0].Description)
	assert.Equal(t, "completed", resp.Data.Transactions[0].Status)
	assert.Equal(t, 2, len(resp.Data.Transactions))

	mockProfile.AssertExpectations(t)
	mockTxRepo.AssertExpectations(t)
}

func TestDeposit_Success(t *testing.T) {
	ctx := context.Background()
	mockProfile := mocks.NewMockProfileRepository(t)
	mockTxRepo := mocks.NewMockTransactionRepository(t)
	mockUow := mocks.NewMockUnitOfWork(t)

	svc := NewWalletService(mockProfile, mockTxRepo, mockUow)

	profile := &entity.Profile{ProfileID: 100, UserID: 500, Balance: 1000}
	mockProfile.On("GetByID", mock.Anything, uint64(100)).Return(profile, nil)

	// Mock unit of work to execute the passed transaction function
	mockUow.EXPECT().Do(mock.Anything, mock.Anything).RunAndReturn(func(ctx context.Context, fn func(ctx context.Context) error) error {
		return fn(ctx)
	})

	// Capture Create to set TransactionID
	mockTxRepo.On("Create", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		tx := args.Get(1).(*entity.Transaction)
		tx.TransactionID = 999
	}).Return(nil)

	// Expect profile update (balance increased)
	mockProfile.On("Update", mock.Anything, mock.MatchedBy(func(p *entity.Profile) bool { return p.Balance == 1500 })).Return(nil)

	req := wallet.DepositRequest{Amount: 500, Description: "deposit"}
	resp, err := svc.Deposit(ctx, 500, 100, req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "999", resp.Data.TransactionID)
	assert.Equal(t, uint64(1500), resp.Data.NewBalance)
	assert.Equal(t, "Deposit successful", resp.Data.Message)

	mockProfile.AssertExpectations(t)
	mockTxRepo.AssertExpectations(t)
	mockUow.AssertExpectations(t)
}

func TestDeposit_Failure_CreateOrUpdate(t *testing.T) {
	ctx := context.Background()
	mockProfile := mocks.NewMockProfileRepository(t)
	mockTxRepo := mocks.NewMockTransactionRepository(t)
	mockUow := mocks.NewMockUnitOfWork(t)

	svc := NewWalletService(mockProfile, mockTxRepo, mockUow)

	profile := &entity.Profile{ProfileID: 200, UserID: 600, Balance: 100}
	mockProfile.On("GetByID", mock.Anything, uint64(200)).Return(profile, nil)

	mockUow.EXPECT().Do(mock.Anything, mock.Anything).RunAndReturn(func(ctx context.Context, fn func(ctx context.Context) error) error {
		return fn(ctx)
	})

	// Make Create fail
	mockTxRepo.On("Create", mock.Anything, mock.Anything).Return(errors.New("insert failed"))

	req := wallet.DepositRequest{Amount: 100, Description: "deposit"}
	resp, err := svc.Deposit(ctx, 600, 200, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "insert failed")

	// Now test update failure path — reset expectations
	mockTxRepo.ExpectedCalls = nil
	mockTxRepo.Test(t)
	mockProfile.ExpectedCalls = nil
	mockProfile.Test(t)

	// Prepare again
	mockProfile.On("GetByID", mock.Anything, uint64(200)).Return(profile, nil)
	mockUow.EXPECT().Do(mock.Anything, mock.Anything).RunAndReturn(func(ctx context.Context, fn func(ctx context.Context) error) error { return fn(ctx) })
	mockTxRepo.On("Create", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		tx := args.Get(1).(*entity.Transaction)
		tx.TransactionID = 333
	}).Return(nil)
	mockProfile.On("Update", mock.Anything, mock.Anything).Return(fmt.Errorf("update failed"))

	resp, err = svc.Deposit(ctx, 600, 200, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "update failed")

	mockProfile.AssertExpectations(t)
	mockTxRepo.AssertExpectations(t)
	mockUow.AssertExpectations(t)
}
