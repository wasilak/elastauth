package cache

import (
	"time"

	gocache "github.com/patrickmn/go-cache"
)

type GoCache struct {
	Cache *gocache.Cache
	TTL   time.Duration
}

func (c *GoCache) Init(cacheDuration time.Duration) {
	c.TTL = cacheDuration
	c.Cache = gocache.New(cacheDuration, c.TTL)
}

func (c *GoCache) GetTTL() time.Duration {
	return c.TTL
}

func (c *GoCache) Get(cacheKey string) (interface{}, bool) {
	return c.Cache.Get(cacheKey)
}

func (c *GoCache) Set(cacheKey string, item interface{}) {
	c.Cache.Set(cacheKey, item, c.TTL)
}

func (c *GoCache) GetItemTTL(cacheKey string) (time.Duration, bool) {
	_, expiration, found := c.Cache.GetWithExpiration(cacheKey)

	now := time.Now()
	difference := expiration.Sub(now)

	return difference, found
}

func (c *GoCache) ExtendTTL(cacheKey string, item interface{}) {
	c.Set(cacheKey, item)
}
