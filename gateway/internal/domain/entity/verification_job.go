package entity

import (
	"time"

	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/enum"
	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type VerificationJob struct {
	JobID             uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"job_id"`
	ProfileID         uint64         `gorm:"not null;index" json:"profile_id"`
	ApiKeyID          uint64         `gorm:"not null;incex" json:"api_key_id"`
	ServiceID         uint64         `gorm:"not null" json:"service_id"`
	Status            enum.JobStatus `gorm:"type:job_status_enum;not null;default:'pending'" json:"status"`
	JobCost           int64          `gorm:"type:decimal(10,2);not null" json:"job_cost"`
	DocVerifiedResult datatypes.JSON `gorm:"type:jsonb" json:"doc_verify_result,omitempty"`
	FaceMatchResult   datatypes.JSON `gorm:"type:jsonb" json:"face_match_result,omitempty"`
	LivenessResult    datatypes.JSON `gorm:"type:jsonb" json:"liveness_result,omitempty"`
	IdCardURL         string         `gorm:"type:text" json:"id_card_url,omitempty"`
	FaceImageURL      string         `gorm:"type:text" json:"face_image_url,omitempty"`
	LivenessVideoURL  string         `gorm:"type:text" json:"liveness_video_url,omitempty"`
	CreatedAt         time.Time      `gorm:"not null;default:now()" json:"created_at"`
	CompletedAt       *time.Time     `json:"completed_at,omitempty"`
}

func (VerificationJob) TableName() string {
	return "verification_jobs"
}
