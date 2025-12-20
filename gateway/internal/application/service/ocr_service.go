package service

import (
	"context"
	"fmt"
	"time"

	ocrDto "github.com/Barbod-Biometrics/Backend/gateway/internal/application/dto/ocr"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/application/usecase"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/entity"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/enum"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/exception"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/logger"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/repository"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/infrastructure/ocr"
)

type OCRService struct {
	client          *ocr.OCRClient
	logger          logger.Logger
	repo            repository.OCRRepository
	serviceRepo     repository.ServiceRepository
	profileRepo     repository.ProfileRepository
	transactionRepo repository.TransactionRepository
	unitOfWork      repository.UnitOfWork
}

func NewOCRService(
	cli *ocr.OCRClient,
	logger logger.Logger,
	repo repository.OCRRepository,
	serviceRepo repository.ServiceRepository,
	profileRepo repository.ProfileRepository,
	transactionRepo repository.TransactionRepository,
	unitOfWork repository.UnitOfWork,
) *OCRService {
	return &OCRService{
		client:          cli,
		logger:          logger,
		repo:            repo,
		serviceRepo:     serviceRepo,
		profileRepo:     profileRepo,
		transactionRepo: transactionRepo,
		unitOfWork:      unitOfWork,
	}
}

var _ usecase.OCRUsecase = (*OCRService)(nil)

func (s *OCRService) ExtractText(ctx context.Context, profileID uint64, image []byte) (*ocrDto.OCRResponse, error) {
	s.logger.Info("Starting OCR text extraction process")

	if len(image) == 0 {
		s.logger.Warn("Image is empty")
		return nil, exception.ErrEmptyImage
	}

	// Get service cost
	service, err := s.serviceRepo.GetByName(ctx, "ocr")
	if err != nil {
		s.logger.Error("Failed to get OCR service",
			logger.Field{Key: "error", Value: err},
		)
		return nil, exception.ErrServiceNotFound
	}

	if !service.IsAvailable {
		s.logger.Warn("OCR service is not available")
		return nil, fmt.Errorf("OCR service is currently unavailable")
	}

	serviceCost := uint64(service.CurrentCost)

	// Check wallet balance and deduct cost
	if profileID != 0 {
		err = s.unitOfWork.Do(ctx, func(txCtx context.Context) error {
			profile, err := s.profileRepo.GetByID(txCtx, profileID)
			if err != nil {
				return fmt.Errorf("failed to get profile: %w", err)
			}

			if profile.Balance < serviceCost {
				s.logger.Warn("Insufficient balance for OCR",
					logger.Field{Key: "profile_id", Value: profileID},
					logger.Field{Key: "balance", Value: profile.Balance},
					logger.Field{Key: "cost", Value: serviceCost},
				)
				return exception.ErrInsufficientFunds
			}

			profile.Balance -= serviceCost
			if err := s.profileRepo.Update(txCtx, profile); err != nil {
				return fmt.Errorf("failed to update profile balance: %w", err)
			}

			transaction := &entity.Transaction{
				ProfileID:       profileID,
				TransactionType: enum.TransactionTypeWithdrawal,
				Amount:          -int64(serviceCost),
				Notes:           "OCR service charge",
				CreatedAt:       time.Now(),
			}

			if err := s.transactionRepo.Create(txCtx, transaction); err != nil {
				return fmt.Errorf("failed to create transaction: %w", err)
			}

			s.logger.Info("Wallet deducted for OCR",
				logger.Field{Key: "profile_id", Value: profileID},
				logger.Field{Key: "amount", Value: serviceCost},
				logger.Field{Key: "new_balance", Value: profile.Balance},
			)

			return nil
		})

		if err != nil {
			return nil, err
		}
	}

	// Perform OCR extraction
	result, err := s.client.ExtractText(ctx, image)
	if err != nil {
		s.logger.Error("OCR extraction failed",
			logger.Field{Key: "error", Value: err},
		)
		return nil, fmt.Errorf("OCR extraction failed: %w", err)
	}

	if !result.Success {
		s.logger.Warn("OCR extraction returned failure",
			logger.Field{Key: "message", Value: result.Message},
		)
	} else {
		s.logger.Info("OCR extraction completed successfully")
	}

	// Save result to database
	if s.repo != nil && profileID != 0 {
		var rec entity.OCRRecord
		rec.Success = result.Success
		rec.Message = result.Message


		if result != nil && result.Stats != nil {
			rec.Stats = result.Stats
		}

		s.logger.Info("Persisting OCR result",
			logger.Field{Key: "rec", Value: rec},
		)

		if err := s.repo.SaveResult(ctx, profileID, &rec); err != nil {
			s.logger.Error("Failed to persist OCR result",
				logger.Field{Key: "error", Value: err},
				logger.Field{Key: "profile_id", Value: profileID},
			)
		}
	}

	return result, nil
}

func (s *OCRService) HealthCheck(ctx context.Context) (*ocrDto.HealthCheckResponseDTO, error) {
	s.logger.Info("Performing health check on OCR service")

	result, err := s.client.HealthCheck(ctx)
	if err != nil {
		s.logger.Error("Health check failed",
			logger.Field{Key: "error", Value: err},
		)
		return nil, fmt.Errorf("health check failed: %w", err)
	}

	s.logger.Info("Health check status",
		logger.Field{Key: "status", Value: result.Status},
	)
	return result, nil
}
