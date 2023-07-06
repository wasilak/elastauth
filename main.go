package main

import (
	"context"

	"github.com/spf13/viper"
	"github.com/wasilak/elastauth/cache"
	"github.com/wasilak/elastauth/libs"
	"github.com/wasilak/elastauth/logger"
	otelgotracer "github.com/wasilak/otelgo/tracing"
	"golang.org/x/exp/slog"
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
		otelgotracer.InitTracer(ctx)
	}

	logger.LoggerInit(viper.GetString("log_level"), viper.GetString("log_format"))

	err = libs.HandleSecretKey(ctx)
	if err != nil {
		panic(err)
	}

	slog.DebugCtx(ctx, "logger", slog.Any("setings", viper.AllSettings()))

	cache.CacheInit(ctx)

	libs.WebserverInit(ctx)
}
