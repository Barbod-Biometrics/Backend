package apikey

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Barbod-Biometrics/Backend/gateway/internal/application/dto/business"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/infrastructure/test/mocks"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRegenerateKey_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockUC := mocks.NewMockAPIKeyUsecase(t)
	mockLogger := mocks.NewMockLogger(t)
	h := NewApiKeyHandler(mockUC, mockLogger)

	r := gin.New()
	r.POST("/api-key/:profile_id/regenerate", h.RegenerateKey)

	// expect logger.Info called
	mockLogger.On("Info", "received regenerate api key request", mock.Anything).Return()
	mockLogger.On("Info", "api key regenerated successfully", mock.Anything).Return()

	mockUC.On("RegenerateKey", mock.Anything, uint64(77)).Return("new-key", nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api-key/77/regenerate", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp business.NewAPIKeyResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, uint64(77), resp.ProfileID)
	assert.Equal(t, "new-key", resp.APIKey)

	mockUC.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

func TestRegenerateKey_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockUC := mocks.NewMockAPIKeyUsecase(t)
	mockLogger := mocks.NewMockLogger(t)
	h := NewApiKeyHandler(mockUC, mockLogger)

	r := gin.New()
	r.POST("/api-key/:profile_id/regenerate", h.RegenerateKey)

	// expect info then a warn log for invalid id
	mockLogger.On("Info", "received regenerate api key request", mock.Anything).Return()
	mockLogger.On("Warn", "regenerate request failed: invalid profile_id", mock.Anything).Return()

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api-key/zz/regenerate", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	mockLogger.AssertExpectations(t)
}

func TestRegenerateKey_UsecaseError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockUC := mocks.NewMockAPIKeyUsecase(t)
	mockLogger := mocks.NewMockLogger(t)
	h := NewApiKeyHandler(mockUC, mockLogger)

	r := gin.New()
	r.POST("/api-key/:profile_id/regenerate", h.RegenerateKey)

	mockLogger.On("Info", "received regenerate api key request", mock.Anything).Return()
	mockUC.On("RegenerateKey", mock.Anything, uint64(99)).Return("", assert.AnError)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api-key/99/regenerate", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	mockUC.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}
