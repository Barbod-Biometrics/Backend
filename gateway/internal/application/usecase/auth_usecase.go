package usecase

import (
	"context"

	"github.com/Barbod-Biometrics/Backend/gateway/internal/application/dto/auth"
)

type AuthUsecase interface {
	RequestOTP(ctx context.Context, req auth.RequestOTPRequest) error
	VerifyOTP(ctx context.Context, req auth.VerifyOTPRequest) (*auth.UserInfoResponse, error)
}
