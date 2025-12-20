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

type FaceVerificationRepository struct {
	db *gorm.DB
}

func NewFaceVerificationRepository(db *gorm.DB) *FaceVerificationRepository {
	return &FaceVerificationRepository{db: db}
}

func (r *FaceVerificationRepository) getDB(ctx context.Context) *gorm.DB {
	type ctxKey string
	const dbTxKey ctxKey = "db_tx"
	if tx, ok := ctx.Value(dbTxKey).(*gorm.DB); ok {
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
		ProcessingTimeSeconds: 0,
		Stats:                 statsBytes,
		CreatedAt:             time.Now(),
	}

	if result != nil {
		model.Reason = result.Reason
		model.Message = result.Message
		model.HighestSimilarity = result.HighestSimilarity
		model.ProcessingTimeSeconds = result.ProcessingTimeSeconds
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
			ProcessingTimeSeconds: rmodel.ProcessingTimeSeconds,
		}

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
