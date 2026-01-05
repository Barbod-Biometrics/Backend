package entity

import (
	"time"

	"gorm.io/datatypes"
)

type OCRRecord struct {
	Success bool   `json:"success"`
	Message string `json:"message"`

	NationalID     string `json:"شماره_ملی,omitempty"`
	FirstName      string `json:"نام,omitempty"`
	LastName       string `json:"نام_خانوادگی,omitempty"`
	FatherName     string `json:"نام_پدر,omitempty"`
	BirthDate      string `json:"تاریخ_تولد,omitempty"`
	ExpirationDate string `json:"پایان_اعتبار,omitempty"`

	Stats map[string]interface{} `json:"stats,omitempty" gorm:"type:jsonb"`
}

type OCRModel struct {
	ID        uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	ProfileID uint64         `gorm:"not null;index" json:"profile_id"`
	Success   bool           `gorm:"not null;index" json:"success"`
	Message   string         `gorm:"type:text" json:"message"`
	Stats     datatypes.JSON `gorm:"type:jsonb" json:"stats"`
	CreatedAt time.Time      `gorm:"not null;default:now()" json:"created_at"`
}

func (OCRModel) TableName() string {
	return "ocr_results"
}
