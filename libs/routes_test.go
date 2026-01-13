package libs

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

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

func TestReadinessRoute(t *testing.T) {
	ctx := context.Background()
	
	// Set up test configuration
	viper.Set("elasticsearch_host", "http://localhost:9200")
	viper.Set("elasticsearch_username", "test")
	viper.Set("elasticsearch_password", "test")
	viper.Set("elasticsearch.dry_run", true) // Use dry run to avoid actual connections
	viper.Set("cache.type", "memory")
	viper.Set("auth_provider", "authelia")
	viper.Set("secret_key", generateTestKey())
	
	// Set up Authelia provider configuration
	viper.Set("headers_username", "Remote-User")
	viper.Set("headers_groups", "Remote-Groups")
	viper.Set("headers_email", "Remote-Email")
	viper.Set("headers_name", "Remote-Name")

	// Initialize cache for the test
	cache.CacheInit(ctx)

	defer func() {
		viper.Set("elasticsearch_host", "")
		viper.Set("elasticsearch_username", "")
		viper.Set("elasticsearch_password", "")
		viper.Set("elasticsearch.dry_run", false)
		viper.Set("cache.type", "")
		viper.Set("auth_provider", "")
		viper.Set("secret_key", "")
		viper.Set("headers_username", "")
		viper.Set("headers_groups", "")
		viper.Set("headers_email", "")
		viper.Set("headers_name", "")
	}()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/ready", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if assert.NoError(t, ReadinessRoute(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		
		var response ReadinessResponse
		err := json.Unmarshal(rec.Body.Bytes(), &response)
		assert.NoError(t, err)
		
		assert.Equal(t, "OK", response.Status)
		assert.Contains(t, response.Checks, "elasticsearch")
		assert.Contains(t, response.Checks, "cache")
		assert.Contains(t, response.Checks, "provider")
		assert.NotEmpty(t, response.Timestamp)
		
		// All checks should be OK
		assert.Equal(t, "OK", response.Checks["elasticsearch"].Status)
		assert.Equal(t, "OK", response.Checks["cache"].Status)
		assert.Equal(t, "OK", response.Checks["provider"].Status)
	}
}

func TestLivenessRoute(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/live", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if assert.NoError(t, LivenessRoute(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		
		var response LivenessResponse
		err := json.Unmarshal(rec.Body.Bytes(), &response)
		assert.NoError(t, err)
		
		assert.Equal(t, "OK", response.Status)
		assert.NotEmpty(t, response.Timestamp)
		assert.NotEmpty(t, response.Uptime)
	}
}

func TestReadinessRoute_ElasticsearchFailure(t *testing.T) {
	// Set up test configuration with invalid Elasticsearch
	viper.Set("elasticsearch_host", "http://localhost:99999") // Invalid port
	viper.Set("elasticsearch_username", "test")
	viper.Set("elasticsearch_password", "test")
	viper.Set("elasticsearch_dry_run", false) // Don't use dry run to test actual failure
	viper.Set("cache.type", "memory")
	viper.Set("auth_provider", "authelia")

	defer func() {
		viper.Set("elasticsearch_host", "")
		viper.Set("elasticsearch_username", "")
		viper.Set("elasticsearch_password", "")
		viper.Set("elasticsearch_dry_run", false)
		viper.Set("cache.type", "")
		viper.Set("auth_provider", "")
	}()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/ready", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if assert.NoError(t, ReadinessRoute(c)) {
		assert.Equal(t, http.StatusServiceUnavailable, rec.Code)
		
		var response ReadinessResponse
		err := json.Unmarshal(rec.Body.Bytes(), &response)
		assert.NoError(t, err)
		
		assert.Equal(t, "NOT_READY", response.Status)
		assert.Equal(t, "ERROR", response.Checks["elasticsearch"].Status)
		assert.NotEmpty(t, response.Checks["elasticsearch"].Error)
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

	// Assertions
	if assert.NoError(t, ConfigRoute(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		
		// Parse the actual JSON response
		var actualJSON map[string]interface{}
		err := json.Unmarshal(rec.Body.Bytes(), &actualJSON)
		assert.NoError(t, err)
		
		// Check that all expected fields are present
		assert.Contains(t, actualJSON, "auth_provider")
		assert.Contains(t, actualJSON, "cache")
		assert.Contains(t, actualJSON, "default_roles")
		assert.Contains(t, actualJSON, "group_mappings")
		assert.Contains(t, actualJSON, "provider_config")
		
		// Check specific values that we set
		assert.Equal(t, "authelia", actualJSON["auth_provider"])
		
		// Check default_roles (convert from []interface{} to []string for comparison)
		defaultRoles, ok := actualJSON["default_roles"].([]interface{})
		assert.True(t, ok)
		assert.Len(t, defaultRoles, 1)
		assert.Equal(t, "your_default_kibana_role", defaultRoles[0])
		
		// Check group_mappings structure
		groupMappings, ok := actualJSON["group_mappings"].(map[string]interface{})
		assert.True(t, ok)
		assert.Contains(t, groupMappings, "your_ad_group")
		
		// Check that cache and provider_config are maps
		assert.IsType(t, map[string]interface{}{}, actualJSON["cache"])
		assert.IsType(t, map[string]interface{}{}, actualJSON["provider_config"])
	}
}

func generateTestKey() string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func setupTestCache(t *testing.T) {
	// Set up memory cache configuration
	viper.Set("cache.type", "memory")
	viper.Set("cache.expiration", "1h")
	
	// Initialize cache using new system
	ctx := context.Background()
	cache.CacheInit(ctx)
}

func TestMainRoute_ValidRequest_ValidationPasses(t *testing.T) {
	setupTestCache(t)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("Remote-User", "validuser")
	req.Header.Set("Remote-Groups", "admin,users")

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	viper.Set("headers_username", "Remote-User")
	viper.Set("headers_groups", "Remote-Groups")
	viper.Set("enable_group_whitelist", false)
	viper.Set("secret_key", generateTestKey())
	viper.Set("default_roles", []string{"user"})
	viper.Set("group_mappings", map[string][]string{})
	viper.Set("elasticsearch_dry_run", true)

	err := MainRoute(c)

	assert.NoError(t, err)
	assert.NotEqual(t, http.StatusBadRequest, rec.Code)
}

func TestMainRoute_InvalidUsername_ValidationFails(t *testing.T) {
	setupTestCache(t)

	e := echo.New()
	invalidUsername := "invalid username!@#$"
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("Remote-User", invalidUsername)
	req.Header.Set("Remote-Groups", "admin")

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	viper.Set("headers_username", "Remote-User")
	viper.Set("headers_groups", "Remote-Groups")
	viper.Set("headers_email", "Remote-Email")
	viper.Set("headers_name", "Remote-Name")
	viper.Set("enable_group_whitelist", false)
	viper.Set("secret_key", generateTestKey())

	err := MainRoute(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestMainRoute_InvalidGroup_ValidationFails(t *testing.T) {
	setupTestCache(t)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("Remote-User", "validuser")
	req.Header.Set("Remote-Groups", "admin,invalid\x00group")

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	viper.Set("headers_username", "Remote-User")
	viper.Set("headers_groups", "Remote-Groups")
	viper.Set("headers_email", "Remote-Email")
	viper.Set("headers_name", "Remote-Name")
	viper.Set("enable_group_whitelist", false)
	viper.Set("secret_key", generateTestKey())

	err := MainRoute(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestMainRoute_MissingUsername_BadRequest(t *testing.T) {
	setupTestCache(t)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", nil)

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	viper.Set("headers_username", "Remote-User")
	viper.Set("headers_groups", "Remote-Groups")
	viper.Set("headers_email", "Remote-Email")
	viper.Set("headers_name", "Remote-Name")

	err := MainRoute(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestMainRoute_GroupWhitelistEnforced(t *testing.T) {
	setupTestCache(t)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("Remote-User", "validuser")
	req.Header.Set("Remote-Groups", "unauthorized_group")

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	viper.Set("headers_username", "Remote-User")
	viper.Set("headers_groups", "Remote-Groups")
	viper.Set("headers_email", "Remote-Email")
	viper.Set("headers_name", "Remote-Name")
	viper.Set("enable_group_whitelist", true)
	viper.Set("group_whitelist", []string{"admin", "users"})
	viper.Set("secret_key", generateTestKey())

	err := MainRoute(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestMainRoute_CacheKeyProperlyEncoded(t *testing.T) {
	setupTestCache(t)

	e := echo.New()
	usernameWithSpecialChar := "user@domain.com"
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("Remote-User", usernameWithSpecialChar)
	req.Header.Set("Remote-Groups", "admin")

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	viper.Set("headers_username", "Remote-User")
	viper.Set("headers_groups", "Remote-Groups")
	viper.Set("enable_group_whitelist", false)
	viper.Set("secret_key", generateTestKey())
	viper.Set("default_roles", []string{"user"})
	viper.Set("group_mappings", map[string][]string{})
	viper.Set("elasticsearch_dry_run", true)

	err := MainRoute(c)

	assert.NoError(t, err)
	encodedKey := EncodeForCacheKey(usernameWithSpecialChar)
	assert.NotEqual(t, usernameWithSpecialChar, encodedKey)
	assert.Contains(t, encodedKey, "%40")
}
