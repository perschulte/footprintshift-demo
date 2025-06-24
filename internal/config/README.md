# Configuration Management

This package provides centralized configuration management for the GreenWeb API. It handles loading configuration from environment variables, validation, and provides sensible defaults for development environments.

## Features

- **Environment Variable Loading**: Automatically loads configuration from environment variables
- **Validation**: Comprehensive validation of all configuration values
- **Sensible Defaults**: Provides development-friendly defaults
- **Security**: Masks sensitive data in logs and string representations
- **Thread Safety**: All operations are thread-safe
- **Environment Detection**: Built-in helpers for production/development detection
- **Comprehensive Coverage**: Supports all aspects of the application (server, Redis, APIs, etc.)

## Usage

### Basic Usage

```go
package main

import (
    "log"
    "github.com/perschulte/greenweb-api/internal/config"
)

func main() {
    // Load configuration from environment variables
    cfg, err := config.Load()
    if err != nil {
        log.Fatalf("Failed to load configuration: %v", err)
    }

    // Use configuration throughout your application
    fmt.Printf("Starting server on %s\n", cfg.GetServerAddress())
    
    // Safe logging (sensitive data is masked)
    log.Printf("Configuration: %s", cfg)
}
```

### Integration with Redis

```go
import "github.com/redis/go-redis/v9"

func setupRedis(cfg *config.Config) *redis.Client {
    return redis.NewClient(&redis.Options{
        Addr:            cfg.Redis.URL[8:], // Remove redis:// prefix
        MaxRetries:      cfg.Redis.MaxRetries,
        MinRetryBackoff: cfg.Redis.MinRetryBackoff,
        MaxRetryBackoff: cfg.Redis.MaxRetryBackoff,
        DialTimeout:     cfg.Redis.DialTimeout,
        ReadTimeout:     cfg.Redis.ReadTimeout,
        WriteTimeout:    cfg.Redis.WriteTimeout,
        PoolSize:        cfg.Redis.PoolSize,
        MinIdleConns:    cfg.Redis.MinIdleConns,
        MaxConnAge:      cfg.Redis.MaxConnAge,
        PoolTimeout:     cfg.Redis.PoolTimeout,
        IdleTimeout:     cfg.Redis.IdleTimeout,
    })
}
```

### Environment-Specific Behavior

```go
if cfg.IsProduction() {
    // Production-specific setup
    setupProductionLogging()
    enableSecurityFeatures()
} else if cfg.IsDevelopment() {
    // Development-specific setup
    enableDebugEndpoints()
    setupVerboseLogging()
}
```

## Configuration Structure

The configuration is organized into logical groups:

### Server Configuration
- `HOST`: Server host address (default: "localhost")
- `PORT`: Server port (default: 8090)
- `ENVIRONMENT`: Environment (development/staging/production)

### Electricity Maps API
- `ELECTRICITY_MAPS_API_KEY`: API key for Electricity Maps
- `ELECTRICITY_MAPS_BASE_URL`: Base URL for the API (default: "https://api.electricitymap.org/v3")

### Redis Configuration
- `REDIS_URL`: Redis connection URL (default: "redis://localhost:6379")
- `REDIS_MAX_RETRIES`: Maximum number of retries (default: 3)
- `REDIS_MIN_RETRY_BACKOFF_MS`: Minimum backoff between retries (default: 100)
- `REDIS_MAX_RETRY_BACKOFF_MS`: Maximum backoff between retries (default: 3000)
- `REDIS_DIAL_TIMEOUT_MS`: Timeout for establishing connections (default: 5000)
- `REDIS_READ_TIMEOUT_MS`: Timeout for reads (default: 3000)
- `REDIS_WRITE_TIMEOUT_MS`: Timeout for writes (default: 3000)
- `REDIS_POOL_SIZE`: Maximum number of connections (default: 10)
- `REDIS_MIN_IDLE_CONNS`: Minimum idle connections (default: 2)
- `REDIS_MAX_CONN_AGE_MINUTES`: Connection age limit (default: 30)
- `REDIS_POOL_TIMEOUT_MS`: Pool timeout (default: 4000)
- `REDIS_IDLE_TIMEOUT_MINUTES`: Idle timeout (default: 5)
- `REDIS_IDLE_CHECK_FREQUENCY_MINUTES`: Idle check frequency (default: 1)

### Application Settings
- `LOG_LEVEL`: Logging level (debug/info/warn/error, default: "info")
- `CACHE_TTL_SECONDS`: Default cache TTL (default: 300)
- `RATE_LIMIT_RPM`: Rate limit requests per minute (default: 100)
- `RATE_LIMIT_BURST`: Rate limit burst size (default: 10)
- `RATE_LIMIT_CLEANUP_INTERVAL_MINUTES`: Rate limit cleanup interval (default: 5)

### Security Settings
- `ALLOWED_ORIGINS`: CORS allowed origins (comma-separated, default: "http://localhost:3000,http://localhost:8090")

### Feature Flags
- `ENABLE_DEMO_MODE`: Enable demo mode with mock data (default: true)

## Environment Files

The configuration automatically loads `.env` files if they exist. Create a `.env` file in your project root:

```bash
# Server
PORT=8090
ENVIRONMENT=development

# Electricity Maps API
ELECTRICITY_MAPS_API_KEY=your_api_key_here

# Redis
REDIS_URL=redis://localhost:6379
REDIS_POOL_SIZE=20

# Application
LOG_LEVEL=debug
CACHE_TTL_SECONDS=600

# Security
ALLOWED_ORIGINS=http://localhost:3000,https://yourdomain.com

# Features
ENABLE_DEMO_MODE=true
```

## Validation

The configuration includes comprehensive validation:

- **Required Fields**: Ensures all required fields are present
- **Value Ranges**: Validates ports, timeouts, and other numeric values
- **Environment-Specific**: Production environments require API keys
- **Logical Consistency**: Ensures related values make sense together

### Validation Examples

```go
// This will fail validation
cfg := &Config{
    Server: ServerConfig{Port: -1}, // Invalid port
}
err := cfg.Validate() // Returns error about invalid port

// Production environment validation
cfg := &Config{
    Server: ServerConfig{Env: "production"},
    ElectricityMaps: ElectricityMapsConfig{APIKey: ""}, // Missing API key
}
err := cfg.Validate() // Returns error about missing API key in production
```

## Security Features

- **Sensitive Data Masking**: API keys and passwords are automatically masked in logs
- **Safe String Representation**: The `String()` method never exposes sensitive data
- **Validation**: Prevents unsafe configurations from being used

## Thread Safety

All configuration operations are thread-safe. The configuration is immutable after loading, and all read operations use appropriate locking.

## Testing

The package includes comprehensive tests covering:
- Default value loading
- Custom environment variable parsing
- Validation scenarios
- Security (sensitive data masking)
- Helper methods

Run tests with:
```bash
go test ./internal/config -v
```

## Error Handling

The configuration system provides clear, actionable error messages:

```go
cfg, err := config.Load()
if err != nil {
    // Error messages are detailed and helpful
    log.Printf("Configuration error: %v", err)
    // Example: "validation errors: server port must be between 1 and 65535; ELECTRICITY_MAPS_API_KEY is required in production"
}
```

## Best Practices

1. **Load Early**: Load configuration at application startup
2. **Validate Always**: Always check for validation errors
3. **Use Helpers**: Use `IsProduction()`, `IsDevelopment()` for environment detection
4. **Log Safely**: Use `cfg.String()` for logging (automatically masks sensitive data)
5. **Handle Errors**: Provide clear error messages for configuration issues
6. **Environment Files**: Use `.env` files for local development
7. **Production Ready**: Ensure production environments have all required values

## Integration Examples

See `example_usage.go` for complete integration examples with:
- Gin web framework
- Redis client setup
- Environment-specific behavior
- Error handling patterns