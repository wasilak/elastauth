package main

import (
	"context"
	"os"

	"log/slog"

	"github.com/spf13/viper"
	"github.com/wasilak/elastauth/cache"
	"github.com/wasilak/elastauth/libs"
	"github.com/wasilak/elastauth/logger"
	otelgotracer "github.com/wasilak/otelgo/tracing"
)

// The main function initializes configuration, logger, secret key, cache, and web server for a Go
// application.
func main() {

	ctx := context.Background()

	err := libs.InitConfiguration()
	if err != nil {
		slog.ErrorContext(ctx, "Failed to initialize configuration", slog.Any("error", err))
		os.Exit(1)
	}

	if viper.GetBool("enableOtel") {
		otelGoTracingConfig := otelgotracer.Config{
			HostMetricsEnabled:    viper.GetBool("enableOtelHostMetrics"),
			RuntimeMetricsEnabled: viper.GetBool("enableOtelRuntimeMetrics"),
		}
		_, _, err := otelgotracer.Init(ctx, otelGoTracingConfig)
		if err != nil {
			slog.ErrorContext(ctx, err.Error())
			os.Exit(1)
		}
	}

	logger.LoggerInit(ctx, viper.GetString("log_level"), viper.GetString("log_format"))

	err = libs.HandleSecretKey(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to handle secret key", slog.Any("error", err))
		os.Exit(1)
	}

	err = libs.ValidateConfiguration(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "Configuration validation failed", slog.Any("error", err))
		os.Exit(1)
	}

	sanitizedSettings := libs.SanitizeForLogging(viper.AllSettings())
	slog.InfoContext(ctx, "Configuration loaded successfully", slog.Any("settings", sanitizedSettings))

	cache.CacheInit(ctx)

	libs.WebserverInit(ctx)
}
