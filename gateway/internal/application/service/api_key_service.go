package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/Barbod-Biometrics/Backend/gateway/internal/application/usecase"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/entity"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/logger"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/repository"
	"golang.org/x/crypto/bcrypt"
)

type APIKeyService struct {
	apiKeyRepo repository.APIKeyRepository
	logger     logger.Logger
}

var _ usecase.APIKeyUsecase = (*APIKeyService)(nil)

func NewAPIKeyService(apiKeyRepo repository.APIKeyRepository, logger logger.Logger) *APIKeyService {
	return &APIKeyService{
		apiKeyRepo: apiKeyRepo,
		logger:     logger,
	}
}

func (s *APIKeyService) GenerateKey(ctx context.Context, profileID uint64) (string, error) {

	s.logger.Info("starting api key generation", logger.Field{Key: "profile_id", Value: profileID})

	// safety check: ensure no active key exists for this profile_id
	existingKey, err := s.apiKeyRepo.GetActiveByProfileID(ctx, profileID)
	if err == nil && existingKey != nil {
		s.logger.Warn("generation blocked: active key exists", logger.Field{Key: "profile_id", Value: profileID})
		return "", errors.New("active key exists for this profile id")
	}

	if err != nil {
		return "", errors.New("record not found in checking active key for the profile id")
	}

	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		s.logger.Error("failed to generate random bytes", logger.Field{Key: "error", Value: err})
		return "", err
	}

	// raw key string (prefix + random hex)
	const visualPrefix = "bb_live_"
	randomPart := hex.EncodeToString(bytes)
	fullRawKey := visualPrefix + randomPart

	dbPrefix := randomPart[:8]

	hasedBytes, err := bcrypt.GenerateFromPassword([]byte(fullRawKey), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Error("failed to has api key", logger.Field{Key: "error", Value: err})
		return "", err
	}

	apiKey := &entity.APIKey{
		ProfileID: profileID,
		KeyHash:   string(hasedBytes),
		KeyPrefix: dbPrefix,
		IsActive:  true,
		CreatedAt: time.Now(),
	}

	err = s.apiKeyRepo.Create(ctx, apiKey)
	if err != nil {
		s.logger.Error("failed to save api key to database", logger.Field{Key: "profile_id", Value: profileID},
			logger.Field{Key: "error", Value: err},
		)
		return "", err
	}

	s.logger.Info("api key generated successfully", logger.Field{Key: "profile_id", Value: profileID})

	return fullRawKey, nil
}

func (s *APIKeyService) RegenerateKey(ctx context.Context, profileID uint64) (string, error) {

	s.logger.Info("starting re-generating api key", logger.Field{Key: "profile_id", Value: profileID})

	oldKey, err := s.apiKeyRepo.GetActiveByProfileID(ctx, profileID)
	if err == nil && oldKey != nil {
		strKeyID := strconv.FormatUint(oldKey.KeyID, 10)

		if revokeErr := s.apiKeyRepo.Revoke(ctx, strKeyID); revokeErr != nil {
			s.logger.Error("failed to revoke old api key", logger.Field{Key: "old_key_id", Value: strKeyID},
				logger.Field{Key: "error", Value: err})
			return "", revokeErr
		}
		s.logger.Info("old api key revoked successfully", logger.Field{Key: "old_key_id", Value: strKeyID})
	}

	return s.GenerateKey(ctx, profileID)
}

func (s *APIKeyService) Authenticate(ctx context.Context, rawKey string) (uint64, error) {
	const visualPrefix = "bb_live_"

	if !strings.HasPrefix(rawKey, visualPrefix) {
		s.logger.Warn("authentication failed: invalid prefix", logger.Field{Key: "key_fragment", Value: rawKey[:4]})
		return 0, errors.New("invalid key format")
	}

	if len(rawKey) <= len(visualPrefix)+8 {
		s.logger.Warn("authentication failed: invalid length")
		return 0, errors.New("invalid key length")
	}

	searchPrefix := rawKey[len(visualPrefix) : len(visualPrefix)+8]

	apiKey, err := s.apiKeyRepo.GetByPrefix(ctx, searchPrefix)
	if err != nil {
		s.logger.Warn("authentication failed: key prefix not found", logger.Field{Key: "prefix", Value: searchPrefix})
		return 0, errors.New("authentication failed")
	}

	if !apiKey.IsActive {
		s.logger.Warn("authenetication failed: key is revoked", logger.Field{Key: "profile_id", Value: apiKey.ProfileID})
		return 0, errors.New("key is revoked")
	}

	err = bcrypt.CompareHashAndPassword([]byte(apiKey.KeyHash), []byte(rawKey))
	if err != nil {
		s.logger.Warn("authentication failed: has mismatch", logger.Field{Key: "profile_id", Value: apiKey.ProfileID})
		return 0, errors.New("authentication failed")
	}

	s.logger.Debug("api key authenticated", logger.Field{Key: "profile_id", Value: apiKey.ProfileID})

	return apiKey.ProfileID, nil

}
