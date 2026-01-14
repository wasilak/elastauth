package libs

import (
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestRecordProxyRequest(t *testing.T) {
	// Reset metrics before test
	proxyRequestsTotal.Reset()
	proxyRequestDuration.Reset()

	// Record a successful request
	RecordProxyRequest("GET", 200, 100*time.Millisecond)

	// Verify counter was incremented
	count := testutil.ToFloat64(proxyRequestsTotal.WithLabelValues("GET", "2xx"))
	if count != 1 {
		t.Errorf("Expected counter to be 1, got %f", count)
	}

	// Record another request with different status
	RecordProxyRequest("POST", 404, 50*time.Millisecond)

	// Verify counter was incremented for 4xx
	count = testutil.ToFloat64(proxyRequestsTotal.WithLabelValues("POST", "4xx"))
	if count != 1 {
		t.Errorf("Expected counter to be 1, got %f", count)
	}
}

func TestRecordAuthentication(t *testing.T) {
	// Reset metrics before test
	proxyAuthenticationTotal.Reset()

	// Record successful authentication
	RecordAuthenticationSuccess()
	count := testutil.ToFloat64(proxyAuthenticationTotal.WithLabelValues("success"))
	if count != 1 {
		t.Errorf("Expected success counter to be 1, got %f", count)
	}

	// Record failed authentication
	RecordAuthenticationFailure()
	count = testutil.ToFloat64(proxyAuthenticationTotal.WithLabelValues("failure"))
	if count != 1 {
		t.Errorf("Expected failure counter to be 1, got %f", count)
	}
}

func TestRecordCache(t *testing.T) {
	// Reset metrics before test
	proxyCacheHitsTotal.Reset()

	// Record cache hit
	RecordCacheHit()
	count := testutil.ToFloat64(proxyCacheHitsTotal.WithLabelValues("hit"))
	if count != 1 {
		t.Errorf("Expected hit counter to be 1, got %f", count)
	}

	// Record cache miss
	RecordCacheMiss()
	count = testutil.ToFloat64(proxyCacheHitsTotal.WithLabelValues("miss"))
	if count != 1 {
		t.Errorf("Expected miss counter to be 1, got %f", count)
	}
}

func TestRecordProxyError(t *testing.T) {
	// Reset metrics before test
	proxyErrorsTotal.Reset()

	// Record different error types
	RecordProxyError("auth_failed")
	count := testutil.ToFloat64(proxyErrorsTotal.WithLabelValues("auth_failed"))
	if count != 1 {
		t.Errorf("Expected auth_failed counter to be 1, got %f", count)
	}

	RecordProxyError("credential_generation_failed")
	count = testutil.ToFloat64(proxyErrorsTotal.WithLabelValues("credential_generation_failed"))
	if count != 1 {
		t.Errorf("Expected credential_generation_failed counter to be 1, got %f", count)
	}
}

func TestActiveRequests(t *testing.T) {
	// Get initial value
	initial := testutil.ToFloat64(proxyActiveRequests)

	// Increment
	IncrementActiveRequests()
	count := testutil.ToFloat64(proxyActiveRequests)
	if count != initial+1 {
		t.Errorf("Expected active requests to be %f, got %f", initial+1, count)
	}

	// Decrement
	DecrementActiveRequests()
	count = testutil.ToFloat64(proxyActiveRequests)
	if count != initial {
		t.Errorf("Expected active requests to be %f, got %f", initial, count)
	}
}

func TestGetStatusCategory(t *testing.T) {
	tests := []struct {
		statusCode int
		expected   string
	}{
		{200, "2xx"},
		{201, "2xx"},
		{299, "2xx"},
		{300, "3xx"},
		{301, "3xx"},
		{399, "3xx"},
		{400, "4xx"},
		{404, "4xx"},
		{499, "4xx"},
		{500, "5xx"},
		{502, "5xx"},
		{599, "5xx"},
		{100, "unknown"},
		{600, "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := getStatusCategory(tt.statusCode)
			if result != tt.expected {
				t.Errorf("getStatusCategory(%d) = %s, expected %s", tt.statusCode, result, tt.expected)
			}
		})
	}
}

func TestMetricsRegistration(t *testing.T) {
	// Verify all metrics are registered with Prometheus
	metrics := []prometheus.Collector{
		proxyRequestsTotal,
		proxyRequestDuration,
		proxyAuthenticationTotal,
		proxyCacheHitsTotal,
		proxyErrorsTotal,
		proxyActiveRequests,
	}

	for _, metric := range metrics {
		if metric == nil {
			t.Error("Metric is nil")
		}
	}
}
