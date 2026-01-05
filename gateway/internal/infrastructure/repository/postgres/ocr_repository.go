package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/entity"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/logger"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type OCRRepository struct {
	db     *gorm.DB
	logger logger.Logger
}

func NewOCRRepository(db *gorm.DB, logger logger.Logger) *OCRRepository {
	return &OCRRepository{
		db:     db,
		logger: logger,
	}
}

func (r *OCRRepository) getDB(ctx context.Context) *gorm.DB {
	type ctxKey string
	const dbTxKey ctxKey = "db_tx"
	if tx, ok := ctx.Value(dbTxKey).(*gorm.DB); ok {
		return tx
	}
	return r.db
}

func (r *OCRRepository) SaveResult(ctx context.Context, profileID uint64, result *entity.OCRRecord) error {
	db := r.getDB(ctx)

	var statsBytes datatypes.JSON
	if result != nil && result.Stats != nil {
		b, err := json.Marshal(result.Stats)
		if err != nil {
			r.logger.Error("Failed to marshal OCR stats to JSON",
				logger.Field{Key: "error", Value: err},
				logger.Field{Key: "profile_id", Value: profileID},
			)
			return err
		}
		statsBytes = datatypes.JSON(b)
	} else {
		statsBytes = datatypes.JSON([]byte("null"))
	}

	model := &entity.OCRModel{
		ProfileID: profileID,
		Success:   result != nil && result.Success,
		Message:   "",
		Stats:     statsBytes,
		CreatedAt: time.Now(),
	}

	if result != nil {
		model.Message = result.Message
	}

	err := db.WithContext(ctx).Create(model).Error
	if err != nil {
		r.logger.Error("Failed to save OCR result to database",
			logger.Field{Key: "error", Value: err},
			logger.Field{Key: "profile_id", Value: profileID},
		)
		return err
	}

	return nil
}

func (r *OCRRepository) GetResultsByProfileID(ctx context.Context, profileID uint64) ([]*entity.OCRRecord, error) {
	db := r.getDB(ctx)

	var rows []*entity.OCRModel
	if err := db.WithContext(ctx).
		Where("profile_id = ?", profileID).
		Order("created_at DESC").
		Find(&rows).Error; err != nil {
		r.logger.Error("Failed to fetch OCR results from database",
			logger.Field{Key: "error", Value: err},
			logger.Field{Key: "profile_id", Value: profileID},
		)
		return nil, err
	}

	results := make([]*entity.OCRRecord, 0, len(rows))
	for _, rmodel := range rows {
		rec := &entity.OCRRecord{
			Success: rmodel.Success,
			Message: rmodel.Message,
		}

		if len(rmodel.Stats) > 0 && string(rmodel.Stats) != "null" {
			var statsMap map[string]interface{}
			if err := json.Unmarshal(rmodel.Stats, &statsMap); err == nil {
				rec.Stats = statsMap

				if v, ok := statsMap["شماره_ملی"]; ok {
					if s, ok2 := v.(string); ok2 {
						rec.NationalID = s
					}
				}
				if v, ok := statsMap["نام"]; ok {
					if s, ok2 := v.(string); ok2 {
						rec.FirstName = s
					}
				}
				if v, ok := statsMap["نام_خانوادگی"]; ok {
					if s, ok2 := v.(string); ok2 {
						rec.LastName = s
					}
				}
				if v, ok := statsMap["نام_پدر"]; ok {
					if s, ok2 := v.(string); ok2 {
						rec.FatherName = s
					}
				}
				if v, ok := statsMap["تاریخ_تولد"]; ok {
					if s, ok2 := v.(string); ok2 {
						rec.BirthDate = s
					}
				}
				if v, ok := statsMap["پایان_اعتبار"]; ok {
					if s, ok2 := v.(string); ok2 {
						rec.ExpirationDate = s
					}
				} else {
					// We don't return error here because we still want to show the partial record
					r.logger.Warn("Failed to unmarshal stored OCR stats JSON",
						logger.Field{Key: "error", Value: err},
						logger.Field{Key: "ocr_record_id", Value: rmodel.ID},
					)
				}
			}
		}

		results = append(results, rec)
	}

	return results, nil
}

func (r *OCRRepository) ApproveOCRResult(ctx context.Context, ocrID uint64, profileID uint64) error {
	db := r.getDB(ctx)

	result := db.WithContext(ctx).
		Model(&entity.OCRModel{}).
		Where("id = ? AND profile_id = ?", ocrID, profileID).
		Update("approved", true)

	if result.Error != nil {
		r.logger.Error("Failed to execute update query for OCR approval",
			logger.Field{Key: "error", Value: result.Error},
			logger.Field{Key: "ocr_id", Value: ocrID},
			logger.Field{Key: "profile_id", Value: profileID},
		)
		return result.Error
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("record not found or access denied")
	}

	return nil
}
