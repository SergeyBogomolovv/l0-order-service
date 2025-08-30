package app

import (
	"context"
	"errors"
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

func (a *Application) Start(ctx context.Context) error {
	for _, c := range a.consumers {
		go c.Consume(ctx)
	}

	go func() {
		a.logger.Info("starting http server", slog.String("addr", a.httpSrv.Addr))
		err := a.httpSrv.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			panic("failed to start http server: " + err.Error())
		}
	}()

	eg, ctx := errgroup.WithContext(ctx)

	for _, s := range a.starters {
		eg.Go(func() error {
			return s.Start(ctx)
		})
	}

	if err := eg.Wait(); err != nil {
		return err
	}

	a.logger.InfoContext(ctx, "Application started")

	return nil
}

const gracefulShutdownTimeout = 5 * time.Second

func (a *Application) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), gracefulShutdownTimeout)
	defer cancel()

	eg, ctx := errgroup.WithContext(ctx)

	for _, c := range a.consumers {
		eg.Go(func() error {
			return c.Close()
		})
	}

	eg.Go(func() error {
		return a.httpSrv.Shutdown(ctx)
	})

	if err := eg.Wait(); err != nil {
		return err
	}

	a.logger.Info("Application stopped")

	return nil
}
