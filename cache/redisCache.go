package cache

import (
	"context"
	"time"

	"log/slog"

	"github.com/go-redis/redis/v8"
	"go.opentelemetry.io/otel/trace"
)

// The RedisCache type represents a Redis cache with a specified time-to-live, context, address, and
// database.
// @property Cache - Cache is a pointer to a Redis client instance that is used to interact with the
// Redis cache.
// @property TTL - TTL stands for "Time To Live" and refers to the amount of time that a cached item
// will remain in the cache before it is considered expired and needs to be refreshed or removed. In
// the context of the RedisCache struct, it represents the duration of time that cached items will be
// stored in
// @property CTX - CTX is a context.Context object that is used to manage the lifecycle of a RedisCache
// instance. It is used to control the cancellation of operations and to pass values between functions.
// It is a part of the standard library in Go and is used extensively in network programming.
// @property {string} Address - Address is a string property that represents the network address of the
// Redis server. It typically includes the hostname or IP address of the server and the port number on
// which Redis is listening. For example, "localhost:6379" or "redis.example.com:6379".
// @property {int} DB - DB stands for "database" and is an integer value that represents the specific
// database within the Redis instance that the RedisCache struct will be interacting with. Redis allows
// for multiple databases to be created within a single instance, each with its own set of keys and
// values. The DB property allows the RedisCache
type RedisCache struct {
	Cache   *redis.Client
	TTL     time.Duration
	Address string
	DB      int
	Tracer  trace.Tracer
}

// `func (c *RedisCache) Init(cacheDuration time.Duration)` is a method of the `RedisCache` struct that
// initializes a new Redis client instance and sets the cache duration (TTL) for the RedisCache
// instance. It takes a `time.Duration` parameter `cacheDuration` which represents the duration of time
// that cached items will be stored in the cache. The method creates a new Redis client instance using
// the `redis.NewClient` function and sets the `Cache` property of the `RedisCache` instance to the new
// client instance. It also sets the `CTX` property to a new `context.Background()` instance. Finally,
// it sets the `TTL` property of the `RedisCache` instance to the `cacheDuration` parameter.
func (c *RedisCache) Init(ctx context.Context, cacheDuration time.Duration) {
	_, span := c.Tracer.Start(ctx, "Init")
	defer span.End()

	c.Cache = redis.NewClient(&redis.Options{
		Addr: c.Address,
		DB:   c.DB,
	})

	c.TTL = cacheDuration
}

// `func (c *RedisCache) GetTTL() time.Duration {` is a method of the `RedisCache` struct that returns
// the `TTL` property of the `RedisCache` instance, which represents the duration of time that cached
// items will be stored in the cache before they are considered expired and need to be refreshed or
// removed. The method returns a `time.Duration` value.
func (c *RedisCache) GetTTL(ctx context.Context) time.Duration {
	_, span := c.Tracer.Start(ctx, "GetTTL")
	defer span.End()

	return c.TTL
}

// `func (c *RedisCache) Get(cacheKey string) (interface{}, bool)` is a method of the `RedisCache`
// struct that retrieves a cached item from the Redis cache using the specified `cacheKey`. It returns
// a tuple containing the cached item as an `interface{}` and a boolean value indicating whether the
// item was successfully retrieved from the cache or not. If the item is not found in the cache or an
// error occurs during retrieval, the method returns an empty `interface{}` and `false`.
func (c *RedisCache) Get(ctx context.Context, cacheKey string) (interface{}, bool) {
	_, span := c.Tracer.Start(ctx, "Get")
	defer span.End()

	item, err := c.Cache.Get(ctx, cacheKey).Result()

	if err != nil || len(item) == 0 {
		slog.ErrorContext(ctx, "Error", slog.Any("message", err))
		return item, false
	}

	return item, true
}

// `func (c *RedisCache) Set(cacheKey string, item interface{})` is a method of the `RedisCache` struct
// that sets a value in the Redis cache with the specified `cacheKey`. It takes two parameters:
// `cacheKey`, which is a string representing the key under which the value will be stored in the
// cache, and `item`, which is an interface{} representing the value to be stored. The method uses the
// `Set` function of the Redis client to set the value in the cache with the specified key and TTL
// (time-to-live) duration. If an error occurs during the set operation, it is logged using the
// `slog.Error` function.
func (c *RedisCache) Set(ctx context.Context, cacheKey string, item interface{}) {
	_, span := c.Tracer.Start(ctx, "Set")
	defer span.End()

	c.Cache.Set(ctx, cacheKey, item, c.TTL).Err()
}

// `func (c *RedisCache) GetItemTTL(cacheKey string) (time.Duration, bool)` is a method of the
// `RedisCache` struct that retrieves the time-to-live (TTL) duration of a cached item with the
// specified `cacheKey`. It returns a tuple containing the TTL duration as a `time.Duration` value and
// a boolean value indicating whether the TTL was successfully retrieved from the cache or not. If the
// TTL is not found in the cache or an error occurs during retrieval, the method returns a zero
// `time.Duration` value and `false`.
func (c *RedisCache) GetItemTTL(ctx context.Context, cacheKey string) (time.Duration, bool) {
	_, span := c.Tracer.Start(ctx, "GetItemTTL")
	defer span.End()

	item, err := c.Cache.TTL(ctx, cacheKey).Result()

	if err != nil {
		slog.ErrorContext(ctx, "Error", slog.Any("message", err))
		return item, false
	}

	return item, true
}

// `func (c *RedisCache) ExtendTTL(cacheKey string, item interface{})` is a method of the `RedisCache`
// struct that extends the time-to-live (TTL) duration of a cached item with the specified `cacheKey`.
// It uses the `Expire` function of the Redis client to set the TTL duration of the cached item to the
// value of the `TTL` property of the `RedisCache` instance. This method is useful for refreshing the
// TTL of a cached item to prevent it from expiring prematurely.
func (c *RedisCache) ExtendTTL(ctx context.Context, cacheKey string, item interface{}) {
	_, span := c.Tracer.Start(ctx, "ExtendTTL")
	defer span.End()

	c.Cache.Expire(ctx, cacheKey, c.TTL)
}
