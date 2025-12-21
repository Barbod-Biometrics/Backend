package usecase

import (
	"context"

	ocrDto "github.com/Barbod-Biometrics/Backend/gateway/internal/application/dto/ocr"
)

type OCRUsecase interface {
	ExtractText(ctx context.Context, profileID uint64, image []byte) (*ocrDto.OCRResponse, error)

	ExtractTextWithIP(ctx context.Context, profileID uint64, image []byte, clientIP string) (*ocrDto.OCRResponse, error)

	HealthCheck(ctx context.Context) (*ocrDto.HealthCheckResponseDTO, error)
}
