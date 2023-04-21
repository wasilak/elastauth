package main

import (
	"github.com/labstack/gommon/log"
	"github.com/spf13/viper"
	"github.com/wasilak/elastauth/cache"
	"github.com/wasilak/elastauth/libs"
	_ "go.uber.org/automaxprocs"
)

func main() {
	err := libs.InitConfiguration()
	if err != nil {
		log.Fatal(err)
	}

	libs.HandleSecretKey()

	log.Debug(viper.AllSettings())

	cache.CacheInit()

	libs.WebserverInit()
}
