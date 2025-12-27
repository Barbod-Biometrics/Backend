package entity

import (
	"time"

	"gorm.io/gorm"
)

type ContactSalesStatus string

const (
	ContactSalesStatusNew  ContactSalesStatus = "new"
	ContactSalesStatusRead ContactSalesStatus = "read"
)

type ContactSalesRequest struct {
	ID           uint64             `gorm:"primaryKey;autoIncrement" json:"id"`
	FirstName    string             `gorm:"size:100;not null" json:"first_name"`
	LastName     string             `gorm:"size:100;not null" json:"last_name"`
	Phone        string             `gorm:"size:15;not null" json:"phone"`
	BusinessName string             `gorm:"size:150;not null" json:"business_name"`
	Email        *string            `gorm:"size:150" json:"email,omitempty"`
	Description  string             `gorm:"type:text;not null" json:"description"`
	Status       ContactSalesStatus `gorm:"type:varchar(20);not null;default:'new'" json:"status"`
	ReadAt       *time.Time         `json:"read_at,omitempty"`
	CreatedAt    time.Time          `gorm:"not null;default:now()" json:"created_at"`
	UpdatedAt    time.Time          `gorm:"not null;default:now()" json:"updated_at"`
	DeletedAt    gorm.DeletedAt     `gorm:"index" json:"-"`
}

func (c *ContactSalesRequest) TableName() string {
	return "contact_sales_requests"
}
