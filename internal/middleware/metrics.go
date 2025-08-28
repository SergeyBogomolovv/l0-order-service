package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	httpInFlight = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "order_service",
		Subsystem: "http",
		Name:      "in_flight_requests",
		Help:      "Current number of in-flight HTTP requests.",
	})

	httpRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "order_service",
		Subsystem: "http",
		Name:      "requests_total",
		Help:      "Total number of HTTP requests processed.",
	}, []string{"method", "route", "status"})

	httpRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "order_service",
		Subsystem: "http",
		Name:      "request_duration",
		Help:      "HTTP request latencies in seconds.",
		Buckets:   prometheus.DefBuckets,
	}, []string{"method", "route", "status"})
)

func Metrics(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		httpInFlight.Inc()
		defer httpInFlight.Dec()

		start := time.Now()
		rw := wrapResponseWriter(w)

		next.ServeHTTP(rw, r)

		route := "unknown"
		if rc := chi.RouteContext(r.Context()); rc != nil && rc.RoutePattern() != "" {
			route = rc.RoutePattern()
		}

		labels := prometheus.Labels{
			"method": r.Method,
			"route":  route,
			"status": strconv.Itoa(rw.status),
		}

		httpRequestsTotal.With(labels).Inc()
		httpRequestDuration.With(labels).Observe(time.Since(start).Seconds())
	})
}
