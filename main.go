package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	// "github.com/labstack/echo-contrib/prometheus"
	_ "net/http/pprof"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/wasilak/elastauth/cache"
	"github.com/wasilak/elastauth/libs"
	_ "go.uber.org/automaxprocs"
)

func main() {
	flag.Bool("debug", false, "Debug")
	flag.Bool("generateKey", false, "Generate valid encryption key for use in app")
	flag.String("listen", "127.0.0.1:5000", "Listen address")
	flag.String("config", "./", "Path to config.yml")

	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()
	viper.BindPFlags(pflag.CommandLine)

	viper.SetEnvPrefix("elastauth")
	viper.AutomaticEnv()

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(viper.GetString("config"))

	viper.SetDefault("cache_type", "memory")
	viper.SetDefault("redis_host", "localhost:6379")
	viper.SetDefault("redis_db", 0)
	viper.SetDefault("cache_expire", "1h")
	viper.SetDefault("elasticsearch_dry_run", false)

	viper.SetDefault("headers_username", "Remote-User")
	viper.SetDefault("headers_groups", "Remote-Groups")
	viper.SetDefault("headers_Email", "Remote-Email")
	viper.SetDefault("headers_name", "Remote-Name")

	viperErr := viper.ReadInConfig()

	if viperErr != nil {
		log.Fatal(viperErr)
		panic(viperErr)
	}

	if viper.GetBool("debug") {
		log.SetLevel(log.DEBUG)
	}

	if viper.GetBool("generateKey") {
		key := libs.GenerateKey()
		fmt.Println(key)
		os.Exit(0)
	}

	if len(viper.GetString("secret_key")) == 0 {
		key := libs.GenerateKey()
		viper.Set("secret_key", key)
		log.Info(fmt.Sprintf("WARNING: No secret key provided. Setting randomly generated: %s", key))
	}

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
