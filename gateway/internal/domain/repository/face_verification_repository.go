package repository

import (
	"context"
	"time"

	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/entity"
)

type FaceReportFilter struct {
	ProfileID uint64

	// filters
	Status   *string
	FromDate *time.Time
	ToDate   *time.Time

	// pagination
	Page  int
	Limit int

	// sorting
	SortBy    string
	SortOrder string
}

type FaceVerificationRepository interface {
	SaveResult(ctx context.Context, profileID uint64, result *entity.FaceVerificationRecord) error
	GetResultsByProfileID(ctx context.Context, profileID uint64) ([]*entity.FaceVerificationRecord, error)
	GetReports(ctx context.Context, filter FaceReportFilter) ([]*entity.FaceVerificationModel, int64, error)
}
