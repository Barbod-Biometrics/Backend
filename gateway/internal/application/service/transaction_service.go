package service

import (
	"context"

	"github.com/Barbod-Biometrics/Backend/gateway/internal/application/dto/profile"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/application/usecase"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/logger"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/repository"
)

type TransactionService struct {
	transactionRepo repository.TransactionRepository
	logger          logger.Logger
}

var _ usecase.TransactionUsecase = &TransactionService{}

func NewTransactionService(transactionRepo repository.TransactionRepository, logger logger.Logger) *TransactionService {
	return &TransactionService{
		transactionRepo: transactionRepo,
		logger:          logger,
	}
}

func (s *TransactionService) GetUsageSummary(ctx context.Context, profileID uint64) (*profile.UsageSummaryResponse, error) {
	s.logger.Info("processing usage summary request",
		logger.Field{Key: "profile_id", Value: profileID},
	)

	summary, err := s.transactionRepo.GetUsageSummary(ctx, profileID)
	if err != nil {
		s.logger.Error("failed to retrieve usage summary from repository",
			logger.Field{Key: "profile_id", Value: profileID},
			logger.Field{Key: "error", Value: err},
		)
		return nil, err
	}

	s.logger.Info("usage summary retrieved successfully",
		logger.Field{Key: "profile_id", Value: profileID},
	)

	return summary, nil

}
