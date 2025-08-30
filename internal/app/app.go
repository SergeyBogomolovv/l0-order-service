package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"

	// swagger docs
	_ "github.com/SergeyBogomolovv/l0-order-service/docs"
	"github.com/SergeyBogomolovv/l0-order-service/internal/config"
	"github.com/SergeyBogomolovv/l0-order-service/internal/middleware"
	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	httpSwagger "github.com/swaggo/http-swagger/v2"
	"golang.org/x/sync/errgroup"
)

type Consumer interface {
	Consume(ctx context.Context)
	Close() error
}

type Starter interface {
	Start(ctx context.Context) error
}

type HTTPHandler interface {
	Init(r chi.Router)
}

type Application struct {
	logger *slog.Logger

	router    chi.Router
	httpSrv   *http.Server
	consumers []Consumer
	starters  []Starter
}

func New(logger *slog.Logger, cfg config.Config) *Application {
	router := chi.NewRouter()
	router.Use(chimw.RequestID)
	router.Use(chimw.RealIP)
	router.Use(chimw.Recoverer)
	router.Use(chimw.Timeout(30 * time.Second))
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins: cfg.Cors.AllowedOrigins,
		AllowedMethods: []string{"GET", "OPTIONS"},
		AllowedHeaders: []string{"Accept", "Content-Type", "Authorization"},
		ExposedHeaders: []string{"Link"},
	}))
	router.Use(middleware.Metrics)
	router.Use(middleware.Logger(logger))

	if cfg.Env != "production" {
		router.Get("/swagger/*", httpSwagger.WrapHandler)
		router.Mount("/debug", chimw.Profiler())
	}

	router.Mount("/metrics", promhttp.Handler())

	httpSrv := &http.Server{
		Handler:           router,
		Addr:              net.JoinHostPort(cfg.HTTP.Host, cfg.HTTP.Port),
		ReadHeaderTimeout: 30 * time.Second,
	}

	return &Application{
		logger:  logger,
		httpSrv: httpSrv,
		router:  router,
	}
}

func (a *Application) SetHTTPHandlers(handlers ...HTTPHandler) {
	for _, h := range handlers {
		h.Init(a.router)
	}
}

func (a *Application) SetConsumers(consumers ...Consumer) {
	a.consumers = consumers
}

func (a *Application) SetStarters(starters ...Starter) {
	a.starters = starters
}

func (a *Application) StartConsumers(ctx context.Context) {
	for _, c := range a.consumers {
		go c.Consume(ctx)
	}
}

func (a *Application) StartStarters(ctx context.Context) error {
	eg, ctx := errgroup.WithContext(ctx)

	for _, s := range a.starters {
		eg.Go(func() error {
			return s.Start(ctx)
		})
	}

	return eg.Wait()
}

func (a *Application) RunServer(ctx context.Context) error {
	a.logger.InfoContext(ctx, "starting http server", slog.String("addr", a.httpSrv.Addr))
	err := a.httpSrv.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("failed to start http server: %w", err)
	}

	return nil
}

func (a *Application) StopServer() error {
	const shutdownTimeout = 5 * time.Second

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	a.logger.Info("stopping http server")

	return a.httpSrv.Shutdown(ctx)
}

func (a *Application) CloseConsumers() error {
	a.logger.Info("closing consumers")
	for _, c := range a.consumers {
		if err := c.Close(); err != nil {
			return fmt.Errorf("failed to close consumer: %w", err)
		}
	}
	return nil
}
