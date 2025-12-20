package middleware_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Barbod-Biometrics/Backend/gateway/internal/infrastructure/test/mocks"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/presentation/middleware"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAPIKeyAuth_Success(t *testing.T) {
	mockUsecase := new(mocks.MockAPIKeyUsecase)
	mockLogger := new(mocks.MockAppLogger)

	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(middleware.APIKeyAuth(mockUsecase, mockLogger))

	r.GET("/test", func(c *gin.Context) {
		pid, _ := c.Get("profile_id")
		c.JSON(200, gin.H{"profile_id": pid})
	})

	apiKey := "bb_live_validkey123"
	mockUsecase.On("Authenticate", mock.Anything, apiKey).Return(uint64(55), nil)
	mockLogger.On("Debug", mock.Anything, mock.Anything).Return()

	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("X-API-KEY", apiKey)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Contains(t, w.Body.String(), `"profile_id":55`)
}

func TestAPIKeyAuth_MissingKey(t *testing.T) {
	mockUsecase := new(mocks.MockAPIKeyUsecase)
	mockLogger := new(mocks.MockAppLogger)

	r := gin.New()
	r.Use(middleware.APIKeyAuth(mockUsecase, mockLogger))
	r.GET("/test", func(c *gin.Context) {})

	mockLogger.On("Warn", mock.Anything, mock.Anything).Return()

	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, 401, w.Code)
}

func TestAPIKeyAuth_InvalidKey(t *testing.T) {
	mockUsecase := new(mocks.MockAPIKeyUsecase)
	mockLogger := new(mocks.MockAppLogger)

	r := gin.New()
	r.Use(middleware.APIKeyAuth(mockUsecase, mockLogger))
	r.GET("/test", func(c *gin.Context) {})

	apiKey := "bb_live_badkey"
	mockUsecase.On("Authenticate", mock.Anything, apiKey).Return(uint64(0), errors.New("auth failed"))
	mockLogger.On("Warn", mock.Anything, mock.Anything).Return()

	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("X-API-KEY", apiKey)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, 401, w.Code)
}
