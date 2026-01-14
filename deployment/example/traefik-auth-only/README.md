# Traefik Integration Example (Auth-Only Mode)

This example demonstrates elastauth in **auth-only mode** with Traefik forward authentication middleware.

> 📚 **Full Documentation**: See the [Auth-Only Mode Guide](https://wasilak.github.io/elastauth/deployment/auth-only-mode/) for complete architecture details, configuration options, and troubleshooting.

## Quick Start

### Prerequisites

- Docker Engine 20.10 or later
- Docker Compose V2 (integrated with Docker CLI)
- At least 2GB RAM available

### 1. Add Hosts Entries

Add to `/etc/hosts`:

```
127.0.0.1 es.localhost auth.localhost elastauth.localhost
```

### 2. Start the Stack

```bash
docker compose up -d
```

### 3. Wait for Services

```bash
# Check status
docker compose ps

# Wait for Elasticsearch (1-2 minutes)
docker compose logs -f elasticsearch
```

### 4. Test Access

1. Open browser: `http://es.localhost`
2. Login with Authelia:
   - Username: `john`
   - Password: `password`
3. View Elasticsearch response

## Configuration

### Test Users

Defined in `authelia/users_database.yml`:

- **john**: Groups: `admins`, `dev`
- **jane**: Groups: `dev`

Both use password: `password`

### Environment Variables

See `.env.example` for all configuration options.

Key settings:
- `ELASTAUTH_PROXY_ENABLED=false` - Auth-only mode
- `ELASTAUTH_AUTH_PROVIDER=authelia` - Use Authelia headers
- `ELASTAUTH_CACHE_TYPE=redis` - Redis caching

## Cleanup

```bash
# Stop services
docker compose down

# Remove volumes
docker compose down -v
```

## Documentation

For detailed information, see:

- **[Auth-Only Mode Guide](https://wasilak.github.io/elastauth/deployment/auth-only-mode/)** - Complete deployment guide
- **[Configuration Reference](https://wasilak.github.io/elastauth/configuration/)** - All configuration options
- **[Troubleshooting](https://wasilak.github.io/elastauth/guides/troubleshooting/)** - Common issues and solutions

## Security Warning

⚠️ **This example uses demo credentials for testing only!**

For production:
1. Change all secrets in `.env.example`
2. Enable TLS/HTTPS
3. Use proper secret management
4. Review security settings in Authelia and Elasticsearch

## Support

- [GitHub Issues](https://github.com/wasilak/elastauth/issues)
- [Full Documentation](https://wasilak.github.io/elastauth/)
