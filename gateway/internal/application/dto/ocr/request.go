package ocr

type OCRRequest struct {
	Image string `json:"image" binding:"required"`
}
