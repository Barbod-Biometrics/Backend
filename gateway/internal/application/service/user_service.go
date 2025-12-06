package service

import (
	"context"
	"errors"
	"fmt"

	userdto "github.com/Barbod-Biometrics/Backend/gateway/internal/application/dto/user"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/application/usecase"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/entity"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/repository"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrEmailAlreadyExist = errors.New("email already exists")
	ErrPhoneAlreadyExist = errors.New("phone number already exists")
	ErrInvalidUpdateData = errors.New("no valid data to update")
)

type UserService struct {
	UserRepo repository.UserRepository
}

func NewUserService(userRepo repository.UserRepository) *UserService {
	return &UserService{
		UserRepo: userRepo,
	}
}

var _ usecase.UserService = (*UserService)(nil)

func (s *UserService) GetUserByID(ctx context.Context, userID uint64) (*entity.User, error) {
	user, err := s.UserRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if user == nil {
		return nil, ErrUserNotFound
	}

	return user, nil
}

func (s *UserService) UpdateProfile(ctx context.Context, req userdto.UpdateProfileRequest) (*userdto.UserInfoResponse, error) {
	user, err := s.UserRepo.GetByID(ctx, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed dto get user: %w", err)
	}

	if user == nil {
		return nil, ErrUserNotFound
	}

	hasChanged := false

	if req.PhoneNumber != "" && req.PhoneNumber != user.PhoneNumber {
		existingUser, err := s.UserRepo.GetByPhoneNumber(ctx, req.PhoneNumber)
		if err != nil {
			return nil, fmt.Errorf("failed to check phone number: %w", err)
		}

		if existingUser != nil && existingUser.UserID != user.UserID {
			return nil, ErrPhoneAlreadyExist
		}

		user.PhoneNumber = req.PhoneNumber
		hasChanged = true
	}

	if req.Email != "" && req.Email != user.Email {
		existingUser, err := s.UserRepo.GetByEmail(ctx, req.Email)
		if err != nil {
			return nil, fmt.Errorf("failed to check email: %w", err)
		}

		if existingUser != nil && existingUser.UserID != user.UserID {
			return nil, ErrEmailAlreadyExist
		}

		user.Email = req.Email
		hasChanged = true
	}

	if !hasChanged {
		return &userdto.UserInfoResponse{
			PhoneNumber: user.PhoneNumber,
			Email:       user.Email,
		}, nil
	}

	if err := s.UserRepo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return &userdto.UserInfoResponse{
		PhoneNumber: user.PhoneNumber,
		Email:       user.Email,
	}, nil
}
