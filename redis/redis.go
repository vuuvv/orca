package redis

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/vuuvv/errors"
	"go.uber.org/zap"
)

func NewRedisClient(config *Config) (client *redis.Client, err error) {
	client = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", config.Host, config.Port),
		Password: config.Passwd,
		DB:       config.DB,
	})
	_, err = client.Ping(context.Background()).Result()
	if err != nil {
		return nil, errors.WithStack(err)
	}
	zap.L().Info("redis客户端初始化成功", zap.String("host", client.Options().Addr))
	return
}

var client *redis.Client

func GetClient() *redis.Client {
	if client == nil {
		panic("Redis client is nil")
	}
	return client
}

func SetClient(c *redis.Client) {
	client = c
}
