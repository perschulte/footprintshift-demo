package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// Service implements the Cacher interface using Redis
type Service struct {
	client      redis.UniversalClient
	config      *Config
	enabled     bool
	stats       *Stats
	mutex       sync.RWMutex
	healthMutex sync.RWMutex
	lastHealth  *HealthStatus
	connectedAt time.Time
}

// New creates a new cache service with the provided configuration
func New(config *Config) *Service {
	if config == nil {
		config = DefaultConfig()
	}
	
	// Validate configuration
	if err := config.Validate(); err != nil {
		log.Printf("Invalid cache configuration: %v", err)
		return &Service{
			config:  config,
			enabled: false,
			stats: &Stats{
				LastReset: time.Now(),
			},
		}
	}

	service := &Service{
		config: config,
		stats: &Stats{
			LastReset: time.Now(),
		},
	}

	// Initialize Redis client
	service.initializeRedisClient()

	return service
}

// NewFromEnv creates a new cache service using environment configuration
func NewFromEnv() *Service {
	config := LoadFromEnv()
	return New(config)
}

// initializeRedisClient sets up the Redis connection
func (s *Service) initializeRedisClient() {
	opts, err := redis.ParseURL(s.config.URL)
	if err != nil {
		log.Printf("Failed to parse Redis URL, disabling cache: %v", err)
		s.setEnabled(false)
		return
	}

	// Apply configuration
	opts.MaxRetries = s.config.MaxRetries
	opts.MinRetryBackoff = s.config.MinRetryBackoff
	opts.MaxRetryBackoff = s.config.MaxRetryBackoff
	opts.DialTimeout = s.config.DialTimeout
	opts.ReadTimeout = s.config.ReadTimeout
	opts.WriteTimeout = s.config.WriteTimeout
	opts.PoolSize = s.config.PoolSize
	opts.MinIdleConns = s.config.MinIdleConns
	opts.ConnMaxLifetime = s.config.MaxConnAge
	opts.PoolTimeout = s.config.PoolTimeout
	opts.MaxIdleConns = s.config.MinIdleConns * 2
	opts.ConnMaxIdleTime = s.config.IdleTimeout

	s.client = redis.NewClient(opts)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	start := time.Now()
	if err := s.client.Ping(ctx).Err(); err != nil {
		log.Printf("Redis connection failed, cache disabled: %v", err)
		s.setEnabled(false)
		s.updateHealth(false, "disconnected", err.Error(), time.Since(start))
		if s.client != nil {
			s.client.Close()
			s.client = nil
		}
	} else {
		log.Printf("Redis cache enabled and connected")
		s.setEnabled(true)
		s.connectedAt = time.Now()
		s.updateHealth(true, "connected", "", time.Since(start))
	}
}

// IsEnabled returns whether the cache is enabled and functioning
func (s *Service) IsEnabled() bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.enabled
}

// setEnabled updates the enabled status thread-safely
func (s *Service) setEnabled(enabled bool) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.enabled = enabled
}

// Get retrieves a value from cache
func (s *Service) Get(ctx context.Context, key string, dest interface{}) error {
	s.incrementTotalRequests()

	if !s.IsEnabled() {
		s.incrementMisses()
		return NewCacheError(ErrorTypeDisabled, "cache disabled", nil)
	}

	data, err := s.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			s.incrementMisses()
			return NewCacheError(ErrorTypeMiss, "cache miss", nil)
		}
		s.incrementErrors()
		return NewCacheError(ErrorTypeConnection, "cache get error", err)
	}

	if err := json.Unmarshal([]byte(data), dest); err != nil {
		s.incrementErrors()
		return NewCacheError(ErrorTypeSerialization, "cache unmarshal error", err)
	}

	s.incrementHits()
	return nil
}

// Set stores a value in cache
func (s *Service) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	if !s.IsEnabled() {
		if s.config.GracefulDegradation {
			return nil // Silently ignore if cache is disabled and graceful degradation is enabled
		}
		return NewCacheError(ErrorTypeDisabled, "cache disabled", nil)
	}

	s.incrementSets()

	data, err := json.Marshal(value)
	if err != nil {
		s.incrementErrors()
		return NewCacheError(ErrorTypeSerialization, "cache marshal error", err)
	}

	if err := s.client.Set(ctx, key, data, ttl).Err(); err != nil {
		s.incrementErrors()
		return NewCacheError(ErrorTypeConnection, "cache set error", err)
	}

	return nil
}

// GetOrSet implements cache-aside pattern: get from cache, or execute function and cache result
func (s *Service) GetOrSet(ctx context.Context, key string, dest interface{}, ttl time.Duration, fn func() (interface{}, error)) error {
	// Try to get from cache first
	if err := s.Get(ctx, key, dest); err == nil {
		return nil // Cache hit
	}

	// Cache miss - execute function
	result, err := fn()
	if err != nil {
		return err
	}

	// Store result in cache (ignore cache errors in graceful degradation mode)
	if setErr := s.Set(ctx, key, result, ttl); setErr != nil {
		if !s.config.GracefulDegradation {
			return setErr
		}
		log.Printf("Warning: failed to set cache for key %s: %v", key, setErr)
	}

	// Marshal result to dest
	data, err := json.Marshal(result)
	if err != nil {
		return NewCacheError(ErrorTypeSerialization, "marshal result error", err)
	}

	if err := json.Unmarshal(data, dest); err != nil {
		return NewCacheError(ErrorTypeSerialization, "unmarshal result error", err)
	}

	return nil
}

// Delete removes a key from cache
func (s *Service) Delete(ctx context.Context, key string) error {
	if !s.IsEnabled() {
		if s.config.GracefulDegradation {
			return nil // Silently ignore if cache is disabled and graceful degradation is enabled
		}
		return NewCacheError(ErrorTypeDisabled, "cache disabled", nil)
	}

	if err := s.client.Del(ctx, key).Err(); err != nil {
		s.incrementErrors()
		return NewCacheError(ErrorTypeConnection, "cache delete error", err)
	}

	return nil
}

// Close gracefully closes the Redis connection
func (s *Service) Close() error {
	if s.client != nil {
		return s.client.Close()
	}
	return nil
}

// GetStats returns current cache statistics
func (s *Service) GetStats() *Stats {
	s.stats.mutex.RLock()
	defer s.stats.mutex.RUnlock()

	stats := &Stats{
		Hits:          s.stats.Hits,
		Misses:        s.stats.Misses,
		Sets:          s.stats.Sets,
		Errors:        s.stats.Errors,
		TotalRequests: s.stats.TotalRequests,
		LastReset:     s.stats.LastReset,
	}

	// Calculate hit rate
	if stats.TotalRequests > 0 {
		stats.HitRate = float64(stats.Hits) / float64(stats.TotalRequests) * 100
	}

	return stats
}

// ResetStats resets cache statistics
func (s *Service) ResetStats() {
	s.stats.mutex.Lock()
	defer s.stats.mutex.Unlock()

	s.stats.Hits = 0
	s.stats.Misses = 0
	s.stats.Sets = 0
	s.stats.Errors = 0
	s.stats.TotalRequests = 0
	s.stats.HitRate = 0
	s.stats.LastReset = time.Now()
}

// Health checks the health of the cache system
func (s *Service) Health(ctx context.Context) *HealthStatus {
	if !s.IsEnabled() || s.client == nil {
		return &HealthStatus{
			Healthy:      false,
			Status:       "disconnected",
			LastError:    "cache disabled or client not initialized",
			ResponseTime: 0,
		}
	}

	start := time.Now()
	err := s.client.Ping(ctx).Err()
	responseTime := time.Since(start)

	if err != nil {
		s.updateHealth(false, "disconnected", err.Error(), responseTime)
		// If health check fails and we're in graceful degradation mode, disable cache
		if s.config.GracefulDegradation {
			s.setEnabled(false)
		}
	} else {
		s.updateHealth(true, "connected", "", responseTime)
		if !s.IsEnabled() {
			s.setEnabled(true) // Re-enable if health check passes
		}
	}

	s.healthMutex.RLock()
	defer s.healthMutex.RUnlock()
	
	// Return a copy of the health status
	return &HealthStatus{
		Healthy:      s.lastHealth.Healthy,
		Status:       s.lastHealth.Status,
		LastError:    s.lastHealth.LastError,
		ResponseTime: s.lastHealth.ResponseTime,
		ConnectedAt:  s.connectedAt,
	}
}

// updateHealth updates the cached health status
func (s *Service) updateHealth(healthy bool, status, lastError string, responseTime time.Duration) {
	s.healthMutex.Lock()
	defer s.healthMutex.Unlock()
	
	s.lastHealth = &HealthStatus{
		Healthy:      healthy,
		Status:       status,
		LastError:    lastError,
		ResponseTime: responseTime,
		ConnectedAt:  s.connectedAt,
	}
}

// Helper methods for cache key generation
func (s *Service) GenerateKey(prefix string, parts ...string) string {
	return s.config.GenerateKey(prefix, parts...)
}

// GetCarbonIntensityKey generates a cache key for carbon intensity data
func (s *Service) GetCarbonIntensityKey(location string) string {
	return s.GenerateKey(CarbonIntensityPrefix, location)
}

// GetOptimizationKey generates a cache key for optimization profiles
func (s *Service) GetOptimizationKey(location, url string) string {
	return s.GenerateKey(OptimizationPrefix, location, url)
}

// GetGreenHoursKey generates a cache key for green hours forecast
func (s *Service) GetGreenHoursKey(location string, hours int) string {
	return s.GenerateKey(GreenHoursPrefix, location, fmt.Sprintf("%d", hours))
}

// Private methods for updating statistics
func (s *Service) incrementHits() {
	s.stats.mutex.Lock()
	defer s.stats.mutex.Unlock()
	s.stats.Hits++
}

func (s *Service) incrementMisses() {
	s.stats.mutex.Lock()
	defer s.stats.mutex.Unlock()
	s.stats.Misses++
}

func (s *Service) incrementSets() {
	s.stats.mutex.Lock()
	defer s.stats.mutex.Unlock()
	s.stats.Sets++
}

func (s *Service) incrementErrors() {
	s.stats.mutex.Lock()
	defer s.stats.mutex.Unlock()
	s.stats.Errors++
}

func (s *Service) incrementTotalRequests() {
	s.stats.mutex.Lock()
	defer s.stats.mutex.Unlock()
	s.stats.TotalRequests++
}

