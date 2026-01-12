package libs

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"os"
	"strings"
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

func TestValidateProviderConfiguration_ConflictingConfiguration(t *testing.T) {
	// Setup: Select OIDC but configure Casdoor
	viper.Reset()
	viper.Set("auth_provider", "oidc")
	viper.Set("casdoor.endpoint", "https://casdoor.example.com")
	viper.Set("casdoor.client_id", "test-client")
	
	ctx := context.Background()
	err := ValidateProviderConfiguration(ctx)
	
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "auth_provider is set to 'oidc' but explicit configuration found for: [casdoor]")
}

func TestValidateProviderConfiguration_MultipleProvidersConfigured(t *testing.T) {
	// Setup: Configure both OIDC and Casdoor explicitly
	viper.Reset()
	viper.Set("auth_provider", "oidc")
	viper.Set("oidc.issuer", "https://oidc.example.com")
	viper.Set("casdoor.endpoint", "https://casdoor.example.com")
	viper.Set("casdoor.client_id", "test-client")
	
	ctx := context.Background()
	err := ValidateProviderConfiguration(ctx)
	
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "explicit configuration found for: [casdoor]")
}

func TestValidateProviderConfiguration_ValidOIDCSelection(t *testing.T) {
	// Setup: Select OIDC and configure OIDC properly
	viper.Reset()
	viper.Set("auth_provider", "oidc")
	viper.Set("oidc.issuer", "https://oidc.example.com")
	viper.Set("oidc.client_id", "test-client")
	viper.Set("oidc.client_secret", "test-secret")
	viper.Set("oidc.scopes", []string{"openid", "profile", "email"})
	viper.Set("oidc.claim_mappings.username", "preferred_username")
	viper.Set("oidc.claim_mappings.email", "email")
	viper.Set("oidc.claim_mappings.groups", "groups")
	viper.Set("oidc.claim_mappings.full_name", "name")
	
	ctx := context.Background()
	err := ValidateProviderConfiguration(ctx)
	
	assert.NoError(t, err)
	assert.Equal(t, "oidc", viper.GetString("auth_provider"))
}

func TestValidateProviderConfiguration_ValidCasdoorSelection(t *testing.T) {
	// Setup: Select Casdoor and configure Casdoor properly
	viper.Reset()
	viper.Set("auth_provider", "casdoor")
	viper.Set("casdoor.endpoint", "https://casdoor.example.com")
	viper.Set("casdoor.client_id", "test-client")
	viper.Set("casdoor.client_secret", "test-secret")
	
	ctx := context.Background()
	err := ValidateProviderConfiguration(ctx)
	
	assert.NoError(t, err)
	assert.Equal(t, "casdoor", viper.GetString("auth_provider"))
}

func TestValidateProviderConfiguration_AutheliaWithDefaults(t *testing.T) {
	// Setup: Select Authelia with default configuration (should work)
	viper.Reset()
	viper.Set("auth_provider", "authelia")
	
	ctx := context.Background()
	err := ValidateProviderConfiguration(ctx)
	
	assert.NoError(t, err)
	assert.Equal(t, "authelia", viper.GetString("auth_provider"))
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

func TestEnvironmentVariableSupport(t *testing.T) {
	// Save original environment
	originalEnv := make(map[string]string)
	envVars := []string{
		"ELASTAUTH_AUTH_PROVIDER",
		"ELASTAUTH_OIDC_CLIENT_SECRET",
		"ELASTAUTH_OIDC_SCOPES",
		"ELASTAUTH_CACHE_TYPE",
		"ELASTAUTH_AUTHELIA_HEADER_USERNAME",
	}
	
	for _, envVar := range envVars {
		originalEnv[envVar] = os.Getenv(envVar)
	}
	
	// Clean up after test
	defer func() {
		for _, envVar := range envVars {
			if originalValue, exists := originalEnv[envVar]; exists && originalValue != "" {
				os.Setenv(envVar, originalValue)
			} else {
				os.Unsetenv(envVar)
			}
		}
		viper.Reset()
	}()
	
	// Set test environment variables
	os.Setenv("ELASTAUTH_AUTH_PROVIDER", "oidc")
	os.Setenv("ELASTAUTH_OIDC_CLIENT_SECRET", "test-secret-from-env")
	os.Setenv("ELASTAUTH_OIDC_SCOPES", "openid,profile,email,custom")
	os.Setenv("ELASTAUTH_CACHE_TYPE", "redis")
	os.Setenv("ELASTAUTH_AUTHELIA_HEADER_USERNAME", "X-Remote-User")
	
	// Reset viper and set up environment variable support
	viper.Reset()
	viper.SetEnvPrefix("elastauth")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	
	// Set defaults and bind environment variables
	setConfigurationDefaults()
	bindProviderEnvironmentVariables()
	
	// Test that environment variables override defaults
	if viper.GetString("auth_provider") != "oidc" {
		t.Errorf("Expected auth_provider to be 'oidc', got '%s'", viper.GetString("auth_provider"))
	}
	
	if viper.GetString("oidc.client_secret") != "test-secret-from-env" {
		t.Errorf("Expected oidc.client_secret to be 'test-secret-from-env', got '%s'", viper.GetString("oidc.client_secret"))
	}
	
	if viper.GetString("cache.type") != "redis" {
		t.Errorf("Expected cache.type to be 'redis', got '%s'", viper.GetString("cache.type"))
	}
	
	if viper.GetString("authelia.header_username") != "X-Remote-User" {
		t.Errorf("Expected authelia.header_username to be 'X-Remote-User', got '%s'", viper.GetString("authelia.header_username"))
	}
	
	// Test OIDC scopes array parsing
	scopes := viper.GetStringSlice("oidc.scopes")
	expectedScopes := []string{"openid", "profile", "email", "custom"}
	if len(scopes) != len(expectedScopes) {
		t.Errorf("Expected %d scopes, got %d", len(expectedScopes), len(scopes))
	}
	for i, expected := range expectedScopes {
		if i >= len(scopes) || scopes[i] != expected {
			t.Errorf("Expected scope[%d] to be '%s', got '%s'", i, expected, scopes[i])
		}
	}
}

func TestOIDCCustomHeadersEnvironmentVariables(t *testing.T) {
	// Save original environment
	originalEnv := make(map[string]string)
	customHeaderEnvVars := []string{
		"ELASTAUTH_OIDC_CUSTOM_HEADERS_X_CUSTOM_HEADER",
		"ELASTAUTH_OIDC_CUSTOM_HEADERS_AUTHORIZATION_EXTRA",
	}
	
	for _, envVar := range customHeaderEnvVars {
		originalEnv[envVar] = os.Getenv(envVar)
	}
	
	// Clean up after test
	defer func() {
		for _, envVar := range customHeaderEnvVars {
			if originalValue, exists := originalEnv[envVar]; exists && originalValue != "" {
				os.Setenv(envVar, originalValue)
			} else {
				os.Unsetenv(envVar)
			}
		}
		viper.Reset()
	}()
	
	// Set test custom header environment variables
	os.Setenv("ELASTAUTH_OIDC_CUSTOM_HEADERS_X_CUSTOM_HEADER", "custom-value")
	os.Setenv("ELASTAUTH_OIDC_CUSTOM_HEADERS_AUTHORIZATION_EXTRA", "Bearer extra-token")
	
	// Reset viper and set up environment variable support
	viper.Reset()
	viper.SetEnvPrefix("elastauth")
	viper.AutomaticEnv()
	
	// Set defaults and bind environment variables (including custom headers)
	setConfigurationDefaults()
	bindProviderEnvironmentVariables()
	
	// Test that custom headers are properly parsed
	customHeaders := viper.GetStringMapString("oidc.custom_headers")
	
	// The headers are stored with the converted names (underscores to hyphens)
	expectedHeaders := map[string]string{
		"X-CUSTOM-HEADER":     "custom-value",
		"AUTHORIZATION-EXTRA": "Bearer extra-token",
	}
	
	if len(customHeaders) != len(expectedHeaders) {
		t.Errorf("Expected %d custom headers, got %d", len(expectedHeaders), len(customHeaders))
	}
	
	for expectedKey, expectedValue := range expectedHeaders {
		if actualValue, exists := customHeaders[expectedKey]; !exists {
			t.Errorf("Expected custom header '%s' not found", expectedKey)
		} else if actualValue != expectedValue {
			t.Errorf("Expected custom header '%s' to be '%s', got '%s'", expectedKey, expectedValue, actualValue)
		}
	}
}

func TestConfigurationPrecedence(t *testing.T) {
	// Save original environment
	originalAuthProvider := os.Getenv("ELASTAUTH_AUTH_PROVIDER")
	
	// Clean up after test
	defer func() {
		if originalAuthProvider != "" {
			os.Setenv("ELASTAUTH_AUTH_PROVIDER", originalAuthProvider)
		} else {
			os.Unsetenv("ELASTAUTH_AUTH_PROVIDER")
		}
		viper.Reset()
	}()
	
	// Test 1: Default value (no config file, no env var)
	viper.Reset()
	setConfigurationDefaults()
	
	if viper.GetString("auth_provider") != "authelia" {
		t.Errorf("Expected default auth_provider to be 'authelia', got '%s'", viper.GetString("auth_provider"))
	}
	
	// Test 2: Environment variable overrides default
	os.Setenv("ELASTAUTH_AUTH_PROVIDER", "oidc")
	viper.Reset()
	viper.SetEnvPrefix("elastauth")
	viper.AutomaticEnv()
	
	// Set defaults first, then bind environment variables
	setConfigurationDefaults()
	bindProviderEnvironmentVariables()
	
	if viper.GetString("auth_provider") != "oidc" {
		t.Errorf("Expected environment variable to override default, got '%s'", viper.GetString("auth_provider"))
	}
}
