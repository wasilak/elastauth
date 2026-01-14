# elastauth Demo Environment

A comprehensive, self-contained Docker Compose demo environment showcasing elastauth with Authelia authentication, Elasticsearch backend, and Redis caching.

## Overview

This demo environment provides a complete authentication proxy setup that demonstrates:

- **Authelia**: File-based authentication server providing user authentication
- **elastauth**: Authentication proxy that validates headers and proxies requests to Elasticsearch
- **Elasticsearch 9.x**: Secure backend with TLS enabled and X-Pack security
- **Redis**: Session storage and caching layer

The environment is designed for local testing and demonstration purposes, with automated certificate generation, configuration management, and convenient Makefile targets for easy operation.

## Architecture

```
┌──────────┐
│  Client  │
└────┬─────┘
     │ 1. HTTP Request
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
└────┬────────────┘
     │ 3. Proxied Request + ES Auth
     ▼
┌─────────────────┐
│ Elasticsearch   │ :9200
│  (Backend)      │
│                 │
│ - Processes     │
│   request       │
│ - Returns data  │
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

The demo environment demonstrates a complete authentication flow:

1. **Client Request**: Client sends HTTP request to Authelia
2. **Authentication**: Authelia validates credentials and adds authentication headers:
   - `Remote-User`: Username
   - `Remote-Groups`: User groups
   - `Remote-Email`: User email
   - `Remote-Name`: User display name
3. **Proxy**: Request is forwarded to elastauth with authentication headers
4. **Validation**: elastauth validates headers and maps to Elasticsearch user
5. **Backend**: Request is proxied to Elasticsearch with proper authentication
6. **Response**: Elasticsearch processes request and returns data

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

### Query Elasticsearch Through elastauth

```bash
curl -H "Remote-User: admin" \
     -H "Remote-Groups: admins" \
     http://localhost:3000/_cluster/health
```

### Search Elasticsearch Indices

```bash
curl -H "Remote-User: admin" \
     -H "Remote-Groups: admins" \
     http://localhost:3000/_search
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

3. Restart Authelia: `docker-compose restart authelia`

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

Restart Authelia: `docker-compose restart authelia`

### Modifying Elasticsearch Configuration

Edit `docker-compose.yml` under the `elasticsearch` service to add or modify environment variables:

```yaml
environment:
  - node.name=es-demo-node
  - cluster.name=elastauth-demo
  # Add your custom settings here
```

Restart Elasticsearch: `docker-compose restart elasticsearch`

### Changing Service Ports

Edit `docker-compose.yml` to modify port mappings:

```yaml
ports:
  - "9200:9200"  # Change to "9201:9200" to use port 9201 on host
```

Restart services: `make restart`

### Customizing elastauth Configuration

Edit `configs/elastauth/config.yml` to modify:

- Server port and host
- Authentication provider settings
- Elasticsearch connection details
- Cache configuration
- Logging settings

Restart elastauth: `docker-compose restart elastauth`

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
2. Restart affected service: `docker-compose restart <service>`
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

## Additional Resources

- [elastauth Documentation](../../README.md)
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
