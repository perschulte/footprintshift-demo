# GreenWeb Cache Service

A comprehensive, production-ready cache service built for the GreenWeb API that provides Redis-based caching with graceful degradation, health monitoring, and performance metrics.

## Features

- **Redis Integration**: Full Redis support with connection pooling and retry logic
- **Graceful Degradation**: Continues operation when cache is unavailable
- **Health Monitoring**: Real-time health checks and status reporting
- **Performance Metrics**: Hit rates, error rates, and efficiency scoring
- **Type Safety**: Strong typing with proper error handling
- **Middleware Support**: Gin middleware for HTTP integration
- **Batch Operations**: Efficient batch get/set/delete operations
- **Cache-Aside Pattern**: Built-in support for cache-aside pattern
- **Management API**: HTTP endpoints for cache administration

## Quick Start

### Basic Usage

```go
package main

import (
    "context"
    "time"
    "github.com/perschulte/greenweb-api/internal/cache"
)

func main() {
    // Initialize cache service from environment variables
    cacheService := cache.NewFromEnv()
    defer cacheService.Close()
    
    ctx := context.Background()
    
    // Set a value
    err := cacheService.Set(ctx, "my-key", "my-value", 10*time.Minute)
    if err != nil {
        // Handle error (cache might be disabled)
    }
    
    // Get a value
    var value string
    err = cacheService.Get(ctx, "my-key", &value)
    if cache.IsCacheMiss(err) {
        // Cache miss - fetch from source
    } else if err != nil {
        // Other error
    }
    
    // Use cache-aside pattern
    var result MyStruct
    err = cacheService.GetOrSet(ctx, "complex-key", &result, time.Hour, func() (interface{}, error) {
        // This function is called only on cache miss
        return fetchFromDatabase()
    })
}
```

### Configuration

The cache service can be configured through environment variables or programmatically:

#### Environment Variables

```bash
# Redis connection
REDIS_URL=redis://localhost:6379
REDIS_MAX_RETRIES=3
REDIS_MIN_RETRY_BACKOFF_MS=100
REDIS_MAX_RETRY_BACKOFF_MS=1000

# Timeouts
REDIS_DIAL_TIMEOUT_MS=5000
REDIS_READ_TIMEOUT_MS=3000
REDIS_WRITE_TIMEOUT_MS=3000

# Connection pool
REDIS_POOL_SIZE=10
REDIS_MIN_IDLE_CONNS=5
REDIS_MAX_CONN_AGE_MINUTES=60
REDIS_POOL_TIMEOUT_MS=4000
REDIS_IDLE_TIMEOUT_MINUTES=5
REDIS_IDLE_CHECK_FREQ_MINUTES=1

# Cache behavior
CACHE_KEY_PREFIX=greenweb
CACHE_DEFAULT_TTL_MINUTES=15
CACHE_ENABLE_HEALTH_CHECK=true
CACHE_HEALTH_CHECK_INTERVAL_SECONDS=30
CACHE_ENABLE_FALLBACK=true
CACHE_GRACEFUL_DEGRADATION=true
```

#### Programmatic Configuration

```go
config := &cache.Config{
    URL:                 "redis://localhost:6379",
    MaxRetries:          3,
    PoolSize:            10,
    MinIdleConns:        5,
    KeyPrefix:           "myapp",
    DefaultTTL:          15 * time.Minute,
    EnableHealthCheck:   true,
    GracefulDegradation: true,
}

cacheService := cache.New(config)
```

## Integration with Gin

### Middleware

```go
package main

import (
    "github.com/gin-gonic/gin"
    "github.com/perschulte/greenweb-api/internal/cache"
)

func main() {
    r := gin.Default()
    cacheService := cache.NewFromEnv()
    
    // Add cache middleware
    cacheMiddleware := cache.NewCacheMiddleware(cacheService)
    r.Use(cacheMiddleware.HeaderMiddleware())
    r.Use(cacheMiddleware.ConditionalCacheMiddleware())
    
    // Optional: Rate limiting using cache
    r.Use(cacheMiddleware.RateLimitByCache(100, time.Hour))
    
    // Add cache management routes
    cacheHandler := cache.NewManagementHandler(cacheService)
    cacheHandler.RegisterRoutes(r) // Adds /cache/* routes
}
```

### Cache Management Endpoints

The cache service provides the following HTTP endpoints for management:

- `GET /cache/stats` - Get cache statistics
- `POST /cache/stats/reset` - Reset cache statistics
- `GET /cache/health` - Get cache health status
- `DELETE /cache/keys/:key` - Delete a specific cache key
- `GET /cache/status` - Get comprehensive cache status

## Advanced Usage

### Cache-Aside Pattern with Services

```go
// Electricity Maps service with caching
func getCarbonIntensityWithCache(location string) (*CarbonIntensity, error) {
    cachedElectricityMaps := cache.NewCachedElectricityMaps(cacheService)
    
    result, err := cachedElectricityMaps.CachedCarbonIntensity(ctx, location, func() (interface{}, error) {
        // This is only called on cache miss
        return electricityMapsClient.GetCarbonIntensity(ctx, location)
    })
    
    if err != nil {
        return nil, err
    }
    
    return result.(*CarbonIntensity), nil
}
```

### Batch Operations

```go
batchOps := cache.NewBatchCacheOperations(cacheService)

// Batch get
keys := []string{"key1", "key2", "key3"}
results, err := batchOps.BatchGet(ctx, keys)

// Batch set
entries := map[string]interface{}{
    "key1": "value1",
    "key2": "value2", 
    "key3": "value3",
}
err = batchOps.BatchSet(ctx, entries, time.Hour)

// Batch delete
err = batchOps.BatchDelete(ctx, keys)
```

### Cache Management

```go
cacheManager := cache.NewCacheManager(cacheService)

// Get comprehensive metrics
metrics, err := cacheManager.GetCacheMetrics(ctx)
fmt.Printf("Cache efficiency: %.1f%%\n", metrics.EfficiencyScore)

// Clear location-specific cache
err = cacheManager.ClearLocationCache(ctx, "Berlin")

// Warm up cache for common locations
locations := []string{"Berlin", "London", "Paris"}
err = cacheManager.WarmUpCache(ctx, locations)
```

## Error Handling

The cache service provides specific error types for better error handling:

```go
err := cacheService.Get(ctx, "my-key", &value)

switch {
case cache.IsCacheMiss(err):
    // Cache miss - normal case, fetch from source
    value = fetchFromSource()
    
case cache.IsCacheDisabled(err):
    // Cache is disabled - continue without caching
    value = fetchFromSource()
    
case err != nil:
    // Other error (connection, serialization, etc.)
    return fmt.Errorf("cache error: %w", err)
}
```

## Monitoring and Metrics

### Built-in Metrics

The cache service tracks the following metrics:

- **Hits**: Number of successful cache retrievals
- **Misses**: Number of cache misses  
- **Sets**: Number of cache writes
- **Errors**: Number of cache errors
- **Total Requests**: Total number of cache operations
- **Hit Rate**: Percentage of successful hits
- **Efficiency Score**: Overall cache effectiveness (0-100)

### Health Monitoring

```go
health := cacheService.Health(ctx)

fmt.Printf("Cache healthy: %v\n", health.Healthy)
fmt.Printf("Status: %s\n", health.Status)
fmt.Printf("Response time: %v\n", health.ResponseTime)

if !health.Healthy {
    fmt.Printf("Last error: %s\n", health.LastError)
}
```

## Performance Characteristics

### TTL Settings

The cache service uses different TTL values for different data types:

- **Carbon Intensity**: 5 minutes (frequently changing)
- **Optimization Profiles**: 10 minutes (semi-static)
- **Green Hours Forecast**: 1 hour (predictive data)

### Connection Pool

Default connection pool settings optimized for typical workloads:

- **Pool Size**: 10 connections
- **Min Idle Connections**: 5
- **Max Connection Age**: 60 minutes
- **Idle Timeout**: 5 minutes

## Best Practices

### 1. Always Handle Cache Misses

```go
var data MyData
err := cache.Get(ctx, key, &data)
if cache.IsCacheMiss(err) {
    // Fetch from primary source
    data = fetchFromPrimarySource()
    // Optionally cache the result
    cache.Set(ctx, key, data, ttl)
}
```

### 2. Use Appropriate TTL Values

```go
// Fast-changing data
cache.Set(ctx, key, data, 5*time.Minute)

// Semi-static data  
cache.Set(ctx, key, data, 1*time.Hour)

// Static reference data
cache.Set(ctx, key, data, 24*time.Hour)
```

### 3. Enable Graceful Degradation

```go
config := cache.LoadFromEnv()
config.GracefulDegradation = true // Application continues if cache fails
```

### 4. Monitor Cache Performance

```go
// Regular monitoring
stats := cacheService.GetStats()
if stats.HitRate < 70 { // Less than 70% hit rate
    log.Warn("Low cache hit rate", "hit_rate", stats.HitRate)
}

if stats.ErrorRate > 5 { // More than 5% error rate
    log.Error("High cache error rate", "error_rate", stats.ErrorRate)
}
```

### 5. Use Structured Keys

```go
// Good: structured, predictable keys
key := cacheService.GenerateKey("carbon_intensity", "Berlin")
key := cacheService.GenerateKey("optimization", "Berlin", "shop.com")

// Avoid: unstructured keys
key := "data_Berlin_123_temp"
```

## Testing

The cache service is designed for easy testing:

```go
func TestMyService(t *testing.T) {
    // Use in-memory cache for testing
    config := cache.DefaultConfig()
    config.URL = "redis://localhost:6379/1" // Use test database
    
    testCache := cache.New(config)
    defer testCache.Close()
    
    // Your tests here
}
```

## Deployment Considerations

### Production Settings

```bash
# Production Redis settings
REDIS_URL=redis://prod-redis:6379
REDIS_POOL_SIZE=20
REDIS_MIN_IDLE_CONNS=10
CACHE_GRACEFUL_DEGRADATION=true
CACHE_ENABLE_HEALTH_CHECK=true
```

### High Availability

- Use Redis Cluster or Redis Sentinel for HA
- Set `GracefulDegradation=true` for resilience
- Monitor cache health endpoints
- Set up alerts for low hit rates or high error rates

### Security

- Use Redis AUTH if available
- Use TLS for Redis connections in production
- Implement proper network security for Redis instances
- Regularly rotate Redis credentials

## Troubleshooting

### Common Issues

1. **Cache Always Disabled**
   - Check Redis URL and connectivity
   - Verify Redis instance is running
   - Check network/firewall settings

2. **Low Hit Rate**
   - Review TTL settings (might be too short)
   - Check if keys are being generated consistently
   - Monitor for high eviction rates

3. **High Error Rate**
   - Check Redis instance health
   - Review connection pool settings
   - Check for serialization issues

4. **Performance Issues**
   - Increase connection pool size
   - Reduce TTL for less important data
   - Implement batch operations where possible

### Debug Mode

Enable debug headers to troubleshoot caching:

```bash
curl -H "Cache-Debug: true" http://localhost:8090/api/v1/carbon-intensity
```

This adds debug headers showing cache performance metrics.

## Contributing

When contributing to the cache service:

1. Maintain backward compatibility
2. Add comprehensive tests
3. Update documentation
4. Follow Go best practices
5. Consider performance implications

## License

This cache service is part of the GreenWeb project and follows the same licensing terms.