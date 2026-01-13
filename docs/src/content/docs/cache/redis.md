---
title: Redis Cache Provider
description: Configure distributed Redis caching for multi-instance elastauth deployments
---

Redis cache enables distributed caching for elastauth deployments with multiple instances. It allows sharing cached authentication credentials across instances for consistent user experience.

## Configuration

### Basic Redis Cache

```yaml
cache:
  type: "redis"
  expiration: "1h"
  redis_host: "localhost:6379"
  redis_db: 0
```

### Redis with Authentication

```yaml
cache:
  type: "redis"
  expiration: "2h"
  redis_host: "redis.example.com:6379"
  redis_db: 0
  redis_password: "${REDIS_PASSWORD}"
  redis_username: "elastauth"
```

### Environment Variables

You can override Redis configuration using environment variables:

```bash
CACHE_TYPE=redis
CACHE_EXPIRATION=4h
CACHE_REDIS_HOST=redis:6379
CACHE_REDIS_DB=0
CACHE_REDIS_PASSWORD=your-redis-password
CACHE_REDIS_USERNAME=elastauth
```

## Configuration Options

| Option | Description | Default | Required |
|--------|-------------|---------|----------|
| `type` | Must be "redis" | - | Yes |
| `expiration` | Cache TTL (e.g., "1h", "30m") | "1h" | No |
| `redis_host` | Redis server host:port | "localhost:6379" | No |
| `redis_db` | Redis database number | 0 | No |
| `redis_password` | Redis password | - | No |
| `redis_username` | Redis username | - | No |

## Multi-Instance Considerations

When running multiple elastauth instances with Redis cache:

1. **Shared Secret Key**: All instances must use the same `secret_key` for credential encryption
2. **Same Database**: All instances should use the same `redis_db` number
3. **Consistent Configuration**: Cache expiration and Redis connection settings should be identical

### Example Multi-Instance Setup

```yaml
# Configuration for all elastauth instances
cache:
  type: "redis"
  expiration: "2h"
  redis_host: "${REDIS_HOST}"
  redis_db: 0
  redis_password: "${REDIS_PASSWORD}"

# CRITICAL: Must be identical across all instances
secret_key: "${SECRET_KEY}"
```

## Troubleshooting

### Connection Issues

**Symptoms**: elastauth fails to start or logs Redis connection errors

**Solutions**:
1. Verify Redis server is running and accessible
2. Check Redis host and port configuration
3. Verify Redis authentication credentials
4. Test Redis connectivity: `redis-cli -h host -p port ping`

### Cache Misses

**Symptoms**: Frequent Elasticsearch user creation requests

**Solutions**:
1. Check Redis memory usage and eviction policy
2. Verify cache expiration settings are appropriate
3. Ensure all elastauth instances use the same Redis database
4. Confirm `secret_key` is identical across instances

### Performance Issues

**Symptoms**: Slow authentication responses

**Solutions**:
1. Monitor Redis response times
2. Check Redis memory usage
3. Verify network latency between elastauth and Redis
4. Consider adjusting cache expiration times

## Related Documentation

- **[Cache Configuration](/elastauth/cache/)** - Cache configuration options
- **[Troubleshooting](/elastauth/guides/troubleshooting)** - Common issues and solutions