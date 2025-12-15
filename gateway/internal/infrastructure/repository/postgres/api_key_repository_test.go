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
)

func TestApiKeyRepository_Create(t *testing.T) {
	ctx := context.Background()
	pg, err := testcontainer.SetupPostgres(ctx)
	if err != nil {
		t.Fatalf("failed to setup postgres container: %v", err)
	}
	defer pg.Terminate(ctx)

	repo := repoPostgres.NewApiKeyRepository(pg.DB)

	apiKey := &entity.APIKey{
		ProfileID: 10,
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

	prefix := "abcdef12" // 8 chars
	// seed
	apiKey := entity.APIKey{
		ProfileID: 100,
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
		assert.Equal(t, uint64(100), result.ProfileID)
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

	profileID := uint64(55)

	// seed active
	apiKey := entity.APIKey{
		ProfileID: profileID,
		KeyHash:   "hash_act",
		KeyPrefix: "pref_act", // 8 chars
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

	// seed
	apiKey := entity.APIKey{
		ProfileID: 123,
		KeyHash:   "revoke_h",
		KeyPrefix: "revoke_p", // 8 chars
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
