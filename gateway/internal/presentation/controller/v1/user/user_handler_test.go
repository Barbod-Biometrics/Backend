package user

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	authdto "github.com/Barbod-Biometrics/Backend/gateway/internal/application/dto/auth"
	userdto "github.com/Barbod-Biometrics/Backend/gateway/internal/application/dto/user"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/application/service"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/infrastructure/test/mocks"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/presentation/middleware"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetUserProfileHandler_SuccessAndErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)
	// Success
	{
		mockAuth := mocks.NewMockAuthUsecase(t)
		mockUser := mocks.NewMockUserUsecase(t)
		h := NewUserController(mockAuth, mockUser)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set(middleware.ContextKeyUserID, uint64(77))
		mockUser.On("GetUserByID", mock.Anything, uint64(77)).Return(&userdto.UserInfoResponse{PhoneNumber: "0912", Email: "a@b.c"}, nil)
		c.Request = httptest.NewRequest("GET", "/user/info", nil)
		h.GetUserProfileHandler(c)
		assert.Equal(t, http.StatusOK, w.Code)
		var resp userdto.UserInfoResponse
		assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.Equal(t, "0912", resp.PhoneNumber)
		mockUser.AssertExpectations(t)
	}

	// Unauthorized (no user id)
	{
		mockAuth := mocks.NewMockAuthUsecase(t)
		mockUser := mocks.NewMockUserUsecase(t)
		h := NewUserController(mockAuth, mockUser)
		w2 := httptest.NewRecorder()
		c2, _ := gin.CreateTestContext(w2)
		c2.Request = httptest.NewRequest("GET", "/user/info", nil)
		h.GetUserProfileHandler(c2)
		assert.Equal(t, http.StatusUnauthorized, w2.Code)
	}

	// Not found
	{
		mockAuth := mocks.NewMockAuthUsecase(t)
		mockUser := mocks.NewMockUserUsecase(t)
		h := NewUserController(mockAuth, mockUser)
		w3 := httptest.NewRecorder()
		c3, _ := gin.CreateTestContext(w3)
		c3.Set(middleware.ContextKeyUserID, uint64(77))
		mockUser.On("GetUserByID", mock.Anything, uint64(77)).Return((*userdto.UserInfoResponse)(nil), service.ErrUserNotFound)
		c3.Request = httptest.NewRequest("GET", "/user/info", nil)
		h.GetUserProfileHandler(c3)
		assert.Equal(t, http.StatusNotFound, w3.Code)
		mockUser.AssertExpectations(t)
	}

	// Internal error
	{
		mockAuth := mocks.NewMockAuthUsecase(t)
		mockUser := mocks.NewMockUserUsecase(t)
		h := NewUserController(mockAuth, mockUser)
		w4 := httptest.NewRecorder()
		c4, _ := gin.CreateTestContext(w4)
		c4.Set(middleware.ContextKeyUserID, uint64(77))
		mockUser.On("GetUserByID", mock.Anything, uint64(77)).Return((*userdto.UserInfoResponse)(nil), assert.AnError)
		c4.Request = httptest.NewRequest("GET", "/user/info", nil)
		h.GetUserProfileHandler(c4)
		assert.Equal(t, http.StatusInternalServerError, w4.Code)
		mockUser.AssertExpectations(t)
	}
}

func TestUpdateProfileHandler_Cases(t *testing.T) {
	gin.SetMode(gin.TestMode)
	body := `{"phone_number":"09120001111"}`
	// Success
	{
		mockAuth := mocks.NewMockAuthUsecase(t)
		mockUser := mocks.NewMockUserUsecase(t)
		h := NewUserController(mockAuth, mockUser)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set(middleware.ContextKeyUserID, uint64(88))
		body := `{"phone_number":"09120001111"}`
		expected := &userdto.UserInfoResponse{PhoneNumber: "09120001111", Email: "e@f.g"}
		mockUser.On("UpdateProfile", mock.Anything, mock.MatchedBy(func(r userdto.UpdateProfileRequest) bool { return r.UserID == 88 && r.PhoneNumber == "09120001111" })).Return(expected, nil)
		c.Request = httptest.NewRequest("POST", "/user/update-info", strings.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")
		h.UpdateProfileHandler(c)
		assert.Equal(t, http.StatusOK, w.Code)
		mockUser.AssertExpectations(t)
	}

	// Unauthorized
	{
		mockAuth := mocks.NewMockAuthUsecase(t)
		mockUser := mocks.NewMockUserUsecase(t)
		h := NewUserController(mockAuth, mockUser)
		w2 := httptest.NewRecorder()
		c2, _ := gin.CreateTestContext(w2)
		body := `{"phone_number":"09120001111"}`
		c2.Request = httptest.NewRequest("POST", "/user/update-info", strings.NewReader(body))
		c2.Request.Header.Set("Content-Type", "application/json")
		h.UpdateProfileHandler(c2)
		assert.Equal(t, http.StatusUnauthorized, w2.Code)
	}

	// Invalid body
	{
		mockAuth := mocks.NewMockAuthUsecase(t)
		mockUser := mocks.NewMockUserUsecase(t)
		h := NewUserController(mockAuth, mockUser)
		w3 := httptest.NewRecorder()
		c3, _ := gin.CreateTestContext(w3)
		c3.Set(middleware.ContextKeyUserID, uint64(88))
		c3.Request = httptest.NewRequest("POST", "/user/update-info", strings.NewReader(`{"phone_number": 123}`))
		c3.Request.Header.Set("Content-Type", "application/json")
		h.UpdateProfileHandler(c3)
		assert.Equal(t, http.StatusBadRequest, w3.Code)
	}

	// Not found
	{
		mockAuth := mocks.NewMockAuthUsecase(t)
		mockUser := mocks.NewMockUserUsecase(t)
		h := NewUserController(mockAuth, mockUser)
		w4 := httptest.NewRecorder()
		c4, _ := gin.CreateTestContext(w4)
		c4.Set(middleware.ContextKeyUserID, uint64(88))
		mockUser.On("UpdateProfile", mock.Anything, mock.Anything).Return((*userdto.UserInfoResponse)(nil), service.ErrUserNotFound)
		c4.Request = httptest.NewRequest("POST", "/user/update-info", strings.NewReader(body))
		c4.Request.Header.Set("Content-Type", "application/json")
		h.UpdateProfileHandler(c4)
		assert.Equal(t, http.StatusNotFound, w4.Code)
		mockUser.AssertExpectations(t)
	}

	// Email exists
	{
		mockAuth := mocks.NewMockAuthUsecase(t)
		mockUser := mocks.NewMockUserUsecase(t)
		h := NewUserController(mockAuth, mockUser)
		w5 := httptest.NewRecorder()
		c5, _ := gin.CreateTestContext(w5)
		c5.Set(middleware.ContextKeyUserID, uint64(88))
		mockUser.On("UpdateProfile", mock.Anything, mock.Anything).Return((*userdto.UserInfoResponse)(nil), service.ErrEmailAlreadyExist)
		c5.Request = httptest.NewRequest("POST", "/user/update-info", strings.NewReader(body))
		c5.Request.Header.Set("Content-Type", "application/json")
		h.UpdateProfileHandler(c5)
		assert.Equal(t, http.StatusConflict, w5.Code)
		mockUser.AssertExpectations(t)
	}

	// Phone exists
	{
		mockAuth := mocks.NewMockAuthUsecase(t)
		mockUser := mocks.NewMockUserUsecase(t)
		h := NewUserController(mockAuth, mockUser)
		w6 := httptest.NewRecorder()
		c6, _ := gin.CreateTestContext(w6)
		c6.Set(middleware.ContextKeyUserID, uint64(88))
		mockUser.On("UpdateProfile", mock.Anything, mock.Anything).Return((*userdto.UserInfoResponse)(nil), service.ErrPhoneAlreadyExist)
		c6.Request = httptest.NewRequest("POST", "/user/update-info", strings.NewReader(body))
		c6.Request.Header.Set("Content-Type", "application/json")
		h.UpdateProfileHandler(c6)
		assert.Equal(t, http.StatusConflict, w6.Code)
		mockUser.AssertExpectations(t)
	}

	// Internal error
	{
		mockAuth := mocks.NewMockAuthUsecase(t)
		mockUser := mocks.NewMockUserUsecase(t)
		h := NewUserController(mockAuth, mockUser)
		w7 := httptest.NewRecorder()
		c7, _ := gin.CreateTestContext(w7)
		c7.Set(middleware.ContextKeyUserID, uint64(88))
		mockUser.On("UpdateProfile", mock.Anything, mock.Anything).Return((*userdto.UserInfoResponse)(nil), assert.AnError)
		c7.Request = httptest.NewRequest("POST", "/user/update-info", strings.NewReader(body))
		c7.Request.Header.Set("Content-Type", "application/json")
		h.UpdateProfileHandler(c7)
		assert.Equal(t, http.StatusInternalServerError, w7.Code)
		mockUser.AssertExpectations(t)
	}
}

func TestAuthHandlers_RequestVerifyOTP(t *testing.T) {
	gin.SetMode(gin.TestMode)
	// Request OTP success
	{
		mockAuth := mocks.NewMockAuthUsecase(t)
		mockUser := mocks.NewMockUserUsecase(t)
		h := NewUserController(mockAuth, mockUser)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		body := `{"phone_number":"09120001111"}`
		mockAuth.On("RequestOTP", mock.Anything, authdto.RequestOTPRequest{PhoneNumber: "09120001111"}).Return(nil)
		c.Request = httptest.NewRequest("POST", "/auth/request-otp", strings.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")
		h.RequestOTPHandler(c)
		assert.Equal(t, http.StatusOK, w.Code)
		mockAuth.AssertExpectations(t)
	}

	// Request OTP bind error
	{
		mockAuth := mocks.NewMockAuthUsecase(t)
		mockUser := mocks.NewMockUserUsecase(t)
		h := NewUserController(mockAuth, mockUser)
		w2 := httptest.NewRecorder()
		c2, _ := gin.CreateTestContext(w2)
		c2.Request = httptest.NewRequest("POST", "/auth/request-otp", strings.NewReader(`{"phone_number": 111}`))
		c2.Request.Header.Set("Content-Type", "application/json")
		h.RequestOTPHandler(c2)
		assert.Equal(t, http.StatusBadRequest, w2.Code)
	}

	// Request OTP usecase error
	{
		mockAuth := mocks.NewMockAuthUsecase(t)
		mockUser := mocks.NewMockUserUsecase(t)
		h := NewUserController(mockAuth, mockUser)
		w3 := httptest.NewRecorder()
		c3, _ := gin.CreateTestContext(w3)
		mockAuth.On("RequestOTP", mock.Anything, mock.Anything).Return(assert.AnError)
		body := `{"phone_number":"09120001111"}`
		c3.Request = httptest.NewRequest("POST", "/auth/request-otp", strings.NewReader(body))
		c3.Request.Header.Set("Content-Type", "application/json")
		h.RequestOTPHandler(c3)
		assert.Equal(t, http.StatusInternalServerError, w3.Code)
		mockAuth.AssertExpectations(t)
	}
}

func TestAuthHandlers_VerifyOTP(t *testing.T) {
	gin.SetMode(gin.TestMode)
	// Verify OTP success
	{
		mockAuth := mocks.NewMockAuthUsecase(t)
		mockUser := mocks.NewMockUserUsecase(t)
		h := NewUserController(mockAuth, mockUser)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		body := `{"phone_number":"09120001111","otp":"123456"}`
		mockAuth.On("VerifyOTP", mock.Anything, authdto.VerifyOTPRequest{PhoneNumber: "09120001111", OTP: "123456"}).Return(&authdto.UserInfoResponse{AccessToken: "at", RefereshToken: "rt", PhoneNumber: "09120001111", IsAdmin: false}, nil)
		c.Request = httptest.NewRequest("POST", "/auth/verify-otp", strings.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")
		h.VerifyOTPHandler(c)
		assert.Equal(t, http.StatusOK, w.Code)
		mockAuth.AssertExpectations(t)
	}

	// bind error
	{
		mockAuth := mocks.NewMockAuthUsecase(t)
		mockUser := mocks.NewMockUserUsecase(t)
		h := NewUserController(mockAuth, mockUser)
		w2 := httptest.NewRecorder()
		c2, _ := gin.CreateTestContext(w2)
		c2.Request = httptest.NewRequest("POST", "/auth/verify-otp", strings.NewReader(`{"phone_number":"x"}`))
		c2.Request.Header.Set("Content-Type", "application/json")
		h.VerifyOTPHandler(c2)
		assert.Equal(t, http.StatusBadRequest, w2.Code)
	}

	// verify error -> unauthorized
	{
		mockAuth := mocks.NewMockAuthUsecase(t)
		mockUser := mocks.NewMockUserUsecase(t)
		h := NewUserController(mockAuth, mockUser)
		w3 := httptest.NewRecorder()
		c3, _ := gin.CreateTestContext(w3)
		mockAuth.On("VerifyOTP", mock.Anything, mock.Anything).Return((*authdto.UserInfoResponse)(nil), assert.AnError)
		body := `{"phone_number":"09120001111","otp":"123456"}`
		c3.Request = httptest.NewRequest("POST", "/auth/verify-otp", strings.NewReader(body))
		c3.Request.Header.Set("Content-Type", "application/json")
		h.VerifyOTPHandler(c3)
		assert.Equal(t, http.StatusUnauthorized, w3.Code)
		mockAuth.AssertExpectations(t)
	}
}
