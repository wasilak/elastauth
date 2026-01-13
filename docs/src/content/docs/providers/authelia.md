---
title: Authelia Provider
description: Configure header-based authentication with Authelia
---

# Authelia Provider

The Authelia provider extracts user information from HTTP headers set by [Authelia](https://www.authelia.com/), a popular authentication and authorization server.

## Overview

Authelia typically runs as a forward authentication service that:

1. Authenticates users through various methods (LDAP, file, etc.)
2. Sets HTTP headers with user information
3. Forwards requests to protected applications

The elastauth Authelia provider reads these headers and creates Elasticsearch users accordingly.

## Configuration

### Basic Configuration

```yaml
auth_provider: "authelia"

authelia:
  header_username: "Remote-User"     # Header containing username
  header_groups: "Remote-Groups"     # Header containing user groups
  header_email: "Remote-Email"       # Header containing user email
  header_name: "Remote-Name"         # Header containing full name
```

### Default Headers

If not specified, the provider uses these default header names:

- `Remote-User` - Username (required)
- `Remote-Groups` - Comma-separated groups (optional)
- `Remote-Email` - User email address (optional)
- `Remote-Name` - User's full name (optional)

### Environment Variable Overrides

```bash
# Override header names via environment variables
AUTHELIA_HEADER_USERNAME="X-Remote-User"
AUTHELIA_HEADER_GROUPS="X-Remote-Groups"
AUTHELIA_HEADER_EMAIL="X-Remote-Email"
AUTHELIA_HEADER_NAME="X-Remote-Name"
```

## Authelia Integration

### Forward Auth Configuration

Configure Authelia to set the required headers:

```yaml
# authelia/configuration.yml
server:
  headers:
    authelia_url: 'https://auth.example.com'

authentication_backend:
  # Your authentication backend configuration

access_control:
  default_policy: deny
  rules:
    - domain: 'kibana.example.com'
      policy: one_factor
      
# Configure headers to be sent to elastauth
headers:
  Remote-User: '{user}'
  Remote-Groups: '{groups}'
  Remote-Email: '{email}'
  Remote-Name: '{name}'
```

### Traefik Configuration

If using Traefik with Authelia:

```yaml
# docker-compose.yml or traefik config
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
```

## Header Format

### Username Header

The username header must contain a single username:

```
Remote-User: john.doe
```

### Groups Header

Groups can be provided in several formats:

```
# Comma-separated
Remote-Groups: admin,developers,users

# Single group
Remote-Groups: admin

# Empty (no groups)
Remote-Groups: 
```

### Email Header

Standard email format:

```
Remote-Email: john.doe@example.com
```

### Name Header

Full name of the user:

```
Remote-Name: John Doe
```

## Validation

The Authelia provider validates:

- **Username**: Required, must be non-empty and valid format
- **Groups**: Optional, validated if present
- **Email**: Optional, must be valid email format if present
- **Name**: Optional, validated if present

## Error Handling

Common error scenarios:

### Missing Username Header

```json
{
  "message": "Remote-User header not found",
  "code": 400,
  "timestamp": "2024-01-15T10:30:00Z"
}
```

### Invalid Email Format

```json
{
  "message": "Invalid email format",
  "code": 400,
  "timestamp": "2024-01-15T10:30:00Z"
}
```

## Configuration Examples

### Basic Single-Instance Setup

```yaml
# Simple Authelia setup with memory cache
auth_provider: "authelia"

authelia:
  header_username: "Remote-User"
  header_groups: "Remote-Groups"
  header_email: "Remote-Email"
  header_name: "Remote-Name"

cache:
  type: "memory"
  expiration: "1h"

elasticsearch:
  hosts: ["https://elasticsearch:9200"]
  username: "elastauth"
  password: "${ELASTICSEARCH_PASSWORD}"

default_roles:
  - "kibana_user"

group_mappings:
  admin:
    - "kibana_admin"
    - "superuser"
  users:
    - "kibana_user"
```

### Production Multi-Instance Setup

```yaml
# Production Authelia setup with Redis cache for scaling
auth_provider: "authelia"

authelia:
  header_username: "Remote-User"
  header_groups: "Remote-Groups"
  header_email: "Remote-Email"
  header_name: "Remote-Name"

cache:
  type: "redis"
  expiration: "2h"
  redis_host: "${REDIS_HOST}"
  redis_db: 0

elasticsearch:
  hosts:
    - "${ELASTICSEARCH_HOST_1}"
    - "${ELASTICSEARCH_HOST_2}"
    - "${ELASTICSEARCH_HOST_3}"
  username: "${ELASTICSEARCH_USERNAME}"
  password: "${ELASTICSEARCH_PASSWORD}"

# Encryption key must be identical across all instances
secret_key: "${SECRET_KEY}"

default_roles:
  - "kibana_user"

group_mappings:
  # Administrative groups
  admins:
    - "kibana_admin"
    - "superuser"
  security-team:
    - "kibana_admin"
    - "security_manager"
  
  # Developer groups
  senior-developers:
    - "kibana_user"
    - "dev_full_access"
  developers:
    - "kibana_user"
    - "dev_limited_access"
  
  # Analyst groups
  data-analysts:
    - "kibana_user"
    - "read_only"
  business-users:
    - "kibana_user"
    - "dashboard_only"
  
  # Operations groups
  sre:
    - "kibana_user"
    - "monitoring_user"
  support:
    - "kibana_user"
    - "read_only"
```

### Custom Headers Configuration

```yaml
# Custom header names for specific Authelia setups
auth_provider: "authelia"

authelia:
  header_username: "X-Forwarded-User"      # Custom username header
  header_groups: "X-Forwarded-Groups"      # Custom groups header
  header_email: "X-Forwarded-Email"        # Custom email header
  header_name: "X-Forwarded-Name"          # Custom name header

cache:
  type: "file"
  expiration: "30m"
  path: "/var/cache/elastauth"

elasticsearch:
  hosts: ["https://elasticsearch:9200"]
  username: "elastauth"
  password: "${ELASTICSEARCH_PASSWORD}"

default_roles:
  - "kibana_user"

group_mappings:
  administrators:
    - "kibana_admin"
  standard-users:
    - "kibana_user"
```

### High-Availability Configuration

```yaml
# Enterprise-grade Authelia setup with comprehensive settings
auth_provider: "authelia"

authelia:
  header_username: "${AUTHELIA_USERNAME_HEADER:-Remote-User}"
  header_groups: "${AUTHELIA_GROUPS_HEADER:-Remote-Groups}"
  header_email: "${AUTHELIA_EMAIL_HEADER:-Remote-Email}"
  header_name: "${AUTHELIA_NAME_HEADER:-Remote-Name}"

# High-availability cache
cache:
  type: "redis"
  expiration: "${CACHE_EXPIRATION:-4h}"
  redis_host: "${REDIS_HOST}"
  redis_db: "${REDIS_DB:-0}"

# Multi-node Elasticsearch cluster
elasticsearch:
  hosts:
    - "${ELASTICSEARCH_HOST_1}"
    - "${ELASTICSEARCH_HOST_2}"
    - "${ELASTICSEARCH_HOST_3}"
  username: "${ELASTICSEARCH_USERNAME}"
  password: "${ELASTICSEARCH_PASSWORD}"
  dry_run: false

# Security configuration
secret_key: "${SECRET_KEY}"

# Comprehensive role mappings
default_roles:
  - "kibana_user"

group_mappings:
  # C-level and executives
  "c-level":
    - "kibana_admin"
    - "superuser"
  
  # IT and platform teams
  "platform-engineering":
    - "kibana_admin"
    - "cluster_admin"
  "security-engineering":
    - "kibana_admin"
    - "security_manager"
  "devops":
    - "kibana_user"
    - "monitoring_user"
    - "watcher_user"
  
  # Development teams
  "tech-leads":
    - "kibana_user"
    - "dev_full_access"
  "senior-engineers":
    - "kibana_user"
    - "dev_full_access"
  "engineers":
    - "kibana_user"
    - "dev_limited_access"
  "qa-engineers":
    - "kibana_user"
    - "test_data_access"
  
  # Data and analytics teams
  "data-engineering":
    - "kibana_user"
    - "ml_user"
    - "transform_user"
  "data-science":
    - "kibana_user"
    - "ml_user"
  "business-intelligence":
    - "kibana_user"
    - "dashboard_only"
  
  # Business teams
  "product-managers":
    - "kibana_user"
    - "dashboard_only"
  "business-analysts":
    - "kibana_user"
    - "read_only"
  "customer-success":
    - "kibana_user"
    - "read_only"
  
  # Support and operations
  "support-tier1":
    - "kibana_user"
    - "read_only"
  "support-tier2":
    - "kibana_user"
    - "support_advanced"
  "operations":
    - "kibana_user"
    - "monitoring_user"
```

## Complete Example

### elastauth Configuration

```yaml
# config.yml
auth_provider: "authelia"

authelia:
  header_username: "Remote-User"
  header_groups: "Remote-Groups"
  header_email: "Remote-Email"
  header_name: "Remote-Name"

# Cache configuration
cache:
  type: "redis"
  expiration: "1h"
  redis_host: "redis:6379"
  redis_db: 0

# Elasticsearch configuration
elasticsearch:
  hosts:
    - "https://elasticsearch:9200"
  username: "elastauth"
  password: "${ELASTICSEARCH_PASSWORD}"

# Role mappings
default_roles:
  - "kibana_user"

group_mappings:
  admin:
    - "kibana_admin"
    - "superuser"
  developers:
    - "kibana_user"
    - "dev_role"
```

### Docker Compose Example

```yaml
version: '3.8'

services:
  authelia:
    image: authelia/authelia:latest
    volumes:
      - ./authelia:/config
    environment:
      - AUTHELIA_JWT_SECRET=${AUTHELIA_JWT_SECRET}
    
  elastauth:
    image: elastauth:latest
    environment:
      - ELASTICSEARCH_PASSWORD=${ELASTICSEARCH_PASSWORD}
      - SECRET_KEY=${SECRET_KEY}
    volumes:
      - ./elastauth-config.yml:/config.yml
    depends_on:
      - authelia
      - elasticsearch
      - redis
      
  traefik:
    image: traefik:latest
    command:
      - --api.insecure=true
      - --providers.docker=true
    ports:
      - "80:80"
      - "8080:8080"
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
    labels:
      - "traefik.http.middlewares.authelia.forwardauth.address=http://authelia:9091/api/verify?rd=https://auth.example.com"
      - "traefik.http.middlewares.authelia.forwardauth.trustForwardHeader=true"
      - "traefik.http.middlewares.authelia.forwardauth.authResponseHeaders=Remote-User,Remote-Groups,Remote-Email,Remote-Name"
```

## Troubleshooting

### Headers Not Received

1. **Check Authelia Configuration**: Ensure headers are configured in Authelia
2. **Verify Proxy Settings**: Confirm proxy (Traefik/nginx) forwards headers
3. **Test Headers**: Use curl to verify headers are present:

```bash
curl -H "Remote-User: testuser" \
     -H "Remote-Groups: admin" \
     http://elastauth:5000/
```

### Authentication Failures

1. **Check Logs**: Review elastauth logs for specific error messages
2. **Validate Headers**: Ensure header names match configuration
3. **Test Connectivity**: Verify Elasticsearch connectivity

### Group Mapping Issues

1. **Check Group Format**: Ensure groups are comma-separated
2. **Verify Mappings**: Check `group_mappings` configuration
3. **Default Roles**: Ensure `default_roles` are configured

## Migration from Direct Integration

If migrating from direct Authelia-Elasticsearch integration:

1. **Update Authelia**: Configure headers instead of direct Elasticsearch integration
2. **Deploy elastauth**: Add elastauth between Authelia and Elasticsearch
3. **Update Clients**: Point Kibana/clients to elastauth instead of Elasticsearch
4. **Test Authentication**: Verify user authentication and role mapping

## Security Considerations

- **Header Validation**: elastauth validates all header content
- **Input Sanitization**: All user input is sanitized before processing
- **Credential Encryption**: Cached credentials are encrypted
- **Network Security**: Use HTTPS for all communications

## Next Steps

### Alternative Authentication
- **[OAuth2/OIDC Provider](/providers/oidc)** - JWT token-based authentication
  - [Keycloak Integration](/providers/oidc#keycloak-configuration) - Keycloak setup
  - [Casdoor Integration](/providers/oidc#casdoor-configuration) - Casdoor setup
  - [Authentik Integration](/providers/oidc#authentik-configuration) - Authentik setup

### Configuration and Optimization
- **[Cache Configuration](/cache/)** - Optimize performance with caching
  - [Redis Cache](/cache/redis) - Distributed caching for production
  - [Memory Cache](/cache/) - Simple single-instance caching
- **[Configuration Reference](/guides/configuration)** - Complete configuration options
- **[Environment Variables](/guides/environment)** - Environment-based configuration

### Deployment and Scaling
- **[Docker Deployment](/deployment/docker)** - Container deployment with Authelia
- **[Kubernetes Deployment](/deployment/kubernetes)** - Production Kubernetes setup
- **[Scaling Guide](/guides/scaling)** - Horizontal scaling considerations
- **[Load Balancer Setup](/guides/load-balancer)** - Production load balancing

### Integration and Setup
- **[Authelia Integration Guide](/guides/authelia)** - Detailed Authelia setup
- **[Traefik Configuration](/guides/traefik)** - Traefik forward auth setup
- **[Nginx Configuration](/guides/nginx)** - Nginx auth_request setup
- **[Elasticsearch Setup](/guides/elasticsearch)** - Elasticsearch configuration

### Operations and Troubleshooting
- **[Troubleshooting Authelia](/guides/troubleshooting#authelia-issues)** - Common Authelia issues
- **[Header Debugging](/guides/header-debugging)** - HTTP header troubleshooting
- **[Monitoring](/guides/monitoring)** - Health checks and observability
- **[Security Best Practices](/guides/security)** - Authelia security considerations