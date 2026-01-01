package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/entity"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/enum"
	"gorm.io/gorm"
)

type TicketRepository struct {
	db *gorm.DB
}

func NewTicketRepository(db *gorm.DB) *TicketRepository {
	return &TicketRepository{db: db}
}

func (r *TicketRepository) getDB(ctx context.Context) *gorm.DB {
	if tx, ok := ctx.Value("db_tx").(*gorm.DB); ok {
		return tx
	}
	return r.db
}

// Create creates a new ticket
func (r *TicketRepository) Create(ctx context.Context, ticket *entity.Ticket) error {
	db := r.getDB(ctx)
	return db.WithContext(ctx).Create(ticket).Error
}

// GetByID retrieves a ticket by ID
func (r *TicketRepository) GetByID(ctx context.Context, ticketID uint64) (*entity.Ticket, error) {
	db := r.getDB(ctx)
	var ticket entity.Ticket

	err := db.WithContext(ctx).First(&ticket, ticketID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &ticket, nil
}

// GetByIDWithUser retrieves a ticket by ID with user information
func (r *TicketRepository) GetByIDWithUser(ctx context.Context, ticketID uint64) (*entity.Ticket, error) {
	db := r.getDB(ctx)
	var ticket entity.Ticket

	err := db.WithContext(ctx).First(&ticket, ticketID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	// Manually load user
	var user entity.User
	if err := db.WithContext(ctx).First(&user, ticket.UserID).Error; err == nil {
		ticket.User = &user
	}

	return &ticket, nil
}

// GetByUserID retrieves all tickets for a specific user
func (r *TicketRepository) GetByUserID(ctx context.Context, userID uint64) ([]*entity.Ticket, error) {
	db := r.getDB(ctx)
	var tickets []*entity.Ticket

	err := db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&tickets).Error

	return tickets, err
}

// GetAll retrieves all tickets
func (r *TicketRepository) GetAll(ctx context.Context) ([]*entity.Ticket, error) {
	db := r.getDB(ctx)
	var tickets []*entity.Ticket

	err := db.WithContext(ctx).
		Order("created_at DESC").
		Find(&tickets).Error

	return tickets, err
}

// GetAllWithUsers retrieves all tickets with user information
func (r *TicketRepository) GetAllWithUsers(ctx context.Context) ([]*entity.Ticket, error) {
	db := r.getDB(ctx)
	var tickets []*entity.Ticket

	err := db.WithContext(ctx).
		Order("created_at DESC").
		Find(&tickets).Error

	if err != nil {
		return nil, err
	}

	// Collect unique user IDs
	userIDs := make([]uint64, 0)
	userIDMap := make(map[uint64]bool)
	for _, t := range tickets {
		if !userIDMap[t.UserID] {
			userIDMap[t.UserID] = true
			userIDs = append(userIDs, t.UserID)
		}
	}

	// Fetch all users in one query
	var users []entity.User
	if len(userIDs) > 0 {
		if err := db.WithContext(ctx).Where("user_id IN ?", userIDs).Find(&users).Error; err != nil {
			return tickets, nil // Return tickets without users on error
		}
	}

	// Create user map
	userMap := make(map[uint64]*entity.User)
	for i := range users {
		userMap[users[i].UserID] = &users[i]
	}

	// Assign users to tickets
	for _, t := range tickets {
		if user, ok := userMap[t.UserID]; ok {
			t.User = user
		}
	}

	return tickets, nil
}

// UpdateStatus updates the status of a ticket
func (r *TicketRepository) UpdateStatus(ctx context.Context, ticketID uint64, status enum.TicketStatus) error {
	db := r.getDB(ctx)
	return db.WithContext(ctx).
		Model(&entity.Ticket{}).
		Where("ticket_id = ?", ticketID).
		Updates(map[string]interface{}{
			"status":     status,
			"updated_at": time.Now(),
		}).Error
}

// Update updates a ticket
func (r *TicketRepository) Update(ctx context.Context, ticket *entity.Ticket) error {
	db := r.getDB(ctx)
	ticket.UpdatedAt = time.Now()
	return db.WithContext(ctx).Save(ticket).Error
}

// CreateMessage creates a new message for a ticket
func (r *TicketRepository) CreateMessage(ctx context.Context, message *entity.TicketMessage) error {
	db := r.getDB(ctx)
	return db.WithContext(ctx).Create(message).Error
}

// GetMessagesByTicketID retrieves all messages for a ticket
func (r *TicketRepository) GetMessagesByTicketID(ctx context.Context, ticketID uint64) ([]*entity.TicketMessage, error) {
	db := r.getDB(ctx)
	var messages []*entity.TicketMessage

	err := db.WithContext(ctx).
		Where("ticket_id = ?", ticketID).
		Order("created_at ASC").
		Find(&messages).Error

	return messages, err
}

// GetLastMessageByTicketID retrieves the last message for a ticket
func (r *TicketRepository) GetLastMessageByTicketID(ctx context.Context, ticketID uint64) (*entity.TicketMessage, error) {
	db := r.getDB(ctx)
	var message entity.TicketMessage

	err := db.WithContext(ctx).
		Where("ticket_id = ?", ticketID).
		Order("created_at DESC").
		First(&message).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &message, nil
}
