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

// This function initializes the configuration for an application using flags, environment variables,
// and a YAML configuration file.
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

	viper.SetDefault("cache_type", "memory")
	viper.SetDefault("redis_host", "localhost:6379")
	viper.SetDefault("redis_db", 0)
	viper.SetDefault("cache_expire", "1h")
	viper.SetDefault("elasticsearch_dry_run", false)

	viper.SetDefault("headers_username", "Remote-User")
	viper.SetDefault("headers_groups", "Remote-Groups")
	viper.SetDefault("headers_Email", "Remote-Email")
	viper.SetDefault("headers_name", "Remote-Name")

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

// The function generates and sets a secret key if one is not provided or generates and prints a secret
// key if the "generateKey" flag is set to true.
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

func ValidateConfiguration(ctx context.Context) error {
	if err := ValidateRequiredConfig(ctx); err != nil {
		return err
	}

	if err := ValidateSecretKey(viper.GetString("secret_key")); err != nil {
		return err
	}

	cacheType := viper.GetString("cache_type")
	if cacheType == "redis" {
		if len(viper.GetString("redis_host")) == 0 {
			return fmt.Errorf("redis_host is required when cache_type is 'redis' (set via ELASTAUTH_REDIS_HOST)")
		}
	} else if cacheType != "memory" {
		return fmt.Errorf("invalid cache_type: %s (must be 'memory' or 'redis')", cacheType)
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
