# Breaking Change: /elastauth/* Endpoint Prefix

## Summary

All elastauth endpoints now require the `/elastauth/` prefix. This enables transparent proxying where all `/*` paths can be forwarded to Elasticsearch without conflicts.

## What Changed

### Before (Old Endpoints)

```
/                    → Main authentication endpoint
/health              → Health check
/ready               → Readiness probe
/live                → Liveness probe
/config              → Configuration endpoint
/docs                → Swagger UI
/api/openapi.yaml    → OpenAPI specification
```

### After (New Endpoints)

```
/elastauth           → Main authentication endpoint
/elastauth/health    → Health check
/elastauth/ready     → Readiness probe
/elastauth/live      → Liveness probe
/elastauth/config    → Configuration endpoint
/elastauth/docs      → Swagger UI
/elastauth/api/openapi.yaml → OpenAPI specification
```

## Why This Change?

### Problem

Previously, elastauth's endpoints overlapped with potential Elasticsearch paths:
- `/health` could conflict with Elasticsearch's health endpoint
- `/_search` and other ES paths had to be explicitly handled
- No clean separation between elastauth and proxied content

### Solution

By scoping all elastauth endpoints under `/elastauth/*`:
- **Clean separation**: elastauth endpoints are clearly distinct
- **Transparent proxying**: All `/*` paths can go directly to Elasticsearch
- **Simpler routing**: Just check if path starts with `/elastauth/`
- **Consistent**: Same endpoint structure in all operating modes

## Migration Guide

### 1. Update Health Check URLs

**Kubernetes/Docker health checks:**

```yaml
# Before
healthcheck:
  test: ["CMD", "wget", "--spider", "-q", "http://localhost:8080/health"]

# After
healthcheck:
  test: ["CMD", "wget", "--spider", "-q", "http://localhost:8080/elastauth/health"]
```

**Kubernetes probes:**

```yaml
# Before
livenessProbe:
  httpGet:
    path: /live
    port: 8080

readinessProbe:
  httpGet:
    path: /ready
    port: 8080

# After
livenessProbe:
  httpGet:
    path: /elastauth/live
    port: 8080

readinessProbe:
  httpGet:
    path: /elastauth/ready
    port: 8080
```

### 2. Update Traefik Configuration (Auth-Only Mode)

**Traefik ForwardAuth middleware:**

```yaml
# Before
http:
  middlewares:
    elastauth:
      forwardAuth:
        address: "http://elastauth:5000/"

# After
http:
  middlewares:
    elastauth:
      forwardAuth:
        address: "http://elastauth:5000/elastauth"
```

### 3. Update Monitoring/Metrics

**Prometheus scrape configs:**

```yaml
# Before
- job_name: 'elastauth'
  metrics_path: '/metrics'
  static_configs:
    - targets: ['elastauth:8080']

# After
- job_name: 'elastauth'
  metrics_path: '/elastauth/metrics'
  static_configs:
    - targets: ['elastauth:8080']
```

### 4. Update API Clients

**Configuration endpoint:**

```bash
# Before
curl http://elastauth:8080/config

# After
curl http://elastauth:8080/elastauth/config
```

**API documentation:**

```bash
# Before
curl http://elastauth:8080/docs

# After
curl http://elastauth:8080/elastauth/docs
```

### 5. Update Scripts and Automation

Search your codebase for references to elastauth endpoints:

```bash
# Find all references to old endpoints
grep -r "localhost:8080/health" .
grep -r "elastauth:5000/ready" .
grep -r "/live" .
grep -r "/config" .
```

Replace with new `/elastauth/*` paths.

## Benefits

### For Transparent Proxy Mode

```
Client Request Flow:
- /_search              → Proxied to Elasticsearch
- /_cluster/health      → Proxied to Elasticsearch
- /my-index/_doc/1      → Proxied to Elasticsearch
- /elastauth/health     → Handled by elastauth (not proxied)
```

**Clean separation**: No ambiguity about which service handles which path.

### For Auth-Only Mode

```
Traefik Flow:
- Client → Traefik → elastauth (/elastauth) → Traefik → Elasticsearch
```

**Consistent**: Same endpoint structure regardless of operating mode.

### For Operations

- **Easier debugging**: Clear distinction in logs between elastauth and ES requests
- **Better monitoring**: Separate metrics for elastauth vs proxied requests
- **Simpler configuration**: No need for complex path matching rules

## Testing Your Migration

### 1. Test Health Checks

```bash
# Test liveness
curl http://localhost:8080/elastauth/live
# Expected: {"status":"OK","timestamp":"...","uptime":"...","mode":"..."}

# Test readiness
curl http://localhost:8080/elastauth/ready
# Expected: {"status":"OK","checks":{...},"timestamp":"..."}

# Test health
curl http://localhost:8080/elastauth/health
# Expected: {"status":"OK"}
```

### 2. Test Proxy Mode (if enabled)

```bash
# Test elastauth endpoint (should NOT proxy)
curl http://localhost:8080/elastauth/config
# Expected: elastauth configuration JSON

# Test Elasticsearch path (should proxy)
curl -H "Remote-User: testuser" http://localhost:8080/_cluster/health
# Expected: Elasticsearch cluster health (proxied)
```

### 3. Test Auth-Only Mode (if using Traefik)

```bash
# Test through Traefik
curl http://localhost:80/
# Expected: Traefik forwards to elastauth at /elastauth, then to Elasticsearch
```

## Rollback Plan

If you need to rollback:

1. **Revert to previous version:**
   ```bash
   git checkout <previous-commit>
   docker compose down
   docker compose build
   docker compose up -d
   ```

2. **Restore old endpoint configurations** in:
   - Health check definitions
   - Kubernetes probe configurations
   - Traefik middleware configurations
   - Monitoring scrape configs

## Timeline

- **Implemented**: January 15, 2026
- **Commit**: f12303653f3062cc06e7fa297415d4045ea6a4ec
- **Affects**: All elastauth deployments (both auth-only and proxy modes)

## Support

If you encounter issues during migration:

1. Check logs: `docker compose logs elastauth`
2. Verify endpoint accessibility: `curl http://localhost:8080/elastauth/health`
3. Review [GitHub Issues](https://github.com/wasilak/elastauth/issues)
4. Consult [Documentation](https://wasilak.github.io/elastauth/)

## Related Changes

- Simplified router logic (removed `isSpecialPath()` function)
- Updated all tests to use new endpoint paths
- Updated all documentation and examples
- Updated deployment examples (docker-compose, Kubernetes)

## Questions?

**Q: Do I need to change my Elasticsearch queries?**
A: No. Elasticsearch paths (like `/_search`, `/_cluster/health`) are unchanged. Only elastauth's own endpoints changed.

**Q: Will this affect my existing authentication flow?**
A: The authentication flow is unchanged. Only the endpoint URLs changed.

**Q: Can I use both old and new endpoints during migration?**
A: No. This is a breaking change. You must update all references to use the new `/elastauth/*` paths.

**Q: Does this affect the Elasticsearch proxy functionality?**
A: No. In proxy mode, all non-`/elastauth/*` paths are still proxied to Elasticsearch as before.

**Q: What about custom middleware or plugins?**
A: Update any custom code that references elastauth endpoints to use the new `/elastauth/*` prefix.
