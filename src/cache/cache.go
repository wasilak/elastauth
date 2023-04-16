package cache

import "time"

type CacheInterface interface {
	Init(cacheDuration time.Duration)
	Get(cacheKey string) (interface{}, bool)
	Set(cacheKey string, item interface{})
	GetItemTTL(cacheKey string) (time.Duration, bool)
	GetTTL() time.Duration
	ExtendTTL(cacheKey string, item interface{})
}

var CacheInstance CacheInterface
