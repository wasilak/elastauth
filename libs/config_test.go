package libs

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func generateValidSecretKey() string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func TestValidateSecretKey_Valid64CharHex(t *testing.T) {
	key := generateValidSecretKey()

	err := ValidateSecretKey(key)

	assert.NoError(t, err)
}

func TestValidateSecretKey_Empty(t *testing.T) {
	err := ValidateSecretKey("")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "secret_key is required")
	assert.Contains(t, err.Error(), "ELASTAUTH_SECRET_KEY")
}

func TestValidateSecretKey_InvalidHex(t *testing.T) {
	invalidKey := "not-valid-hex-string-zzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzz"

	err := ValidateSecretKey(invalidKey)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be hex-encoded")
}

func TestValidateSecretKey_TooShort(t *testing.T) {
	shortKey := hex.EncodeToString([]byte("short"))

	err := ValidateSecretKey(shortKey)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be 64 hex characters")
	assert.Contains(t, err.Error(), "32 bytes")
}

func TestValidateSecretKey_TooLong(t *testing.T) {
	bytes := make([]byte, 64)
	rand.Read(bytes)
	longKey := hex.EncodeToString(bytes)

	err := ValidateSecretKey(longKey)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be 64 hex characters")
}

func TestValidateSecretKey_ExactlyRightLength(t *testing.T) {
	key := generateValidSecretKey()

	err := ValidateSecretKey(key)

	assert.NoError(t, err)
}

func TestValidateSecretKey_PartiallyInvalidHex(t *testing.T) {
	validPart := hex.EncodeToString(make([]byte, 16))
	invalidPart := "zzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzz"
	invalidKey := validPart + invalidPart

	err := ValidateSecretKey(invalidKey)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be hex-encoded")
}

func TestValidateRequiredConfig_AllFieldsPresent(t *testing.T) {
	ctx := context.Background()

	viper.Set("elasticsearch_host", "localhost:9200")
	viper.Set("elasticsearch_username", "user")
	viper.Set("elasticsearch_password", "pass")
	viper.Set("secret_key", generateValidSecretKey())

	defer func() {
		viper.Set("elasticsearch_host", "")
		viper.Set("elasticsearch_username", "")
		viper.Set("elasticsearch_password", "")
		viper.Set("secret_key", "")
	}()

	err := ValidateRequiredConfig(ctx)

	assert.NoError(t, err)
}

func TestValidateRequiredConfig_MissingElasticsearchHost(t *testing.T) {
	ctx := context.Background()

	viper.Set("elasticsearch_host", "")
	viper.Set("elasticsearch_username", "user")
	viper.Set("elasticsearch_password", "pass")
	viper.Set("secret_key", generateValidSecretKey())

	defer func() {
		viper.Set("elasticsearch_host", "")
		viper.Set("elasticsearch_username", "")
		viper.Set("elasticsearch_password", "")
		viper.Set("secret_key", "")
	}()

	err := ValidateRequiredConfig(ctx)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "required configuration missing")
	assert.Contains(t, err.Error(), "elasticsearch_host")
	assert.Contains(t, err.Error(), "ELASTAUTH_ELASTICSEARCH_HOST")
}

func TestValidateRequiredConfig_MissingElasticsearchUsername(t *testing.T) {
	ctx := context.Background()

	viper.Set("elasticsearch_host", "localhost:9200")
	viper.Set("elasticsearch_username", "")
	viper.Set("elasticsearch_password", "pass")
	viper.Set("secret_key", generateValidSecretKey())

	defer func() {
		viper.Set("elasticsearch_host", "")
		viper.Set("elasticsearch_username", "")
		viper.Set("elasticsearch_password", "")
		viper.Set("secret_key", "")
	}()

	err := ValidateRequiredConfig(ctx)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "required configuration missing")
	assert.Contains(t, err.Error(), "elasticsearch_username")
}

func TestValidateRequiredConfig_MissingElasticsearchPassword(t *testing.T) {
	ctx := context.Background()

	viper.Set("elasticsearch_host", "localhost:9200")
	viper.Set("elasticsearch_username", "user")
	viper.Set("elasticsearch_password", "")
	viper.Set("secret_key", generateValidSecretKey())

	defer func() {
		viper.Set("elasticsearch_host", "")
		viper.Set("elasticsearch_username", "")
		viper.Set("elasticsearch_password", "")
		viper.Set("secret_key", "")
	}()

	err := ValidateRequiredConfig(ctx)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "required configuration missing")
	assert.Contains(t, err.Error(), "elasticsearch_password")
}

func TestValidateRequiredConfig_MissingSecretKey(t *testing.T) {
	ctx := context.Background()

	viper.Set("elasticsearch_host", "localhost:9200")
	viper.Set("elasticsearch_username", "user")
	viper.Set("elasticsearch_password", "pass")
	viper.Set("secret_key", "")

	defer func() {
		viper.Set("elasticsearch_host", "")
		viper.Set("elasticsearch_username", "")
		viper.Set("elasticsearch_password", "")
		viper.Set("secret_key", "")
	}()

	err := ValidateRequiredConfig(ctx)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "required configuration missing")
	assert.Contains(t, err.Error(), "secret_key")
}

func TestValidateRequiredConfig_MultipleFieldsMissing(t *testing.T) {
	ctx := context.Background()

	viper.Set("elasticsearch_host", "")
	viper.Set("elasticsearch_username", "")
	viper.Set("elasticsearch_password", "pass")
	viper.Set("secret_key", "")

	defer func() {
		viper.Set("elasticsearch_host", "")
		viper.Set("elasticsearch_username", "")
		viper.Set("elasticsearch_password", "")
		viper.Set("secret_key", "")
	}()

	err := ValidateRequiredConfig(ctx)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "required configuration missing")
	assert.Contains(t, err.Error(), "elasticsearch_host")
	assert.Contains(t, err.Error(), "elasticsearch_username")
	assert.Contains(t, err.Error(), "secret_key")
}

func TestValidateConfiguration_AllValid(t *testing.T) {
	ctx := context.Background()

	viper.Set("elasticsearch_host", "localhost:9200")
	viper.Set("elasticsearch_username", "user")
	viper.Set("elasticsearch_password", "pass")
	viper.Set("secret_key", generateValidSecretKey())
	viper.Set("cache_type", "memory")
	viper.Set("log_level", "info")

	defer func() {
		viper.Set("elasticsearch_host", "")
		viper.Set("elasticsearch_username", "")
		viper.Set("elasticsearch_password", "")
		viper.Set("secret_key", "")
		viper.Set("cache_type", "memory")
		viper.Set("log_level", "info")
	}()

	err := ValidateConfiguration(ctx)

	assert.NoError(t, err)
}

func TestValidateConfiguration_RedisCacheWithoutRedisHost(t *testing.T) {
	ctx := context.Background()

	viper.Set("elasticsearch_host", "localhost:9200")
	viper.Set("elasticsearch_username", "user")
	viper.Set("elasticsearch_password", "pass")
	viper.Set("secret_key", generateValidSecretKey())
	viper.Set("cache_type", "redis")
	viper.Set("redis_host", "")
	viper.Set("log_level", "info")

	defer func() {
		viper.Set("elasticsearch_host", "")
		viper.Set("elasticsearch_username", "")
		viper.Set("elasticsearch_password", "")
		viper.Set("secret_key", "")
		viper.Set("cache_type", "memory")
		viper.Set("redis_host", "")
		viper.Set("log_level", "info")
	}()

	err := ValidateConfiguration(ctx)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "redis_host is required")
	assert.Contains(t, err.Error(), "cache_type is 'redis'")
	assert.Contains(t, err.Error(), "ELASTAUTH_REDIS_HOST")
}

func TestValidateConfiguration_RedisCacheWithRedisHost(t *testing.T) {
	ctx := context.Background()

	viper.Set("elasticsearch_host", "localhost:9200")
	viper.Set("elasticsearch_username", "user")
	viper.Set("elasticsearch_password", "pass")
	viper.Set("secret_key", generateValidSecretKey())
	viper.Set("cache_type", "redis")
	viper.Set("redis_host", "localhost:6379")
	viper.Set("log_level", "info")

	defer func() {
		viper.Set("elasticsearch_host", "")
		viper.Set("elasticsearch_username", "")
		viper.Set("elasticsearch_password", "")
		viper.Set("secret_key", "")
		viper.Set("cache_type", "memory")
		viper.Set("redis_host", "")
		viper.Set("log_level", "info")
	}()

	err := ValidateConfiguration(ctx)

	assert.NoError(t, err)
}

func TestValidateConfiguration_InvalidCacheType(t *testing.T) {
	ctx := context.Background()

	viper.Set("elasticsearch_host", "localhost:9200")
	viper.Set("elasticsearch_username", "user")
	viper.Set("elasticsearch_password", "pass")
	viper.Set("secret_key", generateValidSecretKey())
	viper.Set("cache_type", "memcached")
	viper.Set("log_level", "info")

	defer func() {
		viper.Set("elasticsearch_host", "")
		viper.Set("elasticsearch_username", "")
		viper.Set("elasticsearch_password", "")
		viper.Set("secret_key", "")
		viper.Set("cache_type", "memory")
		viper.Set("log_level", "info")
	}()

	err := ValidateConfiguration(ctx)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid cache_type")
	assert.Contains(t, err.Error(), "memcached")
	assert.Contains(t, err.Error(), "memory")
	assert.Contains(t, err.Error(), "redis")
}

func TestValidateConfiguration_AllValidLogLevels(t *testing.T) {
	ctx := context.Background()
	logLevels := []string{"debug", "info", "warn", "error"}

	for _, level := range logLevels {
		t.Run("log_level_"+level, func(t *testing.T) {
			viper.Set("elasticsearch_host", "localhost:9200")
			viper.Set("elasticsearch_username", "user")
			viper.Set("elasticsearch_password", "pass")
			viper.Set("secret_key", generateValidSecretKey())
			viper.Set("cache_type", "memory")
			viper.Set("log_level", level)

			defer func() {
				viper.Set("elasticsearch_host", "")
				viper.Set("elasticsearch_username", "")
				viper.Set("elasticsearch_password", "")
				viper.Set("secret_key", "")
				viper.Set("cache_type", "memory")
				viper.Set("log_level", "info")
			}()

			err := ValidateConfiguration(ctx)

			assert.NoError(t, err, "log_level %s should be valid", level)
		})
	}
}

func TestValidateConfiguration_InvalidLogLevel(t *testing.T) {
	ctx := context.Background()

	viper.Set("elasticsearch_host", "localhost:9200")
	viper.Set("elasticsearch_username", "user")
	viper.Set("elasticsearch_password", "pass")
	viper.Set("secret_key", generateValidSecretKey())
	viper.Set("cache_type", "memory")
	viper.Set("log_level", "verbose")

	defer func() {
		viper.Set("elasticsearch_host", "")
		viper.Set("elasticsearch_username", "")
		viper.Set("elasticsearch_password", "")
		viper.Set("secret_key", "")
		viper.Set("cache_type", "memory")
		viper.Set("log_level", "info")
	}()

	err := ValidateConfiguration(ctx)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid log_level")
	assert.Contains(t, err.Error(), "verbose")
	assert.Contains(t, err.Error(), "debug")
	assert.Contains(t, err.Error(), "info")
	assert.Contains(t, err.Error(), "warn")
	assert.Contains(t, err.Error(), "error")
}

func TestValidateConfiguration_InvalidSecretKey(t *testing.T) {
	ctx := context.Background()

	viper.Set("elasticsearch_host", "localhost:9200")
	viper.Set("elasticsearch_username", "user")
	viper.Set("elasticsearch_password", "pass")
	viper.Set("secret_key", "invalid-key")
	viper.Set("cache_type", "memory")
	viper.Set("log_level", "info")

	defer func() {
		viper.Set("elasticsearch_host", "")
		viper.Set("elasticsearch_username", "")
		viper.Set("elasticsearch_password", "")
		viper.Set("secret_key", "")
		viper.Set("cache_type", "memory")
		viper.Set("log_level", "info")
	}()

	err := ValidateConfiguration(ctx)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "secret_key")
}

func TestValidateConfiguration_MissingRequiredField(t *testing.T) {
	ctx := context.Background()

	viper.Set("elasticsearch_host", "")
	viper.Set("elasticsearch_username", "user")
	viper.Set("elasticsearch_password", "pass")
	viper.Set("secret_key", generateValidSecretKey())
	viper.Set("cache_type", "memory")
	viper.Set("log_level", "info")

	defer func() {
		viper.Set("elasticsearch_host", "")
		viper.Set("elasticsearch_username", "")
		viper.Set("elasticsearch_password", "")
		viper.Set("secret_key", "")
		viper.Set("cache_type", "memory")
		viper.Set("log_level", "info")
	}()

	err := ValidateConfiguration(ctx)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "required configuration missing")
}
