package database

import (
	"sync"

	"github.com/Barbod-Biometrics/Backend/gateway/internal/infrastructure/repository/redis"
)

type Cache interface {
	GetRDB() *redis.Client
}

type RedisDatabase struct {
	RDB *redis.Client
}

var (
	rdbOnce     sync.Once
	rdbInstance *RedisDatabase
)
