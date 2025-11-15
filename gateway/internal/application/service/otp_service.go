package service

import "context"

type OTPService interface {
	GenerateOTP(ctx context.Context, phoneNumber string) error
	VerifyOTP(ctx context.Context, phoneNumber string, otp string) error
}
