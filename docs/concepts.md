# elastauth Concepts

## Overview

elastauth is a stateless authentication proxy that bridges authentication providers with Elasticsearch and Kibana. It acts as a middleware layer that:

1. **Receives authentication requests** from clients with various authentication formats
2. **Extracts user information** using pluggable authentication providers
3. **Generates temporary Elasticsearch credentials** for the authenticated user
4. **Returns authorization headers** that clients can use to access Elasticsearch/Kibana

## Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Client        │    │   elastauth     │    │  Elasticsearch  │
│   (Browser/App) │    │   Proxy         │    │   Cluster       │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         │ 1. Auth Request       │                       │
         │ (Headers/JWT/etc)     │                       │
         ├──────────────────────►│                       │
         │                       │ 2. Create/Update User │
         │                       ├──────────────────────►│
         │                       │                       │
         │ 3. Authorization      │ 4. User Created       │
         │    Header             │◄──────────────────────┤
         │◄──────────────────────┤                       │
         │                       │                       │
         │ 5. Access Elasticsearch/Kibana               │
         │ (using Authorization header)                  │
         ├───────────────────────────────────────────────►│
```

## Core Components

### Authentication Providers

Authentication providers are pluggable components that extract user information from different authentication systems:

- **Authelia Provider**: Extracts user info from HTTP headers (Remote-User, Remote-Groups, etc.)
- **OAuth2/OIDC Provider**: Validates JWT tokens and extracts claims from any OAuth2/OIDC-compliant system

### User Information Standardization

All providers return a standardized `UserInfo` structure:

```json
{
  "username": "john.doe",
  "email": "john.doe@example.com", 
  "groups": ["admin", "developers"],
  "full_name": "John Doe"
}
```

### Credential Management

elastauth generates temporary Elasticsearch credentials for each user:

1. **Password Generation**: Creates a secure random password
2. **User Creation**: Creates/updates the user in Elasticsearch with appropriate roles
3. **Credential Caching**: Encrypts and caches credentials to avoid repeated Elasticsearch calls
4. **Authorization Header**: Returns a Basic Auth header for Elasticsearch access

### Caching Layer

elastauth uses a pluggable caching system powered by the [cachego](https://github.com/wasilak/cachego) library:

- **Memory Cache**: In-memory caching for single-instance deployments
- **Redis Cache**: Distributed caching for multi-instance deployments
- **File Cache**: File-based caching for persistent storage
- **No Cache**: Direct Elasticsearch calls on every request

## Key Principles

### Stateless Operation

elastauth maintains no persistent authentication state:

- Authentication decisions are made on each request
- User credentials are temporarily cached but not stored permanently
- Multiple instances can run independently with shared cache

### Security

- **Credential Encryption**: All cached credentials are encrypted using AES
- **Input Validation**: All user input is validated and sanitized
- **Secure Defaults**: Secure configuration defaults with explicit overrides

### Pluggable Architecture

- **Provider Interface**: Common interface for all authentication systems
- **Factory Pattern**: Dynamic provider instantiation based on configuration
- **Single Provider**: Exactly one provider active at runtime for simplicity

## Authentication Flow

### 1. Request Processing

1. Client sends request with authentication information
2. elastauth routes request to configured authentication provider
3. Provider extracts and validates user information

### 2. User Management

1. elastauth checks cache for existing credentials
2. If cache miss, generates new temporary password
3. Creates/updates user in Elasticsearch with:
   - Generated password
   - User metadata (email, full name)
   - Mapped roles based on groups
4. Encrypts and caches credentials

### 3. Response Generation

1. Decrypts cached credentials
2. Generates Basic Auth header
3. Returns JSON response with authorization header

## Integration Points

### Elasticsearch Integration

elastauth integrates with Elasticsearch through:

- **User Management API**: Creates and updates users
- **Role Mapping**: Maps user groups to Elasticsearch roles
- **Security Settings**: Configures user permissions and access

### Kibana Integration

Kibana automatically works with elastauth-managed users:

- Uses the same Elasticsearch security model
- Inherits role mappings and permissions
- No additional configuration required

### External Authentication Systems

elastauth integrates with external authentication through:

- **Header-based Systems**: Authelia, Traefik Forward Auth
- **OAuth2/OIDC Systems**: Keycloak, Casdoor, Authentik, Auth0, Azure AD
- **Custom Providers**: Extensible through the provider interface

## Configuration Philosophy

### Single Provider Selection

elastauth uses exactly one authentication provider at runtime:

- Configured via `auth_provider` setting
- Prevents configuration complexity and conflicts
- Enables clear authentication flow

### Environment Variable Support

All configuration can be overridden via environment variables:

- Supports containerized deployments
- Enables secret management through external systems
- Follows 12-factor app principles

### Validation and Defaults

- Configuration is validated at startup
- Clear error messages for invalid configurations
- Secure defaults with explicit overrides

## Deployment Patterns

### Single Instance

- Memory or file cache
- Simple configuration
- Suitable for small deployments

### Multi-Instance

- Redis cache required
- Shared configuration and encryption keys
- Horizontal scaling support
- Load balancer compatible

### Kubernetes

- ConfigMap and Secret support
- Health check endpoints
- Graceful shutdown handling
- Container-optimized logging

## Third-Party Components

elastauth integrates with several external systems:

### Elasticsearch

- **Purpose**: User and role management, search and analytics platform
- **Integration**: REST API for user management
- **Documentation**: [Elasticsearch Security](https://www.elastic.co/guide/en/elasticsearch/reference/current/security-api.html)

### Kibana

- **Purpose**: Data visualization and management interface
- **Integration**: Automatic through Elasticsearch security
- **Documentation**: [Kibana Security](https://www.elastic.co/guide/en/kibana/current/security.html)

### Authentication Systems

- **Authelia**: [Documentation](https://www.authelia.com/)
- **Keycloak**: [Documentation](https://www.keycloak.org/documentation)
- **Casdoor**: [Documentation](https://casdoor.org/docs/)
- **Authentik**: [Documentation](https://docs.goauthentik.io/)

### Caching

- **Redis**: [Documentation](https://redis.io/documentation)
- **cachego**: [Documentation](https://github.com/wasilak/cachego)

## Next Steps

- [Authentication Providers](providers/README.md) - Detailed provider configuration
- [Cache Providers](cache/README.md) - Caching system configuration  
- [Configuration Guide](configuration.md) - Complete configuration reference
- [Deployment Guide](deployment.md) - Production deployment patterns
- [Troubleshooting](troubleshooting.md) - Common issues and solutions