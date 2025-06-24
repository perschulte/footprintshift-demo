package middleware

import (
	"bytes"
	"context"
	"log/slog"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestCORSMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		config         CORSConfig
		origin         string
		method         string
		expectedStatus int
		expectedHeaders map[string]string
	}{
		{
			name: "allow all origins",
			config: CORSConfig{
				AllowedOrigins: []string{"*"},
				AllowedMethods: []string{"GET", "POST"},
			},
			origin:         "https://example.com",
			method:         "GET",
			expectedStatus: 200,
			expectedHeaders: map[string]string{
				"Access-Control-Allow-Origin": "*",
			},
		},
		{
			name: "specific origin allowed",
			config: CORSConfig{
				AllowedOrigins: []string{"https://example.com"},
				AllowedMethods: []string{"GET", "POST"},
			},
			origin:         "https://example.com",
			method:         "GET",
			expectedStatus: 200,
			expectedHeaders: map[string]string{
				"Access-Control-Allow-Origin": "https://example.com",
			},
		},
		{
			name: "origin not allowed",
			config: CORSConfig{
				AllowedOrigins: []string{"https://allowed.com"},
				AllowedMethods: []string{"GET", "POST"},
			},
			origin:         "https://notallowed.com",
			method:         "GET",
			expectedStatus: 200,
			expectedHeaders: map[string]string{
				"Access-Control-Allow-Origin": "",
			},
		},
		{
			name: "preflight request",
			config: CORSConfig{
				AllowedOrigins: []string{"https://example.com"},
				AllowedMethods: []string{"GET", "POST"},
				AllowedHeaders: []string{"Content-Type"},
			},
			origin:         "https://example.com",
			method:         "OPTIONS",
			expectedStatus: 204,
			expectedHeaders: map[string]string{
				"Access-Control-Allow-Origin":  "https://example.com",
				"Access-Control-Allow-Methods": "GET, POST",
				"Access-Control-Allow-Headers": "Content-Type",
			},
		},
		{
			name: "development mode localhost",
			config: CORSConfig{
				AllowedOrigins:  []string{"https://production.com"},
				DevelopmentMode: true,
			},
			origin:         "http://localhost:3000",
			method:         "GET",
			expectedStatus: 200,
			expectedHeaders: map[string]string{
				"Access-Control-Allow-Origin": "http://localhost:3000",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()
			r.Use(NewCORS(tt.config))
			r.GET("/test", func(c *gin.Context) {
				c.JSON(200, gin.H{"ok": true})
			})
			r.OPTIONS("/test", func(c *gin.Context) {
				c.JSON(200, gin.H{"ok": true})
			})

			req := httptest.NewRequest(tt.method, "/test", nil)
			if tt.origin != "" {
				req.Header.Set("Origin", tt.origin)
			}
			if tt.method == "OPTIONS" {
				req.Header.Set("Access-Control-Request-Method", "POST")
			}

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			for header, expectedValue := range tt.expectedHeaders {
				if expectedValue == "" {
					assert.Empty(t, w.Header().Get(header), "Header %s should be empty", header)
				} else {
					assert.Equal(t, expectedValue, w.Header().Get(header), "Header %s mismatch", header)
				}
			}
		})
	}
}

func TestRateLimitMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		config         RateLimitConfig
		requests       int
		expectedPasses int
	}{
		{
			name: "basic rate limiting",
			config: RateLimitConfig{
				Enabled:      true,
				DefaultRPS:   2,
				DefaultBurst: 2,
				WindowSize:   time.Second,
				UseRedis:     false,
			},
			requests:       4,
			expectedPasses: 2,
		},
		{
			name: "burst capacity",
			config: RateLimitConfig{
				Enabled:      true,
				DefaultRPS:   1,
				DefaultBurst: 3,
				WindowSize:   time.Second,
				UseRedis:     false,
			},
			requests:       5,
			expectedPasses: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()
			r.Use(NewRateLimit(tt.config, nil))
			r.GET("/test", func(c *gin.Context) {
				c.JSON(200, gin.H{"ok": true})
			})

			passes := 0
			for i := 0; i < tt.requests; i++ {
				req := httptest.NewRequest("GET", "/test", nil)
				req.RemoteAddr = "192.168.1.1:12345" // Fixed IP for consistent testing
				w := httptest.NewRecorder()
				r.ServeHTTP(w, req)

				if w.Code == 200 {
					passes++
				}
			}

			assert.Equal(t, tt.expectedPasses, passes)
		})
	}
}

func TestRateLimitWithRedis(t *testing.T) {
	// Skip if Redis is not available
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   1, // Use different DB for testing
	})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := redisClient.Ping(ctx).Err(); err != nil {
		t.Skip("Redis not available for testing")
	}

	// Clean up test keys
	defer redisClient.FlushDB(ctx)

	config := RateLimitConfig{
		Enabled:        true,
		DefaultRPS:     2,
		DefaultBurst:   2,
		WindowSize:     time.Second,
		UseRedis:       true,
		RedisKeyPrefix: "test:ratelimit:",
	}

	r := gin.New()
	r.Use(NewRateLimit(config, redisClient))
	r.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	passes := 0
	for i := 0; i < 4; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code == 200 {
			passes++
		}
	}

	assert.Equal(t, 2, passes)
}

func TestLoggingMiddleware(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	config := LoggingConfig{
		Enabled:       true,
		Level:         slog.LevelDebug,
		SlowThreshold: 100 * time.Millisecond,
		EnableMetrics: true,
	}

	r := gin.New()
	r.Use(RequestID())
	r.Use(NewLogging(config, logger))
	r.GET("/test", func(c *gin.Context) {
		time.Sleep(50 * time.Millisecond) // Simulate processing time
		c.JSON(200, gin.H{"message": "test"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("User-Agent", "test-agent")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.NotEmpty(t, w.Header().Get("X-Request-ID"))

	// Check that logging occurred
	logOutput := buf.String()
	assert.Contains(t, logOutput, "HTTP Request")
	assert.Contains(t, logOutput, "HTTP 200 GET /test")
	assert.Contains(t, logOutput, "test-agent")
}

func TestRequestIDMiddleware(t *testing.T) {
	r := gin.New()
	r.Use(RequestID())
	r.GET("/test", func(c *gin.Context) {
		requestID, exists := c.Get("request_id")
		assert.True(t, exists)
		assert.NotEmpty(t, requestID)
		c.JSON(200, gin.H{"request_id": requestID})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.NotEmpty(t, w.Header().Get("X-Request-ID"))
}

func TestRequestIDWithExistingHeader(t *testing.T) {
	r := gin.New()
	r.Use(RequestID())
	r.GET("/test", func(c *gin.Context) {
		requestID, exists := c.Get("request_id")
		assert.True(t, exists)
		assert.Equal(t, "existing-id", requestID)
		c.JSON(200, gin.H{"request_id": requestID})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Request-ID", "existing-id")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Equal(t, "existing-id", w.Header().Get("X-Request-ID"))
}

func TestSecureMiddleware(t *testing.T) {
	r := gin.New()
	r.Use(Secure())
	r.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Equal(t, "nosniff", w.Header().Get("X-Content-Type-Options"))
	assert.Equal(t, "DENY", w.Header().Get("X-Frame-Options"))
	assert.Equal(t, "1; mode=block", w.Header().Get("X-XSS-Protection"))
	assert.Equal(t, "strict-origin-when-cross-origin", w.Header().Get("Referrer-Policy"))
	assert.Equal(t, "default-src 'self'", w.Header().Get("Content-Security-Policy"))
}

func TestTimeoutMiddleware(t *testing.T) {
	r := gin.New()
	r.Use(Timeout(100 * time.Millisecond))
	r.GET("/fast", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})
	r.GET("/slow", func(c *gin.Context) {
		time.Sleep(200 * time.Millisecond)
		c.JSON(200, gin.H{"ok": true})
	})

	// Test fast request
	req := httptest.NewRequest("GET", "/fast", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)

	// Test slow request (should timeout)
	req = httptest.NewRequest("GET", "/slow", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, 408, w.Code)
}

func TestMiddlewareChain(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	config := Config{
		CORS: CORSConfig{
			AllowedOrigins: []string{"*"},
		},
		RateLimit: RateLimitConfig{
			Enabled:      true,
			DefaultRPS:   10,
			DefaultBurst: 10,
			WindowSize:   time.Minute,
			UseRedis:     false,
		},
		Logging: LoggingConfig{
			Enabled: true,
			Level:   slog.LevelInfo,
		},
		Logger: logger,
	}

	r := gin.New()
	middlewares := Chain(config)
	for _, mw := range middlewares {
		r.Use(mw)
	}

	r.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "https://example.com")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.NotEmpty(t, w.Header().Get("X-Request-ID"))
	assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
}

func TestPerUserRateLimit(t *testing.T) {
	config := RateLimitConfig{
		Enabled:      true,
		DefaultRPS:   2,
		DefaultBurst: 2,
		WindowSize:   time.Second,
		UseRedis:     false,
	}

	r := gin.New()
	r.Use(PerUserRateLimit(config, nil, "X-User-ID"))
	r.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	// Test user 1
	user1Passes := 0
	for i := 0; i < 4; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("X-User-ID", "user1")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code == 200 {
			user1Passes++
		}
	}

	// Test user 2 (should have separate limit)
	user2Passes := 0
	for i := 0; i < 4; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("X-User-ID", "user2")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code == 200 {
			user2Passes++
		}
	}

	assert.Equal(t, 2, user1Passes)
	assert.Equal(t, 2, user2Passes)
}

func TestDefaultConfigurations(t *testing.T) {
	// Test default config
	defaultConfig := DefaultConfig()
	assert.True(t, defaultConfig.RateLimit.Enabled)
	assert.Equal(t, 100, defaultConfig.RateLimit.DefaultRPS)
	assert.Equal(t, []string{"*"}, defaultConfig.CORS.AllowedOrigins)

	// Test development config
	devConfig := DevelopmentConfig()
	assert.True(t, devConfig.CORS.DevelopmentMode)
	assert.Equal(t, 1000, devConfig.RateLimit.DefaultRPS)
	assert.Equal(t, slog.LevelDebug, devConfig.Logging.Level)

	// Test production config
	prodConfig := ProductionConfig([]string{"https://example.com"})
	assert.False(t, prodConfig.CORS.DevelopmentMode)
	assert.Equal(t, []string{"https://example.com"}, prodConfig.CORS.AllowedOrigins)
	assert.True(t, prodConfig.CORS.AllowCredentials)
	assert.Equal(t, 50, prodConfig.RateLimit.DefaultRPS)
}

func TestEndpointSpecificRateLimit(t *testing.T) {
	config := RateLimitConfig{
		Enabled:      true,
		DefaultRPS:   10,
		DefaultBurst: 10,
		WindowSize:   time.Second,
		UseRedis:     false,
		EndpointLimits: map[string]EndpointRate{
			"/limited": {RPS: 1, Burst: 1, Window: time.Second},
		},
	}

	r := gin.New()
	r.Use(NewRateLimit(config, nil))
	r.GET("/limited", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})
	r.GET("/normal", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	// Test limited endpoint
	limitedPasses := 0
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest("GET", "/limited", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code == 200 {
			limitedPasses++
		}
	}

	// Test normal endpoint (should have higher limit)
	normalPasses := 0
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest("GET", "/normal", nil)
		req.RemoteAddr = "192.168.1.2:12345" // Different IP
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code == 200 {
			normalPasses++
		}
	}

	assert.Equal(t, 1, limitedPasses)
	assert.Equal(t, 3, normalPasses)
}

func TestStructuredLoggingHelpers(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	r := gin.New()
	r.Use(RequestID())
	r.GET("/test", func(c *gin.Context) {
		LogInfo(c, logger, "Test message", slog.String("key", "value"))
		LogWarn(c, logger, "Warning message")
		LogDebug(c, logger, "Debug message")
		c.JSON(200, gin.H{"ok": true})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	
	logOutput := buf.String()
	assert.Contains(t, logOutput, "Test message")
	assert.Contains(t, logOutput, "Warning message")
	assert.Contains(t, logOutput, "Debug message")
	assert.Contains(t, logOutput, "request_id")
	assert.Contains(t, logOutput, "method")
	assert.Contains(t, logOutput, "path")
}