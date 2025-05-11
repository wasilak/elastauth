package logger

import (
	"context"
	"log/slog"
	"os"

	"github.com/wasilak/elastauth/libs"
	"github.com/wasilak/loggergo"
)

// The function initializes a logger with a specified log level and format, allowing the user to choose
// between logging in JSON or text format.
func LoggerInit(ctx context.Context, level string, logFormat string) {
	var err error

	loggerConfig := loggergo.Config{
		Level:        loggergo.Types.LogLevelFromString(level),
		Format:       loggergo.Types.LogFormatFromString(logFormat),
		OutputStream: os.Stdout,
		DevMode:      loggergo.Types.LogLevelFromString(level) == slog.LevelDebug && logFormat == "plain",
		Output:       loggergo.Types.OutputConsole,
	}

	ctx, _, err = loggergo.Init(ctx, loggerConfig)
	if err != nil {
		slog.ErrorContext(ctx, err.Error())
		os.Exit(1)
	}

	libs.LogLeveler = loggergo.GetLogLevelAccessor()
}
