package middleware

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// LoggingMiddleware provides structured request/response logging
type LoggingMiddleware struct {
	config LoggingConfig
	logger *slog.Logger
	pool   sync.Pool
}

// responseWriter wraps gin.ResponseWriter to capture response data
type responseWriter struct {
	gin.ResponseWriter
	body       *bytes.Buffer
	statusCode int
	size       int
}

// requestMetrics holds metrics for a request
type requestMetrics struct {
	StartTime    time.Time
	EndTime      time.Time
	Duration     time.Duration
	RequestSize  int64
	ResponseSize int
	StatusCode   int
	RequestID    string
	Method       string
	Path         string
	IP           string
	UserAgent    string
	Referer      string
}

// NewLogging creates a new logging middleware
func NewLogging(config LoggingConfig, logger *slog.Logger) gin.HandlerFunc {
	if logger == nil {
		logger = slog.Default()
	}

	middleware := &LoggingMiddleware{
		config: config,
		logger: logger,
		pool: sync.Pool{
			New: func() interface{} {
				return &bytes.Buffer{}
			},
		},
	}

	return middleware.Handler()
}

// Handler returns the gin middleware handler
func (m *LoggingMiddleware) Handler() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// Skip logging for certain paths
		if m.shouldSkipPath(c.Request.URL.Path) {
			c.Next()
			return
		}

		// Create metrics
		metrics := &requestMetrics{
			StartTime: time.Now(),
			Method:    c.Request.Method,
			Path:      c.Request.URL.Path,
			IP:        getClientIP(c),
			UserAgent: c.Request.UserAgent(),
			Referer:   c.Request.Referer(),
		}

		// Get request ID
		if requestID, exists := c.Get("request_id"); exists {
			metrics.RequestID = requestID.(string)
		}

		// Get request size
		if c.Request.ContentLength > 0 {
			metrics.RequestSize = c.Request.ContentLength
		}

		// Wrap response writer to capture response data
		var rw *responseWriter
		if m.config.EnableBody {
			rw = &responseWriter{
				ResponseWriter: c.Writer,
				body:           m.pool.Get().(*bytes.Buffer),
				statusCode:     200,
			}
			rw.body.Reset()
			c.Writer = rw
		}

		// Process request
		m.logRequest(c, metrics)

		// Process the request
		c.Next()

		// Complete metrics
		metrics.EndTime = time.Now()
		metrics.Duration = metrics.EndTime.Sub(metrics.StartTime)
		metrics.StatusCode = c.Writer.Status()
		metrics.ResponseSize = c.Writer.Size()

		// Log response
		m.logResponse(c, metrics, rw)

		// Return buffer to pool
		if rw != nil && rw.body != nil {
			m.pool.Put(rw.body)
		}
	})
}

// shouldSkipPath checks if logging should be skipped for this path
func (m *LoggingMiddleware) shouldSkipPath(path string) bool {
	for _, skipPath := range m.config.SkipPaths {
		if path == skipPath {
			return true
		}
		// Support wildcard patterns
		if strings.HasSuffix(skipPath, "*") {
			prefix := skipPath[:len(skipPath)-1]
			if strings.HasPrefix(path, prefix) {
				return true
			}
		}
	}
	return false
}

// logRequest logs the incoming request
func (m *LoggingMiddleware) logRequest(c *gin.Context, metrics *requestMetrics) {
	if m.config.Level > slog.LevelDebug {
		return
	}

	attrs := []slog.Attr{
		slog.String("type", "request"),
		slog.String("method", metrics.Method),
		slog.String("path", metrics.Path),
		slog.String("ip", metrics.IP),
		slog.String("user_agent", metrics.UserAgent),
		slog.Time("timestamp", metrics.StartTime),
	}

	if metrics.RequestID != "" {
		attrs = append(attrs, slog.String("request_id", metrics.RequestID))
	}

	if metrics.Referer != "" {
		attrs = append(attrs, slog.String("referer", metrics.Referer))
	}

	if metrics.RequestSize > 0 {
		attrs = append(attrs, slog.Int64("request_size", metrics.RequestSize))
	}

	// Add request headers if configured
	if len(m.config.RequestHeaders) > 0 {
		headers := make(map[string]string)
		for _, header := range m.config.RequestHeaders {
			if value := c.GetHeader(header); value != "" {
				headers[strings.ToLower(header)] = value
			}
		}
		if len(headers) > 0 {
			attrs = append(attrs, slog.Any("request_headers", headers))
		}
	}

	// Add request body if enabled and not too large
	if m.config.EnableBody && c.Request.ContentLength > 0 && c.Request.ContentLength <= m.config.MaxBodySize {
		if body := m.readRequestBody(c); body != "" {
			attrs = append(attrs, slog.String("request_body", body))
		}
	}

	m.logger.LogAttrs(c.Request.Context(), slog.LevelDebug, "HTTP Request", attrs...)
}

// logResponse logs the response
func (m *LoggingMiddleware) logResponse(c *gin.Context, metrics *requestMetrics, rw *responseWriter) {
	level := m.determineLogLevel(metrics.StatusCode, metrics.Duration)
	
	attrs := []slog.Attr{
		slog.String("type", "response"),
		slog.String("method", metrics.Method),
		slog.String("path", metrics.Path),
		slog.String("ip", metrics.IP),
		slog.Int("status_code", metrics.StatusCode),
		slog.Duration("duration", metrics.Duration),
		slog.Int("response_size", metrics.ResponseSize),
		slog.Time("timestamp", metrics.EndTime),
	}

	if metrics.RequestID != "" {
		attrs = append(attrs, slog.String("request_id", metrics.RequestID))
	}

	// Add response headers if configured
	if len(m.config.ResponseHeaders) > 0 {
		headers := make(map[string]string)
		for _, header := range m.config.ResponseHeaders {
			if value := c.Writer.Header().Get(header); value != "" {
				headers[strings.ToLower(header)] = value
			}
		}
		if len(headers) > 0 {
			attrs = append(attrs, slog.Any("response_headers", headers))
		}
	}

	// Add response body if enabled
	if m.config.EnableBody && rw != nil && rw.body.Len() > 0 {
		bodyStr := rw.body.String()
		if int64(len(bodyStr)) <= m.config.MaxBodySize {
			attrs = append(attrs, slog.String("response_body", bodyStr))
		}
	}

	// Add performance metrics if enabled
	if m.config.EnableMetrics {
		attrs = append(attrs, 
			slog.Float64("duration_ms", float64(metrics.Duration.Nanoseconds())/1e6),
			slog.Bool("slow_request", metrics.Duration > m.config.SlowThreshold),
		)
	}

	// Add error information for error responses
	if metrics.StatusCode >= 400 {
		if errors := c.Errors; len(errors) > 0 {
			errorMsgs := make([]string, len(errors))
			for i, err := range errors {
				errorMsgs[i] = err.Error()
			}
			attrs = append(attrs, slog.Any("errors", errorMsgs))
		}
	}

	message := fmt.Sprintf("HTTP %d %s %s", metrics.StatusCode, metrics.Method, metrics.Path)
	m.logger.LogAttrs(c.Request.Context(), level, message, attrs...)
}

// determineLogLevel determines the appropriate log level based on status code and duration
func (m *LoggingMiddleware) determineLogLevel(statusCode int, duration time.Duration) slog.Level {
	// Error responses
	if statusCode >= 500 {
		return slog.LevelError
	}
	
	// Client errors
	if statusCode >= 400 {
		return slog.LevelWarn
	}
	
	// Slow requests
	if duration > m.config.SlowThreshold {
		return slog.LevelWarn
	}
	
	// Normal requests
	return slog.LevelInfo
}

// readRequestBody reads and returns the request body as a string
func (m *LoggingMiddleware) readRequestBody(c *gin.Context) string {
	if c.Request.Body == nil {
		return ""
	}

	// Read body
	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return ""
	}

	// Restore body for further processing
	c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	// Return body as string, truncating if necessary
	bodyStr := string(bodyBytes)
	if int64(len(bodyStr)) > m.config.MaxBodySize {
		bodyStr = bodyStr[:m.config.MaxBodySize] + "... (truncated)"
	}

	return bodyStr
}

// responseWriter implementation

func (rw *responseWriter) Write(data []byte) (int, error) {
	// Write to both the original writer and our buffer
	n, err := rw.ResponseWriter.Write(data)
	
	if rw.body != nil {
		rw.body.Write(data[:n])
	}
	
	rw.size += n
	return n, err
}

func (rw *responseWriter) WriteHeader(statusCode int) {
	rw.statusCode = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

func (rw *responseWriter) Status() int {
	return rw.statusCode
}

func (rw *responseWriter) Size() int {
	return rw.size
}

// Structured logging helpers

// LogWithContext logs a message with request context
func LogWithContext(c *gin.Context, logger *slog.Logger, level slog.Level, message string, attrs ...slog.Attr) {
	contextAttrs := []slog.Attr{
		slog.String("method", c.Request.Method),
		slog.String("path", c.Request.URL.Path),
		slog.String("ip", getClientIP(c)),
	}

	if requestID, exists := c.Get("request_id"); exists {
		contextAttrs = append(contextAttrs, slog.String("request_id", requestID.(string)))
	}

	allAttrs := append(contextAttrs, attrs...)
	logger.LogAttrs(c.Request.Context(), level, message, allAttrs...)
}

// LogError logs an error with request context
func LogError(c *gin.Context, logger *slog.Logger, err error, message string) {
	LogWithContext(c, logger, slog.LevelError, message, 
		slog.String("error", err.Error()),
	)
}

// LogWarn logs a warning with request context
func LogWarn(c *gin.Context, logger *slog.Logger, message string, attrs ...slog.Attr) {
	LogWithContext(c, logger, slog.LevelWarn, message, attrs...)
}

// LogInfo logs an info message with request context
func LogInfo(c *gin.Context, logger *slog.Logger, message string, attrs ...slog.Attr) {
	LogWithContext(c, logger, slog.LevelInfo, message, attrs...)
}

// LogDebug logs a debug message with request context
func LogDebug(c *gin.Context, logger *slog.Logger, message string, attrs ...slog.Attr) {
	LogWithContext(c, logger, slog.LevelDebug, message, attrs...)
}

// Performance monitoring middleware

// PerformanceMonitor creates a middleware that monitors performance metrics
func PerformanceMonitor(logger *slog.Logger, slowThreshold time.Duration) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		start := time.Now()
		
		c.Next()
		
		duration := time.Since(start)
		
		attrs := []slog.Attr{
			slog.String("method", c.Request.Method),
			slog.String("path", c.Request.URL.Path),
			slog.Int("status_code", c.Writer.Status()),
			slog.Duration("duration", duration),
			slog.Float64("duration_ms", float64(duration.Nanoseconds())/1e6),
			slog.Bool("slow_request", duration > slowThreshold),
		}
		
		if requestID, exists := c.Get("request_id"); exists {
			attrs = append(attrs, slog.String("request_id", requestID.(string)))
		}
		
		level := slog.LevelInfo
		if duration > slowThreshold {
			level = slog.LevelWarn
		}
		
		logger.LogAttrs(c.Request.Context(), level, "Request completed", attrs...)
	})
}