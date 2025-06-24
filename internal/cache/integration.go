package cache

import (
	"context"
	"time"
)

// CachedElectricityMaps provides cached wrapper for electricity maps operations
type CachedElectricityMaps struct {
	cache Cacher
}

// NewCachedElectricityMaps creates a new cached electricity maps wrapper
func NewCachedElectricityMaps(cache Cacher) *CachedElectricityMaps {
	return &CachedElectricityMaps{
		cache: cache,
	}
}

// CachedCarbonIntensity caches carbon intensity data
func (c *CachedElectricityMaps) CachedCarbonIntensity(ctx context.Context, location string, fetcher func() (interface{}, error)) (interface{}, error) {
	key := c.cache.(*Service).GetCarbonIntensityKey(location)
	
	var result interface{}
	err := c.cache.GetOrSet(ctx, key, &result, CarbonIntensityTTL, fetcher)
	return result, err
}

// CachedGreenHoursForecast caches green hours forecast data
func (c *CachedElectricityMaps) CachedGreenHoursForecast(ctx context.Context, location string, hours int, fetcher func() (interface{}, error)) (interface{}, error) {
	key := c.cache.(*Service).GetGreenHoursKey(location, hours)
	
	var result interface{}
	err := c.cache.GetOrSet(ctx, key, &result, GreenHoursForecastTTL, fetcher)
	return result, err
}

// CachedOptimizationProfile provides cached wrapper for optimization operations
type CachedOptimizationProfile struct {
	cache Cacher
}

// NewCachedOptimizationProfile creates a new cached optimization wrapper
func NewCachedOptimizationProfile(cache Cacher) *CachedOptimizationProfile {
	return &CachedOptimizationProfile{
		cache: cache,
	}
}

// CachedOptimization caches optimization profile data
func (c *CachedOptimizationProfile) CachedOptimization(ctx context.Context, location, url string, fetcher func() (interface{}, error)) (interface{}, error) {
	key := c.cache.(*Service).GetOptimizationKey(location, url)
	
	var result interface{}
	err := c.cache.GetOrSet(ctx, key, &result, OptimizationTTL, fetcher)
	return result, err
}

// CacheManager provides high-level cache management operations
type CacheManager struct {
	cache Cacher
}

// NewCacheManager creates a new cache manager
func NewCacheManager(cache Cacher) *CacheManager {
	return &CacheManager{
		cache: cache,
	}
}

// WarmUpCache pre-populates cache with commonly accessed data
func (m *CacheManager) WarmUpCache(ctx context.Context, locations []string) error {
	for _, location := range locations {
		// You can add warm-up logic here
		// For example, pre-fetch carbon intensity for common locations
		key := m.cache.(*Service).GetCarbonIntensityKey(location)
		
		// Skip if already cached
		var dummy interface{}
		if err := m.cache.Get(ctx, key, &dummy); err == nil {
			continue // Already cached
		}
		
		// Add warm-up logic here if needed
	}
	return nil
}

// ClearLocationCache clears all cache entries for a specific location
func (m *CacheManager) ClearLocationCache(ctx context.Context, location string) error {
	// Clear carbon intensity cache
	carbonKey := m.cache.(*Service).GetCarbonIntensityKey(location)
	if err := m.cache.Delete(ctx, carbonKey); err != nil && !IsCacheDisabled(err) {
		return err
	}
	
	// Clear green hours cache for common hour ranges
	commonHours := []int{24, 48, 72, 168}
	for _, hours := range commonHours {
		greenKey := m.cache.(*Service).GetGreenHoursKey(location, hours)
		if err := m.cache.Delete(ctx, greenKey); err != nil && !IsCacheDisabled(err) {
			return err
		}
	}
	
	return nil
}

// GetCacheMetrics returns comprehensive cache metrics
func (m *CacheManager) GetCacheMetrics(ctx context.Context) (*CacheMetrics, error) {
	stats := m.cache.GetStats()
	health := m.cache.Health(ctx)
	
	return &CacheMetrics{
		Stats:           stats,
		Health:          health,
		Enabled:         m.cache.IsEnabled(),
		CollectedAt:     time.Now(),
		EfficiencyScore: m.calculateEfficiencyScore(stats),
	}, nil
}

// CacheMetrics provides comprehensive cache metrics
type CacheMetrics struct {
	Stats           *Stats        `json:"stats"`
	Health          *HealthStatus `json:"health"`
	Enabled         bool          `json:"enabled"`
	CollectedAt     time.Time     `json:"collected_at"`
	EfficiencyScore float64       `json:"efficiency_score"` // 0-100
}

// calculateEfficiencyScore calculates a cache efficiency score
func (m *CacheManager) calculateEfficiencyScore(stats *Stats) float64 {
	if stats.TotalRequests == 0 {
		return 0
	}
	
	// Base score from hit rate
	score := stats.HitRate
	
	// Penalize for errors
	errorRate := float64(stats.Errors) / float64(stats.TotalRequests) * 100
	score -= errorRate * 2 // Errors are weighted more heavily
	
	// Ensure score is between 0 and 100
	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}
	
	return score
}

// CacheStatus provides a simple status check
type CacheStatus struct {
	Enabled         bool      `json:"enabled"`
	Healthy         bool      `json:"healthy"`
	Status          string    `json:"status"`
	HitRate         float64   `json:"hit_rate"`
	TotalRequests   int64     `json:"total_requests"`
	EfficiencyScore float64   `json:"efficiency_score"`
	LastChecked     time.Time `json:"last_checked"`
}

// GetStatus returns a simple cache status
func (m *CacheManager) GetStatus(ctx context.Context) *CacheStatus {
	stats := m.cache.GetStats()
	health := m.cache.Health(ctx)
	
	return &CacheStatus{
		Enabled:         m.cache.IsEnabled(),
		Healthy:         health.Healthy,
		Status:          health.Status,
		HitRate:         stats.HitRate,
		TotalRequests:   stats.TotalRequests,
		EfficiencyScore: m.calculateEfficiencyScore(stats),
		LastChecked:     time.Now(),
	}
}

// BatchCacheOperations provides batch operations for cache
type BatchCacheOperations struct {
	cache Cacher
}

// NewBatchCacheOperations creates a new batch operations manager
func NewBatchCacheOperations(cache Cacher) *BatchCacheOperations {
	return &BatchCacheOperations{
		cache: cache,
	}
}

// BatchGet retrieves multiple cache keys
func (b *BatchCacheOperations) BatchGet(ctx context.Context, keys []string) (map[string]interface{}, error) {
	results := make(map[string]interface{})
	
	for _, key := range keys {
		var value interface{}
		if err := b.cache.Get(ctx, key, &value); err == nil {
			results[key] = value
		}
		// Ignore cache misses and errors for batch operations
	}
	
	return results, nil
}

// BatchSet stores multiple cache entries
func (b *BatchCacheOperations) BatchSet(ctx context.Context, entries map[string]interface{}, ttl time.Duration) error {
	for key, value := range entries {
		if err := b.cache.Set(ctx, key, value, ttl); err != nil && !IsCacheDisabled(err) {
			return err
		}
	}
	return nil
}

// BatchDelete removes multiple cache keys
func (b *BatchCacheOperations) BatchDelete(ctx context.Context, keys []string) error {
	for _, key := range keys {
		if err := b.cache.Delete(ctx, key); err != nil && !IsCacheDisabled(err) {
			return err
		}
	}
	return nil
}