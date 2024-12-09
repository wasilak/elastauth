package libs

import (
	"context"
	"log/slog"
	_ "net/http/pprof"
	"strings"

	"github.com/labstack/echo-contrib/prometheus"
	"github.com/labstack/gommon/log"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	slogecho "github.com/samber/slog-echo"
	"github.com/spf13/viper"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"
	"go.opentelemetry.io/otel"
)

var tracerWebserver = otel.Tracer("webserver")

// WebserverInit initializes the webserver and sets up all the routes. It
// configures the server based on settings from the [viper] configuration
// library. It also adds support for metrics if enabled with the
// `enable_metrics` flag. Lastly, it starts the server on the `listen` address
// specified in the configuration.
func WebserverInit(ctx context.Context) {
	_, span := tracerWebserver.Start(ctx, "WebserverInit")
	defer span.End()

	e := echo.New()

	e.HideBanner = true

	e.HidePort = true

	// setting log/slog log level as echo logger level
	e.Logger.SetLevel(log.Lvl(LogLeveler.Level().Level()))

	e.Debug = strings.EqualFold(LogLeveler.Level().String(), "debug") || viper.GetBool("debug")

	e.Use(slogecho.New(slog.Default()))

	if viper.GetBool("enableOtel") {
		e.Use(otelecho.Middleware(GetAppName(), otelecho.WithSkipper(func(c echo.Context) bool {
			return strings.Contains(c.Path(), "health")
		})))
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
