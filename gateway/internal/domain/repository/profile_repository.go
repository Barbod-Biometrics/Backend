package repository

import (
	"context"

	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/entity"
)

// ProfileFilter contains filter options for listing profiles
type ProfileFilter struct {
	Status      *entity.VerificationStatus
	ProfileType *entity.ProfileType
	SearchQuery *string // searches in profile name, national ID, phone number
}

// ProfileSort defines sorting options
type ProfileSort struct {
	Field string // "created_at", "profile_name", "verification_status"
	Order string // "asc" or "desc"
}

// Pagination defines pagination options
type Pagination struct {
	Page     int
	PageSize int
}

// PaginatedResult contains paginated results
type PaginatedResult struct {
	Profiles   []*entity.Profile
	TotalCount int64
	Page       int
	PageSize   int
	TotalPages int
}

type ProfileRepository interface {
	Create(ctx context.Context, profile *entity.Profile) error
	Update(ctx context.Context, profile *entity.Profile) error
	GetByID(ctx context.Context, profileID uint64) (*entity.Profile, error)
	GetByUserID(ctx context.Context, userID uint64) ([]*entity.Profile, error)
	GetPersonalProfileByUserID(ctx context.Context, userID uint64) (*entity.Profile, error)

	// Admin methods
	ListWithFilters(ctx context.Context, filter ProfileFilter, sort ProfileSort, pagination Pagination) (*PaginatedResult, error)
}
