package middleware

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// NewCORS creates a new CORS middleware with the given configuration
func NewCORS(config CORSConfig) gin.HandlerFunc {
	// Validate and set defaults
	if len(config.AllowedOrigins) == 0 {
		config.AllowedOrigins = []string{"*"}
	}
	if len(config.AllowedMethods) == 0 {
		config.AllowedMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "HEAD", "PATCH"}
	}
	if len(config.AllowedHeaders) == 0 {
		config.AllowedHeaders = []string{"Content-Type", "Authorization", "X-Requested-With", "Accept", "Origin"}
	}
	if config.MaxAge == 0 {
		config.MaxAge = 12 * time.Hour
	}

	// Pre-compute header values for performance
	allowOriginAll := containsString(config.AllowedOrigins, "*")
	allowMethods := strings.Join(config.AllowedMethods, ", ")
	allowHeaders := strings.Join(config.AllowedHeaders, ", ")
	exposeHeaders := strings.Join(config.ExposedHeaders, ", ")
	maxAgeStr := strconv.Itoa(int(config.MaxAge.Seconds()))
	allowCredentials := strconv.FormatBool(config.AllowCredentials)

	return gin.HandlerFunc(func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		
		// Set Access-Control-Allow-Origin
		if allowOriginAll && !config.AllowCredentials {
			c.Header("Access-Control-Allow-Origin", "*")
		} else if origin != "" && isOriginAllowed(origin, config.AllowedOrigins, config.DevelopmentMode) {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Vary", "Origin")
		}

		// Set other CORS headers
		if config.AllowCredentials {
			c.Header("Access-Control-Allow-Credentials", allowCredentials)
		}

		// Handle preflight requests
		if c.Request.Method == "OPTIONS" {
			c.Header("Access-Control-Allow-Methods", allowMethods)
			c.Header("Access-Control-Allow-Headers", allowHeaders)
			c.Header("Access-Control-Max-Age", maxAgeStr)
			
			if !config.OptionsPassthrough {
				c.AbortWithStatus(http.StatusNoContent)
				return
			}
		} else {
			// Set headers for actual requests
			if exposeHeaders != "" {
				c.Header("Access-Control-Expose-Headers", exposeHeaders)
			}
		}

		c.Next()
	})
}

// isOriginAllowed checks if the origin is allowed
func isOriginAllowed(origin string, allowedOrigins []string, developmentMode bool) bool {
	// In development mode, allow localhost and 127.0.0.1 with any port
	if developmentMode {
		if strings.Contains(origin, "localhost") || 
		   strings.Contains(origin, "127.0.0.1") ||
		   strings.Contains(origin, "0.0.0.0") {
			return true
		}
	}

	for _, allowedOrigin := range allowedOrigins {
		if allowedOrigin == "*" {
			return true
		}
		if allowedOrigin == origin {
			return true
		}
		// Support wildcard subdomains (e.g., *.example.com)
		if strings.HasPrefix(allowedOrigin, "*.") {
			domain := allowedOrigin[2:]
			if strings.HasSuffix(origin, "."+domain) || origin == domain {
				return true
			}
		}
	}
	
	return false
}

// containsString checks if a slice contains a string
func containsString(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// CORSWithConfig creates CORS middleware with custom configuration
func CORSWithConfig(config CORSConfig) gin.HandlerFunc {
	return NewCORS(config)
}

// DefaultCORS creates CORS middleware with default settings
func DefaultCORS() gin.HandlerFunc {
	return NewCORS(CORSConfig{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "HEAD", "PATCH"},
		AllowedHeaders:   []string{"Content-Type", "Authorization", "X-Requested-With", "Accept", "Origin"},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	})
}

// StrictCORS creates CORS middleware with strict settings for production
func StrictCORS(allowedOrigins []string) gin.HandlerFunc {
	return NewCORS(CORSConfig{
		AllowedOrigins:   allowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization", "X-Requested-With"},
		ExposedHeaders:   []string{"X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           1 * time.Hour,
		DevelopmentMode:  false,
	})
}

// DevelopmentCORS creates permissive CORS middleware for development
func DevelopmentCORS() gin.HandlerFunc {
	return NewCORS(CORSConfig{
		AllowedOrigins:     []string{"*"},
		AllowedMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "HEAD", "PATCH"},
		AllowedHeaders:     []string{"*"},
		ExposedHeaders:     []string{"*"},
		AllowCredentials:   false,
		MaxAge:             24 * time.Hour,
		DevelopmentMode:    true,
		OptionsPassthrough: false,
	})
}

// CORSPreflightHandler handles CORS preflight requests explicitly
func CORSPreflightHandler(config CORSConfig) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		if c.Request.Method != "OPTIONS" {
			c.Next()
			return
		}

		origin := c.Request.Header.Get("Origin")
		method := c.Request.Header.Get("Access-Control-Request-Method")
		headers := c.Request.Header.Get("Access-Control-Request-Headers")

		// Validate origin
		if origin == "" || !isOriginAllowed(origin, config.AllowedOrigins, config.DevelopmentMode) {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}

		// Validate method
		if method != "" && !containsString(config.AllowedMethods, method) {
			c.AbortWithStatus(http.StatusMethodNotAllowed)
			return
		}

		// Validate headers
		if headers != "" {
			requestedHeaders := strings.Split(headers, ",")
			for _, header := range requestedHeaders {
				header = strings.TrimSpace(header)
				if !containsString(config.AllowedHeaders, header) && 
				   !containsString(config.AllowedHeaders, "*") {
					c.AbortWithStatus(http.StatusForbidden)
					return
				}
			}
		}

		// Set CORS headers
		c.Header("Access-Control-Allow-Origin", origin)
		c.Header("Access-Control-Allow-Methods", strings.Join(config.AllowedMethods, ", "))
		c.Header("Access-Control-Allow-Headers", strings.Join(config.AllowedHeaders, ", "))
		c.Header("Access-Control-Max-Age", strconv.Itoa(int(config.MaxAge.Seconds())))
		
		if config.AllowCredentials {
			c.Header("Access-Control-Allow-Credentials", "true")
		}

		c.AbortWithStatus(http.StatusNoContent)
	})
}