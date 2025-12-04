package repository

import (
	"context"

	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/entity"
)

type UserRepository interface {
	CreateUser(ctx context.Context, user *entity.User) error
	GetByID(ctx context.Context, userID uint64) (*entity.User, error)
	GetByPhoneNumber(ctx context.Context, phoneNumber string) (*entity.User, error)
	GetByEmail(ctx context.Context, email string) (*entity.User, error)
	Update(ctx context.Context, user *entity.User) error
}
