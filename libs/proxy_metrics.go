package libs

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// proxyRequestsTotal tracks the total number of proxy requests
	proxyRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "elastauth_proxy_requests_total",
			Help: "Total number of proxy requests",
		},
		[]string{"method", "status"},
	)

	// proxyRequestDuration tracks the duration of proxy requests
	proxyRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "elastauth_proxy_request_duration_seconds",
			Help:    "Duration of proxy requests in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "status"},
	)

	// proxyAuthenticationTotal tracks authentication attempts
	proxyAuthenticationTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "elastauth_proxy_authentication_total",
			Help: "Total number of authentication attempts",
		},
		[]string{"result"}, // "success" or "failure"
	)

	// proxyCacheHitsTotal tracks cache hits and misses
	proxyCacheHitsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "elastauth_proxy_cache_hits_total",
			Help: "Total number of cache hits and misses",
		},
		[]string{"result"}, // "hit" or "miss"
	)

	// proxyErrorsTotal tracks proxy errors by type
	proxyErrorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "elastauth_proxy_errors_total",
			Help: "Total number of proxy errors",
		},
		[]string{"error_type"}, // "auth_failed", "credential_generation_failed", "elasticsearch_unreachable", etc.
	)

	// proxyActiveRequests tracks currently active proxy requests
	proxyActiveRequests = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "elastauth_proxy_active_requests",
			Help: "Number of currently active proxy requests",
		},
	)
)

// RecordProxyRequest records metrics for a completed proxy request
func RecordProxyRequest(method string, statusCode int, duration time.Duration) {
	status := getStatusCategory(statusCode)
	proxyRequestsTotal.WithLabelValues(method, status).Inc()
	proxyRequestDuration.WithLabelValues(method, status).Observe(duration.Seconds())
}

// RecordAuthenticationSuccess records a successful authentication
func RecordAuthenticationSuccess() {
	proxyAuthenticationTotal.WithLabelValues("success").Inc()
}

// RecordAuthenticationFailure records a failed authentication
func RecordAuthenticationFailure() {
	proxyAuthenticationTotal.WithLabelValues("failure").Inc()
}

// RecordCacheHit records a cache hit
func RecordCacheHit() {
	proxyCacheHitsTotal.WithLabelValues("hit").Inc()
}

// RecordCacheMiss records a cache miss
func RecordCacheMiss() {
	proxyCacheHitsTotal.WithLabelValues("miss").Inc()
}

// RecordProxyError records a proxy error by type
func RecordProxyError(errorType string) {
	proxyErrorsTotal.WithLabelValues(errorType).Inc()
}

// IncrementActiveRequests increments the active requests counter
func IncrementActiveRequests() {
	proxyActiveRequests.Inc()
}

// DecrementActiveRequests decrements the active requests counter
func DecrementActiveRequests() {
	proxyActiveRequests.Dec()
}

// getStatusCategory converts HTTP status code to category for metrics
func getStatusCategory(statusCode int) string {
	switch {
	case statusCode >= 200 && statusCode < 300:
		return "2xx"
	case statusCode >= 300 && statusCode < 400:
		return "3xx"
	case statusCode >= 400 && statusCode < 500:
		return "4xx"
	case statusCode >= 500 && statusCode < 600:
		return "5xx"
	default:
		return "unknown"
	}
}
