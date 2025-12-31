package postgres

import (
	"context"
	"errors"

	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/entity"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/repository"
	"gorm.io/gorm"
)

type ProfileRepository struct {
	db *gorm.DB
}

func (r *ProfileRepository) GetPersonalProfileByUserID(ctx context.Context, userID uint64) (*entity.Profile, error) {
	db := r.getDB(ctx)
	var profile entity.Profile

	err := db.WithContext(ctx).
		Preload("PersonDetails").
		Preload("BusinessDetails").
		Where("user_id = ? AND profile_type = ?", userID, "personal").
		First(&profile).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
	}

	return &profile, nil
}

func NewProfileRepository(db *gorm.DB) *ProfileRepository {
	return &ProfileRepository{db: db}
}

func (r *ProfileRepository) getDB(ctx context.Context) *gorm.DB {
	if tx, ok := ctx.Value("db_tx").(*gorm.DB); ok {
		return tx
	}
	return r.db
}

func (r *ProfileRepository) Create(ctx context.Context, profile *entity.Profile) error {
	db := r.getDB(ctx)
	return db.WithContext(ctx).Create(profile).Error
}

func (r *ProfileRepository) Update(ctx context.Context, profile *entity.Profile) error {
	db := r.getDB(ctx)
	return db.WithContext(ctx).Session(&gorm.Session{FullSaveAssociations: true}).Save(profile).Error
}

func (r *ProfileRepository) GetByID(ctx context.Context, profileID uint64) (*entity.Profile, error) {
	db := r.getDB(ctx)
	var profile entity.Profile

	err := db.WithContext(ctx).
		Preload("PersonDetails").
		Preload("BusinessDetails").
		Preload("BusinessDetails.Signatories").
		First(&profile, profileID).Error

	if err != nil {
		return nil, err
	}
	return &profile, nil
}

func (r *ProfileRepository) GetByUserID(ctx context.Context, userID uint64) ([]*entity.Profile, error) {
	db := r.getDB(ctx)
	var profiles []*entity.Profile

	err := db.WithContext(ctx).
		Preload("PersonDetails").
		Preload("BusinessDetails").
		Where("user_id = ?", userID).
		Find(&profiles).Error

	return profiles, err
}

func (r *ProfileRepository) ListWithFilters(ctx context.Context, filter repository.ProfileFilter, sort repository.ProfileSort, pagination repository.Pagination) (*repository.PaginatedResult, error) {
	db := r.getDB(ctx)
	query := db.WithContext(ctx).Model(&entity.Profile{})

	// Apply filters
	if filter.Status != nil {
		query = query.Where("verification_status = ?", *filter.Status)
	}
	if filter.ProfileType != nil {
		query = query.Where("profile_type = ?", *filter.ProfileType)
	}
	if filter.SearchQuery != nil && *filter.SearchQuery != "" {
		searchPattern := "%" + *filter.SearchQuery + "%"
		query = query.Where(
			db.Where("profile_name ILIKE ?", searchPattern).
				Or("profile_id IN (?)",
					db.Model(&entity.ProfilePersonDetails{}).
						Select("profile_id").
						Where("national_id ILIKE ? OR mobile_number ILIKE ? OR first_name ILIKE ? OR last_name ILIKE ?",
							searchPattern, searchPattern, searchPattern, searchPattern),
				).
				Or("profile_id IN (?)",
					db.Model(&entity.ProfileBusinessDetails{}).
						Select("profile_id").
						Where("rep_national_id ILIKE ? OR rep_mobile_number ILIKE ? OR business_national_id ILIKE ?",
							searchPattern, searchPattern, searchPattern),
				),
		)
	}

	// Count total
	var totalCount int64
	if err := query.Count(&totalCount).Error; err != nil {
		return nil, err
	}

	// Apply sorting
	sortField := "created_at"
	sortOrder := "desc"
	if sort.Field != "" {
		allowedFields := map[string]bool{
			"created_at":          true,
			"profile_name":        true,
			"verification_status": true,
			"profile_type":        true,
		}
		if allowedFields[sort.Field] {
			sortField = sort.Field
		}
	}
	if sort.Order == "asc" || sort.Order == "desc" {
		sortOrder = sort.Order
	}
	query = query.Order(sortField + " " + sortOrder)

	// Apply pagination
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
	query = query.Offset(offset).Limit(pagination.PageSize)

	// Execute query
	var profiles []*entity.Profile
	if err := query.
		Preload("PersonDetails").
		Preload("BusinessDetails").
		Preload("BusinessDetails.Signatories").
		Find(&profiles).Error; err != nil {
		return nil, err
	}

	totalPages := int(totalCount) / pagination.PageSize
	if int(totalCount)%pagination.PageSize > 0 {
		totalPages++
	}

	return &repository.PaginatedResult{
		Profiles:   profiles,
		TotalCount: totalCount,
		Page:       pagination.Page,
		PageSize:   pagination.PageSize,
		TotalPages: totalPages,
	}, nil
}

func (r *ProfileRepository) UpdateHasAPIKey(ctx context.Context, profileID uint64, hasKey bool) error {
	db := r.getDB(ctx)

	err := db.Model(&entity.Profile{}).
		Where("id = ?", profileID).
		Update("has_apikey", hasKey).Error

	return err
}
