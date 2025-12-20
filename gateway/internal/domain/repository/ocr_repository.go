package repository

import (
	"context"

	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/entity"
)

type OCRRepository interface {
	SaveResult(ctx context.Context, profileID uint64, result *entity.OCRRecord) error
	GetResultsByProfileID(ctx context.Context, profileID uint64) ([]*entity.OCRRecord, error)
}
