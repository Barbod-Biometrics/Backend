package usecase

import "context"

type OTPService interface {
	GenerateAndStoreOTP(ctx context.Context, phoneNumber string) (plainOTP string, err error)
	VerifyOTP(ctx context.Context, phoneNumber string, otp string) error
}
