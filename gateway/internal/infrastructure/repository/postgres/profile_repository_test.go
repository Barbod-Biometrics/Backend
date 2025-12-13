package postgres

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setupMockDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock, *gorm.DB) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock database: %v", err)
	}

	gormDB, err := gorm.Open(postgres.New(postgres.Config{Conn: db, PreferSimpleProtocol: true}), &gorm.Config{})
	if err != nil {
		_ = db.Close()
		t.Fatalf("failed to open gorm DB: %v", err)
	}
	return db, mock, gormDB
}

func TestGetPersonalProfileByUserID_FoundAndNotFound(t *testing.T) {
	db, mock, gdb := setupMockDB(t)
	defer db.Close()

	repo := NewProfileRepository(gdb)

	// success case
	{
		userID := uint64(42)
		rows := sqlmock.NewRows([]string{"profile_id", "user_id", "profile_type", "profile_name", "balance", "verification_status", "is_active", "created_at"}).
			AddRow(100, userID, "personal", "john", 0, "verified", true, time.Now())

		// GORM will issue a SELECT for profiles with WHERE user_id = ? AND profile_type = ? LIMIT 1
		mock.ExpectQuery(`SELECT .* FROM "profiles" WHERE user_id = \$1 AND profile_type = \$2`).
			WithArgs(userID, "personal", sqlmock.AnyArg()).WillReturnRows(rows)

		// Preloads will query details tables; return empty rows for them (order may vary)
		mock.ExpectQuery(`SELECT .* FROM "(profile_person_details|profile_business_details)"`).WillReturnRows(sqlmock.NewRows([]string{"profile_id"}))
		mock.ExpectQuery(`SELECT .* FROM "(profile_person_details|profile_business_details)"`).WillReturnRows(sqlmock.NewRows([]string{"profile_id"}))

		p, err := repo.GetPersonalProfileByUserID(context.Background(), userID)
		assert.NoError(t, err)
		assert.NotNil(t, p)
		assert.Equal(t, uint64(100), p.ProfileID)
	}

	// not found
	{
		userID := uint64(99)
		mock.ExpectQuery(`SELECT .* FROM "profiles" WHERE .*user_id.*`).WithArgs(userID, "personal", sqlmock.AnyArg()).WillReturnError(gorm.ErrRecordNotFound)
		p, err := repo.GetPersonalProfileByUserID(context.Background(), userID)
		assert.NoError(t, err)
		assert.Nil(t, p)
	}

	// ensure expectations met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("there were unfulfilled expectations: %v", err)
	}
}

func TestGetByID_Success(t *testing.T) {
	db, mock, gdb := setupMockDB(t)
	defer db.Close()

	repo := NewProfileRepository(gdb)

	profileID := uint64(10)
	rows := sqlmock.NewRows([]string{"profile_id", "user_id", "profile_type", "profile_name", "balance", "verification_status", "is_active", "created_at"}).
		AddRow(profileID, 7, "personal", "alice", 10, "verified", true, time.Now())

	mock.ExpectQuery(`SELECT .* FROM "profiles" WHERE .*profile_id.* = \$1`).WithArgs(profileID, sqlmock.AnyArg()).WillReturnRows(rows)
	// preloads (order may vary)
	mock.ExpectQuery(`SELECT .* FROM "(profile_person_details|profile_business_details)"`).WillReturnRows(sqlmock.NewRows([]string{"profile_id"}))
	mock.ExpectQuery(`SELECT .* FROM "(profile_person_details|profile_business_details)"`).WillReturnRows(sqlmock.NewRows([]string{"profile_id"}))

	p, err := repo.GetByID(context.Background(), profileID)
	assert.NoError(t, err)
	assert.NotNil(t, p)
	assert.Equal(t, profileID, p.ProfileID)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unfulfilled expectations: %v", err)
	}
}

func TestGetByUserID_Success(t *testing.T) {
	db, mock, gdb := setupMockDB(t)
	defer db.Close()

	repo := NewProfileRepository(gdb)

	userID := uint64(7)
	rows := sqlmock.NewRows([]string{"profile_id", "user_id", "profile_type", "profile_name", "balance", "verification_status", "is_active", "created_at"}).
		AddRow(11, userID, "personal", "a", 0, "draft", true, time.Now()).
		AddRow(12, userID, "business", "b", 0, "draft", true, time.Now())

	mock.ExpectQuery(`SELECT .* FROM "profiles" WHERE user_id = \$1`).WithArgs(userID).WillReturnRows(rows)
	// preloads for found profiles (order may vary)
	mock.ExpectQuery(`SELECT .* FROM "(profile_person_details|profile_business_details)"`).WillReturnRows(sqlmock.NewRows([]string{"profile_id"}))
	mock.ExpectQuery(`SELECT .* FROM "(profile_person_details|profile_business_details)"`).WillReturnRows(sqlmock.NewRows([]string{"profile_id"}))

	profiles, err := repo.GetByUserID(context.Background(), userID)
	assert.NoError(t, err)
	assert.Len(t, profiles, 2)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unfulfilled expectations: %v", err)
	}
}
