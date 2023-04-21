package logger

import (
	"io"
	"log"
	"os"

	"github.com/spf13/viper"
	"golang.org/x/exp/slog"
)

var LogLevel *slog.LevelVar

var LoggerInstance *slog.Logger

func LoggerInit() {
	LogLevel = new(slog.LevelVar)

	opts := slog.HandlerOptions{
		Level:     LogLevel,
		AddSource: true,
	}

	file, err := os.OpenFile(viper.GetString("log_file"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal(err)
	}
	mw := io.MultiWriter(os.Stdout, file)

	textHandler := opts.NewTextHandler(mw)
	LoggerInstance = slog.New(textHandler)
}
