package usecase

import "context"

type APIKeyService interface {
	GenerateKey(ctx context.Context, profileID uint64) (string, error)
	RegenerateKey(ctx context.Context, profileID uint64) (string, error)
	Authenticate(ctx context.Context, rawKey string) (uint64, error)
}
