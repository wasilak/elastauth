package libs

import (
	"context"
	"encoding/base64"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/spf13/viper"
	"github.com/wasilak/elastauth/cache"
	"github.com/wasilak/elastauth/provider"
)

// ProxyHandler wraps httputil.ReverseProxy with authentication and credential injection
type ProxyHandler struct {
	proxy        *httputil.ReverseProxy
	authProvider provider.AuthProvider
	cacheManager cache.CacheInterface
	config       *ProxyConfig
}

// ServeHTTP implements http.Handler interface
func (h *ProxyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	IncrementActiveRequests()
	defer DecrementActiveRequests()

	// Step 1: Authenticate the request
	authReq := &provider.AuthRequest{Request: r}
	userInfo, err := h.authProvider.GetUser(r.Context(), authReq)
	if err != nil {
		RecordAuthenticationFailure()
		RecordProxyError("auth_failed")
		
		slog.ErrorContext(r.Context(), "Authentication failed",
			slog.String("error", err.Error()),
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
		)
		
		http.Error(w, "Authentication failed: "+err.Error(), http.StatusUnauthorized)
		return
	}

	RecordAuthenticationSuccess()

	// Step 2: Validate user input
	if err := ValidateUsername(userInfo.Username); err != nil {
		RecordProxyError("invalid_username")
		http.Error(w, "Invalid username: "+err.Error(), http.StatusBadRequest)
		return
	}

	if userInfo.Email != "" {
		if err := ValidateEmail(userInfo.Email); err != nil {
			RecordProxyError("invalid_email")
			http.Error(w, "Invalid email: "+err.Error(), http.StatusBadRequest)
			return
		}
	}

	// Step 3: Validate request
	if err := ValidateProxyRequest(r); err != nil {
		RecordProxyError("invalid_request")
		http.Error(w, "Invalid request: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Step 4: Get or generate credentials
	credentials, err := getOrGenerateCredentials(r.Context(), userInfo, h.cacheManager)
	if err != nil {
		RecordProxyError("credential_generation_failed")
		
		slog.ErrorContext(r.Context(), "Failed to get credentials",
			slog.String("error", err.Error()),
			slog.String("username", userInfo.Username),
		)
		
		http.Error(w, "Failed to generate credentials", http.StatusInternalServerError)
		return
	}

	// Step 5: Inject credentials into the request
	r.SetBasicAuth(credentials.Username, credentials.Password)

	// Step 6: Modify the request to add custom director logic
	originalDirector := h.proxy.Director
	h.proxy.Director = func(req *http.Request) {
		// Call original director to set up the proxy request
		originalDirector(req)
		
		// Ensure credentials are set (in case director overwrites headers)
		req.SetBasicAuth(credentials.Username, credentials.Password)
		
		slog.DebugContext(req.Context(), "Proxying request",
			slog.String("username", userInfo.Username),
			slog.String("method", req.Method),
			slog.String("path", req.URL.Path),
			slog.String("target_host", req.URL.Host),
		)
	}

	// Step 7: Modify response handler
	h.proxy.ModifyResponse = func(resp *http.Response) error {
		duration := time.Since(startTime)
		RecordProxyRequest(r.Method, resp.StatusCode, duration)
		
		slog.InfoContext(r.Context(), "Proxy response",
			slog.String("username", userInfo.Username),
			slog.Int("status_code", resp.StatusCode),
			slog.Duration("duration", duration),
		)
		
		// Sanitize response headers
		sanitizeResponseHeadersInPlace(resp)
		
		return nil
	}

	// Step 8: Proxy the request
	h.proxy.ServeHTTP(w, r)
}

// InitProxyServer creates and configures a reverse proxy with authentication
// and credential injection. It returns a configured HTTP handler ready
// to handle requests and proxy them to Elasticsearch.
//
// The proxy handler:
// - Authenticates requests using the auth provider
// - Generates/retrieves Elasticsearch credentials
// - Injects credentials into proxied requests
// - Logs responses and collects metrics
//
// Parameters:
//   - config: Proxy configuration including Elasticsearch URL and timeouts
//   - authProvider: Authentication provider for validating requests
//   - cacheManager: Cache interface for storing/retrieving credentials
//
// Returns:
//   - http.Handler: Configured proxy handler
//   - error: Any error encountered during initialization
func InitProxyServer(config *ProxyConfig, authProvider provider.AuthProvider, cacheManager cache.CacheInterface) (http.Handler, error) {
	if config == nil {
		return nil, fmt.Errorf("proxy config cannot be nil")
	}

	if authProvider == nil {
		return nil, fmt.Errorf("auth provider cannot be nil")
	}

	if cacheManager == nil {
		return nil, fmt.Errorf("cache manager cannot be nil")
	}

	// Parse target Elasticsearch URL
	targetURL, err := parseElasticsearchURL(config.ElasticsearchURL)
	if err != nil {
		return nil, fmt.Errorf("invalid Elasticsearch URL: %w", err)
	}

	// Create reverse proxy
	proxy := httputil.NewSingleHostReverseProxy(targetURL)

	// Configure transport with timeouts and connection pooling
	proxy.Transport = &http.Transport{
		MaxIdleConns:        config.MaxIdleConns,
		IdleConnTimeout:     config.IdleConnTimeout,
		TLSHandshakeTimeout: 10 * time.Second,
		// TLS configuration would go here if needed
	}

	// Create proxy handler
	handler := &ProxyHandler{
		proxy:        proxy,
		authProvider: authProvider,
		cacheManager: cacheManager,
		config:       config,
	}

	slog.Info("Proxy server initialized",
		"elasticsearch_url", config.ElasticsearchURL,
		"timeout", config.Timeout,
		"max_idle_conns", config.MaxIdleConns,
		"tls_enabled", config.TLS.Enabled,
	)

	return handler, nil
}

// UserCredentials holds Elasticsearch credentials for a user
type UserCredentials struct {
	Username string
	Password string
}

// sanitizeResponseHeadersInPlace removes or redacts sensitive headers from the response.
// This modifies the response headers in place to prevent sensitive data from reaching the client.
//
// Parameters:
//   - r: The HTTP response to sanitize
func sanitizeResponseHeadersInPlace(r *http.Response) {
	if r == nil || r.Header == nil {
		return
	}

	// Remove sensitive headers that should not be forwarded to the client
	// These are headers that might contain Elasticsearch credentials or internal information
	headersToRemove := []string{
		"WWW-Authenticate",   // Elasticsearch auth challenges
		"Proxy-Authenticate", // Proxy auth challenges
		"X-Elastic-Product",  // Internal Elasticsearch header (optional, but good practice)
	}

	for _, header := range headersToRemove {
		r.Header.Del(header)
	}
}

// getOrGenerateCredentials retrieves cached credentials or generates new ones.
// It checks the cache first, and if credentials are not found, generates new ones,
// upserts the user in Elasticsearch (if not in dry run mode), and caches the encrypted password.
//
// Parameters:
//   - ctx: Context for the request
//   - userInfo: User information from authentication provider
//   - cacheManager: Cache interface for storing/retrieving credentials
//
// Returns:
//   - *UserCredentials: The user's Elasticsearch credentials
//   - error: Any error encountered during credential generation or caching
func getOrGenerateCredentials(ctx context.Context, userInfo *provider.UserInfo, cacheManager cache.CacheInterface) (*UserCredentials, error) {
	// Build cache key
	cacheKey := "elastauth-" + EncodeForCacheKey(userInfo.Username)
	
	// Get encryption key from configuration
	key := viper.GetString("secret_key")
	if key == "" {
		return nil, fmt.Errorf("secret_key not configured")
	}

	// Check cache first
	if cacheManager != nil {
		encryptedPasswordBase64, exists := cacheManager.Get(ctx, cacheKey)
		if exists {
			RecordCacheHit()
			
			slog.DebugContext(ctx, "Cache hit for credentials", slog.String("username", userInfo.Username))
			
			// Decode from base64
			decryptedPasswordBase64, err := base64.URLEncoding.DecodeString(encryptedPasswordBase64.(string))
			if err != nil {
				slog.WarnContext(ctx, "Failed to decode cached password, regenerating",
					slog.String("error", err.Error()),
					slog.String("username", userInfo.Username),
				)
			} else {
				// Decrypt password
				password, err := Decrypt(ctx, string(decryptedPasswordBase64), key)
				if err != nil {
					slog.WarnContext(ctx, "Failed to decrypt cached password, regenerating",
						slog.String("error", err.Error()),
						slog.String("username", userInfo.Username),
					)
				} else {
					// Successfully retrieved from cache
					return &UserCredentials{
						Username: userInfo.Username,
						Password: password,
					}, nil
				}
			}
		} else {
			RecordCacheMiss()
			
			slog.DebugContext(ctx, "Cache miss for credentials", slog.String("username", userInfo.Username))
		}
	} else {
		RecordCacheMiss()
	}

	// Generate new credentials
	slog.DebugContext(ctx, "Generating new credentials", slog.String("username", userInfo.Username))
	
	password, err := GenerateTemporaryUserPassword(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to generate password: %w", err)
	}

	// Get user roles based on groups
	roles := GetUserRoles(ctx, userInfo.Groups)

	// Create Elasticsearch user object
	elasticsearchUserMetadata := ElasticsearchUserMetadata{
		Groups: userInfo.Groups,
	}

	elasticsearchUser := ElasticsearchUser{
		Password: password,
		Enabled:  true,
		Email:    userInfo.Email,
		FullName: userInfo.FullName,
		Roles:    roles,
		Metadata: elasticsearchUserMetadata,
	}

	// Upsert user in Elasticsearch (if not dry run)
	if !GetElasticsearchDryRun() {
		// Initialize Elasticsearch client
		hosts := GetElasticsearchHosts()
		if len(hosts) == 0 {
			return nil, fmt.Errorf("no Elasticsearch hosts configured")
		}

		err := initElasticClient(
			ctx,
			hosts,
			GetElasticsearchUsername(),
			GetElasticsearchPassword(),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize Elasticsearch client: %w", err)
		}

		err = UpsertUser(ctx, userInfo.Username, elasticsearchUser)
		if err != nil {
			return nil, fmt.Errorf("failed to upsert user in Elasticsearch: %w", err)
		}
	}

	// Cache encrypted password
	if cacheManager != nil {
		encryptedPassword, err := Encrypt(ctx, password, key)
		if err != nil {
			slog.WarnContext(ctx, "Failed to encrypt password for caching",
				slog.String("error", err.Error()),
				slog.String("username", userInfo.Username),
			)
		} else {
			encryptedPasswordBase64 := base64.URLEncoding.EncodeToString([]byte(encryptedPassword))
			cacheManager.Set(ctx, cacheKey, encryptedPasswordBase64)
			slog.DebugContext(ctx, "Cached encrypted credentials", slog.String("username", userInfo.Username))
		}
	}

	return &UserCredentials{
		Username: userInfo.Username,
		Password: password,
	}, nil
}

// parseElasticsearchURL parses and validates the Elasticsearch URL from configuration.
// It returns a parsed URL structure that can be used to rewrite request URLs.
//
// Parameters:
//   - elasticsearchURL: The Elasticsearch URL from configuration
//
// Returns:
//   - *url.URL: Parsed URL structure
//   - error: Any error encountered during parsing
func parseElasticsearchURL(elasticsearchURL string) (*url.URL, error) {
	if elasticsearchURL == "" {
		return nil, fmt.Errorf("elasticsearch URL is empty")
	}

	parsedURL, err := url.Parse(elasticsearchURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	if parsedURL.Scheme == "" {
		return nil, fmt.Errorf("URL scheme is missing (must be http or https)")
	}

	if parsedURL.Host == "" {
		return nil, fmt.Errorf("URL host is missing")
	}

	return parsedURL, nil
}
