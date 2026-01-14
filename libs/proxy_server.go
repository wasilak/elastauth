package libs

import (
	"context"
	"encoding/base64"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/elazarl/goproxy"
	"github.com/spf13/viper"
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

	// Add authentication handler
	// This handler runs first and validates the request using the auth provider
	proxy.OnRequest().DoFunc(func(r *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
		return handleAuthentication(r, ctx, authProvider)
	})

	// Add credential injection handler
	// This handler runs second and injects Elasticsearch credentials into the request
	proxy.OnRequest().DoFunc(func(r *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
		return handleCredentialInjection(r, ctx, config, cacheManager)
	})

	slog.Info("Proxy server initialized",
		"elasticsearch_url", config.ElasticsearchURL,
		"timeout", config.Timeout,
		"max_idle_conns", config.MaxIdleConns,
		"tls_enabled", config.TLS.Enabled,
	)

	// TODO: Add response handler in task 5

	return proxy, nil
}

// handleAuthentication performs authentication and stores user info in context.
// It extracts user information using the configured auth provider and validates it.
// If authentication fails, it returns an appropriate HTTP error response (401/403).
// On success, it stores the UserInfo in ctx.UserData for use by subsequent handlers.
//
// Parameters:
//   - r: The incoming HTTP request
//   - ctx: The goproxy context for this request
//   - authProvider: The authentication provider to use for extracting user info
//
// Returns:
//   - *http.Request: The original request (unchanged)
//   - *http.Response: nil on success, or an error response (401/403) on failure
func handleAuthentication(r *http.Request, ctx *goproxy.ProxyCtx, authProvider provider.AuthProvider) (*http.Request, *http.Response) {
	// Create auth request from HTTP request
	authReq := &provider.AuthRequest{
		Request: r,
	}

	// Extract user info from request using auth provider
	userInfo, err := authProvider.GetUser(r.Context(), authReq)
	if err != nil {
		slog.ErrorContext(r.Context(), "Authentication failed",
			slog.String("error", err.Error()),
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.String("remote_addr", r.RemoteAddr),
		)

		// Return 401 Unauthorized response
		return r, goproxy.NewResponse(r,
			goproxy.ContentTypeText,
			http.StatusUnauthorized,
			"Authentication failed: "+err.Error())
	}

	// Validate username
	if err := ValidateUsername(userInfo.Username); err != nil {
		slog.ErrorContext(r.Context(), "Invalid username format",
			slog.String("error", err.Error()),
			slog.String("username", userInfo.Username),
		)

		// Return 400 Bad Request for invalid username format
		return r, goproxy.NewResponse(r,
			goproxy.ContentTypeText,
			http.StatusBadRequest,
			"Invalid username format: "+err.Error())
	}

	// Validate email if provided
	if userInfo.Email != "" {
		if err := ValidateEmail(userInfo.Email); err != nil {
			slog.ErrorContext(r.Context(), "Invalid email format",
				slog.String("error", err.Error()),
				slog.String("email", userInfo.Email),
			)

			// Return 400 Bad Request for invalid email format
			return r, goproxy.NewResponse(r,
				goproxy.ContentTypeText,
				http.StatusBadRequest,
				"Invalid email format: "+err.Error())
		}
	}

	// Validate full name if provided
	if userInfo.FullName != "" {
		if err := ValidateName(userInfo.FullName); err != nil {
			slog.ErrorContext(r.Context(), "Invalid name format",
				slog.String("error", err.Error()),
				slog.String("name", userInfo.FullName),
			)

			// Return 400 Bad Request for invalid name format
			return r, goproxy.NewResponse(r,
				goproxy.ContentTypeText,
				http.StatusBadRequest,
				"Invalid name format: "+err.Error())
		}
	}

	// Validate groups if provided
	if len(userInfo.Groups) > 0 {
		for _, group := range userInfo.Groups {
			if err := ValidateGroupName(group); err != nil {
				slog.ErrorContext(r.Context(), "Invalid group name",
					slog.String("error", err.Error()),
					slog.String("group", group),
				)

				// Return 400 Bad Request for invalid group name
				return r, goproxy.NewResponse(r,
					goproxy.ContentTypeText,
					http.StatusBadRequest,
					"Invalid group name: "+err.Error())
			}
		}
	}

	// Store user info in context for next handler
	ctx.UserData = userInfo

	slog.DebugContext(r.Context(), "Authentication successful",
		slog.String("username", userInfo.Username),
		slog.String("email", userInfo.Email),
		slog.Int("groups_count", len(userInfo.Groups)),
	)

	// Return nil response to continue proxying
	return r, nil
}

// handleCredentialInjection gets/generates ES credentials and injects them.
// It retrieves the authenticated user info from the proxy context, gets or generates
// Elasticsearch credentials (with caching), and injects them as Basic Auth headers.
// It also rewrites the request URL to target the configured Elasticsearch cluster.
//
// Parameters:
//   - r: The incoming HTTP request
//   - ctx: The goproxy context containing UserData from authentication handler
//   - config: Proxy configuration with Elasticsearch URL
//   - cacheManager: Cache interface for storing/retrieving credentials
//
// Returns:
//   - *http.Request: The modified request with credentials and rewritten URL
//   - *http.Response: nil on success, or an error response (500) on failure
func handleCredentialInjection(r *http.Request, ctx *goproxy.ProxyCtx, config *ProxyConfig, cacheManager cache.CacheInterface) (*http.Request, *http.Response) {
	// Retrieve user info from context (set by authentication handler)
	userInfo, ok := ctx.UserData.(*provider.UserInfo)
	if !ok || userInfo == nil {
		slog.ErrorContext(r.Context(), "User info not found in proxy context")
		return r, goproxy.NewResponse(r,
			goproxy.ContentTypeText,
			http.StatusInternalServerError,
			"Internal server error: user info not available")
	}

	// Get or generate Elasticsearch credentials
	credentials, err := getOrGenerateCredentials(r.Context(), userInfo, cacheManager)
	if err != nil {
		slog.ErrorContext(r.Context(), "Failed to get or generate credentials",
			slog.String("error", err.Error()),
			slog.String("username", userInfo.Username),
		)
		return r, goproxy.NewResponse(r,
			goproxy.ContentTypeText,
			http.StatusInternalServerError,
			"Failed to generate credentials: "+err.Error())
	}

	// Inject Basic Auth header
	r.SetBasicAuth(credentials.Username, credentials.Password)

	// Rewrite request URL to target Elasticsearch
	// Parse the target Elasticsearch URL
	targetURL, err := parseElasticsearchURL(config.ElasticsearchURL)
	if err != nil {
		slog.ErrorContext(r.Context(), "Failed to parse Elasticsearch URL",
			slog.String("error", err.Error()),
			slog.String("url", config.ElasticsearchURL),
		)
		return r, goproxy.NewResponse(r,
			goproxy.ContentTypeText,
			http.StatusInternalServerError,
			"Internal server error: invalid Elasticsearch URL")
	}

	// Rewrite the request URL to point to Elasticsearch
	r.URL.Scheme = targetURL.Scheme
	r.URL.Host = targetURL.Host
	r.Host = targetURL.Host

	// Preserve the original path and query parameters
	// The path is already set in r.URL.Path from the original request

	slog.DebugContext(r.Context(), "Credential injection successful",
		slog.String("username", credentials.Username),
		slog.String("target_host", r.URL.Host),
		slog.String("path", r.URL.Path),
	)

	// Return nil response to continue proxying
	return r, nil
}

// handleResponse processes responses for logging and metrics
// This will be implemented in task 5
func handleResponse(r *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
	// Implementation will be added in task 5
	return r
}

// UserCredentials holds Elasticsearch credentials for a user
type UserCredentials struct {
	Username string
	Password string
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
			slog.DebugContext(ctx, "Cache miss for credentials", slog.String("username", userInfo.Username))
		}
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
