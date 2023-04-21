package main

import (
	"github.com/spf13/viper"
	"github.com/wasilak/elastauth/cache"
	"github.com/wasilak/elastauth/libs"
	"github.com/wasilak/elastauth/logger"
	_ "go.uber.org/automaxprocs"
	"golang.org/x/exp/slog"
)

func main() {
	viper.SetDefault("log_file", "./elastauth.log")
	logger.LoggerInit()

	err := libs.InitConfiguration()
	if err != nil {
		panic(err)
	}

	err = libs.HandleSecretKey()
	if err != nil {
		panic(err)
	}

	logger.LoggerInstance.Debug("logger", slog.Any("setings", viper.AllSettings()))

	cache.CacheInit()

	libs.WebserverInit()
}
