package usecase

import (
	"context"

	"github.com/Barbod-Biometrics/Backend/gateway/internal/application/dto/admin"
)

type AdminProfileUsecase interface {
	// ListProfiles returns a paginated list of profiles with filters
	ListProfiles(ctx context.Context, req admin.ListProfilesRequest) (*admin.PaginatedProfilesResponse, error)

	// GetProfileDetail returns full details of a profile for admin review
	GetProfileDetail(ctx context.Context, profileID uint64) (*admin.ProfileDetailResponse, error)

	// ApproveProfile approves a pending profile
	ApproveProfile(ctx context.Context, profileID uint64, req admin.ApproveProfileRequest) error

	// RejectProfile rejects a pending profile with a reason
	RejectProfile(ctx context.Context, profileID uint64, req admin.RejectProfileRequest) error
}
