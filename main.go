package main

import (
	"github.com/spf13/viper"
	"github.com/wasilak/elastauth/cache"
	"github.com/wasilak/elastauth/libs"
	"github.com/wasilak/elastauth/logger"
	"golang.org/x/exp/slog"
)

// The main function initializes configuration, logger, secret key, cache, and web server for a Go
// application.
func main() {

	err := libs.InitConfiguration()
	if err != nil {
		panic(err)
	}

	if viper.GetBool("enableOtel") {
		libs.InitTracer()
	}

	logger.LoggerInit(viper.GetString("log_level"), viper.GetString("log_format"))

	err = libs.HandleSecretKey()
	if err != nil {
		panic(err)
	}

	slog.Debug("logger", slog.Any("setings", viper.AllSettings()))

	cache.CacheInit()

	libs.WebserverInit()
}
