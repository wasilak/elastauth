package cache

import "time"

type CacheInterface interface {
	Init(cacheDuration time.Duration)
	Get(cacheKey string) (interface{}, bool)
	Set(cacheKey string, item interface{})
	GetTTL(cacheKey string) (time.Duration, bool)
	ExtendTTL(cacheKey string, item interface{})
}
