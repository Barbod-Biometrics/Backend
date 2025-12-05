package entity

import (
	"encoding/json"
	"time"

	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/enum"
)

type VerificationJob struct {
	JobID            uint64           `gorm:"primaryKey;autoIncrement" json:"job_id"`
	ProfileID        uint64           `gorm:"not null;index" json:"profile_id"`
	APIKeyID         uint64           `gorm:"type:uuid;not null;index" json:"api_key_id"`
	ServiceID        uint64           `gorm:"not null" json:"service_id"`
	Status           enum.JobStatus   `gorm:"type:job_status_enum;not null;default:'pending'" json:"status"`
	JobCost          float64          `gorm:"type:decimal(10, 2);not null" json:"job_cost"`
	Result           *json.RawMessage `gorm:"type:jsonb" json:"result,omitempty"`
	IDCardURL        string           `gorm:"type:string" json:"id_card_urlmomitempty"`
	FaceImageURL     string           `gorm:"type:string" json:"face_image_url,omitempty"`
	LivenessVideoURL string           `gorm:"type:string" json:"liveness_video_url,omitempty"`
	CreatedAt        time.Time        `gorm:"not null;default:now()" json:"created_at"`
	CompletedAt      *time.Time       `json:"completed_at,omitempty"`
}

func (VerificationJob) TableName() string {
	return "verification_job"
}
