package usecase

import (
	"context"

	ocrDto "github.com/Barbod-Biometrics/Backend/gateway/internal/application/dto/ocr"
)

type OCRUsecase interface {
	ExtractText(ctx context.Context, profileID uint64, image []byte) (*ocrDto.OCRResponse, error)

	ExtractTextWithIP(ctx context.Context, profileID uint64, image []byte, clientIP string) (*ocrDto.OCRResponse, error)

	HealthCheck(ctx context.Context) (*ocrDto.HealthCheckResponseDTO, error)

	GetReports(ctx context.Context, req ocrDto.GetOCRReportRequest, userID uint64) (*ocrDto.OCRReportResponse, error)
	ApproveResult(ctx context.Context, userID, ocrID uint64, profileID uint64) error
}
