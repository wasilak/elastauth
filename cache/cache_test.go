package cache

import (
	"context"
	"testing"
	"time"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestCachegoIntegration(t *testing.T) {
	ctx := context.Background()
	
	// Test memory cache
	t.Run("Memory Cache", func(t *testing.T) {
		viper.Set("cache.type", "memory")
		viper.Set("cache.expiration", "1h")
		
		// Initialize cache
		CacheInit(ctx)
		
		// Verify cache instance is created
		assert.NotNil(t, CacheInstance)
		
		// Test basic operations
		testKey := "test-key"
		testValue := "test-value"
		
		// Set and get
		CacheInstance.Set(ctx, testKey, testValue)
		value, exists := CacheInstance.Get(ctx, testKey)
		assert.True(t, exists)
		assert.Equal(t, testValue, value)
		
		// Test TTL
		ttl := CacheInstance.GetTTL(ctx)
		assert.Greater(t, ttl, time.Duration(0))
		
		// Clean up
		viper.Set("cache.type", "")
	})
	
	// Test no cache configuration
	t.Run("No Cache", func(t *testing.T) {
		viper.Set("cache.type", "")
		
		// Initialize cache
		CacheInit(ctx)
		
		// Verify no cache instance is created
		assert.Nil(t, CacheInstance)
	})
	
	// Test Redis cache configuration (without actually connecting)
	t.Run("Redis Cache Configuration", func(t *testing.T) {
		viper.Set("cache.type", "redis")
		viper.Set("cache.expiration", "30m")
		viper.Set("cache.redis_host", "localhost:6379")
		viper.Set("cache.redis_db", 1)
		
		// Get configuration
		config := GetEffectiveCacheConfig()
		
		assert.Equal(t, "redis", config["type"])
		assert.Equal(t, "30m", config["expiration"])
		assert.Equal(t, "localhost:6379", config["redis_host"])
		assert.Equal(t, 1, config["redis_db"])
		
		// Clean up
		viper.Set("cache.type", "")
		viper.Set("cache.expiration", "")
		viper.Set("cache.redis_host", "")
		viper.Set("cache.redis_db", 0)
	})
}

func TestGetEffectiveCacheConfig(t *testing.T) {
	// Test new format configuration
	t.Run("New Format Configuration", func(t *testing.T) {
		viper.Set("cache.type", "redis")
		viper.Set("cache.expiration", "2h")
		viper.Set("cache.redis_host", "redis.example.com:6379")
		viper.Set("cache.redis_db", 2)
		viper.Set("cache.path", "/custom/cache/path")
		
		config := GetEffectiveCacheConfig()
		
		assert.Equal(t, "redis", config["type"])
		assert.Equal(t, "2h", config["expiration"])
		assert.Equal(t, "redis.example.com:6379", config["redis_host"])
		assert.Equal(t, 2, config["redis_db"])
		assert.Equal(t, "/custom/cache/path", config["path"])
		
		// Clean up
		viper.Set("cache.type", "")
		viper.Set("cache.expiration", "")
		viper.Set("cache.redis_host", "")
		viper.Set("cache.redis_db", 0)
		viper.Set("cache.path", "")
	})
	
	// Test defaults
	t.Run("Defaults", func(t *testing.T) {
		viper.Set("cache.type", "")
		viper.Set("cache.expiration", "")
		viper.Set("cache.redis_host", "")
		viper.Set("cache.redis_db", 0)
		viper.Set("cache.path", "")
		
		config := GetEffectiveCacheConfig()
		
		assert.Equal(t, "", config["type"])
		assert.Equal(t, "1h", config["expiration"])
		assert.Equal(t, "localhost:6379", config["redis_host"])
		assert.Equal(t, 0, config["redis_db"])
		assert.Equal(t, "/tmp/elastauth-cache", config["path"])
	})
}