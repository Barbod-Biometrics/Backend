package face_verification

type FaceVerificationResponse struct {
	Success  bool        `json:"success"`
	Reason   string      `json:"reason"`
	Message  string      `json:"message"`
	Stats    interface{} `json:"stats,omitempty"`
	Results  interface{} `json:"results,omitempty"`
	Messages interface{} `json:"messages,omitempty"`

	RemainingAttempts int `json:"remaining_attempts,omitempty"`
	RechargeInSeconds int `json:"recharge_in_seconds,omitempty"`
}

type CropImageResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
	Image   []byte `json:"image,omitempty"`
}

type HealthCheckResponseDTO struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Models  interface{} `json:"models,omitempty"`
	Config  interface{} `json:"config,omitempty"`
}

type FaceReportItem struct {
	ID       uint64  `json:"id"`
	Date     string  `json:"date"`
	Status   string  `json:"status"`
	Duration float64 `json:"duration"`
	Message  string  `json:"message"`
	Reason   string  `json:"reason"`
	Rate     float32 `json:"rate"`
}

type FaceReportResponse struct {
	Items      []FaceReportItem `json:"items"`
	TotalCount int64            `json:"total_count"`
	Page       int              `json:"page"`
	TotalPages int              `json:"total_pages"`
}
