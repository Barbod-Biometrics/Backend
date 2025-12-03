package mocks

import (
	"context"

	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/entity"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/logger"
	"github.com/stretchr/testify/mock"
)

// --- Mock Repo ---
type MockAPIKeyRepository struct {
	mock.Mock
}

func (m *MockAPIKeyRepository) Create(ctx context.Context, apiKey *entity.APIKey) error {
	args := m.Called(ctx, apiKey)
	return args.Error(0)
}

func (m *MockAPIKeyRepository) GetByPrefix(ctx context.Context, prefix string) (*entity.APIKey, error) {
	args := m.Called(ctx, prefix)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.APIKey), args.Error(1)
}

func (m *MockAPIKeyRepository) GetActiveByProfileID(ctx context.Context, profileID uint64) (*entity.APIKey, error) {
	args := m.Called(ctx, profileID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.APIKey), args.Error(1)
}

func (m *MockAPIKeyRepository) Revoke(ctx context.Context, keyID string) error {
	args := m.Called(ctx, keyID)
	return args.Error(0)
}

// --- Mock Service ---
type MockAPIKeyService struct {
	mock.Mock
}

func (m *MockAPIKeyService) GenerateKey(ctx context.Context, profileID uint64) (string, error) {
	args := m.Called(ctx, profileID)
	return args.String(0), args.Error(1)
}

func (m *MockAPIKeyService) RegenerateKey(ctx context.Context, profileID uint64) (string, error) {
	args := m.Called(ctx, profileID)
	return args.String(0), args.Error(1)
}

func (m *MockAPIKeyService) Authenticate(ctx context.Context, rawKey string) (uint64, error) {
	args := m.Called(ctx, rawKey)
	return args.Get(0).(uint64), args.Error(1)
}

// --- NoOp Logger (Use this in tests to ignore logs!) ---
type NoOpLogger struct{}

func (l *NoOpLogger) Info(msg string, fields ...logger.Field)                {}
func (l *NoOpLogger) Warn(msg string, fields ...logger.Field)                {}
func (l *NoOpLogger) Error(msg string, fields ...logger.Field)               {}
func (l *NoOpLogger) Debug(msg string, fields ...logger.Field)               {}
func (l *NoOpLogger) Fatal(msg string, fields ...logger.Field)               {}
func (l *NoOpLogger) WithFields(fields map[string]interface{}) logger.Logger { return l }
func (l *NoOpLogger) Close()                                                 {}
