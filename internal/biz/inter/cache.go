package inter

import (
	"context"
	"time"
)

type CacheRepo interface {
	Lock(ctx context.Context, lock string, value interface{}, sec time.Duration) (bool, error)
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value interface{}, sec time.Duration) (string, error)
}
