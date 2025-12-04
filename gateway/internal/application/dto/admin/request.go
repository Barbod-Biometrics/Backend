package admin

// ListProfilesRequest contains query parameters for listing profiles
type ListProfilesRequest struct {
	Page        int    `form:"page" binding:"omitempty,min=1"`
	PageSize    int    `form:"page_size" binding:"omitempty,min=1,max=100"`
	Status      string `form:"status" binding:"omitempty,oneof=draft pending verified rejected"`
	ProfileType string `form:"profile_type" binding:"omitempty,oneof=personal business"`
	Search      string `form:"search" binding:"omitempty,max=100"`
	SortBy      string `form:"sort_by" binding:"omitempty,oneof=created_at profile_name verification_status profile_type"`
	SortOrder   string `form:"sort_order" binding:"omitempty,oneof=asc desc"`
}

// ApproveProfileRequest is the request body for approving a profile
type ApproveProfileRequest struct {
	Note string `json:"note" binding:"omitempty,max=500"`
}

// RejectProfileRequest is the request body for rejecting a profile
type RejectProfileRequest struct {
	Reason string `json:"reason" binding:"required,min=10,max=500"`
}
