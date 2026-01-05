package usecase

import (
	"context"

	faceVerificationDto "github.com/Barbod-Biometrics/Backend/gateway/internal/application/dto/face_verification"
)

type FaceVerificationUsecase interface {
	VerifyFace(ctx context.Context, profileID uint64, photo []byte, video []byte) (*faceVerificationDto.FaceVerificationResponse, error)

	VerifyFaceWithIP(ctx context.Context, profileID uint64, photo []byte, video []byte, clientIP string) (*faceVerificationDto.FaceVerificationResponse, error)

	CropImage(ctx context.Context, profileID uint64, image []byte) (*faceVerificationDto.CropImageResponse, error)

	HealthCheck(ctx context.Context) (*faceVerificationDto.HealthCheckResponseDTO, error)
	GetReports(ctx context.Context, req faceVerificationDto.GetFaceReportRequest) (*faceVerificationDto.FaceReportResponse, error)
}
