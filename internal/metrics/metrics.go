package metrics

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// MetricsService handles Prometheus metrics collection
type MetricsService struct {
	registry *prometheus.Registry
	
	// HTTP metrics
	httpRequestsTotal     *prometheus.CounterVec
	httpRequestDuration   *prometheus.HistogramVec
	httpRequestsInFlight  *prometheus.GaugeVec
	
	// Repository metrics
	repositoryOperationsTotal *prometheus.CounterVec
	repositoryOperationDuration *prometheus.HistogramVec
	artifactUploadsTotal      *prometheus.CounterVec
	artifactDownloadsTotal    *prometheus.CounterVec
	artifactStorageSize       *prometheus.GaugeVec
	
	// Database metrics
	databaseConnectionsActive prometheus.Gauge
	databaseQueriesTotal      *prometheus.CounterVec
	databaseQueryDuration     *prometheus.HistogramVec
	
	// System metrics
	systemInfo                *prometheus.GaugeVec
	uptime                   *prometheus.CounterVec
}

// NewMetricsService creates a new metrics service
func NewMetricsService() *MetricsService {
	registry := prometheus.NewRegistry()
	
	// HTTP metrics
	httpRequestsTotal := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ganje_http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status_code"},
	)
	
	httpRequestDuration := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "ganje_http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)
	
	httpRequestsInFlight := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ganje_http_requests_in_flight",
			Help: "Number of HTTP requests currently being processed",
		},
		[]string{"method", "endpoint"},
	)
	
	// Repository metrics
	repositoryOperationsTotal := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ganje_repository_operations_total",
			Help: "Total number of repository operations",
		},
		[]string{"repository", "operation", "artifact_type", "status"},
	)
	
	repositoryOperationDuration := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "ganje_repository_operation_duration_seconds",
			Help:    "Repository operation duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"repository", "operation", "artifact_type"},
	)
	
	artifactUploadsTotal := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ganje_artifact_uploads_total",
			Help: "Total number of artifact uploads",
		},
		[]string{"repository", "artifact_type"},
	)
	
	artifactDownloadsTotal := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ganje_artifact_downloads_total",
			Help: "Total number of artifact downloads",
		},
		[]string{"repository", "artifact_type"},
	)
	
	artifactStorageSize := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ganje_artifact_storage_bytes",
			Help: "Total storage size of artifacts in bytes",
		},
		[]string{"repository", "artifact_type"},
	)
	
	// Database metrics
	databaseConnectionsActive := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "ganje_database_connections_active",
			Help: "Number of active database connections",
		},
	)
	
	databaseQueriesTotal := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ganje_database_queries_total",
			Help: "Total number of database queries",
		},
		[]string{"operation", "status"},
	)
	
	databaseQueryDuration := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "ganje_database_query_duration_seconds",
			Help:    "Database query duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"operation"},
	)
	
	// System metrics
	systemInfo := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ganje_system_info",
			Help: "System information",
		},
		[]string{"version", "go_version"},
	)
	
	uptime := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ganje_uptime_seconds_total",
			Help: "Total uptime in seconds",
		},
		[]string{"instance"},
	)
	
	// Register all metrics
	registry.MustRegister(
		httpRequestsTotal,
		httpRequestDuration,
		httpRequestsInFlight,
		repositoryOperationsTotal,
		repositoryOperationDuration,
		artifactUploadsTotal,
		artifactDownloadsTotal,
		artifactStorageSize,
		databaseConnectionsActive,
		databaseQueriesTotal,
		databaseQueryDuration,
		systemInfo,
		uptime,
	)
	
	// Add Go runtime metrics
	registry.MustRegister(prometheus.NewGoCollector())
	registry.MustRegister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))
	
	return &MetricsService{
		registry:                    registry,
		httpRequestsTotal:           httpRequestsTotal,
		httpRequestDuration:         httpRequestDuration,
		httpRequestsInFlight:        httpRequestsInFlight,
		repositoryOperationsTotal:   repositoryOperationsTotal,
		repositoryOperationDuration: repositoryOperationDuration,
		artifactUploadsTotal:        artifactUploadsTotal,
		artifactDownloadsTotal:      artifactDownloadsTotal,
		artifactStorageSize:         artifactStorageSize,
		databaseConnectionsActive:   databaseConnectionsActive,
		databaseQueriesTotal:        databaseQueriesTotal,
		databaseQueryDuration:       databaseQueryDuration,
		systemInfo:                  systemInfo,
		uptime:                      uptime,
	}
}

// GetRegistry returns the Prometheus registry
func (m *MetricsService) GetRegistry() *prometheus.Registry {
	return m.registry
}

// GetHandler returns the Prometheus HTTP handler
func (m *MetricsService) GetHandler() http.Handler {
	return promhttp.HandlerFor(m.registry, promhttp.HandlerOpts{})
}

// GinMiddleware returns a Gin middleware for HTTP metrics collection
func (m *MetricsService) GinMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}
		
		// Track requests in flight
		m.httpRequestsInFlight.WithLabelValues(c.Request.Method, path).Inc()
		defer m.httpRequestsInFlight.WithLabelValues(c.Request.Method, path).Dec()
		
		// Process request
		c.Next()
		
		// Record metrics
		duration := time.Since(start).Seconds()
		statusCode := fmt.Sprintf("%d", c.Writer.Status())
		
		m.httpRequestsTotal.WithLabelValues(c.Request.Method, path, statusCode).Inc()
		m.httpRequestDuration.WithLabelValues(c.Request.Method, path).Observe(duration)
	}
}

// RecordRepositoryOperation records a repository operation metric
func (m *MetricsService) RecordRepositoryOperation(repository, operation, artifactType, status string, duration time.Duration) {
	m.repositoryOperationsTotal.WithLabelValues(repository, operation, artifactType, status).Inc()
	m.repositoryOperationDuration.WithLabelValues(repository, operation, artifactType).Observe(duration.Seconds())
}

// RecordArtifactUpload records an artifact upload metric
func (m *MetricsService) RecordArtifactUpload(repository, artifactType string) {
	m.artifactUploadsTotal.WithLabelValues(repository, artifactType).Inc()
}

// RecordArtifactDownload records an artifact download metric
func (m *MetricsService) RecordArtifactDownload(repository, artifactType string) {
	m.artifactDownloadsTotal.WithLabelValues(repository, artifactType).Inc()
}

// SetArtifactStorageSize sets the storage size metric for a repository
func (m *MetricsService) SetArtifactStorageSize(repository, artifactType string, sizeBytes float64) {
	m.artifactStorageSize.WithLabelValues(repository, artifactType).Set(sizeBytes)
}

// RecordDatabaseQuery records a database query metric
func (m *MetricsService) RecordDatabaseQuery(operation, status string, duration time.Duration) {
	m.databaseQueriesTotal.WithLabelValues(operation, status).Inc()
	m.databaseQueryDuration.WithLabelValues(operation).Observe(duration.Seconds())
}

// SetDatabaseConnections sets the active database connections metric
func (m *MetricsService) SetDatabaseConnections(count float64) {
	m.databaseConnectionsActive.Set(count)
}

// SetSystemInfo sets system information metrics
func (m *MetricsService) SetSystemInfo(version, goVersion string) {
	m.systemInfo.WithLabelValues(version, goVersion).Set(1)
}

// RecordUptime records uptime metric
func (m *MetricsService) RecordUptime(instance string, seconds float64) {
	m.uptime.WithLabelValues(instance).Add(seconds)
}

// MetricsServer represents a separate metrics server
type MetricsServer struct {
	server  *http.Server
	metrics *MetricsService
}

// NewMetricsServer creates a new metrics server on a separate port
func NewMetricsServer(port int, metrics *MetricsService) *MetricsServer {
	mux := http.NewServeMux()
	mux.Handle("/metrics", metrics.GetHandler())
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}
	
	return &MetricsServer{
		server:  server,
		metrics: metrics,
	}
}

// Start starts the metrics server
func (ms *MetricsServer) Start() error {
	return ms.server.ListenAndServe()
}

// Shutdown gracefully shuts down the metrics server
func (ms *MetricsServer) Shutdown(ctx context.Context) error {
	return ms.server.Shutdown(ctx)
}
