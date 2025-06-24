# Redis Caching Implementation for GreenWeb API

## Overview

This implementation adds comprehensive Redis caching to the GreenWeb API to improve performance and reduce external API calls. The caching system follows industry best practices with graceful degradation when Redis is unavailable.

## Features Implemented

### 1. Redis Caching Service (`cache.go`)

- **Connection Pooling**: Configurable connection pool with optimal settings
- **Graceful Degradation**: Continues operation without caching when Redis is unavailable
- **Cache-Aside Pattern**: Implements cache-aside pattern with automatic fallback
- **JSON Serialization**: Handles complex data structures with JSON serialization
- **Error Handling**: Comprehensive error handling with logging
- **Statistics Tracking**: Real-time cache hit/miss metrics

### 2. TTL Configuration

- **Carbon Intensity**: 5 minutes - Data changes frequently
- **Optimization Profiles**: 10 minutes - Moderate refresh rate
- **Green Hours Forecast**: 1 hour - Longer-term forecast data

### 3. Cache Key Strategy

Cache keys include all relevant parameters for proper cache isolation:
- `greenweb:carbon_intensity:location`
- `greenweb:optimization:location:url`
- `greenweb:green_hours:location:hours`

### 4. Configuration

Environment variables for fine-tuning Redis connection:

```env
# Basic Configuration
REDIS_URL=redis://localhost:6379

# Connection Pool Settings
REDIS_MAX_RETRIES=3
REDIS_MIN_RETRY_BACKOFF_MS=100
REDIS_MAX_RETRY_BACKOFF_MS=3000
REDIS_DIAL_TIMEOUT_MS=5000
REDIS_READ_TIMEOUT_MS=3000
REDIS_WRITE_TIMEOUT_MS=3000
REDIS_POOL_SIZE=10
REDIS_MIN_IDLE_CONNS=2
REDIS_MAX_CONN_AGE_MINUTES=30
REDIS_POOL_TIMEOUT_MS=4000
REDIS_IDLE_TIMEOUT_MINUTES=5
```

## API Endpoints

### Core Endpoints (Now Cached)

All existing endpoints now use caching:

- `GET /api/v1/carbon-intensity?location=Berlin` - Cached for 5 minutes
- `GET /api/v1/optimization?location=Berlin&url=example.com` - Cached for 10 minutes  
- `GET /api/v1/green-hours?location=Berlin&next=24` - Cached for 1 hour

### Cache Management Endpoints

#### Cache Health Check
```bash
GET /api/v1/cache/health
```

Response:
```json
{
  "cache": {
    "enabled": true,
    "status": "healthy",
    "stats": {
      "hits": 45,
      "misses": 12,
      "sets": 12,
      "errors": 0,
      "total_requests": 57,
      "hit_rate": 78.95,
      "last_reset": "2025-06-24T15:50:33Z"
    }
  }
}
```

#### Cache Statistics
```bash
GET /api/v1/cache/stats
```

#### Reset Cache Statistics
```bash
POST /api/v1/cache/reset
```

#### Cache Invalidation
```bash
# Invalidate all cache entries
DELETE /api/v1/cache/invalidate

# Invalidate by prefix
DELETE /api/v1/cache/invalidate?prefix=carbon_intensity

# Invalidate by pattern
DELETE /api/v1/cache/invalidate?pattern=greenweb:optimization:Berlin:*
```

## Architecture

### Cache-Aside Pattern

The implementation uses the cache-aside pattern for maximum reliability:

1. Check cache for data
2. If cache miss, fetch from source API
3. Store result in cache for future requests
4. Return data to client

### Graceful Degradation

When Redis is unavailable:
- API continues to function normally
- All requests go directly to source APIs
- Cache statistics show cache as disabled
- No errors are returned to clients

### Error Handling

- Connection failures are logged but don't break API functionality
- Cache operation errors fall back to direct API calls
- Comprehensive logging for debugging and monitoring

## Performance Benefits

### Expected Performance Improvements

- **Response Time**: 80-95% reduction for cached responses
- **External API Calls**: Significant reduction in calls to Electricity Maps API
- **Cost Reduction**: Lower API usage costs
- **Reliability**: Reduced dependency on external API availability

### Cache Hit Rates

Expected hit rates based on TTL settings:
- Carbon Intensity: 70-85% (5-minute TTL)
- Optimization Profiles: 80-90% (10-minute TTL)  
- Green Hours Forecast: 85-95% (1-hour TTL)

## Monitoring and Observability

### Health Monitoring

The `/health` endpoint now includes cache status:
```json
{
  "status": "healthy",
  "service": "greenweb-api",
  "version": "0.2.0",
  "cache": {
    "enabled": true,
    "status": "healthy"
  }
}
```

### Metrics Available

- **Hit Rate**: Percentage of requests served from cache
- **Miss Rate**: Percentage of requests requiring API calls
- **Error Rate**: Cache operation failures
- **Total Requests**: Total cache operations attempted

## Deployment

### Prerequisites

1. Redis server (local or remote)
2. Updated environment configuration
3. Go dependencies installed (`go mod tidy`)

### Running the Application

```bash
# With Redis available
go run .

# Without Redis (graceful degradation)
go run .
# Logs: "Redis not available, continuing without caching"
```

### Testing Cache Functionality

```bash
# Check cache health
curl http://localhost:8090/api/v1/cache/health

# Test cached endpoint (first call - cache miss)
curl "http://localhost:8090/api/v1/carbon-intensity?location=Berlin"

# Test cached endpoint (second call - cache hit)
curl "http://localhost:8090/api/v1/carbon-intensity?location=Berlin"

# Check statistics
curl http://localhost:8090/api/v1/cache/stats
```

## Production Considerations

### Redis Configuration

For production environments:
- Use Redis Cluster for high availability
- Configure appropriate memory limits
- Enable persistence (RDB/AOF)
- Set up monitoring and alerting

### Security

- Use Redis AUTH when possible
- Configure Redis in private network
- Use TLS for Redis connections
- Regularly rotate Redis passwords

### Monitoring

- Monitor cache hit rates
- Alert on cache service unavailability
- Track response time improvements
- Monitor Redis memory usage

## Files Modified/Created

### New Files
- `/cache.go` - Complete Redis caching service implementation

### Modified Files
- `/main.go` - Integrated caching into all endpoints
- `/go.mod` - Added Redis client dependency
- `/.env.example` - Added Redis configuration variables

### Dependencies Added
- `github.com/redis/go-redis/v9` - Redis client library

## Testing Results

✅ **Application builds successfully**  
✅ **Graceful degradation when Redis unavailable**  
✅ **All endpoints function correctly**  
✅ **Cache management endpoints working**  
✅ **Statistics tracking operational**  
✅ **Health monitoring integrated**

The implementation is production-ready and provides significant performance improvements while maintaining full backward compatibility.