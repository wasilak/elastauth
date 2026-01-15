# Manual Testing Checklist - Transparent Proxy Mode

## Overview

This checklist covers comprehensive manual testing for the transparent proxy mode implementation. Complete all sections to verify the implementation works correctly in real-world scenarios.

## Prerequisites

- Docker and Docker Compose V2 installed
- Elasticsearch cluster available (or use provided examples)
- Authelia configured (or use provided examples)
- elastauth binary built (`go build -o elastauth`)

---

## 1. Authentication-Only Mode Testing

### 1.1 Basic Auth-Only Mode

**Setup:**
```bash
cd deployment/example/traefik-auth-only
docker compose up -d
```

**Tests:**

- [ ] **Test 1.1.1**: Verify elastauth starts in auth-only mode
  ```bash
  docker compose logs elastauth | grep "mode"
  # Expected: Should show auth-only mode or no proxy mode enabled
  ```

- [ ] **Test 1.1.2**: Test authentication with valid Authelia headers
  ```bash
  curl -H "Remote-User: john" \
       -H "Remote-Email: john@example.com" \
       -H "Remote-Name: John Doe" \
       -H "Remote-Groups: admin,users" \
       http://localhost:5000/
  ```
  - Expected: HTTP 200 with Authorization header in response

- [ ] **Test 1.1.3**: Test authentication failure (missing headers)
  ```bash
  curl -v http://localhost:5000/
  ```
  - Expected: HTTP 401 or 400 (missing username)

- [ ] **Test 1.1.4**: Verify no proxying occurs in auth-only mode
  ```bash
  curl -H "Remote-User: john" \
       -H "Remote-Email: john@example.com" \
       http://localhost:5000/_search
  ```
  - Expected: Returns auth headers, does NOT proxy to Elasticsearch

**Cleanup:**
```bash
docker compose down -v
```

---

## 2. Transparent Proxy Mode Testing

### 2.1 Basic Proxy Mode

**Setup:**
```bash
cd deployment/example/direct-proxy
docker compose up -d
```

**Tests:**

- [ ] **Test 2.1.1**: Verify elastauth starts in proxy mode
  ```bash
  docker compose logs elastauth | grep "Proxy server initialized"
  # Expected: Should show proxy initialization with Elasticsearch URL
  ```

- [ ] **Test 2.1.2**: Test proxied request with authentication
  ```bash
  curl -H "Remote-User: john" \
       -H "Remote-Email: john@example.com" \
       -H "Remote-Name: John Doe" \
       -H "Remote-Groups: admin" \
       http://localhost:5000/_cluster/health
  ```
  - Expected: HTTP 200 with Elasticsearch cluster health response
  - Verify: Check elastauth logs for proxy activity

- [ ] **Test 2.1.3**: Test authentication failure prevents proxying
  ```bash
  curl -v http://localhost:5000/_cluster/health
  ```
  - Expected: HTTP 401, request NOT proxied to Elasticsearch

- [ ] **Test 2.1.4**: Test request preservation (method, path, query params)
  ```bash
  curl -H "Remote-User: john" \
       -H "Remote-Email: john@example.com" \
       "http://localhost:5000/_search?q=test&size=10"
  ```
  - Expected: Query parameters preserved in proxied request
  - Verify: Check Elasticsearch logs for the exact query

- [ ] **Test 2.1.5**: Test POST request with body
  ```bash
  curl -X POST \
       -H "Remote-User: john" \
       -H "Remote-Email: john@example.com" \
       -H "Content-Type: application/json" \
       -d '{"query":{"match_all":{}}}' \
       http://localhost:5000/_search
  ```
  - Expected: Request body preserved and proxied correctly

**Cleanup:**
```bash
docker compose down -v
```

---

## 3. Traefik Forward Auth Integration

### 3.1 Traefik + elastauth + Authelia

**Setup:**
```bash
cd deployment/example/traefik-auth-only
docker compose up -d
```

**Tests:**

- [ ] **Test 3.1.1**: Verify Traefik forward auth middleware works
  ```bash
  # Access Elasticsearch through Traefik
  curl -v http://localhost:8080/_cluster/health
  ```
  - Expected: Redirected to Authelia login or 401 if not authenticated

- [ ] **Test 3.1.2**: Test authenticated request through Traefik
  ```bash
  # First authenticate with Authelia (manual browser test)
  # Then test with session cookie
  curl -b cookies.txt http://localhost:8080/_cluster/health
  ```
  - Expected: Request authenticated by Authelia, forwarded by Traefik, authorized by elastauth

- [ ] **Test 3.1.3**: Verify elastauth receives correct headers from Traefik
  ```bash
  docker compose logs elastauth | grep "Remote-User"
  ```
  - Expected: Should see Authelia headers in elastauth logs

- [ ] **Test 3.1.4**: Test middleware chain (Authelia → elastauth → Elasticsearch)
  - Verify: Check logs of all three services for request flow

**Cleanup:**
```bash
docker compose down -v
```

---

## 4. Special Paths Bypass Testing

**Setup:** Use either proxy mode example

**Tests:**

- [ ] **Test 4.1**: Health endpoint bypasses proxy
  ```bash
  curl http://localhost:5000/health
  ```
  - Expected: Returns health status, NOT proxied to Elasticsearch

- [ ] **Test 4.2**: Readiness endpoint bypasses proxy
  ```bash
  curl http://localhost:5000/ready
  ```
  - Expected: Returns readiness status

- [ ] **Test 4.3**: Liveness endpoint bypasses proxy
  ```bash
  curl http://localhost:5000/live
  ```
  - Expected: Returns liveness status

- [ ] **Test 4.4**: Config endpoint bypasses proxy
  ```bash
  curl http://localhost:5000/config
  ```
  - Expected: Returns configuration info

- [ ] **Test 4.5**: Metrics endpoint bypasses proxy (if enabled)
  ```bash
  curl http://localhost:5000/metrics
  ```
  - Expected: Returns Prometheus metrics

---

## 5. Credential Caching Testing

**Setup:** Use proxy mode example with Redis cache

**Tests:**

- [ ] **Test 5.1**: First request generates credentials
  ```bash
  # Clear cache first
  docker compose exec redis redis-cli FLUSHALL
  
  # Make request
  curl -H "Remote-User: testuser" \
       -H "Remote-Email: test@example.com" \
       http://localhost:5000/_cluster/health
  ```
  - Verify: Check elastauth logs for "cache miss" or credential generation

- [ ] **Test 5.2**: Second request uses cached credentials
  ```bash
  curl -H "Remote-User: testuser" \
       -H "Remote-Email: test@example.com" \
       http://localhost:5000/_cluster/health
  ```
  - Verify: Check elastauth logs for "cache hit"
  - Verify: Check Redis for cached entry

- [ ] **Test 5.3**: Cache expiration regenerates credentials
  ```bash
  # Wait for cache TTL to expire (or manually delete from Redis)
  docker compose exec redis redis-cli DEL "elastauth-testuser"
  
  # Make request
  curl -H "Remote-User: testuser" \
       -H "Remote-Email: test@example.com" \
       http://localhost:5000/_cluster/health
  ```
  - Verify: New credentials generated

---

## 6. Error Handling Testing

**Tests:**

- [ ] **Test 6.1**: Elasticsearch unreachable
  ```bash
  # Stop Elasticsearch
  docker compose stop elasticsearch
  
  # Make request
  curl -H "Remote-User: john" \
       -H "Remote-Email: john@example.com" \
       http://localhost:5000/_cluster/health
  ```
  - Expected: HTTP 502 Bad Gateway
  - Verify: Error logged with context

- [ ] **Test 6.2**: Elasticsearch timeout
  ```bash
  # Configure short timeout in elastauth config
  # Make request to slow Elasticsearch endpoint
  ```
  - Expected: HTTP 504 Gateway Timeout

- [ ] **Test 6.3**: Invalid authentication
  ```bash
  curl -H "Remote-User: " \
       http://localhost:5000/_search
  ```
  - Expected: HTTP 400 or 401
  - Verify: Error logged without sensitive data

- [ ] **Test 6.4**: Readiness check fails when Elasticsearch down
  ```bash
  docker compose stop elasticsearch
  curl http://localhost:5000/ready
  ```
  - Expected: HTTP 503 Service Unavailable (in proxy mode)

- [ ] **Test 6.5**: Liveness check succeeds even when Elasticsearch down
  ```bash
  docker compose stop elasticsearch
  curl http://localhost:5000/live
  ```
  - Expected: HTTP 200 (liveness independent of Elasticsearch)

---

## 7. Backward Compatibility Testing

**Tests:**

- [ ] **Test 7.1**: Default mode is auth-only
  ```bash
  # Start elastauth without proxy configuration
  ./elastauth
  ```
  - Expected: Starts in auth-only mode (backward compatible)

- [ ] **Test 7.2**: Existing Authelia configuration works
  ```bash
  # Use old config.yml format (without proxy section)
  # Start elastauth
  ```
  - Expected: Works exactly as before, no breaking changes

- [ ] **Test 7.3**: Existing environment variables work
  ```bash
  export ELASTAUTH_ELASTICSEARCH_HOSTS="http://localhost:9200"
  export ELASTAUTH_ELASTICSEARCH_USERNAME="elastic"
  export ELASTAUTH_ELASTICSEARCH_PASSWORD="changeme"
  ./elastauth
  ```
  - Expected: Starts successfully with existing env vars

---

## 8. Provider Switching Testing

### 8.1 Authelia Provider

**Tests:**

- [ ] **Test 8.1.1**: Authelia works in auth-only mode
  - Use traefik-auth-only example
  - Verify authentication with Authelia headers

- [ ] **Test 8.1.2**: Authelia works in proxy mode
  - Use direct-proxy example with Authelia
  - Verify authentication and proxying

### 8.2 OIDC Provider

**Tests:**

- [ ] **Test 8.2.1**: OIDC works in auth-only mode
  ```bash
  # Configure OIDC provider
  # Send request with JWT token
  curl -H "Authorization: Bearer <jwt_token>" \
       http://localhost:5000/
  ```
  - Expected: Authentication succeeds, returns auth headers

- [ ] **Test 8.2.2**: OIDC works in proxy mode
  ```bash
  # Configure OIDC provider in proxy mode
  # Send request with JWT token
  curl -H "Authorization: Bearer <jwt_token>" \
       http://localhost:5000/_cluster/health
  ```
  - Expected: Authentication succeeds, request proxied

---

## 9. Performance and Scalability Testing

**Tests:**

- [ ] **Test 9.1**: Concurrent requests handling
  ```bash
  # Use Apache Bench or similar
  ab -n 1000 -c 10 \
     -H "Remote-User: john" \
     -H "Remote-Email: john@example.com" \
     http://localhost:5000/_cluster/health
  ```
  - Verify: All requests succeed
  - Verify: Check metrics for request count and latency

- [ ] **Test 9.2**: Large response streaming
  ```bash
  # Query that returns large result set
  curl -H "Remote-User: john" \
       -H "Remote-Email: john@example.com" \
       http://localhost:5000/_search?size=10000
  ```
  - Verify: Response streams without memory issues

- [ ] **Test 9.3**: Connection pooling effectiveness
  - Make multiple requests
  - Verify: Check metrics for connection reuse

---

## 10. Security Testing

**Tests:**

- [ ] **Test 10.1**: Credentials not exposed to client
  ```bash
  curl -v -H "Remote-User: john" \
          -H "Remote-Email: john@example.com" \
          http://localhost:5000/_cluster/health
  ```
  - Verify: Response headers do NOT contain Elasticsearch credentials

- [ ] **Test 10.2**: Sensitive data sanitized in logs
  ```bash
  docker compose logs elastauth | grep -i password
  ```
  - Verify: No passwords or credentials in logs

- [ ] **Test 10.3**: Cached credentials encrypted
  ```bash
  docker compose exec redis redis-cli GET "elastauth-testuser"
  ```
  - Verify: Value is encrypted (not plaintext password)

- [ ] **Test 10.4**: Input validation prevents injection
  ```bash
  # Try malicious inputs
  curl -H "Remote-User: john'; DROP TABLE users; --" \
       -H "Remote-Email: john@example.com" \
       http://localhost:5000/_search
  ```
  - Verify: Request validated and rejected or sanitized

---

## 11. Metrics and Observability Testing

**Tests:**

- [ ] **Test 11.1**: Proxy metrics exposed
  ```bash
  curl http://localhost:5000/metrics | grep proxy
  ```
  - Expected: See proxy-specific metrics (request count, latency, errors)

- [ ] **Test 11.2**: Authentication metrics tracked
  ```bash
  curl http://localhost:5000/metrics | grep auth
  ```
  - Expected: See authentication success/failure metrics

- [ ] **Test 11.3**: Cache metrics tracked
  ```bash
  curl http://localhost:5000/metrics | grep cache
  ```
  - Expected: See cache hit/miss metrics

- [ ] **Test 11.4**: Request IDs in logs
  ```bash
  docker compose logs elastauth | grep request_id
  ```
  - Verify: All log entries have request IDs for tracing

---

## 12. Configuration Testing

**Tests:**

- [ ] **Test 12.1**: Environment variables override config file
  ```bash
  # Set env var different from config.yml
  export ELASTAUTH_PROXY_ENABLED=true
  export ELASTAUTH_PROXY_ELASTICSEARCH_URL="http://different:9200"
  ./elastauth
  ```
  - Verify: Uses environment variable value

- [ ] **Test 12.2**: Invalid configuration fails startup
  ```bash
  # Set invalid proxy config
  export ELASTAUTH_PROXY_ENABLED=true
  export ELASTAUTH_PROXY_ELASTICSEARCH_URL=""
  ./elastauth
  ```
  - Expected: Fails to start with descriptive error

- [ ] **Test 12.3**: Configuration validation at startup
  ```bash
  # Start with valid config
  ./elastauth
  ```
  - Verify: Logs show configuration validation passed

---

## 13. Documentation Verification

**Tests:**

- [ ] **Test 13.1**: README instructions work
  - Follow README.md instructions
  - Verify: Can start elastauth successfully

- [ ] **Test 13.2**: Example deployments work
  - Test traefik-auth-only example
  - Test direct-proxy example
  - Verify: Both examples start and work correctly

- [ ] **Test 13.3**: Configuration examples valid
  - Use config.yml.example
  - Verify: Valid configuration that starts successfully

- [ ] **Test 13.4**: OpenAPI spec accurate
  - Check docs/src/schemas/openapi.yaml
  - Verify: Matches actual API behavior

---

## Summary Checklist

After completing all tests above, verify:

- [ ] All automated tests pass (`go test ./...`)
- [ ] Build succeeds (`go build`)
- [ ] Auth-only mode works (backward compatible)
- [ ] Proxy mode works (new functionality)
- [ ] Traefik integration works
- [ ] Both Authelia and OIDC providers work in both modes
- [ ] Special paths bypass proxying
- [ ] Credential caching works
- [ ] Error handling is comprehensive
- [ ] Security measures in place (no credential leaks, sanitized logs)
- [ ] Metrics exposed and accurate
- [ ] Configuration validation works
- [ ] Documentation accurate and complete
- [ ] Examples work out of the box

---

## Notes

- Document any issues found during testing
- Take screenshots of successful tests for reference
- Note any performance observations
- Record any edge cases discovered
- Update documentation if any gaps found

---

## Sign-off

**Tester:** ___________________  
**Date:** ___________________  
**Result:** [ ] PASS  [ ] FAIL  
**Notes:** ___________________
