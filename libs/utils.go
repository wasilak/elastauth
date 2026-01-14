package libs

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/sethvargo/go-password/password"
	"github.com/spf13/viper"
	"go.opentelemetry.io/otel"
)

var tracerUtils = otel.Tracer("utils")

// The function checks if a given string is present in a slice of strings, ignoring case sensitivity.
func contains(s []string, str string) bool {
	for _, v := range s {
		if strings.EqualFold(strings.ToLower(v), strings.ToLower(str)) {
			return true
		}
	}
	return false
}

// The function returns an array of keys from a given map.
func getMapKeys(itemsMap map[string][]string) []string {
	keys := []string{}

	for k := range itemsMap {
		keys = append(keys, k)
	}

	return keys
}

// GenerateTemporaryUserPassword creates a cryptographically secure temporary password
// for user authentication. The password is 32 characters long and contains a mix of
// digits, symbols, and upper/lower case letters with no repeated characters.
func GenerateTemporaryUserPassword(ctx context.Context) (string, error) {
	_, span := tracerUtils.Start(ctx, "GenerateTemporaryUserPassword")
	defer span.End()

	// `res, err := password.Generate(32, 10, 0, false, false)` is generating a temporary user password
	// that is 32 characters long with a mix of digits, symbols, and upper/lower case letters, disallowing
	// repeat characters. It uses the `go-password` package to generate the password and returns the
	// generated password and any error that occurred during the generation process.
	res, err := password.Generate(32, 10, 0, false, false)
	if err != nil {
		return "", err
	}
	return res, nil
}

// GetUserRoles determines the roles that should be assigned to a user based on their
// group membership. If the user belongs to mapped groups, their roles are retrieved from
// the group_mappings configuration. If no mapped groups are found, default_roles are used.
func GetUserRoles(ctx context.Context, userGroups []string) []string {
	_, span := tracerUtils.Start(ctx, "GetUserRoles")
	defer span.End()

	// This code block is retrieving user roles based on their group mappings or default roles if no
	// mappings are found.
	roles := []string{}
	if len(viper.GetStringMapStringSlice("group_mappings")) > 0 {
		for _, group := range userGroups {
			if contains(getMapKeys(viper.GetStringMapStringSlice("group_mappings")), group) {
				roles = append(roles, viper.GetStringMapStringSlice("group_mappings")[strings.ToLower(group)]...)
			}
		}
	}

	if len(roles) == 0 {
		roles = viper.GetStringSlice("default_roles")
		if roles == nil {
			roles = []string{}
		}
	}

	return roles
}

// The function takes a username and password, combines them into a string, encodes the string using
// base64, and returns the encoded string.
func basicAuth(username, pass string) string {
	auth := username + ":" + pass
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

// GenerateKey generates a cryptographically secure random key suitable for AES-256 encryption.
// The key is returned as a hexadecimal-encoded string (64 characters representing 32 bytes).
func GenerateKey(ctx context.Context) (string, error) {
	_, span := tracerUtils.Start(ctx, "GenerateKey")
	defer span.End()

	bytes := make([]byte, 32) //generate a random 32 byte key for AES-256
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	return hex.EncodeToString(bytes), nil //encode key in bytes to string for saving

}

// GetAppName retrieves the application name from environment variables.
// It first checks OTEL_SERVICE_NAME, then APP_NAME, and defaults to "elastauth" if neither is set.
func GetAppName() string {
	appName := os.Getenv("OTEL_SERVICE_NAME")
	if appName == "" {
		appName = os.Getenv("APP_NAME")
		if appName == "" {
			appName = "elastauth"
		}
	}
	return appName
}

var (
	usernamePattern = regexp.MustCompile(`^[a-zA-Z0-9._\-@]+$`)
	emailPattern    = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
)

// ValidateUsername validates the format and length of a username.
// Valid usernames contain only alphanumeric characters, dots, underscores, hyphens, and at-signs,
// and must be between 1 and 255 characters long.
func ValidateUsername(username string) error {
	if len(username) == 0 {
		return fmt.Errorf("username cannot be empty")
	}
	if len(username) > 255 {
		return fmt.Errorf("username exceeds maximum length of 255 characters (got %d)", len(username))
	}
	if !usernamePattern.MatchString(username) {
		return fmt.Errorf("username contains invalid characters; allowed: alphanumeric, dot, underscore, hyphen, at-sign")
	}
	return nil
}

// ValidateEmail validates the format and length of an email address.
// Valid emails follow standard RFC 5322 basic format requirements and
// must be between 1 and 320 characters long.
func ValidateEmail(email string) error {
	if len(email) == 0 {
		return fmt.Errorf("email cannot be empty")
	}
	if len(email) > 320 {
		return fmt.Errorf("email exceeds maximum length of 320 characters (got %d)", len(email))
	}
	if !emailPattern.MatchString(email) {
		return fmt.Errorf("email format is invalid")
	}
	return nil
}

// ValidateName validates the format and length of a user's full name.
// Valid names must not exceed 500 characters and cannot contain control characters
// (except tab, newline, and carriage return).
func ValidateName(name string) error {
	if len(name) > 500 {
		return fmt.Errorf("name exceeds maximum length of 500 characters (got %d)", len(name))
	}
	for _, r := range name {
		if r < 32 && r != '\t' && r != '\n' && r != '\r' {
			return fmt.Errorf("name contains invalid control characters")
		}
	}
	return nil
}

// ValidateGroupName validates the format and length of a group name.
// Valid group names must be between 1 and 255 characters long and cannot contain control characters.
func ValidateGroupName(group string) error {
	if len(group) == 0 {
		return fmt.Errorf("group name cannot be empty")
	}
	if len(group) > 255 {
		return fmt.Errorf("group name exceeds maximum length of 255 characters (got %d)", len(group))
	}
	for _, r := range group {
		if r < 32 {
			return fmt.Errorf("group name contains invalid control characters")
		}
	}
	return nil
}

// ParseAndValidateGroups parses a comma-separated string of group names and validates each one.
// If enableWhitelist is true, only groups in the provided whitelist are accepted.
// Returns a slice of validated group names or an error if validation fails.
func ParseAndValidateGroups(groupsHeader string, enableWhitelist bool, whitelist []string) ([]string, error) {
	if len(groupsHeader) == 0 {
		return []string{}, nil
	}

	rawGroups := strings.Split(groupsHeader, ",")
	validatedGroups := []string{}

	for _, group := range rawGroups {
		trimmed := strings.TrimSpace(group)
		if len(trimmed) == 0 {
			continue
		}

		if err := ValidateGroupName(trimmed); err != nil {
			return nil, err
		}

		if enableWhitelist {
			if !contains(whitelist, trimmed) {
				return nil, fmt.Errorf("group '%s' is not in whitelist", trimmed)
			}
		}

		validatedGroups = append(validatedGroups, trimmed)
	}

	return validatedGroups, nil
}

// EncodeForCacheKey encodes a username for safe use as a cache key.
// It uses URL encoding to ensure special characters don't interfere with cache key format.
func EncodeForCacheKey(username string) string {
	return url.QueryEscape(username)
}

// IsSensitiveField checks whether a field name represents sensitive data that should be redacted in logs.
// It looks for common sensitive keywords like "password", "secret", "key", "token", "credential", and "auth".
func IsSensitiveField(fieldName string) bool {
	lowerName := strings.ToLower(fieldName)
	sensitiveKeywords := []string{
		"password",
		"secret",
		"key",
		"token",
		"credential",
		"auth",
		"authorization", // HTTP Authorization header
		"bearer",        // Bearer tokens
		"basic",         // Basic auth
		"cookie",        // Session cookies
		"session",       // Session IDs
		"api_key",       // API keys
		"apikey",        // API keys (no underscore)
		"private",       // Private keys
		"cert",          // Certificates
		"x-auth",        // Custom auth headers
		"remote-user",   // Authelia header (contains username, but treated as sensitive)
		"remote-email",  // Authelia header
		"remote-name",   // Authelia header
		"remote-groups", // Authelia header
	}
	for _, keyword := range sensitiveKeywords {
		if strings.Contains(lowerName, keyword) {
			return true
		}
	}
	return false
}

// SanitizeForLogging recursively processes data structures and redacts sensitive fields
// to prevent credentials and secrets from being logged. It handles maps, structs, slices,
// and other types by checking field names against a list of sensitive keywords.
func SanitizeForLogging(data interface{}) interface{} {
	if data == nil {
		return nil
	}

	v := reflect.ValueOf(data)
	switch v.Kind() {
	case reflect.Map:
		result := make(map[string]interface{})
		for _, key := range v.MapKeys() {
			keyStr := fmt.Sprintf("%v", key.Interface())
			val := v.MapIndex(key).Interface()

			if IsSensitiveField(keyStr) {
				result[keyStr] = "***REDACTED***"
			} else {
				result[keyStr] = SanitizeForLogging(val)
			}
		}
		return result

	case reflect.Struct:
		result := make(map[string]interface{})
		t := v.Type()
		for i := 0; i < v.NumField(); i++ {
			field := t.Field(i)
			val := v.Field(i).Interface()

			if IsSensitiveField(field.Name) {
				result[field.Name] = "***REDACTED***"
			} else {
				result[field.Name] = SanitizeForLogging(val)
			}
		}
		return result

	case reflect.Slice, reflect.Array:
		result := make([]interface{}, v.Len())
		for i := 0; i < v.Len(); i++ {
			result[i] = SanitizeForLogging(v.Index(i).Interface())
		}
		return result

	case reflect.String:
		// Check if the string value itself looks like a credential
		// This catches cases where the field name isn't sensitive but the value is
		str := v.String()
		if looksLikeCredential(str) {
			return "***REDACTED***"
		}
		return str

	default:
		return data
	}
}

// looksLikeCredential checks if a string value appears to be a credential
// This provides defense-in-depth by catching credentials even when field names aren't sensitive
func looksLikeCredential(s string) bool {
	// Don't redact empty strings or very short strings
	if len(s) < 8 {
		return false
	}

	// Check for common credential patterns
	lowerStr := strings.ToLower(s)
	
	// Bearer tokens
	if strings.HasPrefix(lowerStr, "bearer ") {
		return true
	}
	
	// Basic auth (Base64 encoded username:password)
	if strings.HasPrefix(lowerStr, "basic ") {
		return true
	}
	
	// JWT tokens (three base64 segments separated by dots)
	if strings.Count(s, ".") == 2 && len(s) > 50 {
		parts := strings.Split(s, ".")
		if len(parts) == 3 && len(parts[0]) > 10 && len(parts[1]) > 10 && len(parts[2]) > 10 {
			return true
		}
	}
	
	// API keys (long alphanumeric strings)
	// This is a heuristic: if it's a long string with high entropy, it might be a key
	if len(s) > 32 && isHighEntropyString(s) {
		return true
	}
	
	return false
}

// isHighEntropyString checks if a string has high entropy (likely a random key/token)
func isHighEntropyString(s string) bool {
	// Count unique characters
	charSet := make(map[rune]bool)
	for _, c := range s {
		charSet[c] = true
	}
	
	// High entropy strings have many unique characters relative to their length
	uniqueRatio := float64(len(charSet)) / float64(len(s))
	
	// If more than 50% of characters are unique, it's likely high entropy
	return uniqueRatio > 0.5
}

// Request ID context key type for type-safe context values
type contextKey string

const requestIDKey contextKey = "request_id"

// GetRequestID extracts the request ID from the context
// Returns an empty string if no request ID is found
func GetRequestID(ctx context.Context) string {
	if reqID, ok := ctx.Value(requestIDKey).(string); ok {
		return reqID
	}
	return ""
}

// SetRequestID adds a request ID to the context
// If the request already has an X-Request-ID header, use that
// Otherwise, generate a new UUID
func SetRequestID(ctx context.Context, r *http.Request) context.Context {
	// Check if request already has a request ID header
	reqID := r.Header.Get("X-Request-ID")
	if reqID == "" {
		// Generate a new UUID for the request
		reqID = generateRequestID()
	}
	return context.WithValue(ctx, requestIDKey, reqID)
}

// generateRequestID generates a simple request ID
// Uses a timestamp and random component for uniqueness
func generateRequestID() string {
	// Use crypto/rand for better randomness
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		// Fallback to timestamp-based ID if random generation fails
		return fmt.Sprintf("req-%d", time.Now().UnixNano())
	}
	return fmt.Sprintf("req-%x", b)
}

// AddRequestIDToLogs returns a new context with the request ID added to slog attributes
// This ensures all logs from this context include the request ID
func AddRequestIDToLogs(ctx context.Context) context.Context {
	reqID := GetRequestID(ctx)
	if reqID == "" {
		return ctx
	}
	
	// Add request ID as a slog attribute to the context
	// This will be included in all logs made with this context
	logger := slog.Default().With(slog.String("request_id", reqID))
	return context.WithValue(ctx, slog.Default(), logger)
}

// SafeLogError returns a generic error message that does not expose implementation details.
// This should be used when displaying errors to end users or in logs where sensitive
// information should not be disclosed.
func SafeLogError(err error) string {
	if err == nil {
		return ""
	}
	return "An error occurred. Please contact support if this persists."
}
