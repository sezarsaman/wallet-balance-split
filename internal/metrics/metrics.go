package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Metrics holds all Prometheus metrics for the wallet service
type Metrics struct {
	// HTTP metrics
	RequestDuration prometheus.HistogramVec
	RequestCount    prometheus.CounterVec
	RequestErrors   prometheus.CounterVec

	// Database metrics
	DBConnections prometheus.Gauge
	DBQueryTime   prometheus.HistogramVec
	DBErrors      prometheus.CounterVec

	// Worker pool metrics
	WorkerQueueLength prometheus.Gauge
	WorkerCount       prometheus.Gauge
	TaskDuration      prometheus.HistogramVec
	TaskErrors        prometheus.CounterVec

	// Business metrics
	ChargeAmount    prometheus.HistogramVec
	WithdrawAmount  prometheus.HistogramVec
	BalanceSnapshot prometheus.GaugeVec
}

// New creates and registers all metrics
func New() *Metrics {
	return &Metrics{
		// HTTP endpoint metrics
		RequestDuration: *promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_request_duration_seconds",
				Help:    "Duration of HTTP requests in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "endpoint", "status"},
		),
		RequestCount: *promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_total",
				Help: "Total number of HTTP requests",
			},
			[]string{"method", "endpoint", "status"},
		),
		RequestErrors: *promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_errors_total",
				Help: "Total number of HTTP errors",
			},
			[]string{"method", "endpoint", "error_type"},
		),

		// Database metrics
		DBConnections: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "db_connections_active",
				Help: "Number of active database connections",
			},
		),
		DBQueryTime: *promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "db_query_duration_seconds",
				Help:    "Duration of database queries in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"query_type"},
		),
		DBErrors: *promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "db_errors_total",
				Help: "Total number of database errors",
			},
			[]string{"query_type", "error_type"},
		),

		// Worker pool metrics
		WorkerQueueLength: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "worker_queue_length",
				Help: "Current length of worker task queue",
			},
		),
		WorkerCount: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "worker_count_active",
				Help: "Number of active worker goroutines",
			},
		),
		TaskDuration: *promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "worker_task_duration_seconds",
				Help:    "Duration of worker tasks in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"task_type", "status"},
		),
		TaskErrors: *promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "worker_task_errors_total",
				Help: "Total number of worker task errors",
			},
			[]string{"task_type", "error_type"},
		),

		// Business metrics
		ChargeAmount: *promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "charge_amount",
				Help:    "Charge amount distribution",
				Buckets: prometheus.ExponentialBuckets(1000, 2, 10),
			},
			[]string{"user_id"},
		),
		WithdrawAmount: *promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "withdraw_amount",
				Help:    "Withdraw amount distribution",
				Buckets: prometheus.ExponentialBuckets(1000, 2, 10),
			},
			[]string{"user_id"},
		),
		BalanceSnapshot: *promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "user_balance",
				Help: "Current balance per user",
			},
			[]string{"user_id"},
		),
	}
}
