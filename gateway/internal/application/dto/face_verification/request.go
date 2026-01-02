package face_verification

import "github.com/Barbod-Biometrics/Backend/gateway/internal/presentation/types"

type FaceVerificationRequest struct {
	Photo []byte `json:"photo" form:"photo"`
	Video []byte `json:"video" form:"video"`
}

type CropImageRequest struct {
	Image []byte `json:"image" form:"image"`
}

type HealthCheckRequest struct {
	// no fields
}

type GetFaceReportRequest struct {
	ProfileID uint64 `form:"profile_id" binding:"required"`
	Page      int    `form:"page,default=1"`
	Limit     int    `form:"limit,default=20"`

	// filters
	Status   string           `form:"status"`
	FromDate types.JalaliDate `form:"from_date"`
	ToDate   types.JalaliDate `form:"to_date"`

	// sorting
	SortBy    string `form:"sort_by"`
	SortOrder string `form:"sort_order"`
}
