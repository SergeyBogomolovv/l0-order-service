package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/SergeyBogomolovv/l0-order-service/internal/app"
	"github.com/SergeyBogomolovv/l0-order-service/internal/config"
	"github.com/SergeyBogomolovv/l0-order-service/internal/handler"
	"github.com/SergeyBogomolovv/l0-order-service/internal/postgres"
	"github.com/SergeyBogomolovv/l0-order-service/internal/repo"
	"github.com/SergeyBogomolovv/l0-order-service/internal/service"
	"github.com/SergeyBogomolovv/l0-order-service/pkg/cache"
	"github.com/SergeyBogomolovv/l0-order-service/pkg/trm"

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
	cache := cache.NewLRUCache(conf.Cache.Capacity, conf.Cache.TTL)

	orderService := service.NewOrderService(logger, txManager, orderRepo, cache)

	kafkaHandler := handler.NewKafkaHandler(logger, conf.Kafka, orderService)
	httpHandler := handler.NewHttpHandler(logger, orderService)

	app := app.New(logger, conf)

	app.SetHttpHandlers(httpHandler)
	app.SetKafkaHandlers(kafkaHandler)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	go orderService.WarmUpCache(ctx, conf.Cache.Capacity)
	cache.StartJanitor(ctx)
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
