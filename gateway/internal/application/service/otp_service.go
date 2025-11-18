package service

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/Barbod-Biometrics/Backend/gateway/bootstrap"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/application/usecase"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/repository"
)

var (
	ErrOTPNotFound         = errors.New("OTP not found or expired")
	ErrOTPInvalid          = errors.New("invalid OTP")
	ErrOTPAlreadyExists    = errors.New("OTP already exists for this phone number")
	ErrMaxAttemptsExceeded = errors.New("maximum OTP verification attempts exceeded")
)

type otpService struct {
	cache  repository.CacheRepository
	config *bootstrap.Config
}

func NewOTPService(cache repository.CacheRepository, config *bootstrap.Config) *otpService {
	return &otpService{
		cache:  cache,
		config: config,
	}
}

var _ usecase.OTPService = (*otpService)(nil)

func (s *otpService) GenerateAndStoreOTP(ctx context.Context, phoneNumber string) (string, error) {

	key := s.config.Constants.RedisKey.GenerateOTPKey(phoneNumber)

	exists, err := s.cache.Exists(ctx, key)
	if err != nil {
		return "", fmt.Errorf("failed to check OTP existence: %w", err)
	}

	if exists {
		return "", ErrOTPAlreadyExists
	}

	otp, err := s.generateOTP(s.config.Env.OTP.Length)
	if err != nil {
		return "", fmt.Errorf("failed to generate OTP: %w", err)
	}

	expiration := time.Duration(s.config.Env.OTP.ExpiryMinute) * time.Minute

	if err := s.cache.Set(ctx, key, otp, expiration); err != nil {
		return "", fmt.Errorf("failed to store OTP: %w", err)
	}

	attemptsKey := fmt.Sprintf("%s:attempts", key)
	if err := s.cache.Set(ctx, attemptsKey, "0", expiration); err != nil {
		return "", fmt.Errorf("failed to initialize attempts counter: %w", err)
	}

	return otp, nil
}

func (s *otpService) VerifyOTP(ctx context.Context, phoneNmber string, otp string) error {
	key := s.config.Constants.RedisKey.GenerateOTPKey(phoneNmber)
	attemptsKey := fmt.Sprintf("%s:attempts", key)

	exists, err := s.cache.Exists(ctx, key)
	if err != nil {
		return fmt.Errorf("failed to check OTP: %w", err)
	}

	if !exists {
		return ErrOTPNotFound
	}

	attempts, err := s.cache.Increment(ctx, attemptsKey)
	if err != nil {
		return fmt.Errorf("failed to increment attempts: %w", err)
	}

	if attempts > int64(s.config.Env.OTP.MaxAttempts) {
		s.cache.Delete(ctx, key, attemptsKey)
		return ErrMaxAttemptsExceeded
	}

	storedOTP, err := s.cache.Get(ctx, key)
	if err != nil {
		return ErrOTPNotFound
	}

	if storedOTP != otp {
		return ErrOTPInvalid
	}

	if err := s.cache.Delete(ctx, key, attemptsKey); err != nil {
		return fmt.Errorf("failed to delete OTP after verification: %w", err)
	}

	return nil
}

func (s *otpService) generateOTP(length int) (string, error) {
	if length <= 0 {
		length = 6
	}

	max := big.NewInt(1)
	for i := 0; i < length; i++ {
		max.Mul(max, big.NewInt(10))
	}

	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}

	format := fmt.Sprintf("%%0%ddd", length)
	return fmt.Sprintf(format, n.Int64()), nil
}
