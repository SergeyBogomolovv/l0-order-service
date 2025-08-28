package app

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/SergeyBogomolovv/l0-order-service/internal/config"
	"github.com/SergeyBogomolovv/l0-order-service/internal/middleware"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	_ "github.com/SergeyBogomolovv/l0-order-service/docs"

	httpSwagger "github.com/swaggo/http-swagger/v2"
)

type KafkaHandler interface {
	Consume(ctx context.Context)
	Close() error
}

type HTTPHandler interface {
	Init(r chi.Router)
}

type application struct {
	logger *slog.Logger

	router    chi.Router
	httpSrv   *http.Server
	consumers []KafkaHandler
}

func New(logger *slog.Logger, cfg config.Config) *application {
	router := chi.NewRouter()
	router.Use(chimw.RequestID)
	router.Use(chimw.RealIP)
	router.Use(middleware.Logger(logger))
	router.Use(chimw.Recoverer)
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins: cfg.Cors.AllowedOrigins,
		AllowedMethods: []string{"GET"},
	}))

	router.Get("/swagger/*", httpSwagger.WrapHandler)

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

func (a *application) SetHTTPHandlers(handlers ...HTTPHandler) {
	for _, h := range handlers {
		h.Init(a.router)
	}
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

	a.logger.Info("kafka consumers stopped")

	ctx, cancel := context.WithTimeout(context.Background(), gracefulShutdownTimeout)
	defer cancel()

	if err := a.httpSrv.Shutdown(ctx); err != nil {
		a.logger.Error("failed to shutdown http server", slog.Any("error", err))
		os.Exit(1)
	}

	a.logger.Info("http server stopped")
}
