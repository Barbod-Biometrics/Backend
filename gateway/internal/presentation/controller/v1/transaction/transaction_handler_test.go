package transaction

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Barbod-Biometrics/Backend/gateway/internal/application/dto/profile"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/entity"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/infrastructure/test/mocks"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/presentation/middleware"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestGetUsageSummary(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockUC := mocks.NewMockTransactionUsecase(t)
		mockProfileRepo := mocks.NewMockProfileRepository(t)
		mockLogger := mocks.NewMockAppLogger(t)
		h := NewTransactionHandler(mockUC, mockProfileRepo, mockLogger)

		userID := uint64(100)
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

		mockProfileRepo.On("GetByUserID", mock.Anything, userID).Return([]*entity.Profile{
			{ProfileID: profileID, UserID: userID},
		}, nil)
		mockUC.On("GetUsageSummary", mock.Anything, profileID).Return(expectedSummary, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set(middleware.ContextKeyUserID, userID)
		c.Request = httptest.NewRequest("GET", "/test?profile_id=1", nil)

		h.GetUsageSummary(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var actualSummary profile.UsageSummaryResponse
		err := json.Unmarshal(w.Body.Bytes(), &actualSummary)
		assert.NoError(t, err)
		assert.Equal(t, expectedSummary.TotalSpend, actualSummary.TotalSpend)
		assert.Len(t, actualSummary.ServiceBreakdown, 1)
	})

	t.Run("invalid request parameters - missing profile_id", func(t *testing.T) {
		mockUC := mocks.NewMockTransactionUsecase(t)
		mockProfileRepo := mocks.NewMockProfileRepository(t)
		mockLogger := mocks.NewMockAppLogger(t)
		h := NewTransactionHandler(mockUC, mockProfileRepo, mockLogger)

		userID := uint64(100)
		mockLogger.On("Warn", "invalid usage summary request format", mock.Anything).Return()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set(middleware.ContextKeyUserID, userID)
		c.Request = httptest.NewRequest("GET", "/test", nil)

		h.GetUsageSummary(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "invalid request parameters")
	})

	t.Run("invalid request parameters - non-numeric profile_id", func(t *testing.T) {
		mockUC := mocks.NewMockTransactionUsecase(t)
		mockProfileRepo := mocks.NewMockProfileRepository(t)
		mockLogger := mocks.NewMockAppLogger(t)
		h := NewTransactionHandler(mockUC, mockProfileRepo, mockLogger)

		userID := uint64(100)
		mockLogger.On("Warn", "invalid usage summary request format", mock.Anything).Return()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set(middleware.ContextKeyUserID, userID)
		c.Request = httptest.NewRequest("GET", "/test?profile_id=abc", nil)

		h.GetUsageSummary(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "invalid request parameters")
	})

	t.Run("usecase error", func(t *testing.T) {
		mockUC := mocks.NewMockTransactionUsecase(t)
		mockProfileRepo := mocks.NewMockProfileRepository(t)
		mockLogger := mocks.NewMockAppLogger(t)
		h := NewTransactionHandler(mockUC, mockProfileRepo, mockLogger)

		userID := uint64(100)
		profileID := uint64(1)

		mockProfileRepo.On("GetByUserID", mock.Anything, userID).Return([]*entity.Profile{
			{ProfileID: profileID, UserID: userID},
		}, nil)
		mockUC.On("GetUsageSummary", mock.Anything, profileID).Return(nil, errors.New("internal error"))
		mockLogger.On("Error", "failed to retrieve usage summary", mock.Anything, mock.Anything).Return()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set(middleware.ContextKeyUserID, userID)
		c.Request = httptest.NewRequest("GET", "/test?profile_id=1", nil)

		h.GetUsageSummary(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "failed to retrieve usage summary")
	})
	t.Run("missing user_id in context", func(t *testing.T) {
		mockUC := mocks.NewMockTransactionUsecase(t)
		mockProfileRepo := mocks.NewMockProfileRepository(t)
		mockLogger := mocks.NewMockAppLogger(t)
		h := NewTransactionHandler(mockUC, mockProfileRepo, mockLogger)

		mockLogger.On("Warn", "failed to extract user_id from context", mock.Anything).Return()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/test?profile_id=1", nil)

		h.GetUsageSummary(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("unauthorized profile access", func(t *testing.T) {
		mockUC := mocks.NewMockTransactionUsecase(t)
		mockProfileRepo := mocks.NewMockProfileRepository(t)
		mockLogger := mocks.NewMockAppLogger(t)
		h := NewTransactionHandler(mockUC, mockProfileRepo, mockLogger)

		userID := uint64(100)
		profileID := uint64(1)

		// Return profiles that don't include the requested profileID
		mockProfileRepo.On("GetByUserID", mock.Anything, userID).Return([]*entity.Profile{
			{ProfileID: 999, UserID: userID},
		}, nil)
		mockLogger.On("Warn", "unauthorized profile access attempt", mock.Anything, mock.Anything).Return()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set(middleware.ContextKeyUserID, userID)
		c.Request = httptest.NewRequest("GET", fmt.Sprintf("/test?profile_id=%d", profileID), nil)

		h.GetUsageSummary(c)

		assert.Equal(t, http.StatusForbidden, w.Code)
		assert.Contains(t, w.Body.String(), "access denied to this profile")
	})
}
