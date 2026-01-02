package support

import (
	"net/http"
	"strconv"

	"github.com/Barbod-Biometrics/Backend/gateway/internal/application/dto/ticket"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/application/usecase"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/logger"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/presentation/middleware"
	"github.com/gin-gonic/gin"
)

type TicketHandler struct {
	ticketUsecase usecase.TicketUsecase
	log           logger.Logger
}

func NewTicketHandler(ticketUsecase usecase.TicketUsecase, log logger.Logger) *TicketHandler {
	return &TicketHandler{
		ticketUsecase: ticketUsecase,
		log:           log,
	}
}

// getUserID extracts the authenticated user's ID from the request context
func (h *TicketHandler) getUserID(c *gin.Context) (uint64, error) {
	return middleware.GetUserIDFromContext(c)
}

// isAdmin checks if the current user is an admin
func (h *TicketHandler) isAdmin(c *gin.Context) bool {
	isAdmin, exists := c.Get("isAdmin")
	if !exists {
		return false
	}
	admin, ok := isAdmin.(bool)
	return ok && admin
}

// parseTicketIDParam parses the ticket ID from URL parameter
func (h *TicketHandler) parseTicketIDParam(c *gin.Context) (uint64, error) {
	idStr := c.Param("ticketId")
	// Remove "tk_" prefix if present
	if len(idStr) > 3 && idStr[:3] == "tk_" {
		idStr = idStr[3:]
	}
	return strconv.ParseUint(idStr, 10, 64)
}

// ==================== USER ENDPOINTS ====================

// GetUserTickets retrieves all tickets for the authenticated user
// @Summary Get user tickets
// @Security BearerAuth
// @Description Get all support tickets for the authenticated user
// @Tags Support
// @Accept json
// @Produce json
// @Success 200 {object} ticket.SuccessListResponse{data=[]ticket.TicketListItemResponse}
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /support/tickets [get]
func (h *TicketHandler) GetUserTickets(c *gin.Context) {
	userID, err := h.getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	tickets, err := h.ticketUsecase.GetUserTickets(c.Request.Context(), userID)
	if err != nil {
		h.log.Error("failed to get user tickets",
			logger.Field{Key: "error", Value: err},
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, ticket.SuccessListResponse{
		Success: true,
		Data:    tickets,
	})
}

// CreateTicket creates a new support ticket
// @Summary Create a new ticket
// @Security BearerAuth
// @Description Create a new support ticket
// @Tags Support
// @Accept json
// @Produce json
// @Param request body ticket.CreateTicketRequest true "Create Ticket Request"
// @Success 201 {object} ticket.SuccessResponse{data=ticket.TicketResponse}
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /support/tickets [post]
func (h *TicketHandler) CreateTicket(c *gin.Context) {
	userID, err := h.getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req ticket.CreateTicketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.ticketUsecase.CreateTicket(c.Request.Context(), userID, req)
	if err != nil {
		h.log.Error("failed to create ticket",
			logger.Field{Key: "error", Value: err},
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, ticket.SuccessResponse{
		Success: true,
		Data:    resp,
	})
}

// GetUploadURL generates a presigned URL for uploading a ticket attachment
// @Summary Get upload URL
// @Security BearerAuth
// @Description Generate a presigned URL for uploading a ticket attachment
// @Tags Support
// @Accept json
// @Produce json
// @Param file_extension query string true "File extension (e.g., .pdf, .png, .jpg)"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /support/tickets/upload-url [get]
func (h *TicketHandler) GetUploadURL(c *gin.Context) {
	userID, err := h.getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	fileExtension := c.Query("file_extension")
	if fileExtension == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file_extension is required"})
		return
	}

	uploadURL, objectKey, err := h.ticketUsecase.GetUploadURL(c.Request.Context(), userID, fileExtension)
	if err != nil {
		h.log.Error("failed to get upload URL",
			logger.Field{Key: "error", Value: err},
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"upload_url": uploadURL,
		"object_key": objectKey,
	})
}

// ==================== ADMIN ENDPOINTS ====================

// GetAllTickets retrieves all tickets for admin
// @Summary Get all tickets (Admin)
// @Security BearerAuth
// @Description Get all support tickets (admin only)
// @Tags Admin - Support
// @Accept json
// @Produce json
// @Success 200 {array} ticket.AdminTicketListItemResponse
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /admin/tickets [get]
func (h *TicketHandler) GetAllTickets(c *gin.Context) {
	tickets, err := h.ticketUsecase.GetAllTickets(c.Request.Context())
	if err != nil {
		h.log.Error("failed to get all tickets",
			logger.Field{Key: "error", Value: err},
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tickets)
}

// GetTicketDetail retrieves detailed ticket info for admin
// @Summary Get ticket detail (Admin)
// @Security BearerAuth
// @Description Get detailed ticket information (admin only)
// @Tags Admin - Support
// @Accept json
// @Produce json
// @Param ticketId path string true "Ticket ID"
// @Success 200 {object} ticket.AdminTicketDetailResponse
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /admin/tickets/{ticketId} [get]
func (h *TicketHandler) GetTicketDetail(c *gin.Context) {
	ticketID, err := h.parseTicketIDParam(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ticket ID"})
		return
	}

	resp, err := h.ticketUsecase.GetTicketDetail(c.Request.Context(), ticketID)
	if err != nil {
		h.log.Error("failed to get ticket detail",
			logger.Field{Key: "error", Value: err},
		)
		if err.Error() == "ticket not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "ticket not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetTicketFile retrieves the file URL for a ticket
// @Summary Get ticket file URL (Admin)
// @Security BearerAuth
// @Description Get the attachment file URL for a ticket (admin only)
// @Tags Admin - Support
// @Accept json
// @Produce json
// @Param ticketId path string true "Ticket ID"
// @Success 200 {object} ticket.FileURLResponse
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /admin/tickets/{ticketId}/file [get]
func (h *TicketHandler) GetTicketFile(c *gin.Context) {
	ticketID, err := h.parseTicketIDParam(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ticket ID"})
		return
	}

	resp, err := h.ticketUsecase.GetTicketFileURL(c.Request.Context(), ticketID)
	if err != nil {
		h.log.Error("failed to get ticket file",
			logger.Field{Key: "error", Value: err},
		)
		if err.Error() == "ticket not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "ticket not found"})
			return
		}
		if err.Error() == "ticket has no attachment" {
			c.JSON(http.StatusNotFound, gin.H{"error": "no attachment found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// ==================== SHARED ENDPOINTS (User & Admin) ====================

// GetTicketMessages retrieves all messages for a ticket
// @Summary Get ticket messages
// @Security BearerAuth
// @Description Get all messages for a ticket (accessible by both user and admin)
// @Tags Support - Messages
// @Accept json
// @Produce json
// @Param ticketId path string true "Ticket ID"
// @Success 200 {array} ticket.MessageResponse
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /tickets/{ticketId}/messages [get]
func (h *TicketHandler) GetTicketMessages(c *gin.Context) {
	userID, err := h.getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	ticketID, err := h.parseTicketIDParam(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ticket ID"})
		return
	}

	isAdmin := h.isAdmin(c)

	messages, err := h.ticketUsecase.GetTicketMessages(c.Request.Context(), ticketID, userID, isAdmin)
	if err != nil {
		h.log.Error("failed to get ticket messages",
			logger.Field{Key: "error", Value: err},
		)
		if err.Error() == "ticket not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "ticket not found"})
			return
		}
		if err.Error() == "unauthorized access to ticket" {
			c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, messages)
}

// SendMessage sends a new message to a ticket
// @Summary Send message
// @Security BearerAuth
// @Description Send a new message to a ticket (accessible by both user and admin)
// @Tags Support - Messages
// @Accept json
// @Produce json
// @Param ticketId path string true "Ticket ID"
// @Param request body ticket.SendMessageRequest true "Send Message Request"
// @Success 200 {object} ticket.MessageResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /tickets/{ticketId}/messages [post]
func (h *TicketHandler) SendMessage(c *gin.Context) {
	userID, err := h.getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	ticketID, err := h.parseTicketIDParam(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ticket ID"})
		return
	}

	var req ticket.SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	isAdmin := h.isAdmin(c)

	resp, err := h.ticketUsecase.SendMessage(c.Request.Context(), ticketID, userID, isAdmin, req)
	if err != nil {
		h.log.Error("failed to send message",
			logger.Field{Key: "error", Value: err},
		)
		if err.Error() == "ticket not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "ticket not found"})
			return
		}
		if err.Error() == "unauthorized access to ticket" {
			c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
			return
		}
		if err.Error() == "ticket is already closed" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "ticket is closed"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// CloseTicket closes a ticket
// @Summary Close ticket
// @Security BearerAuth
// @Description Close a ticket (accessible by both user and admin)
// @Tags Support - Messages
// @Accept json
// @Produce json
// @Param ticketId path string true "Ticket ID"
// @Success 200 {object} ticket.CloseTicketResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /tickets/{ticketId}/close [post]
func (h *TicketHandler) CloseTicket(c *gin.Context) {
	userID, err := h.getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	ticketID, err := h.parseTicketIDParam(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ticket ID"})
		return
	}

	isAdmin := h.isAdmin(c)

	resp, err := h.ticketUsecase.CloseTicket(c.Request.Context(), ticketID, userID, isAdmin)
	if err != nil {
		h.log.Error("failed to close ticket",
			logger.Field{Key: "error", Value: err},
		)
		if err.Error() == "ticket not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "ticket not found"})
			return
		}
		if err.Error() == "unauthorized access to ticket" {
			c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
			return
		}
		if err.Error() == "ticket is already closed" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "ticket is already closed"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}
