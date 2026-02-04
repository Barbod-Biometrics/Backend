package entity

import (
	"time"

	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/enum"
)

type Ticket struct {
	TicketID      uint64             `gorm:"primaryKey;autoIncrement" json:"ticket_id"`
	UserID        uint64             `gorm:"not null;index" json:"user_id"`
	Service       enum.TicketService `gorm:"type:varchar(50);not null" json:"service"`
	Title         string             `gorm:"type:varchar(255);not null" json:"title"`
	Description   string             `gorm:"type:text;not null" json:"description"`
	Status        enum.TicketStatus  `gorm:"type:varchar(20);default:'pending';not null" json:"status"`
	AttachmentKey *string            `gorm:"type:varchar(500)" json:"attachment_key,omitempty"` // MinIO object key
	CreatedAt     time.Time          `gorm:"not null;default:now()" json:"created_at"`
	UpdatedAt     time.Time          `gorm:"not null;default:now()" json:"updated_at"`

	// Relations - User is excluded from GORM (loaded via manual join)
	User     *User           `gorm:"-" json:"user,omitempty"`
	Messages []TicketMessage `gorm:"-" json:"messages,omitempty"` // Loaded manually, no FK constraint
}

func (Ticket) TableName() string {
	return "tickets"
}

type TicketMessage struct {
	MessageID uint64             `gorm:"primaryKey;autoIncrement" json:"message_id"`
	TicketID  uint64             `gorm:"not null;index" json:"ticket_id"`
	Sender    enum.MessageSender `gorm:"type:varchar(20);not null" json:"sender"`
	Message   string             `gorm:"type:text;not null" json:"message"`
	CreatedAt time.Time          `gorm:"not null;default:now()" json:"created_at"`
}

func (TicketMessage) TableName() string {
	return "ticket_messages"
}
