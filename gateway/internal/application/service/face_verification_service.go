package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	faceVerificationDto "github.com/Barbod-Biometrics/Backend/gateway/internal/application/dto/face_verification"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/application/usecase"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/entity"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/enum"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/exception"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/logger"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/repository"
	face_verification "github.com/Barbod-Biometrics/Backend/gateway/internal/infrastructure/face_verificaiton"
)

type FaceVerificationService struct {
	client          *face_verification.FaceVerificationClient
	logger          logger.Logger
	repo            repository.FaceVerificationRepository
	serviceRepo     repository.ServiceRepository
	profileRepo     repository.ProfileRepository
	transactionRepo repository.TransactionRepository
	unitOfWork      repository.UnitOfWork
}

func NewFaceVerificationService(
	cli *face_verification.FaceVerificationClient,
	logger logger.Logger,
	repo repository.FaceVerificationRepository,
	serviceRepo repository.ServiceRepository,
	profileRepo repository.ProfileRepository,
	transactionRepo repository.TransactionRepository,
	unitOfWork repository.UnitOfWork,
) *FaceVerificationService {
	return &FaceVerificationService{
		client:          cli,
		logger:          logger,
		repo:            repo,
		serviceRepo:     serviceRepo,
		profileRepo:     profileRepo,
		transactionRepo: transactionRepo,
		unitOfWork:      unitOfWork,
	}
}

var _ usecase.FaceVerificationUsecase = (*FaceVerificationService)(nil)

func (s *FaceVerificationService) VerifyFace(ctx context.Context, profileID uint64, photo []byte, video []byte) (*faceVerificationDto.FaceVerificationResponse, error) {
	s.logger.Info("Starting face verification process")

	if len(photo) == 0 {
		s.logger.Warn("Photo is empty")
		return nil, exception.ErrEmptyPhoto
	}

	if len(video) == 0 {
		s.logger.Warn("Video is empty")
		return nil, exception.ErrEmptyVideo
	}

	service, err := s.serviceRepo.GetByName(ctx, "face_verification")
	if err != nil {
		s.logger.Error("Failed to get face verification service",
			logger.Field{Key: "error", Value: err},
		)
		return nil, exception.ErrServiceNotFound
	}

	if !service.IsAvailable {
		s.logger.Warn("Face verification service is not available")
		return nil, fmt.Errorf("face verification service is currently unavailable")
	}

	serviceCost := uint64(service.CurrentCost)

	if profileID != 0 {
		err = s.unitOfWork.Do(ctx, func(txCtx context.Context) error {
			profile, err := s.profileRepo.GetByID(txCtx, profileID)
			if err != nil {
				return fmt.Errorf("failed to get profile: %w", err)
			}

			if profile.Balance < serviceCost {
				s.logger.Warn("Insufficient balance for face verification",
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
				TransactionType: enum.TransactionTypeFaceVerification,
				Amount:          -int64(serviceCost),
				Notes:           "Face verification service charge",
				CreatedAt:       time.Now(),
			}

			if err := s.transactionRepo.Create(txCtx, transaction); err != nil {
				return fmt.Errorf("failed to create transaction: %w", err)
			}

			s.logger.Info("Wallet deducted for face verification",
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

	result, err := s.client.VerifyFace(ctx, photo, video)
	if err != nil {
		s.logger.Error("Face verification failed",
			logger.Field{Key: "error", Value: err},
		)
		return nil, fmt.Errorf("face verification failed: %w", err)
	}

	if !result.Success {
		s.logger.Warn("Face verification returned failure",
			logger.Field{Key: "reason", Value: result.Reason},
			logger.Field{Key: "message", Value: result.Message},
		)
	} else {
		s.logger.Info("Face verification completed successfully")
	}

	if s.repo != nil && profileID != 0 {
		var rec entity.FaceVerificationRecord
		rec.Success = result.Success
		rec.Reason = result.Reason
		rec.Message = result.Message

		// Store all stats as JSONB
		if result.Stats != nil {
			bs, _ := json.Marshal(result.Stats)
			var stats entity.FaceVerificationStats
			if err := json.Unmarshal(bs, &stats); err == nil {
				rec.Stats = &stats
				// Extract key metrics for quick querying
				if stats.Duration > 0 {
					rec.ProcessingTimeSeconds = stats.Duration
				} else {
					rec.ProcessingTimeSeconds = stats.ProcessingTimeSeconds
				}
			}
		}

		// Calculate highest similarity from results for indexing
		if result.Results != nil {
			var items []map[string]interface{}
			if bs, err := json.Marshal(result.Results); err == nil {
				if err := json.Unmarshal(bs, &items); err == nil {
					var max float64
					for _, it := range items {
						if v, ok := it["insightface_sim"]; ok {
							switch val := v.(type) {
							case float64:
								if val > max {
									max = val
								}
							case float32:
								if float64(val) > max {
									max = float64(val)
								}
							}
						}
					}
					rec.HighestSimilarity = max
				}
			}
		}

		s.logger.Info("Persisting face verification result",
			logger.Field{Key: "rec", Value: rec},
		)

		if err := s.repo.SaveResult(ctx, profileID, &rec); err != nil {
			s.logger.Error("Failed to persist face verification result",
				logger.Field{Key: "error", Value: err},
				logger.Field{Key: "profile_id", Value: profileID},
			)
		}
	}
	result.Results = nil
	result.Messages = nil

	return result, nil
}

func (s *FaceVerificationService) CropImage(ctx context.Context, profileID uint64, image []byte) (*faceVerificationDto.CropImageResponse, error) {
	s.logger.Info("Starting image crop process")

	if len(image) == 0 {
		s.logger.Warn("Image is empty")
		return nil, exception.ErrEmptyImage
	}

	// Get service cost
	service, err := s.serviceRepo.GetByName(ctx, "image_crop")
	if err != nil {
		s.logger.Error("Failed to get image crop service",
			logger.Field{Key: "error", Value: err},
		)
		return nil, exception.ErrServiceNotFound
	}

	if !service.IsAvailable {
		s.logger.Warn("Image crop service is not available")
		return nil, fmt.Errorf("image crop service is currently unavailable")
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
				s.logger.Warn("Insufficient balance for image crop",
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
				TransactionType: enum.TransactionTypeFaceVerification,
				Amount:          -int64(serviceCost),
				Notes:           "Image crop service charge",
				CreatedAt:       time.Now(),
			}

			if err := s.transactionRepo.Create(txCtx, transaction); err != nil {
				return fmt.Errorf("failed to create transaction: %w", err)
			}

			s.logger.Info("Wallet deducted for image crop",
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

	result, err := s.client.CropImage(ctx, image)
	if err != nil {
		s.logger.Error("Image crop failed",
			logger.Field{Key: "error", Value: err},
		)
		return nil, fmt.Errorf("image crop failed: %w", err)
	}

	if !result.Success {
		s.logger.Warn("Image crop returned failure",
			logger.Field{Key: "error", Value: result.Error},
		)
	} else {
		s.logger.Info("Image crop completed successfully")
	}

	return result, nil
}

func (s *FaceVerificationService) HealthCheck(ctx context.Context) (*faceVerificationDto.HealthCheckResponseDTO, error) {
	s.logger.Info("Performing health check on face verification service")

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
