package usecase

import (
	"context"

	"github.com/Barbod-Biometrics/Backend/gateway/internal/application/dto/profile"
)

type ProfileUsecase interface {
	CreateDraft(ctx context.Context, userID uint64, req profile.CreateProfileRequest) (*profile.ProfileResponse, error)
	UpdateDraft(ctx context.Context, userID uint64, profileID uint64, req profile.UpdateProfileRequest) (*profile.ProfileResponse, error)
	SaveDocument(ctx context.Context, userID uint64, profileID uint64, req profile.SaveDocumentRequest) error
	SubmitProfile(ctx context.Context, userID uint64, profileID uint64) (*profile.ProfileResponse, error)
	GetByID(ctx context.Context, userID uint64, profileID uint64) (*profile.ProfileResponse, error)
	GetUploadUrl(ctx context.Context, userID uint64, req profile.GetUploadUrlRequest) (*profile.UploadUrlResponse, error)
}
