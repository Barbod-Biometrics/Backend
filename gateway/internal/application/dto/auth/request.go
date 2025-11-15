package auth

type RequestOTPRequest struct {
	PhoneNumber string `json:"phone_number" validate:"required"`
}

type VerifyOTPRequest struct {
	PhoneNumber string `json:"phone_number" validate:"required"`
	OTP         string `json:"otp" validate:"required"`
}
