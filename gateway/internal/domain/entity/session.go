package entity

import (
	"sync"
	"time"

	"gorm.io/datatypes"
)

type SessionState string

const (
	StateCreated          SessionState = "STEP_START"
	StatePendingOCR       SessionState = "STEP_OCR"
	StateOCRSuccess       SessionState = "STEP_OCR_SUCCESS"
	StateOCRFailed        SessionState = "STEP_OCR_FAILED"
	StatePendingBaseImage SessionState = "STEP_BASE_IMAGE"
	StateBaseImageSuccess SessionState = "STEP_BASE_IMAGE_SUCCESS"
	StateBaseImageFailed  SessionState = "STEP_BASE_IMAGE_FAILED"
	StatePendingLiveness  SessionState = "STEP_LIVENESS"
	StateLivenessSuccess  SessionState = "STEP_LIVENESS_SUCCESS"
	StateLivenessFailed   SessionState = "STEP_LIVENESS_FAILED"
	StateCompleted        SessionState = "STEP_COMPLETED"
	StateCancelled        SessionState = "STEP_CANCELLED"
	StateExpired          SessionState = "STEP_EXPIRED"
)

type Session struct {
	ID               string                 `gorm:"primaryKey;type:uuid" json:"id"`
	ProfileID        uint64                 `gorm:"index" json:"profile_id"`
	WorkflowConfigID uint64                 `json:"workflow_config_id,omitempty"`
	ClientIP         string                 `json:"client_ip"`
	State            SessionState           `gorm:"index" json:"state"`
	CreatedAt        time.Time              `json:"created_at"`
	UpdatedAt        time.Time              `json:"updated_at"`
	ExpiresAt        time.Time              `json:"expires_at"`

	// for in-memory use only
	OCRResult        interface{}            `gorm:"-" json:"ocr_result,omitempty"`
	FaceResult       interface{}            `gorm:"-" json:"face_result,omitempty"`
	Errors           []string               `gorm:"-" json:"errors,omitempty"`
	Metadata         map[string]interface{} `gorm:"-" json:"metadata,omitempty"`

	// for database storage
	OCRResultDB  datatypes.JSON `gorm:"column:ocr_result;type:jsonb" json:"-"`
	FaceResultDB datatypes.JSON `gorm:"column:face_result;type:jsonb" json:"-"`
	ErrorsDB     datatypes.JSON `gorm:"column:errors;type:jsonb" json:"-"`
	MetadataDB   datatypes.JSON `gorm:"column:metadata;type:jsonb" json:"-"`

	mu sync.Mutex `gorm:"-" json:"-"`
}

func (s *Session) Lock() {
	s.mu.Lock()
}

func (s *Session) Unlock() {
	s.mu.Unlock()
}
