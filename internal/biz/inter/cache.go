package inter

import (
	"context"
	"time"

	"github.com/bsm/redislock"
)

type CacheRepo interface {
	// NIO 拿不到锁立即返回错误 
	Lock(ctx context.Context, lock string, sec time.Duration) (bool, error)
	// BIO 等待获取锁,超时报错
	ObtainLock(ctx context.Context, lockKey string, sec time.Duration) (*redislock.Lock, error)
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value interface{}, sec time.Duration) (string, error)
	Del(ctx context.Context, key []string) (int64, error)

	HGet(ctx context.Context, key, field string) (string, error)
	HGetBool(ctx context.Context, key, field string) (bool, error)
	HSet(ctx context.Context, key, field string, value interface{}) (int64, error)
	Expire(ctx context.Context, key string, dur time.Duration) bool
}
