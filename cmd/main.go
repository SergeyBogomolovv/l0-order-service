package main

import (
	"OrderService/internal/app"
	"OrderService/internal/config"
	"OrderService/internal/handler"
	"OrderService/internal/postgres"
	"OrderService/internal/repo"
	"OrderService/internal/service"
	"OrderService/pkg/trm"
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	_ "OrderService/docs"

	"github.com/joho/godotenv"
)

// @title           Order Service API
// @version         1.0
// @description     Документация HTTP API
func main() {
	conf := config.New()
	logger := newLogger(conf.Env)
	exitIfErr(logger, "invalid config", conf.Validate())

	db, err := postgres.New(conf.Postgres)
	exitIfErr(logger, "failed to connect to db", err)
	logger.Info("postgres connected")

	orderRepo := repo.NewPostgresRepo(db)
	txManager := trm.NewManager(db)

	orderService := service.NewOrderService(logger, txManager, orderRepo)

	kafkaHandler := handler.NewKafkaHandler(logger, conf.Kafka, orderService)
	httpHandler := handler.NewHttpHandler(logger, orderService)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	app := app.New(logger, conf)

	app.SetHttpHandlers(httpHandler)
	app.SetKafkaHandlers(kafkaHandler)

	app.Start(ctx)
	<-ctx.Done()
	app.Stop()
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
