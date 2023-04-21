package main

import (
	"github.com/spf13/viper"
	"github.com/wasilak/elastauth/cache"
	"github.com/wasilak/elastauth/libs"
	"github.com/wasilak/elastauth/logger"
	"golang.org/x/exp/slog"
)

func main() {

	err := libs.InitConfiguration()
	if err != nil {
		panic(err)
	}

	logger.LoggerInit()

	err = libs.HandleSecretKey()
	if err != nil {
		panic(err)
	}

	logger.LoggerInstance.Debug("logger", slog.Any("setings", viper.AllSettings()))

	cache.CacheInit()

	libs.WebserverInit()
}
