package logger

import (
	"os"
	"strings"

	"github.com/spf13/viper"
	"golang.org/x/exp/slog"
)

var LogLevel *slog.LevelVar

// This function initializes a logger with options for log level, log file, and log format.
func LoggerInit() {

	opts := slog.HandlerOptions{
		Level:     LogLevel,
		AddSource: true,
	}

	// This code block is checking the value of the "log_format" configuration setting using the Viper
	// library. If the value is "json" (case-insensitive), it creates a new logger with a JSON log handler
	// and sets it as the default logger using the slog library. Otherwise, it creates a new logger with a
	// text log handler and sets it as the default logger. This allows the user to choose between JSON and
	// text log formats for their application.
	if strings.ToLower(viper.GetString("log_format")) == "json" {
		slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stderr, &opts)))
	} else {
		slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &opts)))
	}
}
