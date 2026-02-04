package entity

import (
	"time"

	"gorm.io/datatypes"
)

type FaceVerificationStats struct {
	Duration float64 `json:"duration,omitempty"`
	FPS      int     `json:"fps,omitempty"`
	Width    int     `json:"width,omitempty"`
	Height   int     `json:"height,omitempty"`

	TotalFrames      int     `json:"total_frames,omitempty"`
	ProcessedFrames  int     `json:"processed_frames,omitempty"`
	RealFrames       int     `json:"real_frames,omitempty"`
	RealRate         float64 `json:"real_rate,omitempty"`
	SpoofedFrames    int     `json:"spoofed_frames,omitempty"`
	SpoofRate        float64 `json:"spoof_rate,omitempty"`
	TwoFacesFrames   int     `json:"two_faces_frames,omitempty"`
	TwoFacesRate     float64 `json:"two_faces_rate,omitempty"`
	VerifiedFrames   int     `json:"verified_frames,omitempty"`
	VerificationRate float64 `json:"verification_rate,omitempty"`

	FacesDetected         int     `json:"faces_detected,omitempty"`
	MatchedFrames         int     `json:"matched_frames,omitempty"`
	HighestSimilarity     float64 `json:"highest_similarity,omitempty"`
	AverageSimilarity     float64 `json:"average_similarity,omitempty"`
	ProcessingTimeSeconds float64 `json:"processing_time_seconds,omitempty"`
	LowestSpoofScore      float64 `json:"lowest_spoof_score,omitempty"`
	HighestSpoofScore     float64 `json:"highest_spoof_score,omitempty"`
	AverageSpoofScore     float64 `json:"average_spoof_score,omitempty"`
	LowestLivenessScore   float64 `json:"lowest_liveness_score,omitempty"`
	HighestLivenessScore  float64 `json:"highest_liveness_score,omitempty"`
	AverageLivenessScore  float64 `json:"average_liveness_score,omitempty"`
}

type FaceVerificationRecord struct {
	Success               bool                   `json:"success"`
	Reason                string                 `json:"reason"`
	Message               string                 `json:"message"`
	Stats                 *FaceVerificationStats `json:"stats,omitempty" gorm:"type:jsonb"`
	HighestSimilarity     float64                `json:"highest_similarity,omitempty"`
	ProcessingTimeSeconds float64                `json:"processing_time_seconds,omitempty"`
}

type FaceVerificationModel struct {
	ID                    uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	ProfileID             uint64         `gorm:"not null;index" json:"profile_id"`
	Success               bool           `gorm:"not null;index" json:"success"`
	Reason                string         `gorm:"type:text" json:"reason"`
	Message               string         `gorm:"type:text" json:"message"`
	HighestSimilarity     float64        `gorm:"type:double precision;index" json:"highest_similarity"`
	ProcessingTimeSeconds float64        `gorm:"type:double precision" json:"processing_time_seconds"`
	Stats                 datatypes.JSON `gorm:"type:jsonb" json:"stats"`
	CreatedAt             time.Time      `gorm:"not null;default:now()" json:"created_at"`
}

func (FaceVerificationModel) TableName() string {
	return "face_verifications"
}
