# Direct Proxy Mode Example

This example demonstrates elastauth in **transparent proxy mode**, where elastauth acts as a complete authentication proxy to Elasticsearch.

> 📚 **Full Documentation**: See the [Proxy Mode Guide](https://wasilak.github.io/elastauth/deployment/proxy-mode/) for complete architecture details, configuration options, and troubleshooting.

## Architecture

In this mode, elastauth handles both authentication and proxying:

```
Client → elastauth (auth + proxy) → Elasticsearch
         ↓
      Authelia (authentication)
```

**Key Difference from Traefik Mode**: No reverse proxy needed. elastauth is the proxy.

## Quick Start

### Prerequisites

- Docker Engine 20.10 or later
- Docker Compose V2 (integrated with Docker CLI)
- At least 2GB RAM available

### 1. Start the Stack

```bash
docker compose up -d
```

### 2. Wait for Services

```bash
# Check status
docker compose ps

# Wait for Elasticsearch (1-2 minutes)
docker compose logs -f elasticsearch
```

### 3. Test Authentication Flow

**Step 1: Get Authelia session**

```bash
# Login to Authelia
curl -X POST http://localhost:9091/api/firstfactor \
  -H "Content-Type: application/json" \
  -d '{"username":"john","password":"password","keepMeLoggedIn":false}'
```

**Step 2: Access Elasticsearch through elastauth**

```bash
# Make request with Authelia headers
curl -H "Remote-User: john" \
     -H "Remote-Groups: admins,dev" \
     -H "Remote-Email: john.doe@authelia.com" \
     -H "Remote-Name: John Doe" \
     http://localhost:8080/
```

**Expected**: Elasticsearch cluster information (authenticated as john)

### 4. Test Health Checks

```bash
# elastauth health
curl http://localhost:8080/elastauth/health

# elastauth readiness (includes Elasticsearch check)
curl http://localhost:8080/elastauth/ready

# elastauth configuration
curl http://localhost:8080/elastauth/config
```

## Configuration

### Test Users

Defined in `authelia/users_database.yml`:

- **john**: Groups: `admins`, `dev`
- **jane**: Groups: `dev`

Both use password: `password`

### Key Settings

In `docker-compose.yml` environment variables:

```yaml
# Proxy mode
ELASTAUTH_PROXY_ENABLED=true
ELASTAUTH_PROXY_ELASTICSEARCH_URL=http://elasticsearch:9200

# Default roles for all users
ELASTAUTH_DEFAULT_ROLES=kibana_user,monitoring_user

# Group to role mappings
ELASTAUTH_GROUP_MAPPINGS_ADMIN=superuser
ELASTAUTH_GROUP_MAPPINGS_DEV=kibana_admin,monitoring_user
```

**How it works:**
- Users in the `admin` group get the `superuser` role (full Elasticsearch access)
- Users in the `dev` group get `kibana_admin` and `monitoring_user` roles
- Users without groups get the default roles (`kibana_user`, `monitoring_user`)

**Available Elasticsearch Roles:**
- `superuser` - Full cluster access
- `kibana_admin` - Kibana administration
- `kibana_user` - Kibana read/write access
- `monitoring_user` - Monitoring data access
- `viewer` - Read-only access
- `editor` - Read/write access

## Request Flow

1. **Client** sends request to elastauth (port 8080)
2. **elastauth** extracts Authelia headers
3. **elastauth** authenticates user via Authelia headers
4. **elastauth** generates/retrieves Elasticsearch credentials
5. **elastauth** proxies request to Elasticsearch with credentials
6. **Elasticsearch** processes request
7. **elastauth** forwards response to client

## Comparison with Traefik Mode

| Aspect | Direct Proxy Mode | Traefik Mode |
|--------|------------------|--------------|
| **Architecture** | Client → elastauth → ES | Client → Traefik → elastauth → Traefik → ES |
| **Components** | elastauth, Authelia, ES | elastauth, Authelia, Traefik, ES |
| **Complexity** | Simpler (fewer components) | More complex (more components) |
| **Use Case** | Single backend (Elasticsearch) | Multiple backends, advanced routing |
| **Proxy Features** | Basic HTTP proxy | Advanced routing, load balancing, TLS |
| **Configuration** | `proxy.enabled=true` | `proxy.enabled=false` + Traefik config |

**When to use Direct Proxy Mode**:
- Simple deployments with only Elasticsearch
- Want fewer moving parts
- Don't need advanced reverse proxy features

**When to use Traefik Mode**:
- Multiple services behind same proxy
- Need advanced routing/load balancing
- Already using Traefik infrastructure

## Cleanup

```bash
# Stop services
docker compose down

# Remove volumes
docker compose down -v
```

## Known Issues

### Docker Image Needs Rebuild

The current Docker image (`ghcr.io/wasilak/elastauth:main`) was built before recent fixes were committed. This causes:

1. **Cache Configuration Bug**: Environment variable `ELASTAUTH_CACHE_TYPE` triggers validation error
2. **Status**: Fixed in source code (commit 4213c3c), waiting for new Docker image build

**Workaround Options**:

**Option 1: Build Locally** (Recommended)
```bash
# Build elastauth from source
go build -o elastauth

# Start just the dependencies
docker compose up -d elasticsearch redis

# Run elastauth locally
export ELASTAUTH_PROXY_ENABLED=true
export ELASTAUTH_PROXY_ELASTICSEARCH_URL=http://localhost:9200
export ELASTAUTH_AUTH_PROVIDER=authelia
export ELASTAUTH_ELASTICSEARCH_HOST=http://localhost:9200
export ELASTAUTH_ELASTICSEARCH_USERNAME=elastic
export ELASTAUTH_ELASTICSEARCH_PASSWORD=changeme
export ELASTAUTH_SECRET_KEY=0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef
export ELASTAUTH_CACHE_TYPE=redis
export ELASTAUTH_CACHE_REDIS_HOST=localhost:6379
export ELASTAUTH_LISTEN=0.0.0.0:8080

./elastauth
```

**Option 2: Wait for New Docker Image**

Once the Docker image is rebuilt with the latest source code, `docker compose up -d` will work as expected.

### Single-Command Example Status

✅ **Will work** once Docker image is rebuilt
✅ **Works now** if you build from source (Option 1 above)
❌ **Doesn't work** with current Docker image due to cache validation bug

The docker-compose.yml is correct and ready - it just needs the updated Docker image.

## Documentation

For detailed information, see:

- **[Proxy Mode Guide](https://wasilak.github.io/elastauth/deployment/proxy-mode/)** - Complete deployment guide
- **[Configuration Reference](https://wasilak.github.io/elastauth/configuration/)** - All configuration options
- **[Troubleshooting](https://wasilak.github.io/elastauth/guides/troubleshooting/)** - Common issues and solutions
- **[Mode Comparison](https://wasilak.github.io/elastauth/deployment/modes/)** - Choosing the right mode

## Security Warning

⚠️ **This example uses demo credentials for testing only!**

For production:
1. Change all secrets in `.env.example`
2. Enable TLS/HTTPS for Elasticsearch
3. Use proper secret management
4. Configure TLS in `proxy.tls` settings
5. Review security settings in Authelia and Elasticsearch
6. Use strong passwords and rotate credentials regularly

## Support

- [GitHub Issues](https://github.com/wasilak/elastauth/issues)
- [Full Documentation](https://wasilak.github.io/elastauth/)
