package entity

import "time"

type WorkflowConfig struct {
	ID                    uint64 `gorm:"primaryKey"`
	ProfileID             uint64 `gorm:"index"`
	Name                  string `gorm:"type:varchar(255)"`
	Instruction           string `gorm:"type:text"`
	LivenessSentence      string `gorm:"type:text"`
	OCRAcceptanceRequired bool   `gorm:"default:false"`
	CreatedAt             time.Time
	UpdatedAt             time.Time
}
