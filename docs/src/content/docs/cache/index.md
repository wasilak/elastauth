---
title: Cache Providers
description: Configure caching for elastauth to improve performance
---

elastauth uses the [cachego](https://github.com/wasilak/cachego) library to provide flexible caching of encrypted user credentials. Caching reduces load on Elasticsearch and improves response times.

## Available Cache Providers

- **[Memory Cache](/elastauth/cache/memory)** - In-memory caching for single-instance deployments
- **[Redis Cache](/elastauth/cache/redis)** - Distributed caching for multi-instance deployments  
- **[File Cache](/elastauth/cache/file)** - File-based persistent caching
- **[No Cache](/elastauth/cache/disabled)** - Direct Elasticsearch calls on every request

## Cache Selection

Configure exactly zero or one cache provider:

```yaml
# Choose one cache type or omit for no caching
cache:
  type: "redis"  # or "memory", "file", or omit entirely
```

## Cache Behavior

### With Caching Enabled

1. **Cache Hit**: Return cached encrypted credentials
2. **Cache Miss**: Generate new credentials, store in Elasticsearch, cache encrypted credentials
3. **Cache Expiry**: Automatic cleanup based on TTL settings

### Without Caching

1. **Every Request**: Generate new credentials and call Elasticsearch API
2. **No Storage**: No persistent credential storage
3. **Higher Load**: More Elasticsearch API calls

## Security

### Credential Encryption

All cached credentials are encrypted using AES encryption:

- **Encryption Key**: Configured via `secret_key` setting
- **Secure Storage**: Credentials never stored in plain text
- **Key Management**: Same key required across all instances

### Cache Isolation

- **User Isolation**: Each user's credentials cached separately
- **Key Prefixing**: Cache keys prefixed with "elastauth-"
- **Secure Defaults**: No sensitive data in cache keys

## Horizontal Scaling

### Single Instance Deployments

All cache types supported:

```yaml
cache:
  type: "memory"    # ✅ Supported
  # or
  type: "file"      # ✅ Supported  
  # or
  type: "redis"     # ✅ Supported
```

### Multi-Instance Deployments

Only Redis cache supported for shared state:

```yaml
cache:
  type: "redis"     # ✅ Required for multi-instance
  redis_host: "redis:6379"
  
# These are NOT supported for multi-instance:
# type: "memory"    # ❌ Instance-local only
# type: "file"      # ❌ Instance-local only
```

## Configuration Validation

elastauth validates cache configuration at startup:

- **Single Cache Type**: Exactly zero or one cache type allowed
- **Required Settings**: Missing required settings cause startup failure
- **Connection Testing**: Cache connectivity verified at startup

### Valid Configurations

```yaml
# No caching
# (omit cache section entirely)

# Memory caching
cache:
  type: "memory"
  expiration: "1h"

# Redis caching  
cache:
  type: "redis"
  expiration: "1h"
  redis_host: "localhost:6379"
  redis_db: 0

# File caching
cache:
  type: "file"
  expiration: "1h"
  path: "/tmp/elastauth-cache"
```

### Invalid Configurations

```yaml
# ❌ Multiple cache types
cache:
  type: "redis"
redis:
  host: "localhost:6379"
memory:
  expiration: "1h"

# ❌ Missing required settings
cache:
  type: "redis"
  # Missing redis_host
```

## Performance Considerations

### Cache Hit Rates

- **High Hit Rate**: Fewer Elasticsearch calls, better performance
- **Low Hit Rate**: More Elasticsearch calls, consider TTL adjustment
- **Monitoring**: Monitor cache hit/miss ratios

### TTL Configuration

```yaml
cache:
  expiration: "1h"    # 1 hour TTL
  # or
  expiration: "30m"   # 30 minutes
  # or  
  expiration: "2h"    # 2 hours
```

### Memory Usage

- **Memory Cache**: Grows with active users
- **Redis Cache**: Shared memory across instances
- **File Cache**: Disk space usage

## Environment Variables

Override cache configuration via environment variables:

```bash
# Cache type selection
CACHE_TYPE="redis"

# Redis configuration
CACHE_REDIS_HOST="redis:6379"
CACHE_REDIS_DB="0"

# TTL configuration
CACHE_EXPIRATION="1h"

# File cache path
CACHE_PATH="/var/cache/elastauth"
```

## Migration Between Cache Types

### From No Cache to Cached

1. **Add Cache Configuration**: Configure desired cache type
2. **Restart elastauth**: Cache will be populated on new requests
3. **Monitor Performance**: Verify improved response times

### Between Cache Types

1. **Update Configuration**: Change cache type in configuration
2. **Restart elastauth**: Old cache data will be lost
3. **Warm Cache**: Cache will repopulate on new requests

### From Cached to No Cache

1. **Remove Cache Configuration**: Remove cache section from config
2. **Restart elastauth**: All requests will hit Elasticsearch directly
3. **Monitor Load**: Verify Elasticsearch can handle increased load

## Troubleshooting

### Cache Connection Issues

1. **Check Connectivity**: Verify cache service is accessible
2. **Validate Configuration**: Ensure all required settings are present
3. **Review Logs**: Check elastauth logs for specific errors

### Performance Issues

1. **Monitor Hit Rates**: Low hit rates indicate TTL too short
2. **Check Memory Usage**: High memory usage may indicate TTL too long
3. **Network Latency**: High latency to cache service affects performance

### Data Consistency

1. **Encryption Keys**: Ensure same key across all instances
2. **Cache Invalidation**: Consider manual cache clearing if needed
3. **Clock Synchronization**: Ensure system clocks are synchronized

## Next Steps

- [Redis Cache](/elastauth/cache/redis) - Distributed caching setup