package admin

import (
	"net/http"
	"strconv"

	"github.com/Barbod-Biometrics/Backend/gateway/internal/application/dto/admin"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/application/usecase"
	"github.com/gin-gonic/gin"
)

type AdminProfileHandler struct {
	adminUsecase usecase.AdminProfileUsecase
}

func NewAdminProfileHandler(u usecase.AdminProfileUsecase) *AdminProfileHandler {
	return &AdminProfileHandler{
		adminUsecase: u,
	}
}

// ListProfiles lists profiles with search/sort/filter/pagination
// @Summary List Profiles (Admin)
// @Security BearerAuth
// @Description List all profiles with optional filters, sorting, and pagination. Admin only.
// @Tags Admin - Profiles
// @Accept json
// @Produce json
// @Param page query int false "Page number (default: 1)"
// @Param page_size query int false "Items per page (default: 10, max: 100)"
// @Param status query string false "Filter by status" Enums(draft, pending, verified, rejected)
// @Param profile_type query string false "Filter by profile type" Enums(personal, business)
// @Param search query string false "Search in name, national ID, phone number"
// @Param sort_by query string false "Sort field" Enums(created_at, profile_name, verification_status, profile_type)
// @Param sort_order query string false "Sort order" Enums(asc, desc)
// @Success 200 {object} admin.PaginatedProfilesResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Router /admin/profiles [get]
func (h *AdminProfileHandler) ListProfiles(c *gin.Context) {
	var req admin.ListProfilesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.adminUsecase.ListProfiles(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetProfileDetail returns full details of a profile
// @Summary Get Profile Detail (Admin)
// @Security BearerAuth
// @Description Get full details of a profile for admin review. Admin only.
// @Tags Admin - Profiles
// @Accept json
// @Produce json
// @Param id path int true "Profile ID"
// @Success 200 {object} admin.ProfileDetailResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /admin/profiles/{id} [get]
func (h *AdminProfileHandler) GetProfileDetail(c *gin.Context) {
	idParam := c.Param("id")
	profileID, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid profile id"})
		return
	}

	resp, err := h.adminUsecase.GetProfileDetail(c.Request.Context(), profileID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// ApproveProfile approves a pending profile
// @Summary Approve Profile (Admin)
// @Security BearerAuth
// @Description Approve a pending profile verification request. Admin only.
// @Tags Admin - Profiles
// @Accept json
// @Produce json
// @Param id path int true "Profile ID"
// @Param request body admin.ApproveProfileRequest false "Approval details"
// @Success 200 {object} admin.ActionResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /admin/profiles/{id}/approve [post]
func (h *AdminProfileHandler) ApproveProfile(c *gin.Context) {
	idParam := c.Param("id")
	profileID, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid profile id"})
		return
	}

	var req admin.ApproveProfileRequest
	// Note and body are optional for approval
	_ = c.ShouldBindJSON(&req)

	if err := h.adminUsecase.ApproveProfile(c.Request.Context(), profileID, req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, admin.ActionResponse{
		Success: true,
		Message: "Profile approved successfully",
	})
}

// RejectProfile rejects a pending profile
// @Summary Reject Profile (Admin)
// @Security BearerAuth
// @Description Reject a pending profile verification request with a reason. Admin only.
// @Tags Admin - Profiles
// @Accept json
// @Produce json
// @Param id path int true "Profile ID"
// @Param request body admin.RejectProfileRequest true "Rejection reason"
// @Success 200 {object} admin.ActionResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /admin/profiles/{id}/reject [post]
func (h *AdminProfileHandler) RejectProfile(c *gin.Context) {
	idParam := c.Param("id")
	profileID, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid profile id"})
		return
	}

	var req admin.RejectProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.adminUsecase.RejectProfile(c.Request.Context(), profileID, req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, admin.ActionResponse{
		Success: true,
		Message: "Profile rejected successfully",
	})
}
