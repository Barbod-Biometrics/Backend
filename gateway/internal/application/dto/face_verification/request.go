package face_verification

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

	// filters
	Status   string `form:"status"`
	FromDate string `form:"from_date"` // miladi or shamsi? which one?
	ToDate   string `form:"to_date"`

	// sorting
	SortBy    string `form:"sort_by"`
	SortOrder string `form:"sort_order"`
}
