---
title: Authelia Provider
description: Configure header-based authentication with Authelia
---

The Authelia provider extracts user information from HTTP headers set by Authelia forward authentication. This provider is ideal when you already have Authelia handling authentication and want elastauth to create Elasticsearch users based on the authenticated user information.

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

If not specified, elastauth uses these default header names:

- `Remote-User` - Username (required)
- `Remote-Groups` - Comma-separated groups (optional)
- `Remote-Email` - User email address (optional)
- `Remote-Name` - User's full name (optional)

### Environment Variable Overrides

```bash
AUTHELIA_HEADER_USERNAME="X-Remote-User"
AUTHELIA_HEADER_GROUPS="X-Remote-Groups"
AUTHELIA_HEADER_EMAIL="X-Remote-Email"
AUTHELIA_HEADER_NAME="X-Remote-Name"
```

## Header Format

### Username Header
Must contain a single username:
```
Remote-User: john.doe
```

### Groups Header
Groups can be comma-separated or single:
```
Remote-Groups: admin,developers,users
Remote-Groups: admin
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

## Complete Configuration Example

```yaml
auth_provider: "authelia"

authelia:
  header_username: "Remote-User"
  header_groups: "Remote-Groups"
  header_email: "Remote-Email"
  header_name: "Remote-Name"

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
  users:
    - "kibana_user"
```

## Validation

The Authelia provider validates:

- **Username**: Required, must be non-empty
- **Groups**: Optional, validated if present
- **Email**: Optional, must be valid email format if present
- **Name**: Optional, validated if present

## Error Handling

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

## Troubleshooting

### Headers Not Received

**Symptoms**: elastauth returns 400 errors about missing headers

**Solutions**:
1. Verify Authelia is setting the headers correctly
2. Check reverse proxy configuration forwards headers
3. Verify header names match elastauth configuration
4. Test with curl: `curl -H "Remote-User: testuser" http://elastauth:5000/`

### Authentication Failures

**Symptoms**: Users can't access Kibana despite successful Authelia authentication

**Solutions**:
1. Check elastauth logs for specific error messages
2. Verify Elasticsearch connectivity and credentials
3. Confirm user roles are properly mapped
4. Test elastauth endpoint directly

## Integration Notes

### Reverse Proxy Setup
Ensure your reverse proxy (Traefik, Nginx, etc.) forwards the Authelia headers to elastauth.

### Security Considerations
- Headers should only be set by trusted Authelia instances
- Use network-level security to prevent header spoofing
- Consider using custom header names for additional security

## Related Documentation

- **[Cache Configuration](/elastauth/cache/)** - Optimize performance with caching
  - [Redis Cache](/elastauth/cache/redis) - Distributed caching for production
- **[Troubleshooting](/elastauth/guides/troubleshooting#authelia-issues)** - Common Authelia issues