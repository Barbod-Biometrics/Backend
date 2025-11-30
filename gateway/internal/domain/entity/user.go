package entity

import (
	"time"
)

type User struct {
	UserID      uint64    `gorm:"primaryKey;autoIncrement" json:"user_id"`
	PhoneNumber string    `gorm:"type:varchar(15);unique;not null" json:"phone_number"`
	Email       *string   `gorm:"type:varchar(100);unique" json:"email,omitempty"`
	CreatedAt   time.Time `gorm:"not null;default:now()" json:"created_at"`
}

func (u *User) TableName() string {
	return "users"
}
