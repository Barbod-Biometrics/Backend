package usecase

import (
	"context"
)

type TokenUsecase interface {
	GenerateTokens(ctx context.Context, userID uint64, isAdmin bool) (accessToken string, refreshToken string, err error)
}
