package entity

import (
	"time"

	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/enum"
	"github.com/google/uuid"
)

// Transaction represents a wallet transaction (deposit or withdrawal)
// Amount is stored in Iranian Rials (no decimals, can be very large numbers)
type Transaction struct {
	TransactionID   uint64               `gorm:"primaryKey;autoIncrement" json:"transaction_id"`
	ProfileID       uint64               `gorm:"not null;index;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;references:ProfileID" json:"profile_id"`
	JobID           *uuid.UUID           `gorm:"type:uuid" json:"job_id,omitempty"`
	TransactionType enum.TransactionType `gorm:"type:varchar(20);not null" json:"transaction_type"`
	Amount          int64                `gorm:"type:bigint;not null" json:"amount"` // Positive for deposit, negative for withdrawal
	Notes           string               `gorm:"type:text" json:"notes,omitempty"`
	CreatedAt       time.Time            `gorm:"not null;default:now()" json:"created_at"`
}

func (Transaction) TableName() string {
	return "transactions"
}
