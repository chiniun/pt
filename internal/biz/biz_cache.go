package biz

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

type CacheRepo interface {
	GetCache() redis.Client
	Lock(ctx context.Context, lockKey, lockString string, sec time.Duration) (bool, error)
	Get(ctx context.Context, pre, body string) (string, error)
	GetByKey(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key, value string, sec time.Duration) (string, error)
}
