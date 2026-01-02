package repository

import (
	"context"

	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/entity"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/enum"
)

type TicketRepository interface {
	// Ticket operations
	Create(ctx context.Context, ticket *entity.Ticket) error
	GetByID(ctx context.Context, ticketID uint64) (*entity.Ticket, error)
	GetByIDWithUser(ctx context.Context, ticketID uint64) (*entity.Ticket, error)
	GetByUserID(ctx context.Context, userID uint64) ([]*entity.Ticket, error)
	GetAll(ctx context.Context) ([]*entity.Ticket, error)
	GetAllWithUsers(ctx context.Context) ([]*entity.Ticket, error)
	UpdateStatus(ctx context.Context, ticketID uint64, status enum.TicketStatus) error
	Update(ctx context.Context, ticket *entity.Ticket) error

	// Message operations
	CreateMessage(ctx context.Context, message *entity.TicketMessage) error
	GetMessagesByTicketID(ctx context.Context, ticketID uint64) ([]*entity.TicketMessage, error)
	GetLastMessageByTicketID(ctx context.Context, ticketID uint64) (*entity.TicketMessage, error)
}
