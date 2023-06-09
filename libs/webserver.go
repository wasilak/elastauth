package libs

import (
	_ "net/http/pprof"
	"os"
	"strings"

	"github.com/labstack/echo-contrib/prometheus"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/spf13/viper"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"
)

// WebserverInit initializes the webserver and sets up all the routes. It
// configures the server based on settings from the [viper] configuration
// library. It also adds support for metrics if enabled with the
// `enable_metrics` flag. Lastly, it starts the server on the `listen` address
// specified in the configuration.
func WebserverInit() {
	e := echo.New()

	e.HideBanner = true

	e.HidePort = true

	e.Debug = viper.GetBool("debug")

	e.Use(middleware.Logger())

	if viper.GetBool("enableOtel") {
		e.Use(otelecho.Middleware(os.Getenv("OTEL_SERVICE_NAME")))
	}

	// This code block is checking if the `enable_metrics` flag is set to true in the configuration file
	// using the `viper` library. If it is true, it adds middleware to compress the response using Gzip,
	// but skips compression for requests that contain the word "metrics" in the path. It then creates a
	// new instance of the Prometheus middleware with the application name "elastauth" and adds it to the
	// Echo server. This middleware will collect metrics for all HTTP requests and expose them on a
	// `/metrics` endpoint.
	if viper.GetBool("enable_metrics") {
		e.Use(middleware.GzipWithConfig(middleware.GzipConfig{
			Skipper: func(c echo.Context) bool {
				return strings.Contains(c.Path(), "metrics")
			},
		}))

		// Enable metrics middleware
		p := prometheus.NewPrometheus("elastauth", nil)
		p.Use(e)
	}

	e.GET("/", MainRoute)
	e.GET("/health", HealthRoute)
	e.GET("/config", ConfigRoute)

	e.Logger.Fatal(e.Start(viper.GetString("listen")))
}
