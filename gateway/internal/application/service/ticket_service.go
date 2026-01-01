package service

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/Barbod-Biometrics/Backend/gateway/internal/application/dto/ticket"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/application/usecase"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/entity"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/enum"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/repository"
	"github.com/Barbod-Biometrics/Backend/gateway/pkg/storage"
	"github.com/jalaali/go-jalaali"
)

var (
	ErrTicketNotFound = errors.New("ticket not found")
	ErrUnauthorized   = errors.New("unauthorized access to ticket")
	ErrTicketClosed   = errors.New("ticket is already closed")
	ErrNoAttachment   = errors.New("ticket has no attachment")
)

type TicketService struct {
	ticketRepo repository.TicketRepository
	userRepo   repository.UserRepository
	storage    *storage.MinioClient
}

func NewTicketService(
	ticketRepo repository.TicketRepository,
	userRepo repository.UserRepository,
	storage *storage.MinioClient,
) *TicketService {
	return &TicketService{
		ticketRepo: ticketRepo,
		userRepo:   userRepo,
		storage:    storage,
	}
}

var _ usecase.TicketUsecase = (*TicketService)(nil)

// CreateTicket creates a new support ticket
func (s *TicketService) CreateTicket(ctx context.Context, userID uint64, req ticket.CreateTicketRequest) (*ticket.TicketResponse, error) {
	var attachmentKey *string
	if req.Attachment != "" {
		attachmentKey = &req.Attachment
	}

	newTicket := &entity.Ticket{
		UserID:        userID,
		Service:       enum.TicketService(req.Service),
		Title:         req.Title,
		Description:   req.Description,
		Status:        enum.TicketStatusPending,
		AttachmentKey: attachmentKey,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := s.ticketRepo.Create(ctx, newTicket); err != nil {
		return nil, fmt.Errorf("failed to create ticket: %w", err)
	}

	// Get attachment URL if exists
	attachmentURL := ""
	if attachmentKey != nil && s.storage != nil {
		url, err := s.storage.GetFileURL(ctx, *attachmentKey, time.Hour*24)
		if err == nil {
			attachmentURL = url
		}
	}

	return &ticket.TicketResponse{
		ID:          fmt.Sprintf("%d", newTicket.TicketID),
		Service:     enum.TicketService(req.Service).DisplayName(),
		Title:       newTicket.Title,
		Description: newTicket.Description,
		Status:      string(newTicket.Status),
		Attachment:  attachmentURL,
		CreatedAt:   toJalaliDate(newTicket.CreatedAt),
	}, nil
}

// GetUserTickets retrieves all tickets for a user
func (s *TicketService) GetUserTickets(ctx context.Context, userID uint64) ([]*ticket.TicketListItemResponse, error) {
	tickets, err := s.ticketRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tickets: %w", err)
	}

	response := make([]*ticket.TicketListItemResponse, 0, len(tickets))
	for _, t := range tickets {
		response = append(response, &ticket.TicketListItemResponse{
			ID:        fmt.Sprintf("%d", t.TicketID),
			Service:   t.Service.DisplayName(),
			Title:     t.Title,
			Status:    string(t.Status),
			CreatedAt: toJalaliDate(t.CreatedAt),
		})
	}

	return response, nil
}

// GetAllTickets retrieves all tickets for admin
func (s *TicketService) GetAllTickets(ctx context.Context) ([]*ticket.AdminTicketListItemResponse, error) {
	tickets, err := s.ticketRepo.GetAllWithUsers(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get tickets: %w", err)
	}

	response := make([]*ticket.AdminTicketListItemResponse, 0, len(tickets))
	for _, t := range tickets {
		item := &ticket.AdminTicketListItemResponse{
			ID:        fmt.Sprintf("tk_%d", t.TicketID),
			Service:   string(t.Service),
			Title:     t.Title,
			Status:    string(t.Status),
			CreatedAt: t.CreatedAt,
		}

		// Get user contact info
		if t.User != nil {
			item.UserContact = &ticket.UserContactResponse{
				Phone: t.User.PhoneNumber,
			}
			if t.User.Email != nil {
				item.UserContact.Email = *t.User.Email
			}
		}

		// Get last message time
		lastMessage, _ := s.ticketRepo.GetLastMessageByTicketID(ctx, t.TicketID)
		if lastMessage != nil {
			item.LastMessageAt = &lastMessage.CreatedAt
		}

		response = append(response, item)
	}

	return response, nil
}

// GetTicketDetail retrieves detailed ticket info for admin
func (s *TicketService) GetTicketDetail(ctx context.Context, ticketID uint64) (*ticket.AdminTicketDetailResponse, error) {
	t, err := s.ticketRepo.GetByIDWithUser(ctx, ticketID)
	if err != nil {
		return nil, fmt.Errorf("failed to get ticket: %w", err)
	}
	if t == nil {
		return nil, ErrTicketNotFound
	}

	response := &ticket.AdminTicketDetailResponse{
		ID:        fmt.Sprintf("tk_%d", t.TicketID),
		UserID:    fmt.Sprintf("u_%d", t.UserID),
		Service:   string(t.Service),
		Title:     t.Title,
		Message:   t.Description,
		Status:    string(t.Status),
		CreatedAt: t.CreatedAt,
		UpdatedAt: t.UpdatedAt,
	}

	// Get user contact info
	if t.User != nil {
		response.UserContact = &ticket.UserContactResponse{
			Phone: t.User.PhoneNumber,
		}
		if t.User.Email != nil {
			response.UserContact.Email = *t.User.Email
		}
	}

	// Get file URL if exists
	if t.AttachmentKey != nil && *t.AttachmentKey != "" && s.storage != nil {
		url, err := s.storage.GetFileURL(ctx, *t.AttachmentKey, time.Hour*24)
		if err == nil {
			response.FileURL = url
		}
	}

	return response, nil
}

// GetTicketFileURL retrieves the file URL for a ticket
func (s *TicketService) GetTicketFileURL(ctx context.Context, ticketID uint64) (*ticket.FileURLResponse, error) {
	t, err := s.ticketRepo.GetByID(ctx, ticketID)
	if err != nil {
		return nil, fmt.Errorf("failed to get ticket: %w", err)
	}
	if t == nil {
		return nil, ErrTicketNotFound
	}

	if t.AttachmentKey == nil || *t.AttachmentKey == "" {
		return nil, ErrNoAttachment
	}

	url, err := s.storage.GetFileURL(ctx, *t.AttachmentKey, time.Hour*24)
	if err != nil {
		return nil, fmt.Errorf("failed to get file URL: %w", err)
	}

	return &ticket.FileURLResponse{
		URL: url,
	}, nil
}

// GetTicketMessages retrieves all messages for a ticket
func (s *TicketService) GetTicketMessages(ctx context.Context, ticketID uint64, userID uint64, isAdmin bool) ([]*ticket.MessageResponse, error) {
	// Verify access
	t, err := s.ticketRepo.GetByID(ctx, ticketID)
	if err != nil {
		return nil, fmt.Errorf("failed to get ticket: %w", err)
	}
	if t == nil {
		return nil, ErrTicketNotFound
	}

	// If not admin, check if user owns the ticket
	if !isAdmin && t.UserID != userID {
		return nil, ErrUnauthorized
	}

	messages, err := s.ticketRepo.GetMessagesByTicketID(ctx, ticketID)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages: %w", err)
	}

	response := make([]*ticket.MessageResponse, 0, len(messages))
	for _, m := range messages {
		response = append(response, &ticket.MessageResponse{
			ID:        fmt.Sprintf("msg_%d", m.MessageID),
			TicketID:  fmt.Sprintf("tk_%d", m.TicketID),
			Sender:    string(m.Sender),
			Message:   m.Message,
			CreatedAt: m.CreatedAt,
		})
	}

	return response, nil
}

// SendMessage sends a new message to a ticket
func (s *TicketService) SendMessage(ctx context.Context, ticketID uint64, userID uint64, isAdmin bool, req ticket.SendMessageRequest) (*ticket.MessageResponse, error) {
	// Get ticket and verify access
	t, err := s.ticketRepo.GetByID(ctx, ticketID)
	if err != nil {
		return nil, fmt.Errorf("failed to get ticket: %w", err)
	}
	if t == nil {
		return nil, ErrTicketNotFound
	}

	// If not admin, check if user owns the ticket
	if !isAdmin && t.UserID != userID {
		return nil, ErrUnauthorized
	}

	// Check if ticket is closed
	if t.Status == enum.TicketStatusClosed {
		return nil, ErrTicketClosed
	}

	// Determine sender
	sender := enum.MessageSenderUser
	if isAdmin {
		sender = enum.MessageSenderAdmin
	}

	// Create message
	message := &entity.TicketMessage{
		TicketID:  ticketID,
		Sender:    sender,
		Message:   req.Message,
		CreatedAt: time.Now(),
	}

	if err := s.ticketRepo.CreateMessage(ctx, message); err != nil {
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	// Update ticket status based on sender
	newStatus := enum.TicketStatusPending
	if isAdmin {
		newStatus = enum.TicketStatusAnswered
	}

	if err := s.ticketRepo.UpdateStatus(ctx, ticketID, newStatus); err != nil {
		return nil, fmt.Errorf("failed to update ticket status: %w", err)
	}

	return &ticket.MessageResponse{
		ID:        fmt.Sprintf("msg_%d", message.MessageID),
		TicketID:  fmt.Sprintf("tk_%d", message.TicketID),
		Sender:    string(message.Sender),
		Message:   message.Message,
		CreatedAt: message.CreatedAt,
	}, nil
}

// CloseTicket closes a ticket
func (s *TicketService) CloseTicket(ctx context.Context, ticketID uint64, userID uint64, isAdmin bool) (*ticket.CloseTicketResponse, error) {
	// Get ticket and verify access
	t, err := s.ticketRepo.GetByID(ctx, ticketID)
	if err != nil {
		return nil, fmt.Errorf("failed to get ticket: %w", err)
	}
	if t == nil {
		return nil, ErrTicketNotFound
	}

	// If not admin, check if user owns the ticket
	if !isAdmin && t.UserID != userID {
		return nil, ErrUnauthorized
	}

	// Check if already closed
	if t.Status == enum.TicketStatusClosed {
		return nil, ErrTicketClosed
	}

	if err := s.ticketRepo.UpdateStatus(ctx, ticketID, enum.TicketStatusClosed); err != nil {
		return nil, fmt.Errorf("failed to close ticket: %w", err)
	}

	return &ticket.CloseTicketResponse{
		TicketID: fmt.Sprintf("tk_%d", ticketID),
		Status:   string(enum.TicketStatusClosed),
	}, nil
}

// GetUploadURL generates a presigned URL for uploading a ticket attachment
func (s *TicketService) GetUploadURL(ctx context.Context, userID uint64, fileExtension string) (string, string, error) {
	// Generate unique object key
	objectKey := fmt.Sprintf("tickets/%d/%d%s", userID, time.Now().UnixNano(), fileExtension)

	url, err := s.storage.GeneratePresignedUploadURL(ctx, objectKey, time.Minute*15)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate upload URL: %w", err)
	}

	return url, objectKey, nil
}

// toJalaliDate converts a time.Time to Jalali date string (1403/10/05)
func toJalaliDate(t time.Time) string {
	jy, jm, jd, err := jalaali.ToJalaali(t.Year(), t.Month(), t.Day())
	if err != nil {
		// Fallback to Gregorian if conversion fails
		return t.Format("2006/01/02")
	}
	return fmt.Sprintf("%04d/%02d/%02d", jy, jm, jd)
}

// parseTicketID parses a ticket ID string (with or without "tk_" prefix) to uint64
func parseTicketID(idStr string) (uint64, error) {
	// Remove "tk_" prefix if present
	if len(idStr) > 3 && idStr[:3] == "tk_" {
		idStr = idStr[3:]
	}
	return strconv.ParseUint(idStr, 10, 64)
}
