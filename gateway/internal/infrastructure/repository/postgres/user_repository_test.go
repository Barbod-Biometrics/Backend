package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/entity"
	repoPostgres "github.com/Barbod-Biometrics/Backend/gateway/internal/infrastructure/repository/postgres"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/test/testcontainer"
	"github.com/stretchr/testify/assert"
)

func TestUserRepository_CreateUser(t *testing.T) {
	ctx := context.Background()
	pg, err := testcontainer.SetupPostgres(ctx)
	if err != nil {
		t.Fatalf("failed to setup postgres container: %v", err)
	}
	defer pg.Terminate(ctx)

	repo := repoPostgres.NewUserRepository(pg.DB)

	user := &entity.User{
		PhoneNumber: "09121234567",
		CreatedAt:   time.Now(),
	}

	err = repo.CreateUser(ctx, user)
	assert.NoError(t, err)
	assert.NotZero(t, user.UserID)

	// Verify persistence
	var savedUser entity.User
	err = pg.DB.First(&savedUser, user.UserID).Error
	assert.NoError(t, err)
	assert.Equal(t, user.PhoneNumber, savedUser.PhoneNumber)
}

func TestUserRepository_GetByPhoneNumber(t *testing.T) {
	ctx := context.Background()
	pg, err := testcontainer.SetupPostgres(ctx)
	if err != nil {
		t.Fatalf("failed to setup postgres container: %v", err)
	}
	defer pg.Terminate(ctx)

	repo := repoPostgres.NewUserRepository(pg.DB)

	phone := "09129876543"
	// Seed
	user := entity.User{
		PhoneNumber: phone,
		CreatedAt:   time.Now(),
	}
	err = pg.DB.Create(&user).Error
	assert.NoError(t, err)

	// Found
	{
		foundUser, err := repo.GetByPhoneNumber(ctx, phone)
		assert.NoError(t, err)
		assert.NotNil(t, foundUser)
		assert.Equal(t, user.UserID, foundUser.UserID)
	}

	// Not Found
	{
		foundUser, err := repo.GetByPhoneNumber(ctx, "09000000000")
		assert.NoError(t, err)
		assert.Nil(t, foundUser)
	}
}

func TestUserRepository_GetByEmail(t *testing.T) {
	ctx := context.Background()
	pg, err := testcontainer.SetupPostgres(ctx)
	if err != nil {
		t.Fatalf("failed to setup postgres container: %v", err)
	}
	defer pg.Terminate(ctx)

	repo := repoPostgres.NewUserRepository(pg.DB)

	email := "test@example.com"
	// Seed
	user := entity.User{
		PhoneNumber: "09121111111",
		Email:       &email,
		CreatedAt:   time.Now(),
	}
	err = pg.DB.Create(&user).Error
	assert.NoError(t, err)

	// Found
	{
		foundUser, err := repo.GetByEmail(ctx, email)
		assert.NoError(t, err)
		assert.NotNil(t, foundUser)
		assert.Equal(t, user.UserID, foundUser.UserID)
		assert.Equal(t, email, *foundUser.Email)
	}

	// Not Found
	{
		foundUser, err := repo.GetByEmail(ctx, "notfound@example.com")
		assert.NoError(t, err)
		assert.Nil(t, foundUser)
	}
}

func TestUserRepository_GetByID(t *testing.T) {
	ctx := context.Background()
	pg, err := testcontainer.SetupPostgres(ctx)
	if err != nil {
		t.Fatalf("failed to setup postgres container: %v", err)
	}
	defer pg.Terminate(ctx)

	repo := repoPostgres.NewUserRepository(pg.DB)

	// Seed
	user := entity.User{
		PhoneNumber: "09122222222",
		CreatedAt:   time.Now(),
	}
	err = pg.DB.Create(&user).Error
	assert.NoError(t, err)

	// Found
	{
		foundUser, err := repo.GetByID(ctx, user.UserID)
		assert.NoError(t, err)
		assert.NotNil(t, foundUser)
		assert.Equal(t, user.UserID, foundUser.UserID)
	}

	// Not Found
	{
		foundUser, err := repo.GetByID(ctx, 999999)
		assert.NoError(t, err)
		assert.Nil(t, foundUser)
	}
}

func TestUserRepository_Update(t *testing.T) {
	ctx := context.Background()
	pg, err := testcontainer.SetupPostgres(ctx)
	if err != nil {
		t.Fatalf("failed to setup postgres container: %v", err)
	}
	defer pg.Terminate(ctx)

	repo := repoPostgres.NewUserRepository(pg.DB)

	// Seed
	user := entity.User{
		PhoneNumber: "09123333333",
		IsAdmin:     false,
		CreatedAt:   time.Now(),
	}
	err = pg.DB.Create(&user).Error
	assert.NoError(t, err)

	// Update
	newEmail := "updated@example.com"
	user.Email = &newEmail
	user.IsAdmin = true

	err = repo.Update(ctx, &user)
	assert.NoError(t, err)

	// Verify
	var updatedUser entity.User
	err = pg.DB.First(&updatedUser, user.UserID).Error
	assert.NoError(t, err)
	assert.True(t, updatedUser.IsAdmin)
	assert.Equal(t, newEmail, *updatedUser.Email)
}
