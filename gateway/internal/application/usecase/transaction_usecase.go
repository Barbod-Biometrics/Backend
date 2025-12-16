package usecase

import (
	"context"

	"github.com/Barbod-Biometrics/Backend/gateway/internal/application/dto/profile"
)

type TransactionUsecase interface {
	GetUsageSummary(ctx context.Context, profileID uint64) (*profile.UsageSummaryResponse, error)
}
