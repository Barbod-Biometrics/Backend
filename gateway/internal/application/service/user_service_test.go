package service

import (
	"context"
	"errors"
	"testing"

	userdto "github.com/Barbod-Biometrics/Backend/gateway/internal/application/dto/user"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/entity"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/infrastructure/test/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetUserByID_Success_NoEmail(t *testing.T) {
	ctx := context.Background()
	repo := mocks.NewMockUserRepository(t)
	svc := NewUserService(repo)

	user := &entity.User{UserID: 1, PhoneNumber: "09120000000", Email: nil}
	repo.On("GetByID", mock.Anything, uint64(1)).Return(user, nil)

	resp, err := svc.GetUserByID(ctx, 1)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "09120000000", resp.PhoneNumber)
	assert.Equal(t, "", resp.Email)

	repo.AssertExpectations(t)
}

func TestGetUserByID_Success_WithEmail(t *testing.T) {
	ctx := context.Background()
	repo := mocks.NewMockUserRepository(t)
	svc := NewUserService(repo)

	email := "a@b.c"
	user := &entity.User{UserID: 2, PhoneNumber: "09121111111", Email: &email}
	repo.On("GetByID", mock.Anything, uint64(2)).Return(user, nil)

	resp, err := svc.GetUserByID(ctx, 2)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "09121111111", resp.PhoneNumber)
	assert.Equal(t, "a@b.c", resp.Email)

	repo.AssertExpectations(t)
}

func TestGetUserByID_NotFound(t *testing.T) {
	ctx := context.Background()
	repo := mocks.NewMockUserRepository(t)
	svc := NewUserService(repo)

	repo.On("GetByID", mock.Anything, uint64(3)).Return((*entity.User)(nil), nil)

	resp, err := svc.GetUserByID(ctx, 3)
	assert.ErrorIs(t, err, ErrUserNotFound)
	assert.Nil(t, resp)

	repo.AssertExpectations(t)
}

func TestGetUserByID_RepoError(t *testing.T) {
	ctx := context.Background()
	repo := mocks.NewMockUserRepository(t)
	svc := NewUserService(repo)

	repo.On("GetByID", mock.Anything, uint64(4)).Return((*entity.User)(nil), errors.New("db error"))

	resp, err := svc.GetUserByID(ctx, 4)
	assert.Error(t, err)
	assert.Nil(t, resp)

	repo.AssertExpectations(t)
}

func TestUpdateProfile_NoChange(t *testing.T) {
	ctx := context.Background()
	repo := mocks.NewMockUserRepository(t)
	svc := NewUserService(repo)

	email := "ok@ok.com"
	user := &entity.User{UserID: 10, PhoneNumber: "09121212121", Email: &email}
	repo.On("GetByID", mock.Anything, uint64(10)).Return(user, nil)

	req := userdto.UpdateProfileRequest{UserID: 10, PhoneNumber: "09121212121", Email: "ok@ok.com"}
	resp, err := svc.UpdateProfile(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "09121212121", resp.PhoneNumber)
	assert.Equal(t, "ok@ok.com", resp.Email)

	repo.AssertExpectations(t)
}

func TestUpdateProfile_PhoneConflict(t *testing.T) {
	ctx := context.Background()
	repo := mocks.NewMockUserRepository(t)
	svc := NewUserService(repo)

	user := &entity.User{UserID: 20, PhoneNumber: "09120000000", Email: nil}
	repo.On("GetByID", mock.Anything, uint64(20)).Return(user, nil)
	// Return a different user for the same phone
	existing := &entity.User{UserID: 21, PhoneNumber: "09123333333"}
	repo.On("GetByPhoneNumber", mock.Anything, "09123333333").Return(existing, nil)

	req := userdto.UpdateProfileRequest{UserID: 20, PhoneNumber: "09123333333"}
	resp, err := svc.UpdateProfile(ctx, req)
	assert.ErrorIs(t, err, ErrPhoneAlreadyExist)
	assert.Nil(t, resp)

	repo.AssertExpectations(t)
}

func TestUpdateProfile_EmailConflict(t *testing.T) {
	ctx := context.Background()
	repo := mocks.NewMockUserRepository(t)
	svc := NewUserService(repo)

	user := &entity.User{UserID: 30, PhoneNumber: "09120000000", Email: nil}
	repo.On("GetByID", mock.Anything, uint64(30)).Return(user, nil)
	// Return a different user for the same email
	existing := &entity.User{UserID: 31, PhoneNumber: "09123333333", Email: &[]string{"a@b.c"}[0]}
	repo.On("GetByEmail", mock.Anything, "a@b.c").Return(existing, nil)

	req := userdto.UpdateProfileRequest{UserID: 30, Email: "a@b.c"}
	resp, err := svc.UpdateProfile(ctx, req)
	assert.ErrorIs(t, err, ErrEmailAlreadyExist)
	assert.Nil(t, resp)

	repo.AssertExpectations(t)
}

func TestUpdateProfile_Success_Update(t *testing.T) {
	ctx := context.Background()
	repo := mocks.NewMockUserRepository(t)
	svc := NewUserService(repo)

	user := &entity.User{UserID: 40, PhoneNumber: "09120000000", Email: nil}
	repo.On("GetByID", mock.Anything, uint64(40)).Return(user, nil)
	// No conflicts
	repo.On("GetByPhoneNumber", mock.Anything, "09129999999").Return((*entity.User)(nil), nil)
	repo.On("GetByEmail", mock.Anything, "new@e.c").Return((*entity.User)(nil), nil)
	repo.On("Update", mock.Anything, mock.MatchedBy(func(u *entity.User) bool {
		return u.UserID == 40 && u.PhoneNumber == "09129999999" && u.Email != nil && *u.Email == "new@e.c"
	})).Return(nil)

	req := userdto.UpdateProfileRequest{UserID: 40, PhoneNumber: "09129999999", Email: "new@e.c"}
	resp, err := svc.UpdateProfile(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "09129999999", resp.PhoneNumber)
	assert.Equal(t, "new@e.c", resp.Email)

	repo.AssertExpectations(t)
}

func TestUpdateProfile_RepoUpdateError(t *testing.T) {
	ctx := context.Background()
	repo := mocks.NewMockUserRepository(t)
	svc := NewUserService(repo)

	user := &entity.User{UserID: 50, PhoneNumber: "09120000000", Email: nil}
	repo.On("GetByID", mock.Anything, uint64(50)).Return(user, nil)
	repo.On("GetByPhoneNumber", mock.Anything, "09129999999").Return((*entity.User)(nil), nil)
	repo.On("GetByEmail", mock.Anything, "new@e.c").Return((*entity.User)(nil), nil)
	repo.On("Update", mock.Anything, mock.Anything).Return(errors.New("update failed"))

	req := userdto.UpdateProfileRequest{UserID: 50, PhoneNumber: "09129999999", Email: "new@e.c"}
	resp, err := svc.UpdateProfile(ctx, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "update failed")

	repo.AssertExpectations(t)
}
