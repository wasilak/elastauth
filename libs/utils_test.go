package libs

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestValidateUsername(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid username with letters",
			input:   "john",
			wantErr: false,
		},
		{
			name:    "valid username with numbers",
			input:   "john123",
			wantErr: false,
		},
		{
			name:    "valid username with dot",
			input:   "john.doe",
			wantErr: false,
		},
		{
			name:    "valid username with underscore",
			input:   "john_doe",
			wantErr: false,
		},
		{
			name:    "valid username with hyphen",
			input:   "john-doe",
			wantErr: false,
		},
		{
			name:    "valid username with at-sign",
			input:   "john@example",
			wantErr: false,
		},
		{
			name:    "valid username mixed characters",
			input:   "john.doe_123-test@domain",
			wantErr: false,
		},
		{
			name:    "empty username",
			input:   "",
			wantErr: true,
			errMsg:  "cannot be empty",
		},
		{
			name:    "username exceeding max length",
			input:   "a" + string(make([]byte, 256)),
			wantErr: true,
			errMsg:  "exceeds maximum length",
		},
		{
			name:    "username exactly at max length",
			input:   "a" + strings.Repeat("b", 254),
			wantErr: false,
		},
		{
			name:    "username with space",
			input:   "john doe",
			wantErr: true,
			errMsg:  "invalid characters",
		},
		{
			name:    "username with special character $",
			input:   "john$doe",
			wantErr: true,
			errMsg:  "invalid characters",
		},
		{
			name:    "username with special character !",
			input:   "john!doe",
			wantErr: true,
			errMsg:  "invalid characters",
		},
		{
			name:    "username with special character #",
			input:   "john#doe",
			wantErr: true,
			errMsg:  "invalid characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateUsername(tt.input)
			if tt.wantErr {
				assert.Error(t, err, "expected error")
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err, "expected no error")
			}
		})
	}
}

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid email",
			input:   "user@example.com",
			wantErr: false,
		},
		{
			name:    "valid email with dots",
			input:   "user.name@example.co.uk",
			wantErr: false,
		},
		{
			name:    "valid email with numbers",
			input:   "user123@test456.com",
			wantErr: false,
		},
		{
			name:    "valid email with plus",
			input:   "user+tag@example.com",
			wantErr: false,
		},
		{
			name:    "valid email with hyphen",
			input:   "user-name@ex-ample.com",
			wantErr: false,
		},
		{
			name:    "empty email",
			input:   "",
			wantErr: true,
			errMsg:  "cannot be empty",
		},
		{
			name:    "email exceeding max length",
			input:   "a" + string(make([]byte, 320)),
			wantErr: true,
			errMsg:  "exceeds maximum length",
		},
		{
			name:    "email exactly at max length",
			input:   "a" + string(make([]byte, 319)) + "@b.c",
			wantErr: true,
			errMsg:  "exceeds maximum length",
		},
		{
			name:    "email without @",
			input:   "userexample.com",
			wantErr: true,
			errMsg:  "format is invalid",
		},
		{
			name:    "email without domain",
			input:   "user@",
			wantErr: true,
			errMsg:  "format is invalid",
		},
		{
			name:    "email without TLD",
			input:   "user@example",
			wantErr: true,
			errMsg:  "format is invalid",
		},
		{
			name:    "email with space",
			input:   "user @example.com",
			wantErr: true,
			errMsg:  "format is invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateEmail(tt.input)
			if tt.wantErr {
				assert.Error(t, err, "expected error")
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err, "expected no error")
			}
		})
	}
}

func TestValidateName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid name with letters",
			input:   "John Doe",
			wantErr: false,
		},
		{
			name:    "valid name with numbers",
			input:   "John123",
			wantErr: false,
		},
		{
			name:    "valid name with special characters",
			input:   "John O'Brien-Smith",
			wantErr: false,
		},
		{
			name:    "empty name",
			input:   "",
			wantErr: false,
		},
		{
			name:    "name exactly at max length",
			input:   strings.Repeat("a", 500),
			wantErr: false,
		},
		{
			name:    "name exceeding max length",
			input:   "a" + string(make([]byte, 500)),
			wantErr: true,
			errMsg:  "exceeds maximum length",
		},
		{
			name:    "name with control character NUL",
			input:   "John\x00Doe",
			wantErr: true,
			errMsg:  "invalid control characters",
		},
		{
			name:    "name with control character SOH",
			input:   "John\x01Doe",
			wantErr: true,
			errMsg:  "invalid control characters",
		},
		{
			name:    "name with tab (allowed)",
			input:   "John\tDoe",
			wantErr: false,
		},
		{
			name:    "name with newline (allowed)",
			input:   "John\nDoe",
			wantErr: false,
		},
		{
			name:    "name with carriage return (allowed)",
			input:   "John\rDoe",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateName(tt.input)
			if tt.wantErr {
				assert.Error(t, err, "expected error")
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err, "expected no error")
			}
		})
	}
}

func TestValidateGroupName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid group name",
			input:   "admin",
			wantErr: false,
		},
		{
			name:    "valid group name with spaces",
			input:   "Domain Admins",
			wantErr: false,
		},
		{
			name:    "valid group name with special characters",
			input:   "admin-group_test",
			wantErr: false,
		},
		{
			name:    "empty group name",
			input:   "",
			wantErr: true,
			errMsg:  "cannot be empty",
		},
		{
			name:    "group name exactly at max length",
			input:   strings.Repeat("a", 255),
			wantErr: false,
		},
		{
			name:    "group name exceeding max length",
			input:   "a" + string(make([]byte, 255)),
			wantErr: true,
			errMsg:  "exceeds maximum length",
		},
		{
			name:    "group name with control character NUL",
			input:   "admin\x00group",
			wantErr: true,
			errMsg:  "invalid control characters",
		},
		{
			name:    "group name with control character SOH",
			input:   "admin\x01group",
			wantErr: true,
			errMsg:  "invalid control characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateGroupName(tt.input)
			if tt.wantErr {
				assert.Error(t, err, "expected error")
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err, "expected no error")
			}
		})
	}
}

func TestParseAndValidateGroups(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		enableWhitelist bool
		whitelist       []string
		wantGroups      []string
		wantErr         bool
		errMsg          string
	}{
		{
			name:            "single valid group",
			input:           "admin",
			enableWhitelist: false,
			wantGroups:      []string{"admin"},
			wantErr:         false,
		},
		{
			name:            "multiple valid groups",
			input:           "admin, users, guests",
			enableWhitelist: false,
			wantGroups:      []string{"admin", "users", "guests"},
			wantErr:         false,
		},
		{
			name:            "groups with whitespace",
			input:           "  admin  ,  users  ,  guests  ",
			enableWhitelist: false,
			wantGroups:      []string{"admin", "users", "guests"},
			wantErr:         false,
		},
		{
			name:            "empty groups string",
			input:           "",
			enableWhitelist: false,
			wantGroups:      []string{},
			wantErr:         false,
		},
		{
			name:            "groups with only whitespace",
			input:           "  ,  ,  ",
			enableWhitelist: false,
			wantGroups:      []string{},
			wantErr:         false,
		},
		{
			name:            "group with invalid control character",
			input:           "admin\x00group",
			enableWhitelist: false,
			wantErr:         true,
			errMsg:          "invalid control characters",
		},
		{
			name:            "whitelist enabled, group in whitelist",
			input:           "admin",
			enableWhitelist: true,
			whitelist:       []string{"admin", "users"},
			wantGroups:      []string{"admin"},
			wantErr:         false,
		},
		{
			name:            "whitelist enabled, group not in whitelist",
			input:           "unknown",
			enableWhitelist: true,
			whitelist:       []string{"admin", "users"},
			wantErr:         true,
			errMsg:          "not in whitelist",
		},
		{
			name:            "whitelist enabled, multiple groups some in whitelist",
			input:           "admin, unknown",
			enableWhitelist: true,
			whitelist:       []string{"admin", "users"},
			wantErr:         true,
			errMsg:          "not in whitelist",
		},
		{
			name:            "single group exceeding max length",
			input:           "a" + string(make([]byte, 255)),
			enableWhitelist: false,
			wantErr:         true,
			errMsg:          "exceeds maximum length",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			groups, err := ParseAndValidateGroups(tt.input, tt.enableWhitelist, tt.whitelist)
			if tt.wantErr {
				assert.Error(t, err, "expected error")
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err, "expected no error")
				assert.Equal(t, tt.wantGroups, groups, "expected groups to match")
			}
		})
	}
}

func TestEncodeForCacheKey(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple username",
			input:    "john",
			expected: "john",
		},
		{
			name:     "username with special characters",
			input:    "john@example.com",
			expected: "john%40example.com",
		},
		{
			name:     "username with space",
			input:    "john doe",
			expected: "john+doe",
		},
		{
			name:     "username with dots",
			input:    "john.doe",
			expected: "john.doe",
		},
		{
			name:     "username with underscore",
			input:    "john_doe",
			expected: "john_doe",
		},
		{
			name:     "username with hyphen",
			input:    "john-doe",
			expected: "john-doe",
		},
		{
			name:     "username with multiple special characters",
			input:    "user@domain.com@example.org",
			expected: "user%40domain.com%40example.org",
		},
		{
			name:     "empty username",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := EncodeForCacheKey(tt.input)
			assert.Equal(t, tt.expected, result, "expected encoded key to match")
		})
	}
}

func TestIsSensitiveField(t *testing.T) {
	tests := []struct {
		name      string
		fieldName string
		expected  bool
	}{
		{
			name:      "password field",
			fieldName: "password",
			expected:  true,
		},
		{
			name:      "secret field",
			fieldName: "secret",
			expected:  true,
		},
		{
			name:      "api_key field",
			fieldName: "api_key",
			expected:  true,
		},
		{
			name:      "token field",
			fieldName: "token",
			expected:  true,
		},
		{
			name:      "credential field",
			fieldName: "credential",
			expected:  true,
		},
		{
			name:      "auth field",
			fieldName: "auth",
			expected:  true,
		},
		{
			name:      "secret_key field",
			fieldName: "secret_key",
			expected:  true,
		},
		{
			name:      "authentication field",
			fieldName: "authentication",
			expected:  true,
		},
		{
			name:      "username field",
			fieldName: "username",
			expected:  false,
		},
		{
			name:      "email field",
			fieldName: "email",
			expected:  false,
		},
		{
			name:      "name field",
			fieldName: "name",
			expected:  false,
		},
		{
			name:      "empty field name",
			fieldName: "",
			expected:  false,
		},
		{
			name:      "mixed case PASSWORD",
			fieldName: "PASSWORD",
			expected:  true,
		},
		{
			name:      "mixed case SecretKey",
			fieldName: "SecretKey",
			expected:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsSensitiveField(tt.fieldName)
			assert.Equal(t, tt.expected, result, "expected sensitivity check to match")
		})
	}
}

func TestSanitizeForLogging(t *testing.T) {
	tests := []struct {
		name             string
		input            interface{}
		shouldContain    []string
		shouldNotContain []string
	}{
		{
			name: "map with password",
			input: map[string]interface{}{
				"username": "john",
				"password": "secret123",
			},
			shouldContain:    []string{"john", "REDACTED"},
			shouldNotContain: []string{"secret123"},
		},
		{
			name: "map with multiple sensitive fields",
			input: map[string]interface{}{
				"username": "john",
				"password": "secret123",
				"api_key":  "key123",
				"email":    "john@example.com",
			},
			shouldContain:    []string{"john", "john@example.com", "REDACTED"},
			shouldNotContain: []string{"secret123", "key123"},
		},
		{
			name: "nested map with sensitive fields",
			input: map[string]interface{}{
				"user": map[string]interface{}{
					"username": "john",
					"password": "secret123",
				},
				"config": map[string]interface{}{
					"api_key": "key123",
				},
			},
			shouldContain:    []string{"john", "REDACTED"},
			shouldNotContain: []string{"secret123", "key123"},
		},
		{
			name:          "nil value",
			input:         nil,
			shouldContain: []string{},
		},
		{
			name:          "string value",
			input:         "test_string",
			shouldContain: []string{"test_string"},
		},
		{
			name: "slice with maps",
			input: []interface{}{
				map[string]interface{}{
					"username": "john",
					"password": "secret",
				},
			},
			shouldContain:    []string{"john", "REDACTED"},
			shouldNotContain: []string{"secret"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeForLogging(tt.input)
			resultStr := fmt.Sprint(result)

			for _, shouldContain := range tt.shouldContain {
				if shouldContain != "" {
					assert.Contains(t, resultStr, shouldContain)
				}
			}

			for _, shouldNotContain := range tt.shouldNotContain {
				assert.NotContains(t, resultStr, shouldNotContain)
			}
		})
	}
}

func TestSafeLogError(t *testing.T) {
	tests := []struct {
		name    string
		input   error
		wantMsg string
	}{
		{
			name:    "nil error",
			input:   nil,
			wantMsg: "",
		},
		{
			name:    "non-nil error",
			input:   assert.AnError,
			wantMsg: "An error occurred. Please contact support if this persists.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SafeLogError(tt.input)
			assert.Equal(t, tt.wantMsg, result)
		})
	}
}

func TestGetUserRoles(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name            string
		userGroups      []string
		groupMappings   map[string][]string
		defaultRoles    []string
		expectedRoles   []string
		expectedContent []string
	}{
		{
			name:            "empty groups with default roles",
			userGroups:      []string{},
			groupMappings:   map[string][]string{},
			defaultRoles:    []string{"user", "viewer"},
			expectedRoles:   []string{"user", "viewer"},
			expectedContent: []string{"user", "viewer"},
		},
		{
			name:            "single group with mapping",
			userGroups:      []string{"admin_group"},
			groupMappings:   map[string][]string{"admin_group": {"admin", "superuser"}},
			defaultRoles:    []string{"user"},
			expectedRoles:   []string{"admin", "superuser"},
			expectedContent: []string{"admin", "superuser"},
		},
		{
			name:            "multiple groups with mappings",
			userGroups:      []string{"admin_group", "users_group"},
			groupMappings:   map[string][]string{"admin_group": {"admin"}, "users_group": {"user", "viewer"}},
			defaultRoles:    []string{},
			expectedRoles:   []string{"admin", "user", "viewer"},
			expectedContent: []string{"admin", "user", "viewer"},
		},
		{
			name:            "groups with partial mapping (fallback to default)",
			userGroups:      []string{"unmapped_group"},
			groupMappings:   map[string][]string{"admin_group": {"admin"}},
			defaultRoles:    []string{"user"},
			expectedRoles:   []string{"user"},
			expectedContent: []string{"user"},
		},
		{
			name:            "mixed mapped and unmapped groups",
			userGroups:      []string{"admin_group", "unmapped_group"},
			groupMappings:   map[string][]string{"admin_group": {"admin"}},
			defaultRoles:    []string{"user"},
			expectedRoles:   []string{"admin"},
			expectedContent: []string{"admin"},
		},
		{
			name:            "no groups, no mappings, with default roles",
			userGroups:      []string{},
			groupMappings:   map[string][]string{},
			defaultRoles:    []string{"default_user"},
			expectedRoles:   []string{"default_user"},
			expectedContent: []string{"default_user"},
		},
		{
			name:            "no groups, no mappings, no default roles",
			userGroups:      []string{},
			groupMappings:   map[string][]string{},
			defaultRoles:    []string{},
			expectedRoles:   []string{},
			expectedContent: []string{},
		},
		{
			name:            "case insensitive group mapping",
			userGroups:      []string{"Admin_Group"},
			groupMappings:   map[string][]string{"admin_group": {"admin"}},
			defaultRoles:    []string{"user"},
			expectedRoles:   []string{"admin"},
			expectedContent: []string{"admin"},
		},
		{
			name:            "group mapping with multiple roles per group",
			userGroups:      []string{"developers"},
			groupMappings:   map[string][]string{"developers": {"developer", "code_reviewer", "deployer"}},
			defaultRoles:    []string{},
			expectedRoles:   []string{"developer", "code_reviewer", "deployer"},
			expectedContent: []string{"developer", "code_reviewer", "deployer"},
		},
		{
			name:            "empty roles from mapping uses default",
			userGroups:      []string{"empty_group"},
			groupMappings:   map[string][]string{"empty_group": {}},
			defaultRoles:    []string{"default_role"},
			expectedRoles:   []string{"default_role"},
			expectedContent: []string{"default_role"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			viper.Set("group_mappings", tt.groupMappings)
			viper.Set("default_roles", tt.defaultRoles)

			result := GetUserRoles(ctx, tt.userGroups)

			assert.Len(t, result, len(tt.expectedContent))
			for _, expectedRole := range tt.expectedContent {
				assert.Contains(t, result, expectedRole)
			}

			viper.Set("group_mappings", map[string][]string{})
			viper.Set("default_roles", []string{})
		})
	}
}
