package service

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/Barbod-Biometrics/Backend/gateway/bootstrap"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/infrastructure/test/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func buildConfig(otpLength, expiryMinutes, maxAttempts int, backdoor string) *bootstrap.Config {
	cfg := &bootstrap.Config{
		Constants: bootstrap.NewConstants(),
		Env:       bootstrap.NewEnvironment(),
	}
	cfg.Env.OTP.Length = otpLength
	cfg.Env.OTP.ExpiryMinute = expiryMinutes
	cfg.Env.OTP.MaxAttempts = maxAttempts
	cfg.Env.OTP.BackdoorCode = backdoor
	return cfg
}

func TestGenerateAndStoreOTP_Success(t *testing.T) {
	ctx := context.Background()
	mockCache := mocks.NewMockCacheRepository(t)
	cfg := buildConfig(6, 5, 3, "")
	svc := NewOTPService(mockCache, cfg)

	phone := "09120001111"
	key := cfg.Constants.RedisKey.GenerateOTPKey(phone)
	attemptsKey := fmt.Sprintf("%s:attempts", key)

	mockCache.On("Exists", mock.Anything, key).Return(false, nil)
	// When Set is called, it should receive the generated OTP; we accept any value and return nil
	mockCache.On("Set", mock.Anything, key, mock.Anything, mock.Anything).Return(nil)
	mockCache.On("Set", mock.Anything, attemptsKey, "0", mock.Anything).Return(nil)

	otp, err := svc.GenerateAndStoreOTP(ctx, phone)
	assert.NoError(t, err)
	assert.Equal(t, 6, len(otp))

	mockCache.AssertExpectations(t)
}

func TestGenerateAndStoreOTP_AlreadyExists(t *testing.T) {
	ctx := context.Background()
	mockCache := mocks.NewMockCacheRepository(t)
	cfg := buildConfig(6, 5, 3, "")
	svc := NewOTPService(mockCache, cfg)

	phone := "09129992222"
	key := cfg.Constants.RedisKey.GenerateOTPKey(phone)

	mockCache.On("Exists", mock.Anything, key).Return(true, nil)

	otp, err := svc.GenerateAndStoreOTP(ctx, phone)
	assert.ErrorIs(t, err, ErrOTPAlreadyExists)
	assert.Equal(t, "", otp)

	mockCache.AssertExpectations(t)
}

func TestGenerateAndStoreOTP_ExistsError(t *testing.T) {
	ctx := context.Background()
	mockCache := mocks.NewMockCacheRepository(t)
	cfg := buildConfig(6, 5, 3, "")
	svc := NewOTPService(mockCache, cfg)

	phone := "09121112222"
	key := cfg.Constants.RedisKey.GenerateOTPKey(phone)
	mockCache.On("Exists", mock.Anything, key).Return(false, errors.New("redis error"))

	otp, err := svc.GenerateAndStoreOTP(ctx, phone)
	assert.Error(t, err)
	assert.Equal(t, "", otp)

	mockCache.AssertExpectations(t)
}

func TestGenerateAndStoreOTP_SetError(t *testing.T) {
	ctx := context.Background()
	mockCache := mocks.NewMockCacheRepository(t)
	cfg := buildConfig(6, 5, 3, "")
	svc := NewOTPService(mockCache, cfg)

	phone := "09123334444"
	key := cfg.Constants.RedisKey.GenerateOTPKey(phone)

	mockCache.On("Exists", mock.Anything, key).Return(false, nil)
	mockCache.On("Set", mock.Anything, key, mock.Anything, mock.Anything).Return(errors.New("set error"))

	otp, err := svc.GenerateAndStoreOTP(ctx, phone)
	assert.Error(t, err)
	assert.Equal(t, "", otp)

	mockCache.AssertExpectations(t)
}

func TestGenerateAndStoreOTP_AttemptsSetError(t *testing.T) {
	ctx := context.Background()
	mockCache := mocks.NewMockCacheRepository(t)
	cfg := buildConfig(6, 5, 3, "")
	svc := NewOTPService(mockCache, cfg)

	phone := "09123334445"
	key := cfg.Constants.RedisKey.GenerateOTPKey(phone)
	attemptsKey := fmt.Sprintf("%s:attempts", key)

	mockCache.On("Exists", mock.Anything, key).Return(false, nil)
	mockCache.On("Set", mock.Anything, key, mock.Anything, mock.Anything).Return(nil)
	mockCache.On("Set", mock.Anything, attemptsKey, "0", mock.Anything).Return(errors.New("attempts set error"))

	otp, err := svc.GenerateAndStoreOTP(ctx, phone)
	assert.Error(t, err)
	assert.Equal(t, "", otp)

	mockCache.AssertExpectations(t)
}

func TestVerifyOTP_Success(t *testing.T) {
	ctx := context.Background()
	mockCache := mocks.NewMockCacheRepository(t)
	cfg := buildConfig(6, 5, 3, "")
	svc := NewOTPService(mockCache, cfg)

	phone := "09110001111"
	key := cfg.Constants.RedisKey.GenerateOTPKey(phone)
	attemptsKey := fmt.Sprintf("%s:attempts", key)
	otp := "123456"

	mockCache.On("Exists", mock.Anything, key).Return(true, nil)
	mockCache.On("Increment", mock.Anything, attemptsKey).Return(int64(1), nil)
	mockCache.On("Get", mock.Anything, key).Return(otp, nil)
	mockCache.On("Delete", mock.Anything, []string{key, attemptsKey}).Return(nil)

	err := svc.VerifyOTP(ctx, phone, otp)
	assert.NoError(t, err)

	mockCache.AssertExpectations(t)
}

func TestVerifyOTP_Backdoor(t *testing.T) {
	ctx := context.Background()
	mockCache := mocks.NewMockCacheRepository(t)
	cfg := buildConfig(6, 5, 3, "777777")
	svc := NewOTPService(mockCache, cfg)

	phone := "09112223333"
	mockCache.On("Delete", mock.Anything, []string{cfg.Constants.RedisKey.GenerateOTPKey(phone), cfg.Constants.RedisKey.GenerateOTPKey(phone) + ":attempts"}).Return(nil)
	err := svc.VerifyOTP(ctx, phone, "777777")
	assert.NoError(t, err)
}

func TestVerifyOTP_NotFound(t *testing.T) {
	ctx := context.Background()
	mockCache := mocks.NewMockCacheRepository(t)
	cfg := buildConfig(6, 5, 3, "")
	svc := NewOTPService(mockCache, cfg)

	phone := "09110002222"
	key := cfg.Constants.RedisKey.GenerateOTPKey(phone)

	mockCache.On("Exists", mock.Anything, key).Return(false, nil)

	err := svc.VerifyOTP(ctx, phone, "000000")
	assert.ErrorIs(t, err, ErrOTPNotFound)

	mockCache.AssertExpectations(t)
}

func TestVerifyOTP_InvalidOTP(t *testing.T) {
	ctx := context.Background()
	mockCache := mocks.NewMockCacheRepository(t)
	cfg := buildConfig(6, 5, 3, "")
	svc := NewOTPService(mockCache, cfg)

	phone := "09110003333"
	key := cfg.Constants.RedisKey.GenerateOTPKey(phone)
	attemptsKey := fmt.Sprintf("%s:attempts", key)

	mockCache.On("Exists", mock.Anything, key).Return(true, nil)
	mockCache.On("Increment", mock.Anything, attemptsKey).Return(int64(1), nil)
	mockCache.On("Get", mock.Anything, key).Return("654321", nil)

	err := svc.VerifyOTP(ctx, phone, "123456")
	assert.ErrorIs(t, err, ErrOTPInvalid)

	mockCache.AssertExpectations(t)
}

func TestVerifyOTP_MaxAttemptsExceeded(t *testing.T) {
	ctx := context.Background()
	mockCache := mocks.NewMockCacheRepository(t)
	cfg := buildConfig(6, 5, 1, "") // maxAttempts = 1
	svc := NewOTPService(mockCache, cfg)

	phone := "09110004444"
	key := cfg.Constants.RedisKey.GenerateOTPKey(phone)
	attemptsKey := fmt.Sprintf("%s:attempts", key)

	mockCache.On("Exists", mock.Anything, key).Return(true, nil)
	// First increment returns 2 (exceeds maxAttempts)
	mockCache.On("Increment", mock.Anything, attemptsKey).Return(int64(2), nil)
	mockCache.On("Delete", mock.Anything, []string{key, attemptsKey}).Return(nil)

	err := svc.VerifyOTP(ctx, phone, "whatever")
	assert.ErrorIs(t, err, ErrMaxAttemptsExceeded)

	mockCache.AssertExpectations(t)
}

func TestVerifyOTP_IncrementError(t *testing.T) {
	ctx := context.Background()
	mockCache := mocks.NewMockCacheRepository(t)
	cfg := buildConfig(6, 5, 3, "")
	svc := NewOTPService(mockCache, cfg)

	phone := "09110005555"
	key := cfg.Constants.RedisKey.GenerateOTPKey(phone)
	attemptsKey := fmt.Sprintf("%s:attempts", key)

	mockCache.On("Exists", mock.Anything, key).Return(true, nil)
	mockCache.On("Increment", mock.Anything, attemptsKey).Return(int64(0), errors.New("incr error"))

	err := svc.VerifyOTP(ctx, phone, "whatever")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to increment attempts")

	mockCache.AssertExpectations(t)
}

func TestVerifyOTP_GetError(t *testing.T) {
	ctx := context.Background()
	mockCache := mocks.NewMockCacheRepository(t)
	cfg := buildConfig(6, 5, 3, "")
	svc := NewOTPService(mockCache, cfg)

	phone := "09110007777"
	key := cfg.Constants.RedisKey.GenerateOTPKey(phone)
	attemptsKey := fmt.Sprintf("%s:attempts", key)

	mockCache.On("Exists", mock.Anything, key).Return(true, nil)
	mockCache.On("Increment", mock.Anything, attemptsKey).Return(int64(1), nil)
	mockCache.On("Get", mock.Anything, key).Return("", errors.New("get error"))

	err := svc.VerifyOTP(ctx, phone, "whatever")
	assert.ErrorIs(t, err, ErrOTPNotFound)

	mockCache.AssertExpectations(t)
}

func TestVerifyOTP_DeleteAfterVerifyError(t *testing.T) {
	ctx := context.Background()
	mockCache := mocks.NewMockCacheRepository(t)
	cfg := buildConfig(6, 5, 3, "")
	svc := NewOTPService(mockCache, cfg)

	phone := "09110006666"
	key := cfg.Constants.RedisKey.GenerateOTPKey(phone)
	attemptsKey := fmt.Sprintf("%s:attempts", key)
	otp := "222222"

	mockCache.On("Exists", mock.Anything, key).Return(true, nil)
	mockCache.On("Increment", mock.Anything, attemptsKey).Return(int64(1), nil)
	mockCache.On("Get", mock.Anything, key).Return(otp, nil)
	mockCache.On("Delete", mock.Anything, []string{key, attemptsKey}).Return(errors.New("del error"))

	err := svc.VerifyOTP(ctx, phone, otp)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to delete OTP after verification")

	mockCache.AssertExpectations(t)
}

func TestVerifyOTP_ExistsError(t *testing.T) {
	ctx := context.Background()
	mockCache := mocks.NewMockCacheRepository(t)
	cfg := buildConfig(6, 5, 3, "")
	svc := NewOTPService(mockCache, cfg)

	phone := "09110008888"
	key := cfg.Constants.RedisKey.GenerateOTPKey(phone)

	mockCache.On("Exists", mock.Anything, key).Return(false, errors.New("exists err"))

	err := svc.VerifyOTP(ctx, phone, "whatever")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to check OTP")

	mockCache.AssertExpectations(t)
}
