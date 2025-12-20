package service

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/exception"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/logger"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/infrastructure/database/redis"
)

const (
	maxTrialAttempts = 2 // Start at 2 (gives 3 total attempts: 2, 1, 0)
	trialExpiration  = 24 * time.Hour
)

type TrialService struct {
	redisClient *redis.RedisClient
	logger      logger.Logger
}

func NewTrialService(redisClient *redis.RedisClient, logger logger.Logger) *TrialService {
	return &TrialService{
		redisClient: redisClient,
		logger:      logger,
	}
}

func (s *TrialService) CheckAndDecrementTrial(ctx context.Context, ip string, serviceType string) (int, error) {
	key := fmt.Sprintf("trial:%s:%s", serviceType, ip)

	exists, err := s.redisClient.Exists(ctx, key)
	if err != nil {
		s.logger.Error("Failed to check trial key existence",
			logger.Field{Key: "error", Value: err},
			logger.Field{Key: "ip", Value: ip},
			logger.Field{Key: "service", Value: serviceType},
		)
		return 0, fmt.Errorf("failed to check trial status: %w", err)
	}

	if !exists {
		err = s.redisClient.Set(ctx, key, maxTrialAttempts, trialExpiration)
		if err != nil {
			s.logger.Error("Failed to initialize trial for new IP",
				logger.Field{Key: "error", Value: err},
				logger.Field{Key: "ip", Value: ip},
				logger.Field{Key: "service", Value: serviceType},
			)
			return 0, fmt.Errorf("failed to initialize trial: %w", err)
		}

		s.logger.Info("Initialized new trial for IP",
			logger.Field{Key: "ip", Value: ip},
			logger.Field{Key: "service", Value: serviceType},
			logger.Field{Key: "attempts", Value: maxTrialAttempts},
		)
	}

	attemptsStr, err := s.redisClient.Get(ctx, key)
	if err != nil {
		s.logger.Error("Failed to get trial attempts",
			logger.Field{Key: "error", Value: err},
			logger.Field{Key: "ip", Value: ip},
			logger.Field{Key: "service", Value: serviceType},
		)
		return 0, fmt.Errorf("failed to get trial attempts: %w", err)
	}

	attempts, err := strconv.Atoi(attemptsStr)
	if err != nil {
		s.logger.Error("Failed to parse trial attempts",
			logger.Field{Key: "error", Value: err},
			logger.Field{Key: "ip", Value: ip},
			logger.Field{Key: "service", Value: serviceType},
			logger.Field{Key: "attempts_str", Value: attemptsStr},
		)
		return 0, fmt.Errorf("failed to parse trial attempts: %w", err)
	}

	if attempts < 0 {
		s.logger.Warn("Trial limit exceeded",
			logger.Field{Key: "ip", Value: ip},
			logger.Field{Key: "service", Value: serviceType},
			logger.Field{Key: "attempts", Value: attempts},
		)
		return 0, exception.ErrTrialExceeded
	}

	newAttempts := attempts - 1
	err = s.redisClient.Set(ctx, key, newAttempts, trialExpiration)
	if err != nil {
		s.logger.Error("Failed to decrement trial attempts",
			logger.Field{Key: "error", Value: err},
			logger.Field{Key: "ip", Value: ip},
			logger.Field{Key: "service", Value: serviceType},
		)
		return 0, fmt.Errorf("failed to update trial attempts: %w", err)
	}

	s.logger.Info("Trial attempt used",
		logger.Field{Key: "ip", Value: ip},
		logger.Field{Key: "service", Value: serviceType},
		logger.Field{Key: "remaining_attempts", Value: newAttempts},
	)

	return newAttempts, nil
}

func (s *TrialService) GetRemainingAttempts(ctx context.Context, ip string, serviceType string) (int, error) {
	key := fmt.Sprintf("trial:%s:%s", serviceType, ip)

	exists, err := s.redisClient.Exists(ctx, key)
	if err != nil {
		return 0, fmt.Errorf("failed to check trial status: %w", err)
	}

	if !exists {
		return maxTrialAttempts + 1, nil // +1 because we start at 2 (3 total attempts)
	}

	attemptsStr, err := s.redisClient.Get(ctx, key)
	if err != nil {
		return 0, fmt.Errorf("failed to get trial attempts: %w", err)
	}

	attempts, err := strconv.Atoi(attemptsStr)
	if err != nil {
		return 0, fmt.Errorf("failed to parse trial attempts: %w", err)
	}

	return attempts + 1, nil
}
