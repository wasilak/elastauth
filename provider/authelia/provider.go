package authelia

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/viper"
	"github.com/wasilak/elastauth/provider"
)

// init registers the Authelia provider with the default factory
func init() {
	provider.DefaultFactory.Register("authelia", NewProvider)
}

// Provider implements the AuthProvider interface for Authelia
type Provider struct {
	config Config
}

// Config holds the configuration for the Authelia provider
type Config struct {
	HeaderUsername string `mapstructure:"header_username"`
	HeaderGroups   string `mapstructure:"header_groups"`
	HeaderEmail    string `mapstructure:"header_email"`
	HeaderName     string `mapstructure:"header_name"`
}

// NewProvider creates a new Authelia provider instance
func NewProvider(config interface{}) (provider.AuthProvider, error) {
	// Read configuration from Viper using the new nested structure
	autheliaConfig := Config{
		HeaderUsername: viper.GetString("authelia.header_username"),
		HeaderGroups:   viper.GetString("authelia.header_groups"),
		HeaderEmail:    viper.GetString("authelia.header_email"),
		HeaderName:     viper.GetString("authelia.header_name"),
	}
	
	// If not found in new structure, try old keys for backward compatibility
	if autheliaConfig.HeaderUsername == "" {
		autheliaConfig.HeaderUsername = viper.GetString("headers_username")
	}
	if autheliaConfig.HeaderGroups == "" {
		autheliaConfig.HeaderGroups = viper.GetString("headers_groups")
	}
	if autheliaConfig.HeaderEmail == "" {
		autheliaConfig.HeaderEmail = viper.GetString("headers_email")
	}
	if autheliaConfig.HeaderName == "" {
		autheliaConfig.HeaderName = viper.GetString("headers_name")
	}
	
	p := &Provider{config: autheliaConfig}
	if err := p.Validate(); err != nil {
		return nil, fmt.Errorf("invalid Authelia provider configuration: %w", err)
	}
	
	return p, nil
}

// GetUser extracts user information from Authelia headers
func (p *Provider) GetUser(ctx context.Context, req *provider.AuthRequest) (*provider.UserInfo, error) {
	username := req.GetHeader(p.config.HeaderUsername)
	if username == "" {
		return nil, fmt.Errorf("username header %s not found", p.config.HeaderUsername)
	}
	
	groupsHeader := req.GetHeader(p.config.HeaderGroups)
	groups := parseGroups(groupsHeader)
	
	email := req.GetHeader(p.config.HeaderEmail)
	fullName := req.GetHeader(p.config.HeaderName)
	
	return &provider.UserInfo{
		Username: username,
		Email:    email,
		Groups:   groups,
		FullName: fullName,
	}, nil
}

// Type returns the provider type identifier
func (p *Provider) Type() string {
	return "authelia"
}

// Validate checks if the provider configuration is valid
func (p *Provider) Validate() error {
	if p.config.HeaderUsername == "" {
		return fmt.Errorf("header_username is required")
	}
	if p.config.HeaderGroups == "" {
		return fmt.Errorf("header_groups is required")
	}
	if p.config.HeaderEmail == "" {
		return fmt.Errorf("header_email is required")
	}
	if p.config.HeaderName == "" {
		return fmt.Errorf("header_name is required")
	}
	return nil
}

// parseGroups parses the comma-separated groups header into a slice
func parseGroups(groupsHeader string) []string {
	if groupsHeader == "" {
		return []string{}
	}
	
	groups := strings.Split(groupsHeader, ",")
	result := make([]string, 0, len(groups))
	
	for _, group := range groups {
		group = strings.TrimSpace(group)
		if group != "" {
			result = append(result, group)
		}
	}
	
	return result
}

// DefaultConfig returns the default configuration for Authelia provider
func DefaultConfig() Config {
	return Config{
		HeaderUsername: "Remote-User",
		HeaderGroups:   "Remote-Groups",
		HeaderEmail:    "Remote-Email",
		HeaderName:     "Remote-Name",
	}
}