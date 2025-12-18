package postgres

import (
	"context"
	"encoding/json"
	"time"

	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/entity"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type FaceVerificationModel struct {
	ID                    uint64  `gorm:"primaryKey;autoIncrement" json:"id"`
	ProfileID             uint64  `gorm:"not null;index" json:"profile_id"`
	Success               bool    `gorm:"not null;index" json:"success"`
	Reason                string  `gorm:"type:text" json:"reason"`
	Message               string  `gorm:"type:text" json:"message"`
	HighestSimilarity     float64 `gorm:"type:double precision" json:"highest_similarity"`
	AverageSimilarity     float64 `gorm:"type:double precision" json:"average_similarity"`
	MatchedFrames         int     `gorm:"type:int" json:"matched_frames"`
	FacesDetected         int     `gorm:"type:int" json:"faces_detected"`
	ProcessingTimeSeconds float64 `gorm:"type:double precision" json:"processing_time_seconds"`

	RealFrames       int            `gorm:"type:int" json:"real_frames"`
	RealRate         float64        `gorm:"type:double precision" json:"real_rate"`
	SpoofRate        float64        `gorm:"type:double precision" json:"spoof_rate"`
	SpoofedFrames    int            `gorm:"type:int" json:"spoofed_frames"`
	TotalFrames      int            `gorm:"type:int" json:"total_frames"`
	TwoFacesFrames   int            `gorm:"type:int" json:"two_faces_frames"`
	TwoFacesRate     float64        `gorm:"type:double precision" json:"two_faces_rate"`
	VerificationRate float64        `gorm:"type:double precision" json:"verification_rate"`
	VerifiedFrames   int            `gorm:"type:int" json:"verified_frames"`
	Stats            datatypes.JSON `gorm:"type:jsonb" json:"stats"`
	CreatedAt        time.Time      `gorm:"not null;default:now()" json:"created_at"`
}

func (FaceVerificationModel) TableName() string {
	return "face_verifications"
}

type FaceVerificationRepository struct {
	db *gorm.DB
}

func NewFaceVerificationRepository(db *gorm.DB) *FaceVerificationRepository {
	return &FaceVerificationRepository{db: db}
}

func (r *FaceVerificationRepository) getDB(ctx context.Context) *gorm.DB {
	if tx, ok := ctx.Value("db_tx").(*gorm.DB); ok {
		return tx
	}
	return r.db
}

func (r *FaceVerificationRepository) SaveResult(ctx context.Context, profileID uint64, result *entity.FaceVerificationRecord) error {
	db := r.getDB(ctx)

	var statsBytes datatypes.JSON
	if result != nil && result.Stats != nil {
		b, err := json.Marshal(result.Stats)
		if err != nil {
			return err
		}
		statsBytes = datatypes.JSON(b)
	} else {
		statsBytes = datatypes.JSON([]byte("null"))
	}

	model := &FaceVerificationModel{
		ProfileID:             profileID,
		Success:               result != nil && result.Success,
		Reason:                "",
		Message:               "",
		HighestSimilarity:     0,
		AverageSimilarity:     0,
		MatchedFrames:         0,
		FacesDetected:         0,
		ProcessingTimeSeconds: 0,
		RealFrames:            0,
		RealRate:              0,
		SpoofRate:             0,
		SpoofedFrames:         0,
		TotalFrames:           0,
		TwoFacesFrames:        0,
		TwoFacesRate:          0,
		VerificationRate:      0,
		VerifiedFrames:        0,
		Stats:                 statsBytes,
		CreatedAt:             time.Now(),
	}

	if result != nil {
		model.Reason = result.Reason
		model.Message = result.Message

		model.HighestSimilarity = result.HighestSimilarity
		model.AverageSimilarity = result.AverageSimilarity
		model.MatchedFrames = result.MatchedFrames
		model.FacesDetected = result.FacesDetected
		model.ProcessingTimeSeconds = result.ProcessingTimeSeconds

		model.RealFrames = result.RealFrames
		model.RealRate = result.RealRate
		model.SpoofRate = result.SpoofRate
		model.SpoofedFrames = result.SpoofedFrames
		model.TotalFrames = result.TotalFrames
		model.TwoFacesFrames = result.TwoFacesFrames
		model.TwoFacesRate = result.TwoFacesRate
		model.VerificationRate = result.VerificationRate
		model.VerifiedFrames = result.VerifiedFrames

		if result.Stats != nil {
			if model.RealFrames == 0 {
				model.RealFrames = result.Stats.RealFrames
			}
			if model.RealRate == 0 {
				model.RealRate = result.Stats.RealRate
			}
			if model.SpoofRate == 0 {
				model.SpoofRate = result.Stats.SpoofRate
			}
			if model.SpoofedFrames == 0 {
				model.SpoofedFrames = result.Stats.SpoofedFrames
			}
			if model.TotalFrames == 0 {
				model.TotalFrames = result.Stats.TotalFrames
			}
			if model.TwoFacesFrames == 0 {
				model.TwoFacesFrames = result.Stats.TwoFacesFrames
			}
			if model.TwoFacesRate == 0 {
				model.TwoFacesRate = result.Stats.TwoFacesRate
			}
			if model.VerificationRate == 0 {
				model.VerificationRate = result.Stats.VerificationRate
			}
			if model.VerifiedFrames == 0 {
				model.VerifiedFrames = result.Stats.VerifiedFrames
			}
		}
	}

	return db.WithContext(ctx).Create(model).Error
}

func (r *FaceVerificationRepository) GetResultsByProfileID(ctx context.Context, profileID uint64) ([]*entity.FaceVerificationRecord, error) {
	db := r.getDB(ctx)

	var rows []FaceVerificationModel
	if err := db.WithContext(ctx).
		Where("profile_id = ?", profileID).
		Order("created_at DESC").
		Find(&rows).Error; err != nil {
		return nil, err
	}

	results := make([]*entity.FaceVerificationRecord, 0, len(rows))
	for _, rmodel := range rows {
		rec := &entity.FaceVerificationRecord{
			Success:               rmodel.Success,
			Reason:                rmodel.Reason,
			Message:               rmodel.Message,
			HighestSimilarity:     rmodel.HighestSimilarity,
			AverageSimilarity:     rmodel.AverageSimilarity,
			MatchedFrames:         rmodel.MatchedFrames,
			FacesDetected:         rmodel.FacesDetected,
			ProcessingTimeSeconds: rmodel.ProcessingTimeSeconds,
		}

		// populate model-provided aggregates
		rec.RealFrames = rmodel.RealFrames
		rec.RealRate = rmodel.RealRate
		rec.SpoofRate = rmodel.SpoofRate
		rec.SpoofedFrames = rmodel.SpoofedFrames
		rec.TotalFrames = rmodel.TotalFrames
		rec.TwoFacesFrames = rmodel.TwoFacesFrames
		rec.TwoFacesRate = rmodel.TwoFacesRate
		rec.VerificationRate = rmodel.VerificationRate
		rec.VerifiedFrames = rmodel.VerifiedFrames

		if len(rmodel.Stats) > 0 && string(rmodel.Stats) != "null" {
			var stats entity.FaceVerificationStats
			if err := json.Unmarshal(rmodel.Stats, &stats); err == nil {
				rec.Stats = &stats
			}
		}

		results = append(results, rec)
	}

	return results, nil
}
