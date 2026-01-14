---
title: Authentication-Only Mode
description: Deploy elastauth with Traefik forward authentication
---

Authentication-only mode is elastauth's default operating mode, designed to work seamlessly with reverse proxies like Traefik using the forward authentication pattern.

## Overview

In this mode, elastauth validates authentication and returns authorization headers. The reverse proxy (Traefik) handles the actual proxying to Elasticsearch.

### Architecture

```
┌─────────┐    ┌─────────┐    ┌───────────┐    ┌─────────┐    ┌──────────────┐
│ Client  │───►│ Traefik │───►│ elastauth │───►│ Traefik │───►│Elasticsearch │
└─────────┘    │ (Proxy) │    │  (Auth)   │    │ (Proxy) │    └──────────────┘
               └─────────┘    └───────────┘    └─────────┘
                    │              │                 │
                    └──────────────┴─────────────────┘
                         Forward Auth Flow
```

### Request Flow

1. Client sends request to Traefik
2. Traefik forwards authentication headers to elastauth (forward auth middleware)
3. elastauth validates authentication with configured provider (Authelia/OIDC)
4. elastauth creates or updates Elasticsearch user with appropriate roles
5. elastauth returns `Authorization` header to Traefik
6. Traefik forwards original request to Elasticsearch with injected auth header
7. Elasticsearch processes request and returns response
8. Traefik forwards response to client

## When to Use

Choose authentication-only mode when:

- ✅ You already use Traefik or another reverse proxy
- ✅ You need advanced routing, load balancing, or middleware
- ✅ You protect multiple services with the same proxy
- ✅ You want centralized TLS termination
- ✅ You need Traefik's ecosystem (rate limiting, circuit breakers, plugins)

## Configuration

### elastauth Configuration

Create `config.yml`:

```yaml
# Proxy mode disabled (default)
proxy:
  enabled: false

# Authentication provider
auth_provider: "authelia"

# Elasticsearch connection
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
# Core settings
ELASTAUTH_AUTH_PROVIDER=authelia
ELASTAUTH_SECRET_KEY=your-64-char-hex-key
ELASTAUTH_LISTEN=0.0.0.0:5000

# Elasticsearch
ELASTAUTH_ELASTICSEARCH_USERNAME=elastic
ELASTAUTH_ELASTICSEARCH_PASSWORD=changeme

# Cache
ELASTAUTH_CACHE_TYPE=redis
ELASTAUTH_CACHE_REDIS_HOST=redis:6379
ELASTAUTH_CACHE_EXPIRATION=1h
```

## Traefik Integration

### Traefik Middleware Configuration

Create a forward auth middleware in Traefik:

#### Static Configuration (traefik.yml)

```yaml
entryPoints:
  web:
    address: ":80"
  websecure:
    address: ":443"

providers:
  file:
    filename: /etc/traefik/dynamic-config.yml
    watch: true

api:
  dashboard: true
  insecure: true

log:
  level: INFO
```

#### Dynamic Configuration (dynamic-config.yml)

```yaml
http:
  middlewares:
    # elastauth forward auth middleware
    elastauth:
      forwardAuth:
        address: "http://elastauth:5000/"
        trustForwardHeader: true
        authResponseHeaders:
          - "Authorization"
    
    # Optional: Chain with Authelia
    authelia:
      forwardAuth:
        address: "http://authelia:9091/api/verify?rd=https://auth.example.com"
        trustForwardHeader: true
        authResponseHeaders:
          - "Remote-User"
          - "Remote-Groups"
          - "Remote-Email"
          - "Remote-Name"
    
    # Middleware chain: Authelia → elastauth
    auth-chain:
      chain:
        middlewares:
          - authelia
          - elastauth

  routers:
    # Elasticsearch/Kibana router
    elasticsearch:
      rule: "Host(`kibana.example.com`)"
      entryPoints:
        - websecure
      middlewares:
        - auth-chain
      service: elasticsearch
      tls:
        certResolver: letsencrypt

  services:
    elasticsearch:
      loadBalancer:
        servers:
          - url: "https://elasticsearch:9200"
```

### Docker Compose Example

```yaml
version: '3.8'

services:
  traefik:
    image: traefik:v2.10
    command:
      - "--api.insecure=true"
      - "--providers.docker=true"
      - "--providers.file.filename=/etc/traefik/dynamic-config.yml"
      - "--entrypoints.web.address=:80"
      - "--entrypoints.websecure.address=:443"
    ports:
      - "80:80"
      - "443:443"
      - "8080:8080"
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock:ro
      - ./traefik/dynamic-config.yml:/etc/traefik/dynamic-config.yml:ro
      - ./certs:/certs:ro
    networks:
      - elastauth-net

  authelia:
    image: authelia/authelia:latest
    volumes:
      - ./authelia:/config
    environment:
      - TZ=UTC
    networks:
      - elastauth-net
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.authelia.rule=Host(`auth.example.com`)"
      - "traefik.http.routers.authelia.entrypoints=websecure"
      - "traefik.http.routers.authelia.tls=true"

  elastauth:
    image: ghcr.io/wasilak/elastauth:latest
    environment:
      - ELASTAUTH_AUTH_PROVIDER=authelia
      - ELASTAUTH_SECRET_KEY=${ELASTAUTH_SECRET_KEY}
      - ELASTAUTH_ELASTICSEARCH_USERNAME=elastic
      - ELASTAUTH_ELASTICSEARCH_PASSWORD=${ELASTICSEARCH_PASSWORD}
      - ELASTAUTH_CACHE_TYPE=redis
      - ELASTAUTH_CACHE_REDIS_HOST=redis:6379
    volumes:
      - ./config.yml:/app/config.yml:ro
    networks:
      - elastauth-net
    depends_on:
      - redis
      - elasticsearch

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

  kibana:
    image: docker.elastic.co/kibana/kibana:8.11.0
    environment:
      - ELASTICSEARCH_HOSTS=https://elasticsearch:9200
      - ELASTICSEARCH_USERNAME=kibana_system
      - ELASTICSEARCH_PASSWORD=${KIBANA_PASSWORD}
    networks:
      - elastauth-net
    depends_on:
      - elasticsearch
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.kibana.rule=Host(`kibana.example.com`)"
      - "traefik.http.routers.kibana.entrypoints=websecure"
      - "traefik.http.routers.kibana.middlewares=auth-chain"
      - "traefik.http.routers.kibana.tls=true"
      - "traefik.http.services.kibana.loadbalancer.server.port=5601"

networks:
  elastauth-net:
    driver: bridge

volumes:
  es-data:
```

## Chaining with Authelia

elastauth works seamlessly with Authelia in a middleware chain:

### Middleware Chain Flow

```
Request → Traefik → Authelia → elastauth → Traefik → Elasticsearch
```

1. **Traefik** receives the request
2. **Authelia** validates user authentication (session, 2FA, etc.)
3. **Authelia** adds user headers (`Remote-User`, `Remote-Groups`, etc.)
4. **elastauth** receives request with Authelia headers
5. **elastauth** creates/updates Elasticsearch user
6. **elastauth** returns `Authorization` header
7. **Traefik** forwards request to Elasticsearch with auth header

### Authelia Configuration

In your Authelia `configuration.yml`:

```yaml
access_control:
  default_policy: deny
  rules:
    - domain: kibana.example.com
      policy: two_factor
      subject:
        - "group:admins"
        - "group:users"
```

### Traefik Middleware Chain

```yaml
http:
  middlewares:
    authelia:
      forwardAuth:
        address: "http://authelia:9091/api/verify?rd=https://auth.example.com"
        trustForwardHeader: true
        authResponseHeaders:
          - "Remote-User"
          - "Remote-Groups"
          - "Remote-Email"
          - "Remote-Name"
    
    elastauth:
      forwardAuth:
        address: "http://elastauth:5000/"
        trustForwardHeader: true
        authResponseHeaders:
          - "Authorization"
    
    auth-chain:
      chain:
        middlewares:
          - authelia
          - elastauth
```

## Testing

### Test Authentication Endpoint

```bash
# Test with Authelia headers
curl -H "Remote-User: testuser" \
     -H "Remote-Groups: admins,users" \
     -H "Remote-Email: test@example.com" \
     http://localhost:5000/

# Expected response
{
  "message": "Authentication successful",
  "username": "testuser",
  "groups": ["admins", "users"],
  "email": "test@example.com"
}

# Check for Authorization header
curl -I -H "Remote-User: testuser" \
        -H "Remote-Groups: admins" \
        http://localhost:5000/

# Expected headers
HTTP/1.1 200 OK
Authorization: Basic dGVzdHVzZXI6Z2VuZXJhdGVkLXBhc3N3b3Jk
```

### Test Through Traefik

```bash
# Access Kibana through Traefik
# (Assumes you're authenticated with Authelia)
curl https://kibana.example.com/_cluster/health

# Should return Elasticsearch cluster health
```

### Test Elasticsearch Access

```bash
# Get authorization header from elastauth
AUTH_HEADER=$(curl -s -H "Remote-User: testuser" \
                      -H "Remote-Groups: admins" \
                      http://localhost:5000/ | \
              grep -o 'Authorization: .*' | \
              cut -d' ' -f2-)

# Use it to access Elasticsearch directly
curl -H "Authorization: $AUTH_HEADER" \
     https://elasticsearch:9200/_cluster/health
```

## Troubleshooting

### Issue: 401 Unauthorized from elastauth

**Symptoms:**
- Traefik returns 401
- elastauth logs show "missing authentication headers"

**Solutions:**

1. **Check Authelia headers are forwarded:**
   ```yaml
   # In Traefik middleware
   authelia:
     forwardAuth:
       authResponseHeaders:
         - "Remote-User"
         - "Remote-Groups"
         - "Remote-Email"
         - "Remote-Name"
   ```

2. **Verify middleware chain order:**
   ```yaml
   auth-chain:
     chain:
       middlewares:
         - authelia  # Must be first
         - elastauth # Must be second
   ```

3. **Check elastauth logs:**
   ```bash
   docker logs elastauth
   ```

### Issue: Authorization header not forwarded

**Symptoms:**
- elastauth returns 200
- Elasticsearch returns 401
- Authorization header missing

**Solutions:**

1. **Ensure elastauth middleware forwards Authorization header:**
   ```yaml
   elastauth:
     forwardAuth:
       authResponseHeaders:
         - "Authorization"  # Must be present
   ```

2. **Check Traefik logs:**
   ```bash
   docker logs traefik
   ```

3. **Verify header in response:**
   ```bash
   curl -I -H "Remote-User: test" http://elastauth:5000/
   ```

### Issue: Elasticsearch user not created

**Symptoms:**
- elastauth returns 200
- Elasticsearch returns 401
- User doesn't exist in Elasticsearch

**Solutions:**

1. **Check Elasticsearch credentials:**
   ```yaml
   elasticsearch:
     username: "elastic"  # Must have user management permissions
     password: "changeme"
   ```

2. **Verify Elasticsearch connectivity:**
   ```bash
   curl -u elastic:changeme https://elasticsearch:9200/_cluster/health
   ```

3. **Check elastauth logs for errors:**
   ```bash
   docker logs elastauth | grep -i error
   ```

4. **Disable dry run mode:**
   ```yaml
   elasticsearch:
     dry_run: false  # Must be false for user creation
   ```

### Issue: Cache not working

**Symptoms:**
- Slow response times
- High Elasticsearch load
- New credentials generated every request

**Solutions:**

1. **Verify Redis connectivity:**
   ```bash
   docker exec elastauth redis-cli -h redis ping
   ```

2. **Check cache configuration:**
   ```yaml
   cache:
     type: "redis"  # Not "memory" or empty
     redis_host: "redis:6379"
     expiration: "1h"
   ```

3. **Check Redis logs:**
   ```bash
   docker logs redis
   ```

### Issue: Traefik can't reach elastauth

**Symptoms:**
- Traefik logs show connection errors
- 502 Bad Gateway errors

**Solutions:**

1. **Verify network connectivity:**
   ```bash
   docker exec traefik ping elastauth
   ```

2. **Check service names match:**
   ```yaml
   # In Traefik config
   address: "http://elastauth:5000/"  # Must match service name
   ```

3. **Ensure services on same network:**
   ```yaml
   # In docker-compose.yml
   services:
     traefik:
       networks:
         - elastauth-net
     elastauth:
       networks:
         - elastauth-net
   ```

### Issue: TLS certificate errors

**Symptoms:**
- Certificate validation errors
- Connection refused to Elasticsearch

**Solutions:**

1. **For development, disable TLS verification:**
   ```yaml
   elasticsearch:
     hosts: ["http://elasticsearch:9200"]  # Use HTTP
   ```

2. **For production, provide CA certificate:**
   ```yaml
   # Mount certificates
   volumes:
     - ./certs/ca.crt:/certs/ca.crt:ro
   ```

3. **Configure Elasticsearch client:**
   ```bash
   # Set environment variable
   ELASTAUTH_ELASTICSEARCH_CA_CERT=/certs/ca.crt
   ```

## Performance Tuning

### Connection Pooling

Traefik handles connection pooling to Elasticsearch. Configure in Traefik:

```yaml
http:
  services:
    elasticsearch:
      loadBalancer:
        servers:
          - url: "https://elasticsearch:9200"
        healthCheck:
          path: /_cluster/health
          interval: "10s"
          timeout: "3s"
```

### Cache Optimization

```yaml
cache:
  type: "redis"
  expiration: "1h"  # Adjust based on security requirements
  redis_host: "redis:6379"
```

**Recommendations:**
- **High security**: 15-30 minutes
- **Balanced**: 1 hour (default)
- **Performance**: 2-4 hours

### Horizontal Scaling

Scale elastauth instances with Redis cache:

```yaml
services:
  elastauth:
    image: ghcr.io/wasilak/elastauth:latest
    deploy:
      replicas: 3  # Multiple instances
    environment:
      - ELASTAUTH_CACHE_TYPE=redis
      - ELASTAUTH_CACHE_REDIS_HOST=redis:6379
```

All instances share the same Redis cache for consistent credentials.

## Security Considerations

### Secret Key Management

```bash
# Generate secure key
openssl rand -hex 32

# Store in environment variable
export ELASTAUTH_SECRET_KEY=your-generated-key

# Use in docker-compose
environment:
  - ELASTAUTH_SECRET_KEY=${ELASTAUTH_SECRET_KEY}
```

### TLS Configuration

Always use TLS in production:

```yaml
# Traefik TLS
entryPoints:
  websecure:
    address: ":443"
    http:
      tls:
        certResolver: letsencrypt

# Elasticsearch TLS
elasticsearch:
  hosts: ["https://elasticsearch:9200"]
```

### Network Isolation

```yaml
networks:
  elastauth-net:
    driver: bridge
    internal: false  # Allow external access through Traefik
  
  backend-net:
    driver: bridge
    internal: true  # Isolate backend services
```

## Next Steps

- [Traefik Integration Example](/deployment/traefik/) - Complete working example
- [OIDC Provider](/providers/oidc/) - Use OIDC instead of Authelia
- [Cache Configuration](/cache/redis/) - Optimize Redis cache
- [Troubleshooting Guide](/guides/troubleshooting/) - More solutions
