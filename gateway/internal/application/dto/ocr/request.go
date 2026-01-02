package ocr

import "github.com/Barbod-Biometrics/Backend/gateway/internal/presentation/types"

type OCRRequest struct {
	Image string `json:"image" binding:"required"`
}

type GetOCRReportRequest struct {
	ProfileID uint64 `form:"profile_id" binding:"required"`
	Page      int    `form:"page,default=1"`
	Limit     int    `form:"limit,default=20"`

	FromDate types.JalaliDate `form:"from_date"`
	ToDate   types.JalaliDate `form:"to_date"`

	Status    string `form:"status"`
	SortBy    string `form:"sort_by"`
	SortOrder string `form:"sort_order"`
}
