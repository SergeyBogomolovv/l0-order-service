package main

import (
	"context"
	"log/slog"
	"os/signal"
	"syscall"

	"github.com/SergeyBogomolovv/l0-order-service/internal/app"
	"github.com/SergeyBogomolovv/l0-order-service/internal/config"
	"github.com/SergeyBogomolovv/l0-order-service/internal/handler"
	"github.com/SergeyBogomolovv/l0-order-service/internal/postgres"
	"github.com/SergeyBogomolovv/l0-order-service/internal/repo"
	"github.com/SergeyBogomolovv/l0-order-service/internal/service"
	"github.com/SergeyBogomolovv/l0-order-service/pkg/cache"
	"github.com/SergeyBogomolovv/l0-order-service/pkg/logger"
	"github.com/SergeyBogomolovv/l0-order-service/pkg/trm"
	"github.com/joho/godotenv"
)

// @title           Order Service API
// @version         1.0
// @description     Документация HTTP API
func main() {
	// optionally load config from .env
	godotenv.Load()

	// load and validate config
	conf := config.New()
	if err := conf.Validate(); err != nil {
		panic("invalid config: " + err.Error())
	}

	// init logger
	log := logger.New(conf.Env)

	// connect to db
	db, err := postgres.New(conf.Postgres)
	if err != nil {
		panic("failed to connect to db: " + err.Error())
	}
	defer db.Close()
	log.Info("postgres connected")

	// init dependencies
	orderRepo := repo.NewPostgresRepo(db)
	txManager := trm.NewManager(db)
	cache := cache.NewLRUCache(conf.Cache.Capacity, conf.Cache.TTL)
	orderService := service.NewOrderService(log, txManager, orderRepo, cache)
	kafkaHandler := handler.NewKafkaHandler(log, conf.Kafka, orderService)
	httpHandler := handler.NewHTTPHandler(log, orderService)

	// init app
	app := app.New(log, conf)
	app.SetHTTPHandlers(httpHandler)
	app.SetConsumers(kafkaHandler)
	app.SetStarters(cache, cacheWarmUpAdapter{svc: orderService, count: conf.Cache.Capacity})

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	// start app
	go func() {
		if err := app.RunServer(ctx); err != nil {
			panic("failed to run server: " + err.Error())
		}
	}()

	if err := app.StartStarters(ctx); err != nil {
		panic("failed to start starters: " + err.Error())
	}

	app.StartConsumers(ctx)

	<-ctx.Done()

	// graceful shutdown
	if err := app.StopServer(); err != nil {
		log.Error("failed to stop server", slog.Any("error", err))
	}
	if err := app.CloseConsumers(); err != nil {
		log.Error("failed to close consumers", slog.Any("error", err))
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
