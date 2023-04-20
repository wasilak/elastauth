package main

import (
	"time"

	// "github.com/labstack/echo-contrib/prometheus"
	_ "net/http/pprof"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
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

	cacheDuration, err := time.ParseDuration(viper.GetString("cache_expire"))
	if err != nil {
		log.Fatal(err)
	}

	if viper.GetString("cache_type") == "redis" {
		cache.CacheInstance = &cache.RedisCache{
			Address: viper.GetString("redis_host"),
			DB:      viper.GetInt("redis_db"),
			TTL:     cacheDuration,
		}
	} else if viper.GetString("cache_type") == "memory" {
		cache.CacheInstance = &cache.GoCache{
			TTL: cacheDuration,
		}
	} else {
		log.Fatal("No cache_type selected or cache type is invalid")
	}

	cache.CacheInstance.Init(cacheDuration)

	e := echo.New()

	e.HideBanner = true

	e.Debug = viper.GetBool("debug")

	e.Use(middleware.Logger())

	// e.Use(middleware.GzipWithConfig(middleware.GzipConfig{
	// 	Skipper: func(c echo.Context) bool {
	// 		return strings.Contains(c.Path(), "metrics")
	// 	},
	// }))

	// // Enable metrics middleware
	// p := prometheus.NewPrometheus("echo", nil)
	// p.Use(e)

	e.GET("/", libs.MainRoute)
	e.GET("/health", libs.HealthRoute)
	e.GET("/config", libs.ConfigRoute)

	e.Logger.Fatal(e.Start(viper.GetString("listen")))
}
