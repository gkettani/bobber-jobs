package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/gkettani/bobber-the-swe/internal/logger"
	"github.com/redis/go-redis/v9"
)

type RedisConfig struct {
	Addr string `env:"REDIS_ADDR" envDefault:"localhost:6379"`
}

type RedisCache struct {
	client     *redis.Client
	defaultTTL time.Duration
	ctx        context.Context
}

func NewRedisCache() (*RedisCache, error) {
	config := getRedisConfig()
	client := redis.NewClient(&redis.Options{
		Addr: config.Addr,
	})

	// Test the connection
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return &RedisCache{
		client:     client,
		defaultTTL: 24 * time.Hour,
		ctx:        ctx,
	}, nil
}

func getRedisConfig() *RedisConfig {
	c := RedisConfig{}
	if err := env.Parse(&c); err != nil {
		fmt.Printf("%+v\n", err)
	}
	return &c
}

func (c *RedisCache) Get(key string) (string, bool) {
	val, err := c.client.Get(c.ctx, key).Result()
	if err == redis.Nil {
		return "", false
	} else if err != nil {
		logger.Error("redis get error", "error", err)
		return "", false
	}

	return val, true
}

func (c *RedisCache) Set(key string, value string) {
	err := c.client.Set(c.ctx, key, value, c.defaultTTL).Err()
	if err != nil {
		logger.Error("redis set error", "error", err)
	}
}

func (c *RedisCache) Exists(key string) bool {
	val, err := c.client.Exists(c.ctx, key).Result()
	if err != nil {
		logger.Error("redis exists error", "error", err)
		return false
	}
	return val > 0
}

func (c *RedisCache) Delete(key string) {
	err := c.client.Del(c.ctx, key).Err()
	if err != nil {
		logger.Error("redis delete error", "error", err)
	}
}

func (c *RedisCache) Close() error {
	return c.client.Close()
}
