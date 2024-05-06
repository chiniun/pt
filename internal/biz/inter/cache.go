package inter

import (
	"context"
	"time"
)

type CacheRepo interface {
	Lock(ctx context.Context, lock string, value interface{}, sec time.Duration) (bool, error)
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value interface{}, sec time.Duration) (string, error)
	Del(ctx context.Context, key []string) (int64, error)

	HGet(ctx context.Context, key, field string) (string, error)
	HGetBool(ctx context.Context, key, field string) (bool, error)
	HSet(ctx context.Context, key, field string, value interface{}) (int64, error)
	Expire(ctx context.Context, key string, dur time.Duration) bool
}
