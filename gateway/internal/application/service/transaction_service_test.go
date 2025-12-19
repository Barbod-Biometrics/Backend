package service

import (
	"context"
	"errors"
	"testing"

	"github.com/Barbod-Biometrics/Backend/gateway/internal/application/dto/profile"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/infrastructure/test/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetUsageSummary_Success(t *testing.T) {
	ctx := context.Background()
	mockTxRepo := mocks.NewMockTransactionRepository(t)
	mockLogger := mocks.NewMockAppLogger(t)

	s := NewTransactionService(mockTxRepo, mockLogger)

	profileID := uint64(1)
	expectedSummary := &profile.UsageSummaryResponse{
		TotalSpend: 5000,
		ServiceBreakdown: []profile.ServiceUsageStats{
			{
				ServiceName: "Verification",
				TotalCount:  10,
				TotalCost:   5000,
			},
		},
	}

	mockLogger.On("Info", "processing usage summary request", mock.Anything).Return()
	mockTxRepo.On("GetUsageSummary", ctx, profileID).Return(expectedSummary, nil)
	mockLogger.On("Info", "usage summary retrieved successfully", mock.Anything).Return()

	resp, err := s.GetUsageSummary(ctx, profileID)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, expectedSummary.TotalSpend, resp.TotalSpend)
	assert.Len(t, resp.ServiceBreakdown, 1)
	assert.Equal(t, expectedSummary.ServiceBreakdown[0].ServiceName, resp.ServiceBreakdown[0].ServiceName)

	mockTxRepo.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

func TestGetUsageSummary_RepositoryError(t *testing.T) {
	ctx := context.Background()
	mockTxRepo := mocks.NewMockTransactionRepository(t)
	mockLogger := mocks.NewMockAppLogger(t)

	s := NewTransactionService(mockTxRepo, mockLogger)

	profileID := uint64(1)
	repoErr := errors.New("repository error")

	mockLogger.On("Info", "processing usage summary request", mock.Anything).Return()
	mockTxRepo.On("GetUsageSummary", ctx, profileID).Return(nil, repoErr)
	mockLogger.On("Error", "failed to retrieve usage summary from repository", mock.Anything).Return()

	resp, err := s.GetUsageSummary(ctx, profileID)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, repoErr, err)

	mockTxRepo.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}
