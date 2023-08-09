package logger

import (
	"os"
	"strings"

	"log/slog"

	otelgoslog "github.com/wasilak/otelgo/slog"
)

// The function initializes a logger with a specified log level and format, allowing the user to choose
// between logging in JSON or text format.
func LoggerInit(level string, logFormat string) {

	// This code block is setting the log level based on the `level` parameter passed to the `LoggerInit`
	// function. It converts the `level` parameter to uppercase using `strings.ToUpper` and then sets the
	// `logLevel` variable to the corresponding `slog.Leveler` value based on the string value of `level`.
	// If `level` is not one of the recognized values ("INFO", "ERROR", "WARN", "DEBUG"), it sets the log
	// level to `slog.LevelInfo` by default. The `logLevel` variable is then used to set the log level in
	// the `opts` variable, which is used to configure the logger handler.
	var logLevel slog.Leveler

	switch strings.ToUpper(level) {
	case "INFO":
		logLevel = slog.LevelInfo
	case "ERROR":
		logLevel = slog.LevelError
	case "WARN":
		logLevel = slog.LevelWarn
	case "DEBUG":
		logLevel = slog.LevelDebug
	default:
		logLevel = slog.LevelInfo
	}

	opts := slog.HandlerOptions{
		Level:     logLevel,
		AddSource: true,
	}

	// This code block is checking if the `logFormat` parameter passed to the `LoggerInit` function is
	// equal to the string "json" in lowercase. If it is, it creates a new logger with a JSON handler and
	// sets it as the default logger using `slog.SetDefault`. If it is not equal to "json", it creates a
	// new logger with a text handler and sets it as the default logger. This allows the user to choose
	// between logging in JSON format or text format.
	if strings.ToLower(logFormat) == "json" {
		slog.SetDefault(slog.New(otelgoslog.NewTracingHandler(slog.NewJSONHandler(os.Stderr, &opts))))
	} else {
		slog.SetDefault(slog.New(otelgoslog.NewTracingHandler(slog.NewTextHandler(os.Stderr, &opts))))
	}
}
