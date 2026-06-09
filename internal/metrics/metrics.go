package metrics

import (
	"net/http"
	"runtime"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// =========================================================
// HTTP Metrics
// =========================================================

var (
	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests processed.",
		},
		[]string{"path", "method", "status"},
	)

	HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Histogram of response latencies for HTTP requests.",
			Buckets: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
		},
		[]string{"path", "method"},
	)

	HTTPRequestsInFlight = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "http_requests_in_flight",
			Help: "Current number of HTTP requests being served.",
		},
	)

	HTTPResponseSize = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_response_size_bytes",
			Help:    "Histogram of response sizes in bytes.",
			Buckets: prometheus.ExponentialBuckets(100, 10, 7), // 100B → ~100MB
		},
		[]string{"path", "method"},
	)
)

// =========================================================
// Database Metrics
// =========================================================

var (
	DBActiveConnections = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "db_active_connections",
			Help: "The number of active connections to the PostgreSQL database.",
		},
	)

	DBQueryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "db_query_duration_seconds",
			Help:    "Duration of database queries in seconds.",
			Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5},
		},
		[]string{"query_name", "status"}, // status: "success" | "error"
	)

	DBErrorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "db_errors_total",
			Help: "Total number of database errors.",
		},
		[]string{"operation"}, // e.g. "select", "insert", "update", "delete"
	)

	DBConnectionPoolSize = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "db_connection_pool_size",
			Help: "Maximum size of the database connection pool.",
		},
	)
)

// =========================================================
// Host / System Metrics
//
// NOTE: Prometheus's default Go collector already publishes the official
// go_goroutines, go_memstats_alloc_bytes, go_memstats_sys_bytes, and
// go_gc_duration_seconds metrics. We use an "app_" prefix here to avoid
// duplicate-registration panics while still exposing the values we sample
// ourselves (useful for cross-referencing or custom dashboards).
// =========================================================

var (
	AppMemAlloc = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "app_mem_alloc_bytes",
			Help: "Bytes of allocated heap objects (sampled by app collector).",
		},
	)

	AppMemSys = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "app_mem_sys_bytes",
			Help: "Total bytes of memory obtained from the OS (sampled by app collector).",
		},
	)

	AppGoroutines = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "app_goroutines",
			Help: "Number of currently running goroutines (sampled by app collector).",
		},
	)

	AppGCPauseTotal = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "app_gc_pause_total_seconds",
			Help: "Cumulative GC stop-the-world pause time in seconds (sampled by app collector).",
		},
	)

	// Application uptime
	AppStartTime = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "app_start_time_seconds",
			Help: "Unix timestamp when the application started.",
		},
	)
)

// =========================================================
// Business / Domain Metrics  (add your own here)
// =========================================================

var (
	// Example: track jobs processed by a background worker
	JobsProcessedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "jobs_processed_total",
			Help: "Total background jobs processed.",
		},
		[]string{"job_type", "status"}, // status: "success" | "failure"
	)

	// Example: cache hit/miss ratio
	CacheOperationsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cache_operations_total",
			Help: "Total cache operations.",
		},
		[]string{"operation", "result"}, // operation: "get"|"set"|"delete", result: "hit"|"miss"
	)
)

// =========================================================
// Initialisation
// =========================================================

func init() {
	AppStartTime.SetToCurrentTime()
}

// RecordRuntimeMetrics snapshots Go runtime stats into gauges.
// Call this in a goroutine on a regular interval, e.g. every 15 seconds.
//
//	go metrics.StartRuntimeCollector(15 * time.Second)
func RecordRuntimeMetrics() {
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)

	AppMemAlloc.Set(float64(ms.Alloc))
	AppMemSys.Set(float64(ms.Sys))
	AppGoroutines.Set(float64(runtime.NumGoroutine()))
	AppGCPauseTotal.Set(float64(ms.PauseTotalNs) / 1e9)
}

// StartRuntimeCollector runs RecordRuntimeMetrics on the given interval.
// Launch it once at startup: go metrics.StartRuntimeCollector(15 * time.Second)
func StartRuntimeCollector(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for range ticker.C {
		RecordRuntimeMetrics()
	}
}

// =========================================================
// responseWriter wrapper — captures status code & size
// =========================================================

type responseWriter struct {
	http.ResponseWriter
	statusCode int
	size       int
}

func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(b)
	rw.size += n
	return n, err
}

// =========================================================
// Middleware
// =========================================================

// MetricsMiddleware instruments any http.HandlerFunc with the full set of
// HTTP metrics: request count, duration, in-flight count, and response size.
//
// Usage:
//
//	mux.HandleFunc("/api/users", metrics.MetricsMiddleware(usersHandler))
func MetricsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		method := r.Method

		HTTPRequestsInFlight.Inc()
		defer HTTPRequestsInFlight.Dec()

		rw := newResponseWriter(w)
		start := time.Now()

		next.ServeHTTP(rw, r)

		duration := time.Since(start).Seconds()
		status := strconv.Itoa(rw.statusCode)

		HTTPRequestsTotal.WithLabelValues(path, method, status).Inc()
		HTTPRequestDuration.WithLabelValues(path, method).Observe(duration)
		HTTPResponseSize.WithLabelValues(path, method).Observe(float64(rw.size))
	}
}

// MetricsHandler wraps an http.Handler (for use with ServeMux patterns).
//
// Usage (stdlib mux):
//
//	mux.Handle("/api/", metrics.MetricsHandler(apiHandler))
func MetricsHandler(next http.Handler) http.Handler {
	return MetricsMiddleware(next.ServeHTTP)
}