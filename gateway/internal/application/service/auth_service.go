package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/Barbod-Biometrics/Backend/gateway/internal/application/dto/auth"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/application/usecase"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/communication"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/entity"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/repository"
)

type AuthUsecase struct {
	userRepo     repository.UserRepository
	otpService   usecase.OTPService
	smsService   communication.SMSService
	tokenService usecase.TokenService
}

func NewAuthUsecase(userRepo repository.UserRepository, otpService usecase.OTPService, smsService communication.SMSService, tokenService usecase.TokenService) *AuthUsecase {
	return &AuthUsecase{
		userRepo:     userRepo,
		otpService:   otpService,
		smsService:   smsService,
		tokenService: tokenService,
	}
}

func (uc *AuthUsecase) RequestOTP(ctx context.Context, req auth.RequestOTPRequest) error {

	plainOTP, err := uc.otpService.GenerateAndStoreOTP(ctx, req.PhoneNumber)
	if err != nil {
		return fmt.Errorf("failed to generate otp: %w", err)
	}

	message := fmt.Sprintf("Your Barbod Biomentrics code is: %s", plainOTP)
	if err := uc.smsService.Send(ctx, req.PhoneNumber, message); err != nil {
		return fmt.Errorf("failed to send sms: %w", err)
	}

	return nil
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
