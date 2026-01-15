---
title: OAuth2/OIDC Provider
description: Configure JWT token authentication with OAuth2/OIDC providers
---

The OAuth2/OIDC provider validates JWT tokens from any OAuth2/OIDC-compliant authentication system. It extracts user information from JWT claims and creates corresponding Elasticsearch users.

## Supported Systems

This provider works with any OAuth2/OIDC-compliant system including:
- Keycloak
- Authentik
- Auth0
- Azure AD
- Pocket-ID
- Ory Hydra
- And many other OIDC-compliant providers

## Configuration

### Basic Configuration

```yaml
auth_provider: "oidc"

oidc:
  issuer: "https://auth.example.com"
  client_id: "elastauth"
  client_secret: "${OIDC_CLIENT_SECRET}"
  
  scopes: ["openid", "profile", "email", "groups"]
  
  claim_mappings:
    username: "preferred_username"
    email: "email"
    groups: "groups"
    full_name: "name"
```

### Advanced Configuration

```yaml
oidc:
  # Standard OAuth2/OIDC settings
  issuer: "https://auth.example.com"
  client_id: "elastauth"
  client_secret: "${OIDC_CLIENT_SECRET}"
  
  # Manual endpoint configuration (optional - uses discovery by default)
  authorization_endpoint: "https://auth.example.com/auth"
  token_endpoint: "https://auth.example.com/token"
  userinfo_endpoint: "https://auth.example.com/userinfo"
  jwks_uri: "https://auth.example.com/.well-known/jwks.json"
  
  # OAuth2 settings
  scopes: ["openid", "profile", "email", "groups"]
  client_auth_method: "client_secret_basic"  # or "client_secret_post"
  
  # Token validation method
  token_validation: "jwks"  # "userinfo", or "both"
  
  # Claim mappings (supports nested claims)
  claim_mappings:
    username: "preferred_username"
    email: "email"
    groups: "realm_access.roles"  # Nested claim example
    full_name: "name"
  
  # Security settings
  use_pkce: true
  
  # Custom headers (if required by provider)
  custom_headers:
    "X-Custom-Header": "value"
```

## Provider Examples

### Keycloak

```yaml
oidc:
  issuer: "https://keycloak.example.com/realms/myrealm"
  client_id: "elastauth"
  client_secret: "${KEYCLOAK_SECRET}"
  scopes: ["openid", "profile", "email"]
  claim_mappings:
    username: "preferred_username"
    email: "email"
    groups: "realm_access.roles"
    full_name: "name"
```

### Authentik

```yaml
oidc:
  issuer: "https://authentik.example.com/application/o/elastauth/"
  client_id: "elastauth"
  client_secret: "${AUTHENTIK_SECRET}"
  scopes: ["openid", "profile", "email", "groups"]
  claim_mappings:
    username: "preferred_username"
    email: "email"
    groups: "groups"
    full_name: "name"
```

### Auth0

```yaml
oidc:
  issuer: "https://your-tenant.auth0.com/"
  client_id: "your-client-id"
  client_secret: "${AUTH0_SECRET}"
  scopes: ["openid", "profile", "email"]
  claim_mappings:
    username: "nickname"
    email: "email"
    groups: "https://your-app.com/roles"  # Custom claim
    full_name: "name"
```

### Azure AD

```yaml
oidc:
  issuer: "https://login.microsoftonline.com/your-tenant-id/v2.0"
  client_id: "your-application-id"
  client_secret: "${AZURE_SECRET}"
  scopes: ["openid", "profile", "email"]
  claim_mappings:
    username: "preferred_username"
    email: "email"
    groups: "groups"
    full_name: "name"
```

## Environment Variables

Override OIDC configuration using environment variables:

```bash
OIDC_ISSUER="https://auth.example.com"
OIDC_CLIENT_ID="elastauth"
OIDC_CLIENT_SECRET="your-secret"
OIDC_SCOPES="openid,profile,email,groups"
OIDC_USERNAME_CLAIM="preferred_username"
OIDC_EMAIL_CLAIM="email"
OIDC_GROUPS_CLAIM="groups"
OIDC_NAME_CLAIM="name"
```

## Token Validation

### JWKS Validation (Recommended)
Validates JWT tokens using the provider's public keys:
```yaml
oidc:
  token_validation: "jwks"
```

### Userinfo Validation
Validates tokens by calling the userinfo endpoint:
```yaml
oidc:
  token_validation: "userinfo"
```

### Both Methods
Uses JWKS first, falls back to userinfo:
```yaml
oidc:
  token_validation: "both"
```

## Complete Configuration Example

```yaml
auth_provider: "oidc"

oidc:
  issuer: "https://auth.example.com"
  client_id: "elastauth"
  client_secret: "${OIDC_CLIENT_SECRET}"
  scopes: ["openid", "profile", "email", "groups"]
  token_validation: "both"
  use_pkce: true
  
  claim_mappings:
    username: "preferred_username"
    email: "email"
    groups: "groups"
    full_name: "name"

cache:
  type: "redis"
  expiration: "2h"
  redis_host: "redis:6379"

elasticsearch:
  hosts: ["https://elasticsearch:9200"]
  username: "elastauth"
  password: "${ELASTICSEARCH_PASSWORD}"

secret_key: "${SECRET_KEY}"

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

## Troubleshooting

### Token Validation Failures

**Symptoms**: 401 errors, "invalid token" messages

**Solutions**:
1. Verify JWT token is valid and not expired
2. Check issuer URL matches exactly (including trailing slashes)
3. Verify JWKS endpoint is accessible
4. Test token validation method (try "both" if having issues)

### Claim Mapping Issues

**Symptoms**: Missing user information, empty groups

**Solutions**:
1. Check JWT token contains expected claims
2. Verify claim mapping paths are correct
3. Test with nested claim paths (e.g., "realm_access.roles")
4. Use token debugging tools to inspect JWT contents

### Connection Issues

**Symptoms**: elastauth fails to start, discovery errors

**Solutions**:
1. Verify issuer URL is accessible
2. Check network connectivity to OIDC provider
3. Verify client credentials are correct
4. Test OIDC discovery endpoint manually

## Related Documentation

- **[Authelia Provider](/elastauth/providers/authelia)** - Alternative header-based authentication
- **[Cache Configuration](/elastauth/cache/)** - Optimize performance with caching
  - [Redis Cache](/elastauth/cache/redis) - Distributed caching for production
- **[Troubleshooting](/elastauth/guides/troubleshooting#oauth2oidc-issues)** - Common OAuth2/OIDC issues