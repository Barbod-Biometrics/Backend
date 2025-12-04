package repository

import (
	"context"

	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/entity"
)

// TransactionFilter contains filter options for listing transactions
type TransactionFilter struct {
	ProfileID       uint64
	TransactionType *string
}

// TransactionPagination defines pagination options for transactions
type TransactionPagination struct {
	Page     int
	PageSize int
}

// TransactionPaginatedResult contains paginated transaction results
type TransactionPaginatedResult struct {
	Transactions []*entity.Transaction
	TotalCount   int64
	Page         int
	PageSize     int
	TotalPages   int
}

// WalletSummary contains aggregated wallet information
type WalletSummary struct {
	Balance          uint64 // Current balance from profile
	TotalDeposits    int64  // Sum of all deposits
	TotalWithdrawals int64  // Sum of all withdrawals (absolute value)
	TransactionCount int64  // Total number of transactions
}

type TransactionRepository interface {
	Create(ctx context.Context, transaction *entity.Transaction) error
	GetByID(ctx context.Context, transactionID uint64) (*entity.Transaction, error)
	GetByProfileID(ctx context.Context, profileID uint64, pagination TransactionPagination) (*TransactionPaginatedResult, error)
	GetWalletSummary(ctx context.Context, profileID uint64) (*WalletSummary, error)
}
