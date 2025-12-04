package handlers

import (
	"net/http"
	"time"
	"wallet-simulator/internal/metrics"

	"github.com/go-chi/chi/v5"
)

// MetricsMiddleware records HTTP metrics for Prometheus
func MetricsMiddleware(m *metrics.Metrics) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rctx := chi.RouteContext(r.Context())
			routePattern := rctx.RoutePattern()

			wrapped := &responseWriter{ResponseWriter: w, statusCode: 200}

			next.ServeHTTP(wrapped, r)

			duration := time.Since(start).Seconds()
			status := http.StatusText(wrapped.statusCode)

			// Record metrics
			m.RequestDuration.WithLabelValues(r.Method, routePattern, status).Observe(duration)
			m.RequestCount.WithLabelValues(r.Method, routePattern, status).Inc()

			if wrapped.statusCode >= 400 {
				m.RequestErrors.WithLabelValues(r.Method, routePattern, "http_error").Inc()
			}
		})
	}
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
