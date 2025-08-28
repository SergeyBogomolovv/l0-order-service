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
	panicIfErr("invalid config", conf.Validate())

	db, err := postgres.New(conf.Postgres)
	panicIfErr("failed to connect to db", err)
	defer db.Close()
	logger.Info("postgres connected")

	orderRepo := repo.NewPostgresRepo(db)
	txManager := trm.NewManager(db)
	cache := cache.NewLRUCache(conf.Cache.Capacity, conf.Cache.TTL)

	orderService := service.NewOrderService(logger, txManager, orderRepo, cache)

	kafkaHandler := handler.NewKafkaHandler(logger, conf.Kafka, orderService)
	httpHandler := handler.NewHTTPHandler(logger, orderService)

	app := app.New(logger, conf)

	app.SetHTTPHandlers(httpHandler)
	app.SetConsumers(kafkaHandler)
	app.SetStarters(cache, cacheWarmUpAdapter{svc: orderService, count: conf.Cache.Capacity})

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	panicIfErr("failed to start app", app.Start(ctx))
	<-ctx.Done()
	panicIfErr("failed to stop app", app.Stop())
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

func panicIfErr(prefix string, err error) {
	if err != nil {
		panic(prefix + ": " + err.Error())
	}
}

type warmUpper interface {
	WarmUpCache(ctx context.Context, count int) error
}

type cacheWarmUpAdapter struct {
	svc   warmUpper
	count int
}

func (a cacheWarmUpAdapter) Start(ctx context.Context) error {
	return a.svc.WarmUpCache(ctx, a.count)
}
