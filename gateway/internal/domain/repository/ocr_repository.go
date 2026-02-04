package repository

import (
	"context"
	"time"

	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/entity"
)

type OCRReportFilter struct {
	ProfileID uint64

	// optional filters
	Status   *string
	FromDate *time.Time
	ToDate   *time.Time

	// pagination
	Page  int
	Limit int

	// sorting
	SortBy    string // default -> date
	SortOrder string // "asc" or "desc"
}

type OCRRepository interface {
	SaveResult(ctx context.Context, profileID uint64, result *entity.OCRRecord) error
	GetResultsByProfileID(ctx context.Context, profileID uint64) ([]*entity.OCRRecord, error)
	GetReports(ctx context.Context, filter OCRReportFilter) ([]*entity.OCRModel, int64, error)
	ApproveOCRResult(ctx context.Context, ocrID uint64, profileID uint64) error
}
