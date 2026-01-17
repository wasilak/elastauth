package libs

import (
	"context"
	"log/slog"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

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
			return strings.Contains(c.Path(), "health") || strings.Contains(c.Path(), "ready") || strings.Contains(c.Path(), "live")
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

	// Main application routes
	e.GET("/", MainRoute)
	e.GET("/config", ConfigRoute)
	e.GET("/docs", SwaggerUIRoute)
	e.GET("/api/openapi.yaml", SwaggerRoute)

	// Health check endpoints for Kubernetes
	e.GET("/health", HealthRoute)      // Basic health check (legacy)
	e.GET("/ready", ReadinessRoute)    // Kubernetes readiness probe
	e.GET("/live", LivenessRoute)      // Kubernetes liveness probe

	// Start server with graceful shutdown
	StartServerWithGracefulShutdown(ctx, e)
}

// StartServerWithGracefulShutdown starts the Echo server with graceful shutdown support
func StartServerWithGracefulShutdown(ctx context.Context, e *echo.Echo) {
	// Start server in a goroutine
	go func() {
		listenAddr := viper.GetString("listen")
		slog.InfoContext(ctx, "Starting server", slog.String("address", listenAddr))
		
		if err := e.Start(listenAddr); err != nil && err != http.ErrServerClosed {
			slog.ErrorContext(ctx, "Server failed to start", slog.Any("error", err))
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	
	sig := <-quit
	slog.InfoContext(ctx, "Received shutdown signal", slog.String("signal", sig.String()))

	// Create a context with timeout for graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	slog.InfoContext(ctx, "Shutting down server gracefully...")
	
	if err := e.Shutdown(shutdownCtx); err != nil {
		slog.ErrorContext(ctx, "Server forced to shutdown", slog.Any("error", err))
		os.Exit(1)
	}

	slog.InfoContext(ctx, "Server shutdown complete")
}
