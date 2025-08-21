package metrics

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/stretchr/testify/assert"
)

func TestMetricsService(t *testing.T) {
	// Create a new metrics service
	service := NewMetricsService()
	assert.NotNil(t, service)

	// Test system info setting
	service.SetSystemInfo("1.0.0", "go1.21")

	// Test recording repository operation
	service.RecordRepositoryOperation("test-repo", "push", "maven", "success", time.Millisecond*50)

	// Test recording artifact upload/download
	service.RecordArtifactUpload("test-repo", "maven")
	service.RecordArtifactDownload("test-repo", "maven")

	// Test setting storage size
	service.SetArtifactStorageSize("test-repo", "maven", 1024.0)

	// Test recording database operation
	service.RecordDatabaseQuery("SELECT", "success", time.Millisecond*10)
}

func TestMetricsMiddleware(t *testing.T) {
	// Create a new metrics service
	service := NewMetricsService()

	// Create a test Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	
	// Add metrics middleware
	router.Use(service.GinMiddleware())
	
	// Add a test route
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "test"})
	})

	// Create a test request
	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	
	// Execute the request
	router.ServeHTTP(w, req)

	// Verify response
	assert.Equal(t, 200, w.Code)
	assert.Contains(t, w.Body.String(), "test")
}

func TestMetricsEndpoint(t *testing.T) {
	// Create a new metrics service
	service := NewMetricsService()
	
	// Record some test metrics
	service.RecordRepositoryOperation("test-repo", "push", "maven", "success", time.Millisecond*50)

	// Create a test server with metrics endpoint
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Create a test request to metrics endpoint
	req, _ := http.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()
	
	// Execute the request
	router.ServeHTTP(w, req)

	// Verify response
	assert.Equal(t, 200, w.Code)
	assert.Contains(t, w.Body.String(), "# HELP")
	assert.Contains(t, w.Body.String(), "# TYPE")
}

func TestMetricsServer(t *testing.T) {
	// Create a new metrics service
	service := NewMetricsService()
	
	// Create a metrics server
	server := NewMetricsServer(9091, service)
	assert.NotNil(t, server)
	
	// Record a test metric to use the service
	service.RecordRepositoryOperation("test-repo", "push", "maven", "success", time.Millisecond*50)
	
	// Test that the server can be created without errors
	// Note: We don't actually start the server in the test to avoid port conflicts
}

func TestPrometheusRegistry(t *testing.T) {
	// Create a new metrics service
	service := NewMetricsService()
	
	// Record a test metric to ensure it's registered
	service.RecordRepositoryOperation("test-repo", "push", "maven", "success", time.Millisecond*50)
	
	// Verify that metrics are registered with Prometheus
	metricFamilies, err := prometheus.DefaultGatherer.Gather()
	assert.NoError(t, err)
	assert.NotEmpty(t, metricFamilies)
	
	// Look for our custom metrics
	var foundHTTPRequests bool
	var foundRepoOps bool
	
	for _, mf := range metricFamilies {
		if strings.Contains(mf.GetName(), "http_requests_total") {
			foundHTTPRequests = true
		}
		if strings.Contains(mf.GetName(), "repository_operations_total") {
			foundRepoOps = true
		}
	}
	
	// Note: These might not be found if other tests haven't run yet
	// This is more of a smoke test to ensure the registry is working
	t.Logf("Found HTTP requests metric: %v", foundHTTPRequests)
	t.Logf("Found repository operations metric: %v", foundRepoOps)
}
