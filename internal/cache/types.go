package cache

import (
	"context"
	"sync"
	"time"
)

// Cacher defines the interface for cache operations
type Cacher interface {
	// Get retrieves a value from cache
	Get(ctx context.Context, key string, dest interface{}) error
	
	// Set stores a value in cache with TTL
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	
	// GetOrSet implements cache-aside pattern
	GetOrSet(ctx context.Context, key string, dest interface{}, ttl time.Duration, fn func() (interface{}, error)) error
	
	// Delete removes a key from cache
	Delete(ctx context.Context, key string) error
	
	// IsEnabled returns whether the cache is enabled and functioning
	IsEnabled() bool
	
	// GetStats returns current cache statistics
	GetStats() *Stats
	
	// ResetStats resets cache statistics
	ResetStats()
	
	// Close gracefully closes the cache connection
	Close() error
	
	// Health checks the health of the cache system
	Health(ctx context.Context) *HealthStatus
}

// Stats tracks cache performance metrics
type Stats struct {
	Hits              int64     `json:"hits"`
	Misses            int64     `json:"misses"`
	Sets              int64     `json:"sets"`
	Errors            int64     `json:"errors"`
	TotalRequests     int64     `json:"total_requests"`
	HitRate           float64   `json:"hit_rate"`
	LastReset         time.Time `json:"last_reset"`
	mutex             sync.RWMutex `json:"-"`
}

// HealthStatus represents the health state of the cache
type HealthStatus struct {
	Healthy      bool          `json:"healthy"`
	Status       string        `json:"status"`       // "connected", "degraded", "disconnected"
	LastError    string        `json:"last_error,omitempty"`
	ResponseTime time.Duration `json:"response_time"`
	ConnectedAt  time.Time     `json:"connected_at,omitempty"`
}

// Error types for cache-specific errors
type Error struct {
	Type    ErrorType
	Message string
	Err     error
}

type ErrorType string

const (
	ErrorTypeMiss         ErrorType = "cache_miss"
	ErrorTypeConnection   ErrorType = "connection_error"
	ErrorTypeSerialization ErrorType = "serialization_error"
	ErrorTypeTimeout      ErrorType = "timeout_error"
	ErrorTypeDisabled     ErrorType = "cache_disabled"
)

func (e *Error) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}

func (e *Error) Unwrap() error {
	return e.Err
}

func NewCacheError(errType ErrorType, message string, err error) *Error {
	return &Error{
		Type:    errType,
		Message: message,
		Err:     err,
	}
}

// IsCacheMiss checks if an error is a cache miss
func IsCacheMiss(err error) bool {
	if cacheErr, ok := err.(*Error); ok {
		return cacheErr.Type == ErrorTypeMiss
	}
	return false
}

// IsCacheDisabled checks if an error is due to cache being disabled
func IsCacheDisabled(err error) bool {
	if cacheErr, ok := err.(*Error); ok {
		return cacheErr.Type == ErrorTypeDisabled
	}
	return false
}

// Cache key prefixes and TTL constants
const (
	// TTL constants
	CarbonIntensityTTL    = 5 * time.Minute  // 5 minutes for carbon intensity data
	OptimizationTTL       = 10 * time.Minute // 10 minutes for optimization profiles
	GreenHoursForecastTTL = 1 * time.Hour    // 1 hour for green hours forecast

	// Cache key prefixes
	CarbonIntensityPrefix = "carbon_intensity"
	OptimizationPrefix    = "optimization"
	GreenHoursPrefix      = "green_hours"
	HealthPrefix          = "health"
)