package cache

import (
	"log"
	"time"

	"github.com/spf13/viper"
)

type CacheInterface interface {
	Init(cacheDuration time.Duration)
	Get(cacheKey string) (interface{}, bool)
	Set(cacheKey string, item interface{})
	GetItemTTL(cacheKey string) (time.Duration, bool)
	GetTTL() time.Duration
	ExtendTTL(cacheKey string, item interface{})
}

var CacheInstance CacheInterface

func CacheInit() {
	cacheDuration, err := time.ParseDuration(viper.GetString("cache_expire"))
	if err != nil {
		log.Fatal(err)
	}

	if viper.GetString("cache_type") == "redis" {
		CacheInstance = &RedisCache{
			Address: viper.GetString("redis_host"),
			DB:      viper.GetInt("redis_db"),
			TTL:     cacheDuration,
		}
	} else if viper.GetString("cache_type") == "memory" {
		CacheInstance = &GoCache{
			TTL: cacheDuration,
		}
	} else {
		log.Fatal("No cache_type selected or cache type is invalid")
	}

	CacheInstance.Init(cacheDuration)
}
