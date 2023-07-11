package libs

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"go.opentelemetry.io/otel"
	"golang.org/x/exp/slog"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var tracerConfig = otel.Tracer("config")

// This function initializes the configuration for an application using flags, environment variables,
// and a YAML configuration file.
func InitConfiguration() error {
	// Define command line flags with their default values and descriptions.
	flag.Bool("generateKey", false, "Generate valid encryption key for use in app")
	flag.String("listen", "127.0.0.1:5000", "Listen address")
	flag.String("config", "./", "Path to config.yml")
	flag.Bool("enableOtel", false, "Enable OTEL (OpenTelemetry)")

	// Add the defined flags to the pflag.CommandLine flag set and parse the command line arguments.
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()

	// Bind the pflags.CommandLine to Viper configuration, so that Viper can automatically read values from command line flags.
	viper.BindPFlags(pflag.CommandLine)

	// Set the environment variable prefix for Viper.
	viper.SetEnvPrefix("elastauth")

	// Automatically read values from environment variables if they have the "ELASTAUTH_" prefix.
	viper.AutomaticEnv()

	// Set the name and type of the configuration file to be used by Viper.
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	// Add the path defined in the "config" flag as a search path for the configuration file.
	viper.AddConfigPath(viper.GetString("config"))

	// Set default values for certain configuration keys.
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

	// Read the configuration file specified by the search paths and unmarshal its contents into Viper.
	err := viper.ReadInConfig()
	if err != nil {
		log.Println(err)
	}

	return nil
}

// The function generates and sets a secret key if one is not provided or generates and prints a secret
// key if the "generateKey" flag is set to true.
func HandleSecretKey(ctx context.Context) error {
	// Start a new span named "HandleSecretKey" using tracerConfig and context.
	_, span := tracerConfig.Start(ctx, "HandleSecretKey")

	// Defer the end of the span to ensure it is ended when the function returns.
	defer span.End()

	// Check if the "generateKey" flag is set to true in the Viper configuration.
	if viper.GetBool("generateKey") {
		// If it is true, generate a new key using the GenerateKey function.
		key, err := GenerateKey(ctx)
		if err != nil {
			panic(err)
		}

		// Print the generated key.
		fmt.Println(key)

		// Exit the program with code 0 (success).
		os.Exit(0)
	}

	// Check if the "secret_key" value in the Viper configuration is empty.
	if len(viper.GetString("secret_key")) == 0 {
		// If it is empty, generate a new key using the GenerateKey function.
		key, err := GenerateKey(ctx)
		if err != nil {
			return err
		}

		// Set the generated key as the value for "secret_key" in the Viper configuration.
		viper.Set("secret_key", key)

		// Log a warning message indicating that no secret key was provided and a randomly generated key is being set.
		slog.InfoCtx(ctx, "WARNING: No secret key provided. Setting randomly generated", slog.String("key", key))
	}

	// Return nil to indicate success.
	return nil
}
