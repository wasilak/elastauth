---
title: Operating Modes
description: Understanding elastauth's two operating modes
---

elastauth supports two distinct operating modes to fit different deployment architectures.

## Mode Comparison

| Feature | Authentication-Only | Transparent Proxy |
|---------|-------------------|-------------------|
| **Use Case** | Forward auth with reverse proxy | Direct authentication proxy |
| **Proxying** | Handled by external proxy (Traefik) | Built-in HTTP proxy |
| **Architecture** | Client → Traefik → elastauth → Traefik → ES | Client → elastauth → ES |
| **Complexity** | More components | Fewer components |
| **Flexibility** | Leverage Traefik features | Direct control |
| **Best For** | Multi-service deployments | Single-service deployments |

## Authentication-Only Mode

### Overview

In authentication-only mode (default), elastauth validates authentication and returns headers. A reverse proxy like Traefik handles the actual proxying to Elasticsearch.

### Architecture

```mermaid
graph LR
    Client[Client] --> Traefik1[Traefik<br/>Proxy]
    Traefik1 --> elastauth[elastauth<br/>Auth]
    elastauth --> Traefik2[Traefik<br/>Proxy]
    Traefik2 --> ES[Elasticsearch]
    
    Traefik1 -.Forward Auth.-> elastauth
    
    style elastauth fill:#e1f5e1
    style Traefik1 fill:#e3f2fd
    style Traefik2 fill:#e3f2fd
```

### Configuration

```yaml
# Proxy mode disabled (default)
proxy:
  enabled: false

auth_provider: "authelia"

elasticsearch:
  hosts: ["https://elasticsearch:9200"]
  username: "elastic"
  password: "changeme"

secret_key: "your-64-char-hex-key"
```

### Request Flow

1. Client sends request to Traefik
2. Traefik forwards to elastauth for authentication (forward auth)
3. elastauth validates authentication headers
4. elastauth creates/updates Elasticsearch user
5. elastauth returns `Authorization` header to Traefik
6. Traefik forwards original request to Elasticsearch with auth header
7. Elasticsearch processes request and returns response
8. Traefik forwards response to client

### When to Use

**Choose authentication-only mode when:**

- You already use Traefik or another reverse proxy
- You need advanced routing, load balancing, or middleware
- You protect multiple services with the same proxy
- You want centralized TLS termination
- You need Traefik's advanced features (rate limiting, circuit breakers, etc.)

### Benefits

- **Leverage existing infrastructure** - Use your existing reverse proxy
- **Advanced routing** - Traefik's powerful routing capabilities
- **Multi-service** - Protect multiple services with one proxy
- **Centralized management** - Single point for TLS, routing, middleware

### Limitations

- **More components** - Requires reverse proxy setup
- **Additional latency** - Extra hop through proxy
- **Configuration complexity** - Must configure both Traefik and elastauth

### Example Deployment

See [Traefik Integration](/deployment/traefik/) for complete setup guide.

## Transparent Proxy Mode

### Overview

In transparent proxy mode, elastauth handles both authentication and proxying. It acts as a complete HTTP proxy to Elasticsearch.

### Architecture

```mermaid
graph LR
    Client[Client] --> elastauth[elastauth<br/>Auth + Proxy]
    elastauth --> ES[Elasticsearch]
    
    style elastauth fill:#e1f5e1
```

### Configuration

```yaml
# Enable proxy mode
proxy:
  enabled: true
  elasticsearch_url: "https://elasticsearch:9200"
  timeout: "30s"
  max_idle_conns: 100
  idle_conn_timeout: "90s"
  
  tls:
    enabled: true
    insecure_skip_verify: false
    ca_cert: "/certs/ca.crt"

auth_provider: "authelia"

elasticsearch:
  hosts: ["https://elasticsearch:9200"]
  username: "elastic"
  password: "changeme"

secret_key: "your-64-char-hex-key"
```

### Request Flow

1. Client sends request to elastauth
2. elastauth validates authentication headers/tokens
3. elastauth creates/updates Elasticsearch user
4. elastauth injects `Authorization` header
5. elastauth proxies request to Elasticsearch
6. Elasticsearch processes request and returns response
7. elastauth forwards response to client

### When to Use

**Choose transparent proxy mode when:**

- You don't need a reverse proxy
- You want simpler architecture
- You're deploying a single service
- You want direct control over proxy behavior
- You prefer fewer moving parts

### Benefits

- **Simpler architecture** - Fewer components to manage
- **Lower latency** - Direct proxy, no extra hops
- **Direct control** - Full control over proxy behavior
- **Easier deployment** - Single service to deploy

### Limitations

- **No advanced routing** - Basic HTTP proxy only
- **Single service** - Designed for Elasticsearch only
- **Limited features** - No rate limiting, circuit breakers, etc.

### Example Deployment

See [Direct Proxy Mode](/deployment/direct-proxy/) for complete setup guide.

## Choosing a Mode

### Decision Tree

```
Do you already use Traefik or another reverse proxy?
├─ Yes → Use Authentication-Only Mode
└─ No
   │
   Do you need to protect multiple services?
   ├─ Yes → Use Authentication-Only Mode (with Traefik)
   └─ No
      │
      Do you need advanced routing/middleware?
      ├─ Yes → Use Authentication-Only Mode (with Traefik)
      └─ No → Use Transparent Proxy Mode
```

### Recommendations

**Use Authentication-Only Mode if:**
- You have existing Traefik infrastructure
- You need to protect multiple services
- You require advanced routing or middleware
- You want centralized TLS termination
- You need Traefik's ecosystem (plugins, middleware, etc.)

**Use Transparent Proxy Mode if:**
- You're deploying elastauth standalone
- You only need to protect Elasticsearch
- You want the simplest possible setup
- You prefer fewer components
- You don't need advanced proxy features

## Configuration Comparison

### Authentication-Only Mode

```yaml
# Minimal configuration
proxy:
  enabled: false  # Default

auth_provider: "authelia"
elasticsearch:
  hosts: ["https://elasticsearch:9200"]
  username: "elastic"
  password: "changeme"
secret_key: "your-64-char-hex-key"
```

### Transparent Proxy Mode

```yaml
# Additional proxy configuration required
proxy:
  enabled: true
  elasticsearch_url: "https://elasticsearch:9200"
  timeout: "30s"
  tls:
    enabled: true
    ca_cert: "/certs/ca.crt"

auth_provider: "authelia"
elasticsearch:
  hosts: ["https://elasticsearch:9200"]
  username: "elastic"
  password: "changeme"
secret_key: "your-64-char-hex-key"
```

## Switching Modes

You can switch between modes by changing the `proxy.enabled` setting:

```bash
# Switch to transparent proxy mode
export ELASTAUTH_PROXY_ENABLED=true
export ELASTAUTH_PROXY_ELASTICSEARCH_URL=https://elasticsearch:9200

# Switch back to authentication-only mode
export ELASTAUTH_PROXY_ENABLED=false
```

No code changes required - just configuration.

## Performance Considerations

### Authentication-Only Mode

- **Latency**: Higher (extra hop through Traefik)
- **Throughput**: Depends on Traefik configuration
- **Scaling**: Scale Traefik and elastauth independently

### Transparent Proxy Mode

- **Latency**: Lower (direct proxy)
- **Throughput**: Depends on elastauth configuration
- **Scaling**: Scale elastauth instances with Redis cache

## Next Steps

- [Traefik Integration](/deployment/traefik/) - Set up authentication-only mode
- [Direct Proxy Mode](/deployment/direct-proxy/) - Set up transparent proxy mode
- [Configuration Reference](/configuration/) - Complete configuration guide
- [Troubleshooting](/guides/troubleshooting/) - Common issues and solutions
