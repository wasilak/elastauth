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

	// Add authentication handler
	// This handler runs first and validates the request using the auth provider
	proxy.OnRequest().DoFunc(func(r *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
		return handleAuthentication(r, ctx, authProvider)
	})

	slog.Info("Proxy server initialized",
		"elasticsearch_url", config.ElasticsearchURL,
		"timeout", config.Timeout,
		"max_idle_conns", config.MaxIdleConns,
		"tls_enabled", config.TLS.Enabled,
	)

	// TODO: Add credential injection handler in task 4
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
