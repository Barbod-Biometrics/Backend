package admin

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Barbod-Biometrics/Backend/gateway/internal/application/dto/admin"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/infrastructure/test/mocks"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestListProfiles_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockUC := mocks.NewMockAdminProfileUsecase(t)
	h := NewAdminProfileHandler(mockUC)

	// prepare expected request and response
	expectedReq := admin.ListProfilesRequest{Page: 1, PageSize: 10, Status: "pending", ProfileType: "personal", Search: "john"}
	resp := &admin.PaginatedProfilesResponse{Profiles: []admin.ProfileListItem{{ID: 42, ProfileName: "john"}}, TotalCount: 1, Page: 1, PageSize: 10, TotalPages: 1}

	mockUC.On("ListProfiles", mock.Anything, mock.MatchedBy(func(r admin.ListProfilesRequest) bool {
		return r.Page == expectedReq.Page && r.PageSize == expectedReq.PageSize && r.Status == expectedReq.Status && r.ProfileType == expectedReq.ProfileType && r.Search == expectedReq.Search
	})).Return(resp, nil)

	r := gin.New()
	r.GET("/admin/profiles", h.ListProfiles)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/admin/profiles?page=1&page_size=10&status=pending&profile_type=personal&search=john", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var got admin.PaginatedProfilesResponse
	err := json.Unmarshal(w.Body.Bytes(), &got)
	assert.NoError(t, err)
	assert.Equal(t, resp.TotalCount, got.TotalCount)
	assert.Equal(t, resp.Profiles[0].ID, got.Profiles[0].ID)

	mockUC.AssertExpectations(t)
}

func TestListProfiles_BindError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockUC := mocks.NewMockAdminProfileUsecase(t)
	h := NewAdminProfileHandler(mockUC)

	r := gin.New()
	r.GET("/admin/profiles", h.ListProfiles)

	// invalid page value should cause binding error (non-numeric)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/admin/profiles?page=abc", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	mockUC.AssertExpectations(t)
}

func TestListProfiles_UsecaseError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockUC := mocks.NewMockAdminProfileUsecase(t)
	h := NewAdminProfileHandler(mockUC)

	r := gin.New()
	r.GET("/admin/profiles", h.ListProfiles)

	// usecase error -> 500
	mockUC.On("ListProfiles", mock.Anything, mock.Anything).Return((*admin.PaginatedProfilesResponse)(nil), assert.AnError)
	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest("GET", "/admin/profiles?page=1", nil)
	r.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusInternalServerError, w2.Code)

	mockUC.AssertExpectations(t)
}

func TestGetProfileDetail_SuccessAndErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockUC := mocks.NewMockAdminProfileUsecase(t)
	h := NewAdminProfileHandler(mockUC)

	r := gin.New()
	r.GET("/admin/profiles/:id", h.GetProfileDetail)

	// success
	resp := &admin.ProfileDetailResponse{ID: 10, ProfileName: "p1"}
	mockUC.On("GetProfileDetail", mock.Anything, uint64(10)).Return(resp, nil)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/admin/profiles/10", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var got admin.ProfileDetailResponse
	err := json.Unmarshal(w.Body.Bytes(), &got)
	assert.NoError(t, err)
	assert.Equal(t, resp.ID, got.ID)

	// invalid id
	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest("GET", "/admin/profiles/notanumber", nil)
	r.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusBadRequest, w2.Code)

	// usecase not found -> 404
	mockUC.On("GetProfileDetail", mock.Anything, uint64(99)).Return((*admin.ProfileDetailResponse)(nil), assert.AnError)
	w3 := httptest.NewRecorder()
	req3 := httptest.NewRequest("GET", "/admin/profiles/99", nil)
	r.ServeHTTP(w3, req3)
	assert.Equal(t, http.StatusNotFound, w3.Code)

	mockUC.AssertExpectations(t)
}

func TestApproveProfile_SuccessAndErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockUC := mocks.NewMockAdminProfileUsecase(t)
	h := NewAdminProfileHandler(mockUC)

	r := gin.New()
	r.POST("/admin/profiles/:id/approve", h.ApproveProfile)

	// success
	mockUC.On("ApproveProfile", mock.Anything, uint64(5), admin.ApproveProfileRequest{Note: "ok"}).Return(nil)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/admin/profiles/5/approve", strings.NewReader(`{"note":"ok"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var res admin.ActionResponse
	err := json.Unmarshal(w.Body.Bytes(), &res)
	assert.NoError(t, err)
	assert.True(t, res.Success)

	// invalid id
	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest("POST", "/admin/profiles/abc/approve", strings.NewReader(`{"note":"ok"}`))
	req2.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusBadRequest, w2.Code)

	// usecase error -> 400
	mockUC.On("ApproveProfile", mock.Anything, uint64(6), admin.ApproveProfileRequest{}).Return(assert.AnError)
	w3 := httptest.NewRecorder()
	req3 := httptest.NewRequest("POST", "/admin/profiles/6/approve", strings.NewReader(`{}`))
	req3.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w3, req3)
	assert.Equal(t, http.StatusBadRequest, w3.Code)

	mockUC.AssertExpectations(t)
}

func TestRejectProfile_SuccessAndErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockUC := mocks.NewMockAdminProfileUsecase(t)
	h := NewAdminProfileHandler(mockUC)

	r := gin.New()
	r.POST("/admin/profiles/:id/reject", h.RejectProfile)

	// invalid id
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/admin/profiles/zz/reject", strings.NewReader(`{"reason":"just wrong"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// invalid body (reason too short)
	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest("POST", "/admin/profiles/7/reject", strings.NewReader(`{"reason":"short"}`))
	req2.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusBadRequest, w2.Code)

	// success
	mockUC.On("RejectProfile", mock.Anything, uint64(7), admin.RejectProfileRequest{Reason: "Valid reason of >10 chars"}).Return(nil)
	w3 := httptest.NewRecorder()
	req3 := httptest.NewRequest("POST", "/admin/profiles/7/reject", strings.NewReader(`{"reason":"Valid reason of >10 chars"}`))
	req3.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w3, req3)
	assert.Equal(t, http.StatusOK, w3.Code)
	var res admin.ActionResponse
	err := json.Unmarshal(w3.Body.Bytes(), &res)
	assert.NoError(t, err)
	assert.True(t, res.Success)

	// usecase error -> 400
	mockUC.On("RejectProfile", mock.Anything, uint64(8), admin.RejectProfileRequest{Reason: "Valid reason of >10 chars"}).Return(assert.AnError)
	w4 := httptest.NewRecorder()
	req4 := httptest.NewRequest("POST", "/admin/profiles/8/reject", strings.NewReader(`{"reason":"Valid reason of >10 chars"}`))
	req4.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w4, req4)
	assert.Equal(t, http.StatusBadRequest, w4.Code)

	mockUC.AssertExpectations(t)
}
