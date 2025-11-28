package cache

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel"
)

func setupTestGoCache(duration time.Duration) *GoCache {
	return &GoCache{
		Tracer: otel.Tracer("test"),
		TTL:    duration,
	}
}

func TestGoCache_Init(t *testing.T) {
	ctx := context.Background()
	cache := setupTestGoCache(5 * time.Minute)
	duration := 5 * time.Minute

	cache.Init(ctx, duration)

	assert.NotNil(t, cache.Cache)
	assert.Equal(t, duration, cache.TTL)
}

func TestGoCache_GetTTL(t *testing.T) {
	ctx := context.Background()
	cache := setupTestGoCache(10 * time.Minute)
	duration := 10 * time.Minute

	cache.Init(ctx, duration)

	result := cache.GetTTL(ctx)

	assert.Equal(t, duration, result)
}

func TestGoCache_Set_And_Get(t *testing.T) {
	ctx := context.Background()
	cache := setupTestGoCache(100 * time.Millisecond)
	cache.Init(ctx, 5*time.Minute)

	cache.Set(ctx, "test_key", "test_value")

	value, found := cache.Get(ctx, "test_key")

	assert.True(t, found)
	assert.Equal(t, "test_value", value)
}

func TestGoCache_Get_NotFound(t *testing.T) {
	ctx := context.Background()
	cache := setupTestGoCache(100 * time.Millisecond)
	cache.Init(ctx, 5*time.Minute)

	value, found := cache.Get(ctx, "nonexistent_key")

	assert.False(t, found)
	assert.Nil(t, value)
}

func TestGoCache_GetItemTTL_AfterSet(t *testing.T) {
	ctx := context.Background()
	cache := setupTestGoCache(100 * time.Millisecond)
	cache.Init(ctx, 5*time.Minute)

	cache.Set(ctx, "test_key", "test_value")
	ttl, found := cache.GetItemTTL(ctx, "test_key")

	assert.True(t, found)
	assert.Greater(t, ttl, time.Duration(0))
	assert.LessOrEqual(t, ttl, 5*time.Minute)
}

func TestGoCache_GetItemTTL_NotFound(t *testing.T) {
	ctx := context.Background()
	cache := setupTestGoCache(5 * time.Minute)
	cache.Init(ctx, 5*time.Minute)

	ttl, found := cache.GetItemTTL(ctx, "nonexistent_key")

	assert.False(t, found)
	assert.LessOrEqual(t, ttl, time.Duration(0))
}

func TestGoCache_ExtendTTL(t *testing.T) {
	ctx := context.Background()
	cache := setupTestGoCache(5 * time.Minute)
	cache.Init(ctx, 5*time.Minute)

	cache.Set(ctx, "test_key", "test_value")

	cache.ExtendTTL(ctx, "test_key", "test_value")
	ttlAfter, _ := cache.GetItemTTL(ctx, "test_key")

	assert.Greater(t, ttlAfter, 4*time.Minute)
	assert.LessOrEqual(t, ttlAfter, 5*time.Minute)
}

func TestGoCache_Set_Multiple_Items(t *testing.T) {
	ctx := context.Background()
	cache := setupTestGoCache(100 * time.Millisecond)
	cache.Init(ctx, 5*time.Minute)

	cache.Set(ctx, "key1", "value1")
	cache.Set(ctx, "key2", "value2")
	cache.Set(ctx, "key3", "value3")

	value1, found1 := cache.Get(ctx, "key1")
	value2, found2 := cache.Get(ctx, "key2")
	value3, found3 := cache.Get(ctx, "key3")

	assert.True(t, found1)
	assert.True(t, found2)
	assert.True(t, found3)
	assert.Equal(t, "value1", value1)
	assert.Equal(t, "value2", value2)
	assert.Equal(t, "value3", value3)
}

func TestGoCache_Set_Override(t *testing.T) {
	ctx := context.Background()
	cache := setupTestGoCache(100 * time.Millisecond)
	cache.Init(ctx, 5*time.Minute)

	cache.Set(ctx, "test_key", "original_value")
	value1, _ := cache.Get(ctx, "test_key")

	cache.Set(ctx, "test_key", "new_value")
	value2, _ := cache.Get(ctx, "test_key")

	assert.Equal(t, "original_value", value1)
	assert.Equal(t, "new_value", value2)
}

func TestGoCache_Set_Different_Types(t *testing.T) {
	ctx := context.Background()
	cache := setupTestGoCache(100 * time.Millisecond)
	cache.Init(ctx, 5*time.Minute)

	cache.Set(ctx, "string_key", "string_value")
	cache.Set(ctx, "int_key", 42)
	cache.Set(ctx, "bool_key", true)
	cache.Set(ctx, "map_key", map[string]string{"name": "test"})

	strVal, _ := cache.Get(ctx, "string_key")
	intVal, _ := cache.Get(ctx, "int_key")
	boolVal, _ := cache.Get(ctx, "bool_key")
	mapVal, _ := cache.Get(ctx, "map_key")

	assert.Equal(t, "string_value", strVal)
	assert.Equal(t, 42, intVal)
	assert.Equal(t, true, boolVal)
	assert.Equal(t, map[string]string{"name": "test"}, mapVal)
}

func TestGoCache_Expiration(t *testing.T) {
	ctx := context.Background()
	cache := setupTestGoCache(100 * time.Millisecond)
	cache.Init(ctx, 100*time.Millisecond)

	cache.Set(ctx, "test_key", "test_value")
	value, found := cache.Get(ctx, "test_key")
	assert.True(t, found)
	assert.Equal(t, "test_value", value)

	time.Sleep(150 * time.Millisecond)

	value, found = cache.Get(ctx, "test_key")
	assert.False(t, found)
	assert.Nil(t, value)
}

func TestGoCache_ExtendTTL_Nonexistent_Key(t *testing.T) {
	ctx := context.Background()
	cache := setupTestGoCache(5 * time.Minute)
	cache.Init(ctx, 5*time.Minute)

	cache.ExtendTTL(ctx, "nonexistent_key", "some_value")

	value, found := cache.Get(ctx, "nonexistent_key")
	assert.True(t, found)
	assert.Equal(t, "some_value", value)
}

func TestGoCache_Empty_String_Value(t *testing.T) {
	ctx := context.Background()
	cache := setupTestGoCache(100 * time.Millisecond)
	cache.Init(ctx, 5*time.Minute)

	cache.Set(ctx, "empty_key", "")

	value, found := cache.Get(ctx, "empty_key")
	assert.True(t, found)
	assert.Equal(t, "", value)
}

func TestGoCache_Nil_Value(t *testing.T) {
	ctx := context.Background()
	cache := setupTestGoCache(100 * time.Millisecond)
	cache.Init(ctx, 5*time.Minute)

	cache.Set(ctx, "nil_key", nil)

	value, found := cache.Get(ctx, "nil_key")
	assert.True(t, found)
	assert.Nil(t, value)
}

func TestGoCache_Large_Value(t *testing.T) {
	ctx := context.Background()
	cache := setupTestGoCache(100 * time.Millisecond)
	cache.Init(ctx, 5*time.Minute)

	largeValue := make([]byte, 1000000)
	for i := range largeValue {
		largeValue[i] = byte(i % 256)
	}

	cache.Set(ctx, "large_key", largeValue)

	value, found := cache.Get(ctx, "large_key")
	assert.True(t, found)
	assert.Equal(t, largeValue, value)
}

func TestGoCache_Special_Characters_In_Key(t *testing.T) {
	ctx := context.Background()
	cache := setupTestGoCache(100 * time.Millisecond)
	cache.Init(ctx, 5*time.Minute)

	specialKeys := []string{
		"key:with:colons",
		"key/with/slashes",
		"key with spaces",
		"key-with-dashes",
		"key_with_underscores",
		"key.with.dots",
		"key@with#special$chars",
	}

	for _, key := range specialKeys {
		cache.Set(ctx, key, "value_for_"+key)
	}

	for _, key := range specialKeys {
		value, found := cache.Get(ctx, key)
		assert.True(t, found, "key %s should be found", key)
		assert.Equal(t, "value_for_"+key, value)
	}
}

func TestGoCache_Complex_Struct(t *testing.T) {
	ctx := context.Background()
	cache := setupTestGoCache(100 * time.Millisecond)
	cache.Init(ctx, 5*time.Minute)

	type User struct {
		Name  string
		Age   int
		Email string
	}

	user := User{Name: "John", Age: 30, Email: "john@example.com"}

	cache.Set(ctx, "user_key", user)
	value, found := cache.Get(ctx, "user_key")

	assert.True(t, found)
	assert.Equal(t, user, value)
}

func TestGoCacheInterface_Implementation(t *testing.T) {
	ctx := context.Background()
	cache := setupTestGoCache(100 * time.Millisecond)
	cache.Init(ctx, 5*time.Minute)

	var cacheInterface CacheInterface = cache

	assert.NotNil(t, cacheInterface)

	cacheInterface.Set(ctx, "test_key", "test_value")
	value, found := cacheInterface.Get(ctx, "test_key")

	assert.True(t, found)
	assert.Equal(t, "test_value", value)
}
