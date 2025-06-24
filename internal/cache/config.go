package cache

import (
	"os"
	"strconv"
	"time"
)

// Config holds cache configuration
type Config struct {
	// Redis connection settings
	URL                  string        `json:"url"`
	MaxRetries          int           `json:"max_retries"`
	MinRetryBackoff     time.Duration `json:"min_retry_backoff"`
	MaxRetryBackoff     time.Duration `json:"max_retry_backoff"`
	DialTimeout         time.Duration `json:"dial_timeout"`
	ReadTimeout         time.Duration `json:"read_timeout"`
	WriteTimeout        time.Duration `json:"write_timeout"`
	
	// Connection pool settings
	PoolSize            int           `json:"pool_size"`
	MinIdleConns        int           `json:"min_idle_conns"`
	MaxConnAge          time.Duration `json:"max_conn_age"`
	PoolTimeout         time.Duration `json:"pool_timeout"`
	IdleTimeout         time.Duration `json:"idle_timeout"`
	IdleCheckFrequency  time.Duration `json:"idle_check_frequency"`
	
	// Cache behavior settings
	KeyPrefix           string        `json:"key_prefix"`
	DefaultTTL          time.Duration `json:"default_ttl"`
	EnableHealthCheck   bool          `json:"enable_health_check"`
	HealthCheckInterval time.Duration `json:"health_check_interval"`
	
	// Fallback behavior
	EnableFallback      bool          `json:"enable_fallback"`
	GracefulDegradation bool          `json:"graceful_degradation"`
}

// DefaultConfig returns a configuration with sensible defaults
func DefaultConfig() *Config {
	return &Config{
		URL:                 "redis://localhost:6379",
		MaxRetries:          3,
		MinRetryBackoff:     100 * time.Millisecond,
		MaxRetryBackoff:     1000 * time.Millisecond,
		DialTimeout:         5 * time.Second,
		ReadTimeout:         3 * time.Second,
		WriteTimeout:        3 * time.Second,
		
		PoolSize:            10,
		MinIdleConns:        5,
		MaxConnAge:          60 * time.Minute,
		PoolTimeout:         4 * time.Second,
		IdleTimeout:         5 * time.Minute,
		IdleCheckFrequency:  1 * time.Minute,
		
		KeyPrefix:           "greenweb",
		DefaultTTL:          15 * time.Minute,
		EnableHealthCheck:   true,
		HealthCheckInterval: 30 * time.Second,
		
		EnableFallback:      true,
		GracefulDegradation: true,
	}
}

// LoadFromEnv loads configuration from environment variables, falling back to defaults
func LoadFromEnv() *Config {
	config := DefaultConfig()
	
	// Redis connection settings
	if url := getEnv("REDIS_URL", ""); url != "" {
		config.URL = url
	}
	
	config.MaxRetries = getEnvInt("REDIS_MAX_RETRIES", config.MaxRetries)
	config.MinRetryBackoff = time.Duration(getEnvInt("REDIS_MIN_RETRY_BACKOFF_MS", 
		int(config.MinRetryBackoff.Milliseconds()))) * time.Millisecond
	config.MaxRetryBackoff = time.Duration(getEnvInt("REDIS_MAX_RETRY_BACKOFF_MS", 
		int(config.MaxRetryBackoff.Milliseconds()))) * time.Millisecond
	config.DialTimeout = time.Duration(getEnvInt("REDIS_DIAL_TIMEOUT_MS", 
		int(config.DialTimeout.Milliseconds()))) * time.Millisecond
	config.ReadTimeout = time.Duration(getEnvInt("REDIS_READ_TIMEOUT_MS", 
		int(config.ReadTimeout.Milliseconds()))) * time.Millisecond
	config.WriteTimeout = time.Duration(getEnvInt("REDIS_WRITE_TIMEOUT_MS", 
		int(config.WriteTimeout.Milliseconds()))) * time.Millisecond
	
	// Connection pool settings
	config.PoolSize = getEnvInt("REDIS_POOL_SIZE", config.PoolSize)
	config.MinIdleConns = getEnvInt("REDIS_MIN_IDLE_CONNS", config.MinIdleConns)
	config.MaxConnAge = time.Duration(getEnvInt("REDIS_MAX_CONN_AGE_MINUTES", 
		int(config.MaxConnAge.Minutes()))) * time.Minute
	config.PoolTimeout = time.Duration(getEnvInt("REDIS_POOL_TIMEOUT_MS", 
		int(config.PoolTimeout.Milliseconds()))) * time.Millisecond
	config.IdleTimeout = time.Duration(getEnvInt("REDIS_IDLE_TIMEOUT_MINUTES", 
		int(config.IdleTimeout.Minutes()))) * time.Minute
	config.IdleCheckFrequency = time.Duration(getEnvInt("REDIS_IDLE_CHECK_FREQ_MINUTES", 
		int(config.IdleCheckFrequency.Minutes()))) * time.Minute
	
	// Cache behavior settings
	if prefix := getEnv("CACHE_KEY_PREFIX", ""); prefix != "" {
		config.KeyPrefix = prefix
	}
	
	config.DefaultTTL = time.Duration(getEnvInt("CACHE_DEFAULT_TTL_MINUTES", 
		int(config.DefaultTTL.Minutes()))) * time.Minute
	config.EnableHealthCheck = getEnvBool("CACHE_ENABLE_HEALTH_CHECK", config.EnableHealthCheck)
	config.HealthCheckInterval = time.Duration(getEnvInt("CACHE_HEALTH_CHECK_INTERVAL_SECONDS", 
		int(config.HealthCheckInterval.Seconds()))) * time.Second
	
	// Fallback behavior
	config.EnableFallback = getEnvBool("CACHE_ENABLE_FALLBACK", config.EnableFallback)
	config.GracefulDegradation = getEnvBool("CACHE_GRACEFUL_DEGRADATION", config.GracefulDegradation)
	
	return config
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.URL == "" {
		return NewCacheError(ErrorTypeConnection, "Redis URL is required", nil)
	}
	
	if c.MaxRetries < 0 {
		return NewCacheError(ErrorTypeConnection, "MaxRetries cannot be negative", nil)
	}
	
	if c.PoolSize <= 0 {
		return NewCacheError(ErrorTypeConnection, "PoolSize must be positive", nil)
	}
	
	if c.MinIdleConns < 0 {
		return NewCacheError(ErrorTypeConnection, "MinIdleConns cannot be negative", nil)
	}
	
	if c.MinIdleConns > c.PoolSize {
		return NewCacheError(ErrorTypeConnection, "MinIdleConns cannot exceed PoolSize", nil)
	}
	
	if c.KeyPrefix == "" {
		c.KeyPrefix = "greenweb" // Set default if empty
	}
	
	return nil
}

// GenerateKey creates a cache key with consistent formatting using the configured prefix
func (c *Config) GenerateKey(prefix string, parts ...string) string {
	key := c.KeyPrefix + ":" + prefix
	for _, part := range parts {
		if part != "" {
			key += ":" + part
		}
	}
	return key
}

// Utility functions for environment variable parsing
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}