package data

import (
	"context"
	"crypto/md5"

	"fmt"
	"time"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-redis/redis/v8"
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

func (o *Cache) GetCache(ctx context.Context) *redis.Client {
	return o.data.redisCli
}

func (o *Cache) Lock(ctx context.Context, lockKey, lockString string, sec time.Duration) (bool, error) {

	key := lockKey + fmt.Sprintf(":%x", md5.Sum([]byte(lockString)))
	setResult, err := o.data.redisCli.SetNX(ctx, key, time.Now(), sec*time.Second).Result()
	if err != nil {
		return true, errors.New(500, "LockErr", err.Error())
	}

	return setResult, nil
}

func (o *Cache) Get(ctx context.Context, pre, body string) (string, error) {
	key := pre + ":" + body
	result, err := o.data.redisCli.Get(ctx, key).Result()
	if err != nil {
		return "", errors.New(500, "Get: "+key, err.Error())
	}
	return result, nil

}

func (o *Cache) GetByKey(ctx context.Context, key string) (string, error) {
	result, err := o.data.redisCli.Get(ctx, key).Result()
	if err != nil {
		return "", errors.New(500, "Get: "+key, err.Error())
	}
	return result, nil

}

func (o *Cache) Set(ctx context.Context, key, value string,sec time.Duration) (string, error) {
	
	result, err := o.data.redisCli.Set(ctx, key,value, sec).Result()
	if err != nil {
		return "", errors.New(500, "Set: "+key, err.Error())
	}
	return result, nil

}

