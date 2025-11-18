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

	Increment(ctx context.Context, key string) (int64, error)
	Expire(ctx context.Context, key string, expiration time.Duration) error
	GetTTL(ctx context.Context, key string) (time.Duration, error)
}
