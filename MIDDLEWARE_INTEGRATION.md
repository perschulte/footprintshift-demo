# GreenWeb Middleware System Integration

This document explains how the modular middleware system has been integrated into the GreenWeb API.

## Architecture Overview

The middleware system consists of modular components located in `internal/middleware/`:

```
internal/middleware/
├── middleware.go      # Core configuration and middleware chain
├── cors.go           # CORS handling middleware
├── ratelimit.go      # Rate limiting with Redis support
├── logging.go        # Structured logging middleware
├── middleware_test.go # Comprehensive test suite
└── README.md         # Detailed documentation
```

## Integration in main.go

The main application now uses the middleware system:

```go
// Setup middleware configuration
var middlewareConfig middleware.Config
if os.Getenv("ENVIRONMENT") == "production" {
    allowedOrigins := getAllowedOrigins()
    middlewareConfig = middleware.ProductionConfig(allowedOrigins)
} else {
    middlewareConfig = middleware.DevelopmentConfig()
}

// Set Redis client if available
middlewareConfig.Redis = redisClient
middlewareConfig.Logger = logger

// Apply middleware chain
middlewares := middleware.Chain(middlewareConfig)
for _, mw := range middlewares {
    r.Use(mw)
}
```

## Key Features Implemented

### 1. CORS Middleware
- **Before**: Simple, hardcoded CORS with `*` origin
- **After**: Configurable CORS with:
  - Environment-specific origins
  - Development mode with automatic localhost support
  - Wildcard subdomain support
  - Proper preflight handling
  - Credential support

### 2. Rate Limiting
- **New Feature**: Advanced rate limiting with:
  - Sliding window algorithm
  - Redis-backed distributed limiting
  - In-memory fallback
  - Per-endpoint configuration
  - Burst capacity support
  - Custom key generation (IP, user ID, API key)

### 3. Structured Logging
- **Before**: Basic Gin logging
- **After**: Comprehensive structured logging with:
  - JSON format with slog
  - Request ID generation and propagation
  - Performance metrics
  - Slow request detection
  - Error context logging
  - Configurable verbosity

### 4. Security Headers
- **New Feature**: Production-ready security headers:
  - X-Content-Type-Options
  - X-Frame-Options
  - X-XSS-Protection
  - Referrer-Policy
  - Content-Security-Policy

### 5. Request Timeout
- **New Feature**: Configurable request timeout to prevent hanging requests

## Configuration Options

### Environment Variables

The middleware system respects these environment variables:

```bash
# Core settings
ENVIRONMENT=production
ALLOWED_ORIGINS=https://yourdomain.com,https://app.yourdomain.com

# Redis (optional)
REDIS_URL=redis://localhost:6379
REDIS_PASSWORD=your_password

# Rate limiting
RATE_LIMIT_ENABLED=true
RATE_LIMIT_DEFAULT_RPS=100
RATE_LIMIT_DEFAULT_BURST=200
RATE_LIMIT_USE_REDIS=true

# Logging
LOG_LEVEL=info
LOG_ENABLE_BODY=false
LOG_MAX_BODY_SIZE_BYTES=1048576
LOG_SLOW_THRESHOLD_MS=2000

# Security
ENABLE_SECURITY_HEADERS=true
REQUEST_TIMEOUT_SECONDS=30
```

### Programmatic Configuration

```go
// Custom configuration
config := middleware.Config{
    CORS: middleware.CORSConfig{
        AllowedOrigins:   []string{"https://yourdomain.com"},
        AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE"},
        AllowCredentials: true,
    },
    RateLimit: middleware.RateLimitConfig{
        Enabled:      true,
        DefaultRPS:   100,
        DefaultBurst: 200,
        EndpointLimits: map[string]middleware.EndpointRate{
            "/api/v1/carbon-intensity": {RPS: 60, Burst: 120, Window: time.Minute},
            "/api/v1/optimization":     {RPS: 30, Burst: 60, Window: time.Minute},
        },
    },
    Logging: middleware.LoggingConfig{
        Enabled:       true,
        Level:         slog.LevelInfo,
        SlowThreshold: 2 * time.Second,
        EnableMetrics: true,
    },
}
```

## Endpoint-Specific Rate Limits

The system includes predefined rate limits for GreenWeb API endpoints:

```go
EndpointLimits: map[string]middleware.EndpointRate{
    "/api/v1/carbon-intensity": {RPS: 60, Burst: 120, Window: time.Minute},
    "/api/v1/optimization":     {RPS: 30, Burst: 60, Window: time.Minute},
    "/api/v1/green-hours":      {RPS: 20, Burst: 40, Window: time.Minute},
    "/health":                  {RPS: 1000, Burst: 2000, Window: time.Minute},
}
```

## Performance Impact

The middleware system is designed for minimal performance overhead:

- **CORS**: Headers pre-computed for optimal performance
- **Rate Limiting**: Redis operations add ~1-2ms per request
- **Logging**: Body logging disabled by default in production
- **Memory**: Efficient sliding windows with automatic cleanup

## Monitoring and Observability

### Structured Logs

```json
{
  "time": "2024-01-15T10:30:00Z",
  "level": "INFO",
  "msg": "HTTP 200 GET /api/v1/carbon-intensity",
  "type": "response",
  "method": "GET",
  "path": "/api/v1/carbon-intensity",
  "status_code": 200,
  "duration": "123ms",
  "request_id": "20240115103000-abc123",
  "ip": "192.168.1.1",
  "slow_request": false
}
```

### Rate Limit Headers

```
X-RateLimit-Limit: 100
X-RateLimit-Reset: 1705320600
X-Request-ID: 20240115103000-abc123
```

### Error Responses

```json
{
  "error": "rate limit exceeded",
  "message": "too many requests, please try again later",
  "retry_after": 45
}
```

## Testing

Comprehensive test suite covers:

- CORS functionality with various origins and methods
- Rate limiting with different configurations
- Structured logging with request tracing
- Security headers application
- Request timeout behavior
- Middleware chain integration

Run tests:
```bash
go test ./internal/middleware -v
```

## Migration Guide

### From Old CORS Implementation

**Before:**
```go
func corsMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
        // ...
    }
}
```

**After:**
```go
config := middleware.DefaultConfig()
middlewares := middleware.Chain(config)
for _, mw := range middlewares {
    r.Use(mw)
}
```

### Benefits of Migration

1. **Security**: Configurable CORS origins instead of wildcard
2. **Performance**: Built-in rate limiting prevents abuse
3. **Observability**: Structured logging with request tracing
4. **Maintainability**: Modular, testable middleware components
5. **Production-Ready**: Security headers, timeouts, and Redis support

## Production Deployment

### Docker Configuration

```dockerfile
ENV ENVIRONMENT=production
ENV ALLOWED_ORIGINS=https://yourdomain.com
ENV REDIS_URL=redis://redis:6379
ENV LOG_LEVEL=info
```

### Kubernetes ConfigMap

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: greenweb-middleware-config
data:
  ENVIRONMENT: "production"
  ALLOWED_ORIGINS: "https://yourdomain.com,https://app.yourdomain.com"
  REDIS_URL: "redis://redis-service:6379"
  RATE_LIMIT_DEFAULT_RPS: "50"
  LOG_LEVEL: "info"
```

## Troubleshooting

### Common Issues

1. **Redis Connection Failures**
   - System gracefully falls back to in-memory rate limiting
   - Check Redis connectivity and credentials

2. **CORS Rejections**
   - Verify `ALLOWED_ORIGINS` environment variable
   - Check for wildcard vs specific domain conflicts

3. **Rate Limiting Too Aggressive**
   - Adjust `RATE_LIMIT_DEFAULT_RPS` and `RATE_LIMIT_DEFAULT_BURST`
   - Configure endpoint-specific limits

### Debug Mode

Enable debug logging:
```bash
export LOG_LEVEL=debug
```

## Future Enhancements

Potential improvements for the middleware system:

1. **Metrics Integration**: Prometheus metrics for monitoring
2. **Circuit Breaker**: Automatic failover for downstream services  
3. **Request Caching**: Cache responses for GET requests
4. **API Versioning**: Version-aware middleware configuration
5. **Webhook Signatures**: Request signature validation
6. **IP Whitelisting**: Allow/deny lists for specific IPs

## Examples

See the `examples/` directory for:

- `middleware_usage.go`: Comprehensive usage examples
- `middleware_demo.go`: Interactive demo server

Run the demo:
```bash
go run examples/middleware_demo.go
```

## Documentation

Full documentation available in:
- `internal/middleware/README.md`: Detailed middleware documentation
- `examples/middleware_usage.go`: Code examples
- This file: Integration guide