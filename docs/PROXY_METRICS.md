# Proxy Metrics

This document describes the Prometheus metrics exposed by elastauth when running in transparent proxy mode.

## Available Metrics

### Request Metrics

#### `elastauth_proxy_requests_total`
- **Type**: Counter
- **Description**: Total number of proxy requests
- **Labels**:
  - `method`: HTTP method (GET, POST, PUT, DELETE, etc.)
  - `status`: HTTP status category (2xx, 3xx, 4xx, 5xx, unknown)

#### `elastauth_proxy_request_duration_seconds`
- **Type**: Histogram
- **Description**: Duration of proxy requests in seconds
- **Labels**:
  - `method`: HTTP method
  - `status`: HTTP status category
- **Buckets**: Default Prometheus buckets (0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10)

#### `elastauth_proxy_active_requests`
- **Type**: Gauge
- **Description**: Number of currently active proxy requests
- **Labels**: None

### Authentication Metrics

#### `elastauth_proxy_authentication_total`
- **Type**: Counter
- **Description**: Total number of authentication attempts
- **Labels**:
  - `result`: Authentication result (`success` or `failure`)

### Cache Metrics

#### `elastauth_proxy_cache_hits_total`
- **Type**: Counter
- **Description**: Total number of cache hits and misses
- **Labels**:
  - `result`: Cache result (`hit` or `miss`)

### Error Metrics

#### `elastauth_proxy_errors_total`
- **Type**: Counter
- **Description**: Total number of proxy errors
- **Labels**:
  - `error_type`: Type of error encountered

**Error Types**:
- `auth_failed`: Authentication failed
- `invalid_username`: Invalid username format
- `invalid_email`: Invalid email format
- `invalid_name`: Invalid name format
- `invalid_group`: Invalid group name
- `invalid_request`: Request validation failed
- `missing_user_info`: User info not found in context
- `credential_generation_failed`: Failed to generate credentials
- `invalid_elasticsearch_url`: Invalid Elasticsearch URL
- `nil_response`: Received nil response from Elasticsearch

## Accessing Metrics

Metrics are exposed on the `/metrics` endpoint when `enable_metrics` is set to `true` in the configuration.

### Configuration

```yaml
enable_metrics: true
```

Or via environment variable:

```bash
export ELASTAUTH_ENABLE_METRICS=true
```

### Example Queries

#### Request Rate by Status
```promql
rate(elastauth_proxy_requests_total[5m])
```

#### Average Request Duration
```promql
rate(elastauth_proxy_request_duration_seconds_sum[5m]) / rate(elastauth_proxy_request_duration_seconds_count[5m])
```

#### Authentication Success Rate
```promql
rate(elastauth_proxy_authentication_total{result="success"}[5m]) / rate(elastauth_proxy_authentication_total[5m])
```

#### Cache Hit Rate
```promql
rate(elastauth_proxy_cache_hits_total{result="hit"}[5m]) / rate(elastauth_proxy_cache_hits_total[5m])
```

#### Error Rate by Type
```promql
rate(elastauth_proxy_errors_total[5m])
```

#### Active Requests
```promql
elastauth_proxy_active_requests
```

## Grafana Dashboard

You can create a Grafana dashboard using these metrics to monitor:

1. **Request Volume**: Track requests per second by method and status
2. **Latency**: Monitor request duration percentiles (p50, p95, p99)
3. **Authentication**: Track authentication success/failure rates
4. **Cache Performance**: Monitor cache hit/miss ratios
5. **Error Rates**: Track errors by type
6. **Active Connections**: Monitor concurrent requests

## Alerting Examples

### High Error Rate
```yaml
- alert: HighProxyErrorRate
  expr: rate(elastauth_proxy_errors_total[5m]) > 0.1
  for: 5m
  annotations:
    summary: "High proxy error rate detected"
```

### Low Cache Hit Rate
```yaml
- alert: LowCacheHitRate
  expr: rate(elastauth_proxy_cache_hits_total{result="hit"}[5m]) / rate(elastauth_proxy_cache_hits_total[5m]) < 0.5
  for: 10m
  annotations:
    summary: "Cache hit rate below 50%"
```

### High Authentication Failure Rate
```yaml
- alert: HighAuthFailureRate
  expr: rate(elastauth_proxy_authentication_total{result="failure"}[5m]) / rate(elastauth_proxy_authentication_total[5m]) > 0.1
  for: 5m
  annotations:
    summary: "Authentication failure rate above 10%"
```

## Best Practices

1. **Monitor Request Latency**: Set up alerts for p95 and p99 latency
2. **Track Error Rates**: Monitor error rates by type to identify issues
3. **Cache Performance**: Ensure cache hit rate is high (>80%)
4. **Authentication Success**: Monitor authentication failures for security issues
5. **Active Requests**: Track concurrent requests to identify capacity issues

## Integration with Existing Metrics

These proxy-specific metrics complement the existing Echo framework metrics that are already exposed when `enable_metrics` is enabled. The Echo metrics provide general HTTP server metrics, while these proxy metrics provide detailed insights into the proxy functionality.
