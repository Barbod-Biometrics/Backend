package service

import (
	"context"
	"encoding/base64"
	"errors"
	"time"

	faceVerificationDto "github.com/Barbod-Biometrics/Backend/gateway/internal/application/dto/face_verification"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/application/session"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/application/usecase"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/entity"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/logger"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/repository"
	"github.com/Barbod-Biometrics/Backend/gateway/pkg/storage"
)

var _ usecase.SessionUsecase = (*SessionInteractor)(nil)

type SessionInteractor struct {
	SessionMgr    *session.SessionManager
	OCR           usecase.OCRUsecase
	FV            usecase.FaceVerificationUsecase
	ConfigRepo    repository.WorkflowConfigRepository
	StorageClient *storage.MinioClient
	Logger        logger.Logger
}

func NewSessionInteractor(mgr *session.SessionManager, ocr usecase.OCRUsecase, fv usecase.FaceVerificationUsecase, configRepo repository.WorkflowConfigRepository, storageClient *storage.MinioClient, l logger.Logger) *SessionInteractor {
	return &SessionInteractor{SessionMgr: mgr, OCR: ocr, FV: fv, ConfigRepo: configRepo, StorageClient: storageClient, Logger: l}
}

func (s *SessionInteractor) StartAndRun(ctx context.Context, profileID uint64, workflowConfigID uint64, clientIP string) (string, error) {
	sess, err := s.SessionMgr.StartSession(profileID, clientIP)
	if err != nil {
		return "", err
	}

	if workflowConfigID > 0 && s.ConfigRepo != nil {
		cfg, err := s.ConfigRepo.GetByID(ctx, workflowConfigID)
		if err == nil {
			sess.WorkflowConfigID = cfg.ID
			sess.Lock()
			if sess.Metadata == nil {
				sess.Metadata = make(map[string]interface{})
			}
			sess.Metadata["instruction"] = cfg.Instruction
			sess.Metadata["liveness_sentence"] = cfg.LivenessSentence
			sess.Metadata["workflow_name"] = cfg.Name
			sess.Unlock()
			s.SessionMgr.UpdateSession(sess)
		}
	}

	sess.State = entity.StatePendingOCR
	_ = s.SessionMgr.UpdateSession(sess)

	return sess.ID, nil
}

func (s *SessionInteractor) ProcessOCR(ctx context.Context, sess *entity.Session, imageBytes []byte) (interface{}, error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()

	ocrRes, ocrErr := s.OCR.ExtractText(ctx, sess.ProfileID, imageBytes)
	sess.Lock()
	defer sess.Unlock()

	if ocrErr != nil {
		sess.State = entity.StateOCRFailed
		sess.Errors = append(sess.Errors, ocrErr.Error())
		sess.OCRResult = nil
		s.SessionMgr.UpdateSession(sess)
		return nil, ocrErr
	}

	sess.OCRResult = ocrRes

	// Check if OCR acceptance is required for this workflow
	var ocrAcceptanceRequired bool
	if sess.WorkflowConfigID > 0 && s.ConfigRepo != nil {
		cfg, err := s.ConfigRepo.GetByID(ctx, sess.WorkflowConfigID)
		if err == nil {
			ocrAcceptanceRequired = cfg.OCRAcceptanceRequired
		}
	}

	if ocrAcceptanceRequired {
		sess.State = entity.StateOCRPendingAcceptance
	} else {
		sess.State = entity.StateOCRSuccess
	}

	s.SessionMgr.UpdateSession(sess)

	go func() {
		s.Logger.Info("Auto-cropping OCR image for face verification",
			logger.Field{Key: "session_id", Value: sess.ID},
		)
		cropRes, cropErr := s.FV.CropImage(context.Background(), sess.ProfileID, imageBytes)
		if cropErr != nil {
			s.Logger.Error("Failed to auto-crop image",
				logger.Field{Key: "session_id", Value: sess.ID},
				logger.Field{Key: "error", Value: cropErr},
			)
			return
		}

		sess.Lock()
		if sess.Metadata == nil {
			sess.Metadata = make(map[string]interface{})
		}
		sess.Metadata["auto_cropped_image"] = cropRes
		sess.Unlock()

		s.SessionMgr.UpdateSession(sess)
		s.Logger.Info("OCR image auto-cropped and stored",
			logger.Field{Key: "session_id", Value: sess.ID},
		)
	}()

	return ocrRes, nil
}

func (s *SessionInteractor) ProcessFaceVerification(ctx context.Context, sess *entity.Session, baseImageBytes []byte, videoBytes []byte) (interface{}, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	sess.Lock()
	var finalBaseImageBytes []byte
	if autoCropped, ok := sess.Metadata["auto_cropped_image"]; ok {

		s.Logger.Info("Found auto-cropped image in metadata",
			logger.Field{Key: "session_id", Value: sess.ID},
		)

		if cropResMap, ok := autoCropped.(map[string]interface{}); ok {
			if imageBase64, ok := cropResMap["image"].(string); ok && len(imageBase64) > 0 {
				decodedImage, decodeErr := base64.StdEncoding.DecodeString(imageBase64)
				if decodeErr == nil && len(decodedImage) > 0 {
					finalBaseImageBytes = decodedImage
					s.Logger.Info("Using auto-cropped image from OCR",
						logger.Field{Key: "session_id", Value: sess.ID},
						logger.Field{Key: "cropped_image_size", Value: len(finalBaseImageBytes)},
					)
				} else {
					s.Logger.Error("Failed to decode auto-cropped image",
						logger.Field{Key: "session_id", Value: sess.ID},
						logger.Field{Key: "error", Value: decodeErr},
					)
				}
			}
		} else if cropRes, ok := autoCropped.(*faceVerificationDto.CropImageResponse); ok {
			// Handle if it's stored as the struct
			if len(cropRes.Image) > 0 {
				finalBaseImageBytes = cropRes.Image
				s.Logger.Info("Using auto-cropped image from OCR (struct)",
					logger.Field{Key: "session_id", Value: sess.ID},
					logger.Field{Key: "cropped_image_size", Value: len(finalBaseImageBytes)},
				)
			}
		}
	}
	sess.Unlock()

	if len(finalBaseImageBytes) == 0 {
		finalBaseImageBytes = baseImageBytes
	}

	s.Logger.Info("Processing face verification",
		logger.Field{Key: "session_id", Value: sess.ID},
		logger.Field{Key: "video_size", Value: len(videoBytes)},
	)

	fvRes, fvErr := s.FV.VerifyFace(ctx, sess.ProfileID, finalBaseImageBytes, videoBytes)
	s.Logger.Info("Face verification call completed", logger.Field{Key: "session_id", Value: sess.ID}, logger.Field{Key: "has_error", Value: fvErr != nil})

	sess.Lock()
	defer sess.Unlock()

	if fvErr != nil {
		s.Logger.Info("Face verification error, restoring to OCRSuccess state", logger.Field{Key: "session_id", Value: sess.ID})
		sess.State = entity.StateOCRSuccess
		sess.Errors = append(sess.Errors, fvErr.Error())
		sess.FaceResult = nil
		s.SessionMgr.UpdateSession(sess)
		s.Logger.Info("Session updated after error", logger.Field{Key: "session_id", Value: sess.ID})
		return nil, fvErr
	}

	s.Logger.Info("Setting face verification result", logger.Field{Key: "session_id", Value: sess.ID})
	sess.FaceResult = fvRes
	sess.State = entity.StateLivenessSuccess
	s.SessionMgr.UpdateSession(sess)

	s.Logger.Info("Completing session asynchronously", logger.Field{Key: "session_id", Value: sess.ID})
	go func() {
		s.Logger.Info("Async: Completing session", logger.Field{Key: "session_id", Value: sess.ID})
		s.SessionMgr.CompleteSession(sess)
		s.Logger.Info("Async: Session completed", logger.Field{Key: "session_id", Value: sess.ID})
	}()

	s.Logger.Info("Face verification completed",
		logger.Field{Key: "session_id", Value: sess.ID},
		logger.Field{Key: "result", Value: fvRes},
	)

	return fvRes, nil
}

// AcceptOCR moves session from OCR pending acceptance to OCR success state
func (s *SessionInteractor) AcceptOCR(ctx context.Context, sess *entity.Session) error {
	sess.Lock()
	defer sess.Unlock()

	if sess.State != entity.StateOCRPendingAcceptance {
		return errors.New("session is not in OCR pending acceptance state")
	}

	sess.State = entity.StateOCRSuccess
	s.SessionMgr.UpdateSession(sess)

	s.Logger.Info("OCR result accepted", logger.Field{Key: "session_id", Value: sess.ID})
	return nil
}
