package admin

import (
	"time"

	"github.com/Barbod-Biometrics/Backend/gateway/internal/application/dto/profile"
)

// ProfileListItem is a summary of a profile for list views
type ProfileListItem struct {
	ID                 uint64    `json:"id"`
	UserID             uint64    `json:"user_id"`
	ProfileType        string    `json:"profile_type"`
	ProfileName        string    `json:"profile_name"`
	VerificationStatus string    `json:"verification_status"`
	IsActive           bool      `json:"is_active"`
	CreatedAt          time.Time `json:"created_at"`

	// Summary fields for quick view
	OwnerName    string `json:"owner_name,omitempty"`     // First + Last name or Rep name
	NationalID   string `json:"national_id,omitempty"`    // Person or Rep national ID
	MobileNumber string `json:"mobile_number,omitempty"`  // Person or Rep mobile
}

// PaginatedProfilesResponse is the response for listing profiles
type PaginatedProfilesResponse struct {
	Profiles   []ProfileListItem `json:"profiles"`
	TotalCount int64             `json:"total_count"`
	Page       int               `json:"page"`
	PageSize   int               `json:"page_size"`
	TotalPages int               `json:"total_pages"`
}

// ProfileDetailResponse contains full profile details for admin view
type ProfileDetailResponse struct {
	ID                 uint64    `json:"id"`
	UserID             uint64    `json:"user_id"`
	ProfileType        string    `json:"profile_type"`
	ProfileName        string    `json:"profile_name"`
	Balance            uint64    `json:"balance"`
	VerificationStatus string    `json:"verification_status"`
	IsActive           bool      `json:"is_active"`
	CreatedAt          time.Time `json:"created_at"`

	PersonDetails   *profile.PersonDetailsResponse   `json:"person_details,omitempty"`
	BusinessDetails *profile.BusinessDetailsResponse `json:"business_details,omitempty"`
}

// ActionResponse is a generic response for admin actions
type ActionResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}
