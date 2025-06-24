# GreenWeb Middleware System

A comprehensive, modular middleware system for the GreenWeb API built on top of Gin framework. This system provides CORS handling, rate limiting, structured logging, and other essential middleware components with production-ready features.

## Features

- **CORS Middleware**: Configurable Cross-Origin Resource Sharing with support for wildcards and development mode
- **Rate Limiting**: Advanced rate limiting with sliding window algorithm, Redis support, and per-endpoint configuration
- **Structured Logging**: JSON-based logging with request tracing, performance metrics, and configurable verbosity
- **Request ID**: Automatic request ID generation and propagation
- **Security Headers**: Basic security headers for production deployments
- **Request Timeout**: Configurable request timeout middleware
- **Performance Monitoring**: Built-in performance metrics and slow request detection

## Quick Start

### Basic Usage

```go
package main

import (
    "github.com/gin-gonic/gin"
    "github.com/perschulte/greenweb-api/internal/middleware"
)

func main() {
    r := gin.New()
    
    // Use default middleware chain
    config := middleware.DefaultConfig()
    middlewares := middleware.Chain(config)
    for _, mw := range middlewares {
        r.Use(mw)
    }
    
    r.GET("/health", func(c *gin.Context) {
        c.JSON(200, gin.H{"status": "healthy"})
    })
    
    r.Run(":8080")
}
```

### Environment-Specific Configuration

```go
// Development setup
config := middleware.DevelopmentConfig()

// Production setup
allowedOrigins := []string{"https://yourdomain.com"}
config := middleware.ProductionConfig(allowedOrigins)

// Custom configuration
config := middleware.Config{
    CORS: middleware.CORSConfig{
        AllowedOrigins: []string{"*"},
        AllowedMethods: []string{"GET", "POST"},
    },
    RateLimit: middleware.RateLimitConfig{
        Enabled:    true,
        DefaultRPS: 100,
    },
    Logging: middleware.LoggingConfig{
        Enabled: true,
        Level:   slog.LevelInfo,
    },
}
```

## Middleware Components

### 1. CORS Middleware

Handles Cross-Origin Resource Sharing with extensive configuration options.

#### Features
- Configurable allowed origins, methods, and headers
- Wildcard origin support (`*`)
- Subdomain wildcard support (`*.example.com`)
- Development mode with automatic localhost/127.0.0.1 allowance
- Credential support
- Configurable preflight cache duration

#### Configuration

```go
corsConfig := middleware.CORSConfig{
    AllowedOrigins:     []string{"https://example.com", "*.example.com"},
    AllowedMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
    AllowedHeaders:     []string{"Content-Type", "Authorization", "X-API-Key"},
    ExposedHeaders:     []string{"X-Request-ID", "X-Total-Count"},
    AllowCredentials:   true,
    MaxAge:             24 * time.Hour,
    DevelopmentMode:    false,
    OptionsPassthrough: false,
}

r.Use(middleware.NewCORS(corsConfig))
```

#### Convenience Functions

```go
// Default permissive CORS
r.Use(middleware.DefaultCORS())

// Development CORS (very permissive)
r.Use(middleware.DevelopmentCORS())

// Production CORS (strict)
r.Use(middleware.StrictCORS([]string{"https://yourdomain.com"}))
```

### 2. Rate Limiting Middleware

Advanced rate limiting with multiple algorithms and storage backends.

#### Features
- Sliding window algorithm for accurate rate limiting
- Redis-backed distributed rate limiting
- In-memory fallback for single-instance deployments
- Per-endpoint rate limit configuration
- Custom key generation (IP, user ID, API key)
- Burst capacity support
- Graceful degradation when Redis is unavailable

#### Configuration

```go
rateLimitConfig := middleware.RateLimitConfig{
    Enabled:         true,
    DefaultRPS:      100,
    DefaultBurst:    200,
    WindowSize:      time.Minute,
    UseRedis:        true,
    RedisKeyPrefix:  "api:ratelimit:",
    SkipSuccessful:  false,
    SkipClientErrors: false,
    EndpointLimits: map[string]middleware.EndpointRate{
        "/api/v1/upload":    {RPS: 10, Burst: 20, Window: time.Minute},
        "/api/v1/search":    {RPS: 50, Burst: 100, Window: time.Minute},
        "/api/v1/expensive": {RPS: 5, Burst: 10, Window: time.Minute},
    },
}

r.Use(middleware.NewRateLimit(rateLimitConfig, redisClient))
```

#### Specialized Rate Limiters

```go
// Per-user rate limiting
r.Use(middleware.PerUserRateLimit(config, redisClient, "X-User-ID"))

// Per-API-key rate limiting
r.Use(middleware.PerAPIKeyRateLimit(config, redisClient, "X-API-Key"))

// Simple burst rate limiting
r.Use(middleware.BurstRateLimit(100, 200, time.Minute, redisClient))
```

#### Custom Key Generation

```go
config.KeyGenerator = func(c *gin.Context) string {
    userID := c.GetHeader("X-User-ID")
    if userID == "" {
        return "anonymous:" + getClientIP(c)
    }
    return "user:" + userID
}
```

### 3. Logging Middleware

Structured JSON logging with request tracing and performance metrics.

#### Features
- Structured JSON logging with slog
- Request/response logging with configurable detail level
- Request ID generation and propagation
- Performance metrics and slow request detection
- Configurable log levels and skip paths
- Request/response body logging (configurable)
- Error context logging

#### Configuration

```go
loggingConfig := middleware.LoggingConfig{
    Enabled:          true,
    Level:            slog.LevelInfo,
    SkipPaths:        []string{"/health", "/metrics"},
    RequestHeaders:   []string{"User-Agent", "X-Forwarded-For"},
    ResponseHeaders:  []string{"Content-Type", "X-Request-ID"},
    EnableBody:       false,
    MaxBodySize:      1024 * 1024, // 1MB
    SlowThreshold:    2 * time.Second,
    EnableMetrics:    true,
    RequestIDHeader:  "X-Request-ID",
}

r.Use(middleware.NewLogging(loggingConfig, logger))
```

#### Structured Logging Helpers

```go
// Log with request context
middleware.LogInfo(c, logger, "Processing request", 
    slog.String("user_id", userID),
    slog.Int("count", itemCount))

middleware.LogError(c, logger, err, "Database connection failed")
middleware.LogWarn(c, logger, "Rate limit approaching")
```

### 4. Security Middleware

Adds essential security headers for production deployments.

```go
r.Use(middleware.Secure())
```

Headers added:
- `X-Content-Type-Options: nosniff`
- `X-Frame-Options: DENY`
- `X-XSS-Protection: 1; mode=block`
- `Referrer-Policy: strict-origin-when-cross-origin`
- `Content-Security-Policy: default-src 'self'`

### 5. Request Timeout Middleware

Prevents long-running requests from hanging.

```go
r.Use(middleware.Timeout(30 * time.Second))
```

### 6. Performance Monitoring

Monitors request performance and logs slow requests.

```go
r.Use(middleware.PerformanceMonitor(logger, 1*time.Second))
```

## Configuration Management

### Environment-Based Configuration

The middleware system supports environment-based configuration:

```bash
# Environment variables
ENVIRONMENT=production
ALLOWED_ORIGINS=https://yourdomain.com,https://app.yourdomain.com
REDIS_URL=redis://localhost:6379
REDIS_PASSWORD=your_password
```

### Pre-built Configurations

```go
// Development (permissive, verbose logging)
config := middleware.DevelopmentConfig()

// Production (strict CORS, Redis rate limiting, structured logging)
config := middleware.ProductionConfig(allowedOrigins)

// Default (balanced settings)
config := middleware.DefaultConfig()
```

## Integration Examples

### Basic API Server

```go
func main() {
    logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
    
    r := gin.New()
    
    // Apply middleware
    config := middleware.DefaultConfig()
    config.Logger = logger
    
    middlewares := middleware.Chain(config)
    for _, mw := range middlewares {
        r.Use(mw)
    }
    
    // Routes
    r.GET("/api/v1/data", getData)
    r.POST("/api/v1/data", createData)
    
    r.Run(":8080")
}
```

### Microservice with Redis

```go
func main() {
    logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
    
    redisClient := redis.NewClient(&redis.Options{
        Addr: "localhost:6379",
    })
    
    r := gin.New()
    
    config := middleware.ProductionConfig([]string{"https://api.yourdomain.com"})
    config.Redis = redisClient
    config.Logger = logger
    
    // Custom rate limits for different endpoints
    config.RateLimit.EndpointLimits = map[string]middleware.EndpointRate{
        "/api/v1/heavy":  {RPS: 10, Burst: 20, Window: time.Minute},
        "/api/v1/light":  {RPS: 100, Burst: 200, Window: time.Minute},
        "/health":        {RPS: 1000, Burst: 2000, Window: time.Minute},
    }
    
    middlewares := middleware.Chain(config)
    for _, mw := range middlewares {
        r.Use(mw)
    }
    
    setupRoutes(r)
    r.Run(":8080")
}
```

### Custom Middleware Chain

```go
func customMiddlewareSetup() *gin.Engine {
    r := gin.New()
    
    // Request ID (always first)
    r.Use(middleware.RequestID())
    
    // CORS
    r.Use(middleware.StrictCORS([]string{"https://yourdomain.com"}))
    
    // Logging
    r.Use(middleware.NewLogging(middleware.LoggingConfig{
        Enabled:       true,
        Level:         slog.LevelInfo,
        SlowThreshold: 500 * time.Millisecond,
    }, logger))
    
    // Rate limiting
    r.Use(middleware.NewRateLimit(middleware.RateLimitConfig{
        Enabled:      true,
        DefaultRPS:   50,
        DefaultBurst: 100,
        WindowSize:   time.Minute,
    }, redisClient))
    
    // Security headers
    r.Use(middleware.Secure())
    
    // Timeout
    r.Use(middleware.Timeout(30 * time.Second))
    
    // Recovery (always last)
    r.Use(gin.Recovery())
    
    return r
}
```

## Testing

The middleware system is designed to be testable:

```go
func TestRateLimit(t *testing.T) {
    gin.SetMode(gin.TestMode)
    r := gin.New()
    
    config := middleware.RateLimitConfig{
        Enabled:      true,
        DefaultRPS:   2,
        DefaultBurst: 2,
        WindowSize:   time.Second,
        UseRedis:     false, // Use in-memory for testing
    }
    
    r.Use(middleware.NewRateLimit(config, nil))
    r.GET("/test", func(c *gin.Context) {
        c.JSON(200, gin.H{"ok": true})
    })
    
    // Test rate limiting
    for i := 0; i < 3; i++ {
        req := httptest.NewRequest("GET", "/test", nil)
        w := httptest.NewRecorder()
        r.ServeHTTP(w, req)
        
        if i < 2 {
            assert.Equal(t, 200, w.Code)
        } else {
            assert.Equal(t, 429, w.Code) // Rate limited
        }
    }
}
```

## Performance Considerations

- **Rate Limiting**: Redis-backed rate limiting adds minimal latency (~1-2ms per request)
- **Logging**: Body logging is disabled by default in production for performance
- **CORS**: Headers are pre-computed for optimal performance
- **Memory**: In-memory rate limiting uses efficient sliding windows with automatic cleanup

## Production Deployment

### Environment Variables

```bash
ENVIRONMENT=production
ALLOWED_ORIGINS=https://yourdomain.com,https://app.yourdomain.com
REDIS_URL=redis://redis-cluster:6379
REDIS_PASSWORD=secure_password
LOG_LEVEL=info
```

### Docker Configuration

```dockerfile
ENV ENVIRONMENT=production
ENV ALLOWED_ORIGINS=https://yourdomain.com
ENV REDIS_URL=redis://redis:6379
```

### Kubernetes ConfigMap

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: greenweb-config
data:
  ENVIRONMENT: "production"
  ALLOWED_ORIGINS: "https://yourdomain.com,https://app.yourdomain.com"
  REDIS_URL: "redis://redis-service:6379"
```

## Error Handling

The middleware system provides comprehensive error handling:

- **Rate Limiting**: Returns HTTP 429 with retry-after headers
- **CORS**: Returns HTTP 403 for forbidden origins
- **Timeout**: Returns HTTP 408 for request timeouts
- **Internal Errors**: Logged with full context, returns HTTP 500

## Monitoring and Observability

### Metrics

The middleware automatically logs performance metrics:

```json
{
  "type": "response",
  "method": "GET",
  "path": "/api/v1/data",
  "status_code": 200,
  "duration": "123ms",
  "slow_request": false,
  "request_id": "20240101120000-abc123"
}
```

### Health Checks

```go
r.Use(middleware.HealthCheck("/health"))
```

### Custom Metrics Integration

```go
// Integrate with Prometheus or other metrics systems
config.RateLimit.LimitReachedHandler = func(c *gin.Context, resetTime time.Duration) {
    rateLimitCounter.Inc()
    c.JSON(429, gin.H{"error": "rate limited"})
    c.Abort()
}
```

## Security Best Practices

1. **Always use HTTPS in production**
2. **Configure strict CORS origins**
3. **Use Redis for distributed rate limiting**
4. **Enable request timeout**
5. **Set appropriate rate limits per endpoint**
6. **Monitor and alert on rate limit violations**
7. **Use structured logging for security auditing**

## Troubleshooting

### Common Issues

1. **Redis Connection Issues**
   - Fallback to in-memory rate limiting
   - Check Redis connectivity and credentials

2. **CORS Issues**
   - Verify allowed origins configuration
   - Check for wildcard vs specific domain conflicts

3. **Rate Limiting Too Aggressive**
   - Adjust RPS and burst values
   - Consider per-endpoint limits

4. **Performance Issues**
   - Disable body logging in production
   - Use Redis for rate limiting in high-traffic scenarios

### Debug Mode

Enable debug logging to troubleshoot issues:

```go
config.Logging.Level = slog.LevelDebug
config.Logging.EnableBody = true
```

## Contributing

When contributing to the middleware system:

1. Follow Go best practices
2. Write comprehensive tests
3. Update documentation
4. Consider performance implications
5. Maintain backward compatibility