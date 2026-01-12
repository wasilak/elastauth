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
