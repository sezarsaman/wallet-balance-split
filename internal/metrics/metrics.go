package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type Metrics struct {
	// HTTP metrics
	RequestDuration *prometheus.HistogramVec
	RequestCount    *prometheus.CounterVec
	RequestErrors   *prometheus.CounterVec

	// Database metrics
	DBConnections prometheus.Gauge
	DBQueryTime   *prometheus.HistogramVec
	DBErrors      *prometheus.CounterVec

	// Worker pool metrics
	WorkerQueueLength prometheus.Gauge
	WorkerCount       prometheus.Gauge
	TaskDuration      *prometheus.HistogramVec
	TaskErrors        *prometheus.CounterVec

	// Business metrics (no per-user labels!)
	ChargeAmount    *prometheus.HistogramVec
	WithdrawAmount  *prometheus.HistogramVec
	BalanceSnapshot *prometheus.GaugeVec
}

func New() *Metrics {
	return &Metrics{
		RequestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_request_duration_seconds",
				Help:    "Duration of HTTP requests in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "endpoint", "status"},
		),
		RequestCount: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_total",
				Help: "Total number of HTTP requests",
			},
			[]string{"method", "endpoint", "status"},
		),
		RequestErrors: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_errors_total",
				Help: "Total number of HTTP errors",
			},
			[]string{"method", "endpoint", "error_type"},
		),

		DBConnections: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "db_connections_active",
				Help: "Number of active DB connections",
			},
		),
		DBQueryTime: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "db_query_duration_seconds",
				Help:    "DB query duration",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"query_type"},
		),
		DBErrors: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "db_errors_total",
				Help: "Total database errors",
			},
			[]string{"query_type", "error_type"},
		),

		WorkerQueueLength: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "worker_queue_length",
				Help: "Current worker queue length",
			},
		),
		WorkerCount: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "worker_count_active",
				Help: "Active worker goroutines",
			},
		),
		TaskDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "worker_task_duration_seconds",
				Help:    "Duration of worker tasks",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"task_type", "status"},
		),
		TaskErrors: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "worker_task_errors_total",
				Help: "Total worker task errors",
			},
			[]string{"task_type", "error_type"},
		),

		// Avoid cardinality explosion: NO user-id labels!
		ChargeAmount: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "charge_amount",
				Help:    "Charge amount distribution",
				Buckets: prometheus.ExponentialBuckets(1000, 2, 10),
			},
			[]string{"operation"},
		),
		WithdrawAmount: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "withdraw_amount",
				Help:    "Withdraw amount distribution",
				Buckets: prometheus.ExponentialBuckets(1000, 2, 10),
			},
			[]string{"operation"},
		),
		BalanceSnapshot: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "user_balance_snapshot",
				Help: "Last calculated balance snapshot (no per-user labels)",
			},
			[]string{"type"},
		),
	}
}
