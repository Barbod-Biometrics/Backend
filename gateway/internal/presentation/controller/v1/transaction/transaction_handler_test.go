package transaction

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Barbod-Biometrics/Backend/gateway/internal/application/dto/profile"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/infrastructure/test/mocks"
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
		mockLogger := mocks.NewMockAppLogger(t)
		h := NewTransactionHandler(mockUC, mockLogger)

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

		mockUC.On("GetUsageSummary", mock.Anything, uint64(1)).Return(expectedSummary, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
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
		mockLogger := mocks.NewMockAppLogger(t)
		h := NewTransactionHandler(mockUC, mockLogger)

		mockLogger.On("Warn", "invalid usage summary request format", mock.Anything).Return()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/test", nil)

		h.GetUsageSummary(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "invalid request parameters")
	})

	t.Run("invalid request parameters - non-numeric profile_id", func(t *testing.T) {
		mockUC := mocks.NewMockTransactionUsecase(t)
		mockLogger := mocks.NewMockAppLogger(t)
		h := NewTransactionHandler(mockUC, mockLogger)

		mockLogger.On("Warn", "invalid usage summary request format", mock.Anything).Return()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/test?profile_id=abc", nil)

		h.GetUsageSummary(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "invalid request parameters")
	})

	t.Run("usecase error", func(t *testing.T) {
		mockUC := mocks.NewMockTransactionUsecase(t)
		mockLogger := mocks.NewMockAppLogger(t)
		h := NewTransactionHandler(mockUC, mockLogger)

		mockUC.On("GetUsageSummary", mock.Anything, uint64(1)).Return(nil, errors.New("internal error"))

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/test?profile_id=1", nil)

		h.GetUsageSummary(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "failed to retrieve usage summary")
	})
}
