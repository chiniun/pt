package redis

import (
	"context"

	v1 "pt/internal/conf"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/redis/go-redis/v9"
)

func NewWithPanic(cfg *v1.Redis, logger log.Logger) (client *redis.Client) {
	client, err := New(cfg, logger)
	if err != nil {
		panic(err)
	}
	return
}
func New(cfg *v1.Redis, logger log.Logger) (client *redis.Client, err error) {
	if cfg.Network == "" {
		cfg.Network = "tcp"
	}
	client = redis.NewClient(&redis.Options{
		Network:               cfg.Network,
		Addr:                  cfg.Addr,
		ClientName:            "",
		Protocol:              0,
		Username:              cfg.Username,
		Password:              cfg.Password,
		DB:                    int(cfg.Db),
		MaxRetries:            int(cfg.MaxRetries),
		MinRetryBackoff:       cfg.MinRetryBackoff.AsDuration(),
		MaxRetryBackoff:       cfg.MaxRetryBackoff.AsDuration(),
		DialTimeout:           cfg.DialTimeout.AsDuration(),
		ReadTimeout:           cfg.ReadTimeout.AsDuration(),
		WriteTimeout:          cfg.WriteTimeout.AsDuration(),
		ContextTimeoutEnabled: false,
		PoolFIFO:              cfg.PoolFifo,
		PoolSize:              int(cfg.PoolSize),
		PoolTimeout:           cfg.PoolTimeout.AsDuration(),
		MinIdleConns:          int(cfg.MinIdleConns),
		MaxIdleConns:          0,
		MaxActiveConns:        0,
		ConnMaxIdleTime:       0,
		ConnMaxLifetime:       0,
		Limiter:               nil,
		DisableIndentity:      false,
		IdentitySuffix:        "",
	})

	err = client.Ping(context.Background()).Err()
	if err != nil {
		return
	}
	log.NewHelper(logger).Infow("message", "Connected to Redis")
	return
}

func Close(ctx context.Context, client *redis.Client) (err error) {
	closeErrChan := make(chan error, 1)
	go func() {
		closeErrChan <- client.Close()
		return
	}()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case closeErr := <-closeErrChan:
			return closeErr
		}
	}
}
