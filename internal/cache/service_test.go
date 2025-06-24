package cache

import (
	"context"
	"testing"
	"time"
)

func TestCacheService_BasicOperations(t *testing.T) {
	// Create a test cache service
	config := DefaultConfig()
	config.URL = "redis://localhost:6379/1" // Use test database
	
	cache := New(config)
	defer cache.Close()
	
	// Skip test if Redis is not available
	if !cache.IsEnabled() {
		t.Skip("Redis not available for testing")
	}
	
	ctx := context.Background()
	
	// Test Set operation
	key := "test-key"
	value := "test-value"
	ttl := 1 * time.Minute
	
	err := cache.Set(ctx, key, value, ttl)
	if err != nil {
		t.Fatalf("Failed to set cache value: %v", err)
	}
	
	// Test Get operation
	var retrieved string
	err = cache.Get(ctx, key, &retrieved)
	if err != nil {
		t.Fatalf("Failed to get cache value: %v", err)
	}
	
	if retrieved != value {
		t.Errorf("Expected %s, got %s", value, retrieved)
	}
	
	// Test cache miss
	var missValue string
	err = cache.Get(ctx, "non-existent-key", &missValue)
	if !IsCacheMiss(err) {
		t.Errorf("Expected cache miss error, got: %v", err)
	}
	
	// Test Delete operation
	err = cache.Delete(ctx, key)
	if err != nil {
		t.Fatalf("Failed to delete cache key: %v", err)
	}
	
	// Verify deletion
	err = cache.Get(ctx, key, &retrieved)
	if !IsCacheMiss(err) {
		t.Errorf("Expected cache miss after deletion, got: %v", err)
	}
}

func TestCacheService_GetOrSet(t *testing.T) {
	config := DefaultConfig()
	config.URL = "redis://localhost:6379/1"
	
	cache := New(config)
	defer cache.Close()
	
	if !cache.IsEnabled() {
		t.Skip("Redis not available for testing")
	}
	
	ctx := context.Background()
	key := "test-get-or-set"
	expectedValue := "computed-value"
	
	// Test cache miss - function should be called
	functionCalled := false
	var result string
	
	err := cache.GetOrSet(ctx, key, &result, time.Minute, func() (interface{}, error) {
		functionCalled = true
		return expectedValue, nil
	})
	
	if err != nil {
		t.Fatalf("GetOrSet failed: %v", err)
	}
	
	if !functionCalled {
		t.Error("Function should have been called on cache miss")
	}
	
	if result != expectedValue {
		t.Errorf("Expected %s, got %s", expectedValue, result)
	}
	
	// Test cache hit - function should not be called
	functionCalled = false
	var result2 string
	
	err = cache.GetOrSet(ctx, key, &result2, time.Minute, func() (interface{}, error) {
		functionCalled = true
		return "should-not-be-returned", nil
	})
	
	if err != nil {
		t.Fatalf("GetOrSet failed on cache hit: %v", err)
	}
	
	if functionCalled {
		t.Error("Function should not have been called on cache hit")
	}
	
	if result2 != expectedValue {
		t.Errorf("Expected cached value %s, got %s", expectedValue, result2)
	}
}

func TestCacheService_Stats(t *testing.T) {
	config := DefaultConfig()
	config.URL = "redis://localhost:6379/1"
	
	cache := New(config)
	defer cache.Close()
	
	if !cache.IsEnabled() {
		t.Skip("Redis not available for testing")
	}
	
	ctx := context.Background()
	
	// Reset stats
	cache.ResetStats()
	
	// Perform some operations
	cache.Set(ctx, "key1", "value1", time.Minute)
	cache.Set(ctx, "key2", "value2", time.Minute)
	
	var value string
	cache.Get(ctx, "key1", &value) // Hit
	cache.Get(ctx, "key2", &value) // Hit
	cache.Get(ctx, "key3", &value) // Miss
	
	stats := cache.GetStats()
	
	if stats.Sets != 2 {
		t.Errorf("Expected 2 sets, got %d", stats.Sets)
	}
	
	if stats.Hits != 2 {
		t.Errorf("Expected 2 hits, got %d", stats.Hits)
	}
	
	if stats.Misses != 1 {
		t.Errorf("Expected 1 miss, got %d", stats.Misses)
	}
	
	if stats.TotalRequests != 3 {
		t.Errorf("Expected 3 total requests, got %d", stats.TotalRequests)
	}
	
	expectedHitRate := float64(2) / float64(3) * 100
	if stats.HitRate != expectedHitRate {
		t.Errorf("Expected hit rate %.2f, got %.2f", expectedHitRate, stats.HitRate)
	}
}

func TestCacheService_Health(t *testing.T) {
	config := DefaultConfig()
	config.URL = "redis://localhost:6379/1"
	
	cache := New(config)
	defer cache.Close()
	
	ctx := context.Background()
	
	health := cache.Health(ctx)
	
	if cache.IsEnabled() {
		if !health.Healthy {
			t.Error("Health should be healthy when cache is enabled")
		}
		
		if health.Status != "connected" {
			t.Errorf("Expected status 'connected', got %s", health.Status)
		}
		
		if health.ResponseTime <= 0 {
			t.Error("Response time should be positive")
		}
	} else {
		if health.Healthy {
			t.Error("Health should not be healthy when cache is disabled")
		}
		
		if health.Status != "disconnected" {
			t.Errorf("Expected status 'disconnected', got %s", health.Status)
		}
	}
}

func TestCacheService_KeyGeneration(t *testing.T) {
	config := DefaultConfig()
	cache := New(config)
	
	tests := []struct {
		prefix   string
		parts    []string
		expected string
	}{
		{
			prefix:   "carbon_intensity",
			parts:    []string{"Berlin"},
			expected: "greenweb:carbon_intensity:Berlin",
		},
		{
			prefix:   "optimization",
			parts:    []string{"Berlin", "example.com"},
			expected: "greenweb:optimization:Berlin:example.com",
		},
		{
			prefix:   "green_hours",
			parts:    []string{"Berlin", "24"},
			expected: "greenweb:green_hours:Berlin:24",
		},
		{
			prefix:   "test",
			parts:    []string{"", "valid", ""},
			expected: "greenweb:test:valid",
		},
	}
	
	for _, test := range tests {
		result := cache.GenerateKey(test.prefix, test.parts...)
		if result != test.expected {
			t.Errorf("Expected key %s, got %s", test.expected, result)
		}
	}
}

func TestCacheService_SpecificKeys(t *testing.T) {
	config := DefaultConfig()
	cache := New(config)
	
	// Test specific key generation methods
	carbonKey := cache.GetCarbonIntensityKey("Berlin")
	expected := "greenweb:carbon_intensity:Berlin"
	if carbonKey != expected {
		t.Errorf("Expected carbon key %s, got %s", expected, carbonKey)
	}
	
	optKey := cache.GetOptimizationKey("Berlin", "example.com")
	expected = "greenweb:optimization:Berlin:example.com"
	if optKey != expected {
		t.Errorf("Expected optimization key %s, got %s", expected, optKey)
	}
	
	greenKey := cache.GetGreenHoursKey("Berlin", 24)
	expected = "greenweb:green_hours:Berlin:24"
	if greenKey != expected {
		t.Errorf("Expected green hours key %s, got %s", expected, greenKey)
	}
}

func TestCacheService_GracefulDegradation(t *testing.T) {
	// Test with graceful degradation enabled
	config := DefaultConfig()
	config.URL = "redis://invalid-host:6379"
	config.GracefulDegradation = true
	
	cache := New(config)
	defer cache.Close()
	
	if cache.IsEnabled() {
		t.Skip("Test requires Redis to be unavailable")
	}
	
	ctx := context.Background()
	
	// Set should not return error with graceful degradation
	err := cache.Set(ctx, "key", "value", time.Minute)
	if err != nil {
		t.Errorf("Set should not error with graceful degradation: %v", err)
	}
	
	// Get should return cache disabled error
	var value string
	err = cache.Get(ctx, "key", &value)
	if !IsCacheDisabled(err) {
		t.Errorf("Get should return cache disabled error: %v", err)
	}
	
	// Delete should not return error with graceful degradation
	err = cache.Delete(ctx, "key")
	if err != nil {
		t.Errorf("Delete should not error with graceful degradation: %v", err)
	}
}

func TestCacheService_ErrorTypes(t *testing.T) {
	// Test error type checking functions
	cacheMissErr := NewCacheError(ErrorTypeMiss, "cache miss", nil)
	cacheDisabledErr := NewCacheError(ErrorTypeDisabled, "cache disabled", nil)
	connectionErr := NewCacheError(ErrorTypeConnection, "connection error", nil)
	
	if !IsCacheMiss(cacheMissErr) {
		t.Error("Should detect cache miss error")
	}
	
	if IsCacheMiss(cacheDisabledErr) {
		t.Error("Should not detect cache miss for disabled error")
	}
	
	if !IsCacheDisabled(cacheDisabledErr) {
		t.Error("Should detect cache disabled error")
	}
	
	if IsCacheDisabled(connectionErr) {
		t.Error("Should not detect cache disabled for connection error")
	}
}

func TestConfig_Validation(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name:    "valid config",
			config:  DefaultConfig(),
			wantErr: false,
		},
		{
			name: "empty URL",
			config: &Config{
				URL:      "",
				PoolSize: 10,
			},
			wantErr: true,
		},
		{
			name: "negative MaxRetries",
			config: &Config{
				URL:        "redis://localhost:6379",
				MaxRetries: -1,
				PoolSize:   10,
			},
			wantErr: true,
		},
		{
			name: "zero PoolSize",
			config: &Config{
				URL:      "redis://localhost:6379",
				PoolSize: 0,
			},
			wantErr: true,
		},
		{
			name: "MinIdleConns exceeds PoolSize",
			config: &Config{
				URL:          "redis://localhost:6379",
				PoolSize:     5,
				MinIdleConns: 10,
			},
			wantErr: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Config.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func BenchmarkCacheService_SetGet(b *testing.B) {
	config := DefaultConfig()
	config.URL = "redis://localhost:6379/1"
	
	cache := New(config)
	defer cache.Close()
	
	if !cache.IsEnabled() {
		b.Skip("Redis not available for benchmarking")
	}
	
	ctx := context.Background()
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		key := "bench-key"
		value := "bench-value"
		
		// Set
		cache.Set(ctx, key, value, time.Minute)
		
		// Get
		var retrieved string
		cache.Get(ctx, key, &retrieved)
	}
}

func BenchmarkCacheService_GetOrSet(b *testing.B) {
	config := DefaultConfig()
	config.URL = "redis://localhost:6379/1"
	
	cache := New(config)
	defer cache.Close()
	
	if !cache.IsEnabled() {
		b.Skip("Redis not available for benchmarking")
	}
	
	ctx := context.Background()
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		key := "bench-get-or-set"
		
		var result string
		cache.GetOrSet(ctx, key, &result, time.Minute, func() (interface{}, error) {
			return "computed-value", nil
		})
	}
}