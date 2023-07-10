package libs

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/wasilak/elastauth/cache"
)

func TestHealthRoute(t *testing.T) {
	// Setup
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/health")

	response := "{\"status\":\"OK\"}\n"

	// Assertions
	if assert.NoError(t, HealthRoute(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, response, rec.Body.String())
	}
}

func TestConfigRoute(t *testing.T) {
	// Setup
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/config")

	viperDefaultRolesMock := []string{"your_default_kibana_role"}
	viper.Set("default_roles", viperDefaultRolesMock)

	viperMappingsMock := map[string][]string{
		"your_ad_group": {"your_kibana_role"},
	}
	viper.Set("group_mappings", viperMappingsMock)

	response := "{\"default_roles\":[\"your_default_kibana_role\"],\"group_mappings\":{\"your_ad_group\":[\"your_kibana_role\"]}}\n"

	// Assertions
	if assert.NoError(t, ConfigRoute(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, response, rec.Body.String())
	}
}

func TestMainRoute(t *testing.T) {
	// Set up test environment
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Set up test data
	viper.Set("headers_username", "username")
	viper.Set("headers_groups", "groups")
	viper.Set("headers_email", "email")
	viper.Set("headers_name", "name")
	viper.Set("secret_key", "0123456789ABCDEF")
	viper.Set("elasticsearch_dry_run", false)
	viper.Set("elasticsearch_host", "http://localhost:9200")
	viper.Set("elasticsearch_username", "username")
	viper.Set("elasticsearch_password", "password")
	viper.Set("extend_cache", true)
	viper.Set("cache_expire", "1h")

	// Mock cache instance
	mockCache := &MockCacheInstance{}
	cache.CacheInstance = mockCache

	// Mock encrypt and decrypt functions
	mockEncrypt := func(ctx context.Context, stringToEncrypt string, keyString string) (string, error) {
		return "encrypted", nil
	}
	mockDecrypt := func(ctx context.Context, encryptedString string, keyString string) (string, error) {
		return "decrypted", nil
	}
	Encrypt = mockEncrypt
	Decrypt = mockDecrypt

	// Invoke the MainRoute function
	err := MainRoute(c)
	if err != nil {
		t.Fatalf("MainRoute failed with error: %v", err)
	}

	// Check the response status code
	assert.Equal(t, http.StatusOK, rec.Code)

	// Check the Authorization header
	assert.Equal(t, "Basic "+basicAuth("username", "decrypted"), rec.Header().Get(echo.HeaderAuthorization))
}

// MockCacheInstance is a mock implementation of the CacheInstance interface
type MockCacheInstance struct {
	cacheData map[string]interface{}
}

func (m *MockCacheInstance) Get(ctx context.Context, key string) (value interface{}, exists bool) {
	value, exists = m.cacheData[key]
	return value, exists
}

func (m *MockCacheInstance) Set(ctx context.Context, key string, value interface{}) {
	m.cacheData[key] = value
}

func (m *MockCacheInstance) GetItemTTL(ctx context.Context, key string) (ttl time.Duration, exists bool) {
	return 1 * time.Second, false
}

func (m *MockCacheInstance) ExtendTTL(ctx context.Context, key string, value interface{}) {
}

func (m *MockCacheInstance) GetTTL(ctx context.Context) time.Duration {
	return 1 * time.Second
}

func (m *MockCacheInstance) Init(ctx context.Context, cacheDuration time.Duration) {}

func TestMainRoute_HeaderNotProvided(t *testing.T) {
	// Set up test environment
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Set up test data with missing username header
	viper.Set("headers_username", "username")
	viper.Set("headers_groups", "groups")

	// Invoke the MainRoute function
	err := MainRoute(c)
	if err != nil {
		t.Fatalf("MainRoute failed with error: %v", err)
	}

	// Check the response status code
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	// Check the error response
	expectedResponse := `{"message":"Header not provided: username","code":400}`
	assert.Equal(t, expectedResponse, strings.TrimSpace(rec.Body.String()))
}

// MockEncrypt is a mock implementation of the Encrypt function
func MockEncrypt(ctx context.Context, stringToEncrypt string, keyString string) (string, error) {
	return "", errors.New("mock encrypt error")
}

// MockDecrypt is a mock implementation of the Decrypt function
func MockDecrypt(ctx context.Context, encryptedString string, keyString string) (string, error) {
	return "", errors.New("mock decrypt error")
}

func TestMainRoute_EncryptError(t *testing.T) {
	// Set up test environment
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Set up test data
	viper.Set("headers_username", "username")
	viper.Set("headers_groups", "groups")
	viper.Set("secret_key", "0123456789ABCDEF")

	// Mock encrypt and decrypt functions
	Encrypt = MockEncrypt
	Decrypt = MockDecrypt

	// Invoke the MainRoute function
	err := MainRoute(c)
	if err != nil {
		t.Fatalf("MainRoute failed with error: %v", err)
	}

	// Check the response status code
	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	// Check the error response
	expectedResponse := `{"message":"mock encrypt error","code":500}`
	assert.Equal(t, expectedResponse, strings.TrimSpace(rec.Body.String()))
}
