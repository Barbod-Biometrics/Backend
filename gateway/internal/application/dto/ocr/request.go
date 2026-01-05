package ocr

type OCRRequest struct {
	Image string `json:"image" binding:"required"`
}

type ApproveOCRRequest struct {
	ID        uint64 `json:"id" binding:"required"`
	ProfileID uint64 `json:"profile_id" binding:"required"`
}
