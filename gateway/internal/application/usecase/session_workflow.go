package usecase

import (
	"context"

	"github.com/Barbod-Biometrics/Backend/gateway/internal/application/session"
)

type SessionUsecase interface {
	StartAndRun(ctx context.Context, profileID uint64, workflowConfigID uint64, clientIP string) (string, error)

	ProcessOCR(ctx context.Context, sess *session.Session, imageBytes []byte) (interface{}, error)

	ProcessFaceVerification(ctx context.Context, sess *session.Session, baseImageBytes []byte, videoBytes []byte) (interface{}, error)
}
