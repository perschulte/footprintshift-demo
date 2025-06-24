package middleware

import (
	"context"
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// Config holds all middleware configuration
type Config struct {
	CORS      CORSConfig
	RateLimit RateLimitConfig
	Logging   LoggingConfig
	Redis     *redis.Client // Optional Redis client for distributed features
	Logger    *slog.Logger
}

// CORSConfig configures CORS middleware
type CORSConfig struct {
	AllowedOrigins     []string      `json:"allowed_origins"`
	AllowedMethods     []string      `json:"allowed_methods"`
	AllowedHeaders     []string      `json:"allowed_headers"`
	ExposedHeaders     []string      `json:"exposed_headers"`
	AllowCredentials   bool          `json:"allow_credentials"`
	MaxAge             time.Duration `json:"max_age"`
	DevelopmentMode    bool          `json:"development_mode"`
	OptionsPassthrough bool          `json:"options_passthrough"`
}

// RateLimitConfig configures rate limiting middleware
type RateLimitConfig struct {
	Enabled            bool                    `json:"enabled"`
	DefaultRPS         int                     `json:"default_rps"`
	DefaultBurst       int                     `json:"default_burst"`
	WindowSize         time.Duration           `json:"window_size"`
	EndpointLimits     map[string]EndpointRate `json:"endpoint_limits"`
	UseRedis           bool                    `json:"use_redis"`
	RedisKeyPrefix     string                  `json:"redis_key_prefix"`
	SkipSuccessful     bool                    `json:"skip_successful"`
	SkipClientErrors   bool                    `json:"skip_client_errors"`
	TrustedProxies     []string                `json:"trusted_proxies"`
	KeyGenerator       func(*gin.Context) string
	ErrorHandler       func(*gin.Context, error)
	LimitReachedHandler func(*gin.Context, time.Duration)
}

// EndpointRate defines rate limiting for specific endpoints
type EndpointRate struct {
	RPS    int           `json:"rps"`
	Burst  int           `json:"burst"`
	Window time.Duration `json:"window"`
}

// LoggingConfig configures logging middleware
type LoggingConfig struct {
	Enabled          bool          `json:"enabled"`
	Level            slog.Level    `json:"level"`
	SkipPaths        []string      `json:"skip_paths"`
	RequestHeaders   []string      `json:"request_headers"`
	ResponseHeaders  []string      `json:"response_headers"`
	EnableBody       bool          `json:"enable_body"`
	MaxBodySize      int64         `json:"max_body_size"`
	SlowThreshold    time.Duration `json:"slow_threshold"`
	EnableMetrics    bool          `json:"enable_metrics"`
	RequestIDHeader  string        `json:"request_id_header"`
	DisableColor     bool          `json:"disable_color"`
	DisableTimestamp bool          `json:"disable_timestamp"`
}

// DefaultConfig returns a default middleware configuration
func DefaultConfig() Config {
	return Config{
		CORS: CORSConfig{
			AllowedOrigins:     []string{"*"},
			AllowedMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "HEAD", "PATCH"},
			AllowedHeaders:     []string{"Content-Type", "Authorization", "X-Requested-With", "Accept", "Origin"},
			ExposedHeaders:     []string{"Content-Length", "X-Request-ID"},
			AllowCredentials:   false,
			MaxAge:             12 * time.Hour,
			DevelopmentMode:    false,
			OptionsPassthrough: false,
		},
		RateLimit: RateLimitConfig{
			Enabled:         true,
			DefaultRPS:      100,
			DefaultBurst:    200,
			WindowSize:      time.Minute,
			UseRedis:        false,
			RedisKeyPrefix:  "greenweb:ratelimit:",
			SkipSuccessful:  false,
			SkipClientErrors: false,
			EndpointLimits: map[string]EndpointRate{
				"/api/v1/carbon-intensity": {RPS: 60, Burst: 120, Window: time.Minute},
				"/api/v1/optimization":     {RPS: 30, Burst: 60, Window: time.Minute},
				"/api/v1/green-hours":      {RPS: 20, Burst: 40, Window: time.Minute},
				"/health":                  {RPS: 1000, Burst: 2000, Window: time.Minute},
			},
		},
		Logging: LoggingConfig{
			Enabled:          true,
			Level:            slog.LevelInfo,
			SkipPaths:        []string{"/health", "/favicon.ico"},
			RequestHeaders:   []string{"User-Agent", "X-Forwarded-For", "X-Real-IP"},
			ResponseHeaders:  []string{"Content-Type", "X-Request-ID"},
			EnableBody:       false,
			MaxBodySize:      1024 * 1024, // 1MB
			SlowThreshold:    2 * time.Second,
			EnableMetrics:    true,
			RequestIDHeader:  "X-Request-ID",
			DisableColor:     false,
			DisableTimestamp: false,
		},
	}
}

// DevelopmentConfig returns a development-friendly configuration
func DevelopmentConfig() Config {
	config := DefaultConfig()
	
	// More permissive CORS for development
	config.CORS.DevelopmentMode = true
	config.CORS.AllowedOrigins = []string{"*"}
	config.CORS.AllowCredentials = false
	
	// Higher rate limits for development
	config.RateLimit.DefaultRPS = 1000
	config.RateLimit.DefaultBurst = 2000
	
	// More verbose logging
	config.Logging.Level = slog.LevelDebug
	config.Logging.EnableBody = true
	config.Logging.MaxBodySize = 10 * 1024 * 1024 // 10MB
	config.Logging.SlowThreshold = 1 * time.Second
	
	return config
}

// ProductionConfig returns a production-ready configuration
func ProductionConfig(allowedOrigins []string) Config {
	config := DefaultConfig()
	
	// Strict CORS for production
	config.CORS.DevelopmentMode = false
	config.CORS.AllowedOrigins = allowedOrigins
	config.CORS.AllowCredentials = true
	
	// Conservative rate limits
	config.RateLimit.DefaultRPS = 50
	config.RateLimit.DefaultBurst = 100
	config.RateLimit.UseRedis = true // Recommended for production
	
	// Production logging
	config.Logging.Level = slog.LevelInfo
	config.Logging.EnableBody = false
	config.Logging.DisableColor = true
	config.Logging.SlowThreshold = 500 * time.Millisecond
	
	return config
}

// Chain creates a middleware chain from the configuration
func Chain(config Config) []gin.HandlerFunc {
	var middlewares []gin.HandlerFunc
	
	// Request ID middleware (always first)
	middlewares = append(middlewares, RequestID())
	
	// CORS middleware
	middlewares = append(middlewares, NewCORS(config.CORS))
	
	// Logging middleware
	if config.Logging.Enabled {
		middlewares = append(middlewares, NewLogging(config.Logging, config.Logger))
	}
	
	// Rate limiting middleware
	if config.RateLimit.Enabled {
		middlewares = append(middlewares, NewRateLimit(config.RateLimit, config.Redis))
	}
	
	// Recovery middleware (always last)
	middlewares = append(middlewares, gin.Recovery())
	
	return middlewares
}

// RequestID generates a unique request ID for each request
func RequestID() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
		}
		
		c.Header("X-Request-ID", requestID)
		c.Set("request_id", requestID)
		c.Next()
	})
}

// generateRequestID creates a unique request ID
func generateRequestID() string {
	// Using timestamp + random suffix for uniqueness
	now := time.Now()
	return now.Format("20060102150405") + "-" + randomString(8)
}

// randomString generates a random string of given length
func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
}

// HealthCheck is a simple middleware that responds to health check requests
func HealthCheck(path string) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		if c.Request.URL.Path == path {
			c.JSON(200, gin.H{
				"status":    "healthy",
				"timestamp": time.Now(),
				"service":   "greenweb-api",
			})
			c.Abort()
			return
		}
		c.Next()
	})
}

// Timeout middleware adds request timeout
func Timeout(timeout time.Duration) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()
		
		c.Request = c.Request.WithContext(ctx)
		
		done := make(chan bool, 1)
		go func() {
			c.Next()
			done <- true
		}()
		
		select {
		case <-done:
			return
		case <-ctx.Done():
			c.JSON(408, gin.H{
				"error":   "request timeout",
				"timeout": timeout.String(),
			})
			c.Abort()
			return
		}
	})
}

// Secure adds basic security headers
func Secure() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Content-Security-Policy", "default-src 'self'")
		c.Next()
	})
}