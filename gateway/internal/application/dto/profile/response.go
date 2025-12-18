package profile

import (
	"time"
)

// I'm not sure yet if this struct (ServiceUsageStats) should be here or not
type ServiceUsageStats struct {
	ServiceName string `json:"service_name"`
	TotalCount  int    `json:"total_count:"`
	TotalCost   int64  `json:"total_cost"`
}

type ProfileResponse struct {
	ID                 string    `json:"id"`
	Type               string    `json:"type"`
	Name               string    `json:"name"`
	Balance            uint64    `json:"balance"`
	VerificationStatus string    `json:"verification_status"`
	IsActive           bool      `json:"is_active"`
	CreatedAt          time.Time `json:"created_at"`

	PersonDetails   *PersonDetailsResponse   `json:"person_details,omitempty"`
	BusinessDetails *BusinessDetailsResponse `json:"business_details,omitempty"`
}

type PersonDetailsResponse struct {
	FirstName    string                `json:"first_name"`
	LastName     string                `json:"last_name"`
	NationalID   string                `json:"national_id"`
	DOB          *time.Time            `json:"dob"`
	MobileNumber string                `json:"mobile_number"`
	BusinessInfo *BusinessMetaDTO      `json:"business_info,omitempty"`
	LocationInfo *LocationDTO          `json:"location_info,omitempty"`
	Documents    *PersonalDocumentsDTO `json:"documents,omitempty"`
}

type BusinessDetailsResponse struct {
	RepFirstName       string              `json:"rep_first_name"`
	RepLastName        string              `json:"rep_last_name"`
	RepNationalID      string              `json:"rep_national_id"`
	RepDOB             *time.Time          `json:"rep_dob"`
	RepMobileNumber    string              `json:"rep_mobile_number"`
	BusinessNationalID string              `json:"business_national_id"`
	BusinessInfo       *BusinessMetaDTO    `json:"business_info,omitempty"`
	LocationInfo       *LocationDTO        `json:"location_info,omitempty"`
	Signatories        []SignatoryResponse `json:"signatories,omitempty"`
}

type SignatoryResponse struct {
	SignatoryID  uint64                `json:"signatory_id"`
	FirstName    string                `json:"first_name"`
	LastName     string                `json:"last_name"`
	NationalID   string                `json:"national_id"`
	DOB          *time.Time            `json:"dob"`
	MobileNumber string                `json:"mobile_number"`
	Documents    *PersonalDocumentsDTO `json:"documents,omitempty"`
}

type UploadUrlResponse struct {
	UploadUrl string `json:"upload_url"`
	FileKey   string `json:"file_key"`
	ExpiresAt string `json:"expires_at"`
}

type UsageSummaryResponse struct {
	TotalSpend       int64               `json:"total_spend"`
	ServiceBreakdown []ServiceUsageStats `json:"service_breakdown"`
}
