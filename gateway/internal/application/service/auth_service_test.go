package service

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/Barbod-Biometrics/Backend/gateway/internal/application/dto/auth"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/entity"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/infrastructure/test/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRequestOTP_Success(t *testing.T) {
	ctx := context.Background()
	mockUserRepo := mocks.NewMockUserRepository(t)
	mockOtp := mocks.NewMockOTPUsecase(t)
	mockSMS := mocks.NewMockSMSService(t)
	mockToken := mocks.NewMockTokenUsecase(t)

	svc := NewAuthService(mockUserRepo, mockOtp, mockSMS, mockToken)

	phone := "09120001111"
	otp := "123456"

	mockOtp.On("GenerateAndStoreOTP", mock.Anything, phone).Return(otp, nil)
	expectedMessage := fmt.Sprintf("Your Barbod Biomentrics code is: %s", otp)
	mockSMS.On("Send", mock.Anything, phone, expectedMessage).Return(nil)

	err := svc.RequestOTP(ctx, auth.RequestOTPRequest{PhoneNumber: phone})
	assert.NoError(t, err)

	mockOtp.AssertExpectations(t)
	mockSMS.AssertExpectations(t)
}

func TestRequestOTP_AlreadyExists(t *testing.T) {
	ctx := context.Background()
	mockUserRepo := mocks.NewMockUserRepository(t)
	mockOtp := mocks.NewMockOTPUsecase(t)
	mockSMS := mocks.NewMockSMSService(t)
	mockToken := mocks.NewMockTokenUsecase(t)
	svc := NewAuthService(mockUserRepo, mockOtp, mockSMS, mockToken)

	phone := "09120002222"
	mockOtp.On("GenerateAndStoreOTP", mock.Anything, phone).Return("", ErrOTPAlreadyExists)

	err := svc.RequestOTP(ctx, auth.RequestOTPRequest{PhoneNumber: phone})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "OTP already sent")

	mockOtp.AssertExpectations(t)
}

func TestRequestOTP_GenError(t *testing.T) {
	ctx := context.Background()
	mockUserRepo := mocks.NewMockUserRepository(t)
	mockOtp := mocks.NewMockOTPUsecase(t)
	mockSMS := mocks.NewMockSMSService(t)
	mockToken := mocks.NewMockTokenUsecase(t)
	svc := NewAuthService(mockUserRepo, mockOtp, mockSMS, mockToken)

	phone := "09120003333"
	mockOtp.On("GenerateAndStoreOTP", mock.Anything, phone).Return("", errors.New("redis err"))

	err := svc.RequestOTP(ctx, auth.RequestOTPRequest{PhoneNumber: phone})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to generate otp")

	mockOtp.AssertExpectations(t)
}

func TestRequestOTP_SMSFailure(t *testing.T) {
	ctx := context.Background()
	mockUserRepo := mocks.NewMockUserRepository(t)
	mockOtp := mocks.NewMockOTPUsecase(t)
	mockSMS := mocks.NewMockSMSService(t)
	mockToken := mocks.NewMockTokenUsecase(t)
	svc := NewAuthService(mockUserRepo, mockOtp, mockSMS, mockToken)

	phone := "09120004444"
	otp := "999999"
	mockOtp.On("GenerateAndStoreOTP", mock.Anything, phone).Return(otp, nil)
	expectedMessage := fmt.Sprintf("Your Barbod Biomentrics code is: %s", otp)
	mockSMS.On("Send", mock.Anything, phone, expectedMessage).Return(errors.New("sms down"))

	err := svc.RequestOTP(ctx, auth.RequestOTPRequest{PhoneNumber: phone})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to send sms")

	mockOtp.AssertExpectations(t)
	mockSMS.AssertExpectations(t)
}

func TestVerifyOTP_Success_ExistingUser(t *testing.T) {
	ctx := context.Background()
	mockUserRepo := mocks.NewMockUserRepository(t)
	mockOtp := mocks.NewMockOTPUsecase(t)
	mockSMS := mocks.NewMockSMSService(t)
	mockToken := mocks.NewMockTokenUsecase(t)
	svc := NewAuthService(mockUserRepo, mockOtp, mockSMS, mockToken)

	phone := "09121119999"
	otpVal := "123123"

	mockOtp.On("VerifyOTP", mock.Anything, phone, otpVal).Return(nil)
	// existing user
	user := &entity.User{UserID: 100, PhoneNumber: phone, IsAdmin: true}
	mockUserRepo.On("GetByPhoneNumber", mock.Anything, phone).Return(user, nil)
	mockToken.On("GenerateTokens", mock.Anything, uint64(100), true).Return("at", "rt", nil)

	resp, err := svc.VerifyOTP(ctx, auth.VerifyOTPRequest{PhoneNumber: phone, OTP: otpVal})
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "at", resp.AccessToken)
	assert.Equal(t, "rt", resp.RefereshToken)
	assert.Equal(t, phone, resp.PhoneNumber)
	assert.Equal(t, true, resp.IsAdmin)

	mockOtp.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
	mockToken.AssertExpectations(t)
}

func TestVerifyOTP_Success_NewUserCreated(t *testing.T) {
	ctx := context.Background()
	mockUserRepo := mocks.NewMockUserRepository(t)
	mockOtp := mocks.NewMockOTPUsecase(t)
	mockSMS := mocks.NewMockSMSService(t)
	mockToken := mocks.NewMockTokenUsecase(t)
	svc := NewAuthService(mockUserRepo, mockOtp, mockSMS, mockToken)

	phone := "09122223333"
	otpVal := "222333"

	mockOtp.On("VerifyOTP", mock.Anything, phone, otpVal).Return(nil)
	// user not found
	mockUserRepo.On("GetByPhoneNumber", mock.Anything, phone).Return((*entity.User)(nil), nil)
	// capture CreateUser call and set ID
	mockUserRepo.On("CreateUser", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		u := args.Get(1).(*entity.User)
		u.UserID = 501
	}).Return(nil)
	mockToken.On("GenerateTokens", mock.Anything, uint64(501), false).Return("at2", "rt2", nil)

	resp, err := svc.VerifyOTP(ctx, auth.VerifyOTPRequest{PhoneNumber: phone, OTP: otpVal})
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "at2", resp.AccessToken)
	assert.Equal(t, "rt2", resp.RefereshToken)
	assert.Equal(t, phone, resp.PhoneNumber)
	assert.Equal(t, false, resp.IsAdmin)

	mockOtp.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
	mockToken.AssertExpectations(t)
}

func TestVerifyOTP_ErrCases(t *testing.T) {
	ctx := context.Background()
	mockUserRepo := mocks.NewMockUserRepository(t)
	mockOtp := mocks.NewMockOTPUsecase(t)
	mockSMS := mocks.NewMockSMSService(t)
	mockToken := mocks.NewMockTokenUsecase(t)
	svc := NewAuthService(mockUserRepo, mockOtp, mockSMS, mockToken)

	phone := "09130003333"

	// NotFound
	mockOtp.On("VerifyOTP", mock.Anything, phone, "1").Return(ErrOTPNotFound)
	resp, err := svc.VerifyOTP(ctx, auth.VerifyOTPRequest{PhoneNumber: phone, OTP: "1"})
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "OTP not found or expired")

	// Invalid
	mockOtp.On("VerifyOTP", mock.Anything, phone, "2").Return(ErrOTPInvalid)
	resp, err = svc.VerifyOTP(ctx, auth.VerifyOTPRequest{PhoneNumber: phone, OTP: "2"})
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "invalid OTP code")

	// Max attempts
	mockOtp.On("VerifyOTP", mock.Anything, phone, "3").Return(ErrMaxAttemptsExceeded)
	resp, err = svc.VerifyOTP(ctx, auth.VerifyOTPRequest{PhoneNumber: phone, OTP: "3"})
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "maximum verification attempts exceeded")

	mockOtp.AssertExpectations(t)
}

func TestVerifyOTP_UserRepoErrors(t *testing.T) {
	ctx := context.Background()
	mockUserRepo := mocks.NewMockUserRepository(t)
	mockOtp := mocks.NewMockOTPUsecase(t)
	mockSMS := mocks.NewMockSMSService(t)
	mockToken := mocks.NewMockTokenUsecase(t)
	svc := NewAuthService(mockUserRepo, mockOtp, mockSMS, mockToken)

	phone := "09140004444"
	otpVal := "333444"

	mockOtp.On("VerifyOTP", mock.Anything, phone, otpVal).Return(nil)
	mockUserRepo.On("GetByPhoneNumber", mock.Anything, phone).Return((*entity.User)(nil), errors.New("db error"))

	resp, err := svc.VerifyOTP(ctx, auth.VerifyOTPRequest{PhoneNumber: phone, OTP: otpVal})
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "failed to get user")

	mockOtp.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

func TestVerifyOTP_CreateUserError(t *testing.T) {
	ctx := context.Background()
	mockUserRepo := mocks.NewMockUserRepository(t)
	mockOtp := mocks.NewMockOTPUsecase(t)
	mockSMS := mocks.NewMockSMSService(t)
	mockToken := mocks.NewMockTokenUsecase(t)
	svc := NewAuthService(mockUserRepo, mockOtp, mockSMS, mockToken)

	phone := "09150005555"
	otpVal := "555666"

	mockOtp.On("VerifyOTP", mock.Anything, phone, otpVal).Return(nil)
	mockUserRepo.On("GetByPhoneNumber", mock.Anything, phone).Return((*entity.User)(nil), nil)
	mockUserRepo.On("CreateUser", mock.Anything, mock.Anything).Return(errors.New("insert failed"))

	resp, err := svc.VerifyOTP(ctx, auth.VerifyOTPRequest{PhoneNumber: phone, OTP: otpVal})
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "failed to create user")

	mockOtp.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

func TestVerifyOTP_TokenGenerationError(t *testing.T) {
	ctx := context.Background()
	mockUserRepo := mocks.NewMockUserRepository(t)
	mockOtp := mocks.NewMockOTPUsecase(t)
	mockSMS := mocks.NewMockSMSService(t)
	mockToken := mocks.NewMockTokenUsecase(t)
	svc := NewAuthService(mockUserRepo, mockOtp, mockSMS, mockToken)

	phone := "09160006666"
	otpVal := "666777"
	user := &entity.User{UserID: 404, PhoneNumber: phone, IsAdmin: false}

	mockOtp.On("VerifyOTP", mock.Anything, phone, otpVal).Return(nil)
	mockUserRepo.On("GetByPhoneNumber", mock.Anything, phone).Return(user, nil)
	mockToken.On("GenerateTokens", mock.Anything, uint64(404), false).Return("", "", errors.New("jwt fail"))

	resp, err := svc.VerifyOTP(ctx, auth.VerifyOTPRequest{PhoneNumber: phone, OTP: otpVal})
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "failed to generate tokens")

	mockOtp.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
	mockToken.AssertExpectations(t)
}
