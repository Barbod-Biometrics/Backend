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
		if errors.Is(err, ErrOTPAlreadyExists) {
			return errors.New("OTP already sent. Please wait before requesting a new one")
		}
		return fmt.Errorf("failed to generate otp: %w", err)
	}

	message := fmt.Sprintf("Your Barbod Biomentrics code is: %s", plainOTP)
	if err := uc.smsService.Send(ctx, req.PhoneNumber, message); err != nil {
		// if OTP fials we have to delete it from the cache later
		return fmt.Errorf("failed to send sms: %w", err)
	}

	return nil
}

func (uc *AuthUsecase) VerifyOTP(ctx context.Context, req auth.VerifyOTPRequest) (*auth.UserInfoResponse, error) {

	if err := uc.otpService.VerifyOTP(ctx, req.PhoneNumber, req.OTP); err != nil {
		if errors.Is(err, ErrOTPNotFound) {
			return nil, errors.New("OTP not found or expired. Please request a new one")
		}
		if errors.Is(err, ErrOTPInvalid) {
			return nil, errors.New("invalid OTP code")
		}
		if errors.Is(err, ErrMaxAttemptsExceeded) {
			return nil, errors.New("maximum verification attempts exceeded. Please request a new OTP")
		}
		return nil, fmt.Errorf("OTP verification failed: %w", err)
	}

	user, err := uc.userRepo.GetByPhoneNumber(ctx, req.PhoneNumber)
	if err != nil {

		newUser := &entity.User{
			PhoneNumber: req.PhoneNumber,
		}

		if err := uc.userRepo.CreateUser(ctx, newUser); err != nil {
			return nil, fmt.Errorf("failed to create user: %w", err)
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
