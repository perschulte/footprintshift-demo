package middleware

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// RateLimiter interface defines the rate limiting operations
type RateLimiter interface {
	Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, time.Duration, error)
	Reset(ctx context.Context, key string) error
}

// InMemoryRateLimiter implements rate limiting using in-memory storage
type InMemoryRateLimiter struct {
	windows map[string]*slidingWindow
	mutex   sync.RWMutex
}

// slidingWindow represents a sliding window for rate limiting
type slidingWindow struct {
	requests  []time.Time
	mutex     sync.RWMutex
	lastClean time.Time
}

// RedisRateLimiter implements rate limiting using Redis
type RedisRateLimiter struct {
	client *redis.Client
	prefix string
}

// NewRateLimit creates a new rate limiting middleware
func NewRateLimit(config RateLimitConfig, redisClient *redis.Client) gin.HandlerFunc {
	var limiter RateLimiter
	
	if config.UseRedis && redisClient != nil {
		limiter = &RedisRateLimiter{
			client: redisClient,
			prefix: config.RedisKeyPrefix,
		}
	} else {
		limiter = &InMemoryRateLimiter{
			windows: make(map[string]*slidingWindow),
		}
	}

	// Set default key generator if not provided
	if config.KeyGenerator == nil {
		config.KeyGenerator = defaultKeyGenerator
	}

	// Set default error handler if not provided
	if config.ErrorHandler == nil {
		config.ErrorHandler = defaultErrorHandler
	}

	// Set default limit reached handler if not provided
	if config.LimitReachedHandler == nil {
		config.LimitReachedHandler = defaultLimitReachedHandler
	}

	return gin.HandlerFunc(func(c *gin.Context) {
		// Skip rate limiting for certain conditions
		if shouldSkipRateLimit(c, config) {
			c.Next()
			return
		}

		// Get rate limit configuration for this endpoint
		limit, window := getRateLimitForEndpoint(c.Request.URL.Path, config)
		
		// Generate rate limit key
		key := config.KeyGenerator(c)
		
		// Check rate limit
		allowed, resetTime, err := limiter.Allow(c.Request.Context(), key, limit, window)
		if err != nil {
			config.ErrorHandler(c, err)
			return
		}

		// Set rate limit headers
		setRateLimitHeaders(c, limit, resetTime)

		if !allowed {
			config.LimitReachedHandler(c, resetTime)
			return
		}

		c.Next()
	})
}

// shouldSkipRateLimit determines if rate limiting should be skipped
func shouldSkipRateLimit(c *gin.Context, config RateLimitConfig) bool {
	statusCode := c.Writer.Status()
	
	// Skip successful requests if configured
	if config.SkipSuccessful && statusCode >= 200 && statusCode < 300 {
		return true
	}
	
	// Skip client errors if configured
	if config.SkipClientErrors && statusCode >= 400 && statusCode < 500 {
		return true
	}
	
	return false
}

// getRateLimitForEndpoint gets the rate limit configuration for a specific endpoint
func getRateLimitForEndpoint(path string, config RateLimitConfig) (int, time.Duration) {
	if endpointConfig, exists := config.EndpointLimits[path]; exists {
		// Use burst capacity if available, otherwise fall back to RPS
		limit := endpointConfig.Burst
		if limit == 0 {
			limit = endpointConfig.RPS
		}
		return limit, endpointConfig.Window
	}
	
	// Check for pattern matches
	for pattern, endpointConfig := range config.EndpointLimits {
		if strings.Contains(pattern, "*") {
			if matchesPattern(path, pattern) {
				limit := endpointConfig.Burst
				if limit == 0 {
					limit = endpointConfig.RPS
				}
				return limit, endpointConfig.Window
			}
		}
	}
	
	// Use burst capacity for default limit if available
	limit := config.DefaultBurst
	if limit == 0 {
		limit = config.DefaultRPS
	}
	return limit, config.WindowSize
}

// matchesPattern checks if a path matches a pattern with wildcards
func matchesPattern(path, pattern string) bool {
	if pattern == "*" {
		return true
	}
	
	// Simple wildcard matching (supports * at the end)
	if strings.HasSuffix(pattern, "*") {
		prefix := pattern[:len(pattern)-1]
		return strings.HasPrefix(path, prefix)
	}
	
	return path == pattern
}

// defaultKeyGenerator generates a rate limit key based on IP address
func defaultKeyGenerator(c *gin.Context) string {
	ip := getClientIP(c)
	return fmt.Sprintf("ratelimit:%s:%s", c.Request.URL.Path, ip)
}

// getClientIP extracts the client IP address from the request
func getClientIP(c *gin.Context) string {
	// Check X-Forwarded-For header
	xForwardedFor := c.GetHeader("X-Forwarded-For")
	if xForwardedFor != "" {
		ips := strings.Split(xForwardedFor, ",")
		return strings.TrimSpace(ips[0])
	}
	
	// Check X-Real-IP header
	xRealIP := c.GetHeader("X-Real-IP")
	if xRealIP != "" {
		return xRealIP
	}
	
	// Fall back to RemoteAddr
	ip, _, err := net.SplitHostPort(c.Request.RemoteAddr)
	if err != nil {
		return c.Request.RemoteAddr
	}
	
	return ip
}

// setRateLimitHeaders sets standard rate limit headers
func setRateLimitHeaders(c *gin.Context, limit int, resetTime time.Duration) {
	c.Header("X-RateLimit-Limit", strconv.Itoa(limit))
	c.Header("X-RateLimit-Reset", strconv.FormatInt(time.Now().Add(resetTime).Unix(), 10))
}

// defaultErrorHandler handles rate limiter errors
func defaultErrorHandler(c *gin.Context, err error) {
	c.JSON(http.StatusInternalServerError, gin.H{
		"error":   "rate limiter error",
		"message": err.Error(),
	})
	c.Abort()
}

// defaultLimitReachedHandler handles rate limit exceeded responses
func defaultLimitReachedHandler(c *gin.Context, resetTime time.Duration) {
	c.Header("Retry-After", strconv.FormatInt(int64(resetTime.Seconds()), 10))
	c.JSON(http.StatusTooManyRequests, gin.H{
		"error":     "rate limit exceeded",
		"message":   "too many requests, please try again later",
		"retry_after": int64(resetTime.Seconds()),
	})
	c.Abort()
}

// InMemoryRateLimiter implementation

// Allow checks if a request is allowed under the rate limit
func (r *InMemoryRateLimiter) Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, time.Duration, error) {
	r.mutex.RLock()
	w, exists := r.windows[key]
	r.mutex.RUnlock()
	
	if !exists {
		r.mutex.Lock()
		// Double-check pattern
		if w, exists = r.windows[key]; !exists {
			w = &slidingWindow{
				requests:  make([]time.Time, 0),
				lastClean: time.Now(),
			}
			r.windows[key] = w
		}
		r.mutex.Unlock()
	}
	
	w.mutex.Lock()
	defer w.mutex.Unlock()
	
	now := time.Now()
	cutoff := now.Add(-window)
	
	// Clean old requests
	r.cleanOldRequests(w, cutoff)
	
	// Check if we're within the limit
	if len(w.requests) >= limit {
		// Find the oldest request within the window
		if len(w.requests) > 0 {
			oldestTime := w.requests[0]
			resetTime := oldestTime.Add(window).Sub(now)
			if resetTime < 0 {
				resetTime = 0
			}
			return false, resetTime, nil
		}
		return false, window, nil
	}
	
	// Add current request
	w.requests = append(w.requests, now)
	
	return true, 0, nil
}

// Reset clears the rate limit for a key
func (r *InMemoryRateLimiter) Reset(ctx context.Context, key string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	
	delete(r.windows, key)
	return nil
}

// cleanOldRequests removes requests outside the sliding window
func (r *InMemoryRateLimiter) cleanOldRequests(w *slidingWindow, cutoff time.Time) {
	// Remove requests older than the cutoff
	validRequests := w.requests[:0]
	for _, req := range w.requests {
		if req.After(cutoff) {
			validRequests = append(validRequests, req)
		}
	}
	w.requests = validRequests
	w.lastClean = time.Now()
}

// RedisRateLimiter implementation

// Allow checks if a request is allowed using Redis
func (r *RedisRateLimiter) Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, time.Duration, error) {
	fullKey := r.prefix + key
	
	// Use Redis sliding window log algorithm
	now := time.Now()
	cutoff := now.Add(-window)
	
	pipe := r.client.Pipeline()
	
	// Remove expired entries
	pipe.ZRemRangeByScore(ctx, fullKey, "0", strconv.FormatInt(cutoff.UnixNano(), 10))
	
	// Count current entries
	countCmd := pipe.ZCard(ctx, fullKey)
	
	// Add current request
	pipe.ZAdd(ctx, fullKey, redis.Z{
		Score:  float64(now.UnixNano()),
		Member: fmt.Sprintf("%d", now.UnixNano()),
	})
	
	// Set expiration
	pipe.Expire(ctx, fullKey, window)
	
	_, err := pipe.Exec(ctx)
	if err != nil {
		return false, 0, err
	}
	
	count := countCmd.Val()
	
	if count >= int64(limit) {
		// Get the oldest entry to calculate reset time
		oldest, err := r.client.ZRange(ctx, fullKey, 0, 0).Result()
		if err != nil {
			return false, 0, err
		}
		
		if len(oldest) > 0 {
			oldestNano, _ := strconv.ParseInt(oldest[0], 10, 64)
			oldestTime := time.Unix(0, oldestNano)
			resetTime := oldestTime.Add(window).Sub(now)
			if resetTime < 0 {
				resetTime = 0
			}
			return false, resetTime, nil
		}
		
		return false, window, nil
	}
	
	return true, 0, nil
}

// Reset clears the rate limit for a key in Redis
func (r *RedisRateLimiter) Reset(ctx context.Context, key string) error {
	fullKey := r.prefix + key
	return r.client.Del(ctx, fullKey).Err()
}

// Advanced rate limiting functions

// PerUserRateLimit creates a rate limiter that limits per user ID
func PerUserRateLimit(config RateLimitConfig, redisClient *redis.Client, userIDHeader string) gin.HandlerFunc {
	config.KeyGenerator = func(c *gin.Context) string {
		userID := c.GetHeader(userIDHeader)
		if userID == "" {
			userID = "anonymous"
		}
		return fmt.Sprintf("user_ratelimit:%s:%s", c.Request.URL.Path, userID)
	}
	
	return NewRateLimit(config, redisClient)
}

// PerAPIKeyRateLimit creates a rate limiter that limits per API key
func PerAPIKeyRateLimit(config RateLimitConfig, redisClient *redis.Client, apiKeyHeader string) gin.HandlerFunc {
	config.KeyGenerator = func(c *gin.Context) string {
		apiKey := c.GetHeader(apiKeyHeader)
		if apiKey == "" {
			apiKey = getClientIP(c)
		}
		return fmt.Sprintf("apikey_ratelimit:%s:%s", c.Request.URL.Path, apiKey)
	}
	
	return NewRateLimit(config, redisClient)
}

// BurstRateLimit creates a rate limiter with burst capacity
func BurstRateLimit(rps, burst int, window time.Duration, redisClient *redis.Client) gin.HandlerFunc {
	config := RateLimitConfig{
		Enabled:      true,
		DefaultRPS:   rps,
		DefaultBurst: burst,
		WindowSize:   window,
		UseRedis:     redisClient != nil,
	}
	
	return NewRateLimit(config, redisClient)
}