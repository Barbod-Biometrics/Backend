package usecase

import (
	"context"
)

type TokenService interface {
	GenerateTokens(ctx context.Context, userID uint64) (accessToken string, refreshToken string, err error)
}
