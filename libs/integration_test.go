package libs

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wasilak/elastauth/cache"
)

func generateTestKeyForIntegration() string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func setupTestCacheForIntegration(t *testing.T) {
	// Set up memory cache configuration
	viper.Set("cache.type", "memory")
	viper.Set("cache.expiration", "1h")
	
	// Initialize cache using new system
	ctx := context.Background()
	cache.CacheInit(ctx)
}

func setupElasticsearchMockServer(t *testing.T) *httptest.Server {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		assert.NotEmpty(t, auth, "Authorization header should be present")

		if r.Method == "GET" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			response := map[string]interface{}{
				"name": "elasticsearch",
			}
			json.NewEncoder(w).Encode(response)
		} else if r.Method == "POST" || r.Method == "PUT" {
			assert.Contains(t, r.URL.Path, "/_security/user/", "User endpoint should be called")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			response := map[string]interface{}{
				"acknowledged": true,
			}
			json.NewEncoder(w).Encode(response)
		}
	}))

	return server
}

func TestIntegration_CompleteAuthFlow_CacheMiss(t *testing.T) {
	setupTestCacheForIntegration(t)
	server := setupElasticsearchMockServer(t)
	defer server.Close()

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("Remote-User", "testuser")
	req.Header.Set("Remote-Groups", "admin,users")
	req.Header.Set("Remote-Email", "testuser@example.com")
	req.Header.Set("Remote-Name", "Test User")

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	testKey := generateTestKeyForIntegration()
	viper.Set("headers_username", "Remote-User")
	viper.Set("headers_groups", "Remote-Groups")
	viper.Set("headers_email", "Remote-Email")
	viper.Set("headers_name", "Remote-Name")
	viper.Set("enable_group_whitelist", false)
	viper.Set("secret_key", testKey)
	viper.Set("default_roles", []string{"user"})
	viper.Set("group_mappings", map[string][]string{
		"admin": {"kibana_admin"},
		"users": {"kibana_user"},
	})
	viper.Set("elasticsearch_host", server.URL)
	viper.Set("elasticsearch_username", "elastic")
	viper.Set("elasticsearch_password", "password")
	viper.Set("elasticsearch_dry_run", false)
	viper.Set("extend_cache", false)

	err := MainRoute(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	authHeader := rec.Header().Get("Authorization")
	assert.NotEmpty(t, authHeader)
	assert.Contains(t, authHeader, "Basic ")

	encodedCredentials := authHeader[6:]
	decodedCredentials, err := base64.StdEncoding.DecodeString(encodedCredentials)
	assert.NoError(t, err)
	assert.Contains(t, string(decodedCredentials), "testuser")
}

func TestIntegration_CompleteAuthFlow_CacheHit(t *testing.T) {
	setupTestCacheForIntegration(t)
	server := setupElasticsearchMockServer(t)
	defer server.Close()

	testKey := generateTestKeyForIntegration()
	ctx := context.Background()

	testPassword := "TestPassword123!@#"
	encryptedPassword, err := Encrypt(ctx, testPassword, testKey)
	require.NoError(t, err)

	encryptedPasswordBase64 := base64.URLEncoding.EncodeToString([]byte(encryptedPassword))
	cacheKey := "elastauth-testuser"

	// Set up memory cache configuration
	viper.Set("cache.type", "memory")
	viper.Set("cache.expiration", "1h")
	
	// Initialize cache using new system
	cache.CacheInit(ctx)
	cache.CacheInstance.Set(ctx, cacheKey, encryptedPasswordBase64)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("Remote-User", "testuser")
	req.Header.Set("Remote-Groups", "admin,users")

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	viper.Set("headers_username", "Remote-User")
	viper.Set("headers_groups", "Remote-Groups")
	viper.Set("enable_group_whitelist", false)
	viper.Set("secret_key", testKey)
	viper.Set("default_roles", []string{"user"})
	viper.Set("group_mappings", map[string][]string{})
	viper.Set("elasticsearch_dry_run", true)
	viper.Set("extend_cache", false)

	err = MainRoute(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	authHeader := rec.Header().Get("Authorization")
	assert.NotEmpty(t, authHeader)

	encodedCredentials := authHeader[6:]
	decodedCredentials, err := base64.StdEncoding.DecodeString(encodedCredentials)
	assert.NoError(t, err)

	credentials := string(decodedCredentials)
	assert.Contains(t, credentials, "testuser")
	assert.Contains(t, credentials, testPassword)
}

func TestIntegration_CacheHitMissTransition(t *testing.T) {
	setupTestCacheForIntegration(t)
	server := setupElasticsearchMockServer(t)
	defer server.Close()

	testKey := generateTestKeyForIntegration()

	viper.Set("headers_username", "Remote-User")
	viper.Set("headers_groups", "Remote-Groups")
	viper.Set("headers_email", "Remote-Email")
	viper.Set("enable_group_whitelist", false)
	viper.Set("secret_key", testKey)
	viper.Set("default_roles", []string{"user"})
	viper.Set("group_mappings", map[string][]string{})
	viper.Set("elasticsearch_host", server.URL)
	viper.Set("elasticsearch_username", "elastic")
	viper.Set("elasticsearch_password", "password")
	viper.Set("elasticsearch_dry_run", false)
	viper.Set("extend_cache", false)

	e := echo.New()

	req1 := httptest.NewRequest(http.MethodPost, "/", nil)
	req1.Header.Set("Remote-User", "user1")
	req1.Header.Set("Remote-Groups", "users")
	req1.Header.Set("Remote-Email", "user1@example.com")

	rec1 := httptest.NewRecorder()
	c1 := e.NewContext(req1, rec1)

	err := MainRoute(c1)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec1.Code)

	password1 := rec1.Header().Get("Authorization")
	assert.NotEmpty(t, password1)

	req2 := httptest.NewRequest(http.MethodPost, "/", nil)
	req2.Header.Set("Remote-User", "user1")
	req2.Header.Set("Remote-Groups", "users")

	rec2 := httptest.NewRecorder()
	c2 := e.NewContext(req2, rec2)

	err = MainRoute(c2)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec2.Code)

	password2 := rec2.Header().Get("Authorization")
	assert.NotEmpty(t, password2)

	assert.Equal(t, password1, password2, "Same password should be returned for cache hit")
}

func TestIntegration_CompleteAuthFlow_WithExtendCache(t *testing.T) {
	setupTestCacheForIntegration(t)
	server := setupElasticsearchMockServer(t)
	defer server.Close()

	testKey := generateTestKeyForIntegration()

	viper.Set("headers_username", "Remote-User")
	viper.Set("headers_groups", "Remote-Groups")
	viper.Set("headers_email", "Remote-Email")
	viper.Set("enable_group_whitelist", false)
	viper.Set("secret_key", testKey)
	viper.Set("default_roles", []string{"user"})
	viper.Set("group_mappings", map[string][]string{})
	viper.Set("elasticsearch_host", server.URL)
	viper.Set("elasticsearch_username", "elastic")
	viper.Set("elasticsearch_password", "password")
	viper.Set("elasticsearch_dry_run", false)
	viper.Set("extend_cache", true)
	viper.Set("cache.expiration", "2h")

	e := echo.New()

	req1 := httptest.NewRequest(http.MethodPost, "/", nil)
	req1.Header.Set("Remote-User", "user_extend")
	req1.Header.Set("Remote-Groups", "users")
	req1.Header.Set("Remote-Email", "user@example.com")

	rec1 := httptest.NewRecorder()
	c1 := e.NewContext(req1, rec1)

	err := MainRoute(c1)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec1.Code)

	req2 := httptest.NewRequest(http.MethodPost, "/", nil)
	req2.Header.Set("Remote-User", "user_extend")
	req2.Header.Set("Remote-Groups", "users")

	rec2 := httptest.NewRecorder()
	c2 := e.NewContext(req2, rec2)

	err = MainRoute(c2)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec2.Code)

	authHeader := rec2.Header().Get("Authorization")
	assert.NotEmpty(t, authHeader)
}

func TestIntegration_CompleteAuthFlow_InvalidEmail(t *testing.T) {
	setupTestCacheForIntegration(t)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("Remote-User", "testuser")
	req.Header.Set("Remote-Groups", "admin")
	req.Header.Set("Remote-Email", "invalid-email")

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	testKey := generateTestKeyForIntegration()
	viper.Set("headers_username", "Remote-User")
	viper.Set("headers_groups", "Remote-Groups")
	viper.Set("headers_email", "Remote-Email")
	viper.Set("enable_group_whitelist", false)
	viper.Set("secret_key", testKey)
	viper.Set("elasticsearch_dry_run", true)

	err := MainRoute(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestIntegration_CompleteAuthFlow_InvalidName(t *testing.T) {
	setupTestCacheForIntegration(t)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("Remote-User", "testuser")
	req.Header.Set("Remote-Groups", "admin")
	req.Header.Set("Remote-Name", "invalid\x00name")

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	testKey := generateTestKeyForIntegration()
	viper.Set("headers_username", "Remote-User")
	viper.Set("headers_groups", "Remote-Groups")
	viper.Set("headers_name", "Remote-Name")
	viper.Set("enable_group_whitelist", false)
	viper.Set("secret_key", testKey)
	viper.Set("elasticsearch_dry_run", true)

	err := MainRoute(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestIntegration_CompleteAuthFlow_NoGroups(t *testing.T) {
	setupTestCacheForIntegration(t)
	server := setupElasticsearchMockServer(t)
	defer server.Close()

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("Remote-User", "testuser")

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	testKey := generateTestKeyForIntegration()
	viper.Set("headers_username", "Remote-User")
	viper.Set("headers_groups", "Remote-Groups")
	viper.Set("enable_group_whitelist", false)
	viper.Set("secret_key", testKey)
	viper.Set("default_roles", []string{"user"})
	viper.Set("group_mappings", map[string][]string{})
	viper.Set("elasticsearch_host", server.URL)
	viper.Set("elasticsearch_username", "elastic")
	viper.Set("elasticsearch_password", "password")
	viper.Set("elasticsearch_dry_run", false)
	viper.Set("extend_cache", false)

	err := MainRoute(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	authHeader := rec.Header().Get("Authorization")
	assert.NotEmpty(t, authHeader)
}

func TestIntegration_CompleteAuthFlow_MultipleGropsWithRoleMappings(t *testing.T) {
	setupTestCacheForIntegration(t)
	server := setupElasticsearchMockServer(t)
	defer server.Close()

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("Remote-User", "testuser")
	req.Header.Set("Remote-Groups", "admin,developers,users")
	req.Header.Set("Remote-Email", "testuser@example.com")

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	testKey := generateTestKeyForIntegration()
	viper.Set("headers_username", "Remote-User")
	viper.Set("headers_groups", "Remote-Groups")
	viper.Set("headers_email", "Remote-Email")
	viper.Set("enable_group_whitelist", false)
	viper.Set("secret_key", testKey)
	viper.Set("default_roles", []string{"user"})
	viper.Set("group_mappings", map[string][]string{
		"admin":      {"kibana_admin", "admin"},
		"developers": {"kibana_user", "dev"},
		"users":      {"kibana_user"},
	})
	viper.Set("elasticsearch_host", server.URL)
	viper.Set("elasticsearch_username", "elastic")
	viper.Set("elasticsearch_password", "password")
	viper.Set("elasticsearch_dry_run", false)
	viper.Set("extend_cache", false)

	err := MainRoute(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	authHeader := rec.Header().Get("Authorization")
	assert.NotEmpty(t, authHeader)
}

func TestIntegration_DryRunMode(t *testing.T) {
	setupTestCacheForIntegration(t)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("Remote-User", "testuser")
	req.Header.Set("Remote-Groups", "admin")
	req.Header.Set("Remote-Email", "testuser@example.com")

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	testKey := generateTestKeyForIntegration()
	viper.Set("headers_username", "Remote-User")
	viper.Set("headers_groups", "Remote-Groups")
	viper.Set("headers_email", "Remote-Email")
	viper.Set("enable_group_whitelist", false)
	viper.Set("secret_key", testKey)
	viper.Set("default_roles", []string{"user"})
	viper.Set("group_mappings", map[string][]string{})
	viper.Set("elasticsearch_dry_run", true)
	viper.Set("extend_cache", false)

	err := MainRoute(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	authHeader := rec.Header().Get("Authorization")
	assert.NotEmpty(t, authHeader)
}

func TestIntegration_DecryptionFailure_CorruptedCacheData(t *testing.T) {
	setupTestCacheForIntegration(t)

	ctx := context.Background()

	cacheKey := "elastauth-corrupteduser"
	corruptedData := "this-is-not-valid-base64-or-encrypted-data!!!"
	cache.CacheInstance.Set(ctx, cacheKey, corruptedData)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("Remote-User", "corrupteduser")
	req.Header.Set("Remote-Groups", "admin")

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	testKey := generateTestKeyForIntegration()
	viper.Set("headers_username", "Remote-User")
	viper.Set("headers_groups", "Remote-Groups")
	viper.Set("enable_group_whitelist", false)
	viper.Set("secret_key", testKey)
	viper.Set("elasticsearch_dry_run", true)

	err := MainRoute(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestIntegration_ElasticsearchConnectionError(t *testing.T) {
	setupTestCacheForIntegration(t)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("Remote-User", "testuser")
	req.Header.Set("Remote-Groups", "admin")
	req.Header.Set("Remote-Email", "testuser@example.com")

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	testKey := generateTestKeyForIntegration()
	viper.Set("headers_username", "Remote-User")
	viper.Set("headers_groups", "Remote-Groups")
	viper.Set("headers_email", "Remote-Email")
	viper.Set("enable_group_whitelist", false)
	viper.Set("secret_key", testKey)
	viper.Set("default_roles", []string{"user"})
	viper.Set("group_mappings", map[string][]string{})
	viper.Set("elasticsearch_host", "http://localhost:99999")
	viper.Set("elasticsearch_username", "elastic")
	viper.Set("elasticsearch_password", "password")
	viper.Set("elasticsearch_dry_run", false)
	viper.Set("extend_cache", false)

	err := MainRoute(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestIntegration_SpecialCharactersInUsername(t *testing.T) {
	setupTestCacheForIntegration(t)
	server := setupElasticsearchMockServer(t)
	defer server.Close()

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("Remote-User", "user.name@domain.com")
	req.Header.Set("Remote-Groups", "admin")
	req.Header.Set("Remote-Email", "user.name@domain.com")

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	testKey := generateTestKeyForIntegration()
	viper.Set("headers_username", "Remote-User")
	viper.Set("headers_groups", "Remote-Groups")
	viper.Set("headers_email", "Remote-Email")
	viper.Set("enable_group_whitelist", false)
	viper.Set("secret_key", testKey)
	viper.Set("default_roles", []string{"user"})
	viper.Set("group_mappings", map[string][]string{
		"admin": {"kibana_admin"},
	})
	viper.Set("elasticsearch_host", server.URL)
	viper.Set("elasticsearch_username", "elastic")
	viper.Set("elasticsearch_password", "password")
	viper.Set("elasticsearch_dry_run", false)
	viper.Set("extend_cache", false)

	err := MainRoute(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	authHeader := rec.Header().Get("Authorization")
	assert.NotEmpty(t, authHeader)

	encodedCredentials := authHeader[6:]
	decodedCredentials, err := base64.StdEncoding.DecodeString(encodedCredentials)
	assert.NoError(t, err)
	assert.Contains(t, string(decodedCredentials), "user.name@domain.com")
}
