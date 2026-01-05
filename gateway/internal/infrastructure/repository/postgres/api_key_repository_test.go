package postgres_test

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/entity"
	repoPostgres "github.com/Barbod-Biometrics/Backend/gateway/internal/infrastructure/repository/postgres"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/test/testcontainer"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func createDummyProfile(t *testing.T, db *gorm.DB) uint64 {
	profile := entity.Profile{
		UserID:             100, // Dummy user ID
		ProfileType:        "personal",
		ProfileName:        "Test Profile",
		VerificationStatus: "verified",
		IsActive:           true,
		CreatedAt:          time.Now(),
	}

	// We create it in the DB immediately
	err := db.Create(&profile).Error
	assert.NoError(t, err, "failed to create seed profile")

	return profile.ProfileID
}

func TestApiKeyRepository_Create(t *testing.T) {
	ctx := context.Background()
	pg, err := testcontainer.SetupPostgres(ctx)
	if err != nil {
		t.Fatalf("failed to setup postgres container: %v", err)
	}
	defer pg.Terminate(ctx)

	repo := repoPostgres.NewApiKeyRepository(pg.DB)

	// --- FIX: Create Parent First ---
	realProfileID := createDummyProfile(t, pg.DB)

	apiKey := &entity.APIKey{
		ProfileID: realProfileID, // <--- Use the Real ID
		KeyHash:   "hash123",
		KeyPrefix: "pref123",
		IsActive:  true,
		CreatedAt: time.Now(),
	}

	err = repo.Create(context.Background(), apiKey)

	assert.NoError(t, err)
	assert.NotZero(t, apiKey.KeyID)

	// Verify in DB
	var count int64
	pg.DB.Model(&entity.APIKey{}).Where("key_id = ?", apiKey.KeyID).Count(&count)
	assert.Equal(t, int64(1), count)
}

func TestApiKeyRepository_GetByPrefix(t *testing.T) {
	ctx := context.Background()
	pg, err := testcontainer.SetupPostgres(ctx)
	if err != nil {
		t.Fatalf("failed to setup postgres container: %v", err)
	}
	defer pg.Terminate(ctx)

	repo := repoPostgres.NewApiKeyRepository(pg.DB)

	// --- FIX: Create Parent First ---
	realProfileID := createDummyProfile(t, pg.DB)
	prefix := "abcdef12"

	apiKey := entity.APIKey{
		ProfileID: realProfileID, // <--- Use the Real ID
		KeyHash:   "somehash",
		KeyPrefix: prefix,
		IsActive:  true,
		CreatedAt: time.Now(),
	}
	err = pg.DB.Create(&apiKey).Error
	assert.NoError(t, err)

	result, err := repo.GetByPrefix(context.Background(), prefix)

	assert.NoError(t, err)
	if assert.NotNil(t, result) {
		assert.Equal(t, realProfileID, result.ProfileID)
		assert.Equal(t, prefix, result.KeyPrefix)
	}
}

func TestApiKeyRepository_GetActiveByProfileID(t *testing.T) {
	ctx := context.Background()
	pg, err := testcontainer.SetupPostgres(ctx)
	if err != nil {
		t.Fatalf("failed to setup postgres container: %v", err)
	}
	defer pg.Terminate(ctx)

	repo := repoPostgres.NewApiKeyRepository(pg.DB)

	// --- FIX: Create Parent First ---
	profileID := createDummyProfile(t, pg.DB)

	// seed active
	apiKey := entity.APIKey{
		ProfileID: profileID, // <--- Use the Real ID
		KeyHash:   "hash_act",
		KeyPrefix: "pref_act",
		IsActive:  true,
		CreatedAt: time.Now(),
	}
	err = pg.DB.Create(&apiKey).Error
	assert.NoError(t, err)

	// seed inactive
	apiKeyInactive := entity.APIKey{
		ProfileID: profileID,
		KeyHash:   "hash_ina",
		KeyPrefix: "pref_ina", // 8 chars
		IsActive:  false,
		CreatedAt: time.Now(),
	}
	err = pg.DB.Create(&apiKeyInactive).Error
	assert.NoError(t, err)

	result, err := repo.GetActiveByProfileID(context.Background(), profileID)

	assert.NoError(t, err)
	if assert.NotNil(t, result) {
		assert.Equal(t, apiKey.KeyID, result.KeyID)
	}
}

func TestApiKeyRepository_Revoke(t *testing.T) {
	ctx := context.Background()
	pg, err := testcontainer.SetupPostgres(ctx)
	if err != nil {
		t.Fatalf("failed to setup postgres container: %v", err)
	}
	defer pg.Terminate(ctx)

	repo := repoPostgres.NewApiKeyRepository(pg.DB)

	// --- FIX: Create Parent First ---
	realProfileID := createDummyProfile(t, pg.DB)

	apiKey := entity.APIKey{
		ProfileID: realProfileID, // <--- Use the Real ID
		KeyHash:   "revoke_h",
		KeyPrefix: "revoke_p",
		IsActive:  true,
		CreatedAt: time.Now(),
	}
	err = pg.DB.Create(&apiKey).Error
	assert.NoError(t, err)

	idStr := strconv.FormatUint(apiKey.KeyID, 10)
	err = repo.Revoke(context.Background(), idStr)
	assert.NoError(t, err)

	// Verify it is inactive
	var updatedKey entity.APIKey
	err = pg.DB.First(&updatedKey, apiKey.KeyID).Error
	assert.NoError(t, err)
	assert.False(t, updatedKey.IsActive)
}

func TestApiKeyRepository_Revoke_InvalidID(t *testing.T) {
	ctx := context.Background()
	pg, err := testcontainer.SetupPostgres(ctx)
	if err != nil {
		t.Fatalf("failed to setup postgres container: %v", err)
	}
	defer pg.Terminate(ctx)

	repo := repoPostgres.NewApiKeyRepository(pg.DB)

	err = repo.Revoke(context.Background(), "abc")
	assert.Error(t, err)
	// We can't easily check exact error message unless we know it's not changing,
	// but the original test checked "invalid key id format".
	// Assuming the implementation returns that error for invalid strings.
	assert.Equal(t, "invalid key id format", err.Error())
}
