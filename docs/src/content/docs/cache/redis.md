---
title: Redis Cache Provider
description: Configure distributed Redis caching for multi-instance deployments
---

Redis cache provides distributed caching for multi-instance elastauth deployments. It's the recommended cache provider for production environments with multiple elastauth instances.

## Overview

Redis cache offers:

- **Distributed Caching**: Shared cache across multiple elastauth instances
- **High Performance**: In-memory storage with optional persistence
- **Scalability**: Supports horizontal scaling of elastauth
- **Reliability**: Battle-tested caching solution

## Configuration Examples

### Basic Single-Instance Setup

```yaml
# Simple Redis cache configuration
cache:
  type: "redis"
  expiration: "1h"
  redis_host: "localhost:6379"
  redis_db: 0

# Complete elastauth configuration with Redis
auth_provider: "oidc"

oidc:
  issuer: "https://auth.example.com"
  client_id: "elastauth"
  client_secret: "${OIDC_CLIENT_SECRET}"

cache:
  type: "redis"
  expiration: "1h"
  redis_host: "redis:6379"
  redis_db: 0

elasticsearch:
  hosts: ["https://elasticsearch:9200"]
  username: "elastauth"
  password: "${ELASTICSEARCH_PASSWORD}"

default_roles:
  - "kibana_user"
```

### Production Multi-Instance Setup

```yaml
# Production Redis cache with authentication
cache:
  type: "redis"
  expiration: "2h"
  redis_host: "${REDIS_HOST}"
  redis_db: "${REDIS_DB:-0}"
  redis_password: "${REDIS_PASSWORD}"
  redis_username: "${REDIS_USERNAME}"

# Complete production configuration
auth_provider: "oidc"

oidc:
  issuer: "${OIDC_ISSUER}"
  client_id: "${OIDC_CLIENT_ID}"
  client_secret: "${OIDC_CLIENT_SECRET}"
  token_validation: "both"
  use_pkce: true

cache:
  type: "redis"
  expiration: "${CACHE_EXPIRATION:-4h}"
  redis_host: "${REDIS_HOST}"
  redis_db: "${REDIS_DB:-0}"
  redis_password: "${REDIS_PASSWORD}"
  redis_username: "${REDIS_USERNAME:-elastauth}"

elasticsearch:
  hosts:
    - "${ELASTICSEARCH_HOST_1}"
    - "${ELASTICSEARCH_HOST_2}"
    - "${ELASTICSEARCH_HOST_3}"
  username: "${ELASTICSEARCH_USERNAME}"
  password: "${ELASTICSEARCH_PASSWORD}"

# Must be identical across all instances
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
```

### High-Availability Redis Setup

```yaml
# Enterprise Redis configuration with clustering
cache:
  type: "redis"
  expiration: "6h"
  redis_host: "${REDIS_CLUSTER_HOST}"
  redis_db: 0
  redis_password: "${REDIS_PASSWORD}"
  redis_username: "elastauth-prod"

# Environment-specific database isolation
# Production: redis_db: 0
# Staging: redis_db: 1  
# Development: redis_db: 2

auth_provider: "oidc"

oidc:
  issuer: "${OIDC_ISSUER}"
  client_id: "${OIDC_CLIENT_ID}"
  client_secret: "${OIDC_CLIENT_SECRET}"
  scopes: ["openid", "profile", "email", "groups"]
  claim_mappings:
    username: "${OIDC_USERNAME_CLAIM:-preferred_username}"
    email: "${OIDC_EMAIL_CLAIM:-email}"
    groups: "${OIDC_GROUPS_CLAIM:-groups}"
    full_name: "${OIDC_NAME_CLAIM:-name}"
  token_validation: "both"
  use_pkce: true

cache:
  type: "redis"
  expiration: "${CACHE_EXPIRATION:-6h}"
  redis_host: "${REDIS_CLUSTER_HOST}"
  redis_db: "${REDIS_DB:-0}"
  redis_password: "${REDIS_PASSWORD}"
  redis_username: "${REDIS_USERNAME:-elastauth-prod}"

elasticsearch:
  hosts:
    - "${ELASTICSEARCH_HOST_1}"
    - "${ELASTICSEARCH_HOST_2}"
    - "${ELASTICSEARCH_HOST_3}"
  username: "${ELASTICSEARCH_USERNAME}"
  password: "${ELASTICSEARCH_PASSWORD}"

secret_key: "${SECRET_KEY}"

default_roles:
  - "kibana_user"

group_mappings:
  platform-admins:
    - "kibana_admin"
    - "superuser"
  senior-developers:
    - "kibana_user"
    - "dev_full_access"
  developers:
    - "kibana_user"
    - "dev_limited_access"
  data-analysts:
    - "kibana_user"
    - "read_only"
  sre-team:
    - "kibana_user"
    - "monitoring_user"
```

### Docker Compose Production Setup

```yaml
# docker-compose.yml for production Redis deployment
version: '3.8'

services:
  redis:
    image: redis:7-alpine
    restart: unless-stopped
    ports:
      - "6379:6379"
    environment:
      - REDIS_PASSWORD=${REDIS_PASSWORD}
    command: >
      redis-server 
      --requirepass ${REDIS_PASSWORD}
      --appendonly yes
      --appendfsync everysec
      --maxmemory 512mb
      --maxmemory-policy allkeys-lru
    volumes:
      - redis_data:/data
      - ./redis.conf:/usr/local/etc/redis/redis.conf
    healthcheck:
      test: ["CMD", "redis-cli", "-a", "${REDIS_PASSWORD}", "ping"]
      interval: 30s
      timeout: 10s
      retries: 3
    
  elastauth-1:
    image: elastauth:latest
    restart: unless-stopped
    environment:
      - CACHE_TYPE=redis
      - CACHE_REDIS_HOST=redis:6379
      - CACHE_REDIS_PASSWORD=${REDIS_PASSWORD}
      - CACHE_EXPIRATION=2h
      - OIDC_CLIENT_SECRET=${OIDC_CLIENT_SECRET}
      - ELASTICSEARCH_PASSWORD=${ELASTICSEARCH_PASSWORD}
      - SECRET_KEY=${SECRET_KEY}
    volumes:
      - ./elastauth-config.yml:/config.yml
    depends_on:
      redis:
        condition: service_healthy
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:5000/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      
  elastauth-2:
    image: elastauth:latest
    restart: unless-stopped
    environment:
      - CACHE_TYPE=redis
      - CACHE_REDIS_HOST=redis:6379
      - CACHE_REDIS_PASSWORD=${REDIS_PASSWORD}
      - CACHE_EXPIRATION=2h
      - OIDC_CLIENT_SECRET=${OIDC_CLIENT_SECRET}
      - ELASTICSEARCH_PASSWORD=${ELASTICSEARCH_PASSWORD}
      - SECRET_KEY=${SECRET_KEY}
    volumes:
      - ./elastauth-config.yml:/config.yml
    depends_on:
      redis:
        condition: service_healthy
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:5000/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      
  nginx:
    image: nginx:alpine
    restart: unless-stopped
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf
      - ./ssl:/etc/nginx/ssl
    depends_on:
      - elastauth-1
      - elastauth-2
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost/health"]
      interval: 30s
      timeout: 10s
      retries: 3

volumes:
  redis_data:
    driver: local

networks:
  default:
    driver: bridge
```

### Kubernetes Production Deployment

```yaml
# redis-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: redis
  namespace: elastauth
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
        env:
        - name: REDIS_PASSWORD
          valueFrom:
            secretKeyRef:
              name: redis-secret
              key: password
        command:
        - redis-server
        - --requirepass
        - $(REDIS_PASSWORD)
        - --appendonly
        - "yes"
        - --maxmemory
        - "1gb"
        - --maxmemory-policy
        - "allkeys-lru"
        volumeMounts:
        - name: redis-data
          mountPath: /data
        resources:
          requests:
            memory: "256Mi"
            cpu: "100m"
          limits:
            memory: "1Gi"
            cpu: "500m"
        livenessProbe:
          exec:
            command:
            - redis-cli
            - -a
            - $(REDIS_PASSWORD)
            - ping
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          exec:
            command:
            - redis-cli
            - -a
            - $(REDIS_PASSWORD)
            - ping
          initialDelaySeconds: 5
          periodSeconds: 5
      volumes:
      - name: redis-data
        persistentVolumeClaim:
          claimName: redis-pvc

---
apiVersion: v1
kind: Service
metadata:
  name: redis
  namespace: elastauth
spec:
  selector:
    app: redis
  ports:
  - port: 6379
    targetPort: 6379
  type: ClusterIP

---
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: redis-pvc
  namespace: elastauth
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 10Gi
  storageClassName: fast-ssd

---
apiVersion: v1
kind: Secret
metadata:
  name: redis-secret
  namespace: elastauth
type: Opaque
data:
  password: <base64-encoded-password>

---
# elastauth-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: elastauth
  namespace: elastauth
spec:
  replicas: 3
  selector:
    matchLabels:
      app: elastauth
  template:
    metadata:
      labels:
        app: elastauth
    spec:
      containers:
      - name: elastauth
        image: elastauth:latest
        ports:
        - containerPort: 5000
        env:
        - name: CACHE_TYPE
          value: "redis"
        - name: CACHE_REDIS_HOST
          value: "redis:6379"
        - name: CACHE_REDIS_PASSWORD
          valueFrom:
            secretKeyRef:
              name: redis-secret
              key: password
        - name: CACHE_EXPIRATION
          value: "4h"
        - name: OIDC_CLIENT_SECRET
          valueFrom:
            secretKeyRef:
              name: oidc-secret
              key: client-secret
        - name: ELASTICSEARCH_PASSWORD
          valueFrom:
            secretKeyRef:
              name: elasticsearch-secret
              key: password
        - name: SECRET_KEY
          valueFrom:
            secretKeyRef:
              name: elastauth-secret
              key: secret-key
        volumeMounts:
        - name: config
          mountPath: /config.yml
          subPath: config.yml
        resources:
          requests:
            memory: "128Mi"
            cpu: "100m"
          limits:
            memory: "256Mi"
            cpu: "200m"
        livenessProbe:
          httpGet:
            path: /health
            port: 5000
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health
            port: 5000
          initialDelaySeconds: 5
          periodSeconds: 5
      volumes:
      - name: config
        configMap:
          name: elastauth-config
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

### Alternative Cache Providers
- **[Memory Cache](/elastauth/cache/)** - Simple in-memory caching for single instances
- **[File Cache](/elastauth/cache/file)** - Persistent file-based caching
- **[Cache Comparison](/elastauth/cache/comparison)** - Compare cache provider features

### Configuration and Integration
- **[Authentication Providers](/elastauth/providers/)** - Configure authentication with Redis cache
  - [Authelia with Redis](/elastauth/providers/authelia#production-multi-instance-setup) - Authelia + Redis setup
  - [OAuth2/OIDC with Redis](/elastauth/providers/oidc#production-multi-provider-setup) - OIDC + Redis setup
- **[Configuration Reference](/elastauth/guides/configuration)** - Complete configuration options
- **[Environment Variables](/elastauth/guides/environment)** - Environment-based configuration

### Deployment and Scaling
- **[Docker Deployment](/elastauth/deployment/docker)** - Container deployment with Redis
- **[Kubernetes Deployment](/elastauth/deployment/kubernetes)** - Production Kubernetes setup
- **[Scaling Guide](/elastauth/guides/scaling)** - Horizontal scaling with Redis cache
- **[Load Balancer Setup](/elastauth/guides/load-balancer)** - Production load balancing

### Operations and Monitoring
- **[Redis Monitoring](/elastauth/guides/redis-monitoring)** - Redis performance monitoring
- **[Cache Performance](/elastauth/guides/cache-performance)** - Cache optimization strategies
- **[Troubleshooting Redis](/elastauth/guides/troubleshooting#redis-cache-issues)** - Common Redis issues
- **[Backup and Recovery](/elastauth/guides/backup-recovery)** - Redis backup strategies

### Security and Best Practices
- **[Redis Security](/elastauth/guides/redis-security)** - Redis security configuration
- **[Network Security](/elastauth/guides/network-security)** - Secure Redis networking
- **[Performance Tuning](/elastauth/guides/performance-tuning)** - Redis performance optimization
- **[High Availability](/elastauth/guides/high-availability)** - Redis clustering and failover