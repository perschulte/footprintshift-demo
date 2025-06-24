// Package config provides centralized configuration management for the GreenWeb API.
// It handles loading configuration from environment variables, validation,
// and provides sensible defaults for development environments.
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all configuration values for the GreenWeb API.
// It's designed to be thread-safe and immutable after initialization.
type Config struct {
	// Server configuration
	Server ServerConfig

	// External API configuration
	ElectricityMaps ElectricityMapsConfig

	// Cache configuration
	Redis RedisConfig

	// Application settings
	App AppConfig

	// Security settings
	Security SecurityConfig

	// Feature flags
	Features FeatureConfig

	// Internal state
	mu sync.RWMutex
}

// ServerConfig contains HTTP server configuration.
type ServerConfig struct {
	Host string // Server host address
	Port int    // Server port
	Env  string // Environment (development, staging, production)
}

// ElectricityMapsConfig contains Electricity Maps API configuration.
type ElectricityMapsConfig struct {
	APIKey  string // API key for Electricity Maps
	BaseURL string // Base URL for the API
}

// RedisConfig contains Redis connection and pool configuration.
type RedisConfig struct {
	URL                     string        // Redis connection URL
	MaxRetries              int           // Maximum number of retries for failed commands
	MinRetryBackoff         time.Duration // Minimum backoff between retries
	MaxRetryBackoff         time.Duration // Maximum backoff between retries
	DialTimeout             time.Duration // Timeout for establishing new connections
	ReadTimeout             time.Duration // Timeout for socket reads
	WriteTimeout            time.Duration // Timeout for socket writes
	PoolSize                int           // Maximum number of socket connections
	MinIdleConns            int           // Minimum number of idle connections
	MaxConnAge              time.Duration // Connection age at which client retires connection
	PoolTimeout             time.Duration // Amount of time client waits for connection
	IdleTimeout             time.Duration // Amount of time after which client closes idle connections
	IdleCheckFrequency      time.Duration // Frequency of idle checks made by idle connections reaper
}

// AppConfig contains general application configuration.
type AppConfig struct {
	LogLevel     string        // Logging level (debug, info, warn, error)
	CacheTTL     time.Duration // Default cache TTL
	RateLimit    RateLimitConfig
}

// RateLimitConfig contains rate limiting configuration.
type RateLimitConfig struct {
	RequestsPerMinute int           // Number of requests allowed per minute
	BurstSize         int           // Burst size for rate limiting
	CleanupInterval   time.Duration // Interval for cleaning up rate limit data
}

// SecurityConfig contains security-related configuration.
type SecurityConfig struct {
	AllowedOrigins []string // CORS allowed origins
}

// FeatureConfig contains feature flags.
type FeatureConfig struct {
	EnableDemoMode bool // Enable demo mode with mock data
}

// Load creates a new Config instance by loading values from environment variables.
// It automatically loads .env files if they exist and validates all required fields.
func Load() (*Config, error) {
	// Try to load .env file (ignore errors if file doesn't exist)
	_ = godotenv.Load()

	config := &Config{
		Server: ServerConfig{
			Host: getEnvString("HOST", "localhost"),
			Port: getEnvInt("PORT", 8090),
			Env:  getEnvString("ENVIRONMENT", "development"),
		},
		ElectricityMaps: ElectricityMapsConfig{
			APIKey:  getEnvString("ELECTRICITY_MAPS_API_KEY", ""),
			BaseURL: getEnvString("ELECTRICITY_MAPS_BASE_URL", "https://api.electricitymap.org/v3"),
		},
		Redis: RedisConfig{
			URL:                     getEnvString("REDIS_URL", "redis://localhost:6379"),
			MaxRetries:              getEnvInt("REDIS_MAX_RETRIES", 3),
			MinRetryBackoff:         time.Duration(getEnvInt("REDIS_MIN_RETRY_BACKOFF_MS", 100)) * time.Millisecond,
			MaxRetryBackoff:         time.Duration(getEnvInt("REDIS_MAX_RETRY_BACKOFF_MS", 3000)) * time.Millisecond,
			DialTimeout:             time.Duration(getEnvInt("REDIS_DIAL_TIMEOUT_MS", 5000)) * time.Millisecond,
			ReadTimeout:             time.Duration(getEnvInt("REDIS_READ_TIMEOUT_MS", 3000)) * time.Millisecond,
			WriteTimeout:            time.Duration(getEnvInt("REDIS_WRITE_TIMEOUT_MS", 3000)) * time.Millisecond,
			PoolSize:                getEnvInt("REDIS_POOL_SIZE", 10),
			MinIdleConns:            getEnvInt("REDIS_MIN_IDLE_CONNS", 2),
			MaxConnAge:              time.Duration(getEnvInt("REDIS_MAX_CONN_AGE_MINUTES", 30)) * time.Minute,
			PoolTimeout:             time.Duration(getEnvInt("REDIS_POOL_TIMEOUT_MS", 4000)) * time.Millisecond,
			IdleTimeout:             time.Duration(getEnvInt("REDIS_IDLE_TIMEOUT_MINUTES", 5)) * time.Minute,
			IdleCheckFrequency:      time.Duration(getEnvInt("REDIS_IDLE_CHECK_FREQUENCY_MINUTES", 1)) * time.Minute,
		},
		App: AppConfig{
			LogLevel: getEnvString("LOG_LEVEL", "info"),
			CacheTTL: time.Duration(getEnvInt("CACHE_TTL_SECONDS", 300)) * time.Second,
			RateLimit: RateLimitConfig{
				RequestsPerMinute: getEnvInt("RATE_LIMIT_RPM", 100),
				BurstSize:         getEnvInt("RATE_LIMIT_BURST", 10),
				CleanupInterval:   time.Duration(getEnvInt("RATE_LIMIT_CLEANUP_INTERVAL_MINUTES", 5)) * time.Minute,
			},
		},
		Security: SecurityConfig{
			AllowedOrigins: parseStringSlice(getEnvString("ALLOWED_ORIGINS", "http://localhost:3000,http://localhost:8090")),
		},
		Features: FeatureConfig{
			EnableDemoMode: getEnvBool("ENABLE_DEMO_MODE", true),
		},
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return config, nil
}

// Validate checks that all required configuration values are present and valid.
func (c *Config) Validate() error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var errors []string

	// Validate server configuration
	if c.Server.Port <= 0 || c.Server.Port > 65535 {
		errors = append(errors, "server port must be between 1 and 65535")
	}

	if c.Server.Env == "" {
		errors = append(errors, "ENVIRONMENT must be set")
	}

	// Validate environment-specific requirements
	if c.Server.Env == "production" {
		if c.ElectricityMaps.APIKey == "" {
			errors = append(errors, "ELECTRICITY_MAPS_API_KEY is required in production")
		}
		if c.Features.EnableDemoMode {
			errors = append(errors, "demo mode should not be enabled in production")
		}
	}

	// Validate Redis configuration
	if c.Redis.URL == "" {
		errors = append(errors, "REDIS_URL must be set")
	}

	if c.Redis.PoolSize <= 0 {
		errors = append(errors, "Redis pool size must be greater than 0")
	}

	if c.Redis.MinIdleConns < 0 {
		errors = append(errors, "Redis minimum idle connections cannot be negative")
	}

	if c.Redis.MinIdleConns > c.Redis.PoolSize {
		errors = append(errors, "Redis minimum idle connections cannot exceed pool size")
	}

	// Validate timeouts
	if c.Redis.DialTimeout <= 0 {
		errors = append(errors, "Redis dial timeout must be positive")
	}

	if c.Redis.ReadTimeout <= 0 {
		errors = append(errors, "Redis read timeout must be positive")
	}

	if c.Redis.WriteTimeout <= 0 {
		errors = append(errors, "Redis write timeout must be positive")
	}

	// Validate application configuration
	validLogLevels := map[string]bool{
		"debug": true, "info": true, "warn": true, "error": true,
	}
	if !validLogLevels[c.App.LogLevel] {
		errors = append(errors, "log level must be one of: debug, info, warn, error")
	}

	if c.App.CacheTTL <= 0 {
		errors = append(errors, "cache TTL must be positive")
	}

	// Validate rate limiting
	if c.App.RateLimit.RequestsPerMinute <= 0 {
		errors = append(errors, "rate limit requests per minute must be positive")
	}

	if c.App.RateLimit.BurstSize <= 0 {
		errors = append(errors, "rate limit burst size must be positive")
	}

	// Validate CORS origins
	if len(c.Security.AllowedOrigins) == 0 {
		errors = append(errors, "at least one allowed origin must be specified")
	}

	if len(errors) > 0 {
		return fmt.Errorf("validation errors: %s", strings.Join(errors, "; "))
	}

	return nil
}

// String returns a string representation of the configuration with sensitive data masked.
// This is safe for logging and debugging purposes.
func (c *Config) String() string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	apiKey := c.ElectricityMaps.APIKey
	if apiKey != "" {
		if len(apiKey) <= 8 {
			apiKey = "***"
		} else {
			apiKey = apiKey[:4] + "***" + apiKey[len(apiKey)-4:]
		}
	}

	redisURL := c.Redis.URL
	if redisURL != "" {
		// Mask password in Redis URL if present
		if strings.Contains(redisURL, "@") {
			parts := strings.Split(redisURL, "@")
			if len(parts) >= 2 {
				schemeParts := strings.Split(parts[0], "://")
				if len(schemeParts) == 2 {
					userParts := strings.Split(schemeParts[1], ":")
					if len(userParts) >= 2 {
						redisURL = schemeParts[0] + "://" + userParts[0] + ":***@" + strings.Join(parts[1:], "@")
					}
				}
			}
		}
	}

	return fmt.Sprintf(`Config{
  Server: {Host: %s, Port: %d, Env: %s}
  ElectricityMaps: {APIKey: %s, BaseURL: %s}
  Redis: {URL: %s, PoolSize: %d, MaxRetries: %d}
  App: {LogLevel: %s, CacheTTL: %s, RateLimit: %d rpm}
  Security: {AllowedOrigins: %v}
  Features: {EnableDemoMode: %t}
}`,
		c.Server.Host, c.Server.Port, c.Server.Env,
		apiKey, c.ElectricityMaps.BaseURL,
		redisURL, c.Redis.PoolSize, c.Redis.MaxRetries,
		c.App.LogLevel, c.App.CacheTTL, c.App.RateLimit.RequestsPerMinute,
		c.Security.AllowedOrigins,
		c.Features.EnableDemoMode,
	)
}

// IsProduction returns true if the application is running in production mode.
func (c *Config) IsProduction() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Server.Env == "production"
}

// IsDevelopment returns true if the application is running in development mode.
func (c *Config) IsDevelopment() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Server.Env == "development"
}

// GetServerAddress returns the full server address (host:port).
func (c *Config) GetServerAddress() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return fmt.Sprintf("%s:%d", c.Server.Host, c.Server.Port)
}

// Helper functions for environment variable parsing

// getEnvString returns the value of an environment variable or a default value.
func getEnvString(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvInt returns the integer value of an environment variable or a default value.
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

// getEnvBool returns the boolean value of an environment variable or a default value.
func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseBool(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

// parseStringSlice parses a comma-separated string into a slice of strings.
func parseStringSlice(value string) []string {
	if value == "" {
		return []string{}
	}
	
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	
	return result
}