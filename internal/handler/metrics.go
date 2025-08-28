package handler

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	ordersProcessed = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: "order_service",
			Subsystem: "kafka_consumer",
			Name:      "orders_processed_total",
			Help:      "Total number of successfully processed orders",
		},
	)

	ordersFailed = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: "order_service",
			Subsystem: "kafka_consumer",
			Name:      "orders_failed_total",
			Help:      "Total number of failed order processing attempts",
		},
	)

	ordersDLQ = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: "order_service",
			Subsystem: "kafka_consumer",
			Name:      "orders_dlq_total",
			Help:      "Total number of orders written to DLQ",
		},
	)

	commitErrors = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: "order_service",
			Subsystem: "kafka_consumer",
			Name:      "commit_errors_total",
			Help:      "Total number of Kafka commit errors",
		},
	)

	orderProcessingDuration = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Namespace: "order_service",
			Subsystem: "kafka_consumer",
			Name:      "order_processing_duration_seconds",
			Help:      "Histogram of order processing durations in seconds",
			Buckets:   prometheus.DefBuckets,
		},
	)

	ordersInProgress = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "order_service",
			Subsystem: "kafka_consumer",
			Name:      "orders_in_progress",
			Help:      "Number of orders currently being processed",
		},
	)
)

var (
	orderRequestTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "order_service",
			Subsystem: "http",
			Name:      "order_requests_total",
			Help:      "Total number of requests to get order by UID",
		},
		[]string{"status"},
	)

	orderRequestDuration = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Namespace: "order_service",
			Subsystem: "http",
			Name:      "order_request_duration_seconds",
			Help:      "Histogram of request durations for get order by UID",
			Buckets:   prometheus.DefBuckets,
		},
	)

	orderRequestsInProgress = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "order_service",
			Subsystem: "http",
			Name:      "order_requests_in_progress",
			Help:      "Number of in-progress requests to get order by UID",
		},
	)
)

func RegisterMetrics() {
	prometheus.MustRegister(
		ordersProcessed,
		ordersFailed,
		ordersDLQ,
		commitErrors,
		orderProcessingDuration,
		ordersInProgress,

		orderRequestTotal,
		orderRequestDuration,
		orderRequestsInProgress,
	)
}
