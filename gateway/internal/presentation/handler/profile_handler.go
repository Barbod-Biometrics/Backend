package handler

import (
	"net/http"
	"strconv"

	"github.com/Barbod-Biometrics/Backend/gateway/internal/application/dto/profile"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/application/usecase"
	"github.com/gin-gonic/gin"
)

type ProfileHandler struct {
	profileUsecase usecase.ProfileUsecase
}

func NewProfileHandler(u usecase.ProfileUsecase) *ProfileHandler {
	return &ProfileHandler{
		profileUsecase: u,
	}
}

// --- Helper to get UserID ---
// TODO: Replace this with actual JWT middleware logic later
func (h *ProfileHandler) getUserID(c *gin.Context) uint64 {
	// For testing now, return a static ID (e.g., 1)
	// Later, this will be: c.GetUint64("userID")
	return 1
}

// CreateDraft creates a new profile draft
// @Summary Create Profile Draft
// @Description Create a new profile draft
// @Tags Profiles
// @Accept json
// @Produce json
// @Param request body profile.CreateProfileRequest true "Create Profile Request"
// @Success 201 {object} profile.ProfileResponse
// @Failure 400 {object} map[string]string
// @Router /profiles/ [post]
func (h *ProfileHandler) CreateDraft(c *gin.Context) {
	var req profile.CreateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := h.getUserID(c)
	resp, err := h.profileUsecase.CreateDraft(c.Request.Context(), userID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, resp)
}

// UpdateDraft updates fields of an existing profile draft
// @Summary Update Profile Draft
// @Description Update a profile draft (partial updates allowed)
// @Tags Profiles
// @Accept json
// @Produce json
// @Param id path int true "Profile ID"
// @Param request body profile.UpdateProfileRequest true "Update Profile Request"
// @Success 200 {object} profile.ProfileResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /profiles/{id} [patch]
func (h *ProfileHandler) UpdateDraft(c *gin.Context) {
	idParam := c.Param("id")
	profileID, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid profile id"})
		return
	}

	var req profile.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := h.getUserID(c)
	resp, err := h.profileUsecase.UpdateDraft(c.Request.Context(), userID, profileID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// SaveDocument handles POST /api/v1/profiles/:id/documents
// @Summary Save Document URL
// @Description Save a document URL for a profile (e.g., national card, business docs)
// @Tags Profiles
// @Accept json
// @Produce json
// @Param id path int true "Profile ID"
// @Param request body profile.SaveDocumentRequest true "Save Document Request"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /profiles/{id}/documents [post]
func (h *ProfileHandler) SaveDocument(c *gin.Context) {
	idParam := c.Param("id")
	profileID, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid profile id"})
		return
	}

	var req profile.SaveDocumentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := h.getUserID(c)
	err = h.profileUsecase.SaveDocument(c.Request.Context(), userID, profileID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "document saved successfully"})
}

// Submit handles POST /api/v1/profiles/:id/submit
// @Summary Submit Profile
// @Description Submit a profile for verification (final submission)
// @Tags Profiles
// @Accept json
// @Produce json
// @Param id path int true "Profile ID"
// @Success 200 {object} profile.ProfileResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /profiles/{id}/submit [post]
func (h *ProfileHandler) Submit(c *gin.Context) {
	idParam := c.Param("id")
	profileID, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid profile id"})
		return
	}

	userID := h.getUserID(c)
	resp, err := h.profileUsecase.SubmitProfile(c.Request.Context(), userID, profileID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetProfile handles GET /api/v1/profiles/:id
// @Summary Get Profile
// @Description Get a profile by id
// @Tags Profiles
// @Accept json
// @Produce json
// @Param id path int true "Profile ID"
// @Success 200 {object} profile.ProfileResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /profiles/{id} [get]
func (h *ProfileHandler) GetProfile(c *gin.Context) {
	idParam := c.Param("id")
	profileID, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid profile id"})
		return
	}

	userID := h.getUserID(c)
	resp, err := h.profileUsecase.GetByID(c.Request.Context(), userID, profileID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}
