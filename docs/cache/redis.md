# Redis Cache Provider

Redis cache provides distributed caching for multi-instance elastauth deployments. It's the recommended cache provider for production environments with multiple elastauth instances.

## Overview

Redis cache offers:

- **Distributed Caching**: Shared cache across multiple elastauth instances
- **High Performance**: In-memory storage with optional persistence
- **Scalability**: Supports horizontal scaling of elastauth
- **Reliability**: Battle-tested caching solution

## Configuration

### Basic Configuration

```yaml
cache:
  type: "redis"
  expiration: "1h"
  redis_host: "localhost:6379"
  redis_db: 0
```

### Advanced Configuration

```yaml
cache:
  type: "redis"
  expiration: "1h"
  redis_host: "redis.example.com:6379"
  redis_db: 0
  redis_password: "${REDIS_PASSWORD}"  # If Redis requires authentication
  redis_username: "elastauth"          # Redis 6+ ACL username
```

### Environment Variables

```bash
# Redis connection
CACHE_REDIS_HOST="redis:6379"
CACHE_REDIS_DB="0"
CACHE_REDIS_PASSWORD="your-redis-password"
CACHE_REDIS_USERNAME="elastauth"

# Cache settings
CACHE_EXPIRATION="1h"
```

## Redis Setup

### Docker Compose

```yaml
version: '3.8'

services:
  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data
    command: redis-server --appendonly yes
    
  elastauth:
    image: elastauth:latest
    environment:
      - CACHE_REDIS_HOST=redis:6379
      - CACHE_REDIS_DB=0
      - CACHE_EXPIRATION=1h
    depends_on:
      - redis
    volumes:
      - ./config.yml:/config.yml

volumes:
  redis_data:
```

### Redis with Authentication

```yaml
version: '3.8'

services:
  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    environment:
      - REDIS_PASSWORD=${REDIS_PASSWORD}
    command: redis-server --requirepass ${REDIS_PASSWORD} --appendonly yes
    volumes:
      - redis_data:/data
      
  elastauth:
    image: elastauth:latest
    environment:
      - CACHE_REDIS_HOST=redis:6379
      - CACHE_REDIS_PASSWORD=${REDIS_PASSWORD}
      - CACHE_EXPIRATION=1h
    depends_on:
      - redis
```

### Kubernetes Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: redis
spec:
  replicas: 1
  selector:
    matchLabels:
      app: redis
  template:
    metadata:
      labels:
        app: redis
    spec:
      containers:
      - name: redis
        image: redis:7-alpine
        ports:
        - containerPort: 6379
        command: ["redis-server", "--appendonly", "yes"]
        volumeMounts:
        - name: redis-data
          mountPath: /data
      volumes:
      - name: redis-data
        persistentVolumeClaim:
          claimName: redis-pvc

---
apiVersion: v1
kind: Service
metadata:
  name: redis
spec:
  selector:
    app: redis
  ports:
  - port: 6379
    targetPort: 6379
```

## Multi-Instance Configuration

### Shared Redis Instance

All elastauth instances must use the same Redis instance:

```yaml
# Instance 1
cache:
  type: "redis"
  redis_host: "shared-redis:6379"
  redis_db: 0
  expiration: "1h"

# Instance 2 (same configuration)
cache:
  type: "redis"
  redis_host: "shared-redis:6379"  # Same Redis instance
  redis_db: 0                      # Same database
  expiration: "1h"                 # Same TTL
```

### Load Balancer Configuration

```yaml
# docker-compose.yml
version: '3.8'

services:
  redis:
    image: redis:7-alpine
    volumes:
      - redis_data:/data
    command: redis-server --appendonly yes
    
  elastauth-1:
    image: elastauth:latest
    environment:
      - CACHE_REDIS_HOST=redis:6379
    volumes:
      - ./config.yml:/config.yml
    depends_on:
      - redis
      
  elastauth-2:
    image: elastauth:latest
    environment:
      - CACHE_REDIS_HOST=redis:6379
    volumes:
      - ./config.yml:/config.yml
    depends_on:
      - redis
      
  nginx:
    image: nginx:alpine
    ports:
      - "80:80"
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf
    depends_on:
      - elastauth-1
      - elastauth-2
```

## Redis Database Selection

### Database Isolation

Use different Redis databases for different environments:

```yaml
# Production
cache:
  redis_db: 0

# Staging  
cache:
  redis_db: 1

# Development
cache:
  redis_db: 2
```

### Namespace Isolation

elastauth automatically prefixes cache keys with "elastauth-":

```
# Cache key format
elastauth-{encoded-username}

# Example
elastauth-am9obi5kb2U%3D  # base64 encoded username
```

## Performance Tuning

### Connection Pooling

Redis connections are automatically pooled by the Go Redis client:

```yaml
# Connection pool settings (handled automatically)
# - Max connections: Based on GOMAXPROCS
# - Idle timeout: 5 minutes
# - Connection timeout: 5 seconds
```

### Memory Optimization

Configure Redis memory settings:

```bash
# redis.conf
maxmemory 256mb
maxmemory-policy allkeys-lru  # Evict least recently used keys
```

### Persistence Configuration

Choose Redis persistence strategy:

```bash
# Append-only file (recommended)
appendonly yes
appendfsync everysec

# Or RDB snapshots
save 900 1    # Save if at least 1 key changed in 900 seconds
save 300 10   # Save if at least 10 keys changed in 300 seconds
save 60 10000 # Save if at least 10000 keys changed in 60 seconds
```

## Monitoring

### Redis Metrics

Monitor Redis performance:

```bash
# Redis CLI monitoring
redis-cli monitor

# Redis info
redis-cli info

# Memory usage
redis-cli info memory

# Key statistics
redis-cli info keyspace
```

### elastauth Metrics

Monitor cache performance in elastauth logs:

```json
{
  "level": "debug",
  "msg": "Cache hit",
  "cacheKey": "elastauth-am9obi5kb2U%3D",
  "user": "john.doe"
}

{
  "level": "debug", 
  "msg": "Cache miss",
  "cacheKey": "elastauth-am9obi5kb2U%3D",
  "user": "john.doe"
}
```

## Security

### Redis Authentication

Enable Redis authentication:

```bash
# redis.conf
requirepass your-strong-password

# Or with ACL (Redis 6+)
user elastauth on >your-password ~elastauth-* +@read +@write
```

### Network Security

Secure Redis network access:

```bash
# Bind to specific interfaces
bind 127.0.0.1 10.0.0.1

# Disable dangerous commands
rename-command FLUSHDB ""
rename-command FLUSHALL ""
rename-command CONFIG ""
```

### TLS Encryption

Enable Redis TLS:

```bash
# redis.conf
tls-port 6380
tls-cert-file /path/to/redis.crt
tls-key-file /path/to/redis.key
tls-ca-cert-file /path/to/ca.crt
```

## Troubleshooting

### Connection Issues

```bash
# Test Redis connectivity
redis-cli -h redis-host -p 6379 ping

# Check Redis logs
docker logs redis-container

# Test from elastauth container
telnet redis-host 6379
```

### Authentication Problems

```bash
# Test with password
redis-cli -h redis-host -p 6379 -a your-password ping

# Check ACL users (Redis 6+)
redis-cli ACL LIST
```

### Performance Issues

```bash
# Check slow queries
redis-cli SLOWLOG GET 10

# Monitor commands
redis-cli MONITOR

# Check memory usage
redis-cli INFO MEMORY
```

### Cache Key Issues

```bash
# List elastauth keys
redis-cli KEYS "elastauth-*"

# Check key TTL
redis-cli TTL "elastauth-key"

# Manual key deletion
redis-cli DEL "elastauth-key"
```

## Backup and Recovery

### Redis Backup

```bash
# Create backup
redis-cli BGSAVE

# Copy RDB file
cp /var/lib/redis/dump.rdb /backup/dump-$(date +%Y%m%d).rdb
```

### Cache Recovery

Cache data is automatically regenerated:

1. **Cache Miss**: elastauth generates new credentials
2. **Elasticsearch Call**: Creates/updates user in Elasticsearch  
3. **Cache Population**: Stores encrypted credentials in Redis

## Migration

### From Other Cache Types

1. **Update Configuration**: Change cache type to "redis"
2. **Deploy Redis**: Set up Redis instance
3. **Restart elastauth**: Cache will populate on new requests
4. **Remove Old Cache**: Clean up previous cache configuration

### Redis Version Upgrades

1. **Backup Data**: Create Redis backup
2. **Stop elastauth**: Prevent new cache writes
3. **Upgrade Redis**: Follow Redis upgrade procedures
4. **Start elastauth**: Resume normal operation

## Next Steps

- [Memory Cache](memory.md) - Alternative for single-instance deployments
- [File Cache](file.md) - Persistent file-based caching
- [Configuration Examples](../examples/) - Complete configuration examples
- [Troubleshooting](../troubleshooting.md) - Common issues and solutions