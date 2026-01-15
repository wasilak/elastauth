package libs

import (
	"context"
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"log/slog"

	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"go.opentelemetry.io/otel"
)

var tracerConfig = otel.Tracer("config")
var LogLeveler *slog.LevelVar
var validate *validator.Validate

// Config represents the complete application configuration
type Config struct {
	AuthProvider  string `mapstructure:"auth_provider" validate:"required,oneof=authelia oidc"`
	SecretKey     string `mapstructure:"secret_key" validate:"required,len=64,hexadecimal"`
	Listen        string `mapstructure:"listen"`
	LogLevel      string `mapstructure:"log_level" validate:"oneof=debug info warn error"`
	LogFormat     string `mapstructure:"log_format" validate:"oneof=text json"`
	EnableMetrics bool   `mapstructure:"enable_metrics"`
	EnableOtel    bool   `mapstructure:"enableOtel"`

	Elasticsearch ElasticsearchConfig     `mapstructure:"elasticsearch" validate:"required"`
	Cache         CacheConfig             `mapstructure:"cache"`
	Proxy         ProxyConfig             `mapstructure:"proxy"`
	Authelia      AutheliaConfig          `mapstructure:"authelia"`
	OIDC          OIDCConfig              `mapstructure:"oidc"`
	DefaultRoles  []string                `mapstructure:"default_roles"`
	GroupMappings map[string][]string     `mapstructure:"group_mappings"`
}

// ElasticsearchConfig holds Elasticsearch connection settings
type ElasticsearchConfig struct {
	Hosts    []string `mapstructure:"hosts" validate:"required,min=1"`
	Username string   `mapstructure:"username" validate:"required"`
	Password string   `mapstructure:"password" validate:"required"`
	DryRun   bool     `mapstructure:"dry_run"`
}

// CacheConfig holds cache configuration
type CacheConfig struct {
	Type       string        `mapstructure:"type" validate:"omitempty,oneof=memory redis file"`
	Expiration time.Duration `mapstructure:"expiration"`
	RedisHost  string        `mapstructure:"redis_host" validate:"required_if=Type redis"`
	RedisDB    int           `mapstructure:"redis_db" validate:"gte=0,lte=15"`
	Path       string        `mapstructure:"path" validate:"required_if=Type file"`
}

// ProxyConfig holds transparent proxy mode configuration
type ProxyConfig struct {
	Enabled          bool          `mapstructure:"enabled"`
	ElasticsearchURL string        `mapstructure:"elasticsearch_url" validate:"omitempty,url"`
	Timeout          time.Duration `mapstructure:"timeout"`
	MaxIdleConns     int           `mapstructure:"max_idle_conns"`
	IdleConnTimeout  time.Duration `mapstructure:"idle_conn_timeout"`
	TLS              TLSConfig     `mapstructure:"tls"`
}

// TLSConfig holds TLS configuration
type TLSConfig struct {
	Enabled            bool   `mapstructure:"enabled"`
	InsecureSkipVerify bool   `mapstructure:"insecure_skip_verify"`
	CACert             string `mapstructure:"ca_cert"`
	ClientCert         string `mapstructure:"client_cert"`
	ClientKey          string `mapstructure:"client_key"`
}

// AutheliaConfig holds Authelia provider configuration
type AutheliaConfig struct {
	HeaderUsername string `mapstructure:"header_username"`
	HeaderGroups   string `mapstructure:"header_groups"`
	HeaderEmail    string `mapstructure:"header_email"`
	HeaderName     string `mapstructure:"header_name"`
}

// OIDCConfig holds OIDC provider configuration
type OIDCConfig struct {
	Issuer                string            `mapstructure:"issuer" validate:"omitempty,url"`
	ClientID              string            `mapstructure:"client_id"`
	ClientSecret          string            `mapstructure:"client_secret"`
	AuthorizationEndpoint string            `mapstructure:"authorization_endpoint" validate:"omitempty,url"`
	TokenEndpoint         string            `mapstructure:"token_endpoint" validate:"omitempty,url"`
	UserinfoEndpoint      string            `mapstructure:"userinfo_endpoint" validate:"omitempty,url"`
	JWKSURI               string            `mapstructure:"jwks_uri" validate:"omitempty,url"`
	Scopes                []string          `mapstructure:"scopes"`
	ClientAuthMethod      string            `mapstructure:"client_auth_method" validate:"omitempty,oneof=client_secret_basic client_secret_post"`
	TokenValidation       string            `mapstructure:"token_validation" validate:"omitempty,oneof=jwks userinfo both"`
	UsePKCE               bool              `mapstructure:"use_pkce"`
	ClaimMappings         ClaimMappings     `mapstructure:"claim_mappings"`
	CustomHeaders         map[string]string `mapstructure:"custom_headers"`
}

// ClaimMappings holds OIDC claim mapping configuration
type ClaimMappings struct {
	Username string `mapstructure:"username"`
	Email    string `mapstructure:"email"`
	Groups   string `mapstructure:"groups"`
	FullName string `mapstructure:"full_name"`
}

// InitConfiguration initializes the application configuration
func InitConfiguration() error {
	// Only define flags if they haven't been defined yet
	if flag.Lookup("generateKey") == nil {
		flag.Bool("generateKey", false, "Generate valid encryption key for use in app")
		flag.String("listen", "127.0.0.1:5000", "Listen address")
		flag.String("config", "./", "Path to config.yml")
		flag.Bool("enableOtel", false, "Enable OTEL (OpenTelemetry)")
	}

	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()
	viper.BindPFlags(pflag.CommandLine)

	// Load .env file if it exists
	if err := godotenv.Load(); err == nil {
		log.Println("Loaded environment variables from .env file")
	}

	// Configure Viper
	viper.SetEnvPrefix("elastauth")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(viper.GetString("config"))

	// Set defaults
	setDefaults()

	// Read config file
	if err := viper.ReadInConfig(); err != nil {
		log.Println(err)
	}

	// Handle OIDC scopes from environment variable (comma-separated)
	if scopesEnv := os.Getenv("ELASTAUTH_OIDC_SCOPES"); scopesEnv != "" {
		scopes := strings.Split(scopesEnv, ",")
		for i, scope := range scopes {
			scopes[i] = strings.TrimSpace(scope)
		}
		viper.Set("oidc.scopes", scopes)
	}

	// Handle OIDC custom headers from environment variables
	customHeaders := make(map[string]string)
	for _, env := range os.Environ() {
		pair := strings.SplitN(env, "=", 2)
		if len(pair) == 2 && strings.HasPrefix(pair[0], "ELASTAUTH_OIDC_CUSTOM_HEADERS_") {
			headerName := strings.TrimPrefix(pair[0], "ELASTAUTH_OIDC_CUSTOM_HEADERS_")
			headerName = strings.ReplaceAll(headerName, "_", "-")
			customHeaders[headerName] = pair[1]
		}
	}
	if len(customHeaders) > 0 {
		viper.Set("oidc.custom_headers", customHeaders)
	}

	// Initialize validator
	validate = validator.New()

	return nil
}

// setDefaults sets default configuration values
func setDefaults() {
	viper.SetDefault("auth_provider", "authelia")
	viper.SetDefault("listen", "127.0.0.1:5000")
	viper.SetDefault("log_level", "info")
	viper.SetDefault("log_format", "text")
	viper.SetDefault("enable_metrics", false)
	viper.SetDefault("enableOtel", false)

	// Cache defaults
	viper.SetDefault("cache.type", "")
	viper.SetDefault("cache.expiration", "1h")
	viper.SetDefault("cache.redis_host", "localhost:6379")
	viper.SetDefault("cache.redis_db", 0)
	viper.SetDefault("cache.path", "/tmp/elastauth-cache")

	// Authelia defaults
	viper.SetDefault("authelia.header_username", "Remote-User")
	viper.SetDefault("authelia.header_groups", "Remote-Groups")
	viper.SetDefault("authelia.header_email", "Remote-Email")
	viper.SetDefault("authelia.header_name", "Remote-Name")

	// OIDC defaults
	viper.SetDefault("oidc.scopes", []string{"openid", "profile", "email", "groups"})
	viper.SetDefault("oidc.client_auth_method", "client_secret_basic")
	viper.SetDefault("oidc.token_validation", "jwks")
	viper.SetDefault("oidc.use_pkce", true)
	viper.SetDefault("oidc.claim_mappings.username", "preferred_username")
	viper.SetDefault("oidc.claim_mappings.email", "email")
	viper.SetDefault("oidc.claim_mappings.groups", "groups")
	viper.SetDefault("oidc.claim_mappings.full_name", "name")

	// Proxy defaults
	viper.SetDefault("proxy.enabled", false)
	viper.SetDefault("proxy.timeout", "30s")
	viper.SetDefault("proxy.max_idle_conns", 100)
	viper.SetDefault("proxy.idle_conn_timeout", "90s")
	viper.SetDefault("proxy.tls.enabled", false)
	viper.SetDefault("proxy.tls.insecure_skip_verify", false)
}

// LoadConfig loads and validates the configuration
func LoadConfig() (*Config, error) {
	// Initialize validator if not already done
	if validate == nil {
		validate = validator.New()
	}

	var cfg Config

	// Unmarshal configuration
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Validate using validator
	if err := validate.Struct(cfg); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	// Additional custom validations
	// Validate secret key format (validator can't check hex decoding)
	decodedKey, err := hex.DecodeString(cfg.SecretKey)
	if err != nil {
		return nil, fmt.Errorf("secret_key must be hex-encoded: %w", err)
	}
	if len(decodedKey) != 32 {
		return nil, fmt.Errorf("secret_key must be 64 hex characters (32 bytes for AES-256)")
	}

	// Provider-specific validation
	if cfg.AuthProvider == "oidc" {
		if cfg.OIDC.Issuer == "" {
			return nil, fmt.Errorf("oidc.issuer is required when auth_provider is oidc")
		}
		if cfg.OIDC.ClientID == "" {
			return nil, fmt.Errorf("oidc.client_id is required when auth_provider is oidc")
		}
		if cfg.OIDC.ClientSecret == "" {
			return nil, fmt.Errorf("oidc.client_secret is required when auth_provider is oidc")
		}
		if len(cfg.OIDC.Scopes) == 0 {
			return nil, fmt.Errorf("oidc.scopes must contain at least one scope")
		}
		if cfg.OIDC.ClaimMappings.Username == "" || cfg.OIDC.ClaimMappings.Email == "" ||
			cfg.OIDC.ClaimMappings.Groups == "" || cfg.OIDC.ClaimMappings.FullName == "" {
			return nil, fmt.Errorf("oidc.claim_mappings must specify username, email, groups, and full_name")
		}
	}

	// Proxy validation
	if cfg.Proxy.Enabled {
		if cfg.Proxy.ElasticsearchURL == "" {
			return nil, fmt.Errorf("proxy.elasticsearch_url is required when proxy is enabled")
		}
		if cfg.Proxy.MaxIdleConns < 1 {
			return nil, fmt.Errorf("proxy.max_idle_conns must be at least 1")
		}
		
		// Validate TLS cert/key pairs
		if cfg.Proxy.TLS.ClientCert != "" && cfg.Proxy.TLS.ClientKey == "" {
			return nil, fmt.Errorf("proxy.tls.client_key is required when proxy.tls.client_cert is provided")
		}
		if cfg.Proxy.TLS.ClientKey != "" && cfg.Proxy.TLS.ClientCert == "" {
			return nil, fmt.Errorf("proxy.tls.client_cert is required when proxy.tls.client_key is provided")
		}

		// Validate cert files exist if TLS is enabled
		if cfg.Proxy.TLS.Enabled {
			if cfg.Proxy.TLS.CACert != "" {
				if _, err := os.Stat(cfg.Proxy.TLS.CACert); os.IsNotExist(err) {
					return nil, fmt.Errorf("proxy.tls.ca_cert file does not exist: %s", cfg.Proxy.TLS.CACert)
				}
			}
			if cfg.Proxy.TLS.ClientCert != "" {
				if _, err := os.Stat(cfg.Proxy.TLS.ClientCert); os.IsNotExist(err) {
					return nil, fmt.Errorf("proxy.tls.client_cert file does not exist: %s", cfg.Proxy.TLS.ClientCert)
				}
			}
			if cfg.Proxy.TLS.ClientKey != "" {
				if _, err := os.Stat(cfg.Proxy.TLS.ClientKey); os.IsNotExist(err) {
					return nil, fmt.Errorf("proxy.tls.client_key file does not exist: %s", cfg.Proxy.TLS.ClientKey)
				}
			}
		}
	}

	return &cfg, nil
}

// HandleSecretKey manages the encryption secret key
func HandleSecretKey(ctx context.Context) error {
	if viper.GetBool("generateKey") {
		key, err := GenerateKey(ctx)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error generating key: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(key)
		os.Exit(0)
	}

	if len(viper.GetString("secret_key")) == 0 {
		key, err := GenerateKey(ctx)
		if err != nil {
			return err
		}
		viper.Set("secret_key", key)
		slog.WarnContext(ctx, "No secret key configured. Auto-generated random key. Configure ELASTAUTH_SECRET_KEY for persistent key.")
	}

	return nil
}

// ValidateConfiguration validates the loaded configuration
func ValidateConfiguration(ctx context.Context) error {
	cfg, err := LoadConfig()
	if err != nil {
		return err
	}

	// Log configuration info
	slog.InfoContext(ctx, "Configuration loaded successfully",
		slog.String("auth_provider", cfg.AuthProvider),
		slog.String("cache_type", cfg.Cache.Type),
		slog.Bool("proxy_enabled", cfg.Proxy.Enabled))

	return nil
}

// BuildProxyConfig creates a ProxyConfig from viper
func BuildProxyConfig() (*ProxyConfig, error) {
	cfg, err := LoadConfig()
	if err != nil {
		return nil, err
	}

	if !cfg.Proxy.Enabled {
		return nil, nil
	}

	return &cfg.Proxy, nil
}

// GetEffectiveCacheConfig returns the effective cache configuration
func GetEffectiveCacheConfig() map[string]interface{} {
	cfg, err := LoadConfig()
	if err != nil {
		return map[string]interface{}{"type": "disabled"}
	}

	return map[string]interface{}{
		"type":       cfg.Cache.Type,
		"expiration": cfg.Cache.Expiration.String(),
		"redis_host": cfg.Cache.RedisHost,
		"redis_db":   cfg.Cache.RedisDB,
		"path":       cfg.Cache.Path,
	}
}

// GetEffectiveAutheliaConfig returns the effective Authelia configuration
func GetEffectiveAutheliaConfig() map[string]interface{} {
	cfg, err := LoadConfig()
	if err != nil {
		return map[string]interface{}{}
	}

	return map[string]interface{}{
		"header_username": cfg.Authelia.HeaderUsername,
		"header_groups":   cfg.Authelia.HeaderGroups,
		"header_email":    cfg.Authelia.HeaderEmail,
		"header_name":     cfg.Authelia.HeaderName,
	}
}

// GetEffectiveProxyConfig returns the effective proxy configuration
func GetEffectiveProxyConfig() map[string]interface{} {
	cfg, err := LoadConfig()
	if err != nil {
		return map[string]interface{}{}
	}

	return map[string]interface{}{
		"enabled":           cfg.Proxy.Enabled,
		"elasticsearch_url": cfg.Proxy.ElasticsearchURL,
		"timeout":           cfg.Proxy.Timeout.String(),
		"max_idle_conns":    cfg.Proxy.MaxIdleConns,
		"idle_conn_timeout": cfg.Proxy.IdleConnTimeout.String(),
		"tls": map[string]interface{}{
			"enabled":              cfg.Proxy.TLS.Enabled,
			"insecure_skip_verify": cfg.Proxy.TLS.InsecureSkipVerify,
			"ca_cert":              cfg.Proxy.TLS.CACert,
			"client_cert":          cfg.Proxy.TLS.ClientCert,
			"client_key":           cfg.Proxy.TLS.ClientKey,
		},
	}
}
