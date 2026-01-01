package usecase

import (
	"context"

	"github.com/Barbod-Biometrics/Backend/gateway/internal/application/dto/ticket"
)

type TicketUsecase interface {
	// User operations
	CreateTicket(ctx context.Context, userID uint64, req ticket.CreateTicketRequest) (*ticket.TicketResponse, error)
	GetUserTickets(ctx context.Context, userID uint64) ([]*ticket.TicketListItemResponse, error)

	// Admin operations
	GetAllTickets(ctx context.Context) ([]*ticket.AdminTicketListItemResponse, error)
	GetTicketDetail(ctx context.Context, ticketID uint64) (*ticket.AdminTicketDetailResponse, error)
	GetTicketFileURL(ctx context.Context, ticketID uint64) (*ticket.FileURLResponse, error)

	// Shared operations (both user and admin)
	GetTicketMessages(ctx context.Context, ticketID uint64, userID uint64, isAdmin bool) ([]*ticket.MessageResponse, error)
	SendMessage(ctx context.Context, ticketID uint64, userID uint64, isAdmin bool, req ticket.SendMessageRequest) (*ticket.MessageResponse, error)
	CloseTicket(ctx context.Context, ticketID uint64, userID uint64, isAdmin bool) (*ticket.CloseTicketResponse, error)

	// Upload operations
	GetUploadURL(ctx context.Context, userID uint64, fileExtension string) (string, string, error) // returns uploadURL, objectKey, error
}
