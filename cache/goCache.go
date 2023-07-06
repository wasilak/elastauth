package cache

import (
	"context"
	"time"

	gocache "github.com/patrickmn/go-cache"
	"go.opentelemetry.io/otel/trace"
)

// The GoCache type represents a cache with a specified time-to-live duration.
// @property Cache - Cache is a property of type `*gocache.Cache` which is a pointer to an instance of
// the GoCache library's Cache struct. This property is used to store and manage cached data in memory.
// @property TTL - TTL stands for Time To Live and it is a duration that specifies the amount of time
// for which an item should be considered valid in the cache before it is evicted. After the TTL
// expires, the item is considered stale and will be removed from the cache on the next access or
// eviction.
type GoCache struct {
	Cache  *gocache.Cache
	TTL    time.Duration
	Tracer trace.Tracer
}

// `func (c *GoCache) Init(cacheDuration time.Duration)` is a method of the `GoCache` struct that
// initializes the cache with a specified time-to-live duration. It sets the `TTL` property of the
// `GoCache` instance to the `cacheDuration` parameter and creates a new instance of the
// `gocache.Cache` struct with the same `cacheDuration` and `TTL` properties. This method is called
// when creating a new `GoCache` instance to set up the cache for use.
func (c *GoCache) Init(ctx context.Context, cacheDuration time.Duration) {
	_, span := c.Tracer.Start(ctx, "Init")
	defer span.End()

	c.TTL = cacheDuration
	c.Cache = gocache.New(cacheDuration, c.TTL)
}

// `func (c *GoCache) GetTTL() time.Duration {` is a method of the `GoCache` struct that returns the
// time-to-live duration (`TTL`) of the cache instance. It retrieves the `TTL` property of the
// `GoCache` instance and returns it as a `time.Duration` value. This method can be used to check the
// current `TTL` value of the cache instance.
func (c *GoCache) GetTTL(ctx context.Context) time.Duration {
	_, span := c.Tracer.Start(ctx, "GetTTL")
	defer span.End()

	return c.TTL
}

// `func (c *GoCache) Get(cacheKey string) (interface{}, bool)` is a method of the `GoCache` struct
// that retrieves an item from the cache based on the specified `cacheKey`. It returns two values: the
// cached item (as an `interface{}`) and a boolean value indicating whether the item was found in the
// cache or not. If the item is found in the cache, the boolean value will be `true`, otherwise it will
// be `false`.
func (c *GoCache) Get(ctx context.Context, cacheKey string) (interface{}, bool) {
	_, span := c.Tracer.Start(ctx, "Get")
	defer span.End()

	return c.Cache.Get(cacheKey)
}

// `func (c *GoCache) Set(cacheKey string, item interface{})` is a method of the `GoCache` struct that
// sets a value in the cache with the specified `cacheKey`. The `item` parameter is the value to be
// cached and the `cacheKey` parameter is the key used to identify the cached item. The method sets the
// value in the cache with the specified `cacheKey` and a time-to-live duration (`TTL`) equal to the
// `TTL` property of the `GoCache` instance. This means that the cached item will be considered valid
// for the duration of the `TTL` and will be automatically evicted from the cache after the `TTL`
// expires.
func (c *GoCache) Set(ctx context.Context, cacheKey string, item interface{}) {
	_, span := c.Tracer.Start(ctx, "Set")
	defer span.End()

	c.Cache.Set(cacheKey, item, c.TTL)
}

// `func (c *GoCache) GetItemTTL(cacheKey string) (time.Duration, bool)` is a method of the `GoCache`
// struct that retrieves the time-to-live duration (`TTL`) of a cached item identified by the specified
// `cacheKey`. It returns two values: the time-to-live duration of the cached item (as a
// `time.Duration` value) and a boolean value indicating whether the item was found in the cache or
// not. If the item is found in the cache, the boolean value will be `true`, otherwise it will be
// `false`. This method can be used to check the remaining time-to-live of a cached item.
func (c *GoCache) GetItemTTL(ctx context.Context, cacheKey string) (time.Duration, bool) {
	_, span := c.Tracer.Start(ctx, "GetItemTTL")
	defer span.End()

	_, expiration, found := c.Cache.GetWithExpiration(cacheKey)

	now := time.Now()
	difference := expiration.Sub(now)

	return difference, found
}

// `func (c *GoCache) ExtendTTL(cacheKey string, item interface{})` is a method of the `GoCache` struct
// that extends the time-to-live duration (`TTL`) of a cached item identified by the specified
// `cacheKey`. It does this by calling the `Set` method of the `gocache.Cache` struct with the same
// `cacheKey` and `item` parameters, and with a time-to-live duration (`TTL`) equal to the `TTL`
// property of the `GoCache` instance. This means that the cached item will be considered valid for an
// additional duration of the `TTL` and will be automatically evicted from the cache after the extended
// `TTL` expires. This method can be used to refresh the time-to-live of a cached item to prevent it
// from being evicted from the cache prematurely.
func (c *GoCache) ExtendTTL(ctx context.Context, cacheKey string, item interface{}) {
	_, span := c.Tracer.Start(ctx, "ExtendTTL")
	defer span.End()

	c.Set(ctx, cacheKey, item)
}
