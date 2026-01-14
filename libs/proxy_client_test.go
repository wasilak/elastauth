package libs

import (
	"crypto/tls"
	"net/http"
	"testing"
	"time"
)

func TestNewElasticsearchHTTPClient(t *testing.T) {
	tests := []struct {
		name      string
		config    *ProxyConfig
		wantError bool
		validate  func(*testing.T, *http.Client)
	}{
		{
			name:      "nil config returns error",
			config:    nil,
			wantError: true,
		},
		{
			name: "basic config without TLS",
			config: &ProxyConfig{
				Enabled:          true,
				ElasticsearchURL: "http://localhost:9200",
				Timeout:          30 * time.Second,
				MaxIdleConns:     100,
				IdleConnTimeout:  90 * time.Second,
			},
			wantError: false,
			validate: func(t *testing.T, client *http.Client) {
				if client.Timeout != 30*time.Second {
					t.Errorf("expected timeout 30s, got %v", client.Timeout)
				}
				transport := client.Transport.(*http.Transport)
				if transport.MaxIdleConns != 100 {
					t.Errorf("expected MaxIdleConns 100, got %d", transport.MaxIdleConns)
				}
				if transport.IdleConnTimeout != 90*time.Second {
					t.Errorf("expected IdleConnTimeout 90s, got %v", transport.IdleConnTimeout)
				}
			},
		},
		{
			name: "config with TLS disabled",
			config: &ProxyConfig{
				Enabled:          true,
				ElasticsearchURL: "https://localhost:9200",
				Timeout:          15 * time.Second,
				MaxIdleConns:     50,
				IdleConnTimeout:  60 * time.Second,
				TLS: TLSConfig{
					Enabled: false,
				},
			},
			wantError: false,
			validate: func(t *testing.T, client *http.Client) {
				transport := client.Transport.(*http.Transport)
				if transport.TLSClientConfig != nil {
					t.Error("expected no TLS config when TLS is disabled")
				}
			},
		},
		{
			name: "config with TLS enabled but no certs",
			config: &ProxyConfig{
				Enabled:          true,
				ElasticsearchURL: "https://localhost:9200",
				Timeout:          30 * time.Second,
				MaxIdleConns:     100,
				IdleConnTimeout:  90 * time.Second,
				TLS: TLSConfig{
					Enabled:            true,
					InsecureSkipVerify: false,
				},
			},
			wantError: false,
			validate: func(t *testing.T, client *http.Client) {
				transport := client.Transport.(*http.Transport)
				if transport.TLSClientConfig == nil {
					t.Error("expected TLS config when TLS is enabled")
				}
				if transport.TLSClientConfig.InsecureSkipVerify {
					t.Error("expected InsecureSkipVerify to be false")
				}
			},
		},
		{
			name: "config with InsecureSkipVerify enabled",
			config: &ProxyConfig{
				Enabled:          true,
				ElasticsearchURL: "https://localhost:9200",
				Timeout:          30 * time.Second,
				MaxIdleConns:     100,
				IdleConnTimeout:  90 * time.Second,
				TLS: TLSConfig{
					Enabled:            true,
					InsecureSkipVerify: true,
				},
			},
			wantError: false,
			validate: func(t *testing.T, client *http.Client) {
				transport := client.Transport.(*http.Transport)
				if !transport.TLSClientConfig.InsecureSkipVerify {
					t.Error("expected InsecureSkipVerify to be true")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewElasticsearchHTTPClient(tt.config)
			if (err != nil) != tt.wantError {
				t.Errorf("NewElasticsearchHTTPClient() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if !tt.wantError && tt.validate != nil {
				tt.validate(t, client)
			}
		})
	}
}

func TestBuildTLSConfig(t *testing.T) {
	tests := []struct {
		name      string
		config    *TLSConfig
		wantError bool
		validate  func(*testing.T, *tls.Config)
	}{
		{
			name:      "nil config returns error",
			config:    nil,
			wantError: true,
		},
		{
			name: "basic TLS config",
			config: &TLSConfig{
				Enabled:            true,
				InsecureSkipVerify: false,
			},
			wantError: false,
			validate: func(t *testing.T, tlsConfig *tls.Config) {
				if tlsConfig.InsecureSkipVerify {
					t.Error("expected InsecureSkipVerify to be false")
				}
				if tlsConfig.MinVersion != tls.VersionTLS12 {
					t.Errorf("expected MinVersion TLS 1.2, got %d", tlsConfig.MinVersion)
				}
			},
		},
		{
			name: "TLS config with InsecureSkipVerify",
			config: &TLSConfig{
				Enabled:            true,
				InsecureSkipVerify: true,
			},
			wantError: false,
			validate: func(t *testing.T, tlsConfig *tls.Config) {
				if !tlsConfig.InsecureSkipVerify {
					t.Error("expected InsecureSkipVerify to be true")
				}
			},
		},
		{
			name: "TLS config with non-existent CA cert",
			config: &TLSConfig{
				Enabled:            true,
				InsecureSkipVerify: false,
				CACert:             "/nonexistent/ca.crt",
			},
			wantError: true,
		},
		{
			name: "TLS config with only client cert (missing key)",
			config: &TLSConfig{
				Enabled:            true,
				InsecureSkipVerify: false,
				ClientCert:         "/path/to/cert.crt",
			},
			wantError: true,
		},
		{
			name: "TLS config with only client key (missing cert)",
			config: &TLSConfig{
				Enabled:            true,
				InsecureSkipVerify: false,
				ClientKey:          "/path/to/key.key",
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tlsConfig, err := buildTLSConfig(tt.config)
			if (err != nil) != tt.wantError {
				t.Errorf("buildTLSConfig() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if !tt.wantError && tt.validate != nil {
				tt.validate(t, tlsConfig)
			}
		})
	}
}

func TestDefaultProxyConfig(t *testing.T) {
	config := DefaultProxyConfig()

	if config == nil {
		t.Fatal("DefaultProxyConfig() returned nil")
	}

	if config.Enabled {
		t.Error("expected Enabled to be false by default")
	}

	if config.ElasticsearchURL != "" {
		t.Errorf("expected empty ElasticsearchURL, got %s", config.ElasticsearchURL)
	}

	if config.Timeout != 30*time.Second {
		t.Errorf("expected Timeout 30s, got %v", config.Timeout)
	}

	if config.MaxIdleConns != 100 {
		t.Errorf("expected MaxIdleConns 100, got %d", config.MaxIdleConns)
	}

	if config.IdleConnTimeout != 90*time.Second {
		t.Errorf("expected IdleConnTimeout 90s, got %v", config.IdleConnTimeout)
	}

	if config.TLS.Enabled {
		t.Error("expected TLS.Enabled to be false by default")
	}

	if config.TLS.InsecureSkipVerify {
		t.Error("expected TLS.InsecureSkipVerify to be false by default")
	}
}

func TestHTTPClientRedirectHandling(t *testing.T) {
	config := &ProxyConfig{
		Enabled:          true,
		ElasticsearchURL: "http://localhost:9200",
		Timeout:          30 * time.Second,
		MaxIdleConns:     100,
		IdleConnTimeout:  90 * time.Second,
	}

	client, err := NewElasticsearchHTTPClient(config)
	if err != nil {
		t.Fatalf("NewElasticsearchHTTPClient() error = %v", err)
	}

	// Test that CheckRedirect returns ErrUseLastResponse
	err = client.CheckRedirect(nil, nil)
	if err != http.ErrUseLastResponse {
		t.Errorf("expected CheckRedirect to return ErrUseLastResponse, got %v", err)
	}
}
