package oidc

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/spf13/viper"
	"golang.org/x/oauth2"

	"github.com/wasilak/elastauth/provider"
)

func init() {
	provider.DefaultFactory.Register("oidc", NewProviderFromConfig)
}

// NewProviderFromConfig creates a new OIDC provider instance from Viper configuration
func NewProviderFromConfig(config interface{}) (provider.AuthProvider, error) {
	// Read OIDC configuration from Viper
	oidcConfig := Config{
		Issuer:                viper.GetString("oidc.issuer"),
		ClientID:              viper.GetString("oidc.client_id"),
		ClientSecret:          viper.GetString("oidc.client_secret"),
		AuthorizationEndpoint: viper.GetString("oidc.authorization_endpoint"),
		TokenEndpoint:         viper.GetString("oidc.token_endpoint"),
		UserinfoEndpoint:      viper.GetString("oidc.userinfo_endpoint"),
		JWKSURI:               viper.GetString("oidc.jwks_uri"),
		Scopes:                viper.GetStringSlice("oidc.scopes"),
		ClientAuthMethod:      viper.GetString("oidc.client_auth_method"),
		TokenValidation:       viper.GetString("oidc.token_validation"),
		UsePKCE:               viper.GetBool("oidc.use_pkce"),
	}

	// Read claim mappings
	claimMappings := make(map[string]string)
	if viper.IsSet("oidc.claim_mappings") {
		claimMappingsRaw := viper.GetStringMapString("oidc.claim_mappings")
		for k, v := range claimMappingsRaw {
			claimMappings[k] = v
		}
	}
	oidcConfig.ClaimMappings = claimMappings

	// Read custom headers
	customHeaders := make(map[string]string)
	if viper.IsSet("oidc.custom_headers") {
		customHeadersRaw := viper.GetStringMapString("oidc.custom_headers")
		for k, v := range customHeadersRaw {
			customHeaders[k] = v
		}
	}
	oidcConfig.CustomHeaders = customHeaders

	return NewProvider(oidcConfig)
}

// Provider implements the AuthProvider interface for OAuth2/OIDC authentication
type Provider struct {
	config       Config
	provider     *oidc.Provider
	verifier     *oidc.IDTokenVerifier
	oauth2Config *oauth2.Config
}

// Config holds the configuration for the OIDC provider
type Config struct {
	// Standard OAuth2/OIDC settings
	Issuer       string `mapstructure:"issuer"`
	ClientID     string `mapstructure:"client_id"`
	ClientSecret string `mapstructure:"client_secret"`

	// Optional manual endpoint configuration (overrides discovery)
	AuthorizationEndpoint string `mapstructure:"authorization_endpoint"`
	TokenEndpoint         string `mapstructure:"token_endpoint"`
	UserinfoEndpoint      string `mapstructure:"userinfo_endpoint"`
	JWKSURI               string `mapstructure:"jwks_uri"`

	// OAuth2 settings
	Scopes           []string          `mapstructure:"scopes"`
	ClientAuthMethod string            `mapstructure:"client_auth_method"` // "client_secret_basic" or "client_secret_post"
	TokenValidation  string            `mapstructure:"token_validation"`   // "jwks", "userinfo", or "both"
	ClaimMappings    map[string]string `mapstructure:"claim_mappings"`
	CustomHeaders    map[string]string `mapstructure:"custom_headers"`

	// Security settings
	UsePKCE bool `mapstructure:"use_pkce"`
}

// NewProvider creates a new OIDC provider instance
func NewProvider(config Config) (*Provider, error) {
	ctx := context.Background()

	// Validate required configuration
	if err := config.validate(); err != nil {
		return nil, fmt.Errorf("invalid OIDC configuration: %w", err)
	}

	// Set defaults
	config.setDefaults()

	p := &Provider{
		config: config,
	}

	// Initialize OIDC provider (for discovery)
	if config.Issuer != "" {
		provider, err := oidc.NewProvider(ctx, config.Issuer)
		if err != nil {
			return nil, fmt.Errorf("failed to create OIDC provider: %w", err)
		}
		p.provider = provider

		// Create ID token verifier
		p.verifier = provider.Verifier(&oidc.Config{
			ClientID: config.ClientID,
		})
	}

	// Setup OAuth2 configuration
	p.oauth2Config = &oauth2.Config{
		ClientID:     config.ClientID,
		ClientSecret: config.ClientSecret,
		Scopes:       config.Scopes,
	}

	// Set endpoints (either from discovery or manual configuration)
	if p.provider != nil {
		// Use discovery endpoints
		p.oauth2Config.Endpoint = p.provider.Endpoint()
	} else {
		// Use manual endpoints
		p.oauth2Config.Endpoint = oauth2.Endpoint{
			AuthURL:  config.AuthorizationEndpoint,
			TokenURL: config.TokenEndpoint,
		}
	}

	// Set client authentication method
	switch config.ClientAuthMethod {
	case "client_secret_post":
		p.oauth2Config.Endpoint.AuthStyle = oauth2.AuthStyleInParams
	default: // "client_secret_basic"
		p.oauth2Config.Endpoint.AuthStyle = oauth2.AuthStyleInHeader
	}

	return p, nil
}

// GetUser extracts user information from the authentication request
func (p *Provider) GetUser(ctx context.Context, req *provider.AuthRequest) (*provider.UserInfo, error) {
	// Extract token from request (Bearer token or cookie)
	token, err := p.extractToken(req)
	if err != nil {
		return nil, fmt.Errorf("failed to extract authentication token: %w", err)
	}

	// Validate token and extract claims
	claims, err := p.validateToken(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("token validation failed: %w", err)
	}

	// Map claims to UserInfo
	userInfo := p.mapClaimsToUserInfo(claims)
	if userInfo.Username == "" {
		return nil, fmt.Errorf("username claim not found or empty")
	}

	return userInfo, nil
}

// Type returns the provider type identifier
func (p *Provider) Type() string {
	return "oidc"
}

// Validate checks if the provider configuration is valid
func (p *Provider) Validate() error {
	return p.config.validate()
}

// extractToken extracts the authentication token from the request
func (p *Provider) extractToken(req *provider.AuthRequest) (string, error) {
	// Try Bearer token first
	if token, err := req.GetBearerToken(); err == nil && token != "" {
		return token, nil
	}

	// Try cookie-based authentication
	if cookie, err := req.GetCookie("access_token"); err == nil && cookie.Value != "" {
		return cookie.Value, nil
	}

	// Try ID token cookie (common in OIDC flows)
	if cookie, err := req.GetCookie("id_token"); err == nil && cookie.Value != "" {
		return cookie.Value, nil
	}

	return "", fmt.Errorf("no authentication token found in request")
}

// validateToken validates the token using the configured method
func (p *Provider) validateToken(ctx context.Context, token string) (map[string]interface{}, error) {
	switch p.config.TokenValidation {
	case "jwks":
		return p.validateWithJWKS(ctx, token)
	case "userinfo":
		return p.validateWithUserinfo(ctx, token)
	case "both":
		// Try JWKS first, fallback to userinfo
		if claims, err := p.validateWithJWKS(ctx, token); err == nil {
			return claims, nil
		}
		return p.validateWithUserinfo(ctx, token)
	default:
		return nil, fmt.Errorf("invalid token validation method: %s", p.config.TokenValidation)
	}
}

// validateWithJWKS validates the token using JWKS endpoint
func (p *Provider) validateWithJWKS(ctx context.Context, token string) (map[string]interface{}, error) {
	if p.verifier == nil {
		return nil, fmt.Errorf("JWKS verifier not available")
	}

	idToken, err := p.verifier.Verify(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("JWT verification failed: %w", err)
	}

	var claims map[string]interface{}
	if err := idToken.Claims(&claims); err != nil {
		return nil, fmt.Errorf("failed to extract claims: %w", err)
	}

	return claims, nil
}

// validateWithUserinfo validates the token using userinfo endpoint
func (p *Provider) validateWithUserinfo(ctx context.Context, token string) (map[string]interface{}, error) {
	if p.provider == nil {
		// Manual userinfo endpoint
		if p.config.UserinfoEndpoint == "" {
			return nil, fmt.Errorf("userinfo endpoint not configured")
		}
		return p.callUserinfoEndpoint(ctx, token, p.config.UserinfoEndpoint)
	}

	// Use provider's userinfo endpoint
	tokenSource := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	userInfo, err := p.provider.UserInfo(ctx, tokenSource)
	if err != nil {
		return nil, fmt.Errorf("userinfo request failed: %w", err)
	}

	var claims map[string]interface{}
	if err := userInfo.Claims(&claims); err != nil {
		return nil, fmt.Errorf("failed to extract userinfo claims: %w", err)
	}

	return claims, nil
}

// callUserinfoEndpoint makes a direct call to the userinfo endpoint
func (p *Provider) callUserinfoEndpoint(ctx context.Context, token, endpoint string) (map[string]interface{}, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create userinfo request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)

	// Add custom headers if configured
	for key, value := range p.config.CustomHeaders {
		req.Header.Set(key, value)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("userinfo request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("userinfo endpoint returned status %d", resp.StatusCode)
	}

	var claims map[string]interface{}
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&claims); err != nil {
		return nil, fmt.Errorf("failed to decode userinfo response: %w", err)
	}

	return claims, nil
}

// mapClaimsToUserInfo maps OIDC claims to UserInfo structure
func (p *Provider) mapClaimsToUserInfo(claims map[string]interface{}) *provider.UserInfo {
	userInfo := &provider.UserInfo{}

	// Map claims based on configuration
	if username := p.getClaimValue(claims, p.config.ClaimMappings["username"]); username != "" {
		userInfo.Username = username
	}
	if email := p.getClaimValue(claims, p.config.ClaimMappings["email"]); email != "" {
		userInfo.Email = email
	}
	if fullName := p.getClaimValue(claims, p.config.ClaimMappings["full_name"]); fullName != "" {
		userInfo.FullName = fullName
	}
	if groups := p.getClaimSlice(claims, p.config.ClaimMappings["groups"]); len(groups) > 0 {
		userInfo.Groups = groups
	}

	return userInfo
}

// getClaimValue extracts a string value from claims with support for nested claims
func (p *Provider) getClaimValue(claims map[string]interface{}, claimPath string) string {
	if claimPath == "" {
		return ""
	}

	// Support nested claims like "realm_access.roles"
	parts := strings.Split(claimPath, ".")
	current := claims

	for i, part := range parts {
		if i == len(parts)-1 {
			// Last part - extract the value
			if val, ok := current[part]; ok {
				if str, ok := val.(string); ok {
					return str
				}
			}
		} else {
			// Intermediate part - navigate deeper
			if next, ok := current[part].(map[string]interface{}); ok {
				current = next
			} else {
				return ""
			}
		}
	}

	return ""
}

// getClaimSlice extracts a string slice from claims with support for nested claims
func (p *Provider) getClaimSlice(claims map[string]interface{}, claimPath string) []string {
	if claimPath == "" {
		return nil
	}

	// Support nested claims like "realm_access.roles"
	parts := strings.Split(claimPath, ".")
	current := claims

	for i, part := range parts {
		if i == len(parts)-1 {
			if val, ok := current[part]; ok {
				// Handle different group formats
				switch v := val.(type) {
				case []string:
					return v
				case []interface{}:
					var groups []string
					for _, item := range v {
						if str, ok := item.(string); ok {
							groups = append(groups, str)
						}
					}
					return groups
				case string:
					// Single group as string
					return []string{v}
				}
			}
		} else {
			if next, ok := current[part].(map[string]interface{}); ok {
				current = next
			} else {
				return nil
			}
		}
	}

	return nil
}

// validate validates the OIDC configuration
func (c *Config) validate() error {
	if c.ClientID == "" {
		return fmt.Errorf("client_id is required")
	}

	// Either issuer (for discovery) or manual endpoints must be provided
	if c.Issuer == "" {
		if c.AuthorizationEndpoint == "" || c.TokenEndpoint == "" {
			return fmt.Errorf("either issuer (for discovery) or manual endpoints (authorization_endpoint, token_endpoint) must be provided")
		}
	}

	// Validate token validation method
	switch c.TokenValidation {
	case "jwks", "userinfo", "both":
		// Valid
	case "":
		// Will be set to default
	default:
		return fmt.Errorf("invalid token_validation method: %s (must be 'jwks', 'userinfo', or 'both')", c.TokenValidation)
	}

	// Validate client auth method
	switch c.ClientAuthMethod {
	case "client_secret_basic", "client_secret_post":
		// Valid
	case "":
		// Will be set to default
	default:
		return fmt.Errorf("invalid client_auth_method: %s (must be 'client_secret_basic' or 'client_secret_post')", c.ClientAuthMethod)
	}

	return nil
}

// setDefaults sets default values for optional configuration fields
func (c *Config) setDefaults() {
	if len(c.Scopes) == 0 {
		c.Scopes = []string{oidc.ScopeOpenID, "profile", "email"}
	}

	if c.TokenValidation == "" {
		c.TokenValidation = "jwks"
	}

	if c.ClientAuthMethod == "" {
		c.ClientAuthMethod = "client_secret_basic"
	}

	if c.ClaimMappings == nil {
		c.ClaimMappings = map[string]string{
			"username":  "preferred_username",
			"email":     "email",
			"groups":    "groups",
			"full_name": "name",
		}
	}

	// Ensure required claim mappings exist
	if c.ClaimMappings["username"] == "" {
		c.ClaimMappings["username"] = "preferred_username"
	}
	if c.ClaimMappings["email"] == "" {
		c.ClaimMappings["email"] = "email"
	}
	if c.ClaimMappings["groups"] == "" {
		c.ClaimMappings["groups"] = "groups"
	}
	if c.ClaimMappings["full_name"] == "" {
		c.ClaimMappings["full_name"] = "name"
	}

	// Enable PKCE by default for security
	if !c.UsePKCE {
		c.UsePKCE = true
	}
}