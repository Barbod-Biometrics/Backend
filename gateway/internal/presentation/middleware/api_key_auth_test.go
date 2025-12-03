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
	mockService := new(mocks.MockAPIKeyService)
	noOpLogger := new(mocks.NoOpLogger)

	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(middleware.APIKeyAuth(mockService, noOpLogger))

	r.GET("/test", func(c *gin.Context) {
		pid, _ := c.Get("profile_id")
		c.JSON(200, gin.H{"profile_id": pid})
	})

	apiKey := "bb_live_validkey123"
	mockService.On("Authenticate", mock.Anything, apiKey).Return(uint64(55), nil)

	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("X-API-KEY", apiKey)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Contains(t, w.Body.String(), `"profile_id":55`)
}

func TestAPIKeyAuth_MissingKey(t *testing.T) {
	mockService := new(mocks.MockAPIKeyService)
	noOpLogger := new(mocks.NoOpLogger)

	r := gin.New()
	r.Use(middleware.APIKeyAuth(mockService, noOpLogger))
	r.GET("/test", func(c *gin.Context) {})

	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, 401, w.Code)
}

func TestAPIKeyAuth_InvalidKey(t *testing.T) {
	mockService := new(mocks.MockAPIKeyService)
	noOpLogger := new(mocks.NoOpLogger)

	r := gin.New()
	r.Use(middleware.APIKeyAuth(mockService, noOpLogger))
	r.GET("/test", func(c *gin.Context) {})

	apiKey := "bb_live_badkey"
	mockService.On("Authenticate", mock.Anything, apiKey).Return(uint64(0), errors.New("auth failed"))

	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("X-API-KEY", apiKey)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, 401, w.Code)
}
