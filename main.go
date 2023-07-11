package main

import (
	"context"

	"github.com/spf13/viper"
	"github.com/wasilak/elastauth/cache"
	"github.com/wasilak/elastauth/libs"
	otelgotracer "github.com/wasilak/otelgo/tracing"
	"golang.org/x/exp/slog"

	"github.com/wasilak/loggergo"
)

// The main function initializes configuration, logger, secret key, cache, and web server for a Go
// application.
func main() {

	// Create a new background context.
	ctx := context.Background()

	// Initialize the configuration for the application.
	err := libs.InitConfiguration()
	if err != nil {
		panic(err)
	}

	// Check if OpenTelemetry tracing is enabled.
	if viper.GetBool("enableOtel") {
		// Initialize the OpenTelemetry tracer.
		otelgotracer.InitTracer(ctx, true)
	}

	// Initialize the logger with the specified log level and format from the configuration.
	loggergo.LoggerInit(viper.GetString("log_level"), viper.GetString("log_format"))

	// Handle the secret key using the HandleSecretKey function.
	err = libs.HandleSecretKey(ctx)
	if err != nil {
		panic(err)
	}

	// Log the application settings using the debug log level.
	slog.DebugCtx(ctx, "logger", slog.Any("settings", viper.AllSettings()))

	// Initialize the cache.
	cache.CacheInit(ctx)

	// Initialize the web server.
	libs.WebserverInit(ctx)

}
