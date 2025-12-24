package service

import (
	"context"
	"fmt"
	"time"

	"github.com/Barbod-Biometrics/Backend/gateway/internal/application/dto/wallet"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/application/usecase"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/communication"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/entity"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/enum"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/logger"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/repository"
)

type WalletService struct {
	profileRepo     repository.ProfileRepository
	userRepo        repository.UserRepository
	transactionRepo repository.TransactionRepository
	emailService    communication.EmailService
	unitOfWork      repository.UnitOfWork
	logger          logger.Logger
}

func NewWalletService(
	profileRepo repository.ProfileRepository,
	userRepo repository.UserRepository,
	transactionRepo repository.TransactionRepository,
	emailService communication.EmailService,
	unitOfWork repository.UnitOfWork,
	logger logger.Logger,
) usecase.WalletUsecase {
	return &WalletService{
		profileRepo:     profileRepo,
		userRepo:        userRepo,
		transactionRepo: transactionRepo,
		emailService:    emailService,
		unitOfWork:      unitOfWork,
		logger:          logger,
	}
}

func (s *WalletService) GetWalletSummary(ctx context.Context, userID uint64, profileID uint64) (*wallet.WalletSummaryResponse, error) {
	// Verify the profile belongs to the user
	profile, err := s.getProfileForUser(ctx, userID, profileID)
	if err != nil {
		return nil, err
	}

	// Get wallet summary from transactions
	summary, err := s.transactionRepo.GetWalletSummary(ctx, profileID)
	if err != nil {
		return nil, fmt.Errorf("failed to get wallet summary: %w", err)
	}

	return &wallet.WalletSummaryResponse{
		Success: true,
		Data: wallet.WalletSummaryData{
			Balance:          profile.Balance,
			TotalDeposits:    summary.TotalDeposits,
			TotalWithdrawals: summary.TotalWithdrawals,
			TransactionCount: summary.TransactionCount,
			LastUpdated:      time.Now().Format("2006-01-02"),
		},
	}, nil
}

func (s *WalletService) GetTransactions(ctx context.Context, userID uint64, profileID uint64, page, pageSize int) (*wallet.TransactionsResponse, error) {
	// Verify the profile belongs to the user
	_, err := s.getProfileForUser(ctx, userID, profileID)
	if err != nil {
		return nil, err
	}

	// Get transactions
	result, err := s.transactionRepo.GetByProfileID(ctx, profileID, repository.TransactionPagination{
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get transactions: %w", err)
	}

	// Map to DTOs
	transactions := make([]wallet.TransactionDTO, 0, len(result.Transactions))
	for _, t := range result.Transactions {
		transactions = append(transactions, wallet.TransactionDTO{
			ID:          fmt.Sprintf("%d", t.TransactionID),
			Type:        string(t.TransactionType),
			Amount:      t.Amount,
			Description: t.Notes,
			Status:      "completed", // All stored transactions are completed
			Date:        t.CreatedAt,
		})
	}

	return &wallet.TransactionsResponse{
		Success: true,
		Data: wallet.TransactionsData{
			Transactions: transactions,
			TotalCount:   result.TotalCount,
			Page:         result.Page,
			PageSize:     result.PageSize,
			TotalPages:   result.TotalPages,
		},
	}, nil
}

func (s *WalletService) Deposit(ctx context.Context, userID uint64, profileID uint64, req wallet.DepositRequest) (*wallet.DepositResponse, error) {
	var newBalance uint64
	var transactionID uint64

	err := s.unitOfWork.Do(ctx, func(txCtx context.Context) error {
		// Verify the profile belongs to the user
		profile, err := s.getProfileForUser(txCtx, userID, profileID)
		if err != nil {
			return err
		}

		// Create the transaction record
		transaction := &entity.Transaction{
			ProfileID:       profileID,
			TransactionType: enum.TransactionTypeDeposit,
			Amount:          int64(req.Amount), // Positive for deposit
			Notes:           req.Description,
			CreatedAt:       time.Now(),
		}

		if err := s.transactionRepo.Create(txCtx, transaction); err != nil {
			return fmt.Errorf("failed to create transaction: %w", err)
		}

		// Update the profile balance
		profile.Balance += req.Amount
		if err := s.profileRepo.Update(txCtx, profile); err != nil {
			return fmt.Errorf("failed to update profile balance: %w", err)
		}

		newBalance = profile.Balance
		transactionID = transaction.TransactionID
		return nil
	})

	if err != nil {
		return nil, err
	}

	// Send email notification
	go func() {
		user, err := s.userRepo.GetByID(context.Background(), userID)
		if err != nil || user == nil || user.Email == nil || *user.Email == "" {
			return
		}

		profile, err := s.profileRepo.GetByID(context.Background(), profileID)
		if err != nil || profile == nil {
			return
		}

		name := ""
		if profile.ProfileType == entity.ProfileTypePersonal && profile.PersonDetails != nil {
			name = profile.PersonDetails.FirstName + " " + profile.PersonDetails.LastName
		} else if profile.ProfileType == entity.ProfileTypeBusiness && profile.BusinessDetails != nil {
			name = profile.BusinessDetails.RepFirstName + " " + profile.BusinessDetails.RepLastName
		}

		data := map[string]interface{}{
			"Name":        name,
			"ProfileName": profile.ProfileName,
			"Amount":      req.Amount,
			"NewBalance":  newBalance,
		}

		_ = s.emailService.SendWithTemplate(context.Background(), *user.Email, "Balance Top-up Successful", "balance_topup.html", data)
	}()

	return &wallet.DepositResponse{
		Success: true,
		Data: wallet.DepositDataDTO{
			TransactionID: fmt.Sprintf("%d", transactionID),
			NewBalance:    newBalance,
			Message:       "Deposit successful",
		},
	}, nil
}

func (s *WalletService) getProfileForUser(ctx context.Context, userID uint64, profileID uint64) (*entity.Profile, error) {
	profile, err := s.profileRepo.GetByID(ctx, profileID)
	if err != nil {
		return nil, fmt.Errorf("failed to get profile: %w", err)
	}
	if profile == nil || profile.UserID != userID {
		return nil, fmt.Errorf("profile not found for user ID %d and profile ID %d", userID, profileID)
	}
	return profile, nil
}
