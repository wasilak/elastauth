package libs

import (
	_ "net/http/pprof"
	"strings"

	"github.com/labstack/echo-contrib/prometheus"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/spf13/viper"
)

func WebserverInit() {
	e := echo.New()

	e.HideBanner = true

	e.Debug = viper.GetBool("debug")

	e.Use(middleware.Logger())

	if viper.GetBool("enable_metrics") {
		e.Use(middleware.GzipWithConfig(middleware.GzipConfig{
			Skipper: func(c echo.Context) bool {
				return strings.Contains(c.Path(), "metrics")
			},
		}))

		// Enable metrics middleware
		p := prometheus.NewPrometheus("echo", nil)
		p.Use(e)
	}

	e.GET("/", MainRoute)
	e.GET("/health", HealthRoute)
	e.GET("/config", ConfigRoute)

	e.Logger.Fatal(e.Start(viper.GetString("listen")))
}
