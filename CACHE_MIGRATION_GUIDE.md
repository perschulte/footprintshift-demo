# Cache Service Migration Guide

This guide explains how to migrate from the old monolithic cache implementation to the new modular cache service.

## Overview of Changes

The cache implementation has been completely refactored into a modular, production-ready service with the following improvements:

### ✅ **What's New**
- **Modular Design**: Clean separation of concerns with interfaces, config, and implementation
- **Dependency Injection**: No more global variables, proper constructor-based configuration
- **Enhanced Error Handling**: Specific error types for better error handling
- **Health Monitoring**: Real-time health checks and metrics
- **Graceful Degradation**: Continues operation when Redis is unavailable
- **Management API**: HTTP endpoints for cache administration
- **Middleware Support**: Gin middleware for seamless integration
- **Batch Operations**: Efficient bulk operations
- **Testing Support**: Comprehensive test coverage and benchmarks

### 📁 **New Directory Structure**
```
internal/cache/
├── README.md           # Comprehensive documentation
├── config.go          # Configuration management
├── handlers.go         # HTTP management endpoints
├── integration.go      # Helper utilities for service integration
├── service.go          # Main cache service implementation
├── service_test.go     # Test suite
└── types.go           # Interfaces and type definitions
```

## Migration Steps

### 1. Update Imports

**Before:**
```go
// Old global cache service in main.go
var cacheService *CacheService
```

**After:**
```go
import "github.com/perschulte/greenweb-api/internal/cache"

var cacheService cache.Cacher
```

### 2. Initialize Cache Service

**Before:**
```go
func main() {
    // Old initialization - embedded in main.go
    cacheService = NewCacheService()
}
```

**After:**
```go
func main() {
    // New initialization with dependency injection
    cacheService = cache.NewFromEnv()
    defer cacheService.Close()
    
    // Optional: Create cache manager for advanced operations
    cacheManager := cache.NewCacheManager(cacheService)
}
```

### 3. Update Service Integration

**Before:**
```go
func getCarbonIntensity(c *gin.Context) {
    // Direct API call without caching
    intensity, err := electricityMapsClient.GetCarbonIntensity(ctx, location)
    // ...
}
```

**After:**
```go
func getCarbonIntensityWithCache(c *gin.Context) {
    cachedElectricityMaps := cache.NewCachedElectricityMaps(cacheService)
    
    result, err := cachedElectricityMaps.CachedCarbonIntensity(ctx, location, func() (interface{}, error) {
        return electricityMapsClient.GetCarbonIntensity(ctx, location)
    })
    // ...
}
```

### 4. Add Middleware

**New Feature:**
```go
func main() {
    r := gin.Default()
    
    // Add cache middleware
    cacheMiddleware := cache.NewCacheMiddleware(cacheService)
    r.Use(cacheMiddleware.HeaderMiddleware())
    r.Use(cacheMiddleware.ConditionalCacheMiddleware())
    
    // Add cache management routes
    cacheHandler := cache.NewManagementHandler(cacheService)
    cacheHandler.RegisterRoutes(r) // Adds /cache/* routes
}
```

### 5. Update Error Handling

**Before:**
```go
err := cacheService.Get(ctx, key, &value)
if err != nil {
    // Generic error handling
    log.Printf("Cache error: %v", err)
}
```

**After:**
```go
err := cacheService.Get(ctx, key, &value)
switch {
case cache.IsCacheMiss(err):
    // Cache miss - fetch from source
    value = fetchFromSource()
    
case cache.IsCacheDisabled(err):
    // Cache is disabled - continue without caching
    value = fetchFromSource()
    
case err != nil:
    // Other error (connection, serialization, etc.)
    return fmt.Errorf("cache error: %w", err)
}
```

## Configuration Changes

### Environment Variables

The new cache service uses more comprehensive configuration:

```bash
# Basic Redis connection (unchanged)
REDIS_URL=redis://localhost:6379

# New: Enhanced connection settings
REDIS_MAX_RETRIES=3
REDIS_MIN_RETRY_BACKOFF_MS=100
REDIS_MAX_RETRY_BACKOFF_MS=1000
REDIS_DIAL_TIMEOUT_MS=5000
REDIS_READ_TIMEOUT_MS=3000
REDIS_WRITE_TIMEOUT_MS=3000

# New: Connection pool settings
REDIS_POOL_SIZE=10
REDIS_MIN_IDLE_CONNS=5
REDIS_MAX_CONN_AGE_MINUTES=60
REDIS_POOL_TIMEOUT_MS=4000
REDIS_IDLE_TIMEOUT_MINUTES=5
REDIS_IDLE_CHECK_FREQ_MINUTES=1

# New: Cache behavior settings
CACHE_KEY_PREFIX=greenweb
CACHE_DEFAULT_TTL_MINUTES=15
CACHE_ENABLE_HEALTH_CHECK=true
CACHE_HEALTH_CHECK_INTERVAL_SECONDS=30
CACHE_ENABLE_FALLBACK=true
CACHE_GRACEFUL_DEGRADATION=true
```

## API Changes

### Cache Management Endpoints

The new cache service provides management endpoints:

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/cache/stats` | GET | Get cache statistics |
| `/cache/stats/reset` | POST | Reset cache statistics |
| `/cache/health` | GET | Get cache health status |
| `/cache/status` | GET | Get comprehensive cache status |
| `/cache/keys/:key` | DELETE | Delete a specific cache key |

### Response Changes

API responses now include cache status information:

**Before:**
```json
{
  "carbon_intensity": {...},
  "location": "Berlin"
}
```

**After:**
```json
{
  "carbon_intensity": {...},
  "cache_status": "hit",
  "location": "Berlin"
}
```

## Testing the Migration

### 1. Test Basic Functionality

```bash
# Check if cache is working
curl http://localhost:8090/cache/health

# Get cache statistics
curl http://localhost:8090/cache/stats

# Test API endpoints
curl http://localhost:8090/api/v1/carbon-intensity?location=Berlin
```

### 2. Verify Cache Hits

Look for cache status in API responses:
- `"cache_status": "hit"` - Data served from cache
- `"cache_status": "miss"` - Data fetched from source

### 3. Monitor Performance

```bash
# Check comprehensive cache status
curl http://localhost:8090/cache/status
```

## Rollback Plan

If issues arise, you can rollback by:

1. **Rename files:**
   ```bash
   mv cache_old.go cache.go
   mv main.go main_new.go
   mv main_with_cache.go main.go
   ```

2. **Remove new cache module:**
   ```bash
   rm -rf internal/cache/
   ```

3. **Update imports** back to the old structure

## Performance Improvements

The new cache service offers several performance improvements:

### 1. Connection Pooling
- Configurable pool size (default: 10 connections)
- Idle connection management
- Connection reuse optimization

### 2. Batch Operations
```go
batchOps := cache.NewBatchCacheOperations(cacheService)
results, err := batchOps.BatchGet(ctx, []string{"key1", "key2", "key3"})
```

### 3. Health-Based Circuit Breaking
- Automatic cache disabling on health check failures
- Graceful degradation when Redis is unavailable
- Automatic re-enabling when Redis recovers

### 4. Optimized Serialization
- JSON serialization with error handling
- Type-safe deserialization
- Configurable TTL per operation

## Monitoring and Observability

### Built-in Metrics

The new cache service tracks comprehensive metrics:

```go
stats := cacheService.GetStats()
fmt.Printf("Hit Rate: %.1f%%\n", stats.HitRate)
fmt.Printf("Total Requests: %d\n", stats.TotalRequests)
fmt.Printf("Errors: %d\n", stats.Errors)
```

### Health Monitoring

```go
health := cacheService.Health(ctx)
fmt.Printf("Cache Healthy: %v\n", health.Healthy)
fmt.Printf("Response Time: %v\n", health.ResponseTime)
```

### Cache Efficiency Score

```go
cacheManager := cache.NewCacheManager(cacheService)
metrics, err := cacheManager.GetCacheMetrics(ctx)
fmt.Printf("Efficiency Score: %.1f/100\n", metrics.EfficiencyScore)
```

## Best Practices for New Implementation

### 1. Use Cache-Aside Pattern
```go
var result MyData
err := cacheService.GetOrSet(ctx, key, &result, ttl, func() (interface{}, error) {
    return fetchFromDatabase()
})
```

### 2. Handle Cache Failures Gracefully
```go
if cache.IsCacheDisabled(err) {
    // Continue without caching
    return fetchFromSource()
}
```

### 3. Monitor Cache Performance
```go
stats := cacheService.GetStats()
if stats.HitRate < 70 {
    log.Warn("Low cache hit rate", "rate", stats.HitRate)
}
```

### 4. Use Structured Keys
```go
key := cacheService.(*cache.Service).GetCarbonIntensityKey(location)
```

## Support and Troubleshooting

If you encounter issues during migration:

1. **Check the logs** for cache initialization messages
2. **Verify Redis connectivity** using the health endpoint
3. **Review configuration** environment variables
4. **Test with curl** to verify API functionality
5. **Check cache statistics** to ensure proper operation

For detailed documentation, see [internal/cache/README.md](internal/cache/README.md).

## Migration Checklist

- [ ] Update imports to use new cache package
- [ ] Initialize cache service with dependency injection
- [ ] Update service integration to use cached wrappers
- [ ] Add cache middleware to Gin router
- [ ] Update error handling to use specific error types
- [ ] Configure environment variables for new settings
- [ ] Test cache functionality with health endpoints
- [ ] Verify API responses include cache status
- [ ] Monitor cache performance metrics
- [ ] Update documentation and deployment scripts

---

**Next Steps:** After successful migration, consider implementing additional features like:
- Cache warming strategies
- Custom TTL per endpoint
- Advanced monitoring and alerting
- Cache invalidation patterns
- Performance optimization based on metrics