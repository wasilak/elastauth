# OAuth2/OIDC Provider

The OAuth2/OIDC provider validates JWT tokens and extracts user information from any OAuth2/OIDC-compliant authentication system. This single provider works with multiple authentication systems through standard protocols.

## Overview

The OIDC provider supports:

- **JWT Token Validation**: Validates tokens using JWKS or userinfo endpoints
- **Claim Mapping**: Configurable mapping of JWT claims to user information
- **Multiple Auth Systems**: Works with any OAuth2/OIDC-compliant provider
- **Flexible Configuration**: Supports both discovery and manual endpoint configuration

## Supported Authentication Systems

- **[Keycloak](https://www.keycloak.org/)** - Open source identity management
- **[Casdoor](https://casdoor.org/)** - Web-based identity management
- **[Authentik](https://goauthentik.io/)** - Modern authentication platform
- **[Auth0](https://auth0.com/)** - Cloud identity platform
- **[Azure AD](https://azure.microsoft.com/en-us/services/active-directory/)** - Microsoft identity platform
- **[Pocket-ID](https://github.com/stonith404/pocket-id)** - Self-hosted identity provider
- **[Ory Hydra](https://www.ory.sh/hydra/)** - OAuth2 and OpenID Connect server
- **Authelia (OIDC mode)** - When configured as OIDC provider

## Configuration

### Basic Configuration

```yaml
auth_provider: "oidc"

oidc:
  # Standard OAuth2/OIDC settings
  issuer: "https://auth.example.com"
  client_id: "elastauth"
  client_secret: "${OIDC_CLIENT_SECRET}"
  
  # Scopes to request
  scopes: ["openid", "profile", "email", "groups"]
  
  # Claim mappings
  claim_mappings:
    username: "preferred_username"
    email: "email"
    groups: "groups"
    full_name: "name"
```

### Advanced Configuration

```yaml
oidc:
  # Standard settings
  issuer: "https://auth.example.com"
  client_id: "elastauth"
  client_secret: "${OIDC_CLIENT_SECRET}"
  
  # Manual endpoint configuration (optional)
  authorization_endpoint: "https://auth.example.com/auth"
  token_endpoint: "https://auth.example.com/token"
  userinfo_endpoint: "https://auth.example.com/userinfo"
  jwks_uri: "https://auth.example.com/.well-known/jwks.json"
  
  # OAuth2 settings
  scopes: ["openid", "profile", "email", "groups"]
  client_auth_method: "client_secret_basic"  # or "client_secret_post"
  
  # Token validation method
  token_validation: "jwks"  # or "userinfo", or "both"
  
  # Claim mappings with nested claim support
  claim_mappings:
    username: "preferred_username"
    email: "email"
    groups: "realm_access.roles"  # Nested claim example
    full_name: "name"
  
  # Custom headers for provider-specific requirements
  custom_headers:
    "X-Custom-Header": "value"
  
  # Security settings
  use_pkce: true
```

## Provider-Specific Examples

### Keycloak

```yaml
auth_provider: "oidc"

oidc:
  issuer: "https://keycloak.example.com/realms/myrealm"
  client_id: "elastauth"
  client_secret: "${KEYCLOAK_SECRET}"
  scopes: ["openid", "profile", "email", "roles"]
  claim_mappings:
    username: "preferred_username"
    email: "email"
    groups: "realm_access.roles"  # Keycloak realm roles
    full_name: "name"
  token_validation: "jwks"
```

### Casdoor

```yaml
auth_provider: "oidc"

oidc:
  issuer: "https://casdoor.example.com"
  client_id: "elastauth-app"
  client_secret: "${CASDOOR_SECRET}"
  scopes: ["openid", "profile", "email"]
  claim_mappings:
    username: "name"           # Casdoor uses 'name' for username
    email: "email"
    groups: "roles"            # Casdoor roles
    full_name: "displayName"   # Casdoor display name
  token_validation: "jwks"
```

### Authentik

```yaml
auth_provider: "oidc"

oidc:
  issuer: "https://authentik.example.com/application/o/elastauth/"
  client_id: "elastauth"
  client_secret: "${AUTHENTIK_SECRET}"
  scopes: ["openid", "profile", "email", "groups"]
  claim_mappings:
    username: "preferred_username"
    email: "email"
    groups: "groups"           # Authentik groups
    full_name: "name"
  token_validation: "jwks"
```

### Auth0

```yaml
auth_provider: "oidc"

oidc:
  issuer: "https://your-tenant.auth0.com/"
  client_id: "your-auth0-client-id"
  client_secret: "${AUTH0_SECRET}"
  scopes: ["openid", "profile", "email"]
  claim_mappings:
    username: "sub"            # Auth0 uses 'sub' for username
    email: "email"
    groups: "https://example.com/roles"  # Custom claim for roles
    full_name: "name"
  token_validation: "jwks"
```

### Azure AD

```yaml
auth_provider: "oidc"

oidc:
  issuer: "https://login.microsoftonline.com/your-tenant-id/v2.0"
  client_id: "your-azure-app-id"
  client_secret: "${AZURE_SECRET}"
  scopes: ["openid", "profile", "email"]
  claim_mappings:
    username: "preferred_username"
    email: "email"
    groups: "groups"           # Requires group claims configuration
    full_name: "name"
  token_validation: "jwks"
```

## Authentication Methods

### Bearer Token Authentication

Client sends JWT token in Authorization header:

```bash
curl -H "Authorization: Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9..." \
     http://elastauth:5000/
```

### Cookie-Based Authentication

Client sends JWT token in cookie:

```bash
curl -H "Cookie: access_token=eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9..." \
     http://elastauth:5000/
```

## Token Validation Methods

### JWKS Validation

Validates tokens using the provider's JWKS endpoint:

```yaml
oidc:
  token_validation: "jwks"
  # Automatically discovers JWKS endpoint from issuer
```

### Userinfo Validation

Validates tokens by calling the userinfo endpoint:

```yaml
oidc:
  token_validation: "userinfo"
  # Uses token to call userinfo endpoint
```

### Both Methods

Tries JWKS first, falls back to userinfo:

```yaml
oidc:
  token_validation: "both"
  # More flexible but potentially slower
```

## Claim Mapping

### Simple Claims

Map standard JWT claims:

```yaml
claim_mappings:
  username: "preferred_username"
  email: "email"
  full_name: "name"
```

### Nested Claims

Access nested claims using dot notation:

```yaml
claim_mappings:
  username: "preferred_username"
  groups: "realm_access.roles"        # Keycloak realm roles
  # or
  groups: "resource_access.account.roles"  # Keycloak client roles
```

### Group Formats

Handle different group claim formats:

```yaml
# Array of strings
"groups": ["admin", "users"]

# Single string
"groups": "admin"

# Nested object
"realm_access": {
  "roles": ["admin", "users"]
}
```

## Environment Variables

Override sensitive configuration via environment variables:

```bash
# OAuth2 credentials
OIDC_CLIENT_SECRET="your-client-secret"
OIDC_CLIENT_ID="your-client-id"

# Provider endpoints (if not using discovery)
OIDC_ISSUER="https://auth.example.com"
OIDC_AUTHORIZATION_ENDPOINT="https://auth.example.com/auth"
OIDC_TOKEN_ENDPOINT="https://auth.example.com/token"

# Custom configuration
OIDC_TOKEN_VALIDATION="jwks"
```

## Error Handling

### Token Validation Errors

```json
{
  "message": "Token validation failed: invalid signature",
  "code": 400,
  "timestamp": "2024-01-15T10:30:00Z"
}
```

### Missing Claims

```json
{
  "message": "Required claim 'preferred_username' not found in token",
  "code": 400,
  "timestamp": "2024-01-15T10:30:00Z"
}
```

### Provider Connectivity

```json
{
  "message": "Failed to fetch JWKS from provider",
  "code": 500,
  "timestamp": "2024-01-15T10:30:00Z"
}
```

## Complete Example

### elastauth Configuration

```yaml
# config.yml
auth_provider: "oidc"

oidc:
  issuer: "https://keycloak.example.com/realms/myrealm"
  client_id: "elastauth"
  client_secret: "${KEYCLOAK_SECRET}"
  scopes: ["openid", "profile", "email", "roles"]
  claim_mappings:
    username: "preferred_username"
    email: "email"
    groups: "realm_access.roles"
    full_name: "name"
  token_validation: "jwks"

# Cache configuration
cache:
  type: "redis"
  expiration: "1h"
  redis_host: "redis:6379"

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
  keycloak:
    image: quay.io/keycloak/keycloak:latest
    environment:
      - KEYCLOAK_ADMIN=admin
      - KEYCLOAK_ADMIN_PASSWORD=${KEYCLOAK_ADMIN_PASSWORD}
    ports:
      - "8080:8080"
    command: start-dev
    
  elastauth:
    image: elastauth:latest
    environment:
      - KEYCLOAK_SECRET=${KEYCLOAK_SECRET}
      - ELASTICSEARCH_PASSWORD=${ELASTICSEARCH_PASSWORD}
      - SECRET_KEY=${SECRET_KEY}
    volumes:
      - ./elastauth-config.yml:/config.yml
    ports:
      - "5000:5000"
    depends_on:
      - keycloak
      - elasticsearch
      - redis
```

## Troubleshooting

### Token Validation Issues

1. **Check Token Format**: Ensure JWT token is properly formatted
2. **Verify Issuer**: Confirm issuer URL matches provider configuration
3. **Check Clock Skew**: Ensure system clocks are synchronized
4. **Validate Scopes**: Confirm requested scopes are available

```bash
# Test token validation
curl -H "Authorization: Bearer YOUR_TOKEN" \
     -v http://elastauth:5000/
```

### Claim Mapping Problems

1. **Inspect Token Claims**: Decode JWT to verify claim structure
2. **Check Nested Claims**: Use correct dot notation for nested claims
3. **Verify Claim Names**: Ensure claim names match provider configuration

```bash
# Decode JWT token (header.payload.signature)
echo "YOUR_JWT_PAYLOAD" | base64 -d | jq .
```

### Provider Connectivity

1. **Test Discovery**: Verify OIDC discovery endpoint
2. **Check JWKS**: Ensure JWKS endpoint is accessible
3. **Network Connectivity**: Verify network access to provider

```bash
# Test OIDC discovery
curl https://your-provider.com/.well-known/openid-configuration

# Test JWKS endpoint
curl https://your-provider.com/.well-known/jwks.json
```

## Security Considerations

- **Token Validation**: Always validate JWT signatures
- **Claim Verification**: Verify required claims are present
- **Scope Limitation**: Request minimal required scopes
- **HTTPS Only**: Use HTTPS for all communications
- **Token Expiry**: Respect token expiration times

## Next Steps

- [Authelia Provider](authelia.md) - Header-based authentication
- [Configuration Examples](../examples/) - Complete configuration examples
- [Troubleshooting](../troubleshooting.md) - Common issues and solutions