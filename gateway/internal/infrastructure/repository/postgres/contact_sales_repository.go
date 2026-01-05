package postgres

import (
	"context"
	"time"

	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/entity"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/repository"
	"gorm.io/gorm"
)

type ContactSalesRepository struct {
	db *gorm.DB
}

func NewContactSalesRepository(db *gorm.DB) repository.ContactSalesRepository {
	return &ContactSalesRepository{db: db}
}

func (r *ContactSalesRepository) getDB(ctx context.Context) *gorm.DB {
	if tx, ok := ctx.Value("db_tx").(*gorm.DB); ok {
		return tx
	}
	return r.db
}

func (r *ContactSalesRepository) Create(ctx context.Context, req *entity.ContactSalesRequest) error {
	return r.getDB(ctx).WithContext(ctx).Create(req).Error
}

func (r *ContactSalesRepository) List(ctx context.Context, filter repository.ContactSalesFilter, sort repository.ContactSalesSort, pagination repository.Pagination) (*repository.ContactSalesPaginatedResult, error) {
	db := r.getDB(ctx).WithContext(ctx).Model(&entity.ContactSalesRequest{})

	if filter.Status != nil {
		db = db.Where("status = ?", *filter.Status)
	}

	if filter.Search != nil && *filter.Search != "" {
		pattern := "%" + *filter.Search + "%"
		db = db.Where(
			db.Where("first_name ILIKE ?", pattern).
				Or("last_name ILIKE ?", pattern).
				Or("phone ILIKE ?", pattern).
				Or("business_name ILIKE ?", pattern).
				Or("email ILIKE ?", pattern),
		)
	}

	if filter.HasEmail != nil {
		if *filter.HasEmail {
			db = db.Where("email IS NOT NULL")
		} else {
			db = db.Where("email IS NULL")
		}
	}

	if filter.FromDate != nil {
		db = db.Where("created_at >= ?", *filter.FromDate)
	}
	if filter.ToDate != nil {
		to := filter.ToDate.Add(24 * time.Hour)
		db = db.Where("created_at < ?", to)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, err
	}

	sortField := "created_at"
	allowedFields := map[string]bool{
		"created_at": true,
		"status":     true,
	}
	if allowedFields[sort.Field] {
		sortField = sort.Field
	}

	sortOrder := "desc"
	if sort.Order == "asc" || sort.Order == "desc" {
		sortOrder = sort.Order
	}
	db = db.Order(sortField + " " + sortOrder)

	if pagination.Page < 1 {
		pagination.Page = 1
	}
	if pagination.PageSize < 1 {
		pagination.PageSize = 10
	}
	if pagination.PageSize > 100 {
		pagination.PageSize = 100
	}
	offset := (pagination.Page - 1) * pagination.PageSize
	db = db.Offset(offset).Limit(pagination.PageSize)

	var items []*entity.ContactSalesRequest
	if err := db.Find(&items).Error; err != nil {
		return nil, err
	}

	totalPages := int(total) / pagination.PageSize
	if int(total)%pagination.PageSize > 0 {
		totalPages++
	}

	return &repository.ContactSalesPaginatedResult{
		Items:      items,
		TotalCount: total,
		Page:       pagination.Page,
		PageSize:   pagination.PageSize,
		TotalPages: totalPages,
	}, nil
}

func (r *ContactSalesRepository) MarkRead(ctx context.Context, id uint64) error {
	now := time.Now()
	result := r.getDB(ctx).WithContext(ctx).
		Model(&entity.ContactSalesRequest{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":  entity.ContactSalesStatusRead,
			"read_at": &now,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *ContactSalesRepository) Delete(ctx context.Context, id uint64) error {
	result := r.getDB(ctx).WithContext(ctx).Delete(&entity.ContactSalesRequest{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
