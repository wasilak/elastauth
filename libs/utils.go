package libs

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net/url"
	"os"
	"reflect"
	"regexp"
	"strings"

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

// The function generates a temporary user password that is 32 characters long with a mix of digits,
// symbols, and upper/lower case letters, disallowing repeat characters.
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

// The function retrieves user roles based on their group mappings or default roles if no mappings are
// found.
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

// The function generates a random 32 byte key for AES-256 encryption and returns it as a hexadecimal
// encoded string.
func GenerateKey(ctx context.Context) (string, error) {
	_, span := tracerUtils.Start(ctx, "GenerateKey")
	defer span.End()

	bytes := make([]byte, 32) //generate a random 32 byte key for AES-256
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	return hex.EncodeToString(bytes), nil //encode key in bytes to string for saving

}

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

func EncodeForCacheKey(username string) string {
	return url.QueryEscape(username)
}

func IsSensitiveField(fieldName string) bool {
	lowerName := strings.ToLower(fieldName)
	sensitiveKeywords := []string{"password", "secret", "key", "token", "credential", "auth"}
	for _, keyword := range sensitiveKeywords {
		if strings.Contains(lowerName, keyword) {
			return true
		}
	}
	return false
}

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

	default:
		return data
	}
}

func SafeLogError(err error) string {
	if err == nil {
		return ""
	}
	return "An error occurred. Please contact support if this persists."
}
