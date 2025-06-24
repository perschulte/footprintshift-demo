package cache

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// ManagementHandler provides HTTP endpoints for cache management
type ManagementHandler struct {
	cache Cacher
}

// NewManagementHandler creates a new cache management handler
func NewManagementHandler(cache Cacher) *ManagementHandler {
	return &ManagementHandler{
		cache: cache,
	}
}

// RegisterRoutes registers cache management routes
func (h *ManagementHandler) RegisterRoutes(router *gin.Engine) {
	cacheGroup := router.Group("/cache")
	{
		cacheGroup.GET("/stats", h.GetStats)
		cacheGroup.POST("/stats/reset", h.ResetStats)
		cacheGroup.GET("/health", h.GetHealth)
		cacheGroup.DELETE("/keys/:key", h.DeleteKey)
		cacheGroup.GET("/status", h.GetStatus)
	}
}

// GetStats returns cache statistics
func (h *ManagementHandler) GetStats(c *gin.Context) {
	stats := h.cache.GetStats()
	c.JSON(http.StatusOK, gin.H{
		"cache_stats": stats,
		"timestamp":   time.Now(),
	})
}

// ResetStats resets cache statistics
func (h *ManagementHandler) ResetStats(c *gin.Context) {
	h.cache.ResetStats()
	c.JSON(http.StatusOK, gin.H{
		"message":   "Cache statistics reset successfully",
		"timestamp": time.Now(),
	})
}

// GetHealth returns cache health status
func (h *ManagementHandler) GetHealth(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	health := h.cache.Health(ctx)
	
	statusCode := http.StatusOK
	if !health.Healthy {
		statusCode = http.StatusServiceUnavailable
	}

	c.JSON(statusCode, gin.H{
		"cache_health": health,
		"timestamp":    time.Now(),
	})
}

// DeleteKey deletes a specific cache key
func (h *ManagementHandler) DeleteKey(c *gin.Context) {
	key := c.Param("key")
	if key == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "key parameter is required",
		})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	if err := h.cache.Delete(ctx, key); err != nil {
		if IsCacheDisabled(err) {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error": "cache is disabled",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to delete cache key",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Cache key deleted successfully",
		"key":       key,
		"timestamp": time.Now(),
	})
}

// GetStatus returns overall cache status
func (h *ManagementHandler) GetStatus(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	stats := h.cache.GetStats()
	health := h.cache.Health(ctx)
	enabled := h.cache.IsEnabled()

	status := gin.H{
		"enabled":     enabled,
		"health":      health,
		"stats":       stats,
		"timestamp":   time.Now(),
	}

	// Add performance metrics
	if stats.TotalRequests > 0 {
		status["performance"] = gin.H{
			"hit_rate":         stats.HitRate,
			"total_requests":   stats.TotalRequests,
			"average_hits":     float64(stats.Hits) / float64(stats.TotalRequests),
			"error_rate":       float64(stats.Errors) / float64(stats.TotalRequests) * 100,
		}
	}

	c.JSON(http.StatusOK, status)
}

// CacheMiddleware provides cache-related middleware
type CacheMiddleware struct {
	cache Cacher
}

// NewCacheMiddleware creates a new cache middleware
func NewCacheMiddleware(cache Cacher) *CacheMiddleware {
	return &CacheMiddleware{
		cache: cache,
	}
}

// HeaderMiddleware adds cache status headers to responses
func (m *CacheMiddleware) HeaderMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Add cache status header
		if m.cache.IsEnabled() {
			c.Header("X-Cache-Status", "enabled")
		} else {
			c.Header("X-Cache-Status", "disabled")
		}

		// Add cache statistics headers (optional, for debugging)
		if c.Query("cache_debug") == "true" {
			stats := m.cache.GetStats()
			c.Header("X-Cache-Hit-Rate", strconv.FormatFloat(stats.HitRate, 'f', 2, 64))
			c.Header("X-Cache-Total-Requests", strconv.FormatInt(stats.TotalRequests, 10))
		}

		c.Next()
	}
}

// ConditionalCacheMiddleware adds conditional caching based on cache health
func (m *CacheMiddleware) ConditionalCacheMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check cache health and add context
		ctx, cancel := context.WithTimeout(c.Request.Context(), 1*time.Second)
		defer cancel()

		health := m.cache.Health(ctx)
		
		// Add cache health to request context for handlers to use
		c.Set("cache_healthy", health.Healthy)
		c.Set("cache_enabled", m.cache.IsEnabled())
		
		// Add response headers based on cache status
		if health.Healthy {
			c.Header("X-Cache-Health", "healthy")
		} else {
			c.Header("X-Cache-Health", "unhealthy")
			c.Header("X-Cache-Fallback", "active")
		}

		c.Next()
	}
}

// RateLimitByCache provides simple rate limiting using cache
func (m *CacheMiddleware) RateLimitByCache(maxRequests int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !m.cache.IsEnabled() {
			// Skip rate limiting if cache is disabled
			c.Next()
			return
		}

		clientIP := c.ClientIP()
		key := "rate_limit:" + clientIP
		
		ctx, cancel := context.WithTimeout(c.Request.Context(), 1*time.Second)
		defer cancel()

		var count int
		err := m.cache.Get(ctx, key, &count)
		if err != nil && !IsCacheMiss(err) {
			// Cache error, allow request
			c.Next()
			return
		}

		if count >= maxRequests {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":    "rate limit exceeded",
				"limit":    maxRequests,
				"window":   window.String(),
				"reset_in": window.String(),
			})
			c.Abort()
			return
		}

		// Increment counter
		count++
		m.cache.Set(ctx, key, count, window)

		// Add rate limit headers
		c.Header("X-RateLimit-Limit", strconv.Itoa(maxRequests))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(maxRequests-count))
		c.Header("X-RateLimit-Window", window.String())

		c.Next()
	}
}