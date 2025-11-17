package service

import (
	"context"
	"errors"

	"github.com/Barbod-Biometrics/Backend/gateway/internal/application/dto/auth"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/application/usecase"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/entity"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/repository"
)

type AuthUsecase struct {
	userRepo     repository.UserRepository
	otpService   usecase.OTPService
	tokenService usecase.TokenService
}

func NewAuthUsecase(userRepo repository.UserRepository, otpService usecase.OTPService, tokenService usecase.TokenService) *AuthUsecase {
	return &AuthUsecase{
		userRepo:     userRepo,
		otpService:   otpService,
		tokenService: tokenService,
	}
}

func (uc *AuthUsecase) RequestOTP(ctx context.Context, req auth.RequestOTPRequest) error {

	// Note: There can be a phone number validation here

	return uc.otpService.GenerateOTP(ctx, req.PhoneNumber)
}

func (uc *AuthUsecase) VerifyOTP(ctx context.Context, req auth.VerifyOTPRequest) (*auth.UserInfoResponse, error) {

	if err := uc.otpService.VerifyOTP(ctx, req.PhoneNumber, req.OTP); err != nil {
		return nil, errors.New("invalid OTP")
	}

	user, err := uc.userRepo.GetByPhoneNumber(ctx, req.PhoneNumber)
	if err != nil {
		return nil, errors.New("error retrieving user by phone number from database")
	}

	if user == nil {

		newUser := &entity.User{
			PhoneNumber: req.PhoneNumber,
		}

		if err := uc.userRepo.CreateUser(ctx, newUser); err != nil {
			return nil, errors.New("failed to create user")
		}

		user = newUser
	}

	accessToken, refreshToken, err := uc.tokenService.GenerateTokens(ctx, user.UserID)

	if err != nil {
		return nil, errors.New("failed to generate tokens")
	}

	return &auth.UserInfoResponse{
		AccessToken:   accessToken,
		RefereshToken: refreshToken,
		PhoneNumber:   req.PhoneNumber,
	}, nil

}
