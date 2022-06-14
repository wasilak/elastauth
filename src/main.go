package main

import (
	"crypto/rand"
	"encoding/hex"
	"flag"
	"fmt"
	"os"
	// "github.com/labstack/echo-contrib/prometheus"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/wasilak/elastauth/libs"
	"net/http"
	_ "net/http/pprof"
	// "strings"
)

func main() {
	go func() {
		log.Debug(http.ListenAndServe("localhost:6060", nil))
	}()

	// using standard library "flag" package
	flag.Bool("debug", false, "Debug")
	flag.Bool("generateKey", false, "Generate valid encryption key for use in app")
	flag.String("listen", "127.0.0.1:5000", "Listen address")
	flag.String("config", "./", "Path to config.yml")

	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()
	viper.BindPFlags(pflag.CommandLine)

	viper.SetEnvPrefix("elastauth")
	viper.AutomaticEnv()

	viper.SetConfigName("config")                  // name of config file (without extension)
	viper.SetConfigType("yaml")                    // REQUIRED if the config file does not have the extension in the name
	viper.AddConfigPath(viper.GetString("config")) // path to look for the config file in
	viperErr := viper.ReadInConfig()               // Find and read the config file

	if viperErr != nil { // Handle errors reading the config file
		log.Fatal(viperErr)
		panic(viperErr)
	}

	if len(viper.GetString("secret_key")) == 0 {
		log.Fatal("Secret key for password encryption not provided")
	}

	if viper.GetBool("debug") {
		log.SetLevel(log.DEBUG)
	}

	log.Debug(viper.AllSettings())

	if viper.GetBool("generateKey") {
		bytes := make([]byte, 32) //generate a random 32 byte key for AES-256
		if _, err := rand.Read(bytes); err != nil {
			panic(err.Error())
		}

		key := hex.EncodeToString(bytes) //encode key in bytes to string for saving

		fmt.Println(key)
		os.Exit(0)
	}

	e := echo.New()

	// e.Use(middleware.GzipWithConfig(middleware.GzipConfig{
	// 	Skipper: func(c echo.Context) bool {
	// 		return strings.Contains(c.Path(), "metrics")
	// 	},
	// }))

	e.HideBanner = true

	e.Debug = viper.GetBool("debug")

	e.Use(middleware.Logger())

	// // Enable metrics middleware
	// p := prometheus.NewPrometheus("echo", nil)
	// p.Use(e)

	e.GET("/", libs.MainRoute)
	e.GET("/health", libs.HealthRoute)
	e.GET("/config", libs.ConfigRoute)

	e.Logger.Fatal(e.Start(viper.GetString("listen")))
}
