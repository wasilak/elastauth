package libs

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/elazarl/goproxy"
	"github.com/wasilak/elastauth/cache"
	"github.com/wasilak/elastauth/provider"
)

// InitProxyServer creates and configures a goproxy server with authentication
// and credential injection handlers. It returns a configured proxy server ready
// to handle HTTP requests.
//
// The proxy server is configured with:
// - Authentication handler that validates requests using the auth provider
// - Credential injection handler that adds Elasticsearch credentials
// - Response handler for logging and metrics
//
// Parameters:
//   - config: Proxy configuration including Elasticsearch URL and timeouts
//   - authProvider: Authentication provider for validating requests
//   - cacheManager: Cache interface for storing/retrieving credentials
//
// Returns:
//   - *goproxy.ProxyHttpServer: Configured proxy server
//   - error: Any error encountered during initialization
func InitProxyServer(config *ProxyConfig, authProvider provider.AuthProvider, cacheManager cache.CacheInterface) (*goproxy.ProxyHttpServer, error) {
	if config == nil {
		return nil, fmt.Errorf("proxy config cannot be nil")
	}

	if authProvider == nil {
		return nil, fmt.Errorf("auth provider cannot be nil")
	}

	if cacheManager == nil {
		return nil, fmt.Errorf("cache manager cannot be nil")
	}

	// Create new proxy server
	proxy := goproxy.NewProxyHttpServer()

	// Configure proxy behavior based on log level
	// Set verbose mode to false for production, can be made configurable later
	proxy.Verbose = false

	slog.Info("Proxy server initialized",
		"elasticsearch_url", config.ElasticsearchURL,
		"timeout", config.Timeout,
		"max_idle_conns", config.MaxIdleConns,
		"tls_enabled", config.TLS.Enabled,
	)

	// TODO: Add authentication handler in task 3
	// TODO: Add credential injection handler in task 4
	// TODO: Add response handler in task 5

	return proxy, nil
}

// handleAuthentication performs authentication and stores user info in context
// This will be implemented in task 3
func handleAuthentication(r *http.Request, ctx *goproxy.ProxyCtx, authProvider provider.AuthProvider) (*http.Request, *http.Response) {
	// Implementation will be added in task 3
	return r, nil
}

// handleCredentialInjection gets/generates ES credentials and injects them
// This will be implemented in task 4
func handleCredentialInjection(r *http.Request, ctx *goproxy.ProxyCtx, config *ProxyConfig, cacheManager cache.CacheInterface) (*http.Request, *http.Response) {
	// Implementation will be added in task 4
	return r, nil
}

// handleResponse processes responses for logging and metrics
// This will be implemented in task 5
func handleResponse(r *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
	// Implementation will be added in task 5
	return r
}
