package repository

import (
	"context"
	"time"
)

type CacheRepository interface {

	// Basic Operations
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Get(ctx context.Context, key string) (string, error)
	Delete(ctx context.Context, key ...string) error
	Exists(ctx context.Context, key string) (bool, error)
}
