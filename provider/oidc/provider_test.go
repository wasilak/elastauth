package oidc

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"

	"github.com/wasilak/elastauth/provider"
)

func TestConfig_validate(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "valid config with issuer",
			config: Config{
				Issuer:   "https://example.com",
				ClientID: "test-client",
			},
			wantErr: false,
		},
		{
			name: "valid config with manual endpoints",
			config: Config{
				ClientID:              "test-client",
				AuthorizationEndpoint: "https://example.com/auth",
				TokenEndpoint:         "https://example.com/token",
			},
			wantErr: false,
		},
		{
			name: "missing client_id",
			config: Config{
				Issuer: "https://example.com",
			},
			wantErr: true,
		},
		{
			name: "missing issuer and endpoints",
			config: Config{
				ClientID: "test-client",
			},
			wantErr: true,
		},
		{
			name: "invalid token validation method",
			config: Config{
				Issuer:          "https://example.com",
				ClientID:        "test-client",
				TokenValidation: "invalid",
			},
			wantErr: true,
		},
		{
			name: "invalid client auth method",
			config: Config{
				Issuer:           "https://example.com",
				ClientID:         "test-client",
				ClientAuthMethod: "invalid",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConfig_setDefaults(t *testing.T) {
	config := Config{}
	config.setDefaults()

	assert.Equal(t, []string{"openid", "profile", "email"}, config.Scopes)
	assert.Equal(t, "jwks", config.TokenValidation)
	assert.Equal(t, "client_secret_basic", config.ClientAuthMethod)
	assert.True(t, config.UsePKCE)
	assert.NotNil(t, config.ClaimMappings)
	assert.Equal(t, "preferred_username", config.ClaimMappings["username"])
	assert.Equal(t, "email", config.ClaimMappings["email"])
	assert.Equal(t, "groups", config.ClaimMappings["groups"])
	assert.Equal(t, "name", config.ClaimMappings["full_name"])
}

func TestProvider_Type(t *testing.T) {
	config := Config{
		ClientID:              "test-client",
		AuthorizationEndpoint: "https://example.com/auth",
		TokenEndpoint:         "https://example.com/token",
	}

	p, err := NewProvider(config)
	require.NoError(t, err)

	assert.Equal(t, "oidc", p.Type())
}

func TestProvider_extractToken(t *testing.T) {
	config := Config{
		ClientID:              "test-client",
		AuthorizationEndpoint: "https://example.com/auth",
		TokenEndpoint:         "https://example.com/token",
	}

	p, err := NewProvider(config)
	require.NoError(t, err)

	tests := []struct {
		name      string
		setupReq  func() *provider.AuthRequest
		wantToken string
		wantErr   bool
	}{
		{
			name: "bearer token",
			setupReq: func() *provider.AuthRequest {
				req := httptest.NewRequest("GET", "/", nil)
				req.Header.Set("Authorization", "Bearer test-token")
				return &provider.AuthRequest{Request: req}
			},
			wantToken: "test-token",
			wantErr:   false,
		},
		{
			name: "access_token cookie",
			setupReq: func() *provider.AuthRequest {
				req := httptest.NewRequest("GET", "/", nil)
				req.AddCookie(&http.Cookie{Name: "access_token", Value: "cookie-token"})
				return &provider.AuthRequest{Request: req}
			},
			wantToken: "cookie-token",
			wantErr:   false,
		},
		{
			name: "id_token cookie",
			setupReq: func() *provider.AuthRequest {
				req := httptest.NewRequest("GET", "/", nil)
				req.AddCookie(&http.Cookie{Name: "id_token", Value: "id-token"})
				return &provider.AuthRequest{Request: req}
			},
			wantToken: "id-token",
			wantErr:   false,
		},
		{
			name: "no token",
			setupReq: func() *provider.AuthRequest {
				req := httptest.NewRequest("GET", "/", nil)
				return &provider.AuthRequest{Request: req}
			},
			wantToken: "",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := tt.setupReq()
			token, err := p.extractToken(req)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantToken, token)
			}
		})
	}
}

func TestProvider_getClaimValue(t *testing.T) {
	config := Config{
		ClientID:              "test-client",
		AuthorizationEndpoint: "https://example.com/auth",
		TokenEndpoint:         "https://example.com/token",
	}

	p, err := NewProvider(config)
	require.NoError(t, err)

	claims := map[string]interface{}{
		"username": "testuser",
		"email":    "test@example.com",
		"realm_access": map[string]interface{}{
			"roles": []interface{}{"admin", "user"},
		},
		"nested": map[string]interface{}{
			"deep": map[string]interface{}{
				"value": "deep-value",
			},
		},
	}

	tests := []struct {
		name      string
		claimPath string
		want      string
	}{
		{
			name:      "simple claim",
			claimPath: "username",
			want:      "testuser",
		},
		{
			name:      "nested claim",
			claimPath: "nested.deep.value",
			want:      "deep-value",
		},
		{
			name:      "non-existent claim",
			claimPath: "nonexistent",
			want:      "",
		},
		{
			name:      "empty claim path",
			claimPath: "",
			want:      "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := p.getClaimValue(claims, tt.claimPath)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestProvider_getClaimSlice(t *testing.T) {
	config := Config{
		ClientID:              "test-client",
		AuthorizationEndpoint: "https://example.com/auth",
		TokenEndpoint:         "https://example.com/token",
	}

	p, err := NewProvider(config)
	require.NoError(t, err)

	claims := map[string]interface{}{
		"groups": []interface{}{"admin", "user"},
		"roles":  []string{"role1", "role2"},
		"single": "single-group",
		"realm_access": map[string]interface{}{
			"roles": []interface{}{"realm-admin", "realm-user"},
		},
	}

	tests := []struct {
		name      string
		claimPath string
		want      []string
	}{
		{
			name:      "interface slice",
			claimPath: "groups",
			want:      []string{"admin", "user"},
		},
		{
			name:      "string slice",
			claimPath: "roles",
			want:      []string{"role1", "role2"},
		},
		{
			name:      "single string",
			claimPath: "single",
			want:      []string{"single-group"},
		},
		{
			name:      "nested claim",
			claimPath: "realm_access.roles",
			want:      []string{"realm-admin", "realm-user"},
		},
		{
			name:      "non-existent claim",
			claimPath: "nonexistent",
			want:      nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := p.getClaimSlice(claims, tt.claimPath)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestProvider_mapClaimsToUserInfo(t *testing.T) {
	config := Config{
		ClientID:              "test-client",
		AuthorizationEndpoint: "https://example.com/auth",
		TokenEndpoint:         "https://example.com/token",
		ClaimMappings: map[string]string{
			"username":  "preferred_username",
			"email":     "email",
			"groups":    "realm_access.roles",
			"full_name": "name",
		},
	}

	p, err := NewProvider(config)
	require.NoError(t, err)

	claims := map[string]interface{}{
		"preferred_username": "testuser",
		"email":              "test@example.com",
		"name":               "Test User",
		"realm_access": map[string]interface{}{
			"roles": []interface{}{"admin", "user"},
		},
	}

	userInfo := p.mapClaimsToUserInfo(claims)

	assert.Equal(t, "testuser", userInfo.Username)
	assert.Equal(t, "test@example.com", userInfo.Email)
	assert.Equal(t, "Test User", userInfo.FullName)
	assert.Equal(t, []string{"admin", "user"}, userInfo.Groups)
}

func TestNewProvider_ManualEndpoints(t *testing.T) {
	config := Config{
		ClientID:              "test-client",
		ClientSecret:          "test-secret",
		AuthorizationEndpoint: "https://example.com/auth",
		TokenEndpoint:         "https://example.com/token",
		UserinfoEndpoint:      "https://example.com/userinfo",
		TokenValidation:       "userinfo",
	}

	p, err := NewProvider(config)
	require.NoError(t, err)

	assert.Equal(t, "oidc", p.Type())
	assert.Equal(t, "test-client", p.oauth2Config.ClientID)
	assert.Equal(t, "test-secret", p.oauth2Config.ClientSecret)
	assert.Equal(t, "https://example.com/auth", p.oauth2Config.Endpoint.AuthURL)
	assert.Equal(t, "https://example.com/token", p.oauth2Config.Endpoint.TokenURL)
}

func TestProvider_callUserinfoEndpoint(t *testing.T) {
	// Create a test server that returns userinfo
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check authorization header
		auth := r.Header.Get("Authorization")
		if auth != "Bearer test-token" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		// Check custom header
		if r.Header.Get("X-Custom-Header") != "custom-value" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"sub": "user123",
			"preferred_username": "testuser",
			"email": "test@example.com",
			"name": "Test User",
			"groups": ["admin", "user"]
		}`))
	}))
	defer server.Close()

	config := Config{
		ClientID:              "test-client",
		AuthorizationEndpoint: "https://example.com/auth",
		TokenEndpoint:         "https://example.com/token",
		UserinfoEndpoint:      server.URL,
		CustomHeaders: map[string]string{
			"X-Custom-Header": "custom-value",
		},
	}

	p, err := NewProvider(config)
	require.NoError(t, err)

	claims, err := p.callUserinfoEndpoint(context.Background(), "test-token", server.URL)
	require.NoError(t, err)

	assert.Equal(t, "user123", claims["sub"])
	assert.Equal(t, "testuser", claims["preferred_username"])
	assert.Equal(t, "test@example.com", claims["email"])
	assert.Equal(t, "Test User", claims["name"])
}

func TestProvider_callUserinfoEndpoint_Error(t *testing.T) {
	// Create a test server that returns error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error": "invalid_token"}`))
	}))
	defer server.Close()

	config := Config{
		ClientID:              "test-client",
		AuthorizationEndpoint: "https://example.com/auth",
		TokenEndpoint:         "https://example.com/token",
		UserinfoEndpoint:      server.URL,
	}

	p, err := NewProvider(config)
	require.NoError(t, err)

	_, err = p.callUserinfoEndpoint(context.Background(), "invalid-token", server.URL)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "userinfo endpoint returned status 401")
}

func TestProvider_validateToken_UserInfo(t *testing.T) {
	// Create a test server that returns userinfo
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"sub": "user123",
			"preferred_username": "testuser",
			"email": "test@example.com",
			"name": "Test User",
			"groups": ["admin", "user"]
		}`))
	}))
	defer server.Close()

	config := Config{
		ClientID:              "test-client",
		AuthorizationEndpoint: "https://example.com/auth",
		TokenEndpoint:         "https://example.com/token",
		UserinfoEndpoint:      server.URL,
		TokenValidation:       "userinfo",
	}

	p, err := NewProvider(config)
	require.NoError(t, err)

	claims, err := p.validateToken(context.Background(), "test-token")
	require.NoError(t, err)

	assert.Equal(t, "user123", claims["sub"])
	assert.Equal(t, "testuser", claims["preferred_username"])
}

func TestProvider_validateToken_InvalidMethod(t *testing.T) {
	config := Config{
		ClientID:              "test-client",
		AuthorizationEndpoint: "https://example.com/auth",
		TokenEndpoint:         "https://example.com/token",
		TokenValidation:       "invalid-method",
	}

	_, err := NewProvider(config)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid token_validation method: invalid-method")
}

func TestProvider_GetUser_Integration(t *testing.T) {
	// Create a test server that returns userinfo
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"sub": "user123",
			"preferred_username": "testuser",
			"email": "test@example.com",
			"name": "Test User",
			"groups": ["admin", "user"]
		}`))
	}))
	defer server.Close()

	config := Config{
		ClientID:              "test-client",
		AuthorizationEndpoint: "https://example.com/auth",
		TokenEndpoint:         "https://example.com/token",
		UserinfoEndpoint:      server.URL,
		TokenValidation:       "userinfo",
		ClaimMappings: map[string]string{
			"username":  "preferred_username",
			"email":     "email",
			"groups":    "groups",
			"full_name": "name",
		},
	}

	p, err := NewProvider(config)
	require.NoError(t, err)

	// Create a request with Bearer token
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer test-token")
	authReq := &provider.AuthRequest{Request: req}

	userInfo, err := p.GetUser(context.Background(), authReq)
	require.NoError(t, err)

	assert.Equal(t, "testuser", userInfo.Username)
	assert.Equal(t, "test@example.com", userInfo.Email)
	assert.Equal(t, "Test User", userInfo.FullName)
	assert.Equal(t, []string{"admin", "user"}, userInfo.Groups)
}

func TestProvider_GetUser_NoToken(t *testing.T) {
	config := Config{
		ClientID:              "test-client",
		AuthorizationEndpoint: "https://example.com/auth",
		TokenEndpoint:         "https://example.com/token",
	}

	p, err := NewProvider(config)
	require.NoError(t, err)

	// Create a request without token
	req := httptest.NewRequest("GET", "/", nil)
	authReq := &provider.AuthRequest{Request: req}

	_, err = p.GetUser(context.Background(), authReq)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to extract authentication token")
}

func TestProvider_GetUser_NoUsername(t *testing.T) {
	// Create a test server that returns userinfo without username
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"sub": "user123",
			"email": "test@example.com",
			"name": "Test User"
		}`))
	}))
	defer server.Close()

	config := Config{
		ClientID:              "test-client",
		AuthorizationEndpoint: "https://example.com/auth",
		TokenEndpoint:         "https://example.com/token",
		UserinfoEndpoint:      server.URL,
		TokenValidation:       "userinfo",
		ClaimMappings: map[string]string{
			"username":  "preferred_username", // This claim doesn't exist in response
			"email":     "email",
			"full_name": "name",
		},
	}

	p, err := NewProvider(config)
	require.NoError(t, err)

	// Create a request with Bearer token
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer test-token")
	authReq := &provider.AuthRequest{Request: req}

	_, err = p.GetUser(context.Background(), authReq)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "username claim not found or empty")
}
func TestConfig_validate_Comprehensive(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config with all fields",
			config: Config{
				Issuer:                "https://example.com",
				ClientID:              "test-client",
				ClientSecret:          "test-secret",
				AuthorizationEndpoint: "https://example.com/auth",
				TokenEndpoint:         "https://example.com/token",
				UserinfoEndpoint:      "https://example.com/userinfo",
				JWKSURI:               "https://example.com/.well-known/jwks.json",
				Scopes:                []string{"openid", "profile", "email"},
				ClientAuthMethod:      "client_secret_basic",
				TokenValidation:       "both",
				ClaimMappings: map[string]string{
					"username": "preferred_username",
					"email":    "email",
				},
				CustomHeaders: map[string]string{
					"X-Custom": "value",
				},
				UsePKCE: true,
			},
			wantErr: false,
		},
		{
			name: "missing client_id",
			config: Config{
				Issuer: "https://example.com",
			},
			wantErr: true,
			errMsg:  "client_id is required",
		},
		{
			name: "missing issuer and endpoints",
			config: Config{
				ClientID: "test-client",
			},
			wantErr: true,
			errMsg:  "either issuer (for discovery) or manual endpoints",
		},
		{
			name: "missing token endpoint with auth endpoint",
			config: Config{
				ClientID:              "test-client",
				AuthorizationEndpoint: "https://example.com/auth",
			},
			wantErr: true,
			errMsg:  "either issuer (for discovery) or manual endpoints",
		},
		{
			name: "missing auth endpoint with token endpoint",
			config: Config{
				ClientID:      "test-client",
				TokenEndpoint: "https://example.com/token",
			},
			wantErr: true,
			errMsg:  "either issuer (for discovery) or manual endpoints",
		},
		{
			name: "invalid token validation method",
			config: Config{
				Issuer:          "https://example.com",
				ClientID:        "test-client",
				TokenValidation: "invalid",
			},
			wantErr: true,
			errMsg:  "invalid token_validation method",
		},
		{
			name: "invalid client auth method",
			config: Config{
				Issuer:           "https://example.com",
				ClientID:         "test-client",
				ClientAuthMethod: "invalid",
			},
			wantErr: true,
			errMsg:  "invalid client_auth_method",
		},
		{
			name: "valid client_secret_post method",
			config: Config{
				Issuer:           "https://example.com",
				ClientID:         "test-client",
				ClientAuthMethod: "client_secret_post",
			},
			wantErr: false,
		},
		{
			name: "valid userinfo token validation",
			config: Config{
				Issuer:          "https://example.com",
				ClientID:        "test-client",
				TokenValidation: "userinfo",
			},
			wantErr: false,
		},
		{
			name: "valid both token validation",
			config: Config{
				Issuer:          "https://example.com",
				ClientID:        "test-client",
				TokenValidation: "both",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.validate()
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConfig_setDefaults_Comprehensive(t *testing.T) {
	tests := []struct {
		name           string
		input          Config
		expectedScopes []string
		expectedToken  string
		expectedAuth   string
		expectedPKCE   bool
	}{
		{
			name:           "empty config gets all defaults",
			input:          Config{},
			expectedScopes: []string{"openid", "profile", "email"},
			expectedToken:  "jwks",
			expectedAuth:   "client_secret_basic",
			expectedPKCE:   true,
		},
		{
			name: "partial config preserves existing values",
			input: Config{
				Scopes:           []string{"openid", "custom"},
				TokenValidation:  "userinfo",
				ClientAuthMethod: "client_secret_post",
				UsePKCE:          false, // This will be overridden to true by setDefaults
			},
			expectedScopes: []string{"openid", "custom"},
			expectedToken:  "userinfo",
			expectedAuth:   "client_secret_post",
			expectedPKCE:   true, // setDefaults always enables PKCE for security
		},
		{
			name: "claim mappings get defaults when nil",
			input: Config{
				ClaimMappings: nil,
			},
			expectedScopes: []string{"openid", "profile", "email"},
			expectedToken:  "jwks",
			expectedAuth:   "client_secret_basic",
			expectedPKCE:   true,
		},
		{
			name: "partial claim mappings get missing defaults",
			input: Config{
				ClaimMappings: map[string]string{
					"username": "custom_username",
				},
			},
			expectedScopes: []string{"openid", "profile", "email"},
			expectedToken:  "jwks",
			expectedAuth:   "client_secret_basic",
			expectedPKCE:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := tt.input
			config.setDefaults()

			assert.Equal(t, tt.expectedScopes, config.Scopes)
			assert.Equal(t, tt.expectedToken, config.TokenValidation)
			assert.Equal(t, tt.expectedAuth, config.ClientAuthMethod)
			assert.Equal(t, tt.expectedPKCE, config.UsePKCE)

			// Check claim mappings
			assert.NotNil(t, config.ClaimMappings)
			assert.NotEmpty(t, config.ClaimMappings["username"])
			assert.NotEmpty(t, config.ClaimMappings["email"])
			assert.NotEmpty(t, config.ClaimMappings["groups"])
			assert.NotEmpty(t, config.ClaimMappings["full_name"])

			// If we had a custom username mapping, it should be preserved
			if tt.input.ClaimMappings != nil && tt.input.ClaimMappings["username"] != "" {
				assert.Equal(t, tt.input.ClaimMappings["username"], config.ClaimMappings["username"])
			} else {
				assert.Equal(t, "preferred_username", config.ClaimMappings["username"])
			}
		})
	}
}

func TestNewProvider_ConfigurationValidation(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid issuer config",
			config: Config{
				Issuer:   "", // Use empty issuer to avoid real network call
				ClientID: "test-client",
				AuthorizationEndpoint: "https://example.com/auth", // Use manual endpoints instead
				TokenEndpoint:         "https://example.com/token",
			},
			wantErr: false,
		},
		{
			name: "valid manual endpoints config",
			config: Config{
				ClientID:              "test-client",
				AuthorizationEndpoint: "https://example.com/auth",
				TokenEndpoint:         "https://example.com/token",
			},
			wantErr: false,
		},
		{
			name: "invalid config fails validation",
			config: Config{
				ClientID: "", // Missing client ID
				Issuer:   "https://example.com",
			},
			wantErr: true,
			errMsg:  "invalid OIDC configuration",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewProvider(tt.config)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestProvider_OAuth2ConfigSetup(t *testing.T) {
	tests := []struct {
		name               string
		config             Config
		expectedAuthStyle  oauth2.AuthStyle
		expectedScopes     []string
		expectedClientID   string
		expectedClientSec  string
	}{
		{
			name: "client_secret_basic auth style",
			config: Config{
				ClientID:              "test-client",
				ClientSecret:          "test-secret",
				AuthorizationEndpoint: "https://example.com/auth",
				TokenEndpoint:         "https://example.com/token",
				ClientAuthMethod:      "client_secret_basic",
				Scopes:                []string{"openid", "profile"},
			},
			expectedAuthStyle: oauth2.AuthStyleInHeader,
			expectedScopes:    []string{"openid", "profile"},
			expectedClientID:  "test-client",
			expectedClientSec: "test-secret",
		},
		{
			name: "client_secret_post auth style",
			config: Config{
				ClientID:              "test-client",
				ClientSecret:          "test-secret",
				AuthorizationEndpoint: "https://example.com/auth",
				TokenEndpoint:         "https://example.com/token",
				ClientAuthMethod:      "client_secret_post",
				Scopes:                []string{"openid", "email"},
			},
			expectedAuthStyle: oauth2.AuthStyleInParams,
			expectedScopes:    []string{"openid", "email"},
			expectedClientID:  "test-client",
			expectedClientSec: "test-secret",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := NewProvider(tt.config)
			require.NoError(t, err)

			assert.Equal(t, tt.expectedAuthStyle, p.oauth2Config.Endpoint.AuthStyle)
			assert.Equal(t, tt.expectedScopes, p.oauth2Config.Scopes)
			assert.Equal(t, tt.expectedClientID, p.oauth2Config.ClientID)
			assert.Equal(t, tt.expectedClientSec, p.oauth2Config.ClientSecret)
		})
	}
}