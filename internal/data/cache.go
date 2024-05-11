package data

import (
	"context"
	"time"

	"github.com/bsm/redislock"
	kerr "github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
)

type Cache struct {
	data *Data
	log  *log.Helper
}

func NewCache(data *Data, logger log.Logger) *Cache {
	return &Cache{
		data: data,
		log:  log.NewHelper(logger),
	}
}

func (o *Cache) Lock(ctx context.Context, lock string, sec time.Duration) (bool, error) {

	setResult, err := o.data.redisCli.SetNX(ctx, lock, "1", sec*time.Millisecond).Result()
	if err != nil {
		return true, errors.Wrap(err, "Lock")
	}

	return setResult, nil
}

// ObtainLock
func (o *Cache) ObtainLock(ctx context.Context, lockKey string, sec time.Duration) (*redislock.Lock, error) {

	lock, err := o.data.locker.Obtain(ctx, lockKey, sec*time.Millisecond, nil)
	if err == redislock.ErrNotObtained {
		return nil, errors.Wrapf(err, "ObtainLockKey:%s", lockKey)
	}

	return lock, nil
}

func (o *Cache) Get(ctx context.Context, key string) (string, error) {
	result, err := o.data.redisCli.Get(ctx, key).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		log.Error("redis 获取key异常", "key : ", key, "error : ", err)
		return "", errors.Wrap(err, "Get")
	}
	return result, nil
}

func (o *Cache) Set(ctx context.Context, key string, value interface{}, sec time.Duration) (string, error) {

	result, err := o.data.redisCli.Set(ctx, key, value, sec).Result()
	if err != nil {
		return "", kerr.New(500, "Set: "+key, err.Error())
	}
	return result, nil

}

func (o *Cache) Del(ctx context.Context, keys []string) (int64, error) {

	count, err := o.data.redisCli.Del(ctx, keys...).Result()
	if err != nil {
		return 0, kerr.New(500, "del: ", err.Error())
	}
	return count, nil

}

func (o *Cache) HGet(ctx context.Context, key, field string) (string, error) {
	result, err := o.data.redisCli.HGet(ctx, key, field).Result()
	if err != nil {
		return "", errors.Wrap(err, "HGet")
	}
	return result, nil
}
func (o *Cache) HGetBool(ctx context.Context, key, field string) (bool, error) {
	result, err := o.data.redisCli.HGet(ctx, key, field).Bool()
	if err != nil {
		return false, errors.Wrap(err, "HGetBool")
	}
	return result, nil
}

func (o *Cache) Expire(ctx context.Context, key string, dur time.Duration) bool {
	result := o.data.redisCli.Expire(ctx, key, dur).Val()

	return result
}

func (o *Cache) HSet(ctx context.Context, key, field string, value interface{}) (int64, error) {
	result, err := o.data.redisCli.HSet(ctx, key, field, value).Result()
	if err != nil {
		return 0, errors.Wrap(err, "HSet")
	}
	return result, nil
}
