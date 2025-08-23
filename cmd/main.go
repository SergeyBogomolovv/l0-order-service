package main

import (
	"OrderService/internal/config"
	"OrderService/internal/postgres"
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
)

func main() {
	conf := config.New()
	logger := newLogger(conf.Env)
	exitIfErr(logger, "invalid config", conf.Validate())

	db, err := postgres.New(conf.Postgres)
	exitIfErr(logger, "failed to connect to db", err)
	logger.Info("postgres connected")

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	<-ctx.Done()
	db.Close()
}

func init() {
	godotenv.Load()
}

func newLogger(env string) *slog.Logger {
	switch env {
	case "production":
		return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	default:
		return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	}
}

func exitIfErr(logger *slog.Logger, prefix string, err error) {
	if err != nil {
		logger.Error(prefix, slog.Any("error", err))
		os.Exit(1)
	}
}
