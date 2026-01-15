---
title: Configuration Reference
description: Complete configuration reference for elastauth
---

elastauth supports flexible configuration through YAML files and environment variables.

## Configuration Precedence

Configuration is loaded in this order (highest to lowest priority):

1. **Environment variables** - `ELASTAUTH_*` prefix
2. **Configuration file** - `config.yml`
3. **Default values** - Built-in defaults

This allows you to use config files for base configuration and override sensitive values with environment variables.

## Configuration File

Create a `config.yml` file in your working directory:

```yaml
# Operating mode
proxy:
  enabled: false  # Set to true for transparent proxy mode

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

# Cache
cache:
  type: "redis"
  redis_host: "redis:6379"
  expiration: "1h"
```

See [config.yml.example](https://github.com/wasilak/elastauth/blob/main/config.yml.example) for complete reference.

## Core Configuration

### Operating Mode

```yaml
proxy:
  enabled: false  # true for transparent proxy mode, false for auth-only mode
```

**Environment variable:** `ELASTAUTH_PROXY_ENABLED=true`

### Authentication Provider

```yaml
auth_provider: "authelia"  # Options: authelia, oidc
```

**Environment variable:** `ELASTAUTH_AUTH_PROVIDER=authelia`

### Server Settings

```yaml
listen: "0.0.0.0:3000"  # Server listen address
```

**Environment variable:** `ELASTAUTH_LISTEN=0.0.0.0:3000`

### Logging

```yaml
log_level: "info"   # Options: debug, info, warn, error
log_format: "json"  # Options: json, text
```

**Environment variables:**
- `ELASTAUTH_LOG_LEVEL=info`
- `ELASTAUTH_LOG_FORMAT=json`

## Proxy Mode Configuration

When `proxy.enabled` is `true`, elastauth operates as a transparent proxy.

```yaml
proxy:
  enabled: true
  elasticsearch_url: "https://elasticsearch:9200"
  timeout: "30s"
  max_idle_conns: 100
  idle_conn_timeout: "90s"
  
  tls:
    enabled: true
    insecure_skip_verify: false
    ca_cert: "/path/to/ca.crt"
    client_cert: "/path/to/client.crt"
    client_key: "/path/to/client.key"
```

### Proxy Settings

| Setting | Description | Default | Environment Variable |
|---------|-------------|---------|---------------------|
| `enabled` | Enable transparent proxy mode | `false` | `ELASTAUTH_PROXY_ENABLED` |
| `elasticsearch_url` | Target Elasticsearch URL | - | `ELASTAUTH_PROXY_ELASTICSEARCH_URL` |
| `timeout` | Request timeout | `30s` | `ELASTAUTH_PROXY_TIMEOUT` |
| `max_idle_conns` | Connection pool size | `100` | `ELASTAUTH_PROXY_MAX_IDLE_CONNS` |
| `idle_conn_timeout` | Idle connection timeout | `90s` | `ELASTAUTH_PROXY_IDLE_CONN_TIMEOUT` |

### TLS Configuration

| Setting | Description | Default | Environment Variable |
|---------|-------------|---------|---------------------|
| `tls.enabled` | Enable TLS | `false` | `ELASTAUTH_PROXY_TLS_ENABLED` |
| `tls.insecure_skip_verify` | Skip certificate verification | `false` | `ELASTAUTH_PROXY_TLS_INSECURE_SKIP_VERIFY` |
| `tls.ca_cert` | CA certificate path | - | `ELASTAUTH_PROXY_TLS_CA_CERT` |
| `tls.client_cert` | Client certificate path | - | `ELASTAUTH_PROXY_TLS_CLIENT_CERT` |
| `tls.client_key` | Client key path | - | `ELASTAUTH_PROXY_TLS_CLIENT_KEY` |

:::caution
Never use `insecure_skip_verify: true` in production. This disables certificate validation and is only for development/testing.
:::

## Elasticsearch Configuration

```yaml
elasticsearch:
  hosts:
    - "https://elasticsearch:9200"
    - "https://elasticsearch-2:9200"  # Optional: additional hosts for failover
  username: "elastic"
  password: "changeme"
  dry_run: false
```

### Settings

| Setting | Description | Required | Environment Variable |
|---------|-------------|----------|---------------------|
| `hosts` | Elasticsearch endpoints | Yes | - |
| `username` | Admin username | Yes | `ELASTAUTH_ELASTICSEARCH_USERNAME` |
| `password` | Admin password | Yes | `ELASTAUTH_ELASTICSEARCH_PASSWORD` |
| `dry_run` | Test mode (no user creation) | No | `ELASTAUTH_ELASTICSEARCH_DRY_RUN` |

## Authentication Providers

### Authelia Provider

```yaml
auth_provider: "authelia"

authelia:
  header_username: "Remote-User"
  header_groups: "Remote-Groups"
  header_email: "Remote-Email"
  header_name: "Remote-Name"
```

**Environment variables:**
- `ELASTAUTH_AUTHELIA_HEADER_USERNAME=Remote-User`
- `ELASTAUTH_AUTHELIA_HEADER_GROUPS=Remote-Groups`
- `ELASTAUTH_AUTHELIA_HEADER_EMAIL=Remote-Email`
- `ELASTAUTH_AUTHELIA_HEADER_NAME=Remote-Name`

See [Authelia Provider](/providers/authelia/) for details.

### OIDC Provider

```yaml
auth_provider: "oidc"

oidc:
  issuer: "https://your-provider.com"
  client_id: "elastauth"
  client_secret: "your-secret"
  
  scopes:
    - openid
    - profile
    - email
  
  claim_mappings:
    username: "preferred_username"
    email: "email"
    groups: "groups"
    full_name: "name"
```

**Environment variables:**
- `ELASTAUTH_OIDC_ISSUER=https://provider.com`
- `ELASTAUTH_OIDC_CLIENT_ID=elastauth`
- `ELASTAUTH_OIDC_CLIENT_SECRET=secret`
- `ELASTAUTH_OIDC_SCOPES=openid,profile,email`

See [OIDC Provider](/providers/oidc/) for details.

## Cache Configuration

```yaml
cache:
  type: "redis"  # Options: memory, redis, file, or empty
  expiration: "1h"
  redis_host: "redis:6379"
  redis_db: 0
```

### Cache Types

| Type | Description | Horizontal Scaling | Environment Variable |
|------|-------------|-------------------|---------------------|
| `memory` | In-memory cache | ❌ Single instance only | `ELASTAUTH_CACHE_TYPE=memory` |
| `redis` | Redis cache | ✅ Supports scaling | `ELASTAUTH_CACHE_TYPE=redis` |
| `file` | File-based cache | ❌ Single instance only | `ELASTAUTH_CACHE_TYPE=file` |
| empty | No caching | ✅ Supports scaling | `ELASTAUTH_CACHE_TYPE=` |

### Redis Cache Settings

```yaml
cache:
  type: "redis"
  redis_host: "redis:6379"
  redis_db: 0
  expiration: "1h"
```

**Environment variables:**
- `ELASTAUTH_CACHE_REDIS_HOST=redis:6379`
- `ELASTAUTH_CACHE_REDIS_DB=0`
- `ELASTAUTH_CACHE_EXPIRATION=1h`

See [Cache Configuration](/cache/) for details.

## Security Configuration

### Secret Key

The secret key is used for AES-256 encryption of cached credentials.

```yaml
secret_key: "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
```

**Environment variable:** `ELASTAUTH_SECRET_KEY=your-64-char-hex-key`

**Generate a key:**
```bash
# Using elastauth
elastauth --generateKey

# Using OpenSSL
openssl rand -hex 32
```

:::danger
The secret key must be:
- Exactly 64 hexadecimal characters (32 bytes)
- Kept secret and secure
- Same across all instances when using Redis cache
:::

### Default Roles

```yaml
default_roles:
  - "kibana_user"
  - "monitoring_user"
```

These Elasticsearch roles are assigned to all authenticated users.

## Monitoring Configuration

```yaml
enable_metrics: true  # Enable Prometheus metrics at /metrics
enableOtel: false     # Enable OpenTelemetry tracing
```

**Environment variables:**
- `ELASTAUTH_ENABLE_METRICS=true`
- `ELASTAUTH_ENABLE_OTEL=false`

## Environment Variables Reference

### Complete List

All configuration options support environment variables with `ELASTAUTH_` prefix:

```bash
# Core
ELASTAUTH_AUTH_PROVIDER=authelia
ELASTAUTH_SECRET_KEY=your-hex-key
ELASTAUTH_LISTEN=0.0.0.0:3000
ELASTAUTH_LOG_LEVEL=info
ELASTAUTH_LOG_FORMAT=json

# Proxy Mode
ELASTAUTH_PROXY_ENABLED=true
ELASTAUTH_PROXY_ELASTICSEARCH_URL=https://es:9200
ELASTAUTH_PROXY_TIMEOUT=30s
ELASTAUTH_PROXY_MAX_IDLE_CONNS=100
ELASTAUTH_PROXY_IDLE_CONN_TIMEOUT=90s
ELASTAUTH_PROXY_TLS_ENABLED=true
ELASTAUTH_PROXY_TLS_INSECURE_SKIP_VERIFY=false
ELASTAUTH_PROXY_TLS_CA_CERT=/certs/ca.crt
ELASTAUTH_PROXY_TLS_CLIENT_CERT=/certs/client.crt
ELASTAUTH_PROXY_TLS_CLIENT_KEY=/certs/client.key

# Elasticsearch
ELASTAUTH_ELASTICSEARCH_USERNAME=elastic
ELASTAUTH_ELASTICSEARCH_PASSWORD=changeme
ELASTAUTH_ELASTICSEARCH_DRY_RUN=false

# Cache
ELASTAUTH_CACHE_TYPE=redis
ELASTAUTH_CACHE_EXPIRATION=1h
ELASTAUTH_CACHE_REDIS_HOST=redis:6379
ELASTAUTH_CACHE_REDIS_DB=0

# Authelia
ELASTAUTH_AUTHELIA_HEADER_USERNAME=Remote-User
ELASTAUTH_AUTHELIA_HEADER_GROUPS=Remote-Groups
ELASTAUTH_AUTHELIA_HEADER_EMAIL=Remote-Email
ELASTAUTH_AUTHELIA_HEADER_NAME=Remote-Name

# OIDC
ELASTAUTH_OIDC_ISSUER=https://provider.com
ELASTAUTH_OIDC_CLIENT_ID=elastauth
ELASTAUTH_OIDC_CLIENT_SECRET=secret
ELASTAUTH_OIDC_SCOPES=openid,profile,email
ELASTAUTH_OIDC_CLAIM_MAPPINGS_USERNAME=preferred_username
ELASTAUTH_OIDC_CLAIM_MAPPINGS_EMAIL=email
ELASTAUTH_OIDC_CLAIM_MAPPINGS_GROUPS=groups
ELASTAUTH_OIDC_CLAIM_MAPPINGS_FULL_NAME=name

# Monitoring
ELASTAUTH_ENABLE_METRICS=true
ELASTAUTH_ENABLE_OTEL=false
```

## Configuration Examples

### Authentication-Only Mode

```yaml
auth_provider: "authelia"

elasticsearch:
  hosts: ["https://elasticsearch:9200"]
  username: "elastic"
  password: "changeme"

secret_key: "your-64-char-hex-key"
default_roles: ["kibana_user"]

cache:
  type: "redis"
  redis_host: "redis:6379"
  expiration: "1h"
```

### Transparent Proxy Mode

```yaml
auth_provider: "authelia"

proxy:
  enabled: true
  elasticsearch_url: "https://elasticsearch:9200"
  timeout: "30s"
  tls:
    enabled: true
    ca_cert: "/certs/ca.crt"

elasticsearch:
  hosts: ["https://elasticsearch:9200"]
  username: "elastic"
  password: "changeme"

secret_key: "your-64-char-hex-key"
default_roles: ["kibana_user"]

cache:
  type: "redis"
  redis_host: "redis:6379"
  expiration: "1h"
```

### OIDC with Proxy Mode

```yaml
auth_provider: "oidc"

proxy:
  enabled: true
  elasticsearch_url: "https://elasticsearch:9200"

oidc:
  issuer: "https://keycloak.example.com/realms/myrealm"
  client_id: "elastauth"
  client_secret: "your-secret"
  scopes:
    - openid
    - profile
    - email
    - groups
  claim_mappings:
    username: "preferred_username"
    email: "email"
    groups: "groups"
    full_name: "name"

elasticsearch:
  hosts: ["https://elasticsearch:9200"]
  username: "elastic"
  password: "changeme"

secret_key: "your-64-char-hex-key"
default_roles: ["kibana_user"]

cache:
  type: "redis"
  redis_host: "redis:6379"
  expiration: "1h"
```

## Next Steps

- [Deployment Modes](/deployment/) - Choose between auth-only and proxy mode
- [Authentication Providers](/providers/) - Configure your auth provider
- [Cache Configuration](/cache/) - Set up caching for performance
- [Troubleshooting](/guides/troubleshooting/) - Common issues and solutions
