package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/entity"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/test/testcontainer"
	"github.com/stretchr/testify/assert"
)

func TestGetPersonalProfileByUserID_FoundAndNotFound(t *testing.T) {
	ctx := context.Background()
	pg, err := testcontainer.SetupPostgres(ctx)
	if err != nil {
		t.Fatalf("failed to setup postgres container: %v", err)
	}
	defer pg.Terminate(ctx)

	repo := NewProfileRepository(pg.DB)

	// seed data
	userID := uint64(42)
	profile := entity.Profile{
		ProfileID:          100,
		UserID:             userID,
		ProfileType:        "personal",
		ProfileName:        "john",
		Balance:            0,
		VerificationStatus: "verified",
		IsActive:           true,
		CreatedAt:          time.Now(),
	}
	err = pg.DB.Create(&profile).Error
	assert.NoError(t, err)

	// success case
	{
		p, err := repo.GetPersonalProfileByUserID(context.Background(), userID)
		assert.NoError(t, err)
		assert.NotNil(t, p)
		assert.Equal(t, uint64(100), p.ProfileID)
		assert.Equal(t, userID, p.UserID)
		assert.Equal(t, entity.ProfileTypePersonal, p.ProfileType)
	}

	// not found
	{
		userIDNotFound := uint64(99)
		p, err := repo.GetPersonalProfileByUserID(context.Background(), userIDNotFound)
		// Repository swallows ErrRecordNotFound and returns nil, nil
		assert.NoError(t, err)
		assert.Nil(t, p)
	}
}

func TestGetByID_Success(t *testing.T) {
	ctx := context.Background()
	pg, err := testcontainer.SetupPostgres(ctx)
	if err != nil {
		t.Fatalf("failed to setup postgres container: %v", err)
	}
	defer pg.Terminate(ctx)

	repo := NewProfileRepository(pg.DB)

	profileID := uint64(10)
	profile := entity.Profile{
		ProfileID:          profileID,
		UserID:             7,
		ProfileType:        "personal",
		ProfileName:        "alice",
		Balance:            10,
		VerificationStatus: "verified",
		IsActive:           true,
		CreatedAt:          time.Now(),
	}
	err = pg.DB.Create(&profile).Error
	assert.NoError(t, err)

	p, err := repo.GetByID(context.Background(), profileID)
	assert.NoError(t, err)
	assert.NotNil(t, p)
	assert.Equal(t, profileID, p.ProfileID)
	assert.Equal(t, "alice", p.ProfileName)
}

func TestGetByUserID_Success(t *testing.T) {
	ctx := context.Background()
	pg, err := testcontainer.SetupPostgres(ctx)
	if err != nil {
		t.Fatalf("failed to setup postgres container: %v", err)
	}
	defer pg.Terminate(ctx)

	repo := NewProfileRepository(pg.DB)

	userID := uint64(7)
	profiles := []entity.Profile{
		{
			ProfileID:          11,
			UserID:             userID,
			ProfileType:        "personal",
			ProfileName:        "a",
			Balance:            0,
			VerificationStatus: "draft",
			IsActive:           true,
			CreatedAt:          time.Now(),
		},
		{
			ProfileID:          12,
			UserID:             userID,
			ProfileType:        "business",
			ProfileName:        "b",
			Balance:            0,
			VerificationStatus: "draft",
			IsActive:           true,
			CreatedAt:          time.Now(),
		},
	}
	err = pg.DB.Create(&profiles).Error
	assert.NoError(t, err)

	resultProfiles, err := repo.GetByUserID(context.Background(), userID)
	assert.NoError(t, err)
	assert.Len(t, resultProfiles, 2)
}
