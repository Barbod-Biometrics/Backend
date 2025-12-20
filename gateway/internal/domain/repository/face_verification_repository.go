package repository

import (
	"context"

	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/entity"
)

type FaceVerificationRepository interface {
	SaveResult(ctx context.Context, profileID uint64, result *entity.FaceVerificationRecord) error
	GetResultsByProfileID(ctx context.Context, profileID uint64) ([]*entity.FaceVerificationRecord, error)
}
