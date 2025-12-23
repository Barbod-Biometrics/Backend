package face_verification

type FaceVerificationRequest struct {
	Photo []byte `json:"photo" form:"photo"`
	Video []byte `json:"video" form:"video"`
}

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

type CropImageRequest struct {
	Image []byte `json:"image" form:"image"`
}

type CropImageResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
	Image   []byte `json:"image,omitempty"`
}

type HealthCheckRequest struct {
	// no fields
}

type HealthCheckResponseDTO struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Models  interface{} `json:"models,omitempty"`
	Config  interface{} `json:"config,omitempty"`
}
