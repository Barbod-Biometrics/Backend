package contact

import (
	"net/http"
	"strconv"

	"github.com/Barbod-Biometrics/Backend/gateway/internal/application/dto/contact"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/application/usecase"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/logger"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/presentation/middleware"
	"github.com/gin-gonic/gin"
)

type ContactSalesHandler struct {
	usecase usecase.ContactSalesUsecase
	logger  logger.Logger
}

func NewContactSalesHandler(u usecase.ContactSalesUsecase, l logger.Logger) *ContactSalesHandler {
	return &ContactSalesHandler{
		usecase: u,
		logger:  l,
	}
}

// SubmitContactSales handles landing page form submissions
// @Summary Submit contact sales request
// @Description Submits a contact sales request from the landing page form (protected by Google reCAPTCHA v2).
// @Tags Contact Sales
// @Accept json
// @Produce json
// @Param request body contact.SubmitContactSalesRequest true "Contact sales payload"
// @Success 200 {object} contact.SubmitContactSalesResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /contact-sales [post]
func (h *ContactSalesHandler) SubmitContactSales(c *gin.Context) {
	translator := middleware.GetTranslator(c)
	var req contact.SubmitContactSalesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		msg, _ := translator.Translate("errors.invalid_request_body")
		c.JSON(http.StatusBadRequest, gin.H{"error": msg, "details": err.Error()})
		return
	}

	resp, err := h.usecase.Submit(c.Request.Context(), req, c.ClientIP())
	if err != nil {
		h.logger.Warn("contact sales submission failed", logger.Field{Key: "error", Value: err})
		msg, _ := translator.Translate("errors.contactSalesSubmissionFailed")
		c.JSON(http.StatusBadRequest, gin.H{"error": msg})
		return
	}

	msg, _ := translator.Translate("successMessage.contactSalesSubmitted")
	resp.Message = msg
	c.JSON(http.StatusOK, resp)
}

// AdminListContactSales lists contact sales requests for admins
// @Summary List contact sales requests
// @Security BearerAuth
// @Description Lists contact sales requests with filters and pagination (admin only)
// @Tags Contact Sales
// @Accept json
// @Produce json
// @Param page query int false "Page number (default 1)"
// @Param page_size query int false "Page size (default 10, max 100)"
// @Param status query string false "Status filter" Enums(new, read)
// @Param search query string false "Search in name/phone/email/business"
// @Param has_email query bool false "Filter by presence of email"
// @Param from_date query string false "Filter start date (YYYY-MM-DD)"
// @Param to_date query string false "Filter end date (YYYY-MM-DD)"
// @Param sort_by query string false "Sort field" Enums(created_at, status)
// @Param sort_order query string false "Sort order" Enums(asc, desc)
// @Success 200 {object} contact.ContactSalesListResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Router /admin/contact-sales [get]
func (h *ContactSalesHandler) AdminListContactSales(c *gin.Context) {
	translator := middleware.GetTranslator(c)
	var req contact.AdminListContactSalesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		msg, _ := translator.Translate("errors.invalid_request_body")
		c.JSON(http.StatusBadRequest, gin.H{"error": msg, "details": err.Error()})
		return
	}

	resp, err := h.usecase.AdminList(c.Request.Context(), req)
	if err != nil {
		h.logger.Error("failed to list contact sales", logger.Field{Key: "error", Value: err})
		msg, _ := translator.Translate("errors.contactSalesListFailed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// AdminMarkRead marks a request as read (archives it)
// @Summary Mark contact sales request as read
// @Security BearerAuth
// @Tags Contact Sales
// @Accept json
// @Produce json
// @Param id path int true "Contact sales request ID"
// @Success 200 {object} contact.ContactSalesActionResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Router /admin/contact-sales/{id}/mark-read [post]
func (h *ContactSalesHandler) AdminMarkRead(c *gin.Context) {
	translator := middleware.GetTranslator(c)
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		msg, _ := translator.Translate("errors.invalidId")
		c.JSON(http.StatusBadRequest, gin.H{"error": msg})
		return
	}

	resp, err := h.usecase.MarkRead(c.Request.Context(), id)
	if err != nil {
		h.logger.Error("failed to mark contact sales as read", logger.Field{Key: "error", Value: err}, logger.Field{Key: "id", Value: id})
		msg, _ := translator.Translate("errors.contactSalesMarkReadFailed")
		c.JSON(http.StatusBadRequest, gin.H{"error": msg})
		return
	}

	msg, _ := translator.Translate("successMessage.contactSalesMarkedRead")
	resp.Message = msg
	c.JSON(http.StatusOK, resp)
}

// AdminDelete removes a request (soft delete)
// @Summary Delete contact sales request
// @Security BearerAuth
// @Tags Contact Sales
// @Accept json
// @Produce json
// @Param id path int true "Contact sales request ID"
// @Success 200 {object} contact.ContactSalesActionResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Router /admin/contact-sales/{id} [delete]
func (h *ContactSalesHandler) AdminDelete(c *gin.Context) {
	translator := middleware.GetTranslator(c)
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		msg, _ := translator.Translate("errors.invalidId")
		c.JSON(http.StatusBadRequest, gin.H{"error": msg})
		return
	}

	resp, err := h.usecase.Delete(c.Request.Context(), id)
	if err != nil {
		h.logger.Error("failed to delete contact sales", logger.Field{Key: "error", Value: err}, logger.Field{Key: "id", Value: id})
		msg, _ := translator.Translate("errors.contactSalesDeleteFailed")
		c.JSON(http.StatusBadRequest, gin.H{"error": msg})
		return
	}

	msg, _ := translator.Translate("successMessage.contactSalesDeleted")
	resp.Message = msg
	c.JSON(http.StatusOK, resp)
}
