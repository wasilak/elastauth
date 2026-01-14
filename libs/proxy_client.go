package libs

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"os"
	"time"
)

// NewElasticsearchHTTPClient creates an HTTP client optimized for Elasticsearch connections.
// It configures connection pooling, timeouts, and TLS based on the provided ProxyConfig.
//
// The client is configured with:
// - Connection pooling for efficient connection reuse
// - Configurable timeouts for request handling
// - Optional TLS configuration for secure connections
// - HTTP/2 support when available
// - Redirect handling (passes redirects through to client)
//
// Requirements: 5.1, 5.2, 5.5
func NewElasticsearchHTTPClient(config *ProxyConfig) (*http.Client, error) {
	if config == nil {
		return nil, fmt.Errorf("proxy config cannot be nil")
	}

	// Create transport with connection pooling configuration
	transport := &http.Transport{
		MaxIdleConns:        config.MaxIdleConns,
		MaxIdleConnsPerHost: config.MaxIdleConns,
		IdleConnTimeout:     config.IdleConnTimeout,
		DisableCompression:  false,
		ForceAttemptHTTP2:   true,
	}

	// Configure TLS if enabled
	if config.TLS.Enabled {
		tlsConfig, err := buildTLSConfig(&config.TLS)
		if err != nil {
			return nil, fmt.Errorf("failed to build TLS config: %w", err)
		}
		transport.TLSClientConfig = tlsConfig
	}

	// Create HTTP client with configured transport and timeout
	client := &http.Client{
		Transport: transport,
		Timeout:   config.Timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// Don't follow redirects - pass them through to client
			return http.ErrUseLastResponse
		},
	}

	return client, nil
}

// buildTLSConfig creates a TLS configuration from the provided TLSConfig.
// It supports:
// - Custom CA certificates for server verification
// - Client certificates for mutual TLS authentication
// - InsecureSkipVerify for development/testing (not recommended for production)
//
// Requirements: 5.2, 12.5
func buildTLSConfig(config *TLSConfig) (*tls.Config, error) {
	if config == nil {
		return nil, fmt.Errorf("TLS config cannot be nil")
	}

	tlsConfig := &tls.Config{
		InsecureSkipVerify: config.InsecureSkipVerify,
		MinVersion:         tls.VersionTLS12, // Enforce minimum TLS 1.2
	}

	// Load CA certificate if provided
	if config.CACert != "" {
		caCert, err := os.ReadFile(config.CACert)
		if err != nil {
			return nil, fmt.Errorf("failed to read CA certificate from %s: %w", config.CACert, err)
		}

		caCertPool := x509.NewCertPool()
		if !caCertPool.AppendCertsFromPEM(caCert) {
			return nil, fmt.Errorf("failed to parse CA certificate from %s", config.CACert)
		}
		tlsConfig.RootCAs = caCertPool
	}

	// Load client certificate and key if provided (for mutual TLS)
	if config.ClientCert != "" && config.ClientKey != "" {
		cert, err := tls.LoadX509KeyPair(config.ClientCert, config.ClientKey)
		if err != nil {
			return nil, fmt.Errorf("failed to load client certificate and key: %w", err)
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
	} else if config.ClientCert != "" || config.ClientKey != "" {
		// Both cert and key must be provided together
		return nil, fmt.Errorf("both client_cert and client_key must be provided for mutual TLS")
	}

	return tlsConfig, nil
}

// DefaultProxyConfig returns a ProxyConfig with sensible defaults.
// These defaults are suitable for most Elasticsearch deployments.
func DefaultProxyConfig() *ProxyConfig {
	return &ProxyConfig{
		Enabled:          false,
		ElasticsearchURL: "",
		Timeout:          30 * time.Second,
		MaxIdleConns:     100,
		IdleConnTimeout:  90 * time.Second,
		TLS: TLSConfig{
			Enabled:            false,
			InsecureSkipVerify: false,
		},
	}
}
