# Implementation Plan: Transparent Proxy Mode

## Overview

This implementation plan transforms elastauth into a dual-mode authentication proxy. The approach is incremental, building on the existing authentication infrastructure while adding transparent proxy capabilities. Each phase includes testing tasks to ensure correctness.

## Tasks

- [x] 1. Add proxy configuration structure and validation
  - Create `ProxyConfig` struct in `libs/config.go`
  - Add proxy configuration defaults
  - Add environment variable bindings for proxy settings
  - Implement `ValidateProxyConfiguration` function
  - Add proxy configuration to `GetEffectiveConfig` response
  - _Requirements: 7.1, 7.2, 7.3, 7.4, 7.5_

- [ ]* 1.1 Write property test for configuration precedence
  - **Property 15: Configuration Loading**
  - **Validates: Requirements 7.1, 7.2, 7.3, 7.5**

- [ ]* 1.2 Write property test for configuration validation
  - **Property 16: Configuration Validation**
  - **Validates: Requirements 7.4**

- [ ] 2. Implement HTTP client for Elasticsearch connections
  - Create `NewElasticsearchHTTPClient` function in new file `libs/proxy_client.go`
  - Implement connection pooling configuration
  - Implement TLS configuration support
  - Add timeout configuration
  - _Requirements: 5.1, 5.2, 5.5_

- [ ]* 2.1 Write property test for timeout configuration
  - **Property 11: Timeout Configuration**
  - **Validates: Requirements 5.5**

- [ ]* 2.2 Write property test for connection pool configuration
  - **Property 22: Connection Pool Configuration**
  - **Validates: Requirements 11.4**

- [ ] 3. Create proxy handler structure
  - Create new file `libs/proxy_handler.go`
  - Implement `ProxyHandler` struct
  - Implement `NewProxyHandler` constructor
  - Add basic ServeHTTP method skeleton
  - _Requirements: 2.1, 2.2_

- [ ] 4. Implement authentication in proxy handler
  - Implement `authenticate` method in `ProxyHandler`
  - Reuse existing `getAuthProvider` function
  - Implement `GetOrGenerateCredentials` method
  - Add credential caching logic
  - _Requirements: 2.1, 6.1, 6.3_

- [ ]* 4.1 Write property test for credential caching
  - **Property 12: Credential Caching**
  - **Validates: Requirements 6.1, 6.3**

- [ ]* 4.2 Write property test for credential encryption
  - **Property 14: Credential Encryption**
  - **Validates: Requirements 6.5, 12.2**

- [ ] 5. Implement request proxying
  - Implement `modifyRequest` method to inject credentials
  - Create `httputil.ReverseProxy` instance
  - Implement request forwarding logic
  - Preserve original request method, path, query, and body
  - _Requirements: 2.2, 2.4, 6.2_

- [ ]* 5.1 Write property test for request preservation
  - **Property 5: Request Preservation During Proxying**
  - **Validates: Requirements 2.4**

- [ ]* 5.2 Write property test for credential injection
  - **Property 13: Credential Injection and Security**
  - **Validates: Requirements 6.2, 12.1**

- [ ] 6. Implement response forwarding
  - Implement `modifyResponse` method
  - Forward response status, headers, and body
  - Implement response streaming
  - _Requirements: 2.5_

- [ ]* 6.1 Write property test for response forwarding
  - **Property 6: Response Forwarding**
  - **Validates: Requirements 2.5**

- [ ] 7. Implement error handling in proxy handler
  - Implement `handleProxyError` method
  - Handle authentication failures (401/403)
  - Handle Elasticsearch unreachable (502)
  - Handle Elasticsearch timeout (504)
  - Add comprehensive error logging
  - _Requirements: 2.3, 9.1, 9.2, 9.3_

- [ ]* 7.1 Write property test for authentication failure
  - **Property 4: Authentication Failure Prevents Proxying**
  - **Validates: Requirements 2.3, 8.3**

- [ ]* 7.2 Write property test for error logging
  - **Property 18: Error Logging Context**
  - **Validates: Requirements 9.1, 9.2, 9.3, 9.4**

- [ ] 8. Create request router with mode detection
  - Create new file `libs/router.go`
  - Implement `Router` struct with mode detection
  - Implement `isSpecialPath` method
  - Implement `Route` method for request routing
  - _Requirements: 1.1, 1.2, 4.1, 4.2, 4.3, 4.4, 4.5_

- [ ]* 8.1 Write property test for mode selection
  - **Property 1: Mode Selection Based on Configuration**
  - **Validates: Requirements 1.1, 1.2**

- [ ]* 8.2 Write property test for special path bypass
  - **Property 8: Special Path Bypass**
  - **Validates: Requirements 4.1, 4.2**

- [ ]* 8.3 Write property test for default proxy behavior
  - **Property 9: Default Proxy Behavior**
  - **Validates: Requirements 4.3**

- [ ]* 8.4 Write property test for auth-only mode path handling
  - **Property 10: Auth-Only Mode Path Handling**
  - **Validates: Requirements 4.4, 4.5**

- [ ] 9. Integrate router into webserver
  - Update `WebserverInit` in `libs/webserver.go`
  - Initialize router based on proxy configuration
  - Wire up auth-only handler and proxy handler
  - Update route registration to use router
  - _Requirements: 1.3, 1.4_

- [ ]* 9.1 Write property test for authentication headers in auth-only mode
  - **Property 2: Authentication Headers in Auth-Only Mode**
  - **Validates: Requirements 1.3**

- [ ]* 9.2 Write property test for authentication and proxying in proxy mode
  - **Property 3: Authentication and Proxying in Proxy Mode**
  - **Validates: Requirements 1.4, 2.1, 2.2**

- [ ] 10. Update health checks for proxy mode
  - Update `ReadinessRoute` to check Elasticsearch in proxy mode
  - Update `checkElasticsearchReadiness` to respect proxy mode
  - Add proxy mode information to health check responses
  - Ensure liveness check is independent of Elasticsearch status
  - _Requirements: 10.1, 10.2, 10.3, 10.4, 10.5_

- [ ]* 10.1 Write property test for liveness check independence
  - **Property 20: Liveness Check Independence**
  - **Validates: Requirements 10.4**

- [ ]* 10.2 Write property test for health check mode information
  - **Property 21: Health Check Mode Information**
  - **Validates: Requirements 10.5**

- [ ] 11. Implement log sanitization
  - Update `SanitizeForLogging` in `libs/utils.go`
  - Add sanitization for proxy-specific fields
  - Ensure credentials never appear in logs
  - Add request ID to all log entries
  - _Requirements: 9.4, 9.5, 12.3_

- [ ]* 11.1 Write property test for log sanitization
  - **Property 19: Log Sanitization**
  - **Validates: Requirements 9.5, 12.3**

- [ ] 12. Implement input validation
  - Create `ValidateProxyRequest` function
  - Validate headers, query parameters, and body
  - Sanitize inputs before proxying
  - Prevent injection attacks
  - _Requirements: 12.4_

- [ ]* 12.1 Write property test for input validation
  - **Property 24: Input Validation**
  - **Validates: Requirements 12.4**

- [ ] 13. Add proxy metrics
  - Add proxy-specific metrics to Prometheus
  - Track request count, latency, and error rate
  - Track cache hit/miss rate
  - Track Elasticsearch connection pool stats
  - _Requirements: 11.5_

- [ ]* 13.1 Write property test for metrics exposure
  - **Property 23: Metrics Exposure**
  - **Validates: Requirements 11.5**

- [ ] 14. Update configuration documentation
  - Update `config.yml` example with proxy configuration
  - Document all proxy environment variables
  - Add proxy mode examples
  - Update README with proxy mode usage
  - _Requirements: 1.5, 7.1, 7.2, 7.3, 7.5_

- [ ] 15. Create comprehensive usage documentation
  - Create `docs/src/content/docs/deployment/auth-only-mode.md`
    - Document auth-only mode (Traefik forward auth scenario)
    - Include Traefik middleware configuration examples
    - Document chaining with Authelia
    - Add troubleshooting section
  - Create `docs/src/content/docs/deployment/proxy-mode.md`
    - Document transparent proxy mode (direct proxy scenario)
    - Include configuration examples
    - Document TLS setup
    - Add performance tuning guide
    - Add troubleshooting section
  - Update `docs/src/content/docs/index.mdx` with mode selection guide
  - _Requirements: 1.1, 1.2, 1.3, 1.4, 8.1, 8.2, 8.3, 8.4, 8.5_

- [ ] 16. Create Traefik integration example (auth-only mode)
  - Create `deployment/example/traefik-auth-only/`
  - Create `deployment/example/traefik-auth-only/docker-compose.yml`
    - Include Traefik with forward auth middleware
    - Include elastauth in auth-only mode
    - Include Authelia for authentication
    - Include Elasticsearch as protected backend
  - Create `deployment/example/traefik-auth-only/traefik.yml` (static config)
  - Create `deployment/example/traefik-auth-only/dynamic-config.yml` (dynamic config)
  - Create `deployment/example/traefik-auth-only/README.md`
    - Document the architecture
    - Document how to start and test
    - Document the request flow
  - _Requirements: 8.1, 8.2, 8.3, 8.4, 8.5_

- [ ] 17. Create direct proxy mode example
  - Create `deployment/example/direct-proxy/`
  - Create `deployment/example/direct-proxy/docker-compose.yml`
    - Include elastauth in proxy mode
    - Include Authelia for authentication
    - Include Elasticsearch as backend
    - No Traefik (elastauth is the proxy)
  - Create `deployment/example/direct-proxy/config.yml` (elastauth config)
  - Create `deployment/example/direct-proxy/README.md`
    - Document the architecture
    - Document how to start and test
    - Document the request flow
    - Compare with Traefik scenario
  - _Requirements: 1.4, 2.1, 2.2, 2.4, 2.5_

- [ ] 18. Update main deployment example
  - Update `deployment/example/docker-compose.yml`
    - Add commented-out Traefik service
    - Add comments explaining both modes
    - Add elastauth proxy mode configuration (commented)
  - Update `deployment/example/README.md`
    - Document both deployment scenarios
    - Add decision guide (when to use which mode)
    - Link to detailed mode-specific docs
  - _Requirements: 1.1, 1.2, 8.1_

- [ ] 19. Write integration tests
  - Create `libs/proxy_integration_test.go`
  - Test auth-only mode with mock Elasticsearch
  - Test proxy mode with mock Elasticsearch
  - Test provider switching (Authelia and OIDC)
  - Test error scenarios (Elasticsearch down, invalid auth)
  - _Requirements: 2.6, 3.1, 3.2, 3.3, 3.4, 3.5_

- [ ]* 19.1 Write property test for provider consistency
  - **Property 7: Provider Support in Both Modes**
  - **Validates: Requirements 3.1, 3.2, 3.5**

- [ ] 20. Checkpoint - Ensure all tests pass
  - Run `go build` to verify compilation
  - Run `go test ./...` to verify all tests pass
  - Manually test auth-only mode
  - Manually test proxy mode
  - Ensure all tests pass, ask the user if questions arise.

- [ ] 21. Update OpenAPI specification
  - Update `docs/src/schemas/openapi.yaml`
  - Add proxy mode endpoints
  - Document proxy configuration
  - Add proxy error responses
  - _Requirements: 1.1, 1.2, 1.3, 1.4_

- [ ] 22. Create architecture diagrams
  - Create `docs/src/content/docs/architecture/auth-only-mode.md`
    - Add Mermaid diagram showing Traefik → elastauth → Authelia flow
    - Document request/response flow
    - Document when to use this mode
  - Create `docs/src/content/docs/architecture/proxy-mode.md`
    - Add Mermaid diagram showing Client → elastauth → Elasticsearch flow
    - Document request/response flow with authentication
    - Document when to use this mode
  - Update `docs/src/content/docs/getting-started/concepts.md`
    - Add section on operating modes
    - Add comparison table
  - _Requirements: 1.1, 1.2, 1.3, 1.4_

- [ ] 23. Final checkpoint - Complete testing and validation
  - Run full test suite
  - Perform manual testing with real Elasticsearch
  - Test with Traefik forward auth
  - Test with direct proxy mode
  - Verify backward compatibility
  - Ensure all tests pass, ask the user if questions arise.

## Notes

- Tasks marked with `*` are optional and can be skipped for faster MVP
- Each task references specific requirements for traceability
- Checkpoints ensure incremental validation
- Property tests validate universal correctness properties
- Unit tests validate specific examples and edge cases
- Integration tests validate end-to-end scenarios
- The implementation maintains backward compatibility (default is auth-only mode)
- Proxy mode is opt-in via configuration
