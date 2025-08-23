package app

import (
	"OrderService/internal/config"
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

type application struct {
	logger *slog.Logger

	router    chi.Router
	httpSrv   *http.Server
	consumers []KafkaHandler
}

func New(logger *slog.Logger, cfg config.Config) *application {
	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins: cfg.Cors.AllowedOrigins,
	}))

	httpSrv := &http.Server{
		Handler: router,
		Addr:    net.JoinHostPort(cfg.Http.Host, cfg.Http.Port),
	}

	return &application{
		logger:  logger,
		httpSrv: httpSrv,
		router:  router,
	}
}

type HttpHandler interface {
	Init(r chi.Router)
}

func (a *application) SetHttpHandlers(handlers ...HttpHandler) {
	for _, h := range handlers {
		h.Init(a.router)
	}
}

type KafkaHandler interface {
	Consume(ctx context.Context)
	Close() error
}

func (a *application) SetKafkaHandlers(handlers ...KafkaHandler) {
	a.consumers = handlers
}

func (a *application) Start(ctx context.Context) {
	for _, c := range a.consumers {
		go c.Consume(ctx)
	}

	go a.startServer()

	a.logger.Info("application started")
}

func (a *application) startServer() {
	a.logger.Info("starting http server", slog.String("addr", a.httpSrv.Addr))
	if err := a.httpSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		a.logger.Error("failed to start http server", slog.Any("error", err))
		os.Exit(1)
	}
}

const gracefulShutdownTimeout = 5 * time.Second

func (a *application) Stop() {
	for _, c := range a.consumers {
		if err := c.Close(); err != nil {
			a.logger.Error("failed to close kafka consumer", slog.Any("error", err))
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), gracefulShutdownTimeout)
	defer cancel()

	if err := a.httpSrv.Shutdown(ctx); err != nil {
		a.logger.Error("failed to shutdown http server", slog.Any("error", err))
	}

	a.logger.Info("application stopped")
}
