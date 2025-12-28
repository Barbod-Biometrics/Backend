package service_test

import (
	"context"
	"testing"

	"github.com/Barbod-Biometrics/Backend/gateway/internal/application/service"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/entity"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/infrastructure/test/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

func TestApiKeyService_GenerateKey(t *testing.T) {
	mockRepo := new(mocks.MockAPIKeyRepository)
	mockProfileRepo := new(mocks.MockProfileRepository)

	mockAppLogger := new(mocks.MockAppLogger)

	svc := service.NewAPIKeyService(mockRepo, mockProfileRepo, mockAppLogger)
	ctx := context.Background()
	profileID := uint64(12345)

	mockRepo.On("Create", ctx, mock.AnythingOfType("*entity.APIKey")).Return(nil)
	mockAppLogger.On("Info", mock.Anything, mock.Anything).Return()

	rawKey, err := svc.GenerateKey(ctx, profileID)

	assert.NoError(t, err)
	assert.Contains(t, rawKey, "bb_live_")
	assert.Len(t, rawKey, 72)

	mockRepo.AssertExpectations(t)
}

func TestAPIKeyService_RegenerateKey(t *testing.T) {
	mockRepo := new(mocks.MockAPIKeyRepository)
	mockProfileRepo := new(mocks.MockProfileRepository)
	mockAppLogger := new(mocks.MockAppLogger)
	svc := service.NewAPIKeyService(mockRepo, mockProfileRepo, mockAppLogger)

	ctx := context.Background()
	profileID := uint64(10)
	existingKey := &entity.APIKey{KeyID: 555, ProfileID: 10, IsActive: true}

	mockRepo.On("GetActiveByProfileID", ctx, profileID).Return(existingKey, nil)
	mockAppLogger.On("Info", mock.Anything, mock.Anything).Return()

	mockRepo.On("Revoke", ctx, "555").Return(nil)

	mockRepo.On("Create", ctx, mock.AnythingOfType("*entity.APIKey")).Return(nil)

	rawKey, err := svc.RegenerateKey(ctx, profileID)

	assert.NoError(t, err)
	assert.NotEmpty(t, rawKey)
}

func TestAPIKeyService_Authenticate_Success(t *testing.T) {
	mockRepo := new(mocks.MockAPIKeyRepository)
	mockProfileRepo := new(mocks.MockProfileRepository)
	mockAppLogger := new(mocks.MockAppLogger)
	svc := service.NewAPIKeyService(mockRepo, mockProfileRepo, mockAppLogger)

	rawKey := "bb_live_a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6q7r8s9t0u1v2w3x4y5z6a7b8c9d0e1f2"
	dbPrefix := "a1b2c3d4"
	hash, _ := bcrypt.GenerateFromPassword([]byte(rawKey), bcrypt.DefaultCost)

	mockEntity := &entity.APIKey{
		ProfileID: 100,
		KeyHash:   string(hash),
		KeyPrefix: dbPrefix,
		IsActive:  true,
	}

	mockRepo.On("GetByPrefix", context.Background(), dbPrefix).Return(mockEntity, nil)
	mockAppLogger.On("Debug", mock.Anything, mock.Anything).Return()

	pid, err := svc.Authenticate(context.Background(), rawKey)

	assert.NoError(t, err)
	assert.Equal(t, uint64(100), pid)
}

func TestAPIKeyService_Authenticate_InvalidFormat(t *testing.T) {
	mockRepo := new(mocks.MockAPIKeyRepository)
	mockProfileRepo := new(mocks.MockProfileRepository)
	mockAppLogger := new(mocks.MockAppLogger)
	svc := service.NewAPIKeyService(mockRepo, mockProfileRepo, mockAppLogger)

	rawKey := "bad_format"

	mockAppLogger.On("Warn", mock.Anything, mock.Anything).Return()

	pid, err := svc.Authenticate(context.Background(), rawKey)

	assert.Error(t, err)
	assert.Equal(t, uint64(0), pid)
	assert.Equal(t, "invalid key format", err.Error())
}

func TestAPIKeyService_Authenticate_Revoked(t *testing.T) {
	mockRepo := new(mocks.MockAPIKeyRepository)
	mockProfileRepo := new(mocks.MockProfileRepository)
	mockAppLogger := new(mocks.MockAppLogger)
	svc := service.NewAPIKeyService(mockRepo, mockProfileRepo, mockAppLogger)

	rawKey := "bb_live_a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6q7r8s9t0u1v2w3x4y5z6a7b8c9d0e1f2"
	dbPrefix := "a1b2c3d4"

	mockEntity := &entity.APIKey{IsActive: false, ProfileID: 100}

	mockRepo.On("GetByPrefix", context.Background(), dbPrefix).Return(mockEntity, nil)
	mockAppLogger.On("Warn", mock.Anything, mock.Anything).Return()

	_, err := svc.Authenticate(context.Background(), rawKey)

	assert.Error(t, err)
	assert.Equal(t, "key is revoked", err.Error())
}
