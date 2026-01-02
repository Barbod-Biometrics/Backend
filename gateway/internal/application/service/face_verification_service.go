package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/Barbod-Biometrics/Backend/gateway/bootstrap"
	faceVerificationDto "github.com/Barbod-Biometrics/Backend/gateway/internal/application/dto/face_verification"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/application/usecase"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/communication"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/entity"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/enum"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/exception"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/logger"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/repository"
	face_verification "github.com/Barbod-Biometrics/Backend/gateway/internal/infrastructure/face_verificaiton"
)

type FaceVerificationService struct {
	client              *face_verification.FaceVerificationClient
	logger              logger.Logger
	repo                repository.FaceVerificationRepository
	serviceRepo         repository.ServiceRepository
	profileRepo         repository.ProfileRepository
	userRepo            repository.UserRepository
	transactionRepo     repository.TransactionRepository
	emailService        communication.EmailService
	unitOfWork          repository.UnitOfWork
	trialService        *TrialService
	lowBalanceThreshold uint64
}

func NewFaceVerificationService(
	cli *face_verification.FaceVerificationClient,
	logger logger.Logger,
	repo repository.FaceVerificationRepository,
	serviceRepo repository.ServiceRepository,
	profileRepo repository.ProfileRepository,
	userRepo repository.UserRepository,
	transactionRepo repository.TransactionRepository,
	emailService communication.EmailService,
	unitOfWork repository.UnitOfWork,
	trialService *TrialService,
	cfg *bootstrap.Config,
) *FaceVerificationService {
	return &FaceVerificationService{
		client:              cli,
		logger:              logger,
		repo:                repo,
		serviceRepo:         serviceRepo,
		profileRepo:         profileRepo,
		userRepo:            userRepo,
		transactionRepo:     transactionRepo,
		emailService:        emailService,
		unitOfWork:          unitOfWork,
		trialService:        trialService,
		lowBalanceThreshold: cfg.Env.Wallet.LowBalanceThreshold,
	}
}

var _ usecase.FaceVerificationUsecase = (*FaceVerificationService)(nil)

func (s *FaceVerificationService) VerifyFace(ctx context.Context, profileID uint64, photo []byte, video []byte) (*faceVerificationDto.FaceVerificationResponse, error) {
	return s.VerifyFaceWithIP(ctx, profileID, photo, video, "")
}

func (s *FaceVerificationService) VerifyFaceWithIP(ctx context.Context, profileID uint64, photo []byte, video []byte, clientIP string) (*faceVerificationDto.FaceVerificationResponse, error) {
	s.logger.Info("Starting face verification process")

	if len(photo) == 0 {
		s.logger.Warn("Photo is empty")
		return nil, exception.ErrEmptyPhoto
	}

	if len(video) == 0 {
		s.logger.Warn("Video is empty")
		return nil, exception.ErrEmptyVideo
	}

	// Check trial attempts for non-authenticated users
	if profileID == 0 && clientIP != "" && s.trialService != nil {
		remainingAttempts, ttl, err := s.trialService.CheckAndDecrementTrial(ctx, clientIP, "face_verification")
		if err != nil {
			if errors.Is(err, exception.ErrTrialExceeded) {
				resp := &faceVerificationDto.FaceVerificationResponse{
					Success:           false,
					Reason:            "trial_exceeded",
					Message:           "Trial limit exceeded",
					RemainingAttempts: remainingAttempts,
				}
				if ttl > 0 {
					resp.RechargeInSeconds = int(ttl.Seconds())
				}
				s.logger.Warn("Trial limit exceeded",
					logger.Field{Key: "ip", Value: clientIP},
					logger.Field{Key: "remaining_attempts", Value: remainingAttempts},
					logger.Field{Key: "recharge_in_seconds", Value: resp.RechargeInSeconds},
				)
				return resp, exception.ErrTrialExceeded
			}

			s.logger.Warn("Trial limit check failed",
				logger.Field{Key: "ip", Value: clientIP},
				logger.Field{Key: "error", Value: err},
			)
			return nil, err
		}
		s.logger.Info("Trial attempt recorded",
			logger.Field{Key: "ip", Value: clientIP},
			logger.Field{Key: "remaining_attempts", Value: remainingAttempts},
		)
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
		return nil, exception.NewInternalError("ERR_SERVICE_UNAVAILABLE", "face verification service is currently unavailable", nil)
	}

	serviceCost := uint64(service.CurrentCost)

	if profileID != 0 {
		err = s.unitOfWork.Do(ctx, func(txCtx context.Context) error {
			profile, err := s.profileRepo.GetByID(txCtx, profileID)
			if err != nil {
				return exception.NewInternalError("ERR_PROFILE_FETCH", "failed to get profile", err)
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
				return exception.NewInternalError("ERR_PROFILE_UPDATE", "failed to update profile balance", err)
			}

			transaction := &entity.Transaction{
				ProfileID:       profileID,
				TransactionType: enum.TransactionTypeWithdrawal,
				Amount:          -int64(serviceCost),
				Notes:           "Face verification service charge",
				CreatedAt:       time.Now(),
			}

			if err := s.transactionRepo.Create(txCtx, transaction); err != nil {
				return exception.NewInternalError("ERR_CREATE_TRANSACTION", "failed to create transaction", err)
			}

			s.logger.Info("Wallet deducted for face verification",
				logger.Field{Key: "profile_id", Value: profileID},
				logger.Field{Key: "amount", Value: serviceCost},
				logger.Field{Key: "new_balance", Value: profile.Balance},
			)

			s.checkLowBalance(profileID, profile.Balance)

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
		return nil, exception.NewInternalError("ERR_SERVICE_UNAVAILABLE", "image crop service is currently unavailable", nil)
	}

	serviceCost := uint64(service.CurrentCost)

	// Check wallet balance and deduct cost
	if profileID != 0 {
		err = s.unitOfWork.Do(ctx, func(txCtx context.Context) error {
			profile, err := s.profileRepo.GetByID(txCtx, profileID)
			if err != nil {
				return exception.NewInternalError("ERR_PROFILE_FETCH", "failed to get profile", err)
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
				return exception.NewInternalError("ERR_PROFILE_UPDATE", "failed to update profile balance", err)
			}

			transaction := &entity.Transaction{
				ProfileID:       profileID,
				TransactionType: enum.TransactionTypeWithdrawal,
				Amount:          -int64(serviceCost),
				Notes:           "Image crop service charge",
				CreatedAt:       time.Now(),
			}

			if err := s.transactionRepo.Create(txCtx, transaction); err != nil {
				return exception.NewInternalError("ERR_CREATE_TRANSACTION", "failed to create transaction", err)
			}

			s.logger.Info("Wallet deducted for image crop",
				logger.Field{Key: "profile_id", Value: profileID},
				logger.Field{Key: "amount", Value: serviceCost},
				logger.Field{Key: "new_balance", Value: profile.Balance},
			)

			s.checkLowBalance(profileID, profile.Balance)

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
		return nil, exception.NewInternalError("ERR_IMAGE_CROP", "image crop failed", err)
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
		return nil, exception.NewInternalError("ERR_FACE_HEALTH_CHECK", "health check failed", err)
	}

	s.logger.Info("Health check status",
		logger.Field{Key: "status", Value: result.Status},
	)
	return result, nil
}

func (s *FaceVerificationService) checkLowBalance(profileID uint64, balance uint64) {
	if balance >= s.lowBalanceThreshold {
		return
	}

	go func() {
		ctx := context.Background()
		profile, err := s.profileRepo.GetByID(ctx, profileID)
		if err != nil || profile == nil {
			return
		}

		user, err := s.userRepo.GetByID(ctx, profile.UserID)
		if err != nil || user == nil || user.Email == nil || *user.Email == "" {
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
			"Balance":     balance,
		}

		_ = s.emailService.SendWithTemplate(ctx, *user.Email, "هشدار موجودی کم", "low_balance.html", data)
	}()
}
