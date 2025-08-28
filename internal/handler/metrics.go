package handler

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	ordersProcessed = promauto.NewCounter(
		prometheus.CounterOpts{
			Namespace: "order_service",
			Subsystem: "kafka_consumer",
			Name:      "orders_processed_total",
			Help:      "Total number of successfully processed orders",
		},
	)

	ordersFailed = promauto.NewCounter(
		prometheus.CounterOpts{
			Namespace: "order_service",
			Subsystem: "kafka_consumer",
			Name:      "orders_failed_total",
			Help:      "Total number of failed order processing attempts",
		},
	)

	ordersDLQ = promauto.NewCounter(
		prometheus.CounterOpts{
			Namespace: "order_service",
			Subsystem: "kafka_consumer",
			Name:      "orders_dlq_total",
			Help:      "Total number of orders written to DLQ",
		},
	)

	commitErrors = promauto.NewCounter(
		prometheus.CounterOpts{
			Namespace: "order_service",
			Subsystem: "kafka_consumer",
			Name:      "commit_errors_total",
			Help:      "Total number of Kafka commit errors",
		},
	)

	orderProcessingDuration = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Namespace: "order_service",
			Subsystem: "kafka_consumer",
			Name:      "order_processing_duration_seconds",
			Help:      "Histogram of order processing durations in seconds",
			Buckets:   prometheus.DefBuckets,
		},
	)
)
