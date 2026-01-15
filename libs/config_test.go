package libs

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"os"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func generateValidSecretKey() string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func setupTestConfig() {
	viper.Reset()
	viper.Set("auth_provider", "authelia")
	viper.Set("secret_key", generateValidSecretKey())
	viper.Set("elasticsearch.hosts", []string{"http://localhost:9200"})
	viper.Set("elasticsearch.username", "elastic")
	viper.Set("elasticsearch.password", "changeme")
	viper.Set("log_level", "info")
	viper.Set("log_format", "text")
}

func TestLoadConfig_ValidConfiguration(t *testing.T) {
	setupTestConfig()

	cfg, err := LoadConfig()

	require.NoError(t, err)
	assert.Equal(t, "authelia", cfg.AuthProvider)
	assert.Equal(t, "elastic", cfg.Elasticsearch.Username)
	assert.Len(t, cfg.Elasticsearch.Hosts, 1)
}

func TestLoadConfig_MissingSecretKey(t *testing.T) {
	setupTestConfig()
	viper.Set("secret_key", "")

	_, err := LoadConfig()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "SecretKey")
}

func TestLoadConfig_InvalidSecretKeyLength(t *testing.T) {
	setupTestConfig()
	viper.Set("secret_key", "tooshort")

	_, err := LoadConfig()

	assert.Error(t, err)
}

func TestLoadConfig_InvalidAuthProvider(t *testing.T) {
	setupTestConfig()
	viper.Set("auth_provider", "invalid")

	_, err := LoadConfig()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "AuthProvider")
}

func TestLoadConfig_OIDCProvider_Valid(t *testing.T) {
	setupTestConfig()
	viper.Set("auth_provider", "oidc")
	viper.Set("oidc.issuer", "https://auth.example.com")
	viper.Set("oidc.client_id", "test-client")
	viper.Set("oidc.client_secret", "test-secret")
	viper.Set("oidc.scopes", []string{"openid", "profile", "email"})
	viper.Set("oidc.claim_mappings.username", "preferred_username")
	viper.Set("oidc.claim_mappings.email", "email")
	viper.Set("oidc.claim_mappings.groups", "groups")
	viper.Set("oidc.claim_mappings.full_name", "name")

	cfg, err := LoadConfig()

	require.NoError(t, err)
	assert.Equal(t, "oidc", cfg.AuthProvider)
	assert.Equal(t, "https://auth.example.com", cfg.OIDC.Issuer)
}

func TestLoadConfig_OIDCProvider_MissingIssuer(t *testing.T) {
	setupTestConfig()
	viper.Set("auth_provider", "oidc")
	viper.Set("oidc.client_id", "test-client")
	viper.Set("oidc.client_secret", "test-secret")
	// Missing issuer

	_, err := LoadConfig()

	assert.Error(t, err)
}

func TestLoadConfig_InvalidLogLevel(t *testing.T) {
	setupTestConfig()
	viper.Set("log_level", "invalid")

	_, err := LoadConfig()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "LogLevel")
}

func TestLoadConfig_CacheRedis_Valid(t *testing.T) {
	setupTestConfig()
	viper.Set("cache.type", "redis")
	viper.Set("cache.redis_host", "localhost:6379")
	viper.Set("cache.redis_db", 0)

	cfg, err := LoadConfig()

	require.NoError(t, err)
	assert.Equal(t, "redis", cfg.Cache.Type)
	assert.Equal(t, "localhost:6379", cfg.Cache.RedisHost)
}

func TestLoadConfig_CacheRedis_MissingHost(t *testing.T) {
	setupTestConfig()
	viper.Set("cache.type", "redis")
	viper.Set("cache.redis_host", "")

	_, err := LoadConfig()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "RedisHost")
}

func TestLoadConfig_CacheFile_Valid(t *testing.T) {
	setupTestConfig()
	viper.Set("cache.type", "file")
	viper.Set("cache.path", "/tmp/test-cache")

	cfg, err := LoadConfig()

	require.NoError(t, err)
	assert.Equal(t, "file", cfg.Cache.Type)
	assert.Equal(t, "/tmp/test-cache", cfg.Cache.Path)
}

func TestLoadConfig_ProxyEnabled_Valid(t *testing.T) {
	setupTestConfig()
	viper.Set("proxy.enabled", true)
	viper.Set("proxy.elasticsearch_url", "http://elasticsearch:9200")
	viper.Set("proxy.timeout", "30s")
	viper.Set("proxy.idle_conn_timeout", "90s")
	viper.Set("proxy.max_idle_conns", 100)

	cfg, err := LoadConfig()

	require.NoError(t, err)
	assert.True(t, cfg.Proxy.Enabled)
	assert.Equal(t, "http://elasticsearch:9200", cfg.Proxy.ElasticsearchURL)
}

func TestLoadConfig_ProxyEnabled_MissingURL(t *testing.T) {
	setupTestConfig()
	viper.Set("proxy.enabled", true)
	viper.Set("proxy.elasticsearch_url", "")

	_, err := LoadConfig()

	assert.Error(t, err)
}

func TestValidateConfiguration_Success(t *testing.T) {
	setupTestConfig()

	ctx := context.Background()
	err := ValidateConfiguration(ctx)

	assert.NoError(t, err)
}

func TestHandleSecretKey_GenerateKey(t *testing.T) {
	// This test would exit the process, so we skip it
	t.Skip("Skipping test that would exit process")
}

func TestHandleSecretKey_AutoGenerate(t *testing.T) {
	viper.Reset()
	viper.Set("secret_key", "")

	ctx := context.Background()
	err := HandleSecretKey(ctx)

	assert.NoError(t, err)
	assert.NotEmpty(t, viper.GetString("secret_key"))
	assert.Len(t, viper.GetString("secret_key"), 64)
}

func TestGetElasticsearchHosts(t *testing.T) {
	setupTestConfig()

	hosts := GetElasticsearchHosts()

	assert.Len(t, hosts, 1)
	assert.Equal(t, "http://localhost:9200", hosts[0])
}

func TestGetEffectiveCacheConfig(t *testing.T) {
	setupTestConfig()
	viper.Set("cache.type", "redis")
	viper.Set("cache.redis_host", "localhost:6379")

	config := GetEffectiveCacheConfig()

	assert.Equal(t, "redis", config["type"])
	assert.Equal(t, "localhost:6379", config["redis_host"])
}

func TestGetEffectiveAutheliaConfig(t *testing.T) {
	setupTestConfig()

	config := GetEffectiveAutheliaConfig()

	// Just verify it returns a map without errors
	assert.NotNil(t, config)
}

func TestBuildProxyConfig_Disabled(t *testing.T) {
	setupTestConfig()
	viper.Set("proxy.enabled", false)

	cfg, err := BuildProxyConfig()

	assert.NoError(t, err)
	assert.Nil(t, cfg)
}

func TestBuildProxyConfig_Enabled(t *testing.T) {
	setupTestConfig()
	viper.Set("proxy.enabled", true)
	viper.Set("proxy.elasticsearch_url", "http://elasticsearch:9200")
	viper.Set("proxy.timeout", "30s")
	viper.Set("proxy.idle_conn_timeout", "90s")
	viper.Set("proxy.max_idle_conns", 100)

	cfg, err := BuildProxyConfig()

	require.NoError(t, err)
	require.NotNil(t, cfg)
	assert.True(t, cfg.Enabled)
	assert.Equal(t, "http://elasticsearch:9200", cfg.ElasticsearchURL)
}

// Test environment variable handling
func TestInitConfiguration_EnvVars(t *testing.T) {
	// Set environment variables
	os.Setenv("ELASTAUTH_AUTH_PROVIDER", "oidc")
	os.Setenv("ELASTAUTH_LOG_LEVEL", "debug")
	defer os.Unsetenv("ELASTAUTH_AUTH_PROVIDER")
	defer os.Unsetenv("ELASTAUTH_LOG_LEVEL")

	viper.Reset()
	err := InitConfiguration()
	require.NoError(t, err)

	assert.Equal(t, "oidc", viper.GetString("auth_provider"))
	assert.Equal(t, "debug", viper.GetString("log_level"))
}

// Test OIDC scopes from environment variable
func TestInitConfiguration_OIDCScopes(t *testing.T) {
	os.Setenv("ELASTAUTH_OIDC_SCOPES", "openid,profile,email,groups")
	defer os.Unsetenv("ELASTAUTH_OIDC_SCOPES")

	viper.Reset()
	err := InitConfiguration()
	require.NoError(t, err)

	scopes := viper.GetStringSlice("oidc.scopes")
	assert.Len(t, scopes, 4)
	assert.Contains(t, scopes, "openid")
	assert.Contains(t, scopes, "groups")
}
