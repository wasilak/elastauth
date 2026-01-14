# elastauth Demo Environment

A comprehensive, self-contained Docker Compose demo environment showcasing elastauth with Authelia authentication, Elasticsearch backend, and Redis caching.

> 📚 **Full Documentation**: See [Starlight documentation](../../docs/) for comprehensive guides

## Overview

This demo environment demonstrates elastauth in **authentication-only mode** (default), where elastauth validates authentication and returns headers for use with reverse proxies like Traefik.

**What's included:**

- **Authelia**: File-based authentication server providing user authentication
- **elastauth**: Authentication proxy that validates headers and manages Elasticsearch users
- **Elasticsearch 9.2.4**: Secure backend with TLS enabled and X-Pack security (fully open source under AGPL)
- **Redis**: Session storage and caching layer

## Operating Modes

elastauth supports two operating modes. This demo uses **authentication-only mode** by default.

### Authentication-Only Mode (Default)

elastauth validates authentication and returns headers. A reverse proxy (like Traefik) handles proxying to Elasticsearch.

```
Client → Traefik → elastauth (auth) → Traefik → Elasticsearch
```

**When to use:**
- You already use Traefik or another reverse proxy
- You need advanced routing or middleware
- You protect multiple services

**Example:** See `deployment/example/traefik-auth-only/`

### Transparent Proxy Mode

elastauth handles both authentication and proxying directly to Elasticsearch.

```
Client → elastauth (auth + proxy) → Elasticsearch
```

**When to use:**
- You don't need a reverse proxy
- You want simpler architecture
- You're deploying a single service

**Example:** See `deployment/example/direct-proxy/`

### Decision Guide

**Choose Authentication-Only Mode if:**
- ✅ You have existing Traefik infrastructure
- ✅ You need to protect multiple services
- ✅ You require advanced routing or middleware
- ✅ You want centralized TLS termination

**Choose Transparent Proxy Mode if:**
- ✅ You're deploying elastauth standalone
- ✅ You only need to protect Elasticsearch
- ✅ You want the simplest possible setup
- ✅ You prefer fewer components

> 📚 **Detailed Comparison**: See [Operating Modes Documentation](../../docs/src/content/docs/deployment/modes.md)

## Architecture

### Current Setup (Authentication-Only Mode)

```
┌──────────┐
│  Client  │
└────┬─────┘
     │ 1. HTTP Request with Auth Headers
     ▼
┌─────────────────┐
│   Authelia      │ :9091
│  (Auth Server)  │
│                 │
│ - Validates     │
│   credentials   │
│ - Sets headers  │
└────┬────────────┘
     │ 2. Request + Auth Headers
     │    (Remote-User, Remote-Groups, etc.)
     ▼
┌─────────────────┐
│   elastauth     │ :3000
│  (Auth Proxy)   │
│                 │
│ - Validates     │
│   headers       │
│ - Maps to ES    │
│   user          │
│ - Returns       │
│   Authorization │
└─────────────────┘

     ┌──────────┐
     │  Redis   │ :6379
     │ (Cache)  │
     └──────────┘
          ▲
          │ Session Storage
          │
     ┌────┴────┐
     │         │
  Authelia  elastauth

┌─────────────────┐
│ Elasticsearch   │ :9200
│  (Backend)      │
│                 │
│ - Direct access │
│   for testing   │
└─────────────────┘
```

### Alternative: With Traefik (Forward Auth)

For production deployments with Traefik, see `deployment/example/traefik-auth-only/`:

```
Client → Traefik → elastauth (auth) → Traefik → Elasticsearch
```

### Alternative: Transparent Proxy Mode

For direct proxy mode without Traefik, see `deployment/example/direct-proxy/`:

```
Client → elastauth (auth + proxy) → Elasticsearch
```

## Quick Start

### Prerequisites

- Docker (20.10 or later)
- Docker Compose (2.0 or later)
- OpenSSL (for certificate generation)
- Make (for automation targets)
- curl (for testing)

### 1. Initialize the Environment

Generate certificates and configuration files:

```bash
make init
```

This will:
- Create a self-signed Certificate Authority (CA)
- Generate TLS certificates for Elasticsearch and Authelia
- Create Authelia configuration and user database
- Create elastauth configuration

### 2. Start Services

Start all services in detached mode:

```bash
make up
```

Services will start in the background. Wait 30-60 seconds for all services to become healthy.

### 3. Verify Services

Run health checks on all services:

```bash
make test
```

### 4. View Connection Information

Display service URLs, credentials, and example commands:

```bash
make info
```

## Service Ports and URLs

| Service | URL | Port |
|---------|-----|------|
| Elasticsearch | https://localhost:9200 | 9200 |
| Authelia | http://localhost:9091 | 9091 |
| elastauth | http://localhost:3000 | 3000 |
| Redis | localhost:6379 | 6379 |

## Default Credentials

⚠️ **WARNING**: These are demo credentials only. Change for production use!

### Elasticsearch
- **Username**: `elastic`
- **Password**: `demo-password-change-me`

### Authelia
- **Username**: `admin`
- **Password**: `demo-password`
- **Email**: `admin@example.com`
- **Groups**: `admins`

## Makefile Targets

### Initialization

- `make help` - Display all available targets with descriptions (default)
- `make init` - Bootstrap the demo environment (generate certs and configs)
- `make init-force` - Force regenerate all certificates and configs

### Service Management

- `make up` - Start all services in detached mode
- `make down` - Stop and remove all containers (preserves volumes)
- `make restart` - Restart all services (down + up)
- `make ps` - Show running containers
- `make status` - Display formatted service status (alias for ps)
- `make logs` - Display logs from all services with timestamps
- `make logs-follow` - Follow logs in real-time
- `make logs SERVICE=<name>` - View logs for specific service

### Testing and Information

- `make test` - Run health checks on all services
- `make info` - Display connection information, credentials, and examples

### Cleanup

- `make clean` - Remove containers, volumes, certs, and configs (with confirmation)
- `make clean-all` - Force clean without confirmation

## Authentication Flow

The demo environment demonstrates authentication-only mode:

1. **Client Request**: Client sends HTTP request with authentication headers to elastauth
2. **Authentication**: Authelia validates credentials and adds authentication headers:
   - `Remote-User`: Username
   - `Remote-Groups`: User groups
   - `Remote-Email`: User email
   - `Remote-Name`: User display name
3. **Validation**: elastauth validates headers and maps to Elasticsearch user
4. **User Management**: elastauth creates or updates Elasticsearch user with appropriate roles
5. **Response**: elastauth returns success with user information

**Note**: In this demo, elastauth does NOT proxy requests to Elasticsearch. For proxying, see:
- **With Traefik**: `deployment/example/traefik-auth-only/` - Forward auth pattern
- **Direct Proxy**: `deployment/example/direct-proxy/` - Transparent proxy mode

## Testing Examples

### Test Elasticsearch Directly

```bash
curl -k -u elastic:demo-password-change-me https://localhost:9200/_cluster/health
```

### Test Authelia Health

```bash
curl http://localhost:9091/api/health
```

### Test elastauth Health

```bash
curl http://localhost:3000/health
```

### Test Authentication Flow Through elastauth

```bash
curl -H "Remote-User: admin" \
     -H "Remote-Groups: admins" \
     -H "Remote-Email: admin@example.com" \
     -H "Remote-Name: Admin User" \
     http://localhost:3000/
```

**Note**: This demo uses authentication-only mode. elastauth validates authentication but does not proxy to Elasticsearch. To test proxying:
- **With Traefik**: See `deployment/example/traefik-auth-only/`
- **Direct Proxy**: See `deployment/example/direct-proxy/`

### Access Elasticsearch Directly

In this demo, you can access Elasticsearch directly for testing:

```bash
# Direct access with elastic credentials
curl -k -u elastic:demo-password-change-me \
     https://localhost:9200/_cluster/health
```

## Troubleshooting

### Services Not Starting

**Problem**: Services fail to start or become unhealthy

**Solutions**:
1. Check service status: `make ps`
2. View service logs: `make logs` or `make logs SERVICE=<name>`
3. Ensure Docker has enough resources (at least 4GB RAM for Elasticsearch)
4. Wait longer for services to initialize (Elasticsearch can take 60+ seconds)
5. Restart services: `make restart`

### Certificate Errors

**Problem**: TLS certificate errors or expired certificates

**Solutions**:
1. Regenerate certificates: `make init-force`
2. Verify certificates exist: `ls -la certs/`
3. Check certificate validity: `openssl x509 -in certs/elasticsearch.crt -noout -dates`

### Port Conflicts

**Problem**: Port already in use errors

**Solutions**:
1. Check what's using the port: `lsof -i :9200` (or :9091, :3000, :6379)
2. Stop conflicting services
3. Modify port mappings in `docker-compose.yml` if needed

### Elasticsearch Memory Issues

**Problem**: Elasticsearch fails with memory errors

**Solutions**:
1. Increase Docker memory limit (Docker Desktop → Settings → Resources)
2. Reduce Elasticsearch heap size in `docker-compose.yml` (ES_JAVA_OPTS)
3. Ensure at least 4GB RAM available for Docker

### Authentication Failures

**Problem**: Authentication not working through elastauth

**Solutions**:
1. Verify Authelia is healthy: `curl http://localhost:9091/api/health`
2. Check elastauth logs: `make logs SERVICE=elastauth`
3. Verify authentication headers are being sent correctly
4. Ensure Redis is running: `docker exec redis-demo redis-cli ping`
5. Check Elasticsearch TLS configuration (see TLS Certificate Issues below)

### TLS Certificate Issues

**Problem**: Elasticsearch TLS certificate verification failures

**Current Limitation**: The elastauth application currently does not fully implement the `skip_verify` TLS configuration option. This causes certificate verification failures with self-signed certificates in the demo environment.

**Symptoms**:
- Error in logs: `tls: failed to verify certificate: x509: certificate signed by unknown authority`
- 500 Internal Server Error when accessing elastauth endpoints

**Workaround Options**:

1. **Use HTTP instead of HTTPS for Elasticsearch** (Demo only - not for production):
   - Edit `configs/elastauth/config.yml`:
     ```yaml
     elasticsearch_host: "http://elasticsearch:9200"  # Change https to http
     ```
   - Edit `docker-compose.yml` Elasticsearch service:
     ```yaml
     # Comment out or remove xpack.security.http.ssl settings
     # - xpack.security.http.ssl.enabled=true
     # - xpack.security.http.ssl.key=...
     # - xpack.security.http.ssl.certificate=...
     ```
   - Restart services: `make restart`

2. **Wait for code fix**: The elastauth application needs to be updated to properly handle the `elasticsearch.tls.skip_verify` configuration option by creating an HTTP client with custom TLS configuration.

**Note**: This is a known limitation of the current elastauth implementation and is tracked for future enhancement.

### Elasticsearch Proxy Routes Not Found (404)

**Problem**: This demo uses authentication-only mode, not proxy mode

**Explanation**: The default demo configuration runs elastauth in authentication-only mode. In this mode, elastauth validates authentication and returns headers, but does NOT proxy requests to Elasticsearch.

**Solutions**:

1. **For Traefik Forward Auth**: See `deployment/example/traefik-auth-only/`
   - Complete Traefik integration with forward auth middleware
   - elastauth validates auth, Traefik proxies to Elasticsearch
   - Documentation: `docs/src/content/docs/deployment/auth-only-mode.mdx`

2. **For Direct Proxy Mode**: See `deployment/example/direct-proxy/`
   - elastauth handles both auth and proxying
   - Simpler architecture, fewer components
   - Documentation: `docs/src/content/docs/deployment/proxy-mode.mdx`

3. **Enable Proxy Mode in This Demo** (not recommended):
   - Uncomment proxy environment variables in `docker-compose.yml`
   - See comments in the elastauth service section
   - Better to use the dedicated `direct-proxy` example instead

### Configuration Issues

**Problem**: Services fail due to configuration errors

**Solutions**:
1. Regenerate configurations: `make init-force`
2. Check configuration files in `configs/authelia/` and `configs/elastauth/`
3. Verify YAML syntax is valid
4. Review service logs for specific configuration errors

### Complete Reset

If all else fails, perform a complete cleanup and restart:

```bash
make clean-all
make init
make up
make test
```

## Customizing Configuration

### Changing Passwords

#### Elasticsearch Password

1. Edit `docker-compose.yml`:
   ```yaml
   - ELASTIC_PASSWORD=your-new-password
   ```

2. Update `configs/elastauth/config.yml`:
   ```yaml
   elasticsearch:
     password: "your-new-password"
   ```

3. Update health check in `docker-compose.yml`:
   ```yaml
   test: ["CMD-SHELL", "curl -k -u elastic:your-new-password https://localhost:9200/_cluster/health || exit 1"]
   ```

4. Restart services: `make restart`

#### Authelia Password

1. Generate new password hash:
   ```bash
   docker run --rm authelia/authelia:latest authelia crypto hash generate argon2 --password 'your-new-password'
   ```

2. Update `configs/authelia/users_database.yml` with the new hash

3. Restart Authelia: `docker compose restart authelia`

### Adding Authelia Users

Edit `configs/authelia/users_database.yml`:

```yaml
users:
  admin:
    displayname: "Admin User"
    password: "$argon2id$..."
    email: "admin@example.com"
    groups:
      - admins
  
  newuser:
    displayname: "New User"
    password: "$argon2id$..."  # Generate with authelia crypto hash
    email: "newuser@example.com"
    groups:
      - users
```

Restart Authelia: `docker compose restart authelia`

### Modifying Elasticsearch Configuration

Edit `docker-compose.yml` under the `elasticsearch` service to add or modify environment variables:

```yaml
environment:
  - node.name=es-demo-node
  - cluster.name=elastauth-demo
  # Add your custom settings here
```

Restart Elasticsearch: `docker compose restart elasticsearch`

### Changing Service Ports

Edit `docker-compose.yml` to modify port mappings:

```yaml
ports:
  - "9200:9200"  # Change to "9201:9200" to use port 9201 on host
```

Restart services: `make restart`

### Customizing elastauth Configuration

Edit `configs/elastauth/config.yml` to modify:

**Server Configuration:**
```yaml
listen: "0.0.0.0:3000"  # Server listen address
```

**Authentication Provider:**
```yaml
auth_provider: "authelia"  # Options: authelia, casdoor, oidc

authelia:
  user_header: "Remote-User"
  groups_header: "Remote-Groups"
  email_header: "Remote-Email"
  name_header: "Remote-Name"
```

**Elasticsearch Configuration (Legacy Format):**
```yaml
elasticsearch_host: "https://elasticsearch:9200"
elasticsearch_username: "elastic"
elasticsearch_password: "demo-password-change-me"

elasticsearch:
  tls:
    enabled: true
    ca_cert: "/certs/ca.crt"
    skip_verify: true  # Set to false with proper certificates
```

**Redis Cache Configuration:**
```yaml
cache:
  type: "redis"
  redis_host: "redis:6379"  # Note: redis_host not redis.address
  redis_db: 0
```

**Logging Configuration:**
```yaml
logging:
  level: "info"  # Options: debug, info, warn, error
  format: "json"  # Options: json, text
```

After making changes, restart elastauth: `docker compose restart elastauth`

## Directory Structure

```
deployment/example/
├── Makefile                    # Automation targets
├── docker-compose.yml          # Service definitions
├── README.md                   # This file
├── .env.example                # Environment variables template
├── .gitignore                  # Git ignore patterns
├── certs/                      # Generated certificates (gitignored)
│   ├── ca.crt                  # Certificate Authority certificate
│   ├── ca.key                  # CA private key
│   ├── ca.srl                  # CA serial number
│   ├── elasticsearch.crt       # Elasticsearch certificate
│   ├── elasticsearch.key       # Elasticsearch private key
│   ├── authelia.crt            # Authelia certificate
│   └── authelia.key            # Authelia private key
└── configs/                    # Generated configs (gitignored)
    ├── authelia/
    │   ├── configuration.yml   # Authelia main config
    │   ├── users_database.yml  # User credentials
    │   ├── db.sqlite3          # Authelia database (created at runtime)
    │   └── notification.txt    # Notification log (created at runtime)
    └── elastauth/
        └── config.yml          # elastauth configuration
```

## Security Considerations

### Demo Environment Only

This environment is designed for **demonstration and testing purposes only**. It includes:

- Self-signed certificates (not trusted by browsers)
- Default passwords (publicly documented)
- Simplified security configurations
- Local-only networking

### Production Deployment

For production use, you must:

1. **Use proper certificates**: Obtain certificates from a trusted CA (Let's Encrypt, etc.)
2. **Change all passwords**: Use strong, unique passwords for all services
3. **Enable proper authentication**: Configure Authelia with LDAP, OIDC, or other production auth backends
4. **Secure networking**: Use proper network segmentation and firewalls
5. **Enable monitoring**: Add logging, metrics, and alerting
6. **Regular updates**: Keep all services updated with security patches
7. **Backup data**: Implement proper backup and disaster recovery procedures
8. **Review configurations**: Audit all configuration files for security best practices

### Password Security

The demo uses these default passwords:

- Elasticsearch: `demo-password-change-me`
- Authelia admin: `demo-password`

**These passwords are publicly documented and must be changed for any non-demo use.**

## Advanced Usage

### Viewing Logs for Specific Service

```bash
make logs SERVICE=elasticsearch
make logs SERVICE=authelia
make logs SERVICE=elastauth
make logs SERVICE=redis
```

### Following Logs in Real-Time

```bash
make logs-follow
make logs-follow SERVICE=elastauth
```

### Forcing Certificate Regeneration

```bash
make init-force
# or
make FORCE=1 init
```

### Complete Cleanup and Restart

```bash
make clean-all && make init && make up
```

### Checking Service Health Manually

```bash
# Elasticsearch
curl -k -u elastic:demo-password-change-me https://localhost:9200/_cluster/health

# Authelia
curl http://localhost:9091/api/health

# Redis
docker exec redis-demo redis-cli ping

# elastauth
curl http://localhost:3000/health
```

### Accessing Service Containers

```bash
# Elasticsearch
docker exec -it elasticsearch-demo bash

# Authelia
docker exec -it authelia-demo sh

# Redis
docker exec -it redis-demo sh

# elastauth
docker exec -it elastauth-demo sh
```

## Development and Testing

### Building elastauth from Source

The demo environment builds elastauth from the project root:

```yaml
elastauth:
  build:
    context: ../..
    dockerfile: Dockerfile
```

After making code changes:

```bash
make down
make up  # Rebuilds elastauth image
```

### Testing Configuration Changes

1. Modify configuration files in `configs/`
2. Restart affected service: `docker compose restart <service>`
3. Check logs: `make logs SERVICE=<service>`
4. Verify with health checks: `make test`

### Debugging

Enable debug logging by modifying configuration files:

**Authelia** (`configs/authelia/configuration.yml`):
```yaml
log:
  level: "debug"
```

**elastauth** (`configs/elastauth/config.yml`):
```yaml
logging:
  level: "debug"
```

Restart services and view logs: `make restart && make logs-follow`

## Deployment Scenarios

This demo shows the basic authentication-only mode. For production deployments, see these examples:

### 1. Traefik Forward Auth (Recommended for Production)

**Location**: `deployment/example/traefik-auth-only/`

**Architecture**: Client → Traefik → elastauth (auth) → Traefik → Elasticsearch

**Use when:**
- You already use Traefik
- You need advanced routing/middleware
- You protect multiple services
- You want centralized TLS termination

**Documentation**: `docs/src/content/docs/deployment/auth-only-mode.mdx`

### 2. Direct Proxy Mode (Simpler Architecture)

**Location**: `deployment/example/direct-proxy/`

**Architecture**: Client → elastauth (auth + proxy) → Elasticsearch

**Use when:**
- You don't need a reverse proxy
- You want simpler architecture
- You're deploying a single service
- You prefer fewer components

**Documentation**: `docs/src/content/docs/deployment/proxy-mode.mdx`

### 3. Mode Comparison

**Location**: `docs/src/content/docs/deployment/modes.md`

Complete comparison of both modes with decision tree and recommendations.

## Additional Resources

- [elastauth Documentation](../../README.md)
- [Operating Modes Comparison](../../docs/src/content/docs/deployment/modes.md)
- [Traefik Integration Guide](../../docs/src/content/docs/deployment/auth-only-mode.mdx)
- [Direct Proxy Mode Guide](../../docs/src/content/docs/deployment/proxy-mode.mdx)
- [Authelia Documentation](https://www.authelia.com/)
- [Elasticsearch Documentation](https://www.elastic.co/guide/en/elasticsearch/reference/current/index.html)
- [Docker Compose Documentation](https://docs.docker.com/compose/)

## Support

For issues or questions:

1. Check the [Troubleshooting](#troubleshooting) section
2. Review service logs: `make logs`
3. Run health checks: `make test`
4. Check the main elastauth repository for issues and documentation

## License

This demo environment is part of the elastauth project. See the main project LICENSE file for details.
