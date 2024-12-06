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
		panic(err)
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
		panic(err)
	}

	slog.DebugContext(ctx, "logger", slog.Any("setings", viper.AllSettings()))

	cache.CacheInit(ctx)

	libs.WebserverInit(ctx)
}
