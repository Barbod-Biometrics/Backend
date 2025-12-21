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

var _ usecase.AuthUsecase = (*AuthService)(nil)

type AuthService struct {
	userRepo     repository.UserRepository
	otpService   usecase.OTPUsecase
	SMSService   communication.SMSService
	tokenService usecase.TokenUsecase
}

func NewAuthService(userRepo repository.UserRepository, otpService usecase.OTPUsecase, SMSService communication.SMSService, tokenService usecase.TokenUsecase) *AuthService {
	return &AuthService{
		userRepo:     userRepo,
		otpService:   otpService,
		SMSService:   SMSService,
		tokenService: tokenService,
	}
}

func (uc *AuthService) RequestOTP(ctx context.Context, req auth.RequestOTPRequest) error {

	plainOTP, err := uc.otpService.GenerateAndStoreOTP(ctx, req.PhoneNumber)
	if err != nil {
		return err
	}

	message := fmt.Sprintf("Your Barbod Biomentrics code is: %s", plainOTP)
	if err := uc.SMSService.Send(ctx, req.PhoneNumber, message); err != nil {
		_ = uc.otpService.DeleteOTP(ctx, req.PhoneNumber)
		return fmt.Errorf("failed to send sms: %w", err)
	}

	return nil
}

func (uc *AuthService) VerifyOTP(ctx context.Context, req auth.VerifyOTPRequest) (*auth.UserInfoResponse, error) {

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
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if user == nil {
		newUser := &entity.User{
			PhoneNumber: req.PhoneNumber,
			IsAdmin:     false,
		}

		if err := uc.userRepo.CreateUser(ctx, newUser); err != nil {
			return nil, fmt.Errorf("failed to create user: %w", err)
		}

		user = newUser
	}

	accessToken, refreshToken, err := uc.tokenService.GenerateTokens(ctx, user.UserID, user.IsAdmin)

	if err != nil {
		return nil, errors.New("failed to generate tokens")
	}

	return &auth.UserInfoResponse{
		AccessToken:   accessToken,
		RefereshToken: refreshToken,
		PhoneNumber:   req.PhoneNumber,
		IsAdmin:       user.IsAdmin,
	}, nil

}
