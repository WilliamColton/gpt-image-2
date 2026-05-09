package log

import (
	"log/slog"
	"os"
)

var Logger *slog.Logger

func Init(json bool) {
	if json {
		Logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	} else {
		Logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}
	slog.SetDefault(Logger)
}
