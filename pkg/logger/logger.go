package logger

import (
	"log/slog"
	"os"
)

func New(env string) *slog.Logger {
	var handler slog.Handler

	switch env {
	case "production":
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})
	case "staging":
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})
	default:
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})
	}

	return slog.New(handler)
}
