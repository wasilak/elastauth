package libs

import (
	"context"
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"log/slog"

	"go.opentelemetry.io/otel"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var tracerConfig = otel.Tracer("config")

var LogLeveler *slog.LevelVar

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

	// Casdoor provider defaults
	viper.SetDefault("casdoor.organization", "built-in")
	viper.SetDefault("casdoor.application", "app-built-in")

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
	
	// Casdoor provider configuration
	viper.BindEnv("casdoor.endpoint", "ELASTAUTH_CASDOOR_ENDPOINT")
	viper.BindEnv("casdoor.client_id", "ELASTAUTH_CASDOOR_CLIENT_ID")
	viper.BindEnv("casdoor.client_secret", "ELASTAUTH_CASDOOR_CLIENT_SECRET")
	viper.BindEnv("casdoor.organization", "ELASTAUTH_CASDOOR_ORGANIZATION")
	viper.BindEnv("casdoor.application", "ELASTAUTH_CASDOOR_APPLICATION")
	
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
		
		// Casdoor provider configuration
		"ELASTAUTH_CASDOOR_ENDPOINT",
		"ELASTAUTH_CASDOOR_CLIENT_ID",
		"ELASTAUTH_CASDOOR_CLIENT_SECRET",
		"ELASTAUTH_CASDOOR_ORGANIZATION",
		"ELASTAUTH_CASDOOR_APPLICATION",
		
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
	validProviders := []string{"authelia", "casdoor", "oidc"}
	validProvider := false
	for _, provider := range validProviders {
		if authProvider == provider {
			validProvider = true
			break
		}
	}
	if !validProvider {
		return fmt.Errorf("invalid auth_provider: %s (must be one of: authelia, casdoor, oidc)", authProvider)
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
	
	if authProvider != "casdoor" {
		// Check if Casdoor is explicitly configured when another provider is selected
		if hasExplicitCasdoorConfig() {
			explicitProviders = append(explicitProviders, "casdoor")
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
	case "casdoor":
		return ValidateCasdoorConfiguration(ctx)
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

// hasExplicitCasdoorConfig checks if user has explicitly configured Casdoor
func hasExplicitCasdoorConfig() bool {
	// Casdoor requires explicit configuration - no defaults that would work
	return viper.IsSet("casdoor.endpoint") || viper.IsSet("casdoor.client_id") || viper.IsSet("casdoor.client_secret")
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

// ValidateCasdoorConfiguration validates Casdoor provider configuration.
func ValidateCasdoorConfiguration(ctx context.Context) error {
	required := map[string]string{
		"casdoor.endpoint":      "Casdoor endpoint",
		"casdoor.client_id":     "Casdoor client ID",
		"casdoor.client_secret": "Casdoor client secret",
	}

	for key, description := range required {
		if len(viper.GetString(key)) == 0 {
			return fmt.Errorf("casdoor provider requires %s (set %s)", description, key)
		}
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
	// Check for legacy configuration and reject it
	legacyCacheType := viper.GetString("cache_type")
	legacyRedisHost := viper.GetString("redis_host")
	legacyCacheExpire := viper.GetString("cache_expire")
	legacyRedisDB := viper.GetInt("redis_db")
	
	if legacyCacheType != "" || legacyRedisHost != "" || legacyCacheExpire != "" || legacyRedisDB != 0 {
		return fmt.Errorf("legacy cache configuration detected. Please migrate to new format:\n" +
			"  Old: cache_type, redis_host, cache_expire, redis_db\n" +
			"  New: cache.type, cache.redis_host, cache.expiration, cache.redis_db\n" +
			"See documentation for migration guide")
	}
	
	// Only validate new cachego configuration
	cacheType := viper.GetString("cache.type")
	
	// Validate cache type if specified
	if cacheType != "" {
		validCacheTypes := []string{"memory", "redis", "file"}
		validCacheType := false
		for _, validType := range validCacheTypes {
			if cacheType == validType {
				validCacheType = true
				break
			}
		}
		if !validCacheType {
			return fmt.Errorf("invalid cache type: %s (must be one of: memory, redis, file, or empty for no caching)", cacheType)
		}
		
		// Validate cache-specific configuration
		switch cacheType {
		case "redis":
			return ValidateRedisCacheConfiguration(ctx)
		case "file":
			return ValidateFileCacheConfiguration(ctx)
		}
	}
	
	return nil
}

// ValidateRedisCacheConfiguration validates Redis cache configuration.
func ValidateRedisCacheConfiguration(ctx context.Context) error {
	// Only check new configuration format
	redisHost := viper.GetString("cache.redis_host")
	if redisHost == "" {
		return fmt.Errorf("redis cache requires cache.redis_host configuration")
	}
	
	return nil
}

// ValidateFileCacheConfiguration validates file cache configuration.
func ValidateFileCacheConfiguration(ctx context.Context) error {
	cachePath := viper.GetString("cache.path")
	if cachePath == "" {
		return fmt.Errorf("file cache requires path configuration (set cache.path)")
	}
	
	return nil
}

// GetEffectiveCacheConfig returns the effective cache configuration, handling both legacy and new formats.
func GetEffectiveCacheConfig() map[string]interface{} {
	config := make(map[string]interface{})
	
	// Only use new cachego configuration format
	config["type"] = viper.GetString("cache.type")
	config["expiration"] = viper.GetString("cache.expiration")
	config["redis_host"] = viper.GetString("cache.redis_host")
	config["redis_db"] = viper.GetInt("cache.redis_db")
	config["path"] = viper.GetString("cache.path")
	
	return config
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

// GetEffectiveCasdoorConfig returns the effective Casdoor configuration with all settings.
func GetEffectiveCasdoorConfig() map[string]interface{} {
	config := make(map[string]interface{})
	
	// Basic Casdoor settings
	config["endpoint"] = viper.GetString("casdoor.endpoint")
	config["client_id"] = viper.GetString("casdoor.client_id")
	config["client_secret"] = viper.GetString("casdoor.client_secret")
	
	// Optional Casdoor settings
	config["organization"] = viper.GetString("casdoor.organization")
	config["application"] = viper.GetString("casdoor.application")
	
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
	case "casdoor":
		return GetEffectiveCasdoorConfig()
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
