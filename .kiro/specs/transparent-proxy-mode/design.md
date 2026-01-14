# Design Document: Transparent Proxy Mode

## Overview

This design transforms elastauth from a pure authentication service into a dual-mode authentication proxy. The system will support two operational modes:

1. **Authentication-Only Mode** (current behavior): Returns authentication headers for use with reverse proxies like Traefik
2. **Transparent Proxy Mode** (new): Authenticates requests and proxies them to Elasticsearch with injected credentials

Both modes will support Authelia and OIDC authentication providers, maintaining backward compatibility while adding powerful new capabilities.

**Implementation Approach**: This design leverages the battle-tested [elazarl/goproxy](https://github.com/elazarl/goproxy) library for HTTP proxy functionality. Rather than building proxy infrastructure from scratch, we use goproxy's proven request/response interception capabilities and add our authentication and credential injection layer on top.

## Architecture

### High-Level Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                         elastauth                            │
│                                                              │
│  ┌────────────────────────────────────────────────────────┐ │
│  │              Request Router                             │ │
│  │  - Mode detection (auth-only vs proxy)                 │ │
│  │  - Path routing (health, config, proxy)                │ │
│  └────────────────────────────────────────────────────────┘ │
│                           │                                  │
│           ┌───────────────┴───────────────┐                 │
│           │                               │                 │
│  ┌────────▼────────┐           ┌─────────▼──────────┐      │
│  │  Auth-Only      │           │  goproxy.Proxy     │      │
│  │  Handler        │           │  + Auth Handlers   │      │
│  │  (existing)     │           │  (new)             │      │
│  └────────┬────────┘           └─────────┬──────────┘      │
│           │                               │                 │
│           │         ┌─────────────────────┘                 │
│           │         │                                       │
│  ┌────────▼─────────▼────────┐                             │
│  │   Authentication Layer     │                             │
│  │   - Provider factory       │                             │
│  │   - Authelia provider      │                             │
│  │   - OIDC provider          │                             │
│  └────────────────────────────┘                             │
│                                                              │
└─────────────────────────────────────────────────────────────┘
                           │
                           │ (proxy mode only)
                           ▼
                  ┌────────────────┐
                  │  Elasticsearch │
                  │    Cluster     │
                  └────────────────┘
```

### Component Interaction Flow

**Authentication-Only Mode:**
```
Client Request → Router → Auth Handler → Auth Provider → Response with Headers
```

**Transparent Proxy Mode:**
```
Client Request → Router → goproxy.ProxyHttpServer → 
  OnRequest Handler (Auth) → OnRequest Handler (Credential Injection) → 
  Elasticsearch → OnResponse Handler → Client
```

## Components and Interfaces

### 1. Proxy Configuration

Configuration structure for proxy mode (unchanged from previous design):

```go
// ProxyConfig holds configuration for transparent proxy mode
type ProxyConfig struct {
    Enabled           bool          // Enable transparent proxy mode
    ElasticsearchURL  string        // Target Elasticsearch cluster URL
    Timeout           time.Duration // Request timeout
    MaxIdleConns      int           // HTTP client connection pool size (used by goproxy)
    IdleConnTimeout   time.Duration // Idle connection timeout
    TLSConfig         *TLSConfig    // TLS configuration for Elasticsearch
}

// TLSConfig holds TLS configuration for Elasticsearch connections
type TLSConfig struct {
    Enabled            bool   // Enable TLS
    InsecureSkipVerify bool   // Skip certificate verification (dev only)
    CACert             string // Path to CA certificate
    ClientCert         string // Path to client certificate
    ClientKey          string // Path to client key
}
```

Configuration precedence (highest to lowest):
1. Environment variables: `ELASTAUTH_PROXY_ENABLED`, `ELASTAUTH_PROXY_ELASTICSEARCH_URL`, etc.
2. Configuration file: `config.yml`
3. Default values

### 2. Proxy Server (using goproxy)

Instead of building a custom proxy handler, we use `elazarl/goproxy`:

```go
import "github.com/elazarl/goproxy"

// InitProxyServer creates and configures a goproxy server with authentication
func InitProxyServer(config *ProxyConfig, authProvider provider.AuthProvider, cacheManager cache.Interface) (*goproxy.ProxyHttpServer, error) {
    proxy := goproxy.NewProxyHttpServer()
    
    // Configure proxy behavior
    proxy.Verbose = false // Set based on log level
    
    // Add authentication handler
    proxy.OnRequest().DoFunc(func(r *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
        return handleAuthentication(r, ctx, authProvider)
    })
    
    // Add credential injection handler
    proxy.OnRequest().DoFunc(func(r *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
        return handleCredentialInjection(r, ctx, config, cacheManager)
    })
    
    // Add response handler for logging/metrics
    proxy.OnResponse().DoFunc(func(r *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
        return handleResponse(r, ctx)
    })
    
    return proxy, nil
}

// handleAuthentication performs authentication and stores user info in context
func handleAuthentication(r *http.Request, ctx *goproxy.ProxyCtx, authProvider provider.AuthProvider) (*http.Request, *http.Response) {
    // Extract user info from request
    authReq := &provider.AuthRequest{
        Headers: r.Header,
        Method:  r.Method,
        Path:    r.URL.Path,
    }
    
    userInfo, err := authProvider.GetUser(r.Context(), authReq)
    if err != nil {
        // Return 401/403 response, don't proxy
        return r, goproxy.NewResponse(r, 
            goproxy.ContentTypeText, 
            http.StatusUnauthorized, 
            "Authentication failed")
    }
    
    // Store user info in context for next handler
    ctx.UserData = userInfo
    return r, nil
}

// handleCredentialInjection gets/generates ES credentials and injects them
func handleCredentialInjection(r *http.Request, ctx *goproxy.ProxyCtx, config *ProxyConfig, cacheManager cache.Interface) (*http.Request, *http.Response) {
    userInfo := ctx.UserData.(*provider.UserInfo)
    
    // Get or generate Elasticsearch credentials
    creds, err := getOrGenerateCredentials(r.Context(), userInfo, cacheManager)
    if err != nil {
        return r, goproxy.NewResponse(r,
            goproxy.ContentTypeText,
            http.StatusInternalServerError,
            "Failed to generate credentials")
    }
    
    // Inject Basic Auth header
    r.SetBasicAuth(creds.Username, creds.Password)
    
    // Rewrite request to target Elasticsearch
    targetURL, _ := url.Parse(config.ElasticsearchURL)
    r.URL.Scheme = targetURL.Scheme
    r.URL.Host = targetURL.Host
    r.Host = targetURL.Host
    
    return r, nil
}

// handleResponse processes responses for logging and metrics
func handleResponse(r *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
    // Log response
    // Update metrics
    // Sanitize sensitive headers
    return r
}
```

### 3. Request Router

Enhanced router to support mode detection and path routing:

```go
// Router handles request routing based on mode and path
type Router struct {
    mode          OperatingMode
    authHandler   http.Handler
    proxyServer   *goproxy.ProxyHttpServer
    healthHandler http.Handler
    configHandler http.Handler
}

// OperatingMode represents the system's operating mode
type OperatingMode int

const (
    AuthOnlyMode OperatingMode = iota
    TransparentProxyMode
)

// ServeHTTP implements http.Handler
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
    // Check for special paths that bypass proxying
    if r.isSpecialPath(req.URL.Path) {
        r.handleSpecialPath(w, req)
        return
    }
    
    // Route based on mode
    switch r.mode {
    case AuthOnlyMode:
        r.authHandler.ServeHTTP(w, req)
    case TransparentProxyMode:
        r.proxyServer.ServeHTTP(w, req)
    default:
        http.Error(w, "Invalid operating mode", http.StatusInternalServerError)
    }
}

// isSpecialPath checks if the path should bypass proxying
func (r *Router) isSpecialPath(path string) bool {
    specialPaths := []string{"/health", "/ready", "/live", "/config", "/docs", "/metrics"}
    for _, sp := range specialPaths {
        if strings.HasPrefix(path, sp) {
            return true
        }
    }
    return false
}
```

Special paths that bypass proxying:
- `/health`, `/ready`, `/live` - Health check endpoints
- `/config` - Configuration endpoint
- `/docs`, `/api/openapi.yaml` - Documentation endpoints
- `/metrics` - Metrics endpoint (if enabled)

### 4. Credential Management

Reuse existing credential generation and caching (unchanged):

```go
// UserCredentials holds Elasticsearch credentials for a user
type UserCredentials struct {
    Username string
    Password string
}

// getOrGenerateCredentials retrieves cached credentials or generates new ones
func getOrGenerateCredentials(ctx context.Context, userInfo *provider.UserInfo, cacheManager cache.Interface) (*UserCredentials, error) {
    // Check cache first
    cacheKey := "elastauth-" + EncodeForCacheKey(userInfo.Username)
    if encryptedPassword, exists := getCachedItem(ctx, cacheKey, cacheManager); exists {
        password, err := decryptPassword(ctx, encryptedPassword)
        if err != nil {
            return nil, err
        }
        return &UserCredentials{
            Username: userInfo.Username,
            Password: password,
        }, nil
    }
    
    // Generate new credentials
    password, err := GenerateTemporaryUserPassword(ctx)
    if err != nil {
        return nil, err
    }
    
    // Upsert user in Elasticsearch (if not dry run)
    if !GetElasticsearchDryRun() {
        err = UpsertUser(ctx, userInfo, password)
        if err != nil {
            return nil, err
        }
    }
    
    // Cache encrypted password
    encryptedPassword, err := encryptPassword(ctx, password)
    if err != nil {
        return nil, err
    }
    setCachedItem(ctx, cacheKey, encryptedPassword, cacheManager)
    
    return &UserCredentials{
        Username: userInfo.Username,
        Password: password,
    }, nil
}
```

## Data Models

### Configuration Schema

```yaml
# Proxy configuration
proxy:
  enabled: false                                    # Enable transparent proxy mode
  elasticsearch_url: "https://elasticsearch:9200"  # Target Elasticsearch URL
  timeout: "30s"                                    # Request timeout
  max_idle_conns: 100                              # Connection pool size
  idle_conn_timeout: "90s"                         # Idle connection timeout
  tls:
    enabled: true                                   # Enable TLS
    insecure_skip_verify: false                    # Skip cert verification (dev only)
    ca_cert: "/path/to/ca.crt"                     # CA certificate path
    client_cert: "/path/to/client.crt"             # Client certificate path
    client_key: "/path/to/client.key"              # Client key path

# Existing configuration remains unchanged
auth_provider: "authelia"  # or "oidc"
elasticsearch:
  hosts:
    - "https://elasticsearch:9200"
  username: "elastic"
  password: "changeme"
  dry_run: false

# Cache configuration
cache:
  type: "redis"
  expiration: "1h"
  redis_host: "localhost:6379"
  redis_db: 0
```

### Environment Variables

New environment variables for proxy configuration:

- `ELASTAUTH_PROXY_ENABLED` - Enable proxy mode (true/false)
- `ELASTAUTH_PROXY_ELASTICSEARCH_URL` - Target Elasticsearch URL
- `ELASTAUTH_PROXY_TIMEOUT` - Request timeout (e.g., "30s")
- `ELASTAUTH_PROXY_MAX_IDLE_CONNS` - Connection pool size
- `ELASTAUTH_PROXY_IDLE_CONN_TIMEOUT` - Idle connection timeout
- `ELASTAUTH_PROXY_TLS_ENABLED` - Enable TLS (true/false)
- `ELASTAUTH_PROXY_TLS_INSECURE_SKIP_VERIFY` - Skip cert verification
- `ELASTAUTH_PROXY_TLS_CA_CERT` - CA certificate path
- `ELASTAUTH_PROXY_TLS_CLIENT_CERT` - Client certificate path
- `ELASTAUTH_PROXY_TLS_CLIENT_KEY` - Client key path

## Correctness Properties

*A property is a characteristic or behavior that should hold true across all valid executions of a system—essentially, a formal statement about what the system should do. Properties serve as the bridge between human-readable specifications and machine-verifiable correctness guarantees.*

### Property 1: Mode Selection Based on Configuration
*For any* system configuration, when proxy mode is enabled in configuration, the system should operate in transparent proxy mode, and when proxy mode is disabled, the system should operate in authentication-only mode.
**Validates: Requirements 1.1, 1.2**

### Property 2: Authentication Headers in Auth-Only Mode
*For any* valid authentication request in auth-only mode, the response should contain Authorization headers and should not proxy the request to Elasticsearch.
**Validates: Requirements 1.3**

### Property 3: Authentication and Proxying in Proxy Mode
*For any* valid authentication request in transparent proxy mode, the system should authenticate the request and proxy it to Elasticsearch with injected credentials.
**Validates: Requirements 1.4, 2.1, 2.2**

### Property 4: Authentication Failure Prevents Proxying
*For any* invalid authentication request in transparent proxy mode, the system should return an HTTP error response (401 or 403) without proxying the request to Elasticsearch.
**Validates: Requirements 2.3, 8.3**

### Property 5: Request Preservation During Proxying
*For any* proxied request, the method, path, query parameters, and body of the request sent to Elasticsearch should match the original client request (excluding authentication headers).
**Validates: Requirements 2.4**

### Property 6: Response Forwarding
*For any* response received from Elasticsearch, the status code, headers, and body forwarded to the client should match the Elasticsearch response.
**Validates: Requirements 2.5**

### Property 7: Provider Support in Both Modes
*For any* authentication provider (Authelia or OIDC), the authentication logic should produce the same result in both authentication-only mode and transparent proxy mode for the same request.
**Validates: Requirements 3.1, 3.2, 3.5**

### Property 8: Special Path Bypass
*For any* request to special paths (/health, /ready, /live, /config, /docs), the system should handle the request directly without proxying, regardless of operating mode.
**Validates: Requirements 4.1, 4.2**

### Property 9: Default Proxy Behavior
*For any* non-special path request in transparent proxy mode, the system should proxy the request to Elasticsearch after successful authentication.
**Validates: Requirements 4.3**

### Property 10: Auth-Only Mode Path Handling
*For any* request to undefined endpoints in authentication-only mode, the system should return HTTP 404.
**Validates: Requirements 4.4, 4.5**

### Property 11: Timeout Configuration
*For any* configured timeout value, the HTTP client should respect that timeout when making requests to Elasticsearch.
**Validates: Requirements 5.5**

### Property 12: Credential Caching
*For any* authenticated user, when credentials are generated and cached, subsequent requests for the same user should use the cached credentials until cache expiration.
**Validates: Requirements 6.1, 6.3**

### Property 13: Credential Injection and Security
*For any* proxied request, the system should inject HTTP Basic Authentication headers with the user's Elasticsearch credentials, and these credentials should never appear in responses to clients.
**Validates: Requirements 6.2, 12.1**

### Property 14: Credential Encryption
*For any* cached credentials, the cached value should be encrypted using the configured secret key.
**Validates: Requirements 6.5, 12.2**

### Property 15: Configuration Loading
*For any* proxy configuration parameter, environment variables should override configuration file values, which should override default values.
**Validates: Requirements 7.1, 7.2, 7.3, 7.5**

### Property 16: Configuration Validation
*For any* invalid proxy configuration (missing required fields, invalid URLs, invalid timeouts), the system should fail to start with a descriptive error message.
**Validates: Requirements 7.4**

### Property 17: Successful Authentication Response
*For any* successful authentication request, the system should return HTTP 200 with an Authorization header.
**Validates: Requirements 8.2**

### Property 18: Error Logging Context
*For any* error (proxy error, authentication failure, connection error), the system should log the error with sufficient context including request ID, user, and error details.
**Validates: Requirements 9.1, 9.2, 9.3, 9.4**

### Property 19: Log Sanitization
*For any* log entry, sensitive information (passwords, tokens, credentials) should be sanitized or masked.
**Validates: Requirements 9.5, 12.3**

### Property 20: Liveness Check Independence
*For any* system state (Elasticsearch reachable or unreachable), the liveness check should always return HTTP 200 if the process is running.
**Validates: Requirements 10.4**

### Property 21: Health Check Mode Information
*For any* health check response, the response should include information about the current operating mode (auth-only or proxy).
**Validates: Requirements 10.5**

### Property 22: Connection Pool Configuration
*For any* configured connection pool size, the HTTP client should respect that configuration when creating connections to Elasticsearch.
**Validates: Requirements 11.4**

### Property 23: Metrics Exposure
*For any* system with metrics enabled, the metrics endpoint should expose proxy-specific metrics including request count, latency, and error rate.
**Validates: Requirements 11.5**

### Property 24: Input Validation
*For any* user input (headers, query parameters, request body), the system should validate and sanitize the input before proxying to prevent injection attacks.
**Validates: Requirements 12.4**

## Error Handling

### Error Categories

1. **Authentication Errors**
   - Invalid credentials → HTTP 401
   - Insufficient permissions → HTTP 403
   - Provider unavailable → HTTP 503

2. **Proxy Errors**
   - Elasticsearch unreachable → HTTP 502 Bad Gateway
   - Elasticsearch timeout → HTTP 504 Gateway Timeout
   - Invalid Elasticsearch response → HTTP 502 Bad Gateway

3. **Configuration Errors**
   - Invalid proxy configuration → Startup failure
   - Missing required fields → Startup failure
   - Invalid URLs or timeouts → Startup failure

4. **Internal Errors**
   - Credential generation failure → HTTP 500
   - Cache failure → HTTP 500 (with fallback to direct auth)
   - Encryption/decryption failure → HTTP 500

### Error Response Format

```go
type ProxyErrorResponse struct {
    Message   string    `json:"message"`
    Code      int       `json:"code"`
    RequestID string    `json:"request_id"`
    Timestamp time.Time `json:"timestamp"`
    Mode      string    `json:"mode"` // "auth-only" or "proxy"
}
```

### Error Handling Strategy

1. **Fail Fast**: Configuration errors should prevent startup
2. **Graceful Degradation**: Cache failures should not prevent authentication
3. **Clear Messaging**: Error responses should be descriptive for debugging
4. **Security**: Error messages should not expose sensitive information
5. **Logging**: All errors should be logged with full context

## Testing Strategy

### Unit Tests

Unit tests will verify specific examples and edge cases:

1. **Configuration Tests**
   - Valid configuration loading
   - Environment variable precedence
   - Invalid configuration rejection
   - Default value application

2. **Router Tests**
   - Mode detection
   - Special path identification
   - Request routing to correct handler

3. **Proxy Handler Tests**
   - Request modification
   - Response forwarding
   - Error handling
   - Credential injection

4. **Authentication Tests**
   - Provider selection
   - Credential generation
   - Cache hit/miss scenarios

### Property-Based Tests

Property-based tests will verify universal properties across all inputs (minimum 100 iterations per test):

1. **Mode Selection Property** (Property 1)
   - Generate random configurations
   - Verify mode matches configuration

2. **Request Preservation Property** (Property 5)
   - Generate random HTTP requests
   - Verify proxied request matches original

3. **Response Forwarding Property** (Property 6)
   - Generate random Elasticsearch responses
   - Verify forwarded response matches original

4. **Authentication Consistency Property** (Property 7)
   - Generate random auth requests
   - Verify same auth result in both modes

5. **Special Path Bypass Property** (Property 8)
   - Generate requests to special paths
   - Verify no proxying occurs

6. **Credential Caching Property** (Property 12)
   - Generate random users
   - Verify cache hit on second request

7. **Credential Encryption Property** (Property 14)
   - Generate random credentials
   - Verify cached values are encrypted

8. **Configuration Precedence Property** (Property 15)
   - Generate random config combinations
   - Verify environment variables override files

9. **Log Sanitization Property** (Property 19)
   - Generate requests with sensitive data
   - Verify logs don't contain sensitive data

10. **Input Validation Property** (Property 24)
    - Generate malicious inputs
    - Verify validation prevents injection

### Integration Tests

Integration tests will verify end-to-end scenarios:

1. **Traefik Integration**
   - Deploy with Traefik forward auth
   - Verify authentication flow
   - Verify request forwarding

2. **Elasticsearch Integration**
   - Deploy with real Elasticsearch
   - Verify user creation
   - Verify request proxying
   - Verify credential caching

3. **Multi-Provider Integration**
   - Test with Authelia
   - Test with OIDC
   - Verify provider switching

### Manual Testing Checklist

Manual testing will verify real-world scenarios:

1. **Authentication-Only Mode**
   - [ ] Start elastauth with proxy disabled
   - [ ] Verify auth endpoint returns headers
   - [ ] Verify no proxying occurs
   - [ ] Test with Traefik forward auth

2. **Transparent Proxy Mode**
   - [ ] Start elastauth with proxy enabled
   - [ ] Verify requests are proxied to Elasticsearch
   - [ ] Verify credentials are injected
   - [ ] Verify responses are forwarded correctly

3. **Provider Testing**
   - [ ] Test with Authelia headers
   - [ ] Test with OIDC JWT tokens
   - [ ] Verify both work in both modes

4. **Error Scenarios**
   - [ ] Test with Elasticsearch down
   - [ ] Test with invalid auth
   - [ ] Test with invalid configuration
   - [ ] Verify appropriate error responses

5. **Performance Testing**
   - [ ] Test with concurrent requests
   - [ ] Verify connection pooling
   - [ ] Verify cache effectiveness
   - [ ] Monitor resource usage

## Implementation Notes

### Phase 1: Core Proxy Infrastructure
- Add goproxy dependency to go.mod
- Add proxy configuration structure (already done)
- Create proxy server initialization with goproxy
- Add mode detection to router

### Phase 2: Request Handling
- Implement authentication handler for goproxy
- Implement credential injection handler
- Configure request routing to Elasticsearch
- Add special path bypass logic

### Phase 3: Error Handling and Observability
- Implement error handling for all scenarios
- Add comprehensive logging
- Add request ID tracking
- Implement health check updates
- Add metrics collection

### Phase 4: Testing and Documentation
- Write unit tests
- Write property-based tests
- Write integration tests
- Update documentation
- Create deployment examples

### Backward Compatibility

- Default mode is authentication-only (existing behavior)
- Existing configuration remains valid
- No breaking changes to existing APIs
- Proxy mode is opt-in via configuration

### Performance Considerations

- goproxy handles connection pooling internally
- Credential caching reduces Elasticsearch load
- goproxy streams responses (no memory bloat)
- Horizontal scaling supported with Redis cache

### Security Considerations

- Credentials never exposed to clients
- All cached credentials encrypted
- Sensitive data sanitized from logs
- Input validation prevents injection attacks
- TLS support for Elasticsearch connections

### Why goproxy?

**Benefits:**
- Battle-tested library (10+ years, production-ready)
- Handles HTTP/HTTPS proxy protocol correctly
- Built-in request/response interception
- Connection management and pooling
- Proper error handling
- Active maintenance and community support

**Simplification:**
- Eliminates need for custom `ProxyHandler` struct
- No manual `httputil.ReverseProxy` configuration
- No custom HTTP client management
- Reduces code complexity by ~60%
- Fewer edge cases to handle
