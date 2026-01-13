# Troubleshooting Guide

This guide covers common issues and solutions for elastauth deployments.

## Common Issues

### Authentication Failures

#### Symptoms
- HTTP 400/401 responses from elastauth
- "Authentication failed" errors in logs
- Users cannot access Kibana

#### Diagnosis

1. **Check elastauth logs**:
```bash
docker logs elastauth-container
# or
kubectl logs deployment/elastauth
```

2. **Test authentication directly**:
```bash
# For Authelia provider
curl -H "Remote-User: testuser" \
     -H "Remote-Groups: admin" \
     http://elastauth:5000/

# For OIDC provider  
curl -H "Authorization: Bearer YOUR_JWT_TOKEN" \
     http://elastauth:5000/
```

3. **Verify provider configuration**:
```bash
# Check config endpoint
curl http://elastauth:5000/config
```

#### Solutions

**Authelia Provider Issues**:
- Verify headers are being sent by proxy (Traefik/nginx)
- Check header names match configuration
- Ensure Authelia is setting headers correctly

**OIDC Provider Issues**:
- Verify JWT token is valid and not expired
- Check issuer URL matches provider configuration
- Ensure required claims are present in token
- Validate JWKS endpoint is accessible

### Elasticsearch Connection Issues

#### Symptoms
- "Failed to initialize Elasticsearch client" errors
- HTTP 500 responses from elastauth
- Cannot create/update users in Elasticsearch

#### Diagnosis

1. **Test Elasticsearch connectivity**:
```bash
# From elastauth container
curl -u elastauth:password http://elasticsearch:9200/_cluster/health
```

2. **Check Elasticsearch logs**:
```bash
docker logs elasticsearch-container
```

3. **Verify credentials**:
```bash
# Test authentication
curl -u elastauth:password http://elasticsearch:9200/_security/_authenticate
```

#### Solutions

- **Connection refused**: Check Elasticsearch is running and accessible
- **Authentication failed**: Verify elastauth user exists and has correct permissions
- **SSL/TLS issues**: Check certificate configuration and trust
- **Network issues**: Verify network connectivity between containers/pods

### Cache Connection Issues

#### Symptoms
- "Failed to connect to cache" errors
- Slow response times
- Cache misses in logs

#### Diagnosis

1. **Test cache connectivity**:
```bash
# Redis
redis-cli -h redis-host -p 6379 ping

# From elastauth container
telnet redis-host 6379
```

2. **Check cache logs**:
```bash
docker logs redis-container
```

#### Solutions

**Redis Issues**:
- Verify Redis is running and accessible
- Check Redis authentication if enabled
- Ensure Redis database number is correct
- Verify network connectivity

**Memory Cache Issues**:
- Check available memory on elastauth instance
- Monitor memory usage growth
- Consider TTL adjustment if memory usage is high

### Configuration Issues

#### Symptoms
- elastauth fails to start
- "Configuration validation failed" errors
- Unexpected behavior

#### Diagnosis

1. **Validate configuration**:
```bash
./elastauth --config config.yml --validate
```

2. **Check configuration loading**:
```bash
# Enable debug logging
LOG_LEVEL=debug ./elastauth
```

3. **Review configuration**:
```bash
# Check effective configuration
curl http://elastauth:5000/config
```

#### Solutions

- **Invalid YAML**: Check YAML syntax and indentation
- **Missing required fields**: Add all required configuration fields
- **Environment variable issues**: Verify environment variables are set correctly
- **Provider configuration**: Ensure provider-specific settings are correct

## Performance Issues

### Slow Response Times

#### Symptoms
- High response times from elastauth
- Timeouts from clients
- Poor user experience

#### Diagnosis

1. **Check response times**:
```bash
# Test response time
time curl -H "Remote-User: testuser" http://elastauth:5000/
```

2. **Monitor cache hit rates**:
```bash
# Check logs for cache hits/misses
grep -E "(Cache hit|Cache miss)" elastauth.log
```

3. **Check Elasticsearch performance**:
```bash
# Elasticsearch cluster health
curl http://elasticsearch:9200/_cluster/health
```

#### Solutions

- **Enable caching**: Add Redis cache to reduce Elasticsearch calls
- **Tune cache TTL**: Adjust cache expiration for better hit rates
- **Scale Elasticsearch**: Add more Elasticsearch nodes if needed
- **Optimize network**: Reduce network latency between components

### High Memory Usage

#### Symptoms
- elastauth consuming excessive memory
- Out of memory errors
- Container restarts

#### Diagnosis

1. **Monitor memory usage**:
```bash
# Docker
docker stats elastauth-container

# Kubernetes
kubectl top pod elastauth-pod
```

2. **Check cache size**:
```bash
# Redis memory usage
redis-cli info memory

# Memory cache (check logs for cache size)
grep "cache size" elastauth.log
```

#### Solutions

- **Reduce cache TTL**: Lower cache expiration time
- **Switch to Redis**: Use Redis instead of memory cache
- **Increase memory limits**: Allocate more memory to elastauth
- **Monitor cache growth**: Set up alerts for memory usage

## Security Issues

### Token Validation Failures

#### Symptoms
- "Token validation failed" errors
- Invalid signature errors
- Expired token errors

#### Diagnosis

1. **Decode JWT token**:
```bash
# Decode token payload (replace with actual token)
echo "JWT_PAYLOAD_PART" | base64 -d | jq .
```

2. **Check token expiration**:
```bash
# Check exp claim in decoded token
date -d @EXPIRATION_TIMESTAMP
```

3. **Verify JWKS endpoint**:
```bash
curl https://your-provider.com/.well-known/jwks.json
```

#### Solutions

- **Clock skew**: Synchronize system clocks
- **Expired tokens**: Refresh tokens before expiration
- **Invalid signatures**: Verify JWKS endpoint and issuer configuration
- **Network issues**: Ensure JWKS endpoint is accessible

### Permission Issues

#### Symptoms
- Users cannot access Kibana features
- "Access denied" errors in Kibana
- Missing roles or permissions

#### Diagnosis

1. **Check user roles**:
```bash
# Get user info from Elasticsearch
curl -u elastauth:password \
     http://elasticsearch:9200/_security/user/username
```

2. **Verify role mappings**:
```bash
# Check elastauth configuration
curl http://elastauth:5000/config
```

3. **Test Elasticsearch permissions**:
```bash
# Test user permissions
curl -u username:temp-password \
     http://elasticsearch:9200/_security/_authenticate
```

#### Solutions

- **Update role mappings**: Adjust group_mappings in configuration
- **Check default roles**: Ensure default_roles are appropriate
- **Verify group extraction**: Check that groups are correctly extracted from provider
- **Elasticsearch role configuration**: Verify roles exist in Elasticsearch

## Deployment Issues

### Container Startup Issues

#### Symptoms
- Container fails to start
- Exit code 1 errors
- Configuration validation failures

#### Diagnosis

1. **Check container logs**:
```bash
docker logs elastauth-container
```

2. **Verify configuration mounting**:
```bash
# Check if config file is mounted correctly
docker exec elastauth-container cat /config.yml
```

3. **Test configuration**:
```bash
# Validate configuration outside container
./elastauth --config config.yml --validate
```

#### Solutions

- **Fix configuration**: Correct configuration errors
- **Check file permissions**: Ensure config file is readable
- **Verify mounts**: Check volume mounts are correct
- **Environment variables**: Ensure required environment variables are set

### Kubernetes Deployment Issues

#### Symptoms
- Pods in CrashLoopBackOff state
- Service not accessible
- ConfigMap/Secret issues

#### Diagnosis

1. **Check pod status**:
```bash
kubectl get pods -l app=elastauth
kubectl describe pod elastauth-pod
```

2. **Check logs**:
```bash
kubectl logs deployment/elastauth
```

3. **Verify ConfigMap/Secrets**:
```bash
kubectl get configmap elastauth-config -o yaml
kubectl get secret elastauth-secrets -o yaml
```

#### Solutions

- **Fix resource definitions**: Correct Kubernetes YAML files
- **Check resource limits**: Ensure adequate CPU/memory limits
- **Verify networking**: Check service and ingress configuration
- **ConfigMap/Secret updates**: Restart pods after configuration changes

## Monitoring and Alerting

### Health Check Failures

#### Symptoms
- Health check endpoint returns errors
- Load balancer removes elastauth from rotation
- Service marked as unhealthy

#### Diagnosis

1. **Test health endpoint**:
```bash
curl http://elastauth:5000/health
```

2. **Check dependencies**:
```bash
# Test Elasticsearch
curl http://elasticsearch:9200/_cluster/health

# Test Redis (if used)
redis-cli -h redis-host ping
```

#### Solutions

- **Fix dependency issues**: Resolve Elasticsearch/Redis connectivity
- **Adjust health check timeouts**: Increase timeout values if needed
- **Monitor dependencies**: Set up monitoring for all dependencies

### Log Analysis

#### Common Log Patterns

**Successful authentication**:
```json
{
  "level": "info",
  "msg": "User authenticated successfully",
  "user": "john.doe",
  "provider": "oidc"
}
```

**Cache operations**:
```json
{
  "level": "debug",
  "msg": "Cache hit",
  "cacheKey": "elastauth-encoded-username",
  "user": "john.doe"
}
```

**Errors**:
```json
{
  "level": "error", 
  "msg": "Failed to get user from provider",
  "error": "token validation failed",
  "provider": "oidc"
}
```

## Getting Help

### Information to Collect

When seeking help, collect:

1. **elastauth version**: `./elastauth --version`
2. **Configuration**: Sanitized configuration file
3. **Logs**: Recent elastauth logs with errors
4. **Environment**: Deployment method (Docker, Kubernetes, etc.)
5. **Provider details**: Authentication provider type and version

### Support Channels

- **GitHub Issues**: [elastauth repository](https://github.com/wasilak/elastauth/issues)
- **Documentation**: Check all relevant documentation sections
- **Community**: Search existing issues and discussions

### Debug Mode

Enable debug logging for detailed troubleshooting:

```yaml
# config.yml
log_level: "debug"

# Or via environment variable
LOG_LEVEL=debug ./elastauth
```

## Prevention

### Best Practices

1. **Configuration validation**: Always validate configuration before deployment
2. **Health monitoring**: Set up health checks and monitoring
3. **Log aggregation**: Centralize logs for easier troubleshooting
4. **Testing**: Test authentication flows in staging environment
5. **Documentation**: Keep deployment documentation up to date

### Monitoring Setup

```yaml
# Example monitoring configuration
monitoring:
  health_checks:
    - endpoint: "/health"
      interval: "30s"
      timeout: "5s"
  
  metrics:
    - cache_hit_rate
    - response_time
    - error_rate
    
  alerts:
    - high_error_rate
    - cache_connectivity
    - elasticsearch_connectivity
```