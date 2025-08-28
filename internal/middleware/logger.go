package middleware

import (
	"log/slog"
	"net/http"
	"time"

	chimw "github.com/go-chi/chi/v5/middleware"
)

func Logger(logger *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rw := wrapResponseWriter(w)
			next.ServeHTTP(rw, r)
			reqID := chimw.GetReqID(r.Context())

			logger.Info("request",
				slog.Int("status", rw.status),
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.String("remote", r.RemoteAddr),
				slog.String("duration", time.Since(start).String()),
				slog.String("request_id", reqID),
			)
		})
	}
}
