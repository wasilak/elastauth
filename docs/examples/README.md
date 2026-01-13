# Configuration Examples

This directory contains complete configuration examples for different deployment scenarios and authentication providers.

## Available Examples

### Authentication Providers

- **[Authelia + Redis](authelia-redis.md)** - Authelia with Redis cache
- **[Keycloak + Memory](keycloak-memory.md)** - Keycloak with memory cache
- **[Casdoor + File](casdoor-file.md)** - Casdoor with file cache
- **[Auth0 + No Cache](auth0-nocache.md)** - Auth0 without caching

### Deployment Scenarios

- **[Single Instance](single-instance.md)** - Simple single-instance deployment
- **[Multi-Instance](multi-instance.md)** - Load-balanced multi-instance setup
- **[Kubernetes](kubernetes.md)** - Kubernetes deployment with ConfigMaps/Secrets
- **[Docker Compose](docker-compose.md)** - Complete Docker Compose stack

### Integration Examples

- **[Traefik Integration](traefik.md)** - Using elastauth with Traefik
- **[Nginx Integration](nginx.md)** - Using elastauth with Nginx
- **[Elasticsearch Cluster](elasticsearch-cluster.md)** - Multi-node Elasticsearch setup

## Quick Start Examples

### Minimal Authelia Setup

```yaml
# config.yml
auth_provider: "authelia"

elasticsearch:
  hosts: ["http://localhost:9200"]
  username: "elastauth"
  password: "password"

default_roles: ["kibana_user"]
secret_key: "your-32-character-secret-key-here"
```

### Minimal OIDC Setup

```yaml
# config.yml
auth_provider: "oidc"

oidc:
  issuer: "https://your-provider.com"
  client_id: "elastauth"
  client_secret: "your-client-secret"
  claim_mappings:
    username: "preferred_username"
    email: "email"
    groups: "groups"
    full_name: "name"

elasticsearch:
  hosts: ["http://localhost:9200"]
  username: "elastauth"
  password: "password"

default_roles: ["kibana_user"]
secret_key: "your-32-character-secret-key-here"
```

## Environment Variables

All examples support environment variable overrides:

```bash
# Common environment variables
ELASTICSEARCH_PASSWORD="your-elasticsearch-password"
SECRET_KEY="your-32-character-secret-key-here"

# Provider-specific
OIDC_CLIENT_SECRET="your-oidc-client-secret"
KEYCLOAK_SECRET="your-keycloak-secret"
CASDOOR_SECRET="your-casdoor-secret"

# Cache configuration
CACHE_REDIS_HOST="redis:6379"
REDIS_PASSWORD="your-redis-password"
```

## Configuration Validation

Test your configuration:

```bash
# Validate configuration
./elastauth --config config.yml --validate

# Test with dry-run mode
./elastauth --config config.yml --dry-run
```

## Next Steps

Choose an example that matches your deployment scenario:

1. **Start Simple**: Begin with a single-instance example
2. **Add Caching**: Add Redis cache for better performance
3. **Scale Up**: Move to multi-instance when needed
4. **Production**: Use Kubernetes example for production deployments