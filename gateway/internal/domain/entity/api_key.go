package entity

import "time"

type APIKey struct {
	KeyID     uint64    `gorm:"primaryKey;autoIncrement" json:"key_id"`
	ProfileID uint64    `gorm:"not null;index" json:"profile_id"`
	KeyHash   string    `gorm:"type:varchar(256); not null; unique" json:"key_hash"`
	KeyPrefix string    `gorm:"type:varchar(8); not null; unique" json:"key_prefix"`
	IsActive  bool      `gorm:"not null; default:true" json:"is_active"`
	CreatedAt time.Time `gorm:"not null;default:now()" json:"created_at"`
}

func (APIKey) TableName() string {
	return "api_keys"
}
