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

// InitConfiguration initializes the application configuration from command-line flags,
// environment variables (prefixed with ELASTAUTH_), and a YAML configuration file.
// It sets up viper to read from these sources and establishes default values for various settings.
func InitConfiguration() error {
	flag.Bool("generateKey", false, "Generate valid encryption key for use in app")
	flag.String("listen", "127.0.0.1:5000", "Listen address")
	flag.String("config", "./", "Path to config.yml")
	flag.Bool("enableOtel", false, "Enable OTEL (OpenTelemetry)")

	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()
	viper.BindPFlags(pflag.CommandLine)

	viper.SetEnvPrefix("elastauth")
	viper.AutomaticEnv()

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(viper.GetString("config"))

	// Provider configuration defaults
	viper.SetDefault("auth_provider", "authelia")

	// New cachego configuration defaults
	viper.SetDefault("cache.type", "")
	viper.SetDefault("cache.expiration", "1h")
	viper.SetDefault("cache.redis_host", "localhost:6379")
	viper.SetDefault("cache.redis_db", 0)
	viper.SetDefault("cache.path", "/tmp/elastauth-cache")

	viper.SetDefault("elasticsearch_dry_run", false)

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

	viper.SetDefault("enable_metrics", false)

	viper.SetDefault("enableOtel", false)

	viper.SetDefault("log_level", "info")
	viper.SetDefault("log_format", "text")

	err := viper.ReadInConfig()
	if err != nil {
		log.Println(err)
	}

	return nil
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
