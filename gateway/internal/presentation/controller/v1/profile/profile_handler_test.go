package profile

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	dto "github.com/Barbod-Biometrics/Backend/gateway/internal/application/dto/profile"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/infrastructure/test/mocks"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/presentation/middleware"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreateDraft_SuccessAndErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockUC := mocks.NewMockProfileUsecase(t)
	h := NewProfileHandler(mockUC)

	// create ctx and set user id
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set(middleware.ContextKeyUserID, uint64(123))
	// quick sanity: verify handler middleware helper reads the ID
	uid1, err := middleware.GetUserIDFromContext(c)
	assert.NoError(t, err)
	assert.Equal(t, uint64(123), uid1)

	// success
	reqBody := `{"profile_type":"personal","profile_name":"John's"}`
	expectedResp := &dto.ProfileResponse{ID: "10", Name: "John's"}
	mockUC.On("CreateDraft", mock.Anything, mock.Anything, mock.MatchedBy(func(r dto.CreateProfileRequest) bool {
		return r.ProfileType == "personal" && r.ProfileName == "John's"
	})).Return(expectedResp, nil)

	c.Request = httptest.NewRequest("POST", "/profiles", strings.NewReader(reqBody))
	c.Request.Header.Set("Content-Type", "application/json")
	h.CreateDraft(c)
	assert.Equal(t, http.StatusCreated, w.Code)
	var got dto.ProfileResponse
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	assert.Equal(t, expectedResp.ID, got.ID)

	mockUC.AssertExpectations(t)

	// invalid JSON -> bad request
	w2 := httptest.NewRecorder()
	c2, _ := gin.CreateTestContext(w2)
	c2.Set(middleware.ContextKeyUserID, uint64(123))
	uid2, err2 := middleware.GetUserIDFromContext(c2)
	assert.NoError(t, err2)
	assert.Equal(t, uint64(123), uid2)
	c2.Request = httptest.NewRequest("POST", "/profiles", strings.NewReader(`{"profile_type":"personal"`))
	c2.Request.Header.Set("Content-Type", "application/json")
	h.CreateDraft(c2)
	assert.Equal(t, http.StatusBadRequest, w2.Code)

	// unauthorized (no user id set)
	w3 := httptest.NewRecorder()
	c3, _ := gin.CreateTestContext(w3)
	c3.Request = httptest.NewRequest("POST", "/profiles", strings.NewReader(reqBody))
	c3.Request.Header.Set("Content-Type", "application/json")
	h.CreateDraft(c3)
	assert.Equal(t, http.StatusUnauthorized, w3.Code)
}

func TestUpdateDraft_SuccessAndErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockUC := mocks.NewMockProfileUsecase(t)
	h := NewProfileHandler(mockUC)

	// success
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set(middleware.ContextKeyUserID, uint64(42))
	uid3, err3 := middleware.GetUserIDFromContext(c)
	assert.NoError(t, err3)
	assert.Equal(t, uint64(42), uid3)
	body := `{"profile_name":"NewName"}`
	mockUC.On("UpdateDraft", mock.Anything, mock.Anything, uint64(99), mock.Anything).Return(&dto.ProfileResponse{ID: "99", Name: "NewName"}, nil)
	c.Request = httptest.NewRequest("PATCH", "/profiles/99", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "id", Value: "99"}}
	h.UpdateDraft(c)
	assert.Equal(t, http.StatusOK, w.Code)

	// invalid profile id
	w2 := httptest.NewRecorder()
	c2, _ := gin.CreateTestContext(w2)
	c2.Set(middleware.ContextKeyUserID, uint64(42))
	uid4, err4 := middleware.GetUserIDFromContext(c2)
	assert.NoError(t, err4)
	assert.Equal(t, uint64(42), uid4)
	c2.Request = httptest.NewRequest("PATCH", "/profiles/notanumber", strings.NewReader(body))
	c2.Request.Header.Set("Content-Type", "application/json")
	c2.Params = gin.Params{{Key: "id", Value: "notanumber"}}
	h.UpdateDraft(c2)
	assert.Equal(t, http.StatusBadRequest, w2.Code)

	// bind error
	w3 := httptest.NewRecorder()
	c3, _ := gin.CreateTestContext(w3)
	c3.Set(middleware.ContextKeyUserID, uint64(42))
	uid5, err5 := middleware.GetUserIDFromContext(c3)
	assert.NoError(t, err5)
	assert.Equal(t, uint64(42), uid5)
	c3.Request = httptest.NewRequest("PATCH", "/profiles/99", strings.NewReader(`{"profile_name": "a"}`))
	c3.Request.Header.Set("Content-Type", "application/json")
	c3.Params = gin.Params{{Key: "id", Value: "99"}}
	h.UpdateDraft(c3)
	assert.Equal(t, http.StatusBadRequest, w3.Code)

	// unauthorized
	w4 := httptest.NewRecorder()
	c4, _ := gin.CreateTestContext(w4)
	c4.Request = httptest.NewRequest("PATCH", "/profiles/99", strings.NewReader(body))
	c4.Request.Header.Set("Content-Type", "application/json")
	c4.Params = gin.Params{{Key: "id", Value: "99"}}
	h.UpdateDraft(c4)
	assert.Equal(t, http.StatusUnauthorized, w4.Code)
}

func TestSaveDocument_SuccessAndErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockUC := mocks.NewMockProfileUsecase(t)
	h := NewProfileHandler(mockUC)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set(middleware.ContextKeyUserID, uint64(7))
	uid6, err6 := middleware.GetUserIDFromContext(c)
	assert.NoError(t, err6)
	assert.Equal(t, uint64(7), uid6)

	// success
	body := `{"document_type":"national_card_front","file_url":"https://example.com/file.jpg"}`
	mockUC.On("SaveDocument", mock.Anything, mock.Anything, uint64(11), mock.MatchedBy(func(r dto.SaveDocumentRequest) bool { return r.DocumentType == "national_card_front" })).Return(nil)
	c.Request = httptest.NewRequest("POST", "/profiles/11/documents", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "id", Value: "11"}}
	h.SaveDocument(c)
	assert.Equal(t, http.StatusOK, w.Code)

	// invalid id
	w2 := httptest.NewRecorder()
	c2, _ := gin.CreateTestContext(w2)
	c2.Set(middleware.ContextKeyUserID, uint64(7))
	uid7, err7 := middleware.GetUserIDFromContext(c2)
	assert.NoError(t, err7)
	assert.Equal(t, uint64(7), uid7)
	c2.Request = httptest.NewRequest("POST", "/profiles/abc/documents", strings.NewReader(body))
	c2.Request.Header.Set("Content-Type", "application/json")
	c2.Params = gin.Params{{Key: "id", Value: "abc"}}
	h.SaveDocument(c2)
	assert.Equal(t, http.StatusBadRequest, w2.Code)

	// bind error
	w3 := httptest.NewRecorder()
	c3, _ := gin.CreateTestContext(w3)
	c3.Set(middleware.ContextKeyUserID, uint64(7))
	uid8, err8 := middleware.GetUserIDFromContext(c3)
	assert.NoError(t, err8)
	assert.Equal(t, uint64(7), uid8)
	c3.Request = httptest.NewRequest("POST", "/profiles/11/documents", strings.NewReader(`{"file_url":"not-a-url"}`))
	c3.Request.Header.Set("Content-Type", "application/json")
	c3.Params = gin.Params{{Key: "id", Value: "11"}}
	h.SaveDocument(c3)
	assert.Equal(t, http.StatusBadRequest, w3.Code)

	// usecase error -> 500
	mockUC.On("SaveDocument", mock.Anything, mock.Anything, uint64(12), mock.Anything).Return(assert.AnError)
	w4 := httptest.NewRecorder()
	c4, _ := gin.CreateTestContext(w4)
	c4.Set(middleware.ContextKeyUserID, uint64(7))
	c4.Request = httptest.NewRequest("POST", "/profiles/12/documents", strings.NewReader(body))
	c4.Request.Header.Set("Content-Type", "application/json")
	c4.Params = gin.Params{{Key: "id", Value: "12"}}
	h.SaveDocument(c4)
	assert.Equal(t, http.StatusInternalServerError, w4.Code)

	// unauthorized
	w5 := httptest.NewRecorder()
	c5, _ := gin.CreateTestContext(w5)
	c5.Request = httptest.NewRequest("POST", "/profiles/11/documents", strings.NewReader(body))
	c5.Request.Header.Set("Content-Type", "application/json")
	c5.Params = gin.Params{{Key: "id", Value: "11"}}
	h.SaveDocument(c5)
	assert.Equal(t, http.StatusUnauthorized, w5.Code)
}

func TestSubmitAndGetProfile_GetUpload_GetUserProfiles(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockUC := mocks.NewMockProfileUsecase(t)
	h := NewProfileHandler(mockUC)

	// Submit success
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set(middleware.ContextKeyUserID, uint64(99))
	uid9, err9 := middleware.GetUserIDFromContext(c)
	assert.NoError(t, err9)
	assert.Equal(t, uint64(99), uid9)
	mockUC.On("SubmitProfile", mock.Anything, mock.Anything, uint64(25)).Return(&dto.ProfileResponse{ID: "25"}, nil)
	c.Request = httptest.NewRequest("POST", "/profiles/25/submit", nil)
	c.Params = gin.Params{{Key: "id", Value: "25"}}
	h.Submit(c)
	assert.Equal(t, http.StatusOK, w.Code)

	// Submit invalid id
	w2 := httptest.NewRecorder()
	c2, _ := gin.CreateTestContext(w2)
	c2.Set(middleware.ContextKeyUserID, uint64(99))
	uid10, err10 := middleware.GetUserIDFromContext(c2)
	assert.NoError(t, err10)
	assert.Equal(t, uint64(99), uid10)
	c2.Request = httptest.NewRequest("POST", "/profiles/zz/submit", nil)
	c2.Params = gin.Params{{Key: "id", Value: "zz"}}
	h.Submit(c2)
	assert.Equal(t, http.StatusBadRequest, w2.Code)

	// Get profile success
	w3 := httptest.NewRecorder()
	c3, _ := gin.CreateTestContext(w3)
	c3.Set(middleware.ContextKeyUserID, uint64(99))
	uid11, err11 := middleware.GetUserIDFromContext(c3)
	assert.NoError(t, err11)
	assert.Equal(t, uint64(99), uid11)
	mockUC.On("GetByID", mock.Anything, mock.Anything, uint64(25)).Return(&dto.ProfileResponse{ID: "25"}, nil)
	c3.Request = httptest.NewRequest("GET", "/profiles/25", nil)
	c3.Params = gin.Params{{Key: "id", Value: "25"}}
	h.GetProfile(c3)
	assert.Equal(t, http.StatusOK, w3.Code)

	// Get upload url success
	w4 := httptest.NewRecorder()
	c4, _ := gin.CreateTestContext(w4)
	c4.Set(middleware.ContextKeyUserID, uint64(12))
	uid12, err12 := middleware.GetUserIDFromContext(c4)
	assert.NoError(t, err12)
	assert.Equal(t, uint64(12), uid12)
	mockUC.On("GetUploadUrl", mock.Anything, mock.Anything, mock.Anything).Return(&dto.UploadUrlResponse{UploadUrl: "u", FileKey: "k", ExpiresAt: "soon"}, nil)
	c4.Request = httptest.NewRequest("POST", "/profiles/upload-url", strings.NewReader(`{"document_type":"national_card_front","file_extension":".jpg"}`))
	c4.Request.Header.Set("Content-Type", "application/json")
	h.GetUploadUrl(c4)
	assert.Equal(t, http.StatusOK, w4.Code)

	// Get user profiles success
	w5 := httptest.NewRecorder()
	c5, _ := gin.CreateTestContext(w5)
	c5.Set(middleware.ContextKeyUserID, uint64(12))
	uid13, err13 := middleware.GetUserIDFromContext(c5)
	assert.NoError(t, err13)
	assert.Equal(t, uint64(12), uid13)
	mockUC.On("GetUserProfiles", mock.Anything, mock.Anything).Return([]*dto.ProfileResponse{{ID: "1"}}, nil)
	c5.Request = httptest.NewRequest("GET", "/profiles", nil)
	h.GetUserProfiles(c5)
	assert.Equal(t, http.StatusOK, w5.Code)

	mockUC.AssertExpectations(t)
}
