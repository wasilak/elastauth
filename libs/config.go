package libs

import (
	"context"
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"log/slog"

	"go.opentelemetry.io/otel"

	"github.com/joho/godotenv"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var tracerConfig = otel.Tracer("config")

var LogLeveler *slog.LevelVar

// ProxyConfig holds configuration for transparent proxy mode
type ProxyConfig struct {
	Enabled          bool          `mapstructure:"enabled"`
	ElasticsearchURL string        `mapstructure:"elasticsearch_url"`
	Timeout          time.Duration `mapstructure:"timeout"`
	MaxIdleConns     int           `mapstructure:"max_idle_conns"`
	IdleConnTimeout  time.Duration `mapstructure:"idle_conn_timeout"`
	TLS              TLSConfig     `mapstructure:"tls"`
}

// TLSConfig holds TLS configuration for Elasticsearch connections
type TLSConfig struct {
	Enabled            bool   `mapstructure:"enabled"`
	InsecureSkipVerify bool   `mapstructure:"insecure_skip_verify"`
	CACert             string `mapstructure:"ca_cert"`
	ClientCert         string `mapstructure:"client_cert"`
	ClientKey          string `mapstructure:"client_key"`
}

// InitConfiguration initializes the application configuration with proper precedence:
// 1. Environment variables (highest precedence) - prefixed with ELASTAUTH_
// 2. Configuration file values (middle precedence) - config.yml
// 3. Default values (lowest precedence)
//
// Environment Variable Support:
// - All configuration settings can be overridden via environment variables
// - Provider-specific settings: ELASTAUTH_<PROVIDER>_<SETTING>
// - Nested settings use underscores: ELASTAUTH_OIDC_CLIENT_SECRET
// - Special handling for arrays (OIDC scopes) and maps (custom headers)
// - Sensitive values should be set via environment variables for security
//
// Examples:
// - ELASTAUTH_AUTH_PROVIDER=oidc
// - ELASTAUTH_OIDC_CLIENT_SECRET=secret123
// - ELASTAUTH_OIDC_SCOPES=openid,profile,email
// - ELASTAUTH_OIDC_CUSTOM_HEADERS_X_CUSTOM_HEADER=value
func InitConfiguration() error {
	flag.Bool("generateKey", false, "Generate valid encryption key for use in app")
	flag.String("listen", "127.0.0.1:5000", "Listen address")
	flag.String("config", "./", "Path to config.yml")
	flag.Bool("enableOtel", false, "Enable OTEL (OpenTelemetry)")

	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()
	viper.BindPFlags(pflag.CommandLine)

	// Load .env file if it exists (before other configuration)
	// This allows .env to set environment variables that Viper will read
	// godotenv.Load() silently ignores missing .env files
	if err := godotenv.Load(); err == nil {
		log.Println("Loaded environment variables from .env file")
	}

	// Configure environment variable support with proper precedence
	viper.SetEnvPrefix("elastauth")
	viper.AutomaticEnv()
	
	// Enable environment variable substitution in config files
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(viper.GetString("config"))

	// Set defaults first (lowest precedence)
	setConfigurationDefaults()

	// Read config file (middle precedence)
	err := viper.ReadInConfig()
	if err != nil {
		log.Println(err)
	}

	// Bind specific environment variables for provider-specific settings
	// This ensures environment variables override config file values
	bindProviderEnvironmentVariables()

	return nil
}

// setConfigurationDefaults sets all default configuration values
func setConfigurationDefaults() {
	// Provider configuration defaults
	viper.SetDefault("auth_provider", "authelia")

	// New cachego configuration defaults
	viper.SetDefault("cache.type", "")
	viper.SetDefault("cache.expiration", "1h")
	viper.SetDefault("cache.redis_host", "localhost:6379")
	viper.SetDefault("cache.redis_db", 0)
	viper.SetDefault("cache.path", "/tmp/elastauth-cache")

	// Elasticsearch configuration defaults - support both single and multiple endpoints
	viper.SetDefault("elasticsearch_dry_run", false)
	viper.SetDefault("elasticsearch.hosts", []string{})
	viper.SetDefault("elasticsearch.username", "")
	viper.SetDefault("elasticsearch.password", "")
	viper.SetDefault("elasticsearch.dry_run", false)

	// Authelia provider defaults
	viper.SetDefault("authelia.header_username", "Remote-User")
	viper.SetDefault("authelia.header_groups", "Remote-Groups")
	viper.SetDefault("authelia.header_email", "Remote-Email")
	viper.SetDefault("authelia.header_name", "Remote-Name")

	// Legacy header defaults (for backward compatibility)
	viper.SetDefault("headers_username", "Remote-User")
	viper.SetDefault("headers_groups", "Remote-Groups")
	viper.SetDefault("headers_email", "Remote-Email")
	viper.SetDefault("headers_name", "Remote-Name")

	// OIDC provider defaults
	viper.SetDefault("oidc.scopes", []string{"openid", "profile", "email"})
	viper.SetDefault("oidc.client_auth_method", "client_secret_basic")
	viper.SetDefault("oidc.token_validation", "jwks")
	viper.SetDefault("oidc.use_pkce", true)
	
	// OIDC claim mapping defaults
	viper.SetDefault("oidc.claim_mappings.username", "preferred_username")
	viper.SetDefault("oidc.claim_mappings.email", "email")
	viper.SetDefault("oidc.claim_mappings.groups", "groups")
	viper.SetDefault("oidc.claim_mappings.full_name", "name")

	// Proxy configuration defaults
	viper.SetDefault("proxy.enabled", false)
	viper.SetDefault("proxy.elasticsearch_url", "")
	viper.SetDefault("proxy.timeout", "30s")
	viper.SetDefault("proxy.max_idle_conns", 100)
	viper.SetDefault("proxy.idle_conn_timeout", "90s")
	viper.SetDefault("proxy.tls.enabled", false)
	viper.SetDefault("proxy.tls.insecure_skip_verify", false)
	viper.SetDefault("proxy.tls.ca_cert", "")
	viper.SetDefault("proxy.tls.client_cert", "")
	viper.SetDefault("proxy.tls.client_key", "")

	viper.SetDefault("enable_metrics", false)
	viper.SetDefault("enableOtel", false)
	viper.SetDefault("log_level", "info")
	viper.SetDefault("log_format", "text")
}

// bindProviderEnvironmentVariables explicitly binds environment variables for provider-specific settings
// This ensures proper precedence: environment variables > config files > defaults
func bindProviderEnvironmentVariables() {
	// Core configuration
	viper.BindEnv("auth_provider", "ELASTAUTH_AUTH_PROVIDER")
	viper.BindEnv("secret_key", "ELASTAUTH_SECRET_KEY")
	
	// Legacy Elasticsearch configuration (for backward compatibility)
	viper.BindEnv("elasticsearch_host", "ELASTAUTH_ELASTICSEARCH_HOST")
	viper.BindEnv("elasticsearch_username", "ELASTAUTH_ELASTICSEARCH_USERNAME")
	viper.BindEnv("elasticsearch_password", "ELASTAUTH_ELASTICSEARCH_PASSWORD")
	viper.BindEnv("elasticsearch_dry_run", "ELASTAUTH_ELASTICSEARCH_DRY_RUN")
	
	// New Elasticsearch configuration (multi-endpoint support)
	viper.BindEnv("elasticsearch.username", "ELASTAUTH_ELASTICSEARCH_USERNAME")
	viper.BindEnv("elasticsearch.password", "ELASTAUTH_ELASTICSEARCH_PASSWORD")
	viper.BindEnv("elasticsearch.dry_run", "ELASTAUTH_ELASTICSEARCH_DRY_RUN")
	
	// Cache configuration
	viper.BindEnv("cache.type", "ELASTAUTH_CACHE_TYPE")
	viper.BindEnv("cache.expiration", "ELASTAUTH_CACHE_EXPIRATION")
	viper.BindEnv("cache.redis_host", "ELASTAUTH_CACHE_REDIS_HOST")
	viper.BindEnv("cache.redis_db", "ELASTAUTH_CACHE_REDIS_DB")
	viper.BindEnv("cache.path", "ELASTAUTH_CACHE_PATH")
	
	// Authelia provider configuration
	viper.BindEnv("authelia.header_username", "ELASTAUTH_AUTHELIA_HEADER_USERNAME")
	viper.BindEnv("authelia.header_groups", "ELASTAUTH_AUTHELIA_HEADER_GROUPS")
	viper.BindEnv("authelia.header_email", "ELASTAUTH_AUTHELIA_HEADER_EMAIL")
	viper.BindEnv("authelia.header_name", "ELASTAUTH_AUTHELIA_HEADER_NAME")
	
	// Legacy header configuration (for backward compatibility)
	viper.BindEnv("headers_username", "ELASTAUTH_HEADERS_USERNAME")
	viper.BindEnv("headers_groups", "ELASTAUTH_HEADERS_GROUPS")
	viper.BindEnv("headers_email", "ELASTAUTH_HEADERS_EMAIL")
	viper.BindEnv("headers_name", "ELASTAUTH_HEADERS_NAME")
	
	// OIDC provider configuration
	viper.BindEnv("oidc.issuer", "ELASTAUTH_OIDC_ISSUER")
	viper.BindEnv("oidc.client_id", "ELASTAUTH_OIDC_CLIENT_ID")
	viper.BindEnv("oidc.client_secret", "ELASTAUTH_OIDC_CLIENT_SECRET")
	viper.BindEnv("oidc.authorization_endpoint", "ELASTAUTH_OIDC_AUTHORIZATION_ENDPOINT")
	viper.BindEnv("oidc.token_endpoint", "ELASTAUTH_OIDC_TOKEN_ENDPOINT")
	viper.BindEnv("oidc.userinfo_endpoint", "ELASTAUTH_OIDC_USERINFO_ENDPOINT")
	viper.BindEnv("oidc.jwks_uri", "ELASTAUTH_OIDC_JWKS_URI")
	viper.BindEnv("oidc.client_auth_method", "ELASTAUTH_OIDC_CLIENT_AUTH_METHOD")
	viper.BindEnv("oidc.token_validation", "ELASTAUTH_OIDC_TOKEN_VALIDATION")
	viper.BindEnv("oidc.use_pkce", "ELASTAUTH_OIDC_USE_PKCE")
	
	// OIDC claim mappings
	viper.BindEnv("oidc.claim_mappings.username", "ELASTAUTH_OIDC_CLAIM_MAPPINGS_USERNAME")
	viper.BindEnv("oidc.claim_mappings.email", "ELASTAUTH_OIDC_CLAIM_MAPPINGS_EMAIL")
	viper.BindEnv("oidc.claim_mappings.groups", "ELASTAUTH_OIDC_CLAIM_MAPPINGS_GROUPS")
	viper.BindEnv("oidc.claim_mappings.full_name", "ELASTAUTH_OIDC_CLAIM_MAPPINGS_FULL_NAME")
	
	// OIDC scopes (special handling for string slice)
	bindOIDCScopesEnvironmentVariable()
	
	// OIDC custom headers (special handling for map[string]string)
	bindOIDCCustomHeadersEnvironmentVariables()
	
	// Proxy configuration
	viper.BindEnv("proxy.enabled", "ELASTAUTH_PROXY_ENABLED")
	viper.BindEnv("proxy.elasticsearch_url", "ELASTAUTH_PROXY_ELASTICSEARCH_URL")
	viper.BindEnv("proxy.timeout", "ELASTAUTH_PROXY_TIMEOUT")
	viper.BindEnv("proxy.max_idle_conns", "ELASTAUTH_PROXY_MAX_IDLE_CONNS")
	viper.BindEnv("proxy.idle_conn_timeout", "ELASTAUTH_PROXY_IDLE_CONN_TIMEOUT")
	viper.BindEnv("proxy.tls.enabled", "ELASTAUTH_PROXY_TLS_ENABLED")
	viper.BindEnv("proxy.tls.insecure_skip_verify", "ELASTAUTH_PROXY_TLS_INSECURE_SKIP_VERIFY")
	viper.BindEnv("proxy.tls.ca_cert", "ELASTAUTH_PROXY_TLS_CA_CERT")
	viper.BindEnv("proxy.tls.client_cert", "ELASTAUTH_PROXY_TLS_CLIENT_CERT")
	viper.BindEnv("proxy.tls.client_key", "ELASTAUTH_PROXY_TLS_CLIENT_KEY")
	
	// Logging and metrics
	viper.BindEnv("log_level", "ELASTAUTH_LOG_LEVEL")
	viper.BindEnv("log_format", "ELASTAUTH_LOG_FORMAT")
	viper.BindEnv("enable_metrics", "ELASTAUTH_ENABLE_METRICS")
	viper.BindEnv("enableOtel", "ELASTAUTH_ENABLE_OTEL")
}

// bindOIDCScopesEnvironmentVariable handles the special case of OIDC scopes which is a string slice
func bindOIDCScopesEnvironmentVariable() {
	// Check if OIDC scopes are provided via environment variable
	if scopesEnv := os.Getenv("ELASTAUTH_OIDC_SCOPES"); scopesEnv != "" {
		// Split comma-separated scopes
		scopes := strings.Split(scopesEnv, ",")
		// Trim whitespace from each scope
		for i, scope := range scopes {
			scopes[i] = strings.TrimSpace(scope)
		}
		viper.Set("oidc.scopes", scopes)
	}
}

// bindOIDCCustomHeadersEnvironmentVariables handles OIDC custom headers from environment variables
// Custom headers can be set using ELASTAUTH_OIDC_CUSTOM_HEADERS_<HEADER_NAME> format
// For example: ELASTAUTH_OIDC_CUSTOM_HEADERS_X_CUSTOM_HEADER=value
func bindOIDCCustomHeadersEnvironmentVariables() {
	customHeaders := make(map[string]string)
	
	// Scan all environment variables for OIDC custom headers
	for _, env := range os.Environ() {
		pair := strings.SplitN(env, "=", 2)
		if len(pair) != 2 {
			continue
		}
		
		key := pair[0]
		value := pair[1]
		
		// Check if this is an OIDC custom header environment variable
		if strings.HasPrefix(key, "ELASTAUTH_OIDC_CUSTOM_HEADERS_") {
			// Extract the header name from the environment variable name
			headerName := strings.TrimPrefix(key, "ELASTAUTH_OIDC_CUSTOM_HEADERS_")
			// Convert underscores back to hyphens for HTTP header names
			headerName = strings.ReplaceAll(headerName, "_", "-")
			customHeaders[headerName] = value
		}
	}
	
	// Set the custom headers in viper if any were found
	if len(customHeaders) > 0 {
		viper.Set("oidc.custom_headers", customHeaders)
	}
}

// GetSupportedEnvironmentVariables returns a list of all supported environment variables
// This is useful for documentation and validation purposes
func GetSupportedEnvironmentVariables() []string {
	return []string{
		// Core configuration
		"ELASTAUTH_AUTH_PROVIDER",
		"ELASTAUTH_SECRET_KEY",
		"ELASTAUTH_ELASTICSEARCH_HOST",
		"ELASTAUTH_ELASTICSEARCH_USERNAME", 
		"ELASTAUTH_ELASTICSEARCH_PASSWORD",
		"ELASTAUTH_ELASTICSEARCH_DRY_RUN",
		
		// Cache configuration
		"ELASTAUTH_CACHE_TYPE",
		"ELASTAUTH_CACHE_EXPIRATION",
		"ELASTAUTH_CACHE_REDIS_HOST",
		"ELASTAUTH_CACHE_REDIS_DB",
		"ELASTAUTH_CACHE_PATH",
		
		// Authelia provider configuration
		"ELASTAUTH_AUTHELIA_HEADER_USERNAME",
		"ELASTAUTH_AUTHELIA_HEADER_GROUPS",
		"ELASTAUTH_AUTHELIA_HEADER_EMAIL",
		"ELASTAUTH_AUTHELIA_HEADER_NAME",
		
		// Legacy header configuration (backward compatibility)
		"ELASTAUTH_HEADERS_USERNAME",
		"ELASTAUTH_HEADERS_GROUPS",
		"ELASTAUTH_HEADERS_EMAIL",
		"ELASTAUTH_HEADERS_NAME",
		
		// OIDC provider configuration
		"ELASTAUTH_OIDC_ISSUER",
		"ELASTAUTH_OIDC_CLIENT_ID",
		"ELASTAUTH_OIDC_CLIENT_SECRET",
		"ELASTAUTH_OIDC_AUTHORIZATION_ENDPOINT",
		"ELASTAUTH_OIDC_TOKEN_ENDPOINT",
		"ELASTAUTH_OIDC_USERINFO_ENDPOINT",
		"ELASTAUTH_OIDC_JWKS_URI",
		"ELASTAUTH_OIDC_CLIENT_AUTH_METHOD",
		"ELASTAUTH_OIDC_TOKEN_VALIDATION",
		"ELASTAUTH_OIDC_USE_PKCE",
		"ELASTAUTH_OIDC_SCOPES", // Comma-separated list
		
		// OIDC claim mappings
		"ELASTAUTH_OIDC_CLAIM_MAPPINGS_USERNAME",
		"ELASTAUTH_OIDC_CLAIM_MAPPINGS_EMAIL",
		"ELASTAUTH_OIDC_CLAIM_MAPPINGS_GROUPS",
		"ELASTAUTH_OIDC_CLAIM_MAPPINGS_FULL_NAME",
		
		// OIDC custom headers (pattern: ELASTAUTH_OIDC_CUSTOM_HEADERS_<HEADER_NAME>)
		// Example: ELASTAUTH_OIDC_CUSTOM_HEADERS_X_CUSTOM_HEADER
		
		// Proxy configuration
		"ELASTAUTH_PROXY_ENABLED",
		"ELASTAUTH_PROXY_ELASTICSEARCH_URL",
		"ELASTAUTH_PROXY_TIMEOUT",
		"ELASTAUTH_PROXY_MAX_IDLE_CONNS",
		"ELASTAUTH_PROXY_IDLE_CONN_TIMEOUT",
		"ELASTAUTH_PROXY_TLS_ENABLED",
		"ELASTAUTH_PROXY_TLS_INSECURE_SKIP_VERIFY",
		"ELASTAUTH_PROXY_TLS_CA_CERT",
		"ELASTAUTH_PROXY_TLS_CLIENT_CERT",
		"ELASTAUTH_PROXY_TLS_CLIENT_KEY",
		
		// Logging and metrics
		"ELASTAUTH_LOG_LEVEL",
		"ELASTAUTH_LOG_FORMAT",
		"ELASTAUTH_ENABLE_METRICS",
		"ELASTAUTH_ENABLE_OTEL",
	}
}

// HandleSecretKey manages the encryption secret key configuration.
// If the generateKey flag is set, it generates a new key, prints it, and exits.
// If no secret key is configured, it generates a random one and logs a warning.
func HandleSecretKey(ctx context.Context) error {
	_, span := tracerConfig.Start(ctx, "HandleSecretKey")
	defer span.End()

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

// ValidateSecretKey validates that the secret key is properly configured as a 64-character
// hexadecimal string representing a 32-byte (256-bit) AES key.
func ValidateSecretKey(key string) error {
	if len(key) == 0 {
		return fmt.Errorf("secret_key is required (set via ELASTAUTH_SECRET_KEY environment variable)")
	}

	decodedKey, err := hex.DecodeString(key)
	if err != nil {
		return fmt.Errorf("secret_key must be hex-encoded: %w", err)
	}

	if len(decodedKey) != 32 {
		return fmt.Errorf("secret_key must be 64 hex characters (32 bytes for AES-256), got %d hex characters (%d bytes)", len(key), len(decodedKey))
	}

	return nil
}

// ValidateRequiredConfig checks that all required configuration parameters are set.
// Required parameters include Elasticsearch credentials, host, and the encryption secret key.
func ValidateRequiredConfig(ctx context.Context) error {
	required := map[string]string{
		"elasticsearch_host":     "ELASTAUTH_ELASTICSEARCH_HOST",
		"elasticsearch_username": "ELASTAUTH_ELASTICSEARCH_USERNAME",
		"elasticsearch_password": "ELASTAUTH_ELASTICSEARCH_PASSWORD",
		"secret_key":             "ELASTAUTH_SECRET_KEY",
	}

	missing := []string{}
	for key, envVar := range required {
		if len(viper.GetString(key)) == 0 {
			missing = append(missing, fmt.Sprintf("%s (set via %s)", key, envVar))
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("required configuration missing: %s", strings.Join(missing, ", "))
	}

	return nil
}

// ValidateConfiguration performs comprehensive validation of all configuration parameters.
// It checks required settings, secret key format, cache type, log levels, provider configuration, and other configuration options.
func ValidateConfiguration(ctx context.Context) error {
	if err := ValidateRequiredConfig(ctx); err != nil {
		return err
	}

	if err := ValidateSecretKey(viper.GetString("secret_key")); err != nil {
		return err
	}

	// Validate provider configuration
	if err := ValidateProviderConfiguration(ctx); err != nil {
		return err
	}

	// Validate cache configuration (both legacy and new cachego format)
	if err := ValidateCacheConfiguration(ctx); err != nil {
		return err
	}

	// Validate Elasticsearch configuration
	if err := ValidateElasticsearchConfiguration(); err != nil {
		return err
	}

	// Validate proxy configuration
	if err := ValidateProxyConfiguration(); err != nil {
		return err
	}

	validLogLevels := []string{"debug", "info", "warn", "error"}
	logLevel := viper.GetString("log_level")
	validLevel := false
	for _, level := range validLogLevels {
		if logLevel == level {
			validLevel = true
			break
		}
	}
	if !validLevel {
		return fmt.Errorf("invalid log_level: %s (must be one of: debug, info, warn, error)", logLevel)
	}

	return nil
}
// ValidateProviderConfiguration validates the authentication provider configuration.
// It ensures exactly one provider is configured and validates provider-specific settings.
func ValidateProviderConfiguration(ctx context.Context) error {
	authProvider := viper.GetString("auth_provider")
	
	// If no auth_provider is set, default to "authelia" for backward compatibility
	if authProvider == "" {
		authProvider = "authelia"
		viper.Set("auth_provider", authProvider)
	}
	
	// Validate provider type
	validProviders := []string{"authelia", "oidc"}
	validProvider := false
	for _, provider := range validProviders {
		if authProvider == provider {
			validProvider = true
			break
		}
	}
	if !validProvider {
		return fmt.Errorf("invalid auth_provider: %s (must be one of: authelia, oidc)", authProvider)
	}

	// Validate that exactly one provider is configured by checking for conflicting explicit configuration
	// We only check for explicit configuration that conflicts with the selected provider
	
	// Count how many providers have explicit configuration beyond defaults
	explicitProviders := []string{}
	
	// Check if user explicitly configured a different provider than what's selected
	if authProvider != "authelia" {
		// Check if Authelia is explicitly configured when another provider is selected
		if hasExplicitAutheliaConfig() {
			explicitProviders = append(explicitProviders, "authelia")
		}
	}
	
	if authProvider != "oidc" {
		// Check if OIDC is explicitly configured when another provider is selected
		if hasExplicitOIDCConfig() {
			explicitProviders = append(explicitProviders, "oidc")
		}
	}
	
	// If there are conflicting explicit configurations, return error
	if len(explicitProviders) > 0 {
		return fmt.Errorf("auth_provider is set to '%s' but explicit configuration found for: %v. Only configure the selected provider", authProvider, explicitProviders)
	}

	// Validate provider-specific configuration
	switch authProvider {
	case "authelia":
		return ValidateAutheliaConfiguration(ctx)
	case "oidc":
		return ValidateOIDCConfiguration(ctx)
	}

	return nil
}

// hasExplicitAutheliaConfig checks if user has explicitly configured Authelia beyond defaults
func hasExplicitAutheliaConfig() bool {
	// Check for explicit Authelia configuration that differs from defaults
	return viper.IsSet("authelia.header_username") && viper.GetString("authelia.header_username") != "Remote-User" ||
		   viper.IsSet("authelia.header_groups") && viper.GetString("authelia.header_groups") != "Remote-Groups" ||
		   viper.IsSet("authelia.header_email") && viper.GetString("authelia.header_email") != "Remote-Email" ||
		   viper.IsSet("authelia.header_name") && viper.GetString("authelia.header_name") != "Remote-Name"
}

// hasExplicitOIDCConfig checks if user has explicitly configured OIDC
func hasExplicitOIDCConfig() bool {
	// OIDC requires explicit configuration - no defaults that would work
	return viper.IsSet("oidc.issuer") || viper.IsSet("oidc.client_id") || viper.IsSet("oidc.client_secret")
}

// ValidateAutheliaConfiguration validates Authelia provider configuration.
func ValidateAutheliaConfiguration(ctx context.Context) error {
	// Check both new and legacy configuration formats for backward compatibility
	headerUsername := viper.GetString("authelia.header_username")
	if headerUsername == "" {
		headerUsername = viper.GetString("headers_username")
	}
	// Use default if still empty (for backward compatibility with tests)
	if headerUsername == "" {
		headerUsername = "Remote-User"
	}

	headerGroups := viper.GetString("authelia.header_groups")
	if headerGroups == "" {
		headerGroups = viper.GetString("headers_groups")
	}
	// Use default if still empty (for backward compatibility with tests)
	if headerGroups == "" {
		headerGroups = "Remote-Groups"
	}

	return nil
}

// ValidateOIDCConfiguration validates OIDC provider configuration.
func ValidateOIDCConfiguration(ctx context.Context) error {
	required := map[string]string{
		"oidc.issuer":        "OIDC issuer",
		"oidc.client_id":     "OIDC client ID",
		"oidc.client_secret": "OIDC client secret",
	}

	for key, description := range required {
		if len(viper.GetString(key)) == 0 {
			return fmt.Errorf("oidc provider requires %s (set %s)", description, key)
		}
	}

	// Validate client authentication method
	clientAuthMethod := viper.GetString("oidc.client_auth_method")
	if clientAuthMethod != "" {
		validAuthMethods := []string{"client_secret_basic", "client_secret_post"}
		validAuthMethod := false
		for _, method := range validAuthMethods {
			if clientAuthMethod == method {
				validAuthMethod = true
				break
			}
		}
		if !validAuthMethod {
			return fmt.Errorf("invalid oidc.client_auth_method: %s (must be one of: client_secret_basic, client_secret_post)", clientAuthMethod)
		}
	}

	// Validate token validation method
	tokenValidation := viper.GetString("oidc.token_validation")
	if tokenValidation != "" {
		validTokenValidations := []string{"jwks", "userinfo", "both"}
		validTokenValidation := false
		for _, validation := range validTokenValidations {
			if tokenValidation == validation {
				validTokenValidation = true
				break
			}
		}
		if !validTokenValidation {
			return fmt.Errorf("invalid oidc.token_validation: %s (must be one of: jwks, userinfo, both)", tokenValidation)
		}
	}

	// Validate scopes (must be a slice)
	scopes := viper.GetStringSlice("oidc.scopes")
	if len(scopes) == 0 {
		return fmt.Errorf("oidc.scopes must contain at least one scope (typically 'openid')")
	}

	// Validate that required claim mappings are present
	requiredClaimMappings := []string{"username", "email", "groups", "full_name"}
	for _, mapping := range requiredClaimMappings {
		key := fmt.Sprintf("oidc.claim_mappings.%s", mapping)
		if viper.GetString(key) == "" {
			return fmt.Errorf("oidc provider requires claim mapping for %s (set %s)", mapping, key)
		}
	}

	return nil
}

// ValidateCacheConfiguration validates cache configuration for both legacy and cachego formats.
// It ensures exactly zero or one cache type is configured.
func ValidateCacheConfiguration(ctx context.Context) error {
	// Check for both legacy and new configuration
	legacyCacheType := viper.GetString("cache_type")
	newCacheType := viper.GetString("cache.type")
	
	// If both are set to the same value, it's likely due to env var key replacer
	// (ELASTAUTH_CACHE_TYPE maps to both cache_type and cache.type)
	// In this case, treat it as a single configuration
	if legacyCacheType != "" && newCacheType != "" && legacyCacheType == newCacheType {
		// This is the same configuration, not a conflict
		// Use the legacy format for validation
		newCacheType = ""
	}
	
	// Count configured cache types
	configuredCacheTypes := 0
	var activeCacheType string
	
	// Check legacy configuration
	if legacyCacheType != "" {
		configuredCacheTypes++
		activeCacheType = legacyCacheType
	}
	
	// Check new configuration
	if newCacheType != "" {
		configuredCacheTypes++
		activeCacheType = newCacheType
	}
	
	// Validate exactly zero or one cache type is configured
	if configuredCacheTypes > 1 {
		return fmt.Errorf("multiple cache types configured: found both legacy (%s) and new (%s) cache configuration. Please use only one format", legacyCacheType, newCacheType)
	}
	
	// If no cache is configured, that's valid (cache-disabled scenario)
	if configuredCacheTypes == 0 {
		slog.DebugContext(ctx, "No cache configured - running in cache-disabled mode")
		return nil
	}
	
	// Validate the active cache type
	validCacheTypes := []string{"memory", "redis", "file"}
	validCacheType := false
	for _, validType := range validCacheTypes {
		if activeCacheType == validType {
			validCacheType = true
			break
		}
	}
	if !validCacheType {
		return fmt.Errorf("invalid cache type: %s (must be one of: memory, redis, file, or empty for no caching)", activeCacheType)
	}
	
	// Validate cache-specific configuration based on active type
	switch activeCacheType {
	case "redis":
		return ValidateRedisCacheConfiguration(ctx)
	case "file":
		return ValidateFileCacheConfiguration(ctx)
	case "memory":
		// Memory cache doesn't need additional validation
		return ValidateMemoryCacheConfiguration(ctx)
	}
	
	return nil
}

// ValidateMemoryCacheConfiguration validates memory cache configuration
func ValidateMemoryCacheConfiguration(ctx context.Context) error {
	// Memory cache has horizontal scaling constraints
	slog.WarnContext(ctx, "Memory cache is configured - this limits deployment to single instance only for consistency")
	
	// Check if expiration is set (optional for memory cache)
	expiration := viper.GetString("cache.expiration")
	if expiration != "" {
		// Validate expiration format
		if _, err := time.ParseDuration(expiration); err != nil {
			return fmt.Errorf("invalid cache expiration format: %s (must be a valid duration like '1h', '30m', '300s')", expiration)
		}
	}
	
	return nil
}

// ValidateRedisCacheConfiguration validates Redis cache configuration.
func ValidateRedisCacheConfiguration(ctx context.Context) error {
	// Support both legacy and new configuration formats
	var redisHost string
	var redisDB int
	var expiration string
	
	// Check new format first
	newRedisHost := viper.GetString("cache.redis_host")
	if newRedisHost != "" {
		redisHost = newRedisHost
		redisDB = viper.GetInt("cache.redis_db")
		expiration = viper.GetString("cache.expiration")
	} else {
		// Fall back to legacy format
		redisHost = viper.GetString("redis_host")
		redisDB = viper.GetInt("redis_db")
		expiration = viper.GetString("cache_expire")
	}
	
	if redisHost == "" {
		return fmt.Errorf("redis cache requires redis host configuration (cache.redis_host or redis_host)")
	}
	
	// Validate Redis DB number
	if redisDB < 0 || redisDB > 15 {
		return fmt.Errorf("invalid redis database number: %d (must be between 0 and 15)", redisDB)
	}
	
	// Validate expiration format if set
	if expiration != "" {
		if _, err := time.ParseDuration(expiration); err != nil {
			return fmt.Errorf("invalid cache expiration format: %s (must be a valid duration like '1h', '30m', '300s')", expiration)
		}
	}
	
	// Redis cache supports horizontal scaling
	slog.InfoContext(ctx, "Redis cache configured - supports horizontal scaling with shared Redis instance")
	
	return nil
}

// ValidateFileCacheConfiguration validates file cache configuration.
func ValidateFileCacheConfiguration(ctx context.Context) error {
	cachePath := viper.GetString("cache.path")
	if cachePath == "" {
		return fmt.Errorf("file cache requires path configuration (set cache.path)")
	}
	
	// Validate that the cache directory is writable
	dir := filepath.Dir(cachePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("cannot create cache directory %s: %w", dir, err)
	}
	
	// Test write permissions
	testFile := filepath.Join(dir, ".elastauth-cache-test")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		return fmt.Errorf("cache directory %s is not writable: %w", dir, err)
	}
	os.Remove(testFile) // Clean up test file
	
	// Validate expiration format if set
	expiration := viper.GetString("cache.expiration")
	if expiration != "" {
		if _, err := time.ParseDuration(expiration); err != nil {
			return fmt.Errorf("invalid cache expiration format: %s (must be a valid duration like '1h', '30m', '300s')", expiration)
		}
	}
	
	// File cache has horizontal scaling constraints
	slog.WarnContext(ctx, "File cache configured - this limits deployment to single instance only for consistency")
	
	return nil
}

// GetEffectiveCacheConfig returns the effective cache configuration, handling both legacy and new formats.
func GetEffectiveCacheConfig() map[string]interface{} {
	config := make(map[string]interface{})
	
	// Determine which configuration format is being used
	legacyCacheType := viper.GetString("cache_type")
	newCacheType := viper.GetString("cache.type")
	
	if legacyCacheType != "" {
		// Use legacy configuration format
		config["type"] = legacyCacheType
		config["expiration"] = viper.GetString("cache_expire")
		config["redis_host"] = viper.GetString("redis_host")
		config["redis_db"] = viper.GetInt("redis_db")
		config["format"] = "legacy"
	} else if newCacheType != "" {
		// Use new cachego configuration format
		config["type"] = newCacheType
		config["expiration"] = viper.GetString("cache.expiration")
		config["redis_host"] = viper.GetString("cache.redis_host")
		config["redis_db"] = viper.GetInt("cache.redis_db")
		config["path"] = viper.GetString("cache.path")
		config["format"] = "new"
	} else {
		// No cache configured
		config["type"] = "disabled"
		config["format"] = "none"
	}
	
	return config
}

// ValidateHorizontalScalingConstraints validates cache configuration for horizontal scaling
func ValidateHorizontalScalingConstraints(ctx context.Context) error {
	cacheConfig := GetEffectiveCacheConfig()
	cacheType := cacheConfig["type"].(string)
	
	switch cacheType {
	case "memory", "file":
		slog.WarnContext(ctx, "Cache type limits horizontal scaling", 
			slog.String("cache_type", cacheType),
			slog.String("constraint", "single instance only"))
		return nil
	case "redis":
		slog.InfoContext(ctx, "Cache type supports horizontal scaling", 
			slog.String("cache_type", cacheType),
			slog.String("requirement", "shared Redis instance across all instances"))
		return nil
	case "disabled":
		slog.InfoContext(ctx, "No cache configured - supports horizontal scaling with independent instances")
		return nil
	default:
		return fmt.Errorf("unknown cache type for horizontal scaling validation: %s", cacheType)
	}
}

// GetEffectiveAutheliaConfig returns the effective Authelia configuration, handling both legacy and new formats.
func GetEffectiveAutheliaConfig() map[string]interface{} {
	config := make(map[string]interface{})
	
	// Get header configuration (prefer new format)
	headerUsername := viper.GetString("authelia.header_username")
	if headerUsername == "" {
		headerUsername = viper.GetString("headers_username")
	}
	config["header_username"] = headerUsername
	
	headerGroups := viper.GetString("authelia.header_groups")
	if headerGroups == "" {
		headerGroups = viper.GetString("headers_groups")
	}
	config["header_groups"] = headerGroups
	
	headerEmail := viper.GetString("authelia.header_email")
	if headerEmail == "" {
		headerEmail = viper.GetString("headers_email")
	}
	config["header_email"] = headerEmail
	
	headerName := viper.GetString("authelia.header_name")
	if headerName == "" {
		headerName = viper.GetString("headers_name")
	}
	config["header_name"] = headerName
	
	return config
}

// GetEffectiveOIDCConfig returns the effective OIDC configuration with all settings.
func GetEffectiveOIDCConfig() map[string]interface{} {
	config := make(map[string]interface{})
	
	// Basic OIDC settings
	config["issuer"] = viper.GetString("oidc.issuer")
	config["client_id"] = viper.GetString("oidc.client_id")
	config["client_secret"] = viper.GetString("oidc.client_secret")
	
	// Optional manual endpoint configuration
	config["authorization_endpoint"] = viper.GetString("oidc.authorization_endpoint")
	config["token_endpoint"] = viper.GetString("oidc.token_endpoint")
	config["userinfo_endpoint"] = viper.GetString("oidc.userinfo_endpoint")
	config["jwks_uri"] = viper.GetString("oidc.jwks_uri")
	
	// OAuth2 settings
	config["scopes"] = viper.GetStringSlice("oidc.scopes")
	config["client_auth_method"] = viper.GetString("oidc.client_auth_method")
	config["token_validation"] = viper.GetString("oidc.token_validation")
	config["use_pkce"] = viper.GetBool("oidc.use_pkce")
	
	// Claim mappings
	claimMappings := make(map[string]string)
	claimMappings["username"] = viper.GetString("oidc.claim_mappings.username")
	claimMappings["email"] = viper.GetString("oidc.claim_mappings.email")
	claimMappings["groups"] = viper.GetString("oidc.claim_mappings.groups")
	claimMappings["full_name"] = viper.GetString("oidc.claim_mappings.full_name")
	config["claim_mappings"] = claimMappings
	
	// Custom headers
	customHeaders := make(map[string]string)
	for key, value := range viper.GetStringMapString("oidc.custom_headers") {
		customHeaders[key] = value
	}
	config["custom_headers"] = customHeaders
	
	return config
}

// GetEffectiveProviderConfig returns the effective configuration for the currently selected provider.
func GetEffectiveProviderConfig() map[string]interface{} {
	authProvider := viper.GetString("auth_provider")
	if authProvider == "" {
		authProvider = "authelia" // Default for backward compatibility
	}
	
	switch authProvider {
	case "authelia":
		return GetEffectiveAutheliaConfig()
	case "oidc":
		return GetEffectiveOIDCConfig()
	default:
		return make(map[string]interface{})
	}
}

// ValidateElasticsearchConfiguration validates Elasticsearch configuration
func ValidateElasticsearchConfiguration() error {
	hosts := GetElasticsearchHosts()
	if len(hosts) == 0 {
		return fmt.Errorf("no Elasticsearch hosts configured - set elasticsearch.hosts or elasticsearch_host")
	}

	username := GetElasticsearchUsername()
	if username == "" {
		return fmt.Errorf("Elasticsearch username not configured - set elasticsearch.username or elasticsearch_username")
	}

	password := GetElasticsearchPassword()
	if password == "" {
		return fmt.Errorf("Elasticsearch password not configured - set elasticsearch.password or elasticsearch_password")
	}

	// Validate that all hosts use the same protocol (http or https)
	var protocol string
	for i, host := range hosts {
		if host == "" {
			return fmt.Errorf("empty Elasticsearch host at index %d", i)
		}

		if i == 0 {
			if len(host) > 8 && host[:8] == "https://" {
				protocol = "https"
			} else if len(host) > 7 && host[:7] == "http://" {
				protocol = "http"
			} else {
				return fmt.Errorf("Elasticsearch host %s must start with http:// or https://", host)
			}
		} else {
			expectedPrefix := protocol + "://"
			if len(host) < len(expectedPrefix) || host[:len(expectedPrefix)] != expectedPrefix {
				return fmt.Errorf("all Elasticsearch hosts must use the same protocol (%s), but host %s uses different protocol", protocol, host)
			}
		}
	}

	return nil
}

// ValidateProxyConfiguration validates proxy mode configuration
func ValidateProxyConfiguration() error {
	if !viper.GetBool("proxy.enabled") {
		// Proxy mode is disabled, no validation needed
		return nil
	}

	// Validate Elasticsearch URL is provided
	elasticsearchURL := viper.GetString("proxy.elasticsearch_url")
	if elasticsearchURL == "" {
		return fmt.Errorf("proxy.elasticsearch_url is required when proxy mode is enabled")
	}

	// Validate URL format
	if !strings.HasPrefix(elasticsearchURL, "http://") && !strings.HasPrefix(elasticsearchURL, "https://") {
		return fmt.Errorf("proxy.elasticsearch_url must start with http:// or https://, got: %s", elasticsearchURL)
	}

	// Validate timeout
	timeoutStr := viper.GetString("proxy.timeout")
	if _, err := time.ParseDuration(timeoutStr); err != nil {
		return fmt.Errorf("invalid proxy.timeout value '%s': %w", timeoutStr, err)
	}

	// Validate idle connection timeout
	idleTimeoutStr := viper.GetString("proxy.idle_conn_timeout")
	if _, err := time.ParseDuration(idleTimeoutStr); err != nil {
		return fmt.Errorf("invalid proxy.idle_conn_timeout value '%s': %w", idleTimeoutStr, err)
	}

	// Validate max idle connections
	maxIdleConns := viper.GetInt("proxy.max_idle_conns")
	if maxIdleConns < 1 {
		return fmt.Errorf("proxy.max_idle_conns must be at least 1, got: %d", maxIdleConns)
	}

	// Validate TLS configuration if enabled
	if viper.GetBool("proxy.tls.enabled") {
		caCert := viper.GetString("proxy.tls.ca_cert")
		clientCert := viper.GetString("proxy.tls.client_cert")
		clientKey := viper.GetString("proxy.tls.client_key")

		// If client cert is provided, client key must also be provided
		if clientCert != "" && clientKey == "" {
			return fmt.Errorf("proxy.tls.client_key is required when proxy.tls.client_cert is provided")
		}

		// If client key is provided, client cert must also be provided
		if clientKey != "" && clientCert == "" {
			return fmt.Errorf("proxy.tls.client_cert is required when proxy.tls.client_key is provided")
		}

		// Validate that certificate files exist if provided
		if caCert != "" {
			if _, err := os.Stat(caCert); os.IsNotExist(err) {
				return fmt.Errorf("proxy.tls.ca_cert file does not exist: %s", caCert)
			}
		}

		if clientCert != "" {
			if _, err := os.Stat(clientCert); os.IsNotExist(err) {
				return fmt.Errorf("proxy.tls.client_cert file does not exist: %s", clientCert)
			}
		}

		if clientKey != "" {
			if _, err := os.Stat(clientKey); os.IsNotExist(err) {
				return fmt.Errorf("proxy.tls.client_key file does not exist: %s", clientKey)
			}
		}
	}

	return nil
}

// GetEffectiveProxyConfig returns the effective proxy configuration
func GetEffectiveProxyConfig() map[string]interface{} {
	config := make(map[string]interface{})
	
	config["enabled"] = viper.GetBool("proxy.enabled")
	config["elasticsearch_url"] = viper.GetString("proxy.elasticsearch_url")
	config["timeout"] = viper.GetString("proxy.timeout")
	config["max_idle_conns"] = viper.GetInt("proxy.max_idle_conns")
	config["idle_conn_timeout"] = viper.GetString("proxy.idle_conn_timeout")
	
	// TLS configuration
	tlsConfig := make(map[string]interface{})
	tlsConfig["enabled"] = viper.GetBool("proxy.tls.enabled")
	tlsConfig["insecure_skip_verify"] = viper.GetBool("proxy.tls.insecure_skip_verify")
	tlsConfig["ca_cert"] = viper.GetString("proxy.tls.ca_cert")
	tlsConfig["client_cert"] = viper.GetString("proxy.tls.client_cert")
	tlsConfig["client_key"] = viper.GetString("proxy.tls.client_key")
	
	config["tls"] = tlsConfig
	
	return config
}

// BuildProxyConfig creates a ProxyConfig struct from viper configuration
func BuildProxyConfig() (*ProxyConfig, error) {
	if !viper.GetBool("proxy.enabled") {
		return nil, nil
	}

	// Parse timeout
	timeoutStr := viper.GetString("proxy.timeout")
	timeout, err := time.ParseDuration(timeoutStr)
	if err != nil {
		return nil, fmt.Errorf("invalid proxy.timeout: %w", err)
	}

	// Parse idle connection timeout
	idleTimeoutStr := viper.GetString("proxy.idle_conn_timeout")
	idleTimeout, err := time.ParseDuration(idleTimeoutStr)
	if err != nil {
		return nil, fmt.Errorf("invalid proxy.idle_conn_timeout: %w", err)
	}

	// Build TLS configuration
	tlsConfig := TLSConfig{
		Enabled:            viper.GetBool("proxy.tls.enabled"),
		InsecureSkipVerify: viper.GetBool("proxy.tls.insecure_skip_verify"),
		CACert:             viper.GetString("proxy.tls.ca_cert"),
		ClientCert:         viper.GetString("proxy.tls.client_cert"),
		ClientKey:          viper.GetString("proxy.tls.client_key"),
	}

	// Build proxy configuration
	config := &ProxyConfig{
		Enabled:          true,
		ElasticsearchURL: viper.GetString("proxy.elasticsearch_url"),
		Timeout:          timeout,
		MaxIdleConns:     viper.GetInt("proxy.max_idle_conns"),
		IdleConnTimeout:  idleTimeout,
		TLS:              tlsConfig,
	}

	return config, nil
}
