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

func TestValidateOIDCConfiguration_Valid(t *testing.T) {
	// Setup valid OIDC configuration
	viper.Reset()
	viper.Set("oidc.issuer", "https://auth.example.com")
	viper.Set("oidc.client_id", "test-client")
	viper.Set("oidc.client_secret", "test-secret")
	viper.Set("oidc.scopes", []string{"openid", "profile", "email"})
	viper.Set("oidc.client_auth_method", "client_secret_basic")
	viper.Set("oidc.token_validation", "jwks")
	viper.Set("oidc.claim_mappings.username", "preferred_username")
	viper.Set("oidc.claim_mappings.email", "email")
	viper.Set("oidc.claim_mappings.groups", "groups")
	viper.Set("oidc.claim_mappings.full_name", "name")

	ctx := context.Background()
	err := ValidateOIDCConfiguration(ctx)

	assert.NoError(t, err)
}

func TestValidateOIDCConfiguration_MissingRequired(t *testing.T) {
	// Setup incomplete OIDC configuration
	viper.Reset()
	viper.Set("oidc.client_id", "test-client")
	// Missing issuer and client_secret

	ctx := context.Background()
	err := ValidateOIDCConfiguration(ctx)

	assert.Error(t, err)
	// The error message will mention the first missing required field
	assert.Contains(t, err.Error(), "oidc provider requires")
}

func TestValidateOIDCConfiguration_InvalidAuthMethod(t *testing.T) {
	// Setup OIDC configuration with invalid auth method
	viper.Reset()
	viper.Set("oidc.issuer", "https://auth.example.com")
	viper.Set("oidc.client_id", "test-client")
	viper.Set("oidc.client_secret", "test-secret")
	viper.Set("oidc.client_auth_method", "invalid_method")
	viper.Set("oidc.scopes", []string{"openid"})
	viper.Set("oidc.claim_mappings.username", "preferred_username")
	viper.Set("oidc.claim_mappings.email", "email")
	viper.Set("oidc.claim_mappings.groups", "groups")
	viper.Set("oidc.claim_mappings.full_name", "name")

	ctx := context.Background()
	err := ValidateOIDCConfiguration(ctx)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid oidc.client_auth_method")
}

func TestValidateOIDCConfiguration_InvalidTokenValidation(t *testing.T) {
	// Setup OIDC configuration with invalid token validation
	viper.Reset()
	viper.Set("oidc.issuer", "https://auth.example.com")
	viper.Set("oidc.client_id", "test-client")
	viper.Set("oidc.client_secret", "test-secret")
	viper.Set("oidc.token_validation", "invalid_validation")
	viper.Set("oidc.scopes", []string{"openid"})
	viper.Set("oidc.claim_mappings.username", "preferred_username")
	viper.Set("oidc.claim_mappings.email", "email")
	viper.Set("oidc.claim_mappings.groups", "groups")
	viper.Set("oidc.claim_mappings.full_name", "name")

	ctx := context.Background()
	err := ValidateOIDCConfiguration(ctx)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid oidc.token_validation")
}

func TestValidateProviderConfiguration_DefaultsToAuthelia(t *testing.T) {
	// Setup configuration without auth_provider
	viper.Reset()
	
	ctx := context.Background()
	err := ValidateProviderConfiguration(ctx)

	assert.NoError(t, err)
	assert.Equal(t, "authelia", viper.GetString("auth_provider"))
}

func TestValidateProviderConfiguration_InvalidProvider(t *testing.T) {
	// Setup configuration with invalid provider
	viper.Reset()
	viper.Set("auth_provider", "invalid_provider")
	
	ctx := context.Background()
	err := ValidateProviderConfiguration(ctx)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid auth_provider: invalid_provider")
}

func TestGetEffectiveOIDCConfig(t *testing.T) {
	// Setup OIDC configuration
	viper.Reset()
	viper.Set("oidc.issuer", "https://auth.example.com")
	viper.Set("oidc.client_id", "test-client")
	viper.Set("oidc.client_secret", "test-secret")
	viper.Set("oidc.scopes", []string{"openid", "profile", "email"})
	viper.Set("oidc.claim_mappings.username", "preferred_username")
	viper.Set("oidc.claim_mappings.email", "email")
	// Set custom headers using the proper Viper syntax
	viper.Set("oidc.custom_headers", map[string]string{"X-Custom": "custom-value"})

	config := GetEffectiveOIDCConfig()

	assert.Equal(t, "https://auth.example.com", config["issuer"])
	assert.Equal(t, "test-client", config["client_id"])
	assert.Equal(t, "test-secret", config["client_secret"])
	assert.Equal(t, []string{"openid", "profile", "email"}, config["scopes"])
	
	claimMappings := config["claim_mappings"].(map[string]string)
	assert.Equal(t, "preferred_username", claimMappings["username"])
	assert.Equal(t, "email", claimMappings["email"])
	
	customHeaders := config["custom_headers"].(map[string]string)
	assert.Equal(t, "custom-value", customHeaders["X-Custom"])
}

func TestGetEffectiveProviderConfig(t *testing.T) {
	// Test with OIDC provider
	viper.Reset()
	viper.Set("auth_provider", "oidc")
	viper.Set("oidc.issuer", "https://auth.example.com")
	viper.Set("oidc.client_id", "test-client")

	config := GetEffectiveProviderConfig()

	assert.Equal(t, "https://auth.example.com", config["issuer"])
	assert.Equal(t, "test-client", config["client_id"])
}

func TestGetEffectiveProviderConfig_DefaultsToAuthelia(t *testing.T) {
	// Test without auth_provider set
	viper.Reset()
	viper.Set("authelia.header_username", "Remote-User")

	config := GetEffectiveProviderConfig()

	assert.Equal(t, "Remote-User", config["header_username"])
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
	viper.Set("cache.type", "memory")
	viper.Set("log_level", "info")

	defer func() {
		viper.Set("elasticsearch_host", "")
		viper.Set("elasticsearch_username", "")
		viper.Set("elasticsearch_password", "")
		viper.Set("secret_key", "")
		viper.Set("cache.type", "")
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
	viper.Set("cache.type", "redis")
	viper.Set("cache.redis_host", "")
	viper.Set("log_level", "info")

	defer func() {
		viper.Set("elasticsearch_host", "")
		viper.Set("elasticsearch_username", "")
		viper.Set("elasticsearch_password", "")
		viper.Set("secret_key", "")
		viper.Set("cache.type", "")
		viper.Set("cache.redis_host", "")
		viper.Set("log_level", "info")
	}()

	err := ValidateConfiguration(ctx)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "redis cache requires cache.redis_host configuration")
}

func TestValidateConfiguration_RedisCacheWithRedisHost(t *testing.T) {
	ctx := context.Background()

	viper.Set("elasticsearch_host", "localhost:9200")
	viper.Set("elasticsearch_username", "user")
	viper.Set("elasticsearch_password", "pass")
	viper.Set("secret_key", generateValidSecretKey())
	viper.Set("cache.type", "redis")
	viper.Set("cache.redis_host", "localhost:6379")
	viper.Set("log_level", "info")

	defer func() {
		viper.Set("elasticsearch_host", "")
		viper.Set("elasticsearch_username", "")
		viper.Set("elasticsearch_password", "")
		viper.Set("secret_key", "")
		viper.Set("cache.type", "")
		viper.Set("cache.redis_host", "")
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
	viper.Set("cache.type", "memcached")
	viper.Set("log_level", "info")

	defer func() {
		viper.Set("elasticsearch_host", "")
		viper.Set("elasticsearch_username", "")
		viper.Set("elasticsearch_password", "")
		viper.Set("secret_key", "")
		viper.Set("cache.type", "")
		viper.Set("log_level", "info")
	}()

	err := ValidateConfiguration(ctx)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid cache type")
	assert.Contains(t, err.Error(), "memcached")
	assert.Contains(t, err.Error(), "memory, redis, file")
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
			viper.Set("cache.type", "memory")
			viper.Set("log_level", level)

			defer func() {
				viper.Set("elasticsearch_host", "")
				viper.Set("elasticsearch_username", "")
				viper.Set("elasticsearch_password", "")
				viper.Set("secret_key", "")
				viper.Set("cache.type", "")
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
	viper.Set("cache.type", "memory")
	viper.Set("log_level", "verbose")

	defer func() {
		viper.Set("elasticsearch_host", "")
		viper.Set("elasticsearch_username", "")
		viper.Set("elasticsearch_password", "")
		viper.Set("secret_key", "")
		viper.Set("cache.type", "")
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
	viper.Set("cache.type", "memory")
	viper.Set("log_level", "info")

	defer func() {
		viper.Set("elasticsearch_host", "")
		viper.Set("elasticsearch_username", "")
		viper.Set("elasticsearch_password", "")
		viper.Set("secret_key", "")
		viper.Set("cache.type", "")
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
	viper.Set("cache.type", "memory")
	viper.Set("log_level", "info")

	defer func() {
		viper.Set("elasticsearch_host", "")
		viper.Set("elasticsearch_username", "")
		viper.Set("elasticsearch_password", "")
		viper.Set("secret_key", "")
		viper.Set("cache.type", "")
		viper.Set("log_level", "info")
	}()

	err := ValidateConfiguration(ctx)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "required configuration missing")
}
