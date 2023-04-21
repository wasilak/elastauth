package cache

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/wasilak/elastauth/logger"
	"golang.org/x/exp/slog"
)

type RedisCache struct {
	Cache   *redis.Client
	TTL     time.Duration
	CTX     context.Context
	Address string
	DB      int
}

func (c *RedisCache) Init(cacheDuration time.Duration) {
	c.Cache = redis.NewClient(&redis.Options{
		Addr: c.Address,
		DB:   c.DB,
	})

	c.CTX = context.Background()

	c.TTL = cacheDuration
}

func (c *RedisCache) GetTTL() time.Duration {
	return c.TTL
}

func (c *RedisCache) Get(cacheKey string) (interface{}, bool) {
	item, err := c.Cache.Get(c.CTX, cacheKey).Result()

	if err != nil || len(item) == 0 {
		logger.LoggerInstance.Error("Error", slog.Any("message", err))
		return item, false
	}

	return item, true
}

func (c *RedisCache) Set(cacheKey string, item interface{}) {
	c.Cache.Set(c.CTX, cacheKey, item, c.TTL).Err()
}

func (c *RedisCache) GetItemTTL(cacheKey string) (time.Duration, bool) {
	item, err := c.Cache.TTL(c.CTX, cacheKey).Result()

	if err != nil {
		logger.LoggerInstance.Error("Error", slog.Any("message", err))
		return item, false
	}

	return item, true
}

func (c *RedisCache) ExtendTTL(cacheKey string, item interface{}) {
	c.Cache.Expire(c.CTX, cacheKey, c.TTL)
}
