package cache

import (
	"context"
	"time"

	"github.com/dgraph-io/ristretto"
)

type RistrettoCache struct {
	Cache *ristretto.Cache
	TTL   time.Duration
	CTX   context.Context
}

func (c *RistrettoCache) Init(cacheDuration time.Duration) {
	var cacheErr error
	c.Cache, cacheErr = ristretto.NewCache(&ristretto.Config{
		NumCounters: 1e7,     // number of keys to track frequency of (10M).
		MaxCost:     1 << 28, // maximum cost of cache (256mb).
		BufferItems: 64,      // number of keys per Get buffer.
	})

	if cacheErr != nil {
		panic(cacheErr)
	}

	c.TTL = cacheDuration

	c.CTX = context.Background()
}

func (c *RistrettoCache) Get(cacheKey string) (interface{}, bool) {
	return c.Cache.Get(cacheKey)
}

func (c *RistrettoCache) Set(cacheKey string, item interface{}) {
	c.Cache.SetWithTTL(cacheKey, item, 1, c.TTL)
}

func (c *RistrettoCache) GetTTL(cacheKey string) (time.Duration, bool) {
	return c.Cache.GetTTL(cacheKey)
}

func (c *RistrettoCache) ExtendTTL(cacheKey string, item interface{}) {
	// todo: check if get/del/set is needed or if set is enough
	// _, exists := c.Cache.Get(cacheKey)
	// if exists {
	// 	c.Cache.Del(cacheKey)
	// }
	c.Cache.SetWithTTL(cacheKey, item, 1, c.TTL)
}
