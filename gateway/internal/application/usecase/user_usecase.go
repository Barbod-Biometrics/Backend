package usecase

import (
	"context"

	userdto "github.com/Barbod-Biometrics/Backend/gateway/internal/application/dto/user"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/entity"
)

type UserService interface {
	GetUserByID(ctx context.Context, userID uint64) (*entity.User, error)
	UpdateProfile(ctx context.Context, profileInfo userdto.UpdateProfileRequest) (*userdto.UserInfoResponse, error)
}
