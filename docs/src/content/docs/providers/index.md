---
title: Authentication Providers
description: Configure authentication providers for elastauth
---

elastauth supports multiple authentication providers through a pluggable architecture. Each provider implements a common interface while supporting different authentication mechanisms.

## Available Providers

- **[Authelia](/elastauth/providers/authelia)** - Header-based authentication for Authelia deployments
- **[OAuth2/OIDC](/elastauth/providers/oidc)** - Generic OAuth2/OIDC provider supporting multiple systems

## Provider Selection

Configure exactly one authentication provider using the `auth_provider` setting:

```yaml
# Choose one provider
auth_provider: "authelia"  # or "oidc"
```

## Common Configuration Patterns

### Environment Variable Overrides

All provider configurations support environment variable overrides:

```yaml
oidc:
  client_secret: "${OIDC_CLIENT_SECRET}"
  
# Can be overridden with:
# OIDC_CLIENT_SECRET=your-secret-here
```

### Validation

All providers validate their configuration at startup:

- Missing required fields result in startup errors
- Invalid URLs or formats are caught early
- Clear error messages guide configuration fixes

## Provider Interface

All providers implement the same interface:

```go
type AuthProvider interface {
    GetUser(ctx context.Context, req *AuthRequest) (*UserInfo, error)
    Type() string
    Validate() error
}
```

### UserInfo Structure

All providers return standardized user information:

```json
{
  "username": "john.doe",
  "email": "john.doe@example.com",
  "groups": ["admin", "developers"], 
  "full_name": "John Doe"
}
```

## Adding New Providers

To add a new authentication provider:

1. Implement the `AuthProvider` interface
2. Register the provider in the factory
3. Add configuration validation
4. Update documentation

## Next Steps

- [Authelia Provider](/elastauth/providers/authelia) - Header-based authentication
- [OAuth2/OIDC Provider](/elastauth/providers/oidc) - JWT token authentication