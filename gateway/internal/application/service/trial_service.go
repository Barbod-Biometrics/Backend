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
	maxTrialAttempts = 50 // for development/testing
	trialExpiration  = 24 * time.Hour
)

// Lua script for atomic check-and-decrement operation
// Behavior:
// - If the key does not exist, initialize it and consume one attempt.
// - If attempts <= 0 the script returns {-1, attempts} to indicate exceeded.
// - Otherwise decrement and return {1, newAttempts} where newAttempts is remaining attempts.
const checkAndDecrementScript = `
local key = KEYS[1]
local maxAttempts = tonumber(ARGV[1])
local expiration = tonumber(ARGV[2])

local current = redis.call('GET', key)

if not current then
	local newAttempts = maxAttempts - 1
	redis.call('SET', key, newAttempts, 'EX', expiration)
	if newAttempts < 0 then
		return {-1, newAttempts}
	end
	return {1, newAttempts}
end

local attempts = tonumber(current)
if attempts <= 0 then
	return {-1, attempts}
end

local newAttempts = attempts - 1
redis.call('SET', key, newAttempts, 'EX', expiration)
return {1, newAttempts}
`

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

func (s *TrialService) CheckAndDecrementTrial(ctx context.Context, ip string, serviceType string) (int, time.Duration, error) {
	key := fmt.Sprintf("trial:%s:%s", serviceType, ip)

	// atomic check-and-decrement operation
	result, err := s.redisClient.EvalScript(ctx, checkAndDecrementScript, []string{key}, maxTrialAttempts, int(trialExpiration.Seconds()))
	if err != nil {
		s.logger.Error("Failed to execute trial check-and-decrement",
			logger.Field{Key: "error", Value: err},
			logger.Field{Key: "ip", Value: ip},
			logger.Field{Key: "service", Value: serviceType},
		)
		return 0, 0, exception.NewInternalError("ERR_TRIAL_REDIS", "failed to check trial status", err)
	}

	resultSlice, ok := result.([]interface{})
	if !ok || len(resultSlice) < 2 {
		s.logger.Error("Invalid script result format",
			logger.Field{Key: "ip", Value: ip},
			logger.Field{Key: "service", Value: serviceType},
			logger.Field{Key: "raw_result", Value: result},
		)
		return 0, 0, exception.NewInternalError("ERR_TRIAL_SCRIPT", "invalid script result format", nil)
	}

	toInt := func(v interface{}) (int64, error) {
		switch t := v.(type) {
		case int64:
			return t, nil
		case int:
			return int64(t), nil
		case float64:
			return int64(t), nil
		case string:
			return strconv.ParseInt(t, 10, 64)
		case []byte:
			return strconv.ParseInt(string(t), 10, 64)
		default:
			return 0, fmt.Errorf("unsupported type %T", v)
		}
	}

	status, err := toInt(resultSlice[0])
	if err != nil {
		s.logger.Error("Invalid status from Lua script",
			logger.Field{Key: "error", Value: err},
			logger.Field{Key: "ip", Value: ip},
			logger.Field{Key: "service", Value: serviceType},
		)
		return 0, 0, exception.NewInternalError("ERR_TRIAL_SCRIPT", "invalid script status", err)
	}

	attempts, err := toInt(resultSlice[1])
	if err != nil {
		s.logger.Error("Invalid attempts from Lua script",
			logger.Field{Key: "error", Value: err},
			logger.Field{Key: "ip", Value: ip},
			logger.Field{Key: "service", Value: serviceType},
		)
		return 0, 0, exception.NewInternalError("ERR_TRIAL_SCRIPT", "invalid script attempts", err)
	}

	// get TTL for key (may be -1/ -2 meaning no expire / key missing)
	ttl, ttlErr := s.redisClient.GetTTL(ctx, key)

	if status == -1 {
		s.logger.Warn("Trial limit exceeded",
			logger.Field{Key: "ip", Value: ip},
			logger.Field{Key: "service", Value: serviceType},
			logger.Field{Key: "attempts", Value: attempts},
		)
		// Return remaining attempts (may be <= 0) and TTL alongside the error so callers/middleware can surface it.
		if ttlErr != nil {
			return int(attempts), 0, exception.ErrTrialExceeded
		}
		return int(attempts), ttl, exception.ErrTrialExceeded
	}

	s.logger.Info("Trial attempt used",
		logger.Field{Key: "ip", Value: ip},
		logger.Field{Key: "service", Value: serviceType},
		logger.Field{Key: "remaining_attempts", Value: attempts},
	)

	// on success also return TTL if available
	ttl, _ = s.redisClient.GetTTL(ctx, key)
	return int(attempts), ttl, nil
}

func (s *TrialService) GetRemainingAttempts(ctx context.Context, ip string, serviceType string) (int, error) {
	key := fmt.Sprintf("trial:%s:%s", serviceType, ip)

	exists, err := s.redisClient.Exists(ctx, key)
	if err != nil {
		return 0, fmt.Errorf("failed to check trial status: %w", err)
	}

	if !exists {
		return maxTrialAttempts, nil
	}

	// Note: We don't need a Lua script here since this is just a read operation
	// and doesn't have the race condition issues of check-and-decrement
	attemptsStr, err := s.redisClient.Get(ctx, key)
	if err != nil {
		return 0, fmt.Errorf("failed to get trial attempts: %w", err)
	}

	attempts, err := strconv.Atoi(attemptsStr)
	if err != nil {
		return 0, fmt.Errorf("failed to parse trial attempts: %w", err)
	}

	return attempts, nil
}
