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
	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/repository"
	"golang.org/x/crypto/bcrypt"
)

type APIKeyService struct {
	apiKeyRepo repository.APIKeyRepository
}

var _ usecase.APIKeyService = (*APIKeyService)(nil)

func NewAPIKeyService(apiKeyRepo repository.APIKeyRepository) *APIKeyService {
	return &APIKeyService{
		apiKeyRepo: apiKeyRepo,
	}
}

func (s *APIKeyService) GenerateKey(ctx context.Context, profileID uint64) (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	// raw key string (prefix + random hex)
	const visualPrefix = "bb_live_"
	randomPart := hex.EncodeToString(bytes)
	fullRawKey := visualPrefix + randomPart

	dbPrefix := randomPart[:8]

	hasedBytes, err := bcrypt.GenerateFromPassword([]byte(fullRawKey), bcrypt.DefaultCost)
	if err != nil {
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
		return "", err
	}

	return fullRawKey, nil
}

func (s *APIKeyService) RegenerateKey(ctx context.Context, profileID uint64) (string, error) {
	oldKey, err := s.apiKeyRepo.GetActiveByProfileID(ctx, profileID)
	if err == nil && oldKey != nil {
		strKeyID := strconv.FormatUint(oldKey.KeyID, 10)

		if revokeErr := s.apiKeyRepo.Revoke(ctx, strKeyID); revokeErr != nil {
			return "", revokeErr
		}
	}

	return s.GenerateKey(ctx, profileID)
}

func (s *APIKeyService) Authenticate(ctx context.Context, rawKey string) (uint64, error) {
	const visualPrefix = "bb_live_"

	if !strings.HasPrefix(rawKey, visualPrefix) {
		return 0, errors.New("invalid key format")
	}

	if len(rawKey) <= len(visualPrefix)+8 {
		return 0, errors.New("invalid key length")
	}

	searchPrefix := rawKey[len(visualPrefix) : len(visualPrefix)+8]

	apiKey, err := s.apiKeyRepo.GetByPrefix(ctx, searchPrefix)
	if err != nil {
		return 0, errors.New("authentication failed")
	}

	if !apiKey.IsActive {
		return 0, errors.New("key is revoked")
	}

	err = bcrypt.CompareHashAndPassword([]byte(apiKey.KeyHash), []byte(rawKey))
	if err != nil {
		return 0, errors.New("authentication failed")
	}

	return apiKey.ProfileID, nil

}
