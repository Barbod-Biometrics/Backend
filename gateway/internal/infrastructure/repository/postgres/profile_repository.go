package postgres

import (
	"context"
	"errors"

	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/entity"
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
