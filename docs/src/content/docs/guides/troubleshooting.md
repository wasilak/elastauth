---
title: Upgrading elastauth
description: Guide for upgrading elastauth between major versions
---

This document provides guidance for upgrading elastauth between major versions.

## Cache Configuration Breaking Changes

### Overview

Starting with the next major version, elastauth has migrated to use the [cachego](https://github.com/wasilak/cachego) library for all caching operations. This change provides better performance, more cache provider options, and improved reliability.

**⚠️ BREAKING CHANGE**: Legacy cache configuration format is no longer supported.

### What Changed

The cache configuration format has been completely redesigned:

#### Old Configuration (No Longer Supported)
```yaml
# ❌ These settings will cause elastauth to fail to start
cache_type: "redis"
redis_host: "localhost:6379"
redis_db: 0
cache_expire: "1h"
```

#### New Configuration (Required)
```yaml
# ✅ New cachego-based configuration
cache:
  type: "redis"
  expiration: "1h"
  redis_host: "localhost:6379"
  redis_db: 0
```

### Migration Guide

#### 1. Memory Cache Migration

**Before:**
```yaml
cache_type: "memory"
cache_expire: "30m"
```

**After:**
```yaml
cache:
  type: "memory"
  expiration: "30m"
```

#### 2. Redis Cache Migration

**Before:**
```yaml
cache_type: "redis"
redis_host: "redis.example.com:6379"
redis_db: 1
cache_expire: "2h"
```

**After:**
```yaml
cache:
  type: "redis"
  expiration: "2h"
  redis_host: "redis.example.com:6379"
  redis_db: 1
```

#### 3. File Cache (New Feature)

The new configuration also supports file-based caching:

```yaml
cache:
  type: "file"
  expiration: "1h"
  path: "/var/cache/elastauth"
```

#### 4. No Cache Configuration

To disable caching entirely, simply omit the `cache` section or set `type` to empty:

```yaml
# Option 1: Omit cache section entirely
# (no cache configuration)

# Option 2: Explicitly disable
cache:
  type: ""
```

### Environment Variable Support

All cache settings can be overridden with environment variables:

```bash
# Cache type
ELASTAUTH_CACHE_TYPE=redis

# Cache expiration
ELASTAUTH_CACHE_EXPIRATION=2h

# Redis settings
ELASTAUTH_CACHE_REDIS_HOST=redis.example.com:6379
ELASTAUTH_CACHE_REDIS_DB=1

# File cache path
ELASTAUTH_CACHE_PATH=/var/cache/elastauth
```

### Configuration Validation

If you attempt to use the old configuration format, elastauth will fail to start with a clear error message:

```
ERROR: legacy cache configuration detected. Please migrate to new format:
  Old: cache_type, redis_host, cache_expire, redis_db
  New: cache.type, cache.redis_host, cache.expiration, cache.redis_db
See documentation for migration guide
```

### Benefits of the New System

1. **Better Performance**: The cachego library provides optimized caching operations
2. **More Cache Providers**: Support for memory, Redis, and file-based caching
3. **Improved Reliability**: Better error handling and connection management
4. **Consistent Configuration**: All cache settings are now under the `cache` namespace
5. **Environment Variable Support**: Full support for environment-based configuration

### Deployment Considerations

#### Docker Deployments

Update your Docker configuration files:

```yaml
# docker-compose.yml
version: '3.8'
services:
  elastauth:
    image: elastauth:latest
    environment:
      - ELASTAUTH_CACHE_TYPE=redis
      - ELASTAUTH_CACHE_REDIS_HOST=redis:6379
      - ELASTAUTH_CACHE_EXPIRATION=1h
    depends_on:
      - redis
  
  redis:
    image: redis:alpine
```

#### Kubernetes Deployments

Update your ConfigMap:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: elastauth-config
data:
  config.yml: |
    cache:
      type: "redis"
      expiration: "1h"
      redis_host: "redis-service:6379"
      redis_db: 0
    
    # ... other configuration
```

#### Helm Charts

Update your values.yaml:

```yaml
# values.yaml
elastauth:
  cache:
    type: redis
    expiration: 1h
    redis:
      host: redis-service:6379
      db: 0
```

### Troubleshooting

#### Common Migration Issues

1. **"legacy cache configuration detected" error**
   - **Cause**: Old configuration format is still present
   - **Solution**: Update all cache-related settings to use the new `cache.*` format

2. **"redis cache requires cache.redis_host configuration" error**
   - **Cause**: Redis cache type specified but no host configured
   - **Solution**: Add `cache.redis_host` setting

3. **Cache not working after upgrade**
   - **Cause**: Configuration format mismatch
   - **Solution**: Verify all cache settings use the new format and restart elastauth

#### Validation Checklist

Before upgrading, ensure your configuration:

- [ ] Uses `cache.type` instead of `cache_type`
- [ ] Uses `cache.expiration` instead of `cache_expire`
- [ ] Uses `cache.redis_host` instead of `redis_host`
- [ ] Uses `cache.redis_db` instead of `redis_db`
- [ ] Has no legacy cache configuration keys remaining

### Testing Your Migration

1. **Validate Configuration**: Run elastauth with `--dry-run` to validate configuration
2. **Check Logs**: Look for "Initialized cache with type: X" message
3. **Test Cache Operations**: Verify authentication requests are cached properly
4. **Monitor Performance**: Ensure cache hit rates are as expected

### Rollback Plan

If you need to rollback to the previous version:

1. Keep a backup of your old configuration file
2. Use the previous elastauth version with the old configuration format
3. Plan your migration more carefully before attempting the upgrade again

### Support

If you encounter issues during migration:

1. Check the error messages for specific guidance
2. Verify your configuration against the examples in this guide
3. Test with a minimal configuration first
4. Consult the elastauth documentation for additional examples

---

**Note**: This breaking change was introduced to improve the overall reliability and performance of elastauth's caching system. While it requires configuration updates, the new system provides significant benefits for production deployments.