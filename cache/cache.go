package cache

import (
	"context"
	"log"
	"time"

	"github.com/spf13/viper"
	"go.opentelemetry.io/otel"
)

// The CacheInterface defines methods for initializing, getting, setting, and extending the
// time-to-live (TTL) of cached items.
// @property Init - Init is a method that initializes the cache with a specified cache duration. It
// takes a time.Duration parameter that represents the duration for which the cache items should be
// stored.
// @property Get - Get is a method of the CacheInterface that takes a cacheKey string as input and
// returns an interface{} and a bool. The interface{} represents the cached item associated with the
// cacheKey, and the bool indicates whether the item was found in the cache or not.
// @property Set - Set is a method of the CacheInterface that allows you to store an item in the cache
// with a given cacheKey. The item can be of any type that implements the empty interface {}.
// @property GetItemTTL - GetItemTTL is a method of the CacheInterface that returns the remaining
// time-to-live (TTL) of a cached item identified by its cacheKey. It returns the TTL as a
// time.Duration value and a boolean indicating whether the item exists in the cache or not. The TTL
// represents the time
// @property GetTTL - GetTTL is a method of the CacheInterface that returns the default time-to-live
// (TTL) duration for cached items. This duration specifies how long an item should remain in the cache
// before it is considered stale and needs to be refreshed or removed.
// @property ExtendTTL - ExtendTTL is a method in the CacheInterface that allows you to extend the
// time-to-live (TTL) of a cached item. This means that you can update the expiration time of a cached
// item to keep it in the cache for a longer period of time. This can be useful if you
type CacheInterface interface {
	Init(ctx context.Context, cacheDuration time.Duration)
	Get(ctx context.Context, cacheKey string) (interface{}, bool)
	Set(ctx context.Context, cacheKey string, item interface{})
	GetItemTTL(ctx context.Context, cacheKey string) (time.Duration, bool)
	GetTTL(ctx context.Context) time.Duration
	ExtendTTL(ctx context.Context, cacheKey string, item interface{})
}

// `var CacheInstance CacheInterface` is declaring a variable named `CacheInstance` of type
// `CacheInterface`. This variable will be used to store an instance of a cache that implements the
// `CacheInterface` methods.
var CacheInstance CacheInterface

// The function initializes a cache instance based on the cache type specified in the configuration
// file.
func CacheInit(ctx context.Context) {
	// Create a tracer named "Cache" using OpenTelemetry.
	tracer := otel.Tracer("Cache")

	// Start a new span named "CacheInit" using the tracer and context.
	_, span := tracer.Start(ctx, "CacheInit")

	// Defer the end of the span to ensure it is ended when the function returns.
	defer span.End()

	// Parse the cache expiration duration from the Viper configuration.
	cacheDuration, err := time.ParseDuration(viper.GetString("cache_expire"))
	if err != nil {
		log.Fatal(err)
	}

	// Check if the cache type is set to "redis".
	if viper.GetString("cache_type") == "redis" {
		// If it is, create a new RedisCache instance with the specified redis_host, redis_db, cache duration, and tracer.
		CacheInstance = &RedisCache{
			Address: viper.GetString("redis_host"),
			DB:      viper.GetInt("redis_db"),
			TTL:     cacheDuration,
			Tracer:  otel.Tracer("RedisCache"),
		}
	} else if viper.GetString("cache_type") == "memory" {
		// If the cache type is set to "memory", create a new GoCache instance with the specified cache duration and tracer.
		CacheInstance = &GoCache{
			TTL:    cacheDuration,
			Tracer: otel.Tracer("GoCache"),
		}
	} else {
		// If no cache type is selected or the cache type is invalid, log a fatal error message.
		log.Fatal("No cache_type selected or cache type is invalid")
	}

	// Initialize the CacheInstance using the given context and cache duration.
	CacheInstance.Init(ctx, cacheDuration)

}
