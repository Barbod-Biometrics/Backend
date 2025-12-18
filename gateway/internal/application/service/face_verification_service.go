package service

import (
	"context"
	"encoding/json"
	"fmt"

	faceVerificationDto "github.com/Barbod-Biometrics/Backend/gateway/internal/application/dto/face_verification"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/application/usecase"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/entity"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/exception"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/logger"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/repository"
	face_verification "github.com/Barbod-Biometrics/Backend/gateway/internal/infrastructure/face_verificaiton"
)

type FaceVerificationService struct {
	client *face_verification.FaceVerificationClient
	logger logger.Logger
	repo   repository.FaceVerificationRepository
}

func NewFaceVerificationService(cli *face_verification.FaceVerificationClient, logger logger.Logger, repo repository.FaceVerificationRepository) *FaceVerificationService {
	return &FaceVerificationService{
		client: cli,
		logger: logger,
		repo:   repo,
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
		if result.Stats != nil {
			bs, _ := json.Marshal(result.Stats)
			var stats entity.FaceVerificationStats
			if err := json.Unmarshal(bs, &stats); err == nil {
				rec.Stats = &stats
				if stats.VerifiedFrames > 0 {
					rec.MatchedFrames = stats.VerifiedFrames
				} else {
					rec.MatchedFrames = stats.MatchedFrames
				}
				if stats.RealFrames > 0 {
					rec.FacesDetected = stats.RealFrames
				} else {
					rec.FacesDetected = stats.FacesDetected
				}
				if stats.Duration > 0 {
					rec.ProcessingTimeSeconds = stats.Duration
				} else {
					rec.ProcessingTimeSeconds = stats.ProcessingTimeSeconds
				}

				rec.RealFrames = stats.RealFrames
				rec.RealRate = stats.RealRate
				rec.SpoofRate = stats.SpoofRate
				rec.SpoofedFrames = stats.SpoofedFrames
				rec.TotalFrames = stats.TotalFrames
				rec.TwoFacesFrames = stats.TwoFacesFrames
				rec.TwoFacesRate = stats.TwoFacesRate
				rec.VerificationRate = stats.VerificationRate
				rec.VerifiedFrames = stats.VerifiedFrames
			}
		}

		if result.Results != nil {
			var items []map[string]interface{}
			if bs, err := json.Marshal(result.Results); err == nil {
				if err := json.Unmarshal(bs, &items); err == nil {
					var sum float64
					var max float64
					var cnt int
					for _, it := range items {
						if v, ok := it["insightface_sim"]; ok {
							switch val := v.(type) {
							case float64:
								sum += val
								if val > max {
									max = val
								}
								cnt++
							case float32:
								fv := float64(val)
								sum += fv
								if fv > max {
									max = fv
								}
								cnt++
							case int:
								fv := float64(val)
								sum += fv
								if fv > max {
									max = fv
								}
								cnt++
							}
						}
					}
					if cnt > 0 {
						rec.AverageSimilarity = sum / float64(cnt)
						rec.HighestSimilarity = max
					}
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

func (s *FaceVerificationService) CropImage(ctx context.Context, image []byte) (*faceVerificationDto.CropImageResponse, error) {
	s.logger.Info("Starting image crop process")

	if len(image) == 0 {
		s.logger.Warn("Image is empty")
		return nil, exception.ErrEmptyImage
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
