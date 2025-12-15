package usecase

import (
	"context"

	"github.com/Barbod-Biometrics/Backend/gateway/internal/application/dto/profile"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/application/dto/wallet"
)

type WalletUsecase interface {
	GetWalletSummary(ctx context.Context, userID uint64, profileID uint64) (*wallet.WalletSummaryResponse, error)
	GetTransactions(ctx context.Context, userID uint64, profileID uint64, page, pageSize int) (*wallet.TransactionsResponse, error)
	Deposit(ctx context.Context, userID uint64, profileID uint64, req wallet.DepositRequest) (*wallet.DepositResponse, error)
	GetUsageSummary(ctx context.Context, profileID uint64) (*profile.UsageSummaryResponse, error)
}
