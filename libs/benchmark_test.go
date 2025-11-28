package libs

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	gocache "github.com/patrickmn/go-cache"
	"github.com/spf13/viper"
	"github.com/wasilak/elastauth/cache"
	"go.opentelemetry.io/otel"
)

func generateBenchmarkKey() string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func BenchmarkEncrypt(b *testing.B) {
	ctx := context.Background()
	key := generateBenchmarkKey()
	password := "TestPassword123!@#$%^&*()"

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := Encrypt(ctx, password, key)
		if err != nil {
			b.Fatalf("Encryption failed: %v", err)
		}
	}
}

func BenchmarkDecrypt(b *testing.B) {
	ctx := context.Background()
	key := generateBenchmarkKey()
	password := "TestPassword123!@#$%^&*()"

	encryptedPassword, err := Encrypt(ctx, password, key)
	if err != nil {
		b.Fatalf("Failed to setup benchmark: %v", err)
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := Decrypt(ctx, encryptedPassword, key)
		if err != nil {
			b.Fatalf("Decryption failed: %v", err)
		}
	}
}

func BenchmarkEncryptDecryptRoundTrip(b *testing.B) {
	ctx := context.Background()
	key := generateBenchmarkKey()
	password := "TestPassword123!@#$%^&*()"

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		encrypted, err := Encrypt(ctx, password, key)
		if err != nil {
			b.Fatalf("Encryption failed: %v", err)
		}

		decrypted, err := Decrypt(ctx, encrypted, key)
		if err != nil {
			b.Fatalf("Decryption failed: %v", err)
		}

		if decrypted != password {
			b.Fatalf("Round trip failed: expected %s, got %s", password, decrypted)
		}
	}
}

func BenchmarkGenerateTemporaryUserPassword(b *testing.B) {
	ctx := context.Background()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := GenerateTemporaryUserPassword(ctx)
		if err != nil {
			b.Fatalf("Password generation failed: %v", err)
		}
	}
}

func BenchmarkCacheSet(b *testing.B) {
	ctx := context.Background()
	cache.CacheInstance = &cache.GoCache{
		Cache:  gocache.New(1*time.Hour, 2*time.Hour),
		TTL:    1 * time.Hour,
		Tracer: otel.Tracer("bench"),
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		cacheKey := "benchkey-" + string(rune(i))
		cache.CacheInstance.Set(ctx, cacheKey, "test-value")
	}
}

func BenchmarkCacheGet(b *testing.B) {
	ctx := context.Background()
	cache.CacheInstance = &cache.GoCache{
		Cache:  gocache.New(1*time.Hour, 2*time.Hour),
		TTL:    1 * time.Hour,
		Tracer: otel.Tracer("bench"),
	}

	cache.CacheInstance.Set(ctx, "benchkey", "test-value")

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = cache.CacheInstance.Get(ctx, "benchkey")
	}
}

func BenchmarkCacheSetGet(b *testing.B) {
	ctx := context.Background()
	cache.CacheInstance = &cache.GoCache{
		Cache:  gocache.New(1*time.Hour, 2*time.Hour),
		TTL:    1 * time.Hour,
		Tracer: otel.Tracer("bench"),
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		cacheKey := "benchkey"
		cache.CacheInstance.Set(ctx, cacheKey, "test-value")
		_, _ = cache.CacheInstance.Get(ctx, cacheKey)
	}
}

func BenchmarkValidateUsername(b *testing.B) {
	validUsernames := []string{
		"user",
		"user@domain.com",
		"user.name",
		"user_name",
		"user-name",
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		username := validUsernames[i%len(validUsernames)]
		_ = ValidateUsername(username)
	}
}

func BenchmarkValidateEmail(b *testing.B) {
	validEmails := []string{
		"user@example.com",
		"test.user@example.co.uk",
		"user+tag@example.com",
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		email := validEmails[i%len(validEmails)]
		_ = ValidateEmail(email)
	}
}

func BenchmarkValidateName(b *testing.B) {
	validNames := []string{
		"John Doe",
		"Jane Smith",
		"User Name",
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		name := validNames[i%len(validNames)]
		_ = ValidateName(name)
	}
}

func BenchmarkEncodeForCacheKey(b *testing.B) {
	usernames := []string{
		"user",
		"user@domain.com",
		"user.name",
		"user_name-123",
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		username := usernames[i%len(usernames)]
		_ = EncodeForCacheKey(username)
	}
}

func BenchmarkGetUserRoles(b *testing.B) {
	ctx := context.Background()

	viper.Set("default_roles", []string{"user"})
	viper.Set("group_mappings", map[string][]string{
		"admin":      {"kibana_admin"},
		"developers": {"kibana_user", "dev"},
		"users":      {"kibana_user"},
	})

	userGroups := []string{"admin", "developers", "users"}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = GetUserRoles(ctx, userGroups)
	}
}

func BenchmarkParseAndValidateGroups(b *testing.B) {
	groupHeaders := []string{
		"admin,users",
		"admin,developers,users,guests",
		"",
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		groupHeader := groupHeaders[i%len(groupHeaders)]
		_, _ = ParseAndValidateGroups(groupHeader, false, nil)
	}
}

func BenchmarkMainRouteSimplePath(b *testing.B) {
	cache.CacheInstance = &cache.GoCache{
		Cache:  gocache.New(1*time.Hour, 2*time.Hour),
		TTL:    1 * time.Hour,
		Tracer: otel.Tracer("bench"),
	}

	testKey := generateBenchmarkKey()

	viper.Set("headers_username", "Remote-User")
	viper.Set("headers_groups", "Remote-Groups")
	viper.Set("headers_email", "Remote-Email")
	viper.Set("enable_group_whitelist", false)
	viper.Set("secret_key", testKey)
	viper.Set("default_roles", []string{"user"})
	viper.Set("group_mappings", map[string][]string{})
	viper.Set("elasticsearch_dry_run", true)
	viper.Set("extend_cache", false)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		e := echo.New()
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		userNum := i % 1000
		req.Header.Set("Remote-User", "benchuser"+string(rune(48+userNum%10)))
		req.Header.Set("Remote-Groups", "admin,users")
		req.Header.Set("Remote-Email", "benchuser@example.com")

		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		_ = MainRoute(c)
	}
}

func BenchmarkMainRouteCacheHit(b *testing.B) {
	ctx := context.Background()

	cache.CacheInstance = &cache.GoCache{
		Cache:  gocache.New(1*time.Hour, 2*time.Hour),
		TTL:    1 * time.Hour,
		Tracer: otel.Tracer("bench"),
	}

	testKey := generateBenchmarkKey()

	testPassword := "BenchPassword123!@#"
	encryptedPassword, _ := Encrypt(ctx, testPassword, testKey)
	encryptedPasswordBase64 := base64.URLEncoding.EncodeToString([]byte(encryptedPassword))

	cacheKey := "elastauth-benchuser"
	cache.CacheInstance.Set(ctx, cacheKey, encryptedPasswordBase64)

	viper.Set("headers_username", "Remote-User")
	viper.Set("headers_groups", "Remote-Groups")
	viper.Set("enable_group_whitelist", false)
	viper.Set("secret_key", testKey)
	viper.Set("default_roles", []string{"user"})
	viper.Set("group_mappings", map[string][]string{})
	viper.Set("elasticsearch_dry_run", true)
	viper.Set("extend_cache", false)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		e := echo.New()
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		req.Header.Set("Remote-User", "benchuser")
		req.Header.Set("Remote-Groups", "admin,users")

		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		_ = MainRoute(c)
	}
}

func BenchmarkBasicAuth(b *testing.B) {
	credentials := []struct {
		username string
		password string
	}{
		{"user", "password"},
		{"user@domain.com", "Complex!Password123@#$"},
		{"testuser", "p@ssw0rd!@#$%^&*()"},
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		cred := credentials[i%len(credentials)]
		_ = basicAuth(cred.username, cred.password)
	}
}

func BenchmarkSanitizeForLogging(b *testing.B) {
	errors := []interface{}{
		map[string]interface{}{
			"password": "secret",
			"key":      "private_key",
			"token":    "token123",
		},
		"simple error message",
		map[string]string{"error": "test"},
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		err := errors[i%len(errors)]
		_ = SanitizeForLogging(err)
	}
}
