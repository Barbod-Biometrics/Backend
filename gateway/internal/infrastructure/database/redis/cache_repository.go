package redis

import (
	"context"
	"time"

	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/repository"
)

type CacheRepository struct {
	client *RedisClient
}

func NewCacheRepository(client *RedisClient) *CacheRepository {
	return &CacheRepository{
		client: client,
	}
}

var _ repository.CacheRepository = (*CacheRepository)(nil)

func (r *CacheRepository) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return r.client.Set(ctx, key, value, expiration)
}

func (r *CacheRepository) Get(ctx context.Context, key string) (string, error) {
	return r.client.Get(ctx, key)
}

func (r *CacheRepository) Delete(ctx context.Context, keys ...string) error {
	return r.client.Delete(ctx, keys...)
}

func (r *CacheRepository) Exists(ctx context.Context, key string) (bool, error) {
	return r.client.Exists(ctx, key)
}

func (r *CacheRepository) Increment(ctx context.Context, key string) (int64, error) {
	return r.client.Increment(ctx, key)
}

func (r *CacheRepository) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return r.client.Expire(ctx, key, expiration)
}

func (r *CacheRepository) GetTTL(ctx context.Context, key string) (time.Duration, error) {
	return r.client.GetTTL(ctx, key)
}
