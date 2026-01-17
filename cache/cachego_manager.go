package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/wasilak/cachego"
	"github.com/wasilak/cachego/config"
	"go.opentelemetry.io/otel"
)

// CachegoManager wraps the cachego library to implement our CacheInterface
// while providing enhanced functionality and backward compatibility.
type CachegoManager struct {
	cache  cachego.CacheInterface
	config config.Config
	tracer interface{}
}

// NewCachegoManager creates a new cache manager using the cachego library.
func NewCachegoManager(cacheConfig map[string]interface{}) (*CachegoManager, error) {
	cacheType := ""
	if ct, ok := cacheConfig["type"].(string); ok {
		cacheType = ct
	}

	// If no cache type specified, return nil (no caching)
	if cacheType == "" {
		return nil, nil
	}

	// Build cachego configuration
	cfg := config.Config{
		Type: cacheType,
	}

	// Set expiration
	if exp, ok := cacheConfig["expiration"].(string); ok && exp != "" {
		cfg.Expiration = exp
	}

	// Set Redis-specific configuration
	if cacheType == "redis" {
		if host, ok := cacheConfig["redis_host"].(string); ok {
			cfg.RedisHost = host
		}
		if db, ok := cacheConfig["redis_db"].(int); ok {
			cfg.RedisDB = db
		}
	}

	// Set file cache path
	if cacheType == "file" {
		if path, ok := cacheConfig["path"].(string); ok {
			cfg.Path = path
		}
	}

	// Initialize cachego
	cache, err := cachego.CacheInit(context.Background(), cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize cachego: %w", err)
	}

	return &CachegoManager{
		cache:  cache,
		config: cfg,
		tracer: otel.Tracer("CachegoManager"),
	}, nil
}

// Init initializes the cache with the specified duration.
// For cachego, this is mostly a no-op since initialization happens in NewCachegoManager.
func (cm *CachegoManager) Init(ctx context.Context, cacheDuration time.Duration) {
	// Update the expiration if different from config
	durationStr := cacheDuration.String()
	if durationStr != cm.config.Expiration {
		cm.config.Expiration = durationStr
	}
}

// Get retrieves an item from the cache.
func (cm *CachegoManager) Get(ctx context.Context, cacheKey string) (interface{}, bool) {
	if cm.cache == nil {
		return nil, false
	}

	value, exists, err := cm.cache.Get(cacheKey)
	if err != nil {
		// Log error but don't fail - treat as cache miss
		return nil, false
	}

	if !exists {
		return nil, false
	}

	return string(value), true
}

// Set stores an item in the cache.
func (cm *CachegoManager) Set(ctx context.Context, cacheKey string, item interface{}) {
	if cm.cache == nil {
		return
	}

	// Convert item to byte slice
	var data []byte
	switch v := item.(type) {
	case string:
		data = []byte(v)
	case []byte:
		data = v
	default:
		data = []byte(fmt.Sprintf("%v", v))
	}

	// Ignore errors - cache failures shouldn't break the application
	_ = cm.cache.Set(cacheKey, data)
}

// GetItemTTL returns the remaining TTL for a cached item.
func (cm *CachegoManager) GetItemTTL(ctx context.Context, cacheKey string) (time.Duration, bool) {
	if cm.cache == nil {
		return 0, false
	}

	ttl, exists, err := cm.cache.GetItemTTL(cacheKey)
	if err != nil || !exists {
		return 0, false
	}

	return ttl, true
}

// GetTTL returns the default TTL for the cache.
func (cm *CachegoManager) GetTTL(ctx context.Context) time.Duration {
	duration, err := time.ParseDuration(cm.config.Expiration)
	if err != nil {
		return time.Hour // Default fallback
	}
	return duration
}

// ExtendTTL extends the TTL of a cached item.
func (cm *CachegoManager) ExtendTTL(ctx context.Context, cacheKey string, item interface{}) {
	if cm.cache == nil {
		return
	}

	// Convert item to byte slice
	var data []byte
	switch v := item.(type) {
	case string:
		data = []byte(v)
	case []byte:
		data = v
	default:
		data = []byte(fmt.Sprintf("%v", v))
	}

	// Ignore errors - cache failures shouldn't break the application
	_ = cm.cache.ExtendTTL(cacheKey, data)
}

// Type returns the cache type.
func (cm *CachegoManager) Type() string {
	return cm.config.Type
}

// IsEnabled returns whether caching is enabled.
func (cm *CachegoManager) IsEnabled() bool {
	return cm.cache != nil
}