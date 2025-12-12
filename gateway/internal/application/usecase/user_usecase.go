package usecase

import (
	"context"

	userdto "github.com/Barbod-Biometrics/Backend/gateway/internal/application/dto/user"
)

type UserUsecase interface {
	GetUserByID(ctx context.Context, userID uint64) (*userdto.UserInfoResponse, error)
	UpdateProfile(ctx context.Context, profileInfo userdto.UpdateProfileRequest) (*userdto.UserInfoResponse, error)
}
