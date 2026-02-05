package redis

import (
	"context"
	"encoding/json"
	"time"

	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/entity"
	redisclient "github.com/Barbod-Biometrics/Backend/gateway/internal/infrastructure/database/redis"
)

type RedisSessionStore struct {
	client    *redisclient.RedisClient
	keyPrefix string
	// defaultTTL used when session ExpiresAt is not set or already passed
	defaultTTL time.Duration
}

func NewRedisSessionStore(client *redisclient.RedisClient, keyPrefix string, defaultTTL time.Duration) *RedisSessionStore {
	return &RedisSessionStore{client: client, keyPrefix: keyPrefix, defaultTTL: defaultTTL}
}

func (r *RedisSessionStore) key(id string) string {
	return r.keyPrefix + id
}

func (r *RedisSessionStore) Create(sess *entity.Session) error {
	ctx := context.Background()
	b, err := json.Marshal(sess)
	if err != nil {
		return err
	}
	exp := time.Until(sess.ExpiresAt)
	if exp <= 0 {
		exp = r.defaultTTL
	}
	return r.client.Set(ctx, r.key(sess.ID), b, exp)
}

func (r *RedisSessionStore) Get(id string) (*entity.Session, bool) {
	ctx := context.Background()
	s, err := r.client.Get(ctx, r.key(id))
	if err != nil {
		return nil, false
	}
	var sess entity.Session
	if err := json.Unmarshal([]byte(s), &sess); err != nil {
		return nil, false
	}
	return &sess, true
}

func (r *RedisSessionStore) Update(sess *entity.Session) error {
	return r.Create(sess)
}

func (r *RedisSessionStore) Delete(id string) error {
	ctx := context.Background()
	return r.client.Delete(ctx, r.key(id))
}

func (r *RedisSessionStore) Close() error {
	return nil
}
