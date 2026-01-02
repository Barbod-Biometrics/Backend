package ocr

import "encoding/json"

type OCRResponse struct {
	Success bool                   `json:"success"`
	Message string                 `json:"message,omitempty"`
	Stats   map[string]interface{} `json:"stats,omitempty"`

	NationalID     string `json:"شماره_ملی,omitempty"`
	FirstName      string `json:"نام,omitempty"`
	LastName       string `json:"نام_خانوادگی,omitempty"`
	FatherName     string `json:"نام_پدر,omitempty"`
	BirthDate      string `json:"تاریخ_تولد,omitempty"`
	ExpirationDate string `json:"پایان_اعتبار,omitempty"`

	RemainingAttempts int `json:"remaining_attempts,omitempty"`
	RechargeInSeconds int `json:"recharge_in_seconds,omitempty"`
}

type HealthCheckResponseDTO struct {
	Status               string  `json:"status"`
	ProcessingTimeSecond float64 `json:"processing_time_seconds,omitempty"`
}

type OCRReportItem struct {
	ID      uint64          `json:"id"`
	Date    string          `json:"date"`   // jalali string
	Status  string          `json:"status"` // "Failed" or "Success"
	Message string          `json:"message"`
	Stats   json.RawMessage `json:"stats"`
}

type OCRReportResponse struct {
	Items      []OCRReportItem `json:"items"`
	TotalCount int             `json:"total_count"`
	Page       int             `json:"page"`
	TotalPages int             `json:"total_pages"`
}
