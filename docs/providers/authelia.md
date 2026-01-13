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

- [OAuth2/OIDC Provider](oidc.md) - Alternative authentication method
- [Configuration Examples](../examples/) - Complete configuration examples
- [Troubleshooting](../troubleshooting.md) - Common issues and solutions