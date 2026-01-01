package postgres

import (
	"context"
	"encoding/json"
	"time"

	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/entity"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type OCRRepository struct {
	db *gorm.DB
}

func NewOCRRepository(db *gorm.DB) *OCRRepository {
	return &OCRRepository{db: db}
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

	return db.WithContext(ctx).Create(model).Error
}

func (r *OCRRepository) GetResultsByProfileID(ctx context.Context, profileID uint64) ([]*entity.OCRRecord, error) {
	db := r.getDB(ctx)

	var rows []*entity.OCRModel
	if err := db.WithContext(ctx).
		Where("profile_id = ?", profileID).
		Order("created_at DESC").
		Find(&rows).Error; err != nil {
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
				}
			}
		}

		results = append(results, rec)
	}

	return results, nil
}
