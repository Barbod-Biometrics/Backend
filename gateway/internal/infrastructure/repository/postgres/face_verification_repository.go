package postgres

import (
	"context"
	"encoding/json"
	"time"

	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/entity"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/repository"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

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

	model := &entity.FaceVerificationModel{
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

	var rows []*entity.FaceVerificationModel
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

func (r *FaceVerificationRepository) GetReports(ctx context.Context, filter repository.FaceReportFilter) ([]*entity.FaceVerificationModel, int64, error) {
	db := r.getDB(ctx)

	var jobs []*entity.FaceVerificationModel
	var totalCount int64

	query := db.WithContext(ctx).Model(&entity.FaceVerificationModel{})

	// filter on profile id(madatory)
	query = query.Where("profile_id = ?", filter.PofileID)

	// status filter
	if filter.Status != nil {
		if *filter.Status == "success" {
			query = query.Where("success = ?", true)
		} else if *filter.Status == "failed" {
			query = query.Where("success = ?", false)
		}
	}

	// date filter
	if filter.FromDate != nil {
		query = query.Where("created_at >= ?", *filter.FromDate)
	}
	if filter.ToDate != nil {
		query = query.Where("created_at <= ?", *filter.ToDate)
	}

	// for pagination
	if err := query.Count(&totalCount).Error; err != nil {
		return nil, 0, err
	}

	// sorting
	sortString := "created_at DESC" // default: newwest first

	if filter.SortBy != "date" {
		if filter.SortOrder == "asc" {
			sortString = "created_at ASC"
		} else {
			sortString = "created_at DESC"
		}
	} else if filter.SortBy == "rate" {
		if filter.SortOrder == "asc" {
			sortString = "highest_similarity ASC"
		} else {
			sortString = "highest_similarity DESC"
		}
	}

	// pagination
	offset := (filter.Page - 1) * filter.Limit
	err := query.Order(sortString).
		Limit(filter.Limit).
		Offset(offset).
		Find(&jobs).Error

	return jobs, totalCount, err

}
