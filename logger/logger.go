package logger

import (
	"io"
	"log"
	"os"

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

	// This code is opening a file specified in the configuration file using the `os.OpenFile` function.
	// It creates the file if it does not exist, opens it for writing only, and appends to the file if it
	// already exists. If there is an error opening the file, it logs the error and exits the program.
	file, err := os.OpenFile(viper.GetString("log_file"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal(err)
	}
	mw := io.MultiWriter(os.Stdout, file)

	// This code is creating a `slog.Handler` object based on the value of the `log_format` configuration
	// option. If the `log_format` is set to `"json"`, it creates a new JSON handler using the
	// `NewJSONHandler` method of the `opts` object, passing in `mw` (a `io.MultiWriter` that writes to
	// both `os.Stdout` and a log file) as the output destination. If the `log_format` is not set to
	// `"json"`, it creates a new text handler using the `NewTextHandler` method of the `opts` object,
	// again passing in `mw` as the output destination. The resulting `slog.Handler` object is then
	// assigned to the `handler` variable.
	var handler slog.Handler
	if viper.GetString("log_format") == "json" {
		handler = opts.NewJSONHandler(mw)
	} else {
		handler = opts.NewTextHandler(mw)
	}

	slog.SetDefault(slog.New(handler))
}
