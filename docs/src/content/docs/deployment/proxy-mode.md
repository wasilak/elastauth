---
title: Transparent Proxy Mode
description: Deploy elastauth as a direct authentication proxy
---

Transparent proxy mode allows elastauth to act as a complete HTTP proxy, handling both authentication and request forwarding to Elasticsearch in a single service.

## Overview

In this mode, elastauth authenticates requests and proxies them directly to Elasticsearch, eliminating the need for a separate reverse proxy.

### Architecture

```
┌─────────┐    ┌─────────────────────┐    ┌──────────────┐
│ Client  │───►│     elastauth       │───►│Elasticsearch │
└─────────┘    │ (Auth + Proxy)      │    └──────────────┘
               └─────────────────────┘
```

### Request Flow

1. Client sends request to elastauth
2. elastauth validates authentication headers/tokens
3. elastauth creates or updates Elasticsearch user with appropriate roles
4. elastauth injects `Authorization` header with generated credentials
5. elastauth proxies request to Elasticsearch
6. Elasticsearch processes request and returns response
7. elastauth forwards response to client

## When to Use

Choose transparent proxy mode when:

- ✅ You don't need a reverse proxy
- ✅ You want simpler architecture with fewer components
- ✅ You're deploying a single service (Elasticsearch only)
- ✅ You want direct control over proxy behavior
- ✅ You prefer minimal infrastructure complexity

## Configuration

### Basic Configuration

Create `config.yml`:

```yaml
# Enable proxy mode
proxy:
  enabled: true
  elasticsearch_url: "https://elasticsearch:9200"
  timeout: "30s"
  max_idle_conns: 100
  idle_conn_timeout: "90s"

# Authentication provider
auth_provider: "authelia"

# Elasticsearch connection (for user management)
elasticsearch:
  hosts: ["https://elasticsearch:9200"]
  username: "elastic"
  password: "changeme"

# Security
secret_key: "your-64-char-hex-key"
default_roles: ["kibana_user"]

# Cache for performance
cache:
  type: "redis"
  redis_host: "redis:6379"
  expiration: "1h"

# Server settings
listen: "0.0.0.0:5000"
log_level: "info"
```

### Environment Variables

```bash
# Proxy mode
ELASTAUTH_PROXY_ENABLED=true
ELASTAUTH_PROXY_ELASTICSEARCH_URL=https://elasticsearch:9200
ELASTAUTH_PROXY_TIMEOUT=30s
ELASTAUTH_PROXY_MAX_IDLE_CONNS=100
ELASTAUTH_PROXY_IDLE_CONN_TIMEOUT=90s

# Authentication
ELASTAUTH_AUTH_PROVIDER=authelia
ELASTAUTH_SECRET_KEY=your-64-char-hex-key

# Elasticsearch
ELASTAUTH_ELASTICSEARCH_USERNAME=elastic
ELASTAUTH_ELASTICSEARCH_PASSWORD=changeme

# Cache
ELASTAUTH_CACHE_TYPE=redis
ELASTAUTH_CACHE_REDIS_HOST=redis:6379
ELASTAUTH_CACHE_EXPIRATION=1h
```

## TLS Configuration

### Elasticsearch TLS

Configure TLS for secure connections to Elasticsearch:

```yaml
proxy:
  enabled: true
  elasticsearch_url: "https://elasticsearch:9200"
  
  tls:
    enabled: true
    insecure_skip_verify: false  # Never use true in production
    ca_cert: "/certs/ca.crt"
    client_cert: "/certs/client.crt"
    client_key: "/certs/client.key"
```

### Environment Variables for TLS

```bash
ELASTAUTH_PROXY_TLS_ENABLED=true
ELASTAUTH_PROXY_TLS_INSECURE_SKIP_VERIFY=false
ELASTAUTH_PROXY_TLS_CA_CERT=/certs/ca.crt
ELASTAUTH_PROXY_TLS_CLIENT_CERT=/certs/client.crt
ELASTAUTH_PROXY_TLS_CLIENT_KEY=/certs/client.key
```

### Certificate Setup

#### Generate Self-Signed Certificates (Development)

```bash
# Create certificates directory
mkdir -p certs

# Generate CA key and certificate
openssl req -x509 -newkey rsa:4096 -keyout certs/ca.key \
  -out certs/ca.crt -days 365 -nodes \
  -subj "/CN=Elasticsearch CA"

# Generate server key and CSR
openssl req -newkey rsa:4096 -keyout certs/server.key \
  -out certs/server.csr -nodes \
  -subj "/CN=elasticsearch"

# Sign server certificate with CA
openssl x509 -req -in certs/server.csr -CA certs/ca.crt \
  -CAkey certs/ca.key -CAcreateserial \
  -out certs/server.crt -days 365

# Generate client key and CSR
openssl req -newkey rsa:4096 -keyout certs/client.key \
  -out certs/client.csr -nodes \
  -subj "/CN=elastauth"

# Sign client certificate with CA
openssl x509 -req -in certs/client.csr -CA certs/ca.crt \
  -CAkey certs/ca.key -CAcreateserial \
  -out certs/client.crt -days 365
```

#### Production Certificates

Use certificates from your organization's PKI or Let's Encrypt:

```yaml
proxy:
  tls:
    enabled: true
    ca_cert: "/etc/ssl/certs/company-ca.crt"
    client_cert: "/etc/ssl/certs/elastauth.crt"
    client_key: "/etc/ssl/private/elastauth.key"
```

## Docker Compose Example

### Basic Setup

```yaml
version: '3.8'

services:
  authelia:
    image: authelia/authelia:latest
    volumes:
      - ./authelia:/config
    environment:
      - TZ=UTC
    ports:
      - "9091:9091"
    networks:
      - elastauth-net

  elastauth:
    image: ghcr.io/wasilak/elastauth:latest
    ports:
      - "5000:5000"
    environment:
      # Proxy mode
      - ELASTAUTH_PROXY_ENABLED=true
      - ELASTAUTH_PROXY_ELASTICSEARCH_URL=https://elasticsearch:9200
      - ELASTAUTH_PROXY_TIMEOUT=30s
      
      # Authentication
      - ELASTAUTH_AUTH_PROVIDER=authelia
      - ELASTAUTH_SECRET_KEY=${ELASTAUTH_SECRET_KEY}
      
      # Elasticsearch
      - ELASTAUTH_ELASTICSEARCH_USERNAME=elastic
      - ELASTAUTH_ELASTICSEARCH_PASSWORD=${ELASTICSEARCH_PASSWORD}
      
      # Cache
      - ELASTAUTH_CACHE_TYPE=redis
      - ELASTAUTH_CACHE_REDIS_HOST=redis:6379
    volumes:
      - ./config.yml:/app/config.yml:ro
    networks:
      - elastauth-net
    depends_on:
      - redis
      - elasticsearch
      - authelia

  redis:
    image: redis:7-alpine
    networks:
      - elastauth-net

  elasticsearch:
    image: docker.elastic.co/elasticsearch/elasticsearch:8.11.0
    environment:
      - discovery.type=single-node
      - xpack.security.enabled=true
      - ELASTIC_PASSWORD=${ELASTICSEARCH_PASSWORD}
    networks:
      - elastauth-net
    volumes:
      - es-data:/usr/share/elasticsearch/data

networks:
  elastauth-net:
    driver: bridge

volumes:
  es-data:
```

### With TLS

```yaml
version: '3.8'

services:
  authelia:
    image: authelia/authelia:latest
    volumes:
      - ./authelia:/config
    environment:
      - TZ=UTC
    ports:
      - "9091:9091"
    networks:
      - elastauth-net

  elastauth:
    image: ghcr.io/wasilak/elastauth:latest
    ports:
      - "5000:5000"
    environment:
      # Proxy mode with TLS
      - ELASTAUTH_PROXY_ENABLED=true
      - ELASTAUTH_PROXY_ELASTICSEARCH_URL=https://elasticsearch:9200
      - ELASTAUTH_PROXY_TIMEOUT=30s
      - ELASTAUTH_PROXY_TLS_ENABLED=true
      - ELASTAUTH_PROXY_TLS_CA_CERT=/certs/ca.crt
      - ELASTAUTH_PROXY_TLS_CLIENT_CERT=/certs/client.crt
      - ELASTAUTH_PROXY_TLS_CLIENT_KEY=/certs/client.key
      
      # Authentication
      - ELASTAUTH_AUTH_PROVIDER=authelia
      - ELASTAUTH_SECRET_KEY=${ELASTAUTH_SECRET_KEY}
      
      # Elasticsearch
      - ELASTAUTH_ELASTICSEARCH_USERNAME=elastic
      - ELASTAUTH_ELASTICSEARCH_PASSWORD=${ELASTICSEARCH_PASSWORD}
      
      # Cache
      - ELASTAUTH_CACHE_TYPE=redis
      - ELASTAUTH_CACHE_REDIS_HOST=redis:6379
    volumes:
      - ./config.yml:/app/config.yml:ro
      - ./certs:/certs:ro
    networks:
      - elastauth-net
    depends_on:
      - redis
      - elasticsearch
      - authelia

  redis:
    image: redis:7-alpine
    networks:
      - elastauth-net

  elasticsearch:
    image: docker.elastic.co/elasticsearch/elasticsearch:8.11.0
    environment:
      - discovery.type=single-node
      - xpack.security.enabled=true
      - xpack.security.http.ssl.enabled=true
      - xpack.security.http.ssl.key=/usr/share/elasticsearch/config/certs/server.key
      - xpack.security.http.ssl.certificate=/usr/share/elasticsearch/config/certs/server.crt
      - xpack.security.http.ssl.certificate_authorities=/usr/share/elasticsearch/config/certs/ca.crt
      - ELASTIC_PASSWORD=${ELASTICSEARCH_PASSWORD}
    volumes:
      - es-data:/usr/share/elasticsearch/data
      - ./certs:/usr/share/elasticsearch/config/certs:ro
    networks:
      - elastauth-net

networks:
  elastauth-net:
    driver: bridge

volumes:
  es-data:
```

## Testing

### Test Proxy Endpoint

```bash
# Test with Authelia headers
curl -H "Remote-User: testuser" \
     -H "Remote-Groups: admins,users" \
     -H "Remote-Email: test@example.com" \
     http://localhost:5000/_cluster/health

# Should return Elasticsearch cluster health
{
  "cluster_name": "docker-cluster",
  "status": "green",
  "timed_out": false,
  ...
}
```

### Test with OIDC

```bash
# Get JWT token from your OIDC provider
TOKEN="your-jwt-token"

# Access Elasticsearch through elastauth
curl -H "Authorization: Bearer $TOKEN" \
     http://localhost:5000/_cluster/health
```

### Test Different Endpoints

```bash
# Cluster health
curl -H "Remote-User: testuser" \
     http://localhost:5000/_cluster/health

# Index operations
curl -H "Remote-User: testuser" \
     http://localhost:5000/_cat/indices

# Search
curl -H "Remote-User: testuser" \
     -X POST http://localhost:5000/my-index/_search \
     -H "Content-Type: application/json" \
     -d '{"query": {"match_all": {}}}'

# Document operations
curl -H "Remote-User: testuser" \
     -X PUT http://localhost:5000/my-index/_doc/1 \
     -H "Content-Type: application/json" \
     -d '{"field": "value"}'
```

### Verify Credential Caching

```bash
# First request (generates credentials)
time curl -H "Remote-User: testuser" \
          http://localhost:5000/_cluster/health

# Second request (uses cached credentials - should be faster)
time curl -H "Remote-User: testuser" \
          http://localhost:5000/_cluster/health
```

## Performance Tuning

### Connection Pool Configuration

Optimize connection pooling for your workload:

```yaml
proxy:
  max_idle_conns: 100        # Total idle connections
  idle_conn_timeout: "90s"   # How long idle connections stay open
  timeout: "30s"             # Request timeout
```

**Recommendations by workload:**

#### Low Traffic (< 10 req/s)
```yaml
proxy:
  max_idle_conns: 10
  idle_conn_timeout: "30s"
  timeout: "30s"
```

#### Medium Traffic (10-100 req/s)
```yaml
proxy:
  max_idle_conns: 100
  idle_conn_timeout: "90s"
  timeout: "30s"
```

#### High Traffic (> 100 req/s)
```yaml
proxy:
  max_idle_conns: 200
  idle_conn_timeout: "120s"
  timeout: "60s"
```

### Cache Optimization

```yaml
cache:
  type: "redis"
  expiration: "1h"  # Adjust based on security vs performance
  redis_host: "redis:6379"
```

**Cache expiration guidelines:**

- **High security**: 15-30 minutes (more frequent credential rotation)
- **Balanced**: 1 hour (default, good for most use cases)
- **Performance**: 2-4 hours (fewer Elasticsearch user operations)

### Horizontal Scaling

Scale elastauth instances with shared Redis cache:

```yaml
services:
  elastauth:
    image: ghcr.io/wasilak/elastauth:latest
    deploy:
      replicas: 3  # Multiple instances
    environment:
      - ELASTAUTH_PROXY_ENABLED=true
      - ELASTAUTH_CACHE_TYPE=redis
      - ELASTAUTH_CACHE_REDIS_HOST=redis:6379
```

Add a load balancer:

```yaml
services:
  nginx:
    image: nginx:alpine
    ports:
      - "80:80"
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf:ro
    depends_on:
      - elastauth
    networks:
      - elastauth-net
```

nginx.conf:
```nginx
upstream elastauth {
    least_conn;
    server elastauth-1:5000;
    server elastauth-2:5000;
    server elastauth-3:5000;
}

server {
    listen 80;
    
    location / {
        proxy_pass http://elastauth;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    }
}
```

### Monitoring

Enable metrics for performance monitoring:

```yaml
enable_metrics: true
```

Access metrics:
```bash
curl http://localhost:5000/metrics
```

Key metrics to monitor:
- `elastauth_proxy_requests_total` - Total proxy requests
- `elastauth_proxy_request_duration_seconds` - Request latency
- `elastauth_proxy_errors_total` - Proxy errors
- `elastauth_cache_hits_total` - Cache hit rate
- `elastauth_cache_misses_total` - Cache miss rate

## Troubleshooting

### Issue: 502 Bad Gateway

**Symptoms:**
- elastauth returns 502
- Logs show "connection refused" or "no such host"

**Solutions:**

1. **Verify Elasticsearch URL:**
   ```yaml
   proxy:
     elasticsearch_url: "https://elasticsearch:9200"  # Check hostname and port
   ```

2. **Test connectivity:**
   ```bash
   docker exec elastauth curl https://elasticsearch:9200
   ```

3. **Check Elasticsearch is running:**
   ```bash
   docker ps | grep elasticsearch
   docker logs elasticsearch
   ```

4. **Verify network connectivity:**
   ```bash
   docker exec elastauth ping elasticsearch
   ```

### Issue: TLS Certificate Errors

**Symptoms:**
- "x509: certificate signed by unknown authority"
- "tls: bad certificate"

**Solutions:**

1. **Verify CA certificate path:**
   ```yaml
   proxy:
     tls:
       ca_cert: "/certs/ca.crt"  # Must exist in container
   ```

2. **Check certificate is mounted:**
   ```yaml
   volumes:
     - ./certs:/certs:ro  # Ensure certificates are mounted
   ```

3. **Verify certificate contents:**
   ```bash
   docker exec elastauth cat /certs/ca.crt
   ```

4. **For development only, disable verification:**
   ```yaml
   proxy:
     tls:
       insecure_skip_verify: true  # NEVER use in production
   ```

### Issue: Slow Response Times

**Symptoms:**
- High latency
- Timeouts
- Poor performance

**Solutions:**

1. **Increase connection pool:**
   ```yaml
   proxy:
     max_idle_conns: 200
     idle_conn_timeout: "120s"
   ```

2. **Increase timeout:**
   ```yaml
   proxy:
     timeout: "60s"
   ```

3. **Check cache hit rate:**
   ```bash
   curl http://localhost:5000/metrics | grep cache
   ```

4. **Optimize cache expiration:**
   ```yaml
   cache:
     expiration: "2h"  # Longer expiration = fewer ES operations
   ```

5. **Enable connection keep-alive:**
   ```yaml
   proxy:
     idle_conn_timeout: "120s"  # Keep connections open longer
   ```

### Issue: Authentication Failures

**Symptoms:**
- 401 Unauthorized
- "missing authentication headers"

**Solutions:**

1. **Verify authentication headers:**
   ```bash
   # For Authelia
   curl -H "Remote-User: testuser" http://localhost:5000/_cluster/health
   
   # For OIDC
   curl -H "Authorization: Bearer $TOKEN" http://localhost:5000/_cluster/health
   ```

2. **Check provider configuration:**
   ```yaml
   auth_provider: "authelia"  # Must match your setup
   ```

3. **Review elastauth logs:**
   ```bash
   docker logs elastauth | grep -i auth
   ```

### Issue: Elasticsearch User Not Created

**Symptoms:**
- Authentication succeeds
- Elasticsearch returns 401
- User doesn't exist

**Solutions:**

1. **Verify Elasticsearch credentials:**
   ```yaml
   elasticsearch:
     username: "elastic"  # Must have user management permissions
     password: "changeme"
   ```

2. **Check dry run mode:**
   ```yaml
   elasticsearch:
     dry_run: false  # Must be false for user creation
   ```

3. **Test Elasticsearch connectivity:**
   ```bash
   curl -u elastic:changeme https://elasticsearch:9200/_security/user
   ```

4. **Check elastauth logs:**
   ```bash
   docker logs elastauth | grep -i "user creation"
   ```

### Issue: High Memory Usage

**Symptoms:**
- elastauth consuming excessive memory
- OOM errors

**Solutions:**

1. **Reduce connection pool:**
   ```yaml
   proxy:
     max_idle_conns: 50  # Lower value
   ```

2. **Reduce cache size:**
   ```yaml
   cache:
     expiration: "30m"  # Shorter expiration
   ```

3. **Monitor memory usage:**
   ```bash
   docker stats elastauth
   ```

4. **Set memory limits:**
   ```yaml
   services:
     elastauth:
       deploy:
         resources:
           limits:
             memory: 512M
   ```

## Security Considerations

### Secret Key Management

```bash
# Generate secure key
openssl rand -hex 32

# Store securely
export ELASTAUTH_SECRET_KEY=$(openssl rand -hex 32)

# Use in docker-compose
environment:
  - ELASTAUTH_SECRET_KEY=${ELASTAUTH_SECRET_KEY}
```

### TLS Best Practices

1. **Always use TLS in production:**
   ```yaml
   proxy:
     tls:
       enabled: true
       insecure_skip_verify: false
   ```

2. **Use valid certificates:**
   - Production: Use certificates from trusted CA
   - Development: Use self-signed certificates

3. **Rotate certificates regularly:**
   - Set appropriate expiration dates
   - Automate certificate renewal

### Network Security

```yaml
networks:
  elastauth-net:
    driver: bridge
    internal: false  # Allow external access to elastauth
  
  backend-net:
    driver: bridge
    internal: true  # Isolate Elasticsearch from external access
```

### Credential Security

- Credentials are encrypted in cache using AES-256
- Credentials are never logged
- Credentials are never exposed to clients
- Use strong secret key (32 bytes, random)

## Next Steps

- [Configuration Reference](/configuration/) - Complete configuration options
- [OIDC Provider](/providers/oidc/) - Use OIDC authentication
- [Cache Configuration](/cache/redis/) - Optimize Redis cache
- [Troubleshooting Guide](/guides/troubleshooting/) - More solutions
