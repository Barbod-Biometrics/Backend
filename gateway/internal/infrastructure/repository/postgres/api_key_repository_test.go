package postgres_test

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/entity"
	repoPostgres "github.com/Barbod-Biometrics/Backend/gateway/internal/infrastructure/repository/postgres"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setupMockDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)

	dialector := postgres.New(postgres.Config{
		Conn:       db,
		DriverName: "postgres",
	})

	gormDB, err := gorm.Open(dialector, &gorm.Config{})
	assert.NoError(t, err)

	return gormDB, mock
}

func TestApiKeyRepository_Create(t *testing.T) {
	db, mock := setupMockDB(t)
	repo := repoPostgres.NewApiKeyRepository(db)

	apiKey := &entity.APIKey{
		ProfileID: 10,
		KeyHash:   "hash123",
		KeyPrefix: "pref123",
		IsActive:  true,
		CreatedAt: time.Now(),
	}

	mock.ExpectBegin()

	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "api_keys"`)).
		WithArgs(apiKey.ProfileID, apiKey.KeyHash, apiKey.KeyPrefix, apiKey.IsActive, sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"key_id"}).AddRow(1))

	mock.ExpectCommit()

	err := repo.Create(context.Background(), apiKey)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestApiKeyRepository_GetByPrefix(t *testing.T) {
	db, mock := setupMockDB(t)
	repo := repoPostgres.NewApiKeyRepository(db)

	prefix := "abcdef123"

	rows := sqlmock.NewRows([]string{"key_id", "profile_id", "key_hash", "key_prefix", "is_active"}).
		AddRow(1, 100, "somehash", prefix, true)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "api_keys" WHERE key_prefix = $1 ORDER BY "api_keys"."key_id" LIMIT $2`)).WithArgs(prefix, 1).WillReturnRows(rows)

	result, err := repo.GetByPrefix(context.Background(), prefix)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, uint64(100), result.ProfileID)
}

func TestApiKeyRepository_GetActiveByProfileID(t *testing.T) {
	db, mock := setupMockDB(t)
	repo := repoPostgres.NewApiKeyRepository(db)

	profileID := uint64(55)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "api_keys" WHERE profile_id = $1 AND is_active = $2 ORDER BY "api_keys"."key_id" LIMIT $3`)).
		WithArgs(profileID, true, 1).
		WillReturnRows(sqlmock.NewRows([]string{"key_id"}).AddRow(99))

	result, err := repo.GetActiveByProfileID(context.Background(), profileID)

	assert.NoError(t, err)
	assert.Equal(t, uint64(99), result.KeyID)
}

func TestApiKeyRepository_Revoke(t *testing.T) {
	db, mock := setupMockDB(t)
	repo := repoPostgres.NewApiKeyRepository(db)
	keyIDStr := "123"

	mock.ExpectBegin()

	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "api_keys" SET "is_active"=$1 WHERE key_id = $2`)).
		WithArgs(false, 123).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.Revoke(context.Background(), keyIDStr)
	assert.NoError(t, err)
}

func TestApiKeyRepository_Revoke_InvalidID(t *testing.T) {
	db, _ := setupMockDB(t)
	repo := repoPostgres.NewApiKeyRepository(db)

	err := repo.Revoke(context.Background(), "abc")
	assert.Error(t, err)
	assert.Equal(t, "invalid key id format", err.Error())
}
