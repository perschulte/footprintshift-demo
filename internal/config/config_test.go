package config

import (
	"os"
	"strings"
	"testing"
	"time"
)

func TestLoad(t *testing.T) {
	// Save original environment
	originalEnv := make(map[string]string)
	envVars := []string{
		"PORT", "HOST", "ENVIRONMENT",
		"ELECTRICITY_MAPS_API_KEY", "ELECTRICITY_MAPS_BASE_URL",
		"REDIS_URL", "REDIS_POOL_SIZE",
		"LOG_LEVEL", "CACHE_TTL_SECONDS",
		"ALLOWED_ORIGINS", "ENABLE_DEMO_MODE",
	}
	
	for _, key := range envVars {
		originalEnv[key] = os.Getenv(key)
		os.Unsetenv(key)
	}
	
	// Restore environment after test
	defer func() {
		for key, value := range originalEnv {
			if value != "" {
				os.Setenv(key, value)
			} else {
				os.Unsetenv(key)
			}
		}
	}()

	t.Run("default values", func(t *testing.T) {
		config, err := Load()
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		// Test server defaults
		if config.Server.Host != "localhost" {
			t.Errorf("Expected host 'localhost', got %s", config.Server.Host)
		}
		if config.Server.Port != 8090 {
			t.Errorf("Expected port 8090, got %d", config.Server.Port)
		}
		if config.Server.Env != "development" {
			t.Errorf("Expected environment 'development', got %s", config.Server.Env)
		}

		// Test Redis defaults
		if config.Redis.URL != "redis://localhost:6379" {
			t.Errorf("Expected Redis URL 'redis://localhost:6379', got %s", config.Redis.URL)
		}
		if config.Redis.PoolSize != 10 {
			t.Errorf("Expected Redis pool size 10, got %d", config.Redis.PoolSize)
		}

		// Test app defaults
		if config.App.LogLevel != "info" {
			t.Errorf("Expected log level 'info', got %s", config.App.LogLevel)
		}
		if config.App.CacheTTL != 300*time.Second {
			t.Errorf("Expected cache TTL 300s, got %s", config.App.CacheTTL)
		}
	})

	t.Run("custom values", func(t *testing.T) {
		// Set custom environment variables
		os.Setenv("PORT", "9000")
		os.Setenv("HOST", "0.0.0.0")
		os.Setenv("ENVIRONMENT", "production")
		os.Setenv("ELECTRICITY_MAPS_API_KEY", "test_api_key")
		os.Setenv("REDIS_POOL_SIZE", "20")
		os.Setenv("LOG_LEVEL", "debug")
		os.Setenv("CACHE_TTL_SECONDS", "600")
		os.Setenv("ALLOWED_ORIGINS", "https://example.com,https://app.example.com")
		os.Setenv("ENABLE_DEMO_MODE", "false")

		config, err := Load()
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		// Test custom server values
		if config.Server.Host != "0.0.0.0" {
			t.Errorf("Expected host '0.0.0.0', got %s", config.Server.Host)
		}
		if config.Server.Port != 9000 {
			t.Errorf("Expected port 9000, got %d", config.Server.Port)
		}
		if config.Server.Env != "production" {
			t.Errorf("Expected environment 'production', got %s", config.Server.Env)
		}

		// Test custom API key
		if config.ElectricityMaps.APIKey != "test_api_key" {
			t.Errorf("Expected API key 'test_api_key', got %s", config.ElectricityMaps.APIKey)
		}

		// Test custom Redis values
		if config.Redis.PoolSize != 20 {
			t.Errorf("Expected Redis pool size 20, got %d", config.Redis.PoolSize)
		}

		// Test custom app values
		if config.App.LogLevel != "debug" {
			t.Errorf("Expected log level 'debug', got %s", config.App.LogLevel)
		}
		if config.App.CacheTTL != 600*time.Second {
			t.Errorf("Expected cache TTL 600s, got %s", config.App.CacheTTL)
		}

		// Test custom CORS origins
		expectedOrigins := []string{"https://example.com", "https://app.example.com"}
		if len(config.Security.AllowedOrigins) != len(expectedOrigins) {
			t.Errorf("Expected %d origins, got %d", len(expectedOrigins), len(config.Security.AllowedOrigins))
		}
		for i, origin := range expectedOrigins {
			if i < len(config.Security.AllowedOrigins) && config.Security.AllowedOrigins[i] != origin {
				t.Errorf("Expected origin %s, got %s", origin, config.Security.AllowedOrigins[i])
			}
		}

		// Test custom feature flags
		if config.Features.EnableDemoMode != false {
			t.Errorf("Expected demo mode false, got %t", config.Features.EnableDemoMode)
		}
	})
}

func TestValidate(t *testing.T) {
	t.Run("valid configuration", func(t *testing.T) {
		config := &Config{
			Server: ServerConfig{
				Host: "localhost",
				Port: 8080,
				Env:  "development",
			},
			ElectricityMaps: ElectricityMapsConfig{
				APIKey:  "test_key",
				BaseURL: "https://api.example.com",
			},
			Redis: RedisConfig{
				URL:                "redis://localhost:6379",
				MaxRetries:         3,
				MinRetryBackoff:    100 * time.Millisecond,
				MaxRetryBackoff:    3000 * time.Millisecond,
				DialTimeout:        5000 * time.Millisecond,
				ReadTimeout:        3000 * time.Millisecond,
				WriteTimeout:       3000 * time.Millisecond,
				PoolSize:           10,
				MinIdleConns:       2,
				MaxConnAge:         30 * time.Minute,
				PoolTimeout:        4000 * time.Millisecond,
				IdleTimeout:        5 * time.Minute,
				IdleCheckFrequency: 1 * time.Minute,
			},
			App: AppConfig{
				LogLevel: "info",
				CacheTTL: 300 * time.Second,
				RateLimit: RateLimitConfig{
					RequestsPerMinute: 100,
					BurstSize:         10,
					CleanupInterval:   5 * time.Minute,
				},
			},
			Security: SecurityConfig{
				AllowedOrigins: []string{"http://localhost:3000"},
			},
			Features: FeatureConfig{
				EnableDemoMode: true,
			},
		}

		if err := config.Validate(); err != nil {
			t.Errorf("Expected no validation error, got %v", err)
		}
	})

	t.Run("invalid port", func(t *testing.T) {
		config := &Config{
			Server: ServerConfig{Port: -1, Env: "development"},
			Redis:  RedisConfig{URL: "redis://localhost:6379", PoolSize: 10, DialTimeout: time.Second, ReadTimeout: time.Second, WriteTimeout: time.Second},
			App:    AppConfig{LogLevel: "info", CacheTTL: time.Second, RateLimit: RateLimitConfig{RequestsPerMinute: 100, BurstSize: 10}},
			Security: SecurityConfig{AllowedOrigins: []string{"http://localhost:3000"}},
		}

		err := config.Validate()
		if err == nil {
			t.Error("Expected validation error for invalid port")
		}
		if !strings.Contains(err.Error(), "port must be between") {
			t.Errorf("Expected port validation error, got %v", err)
		}
	})

	t.Run("production without API key", func(t *testing.T) {
		config := &Config{
			Server: ServerConfig{Port: 8080, Env: "production"},
			ElectricityMaps: ElectricityMapsConfig{APIKey: ""},
			Redis:  RedisConfig{URL: "redis://localhost:6379", PoolSize: 10, DialTimeout: time.Second, ReadTimeout: time.Second, WriteTimeout: time.Second},
			App:    AppConfig{LogLevel: "info", CacheTTL: time.Second, RateLimit: RateLimitConfig{RequestsPerMinute: 100, BurstSize: 10}},
			Security: SecurityConfig{AllowedOrigins: []string{"http://localhost:3000"}},
			Features: FeatureConfig{EnableDemoMode: false},
		}

		err := config.Validate()
		if err == nil {
			t.Error("Expected validation error for missing API key in production")
		}
		if !strings.Contains(err.Error(), "API_KEY is required in production") {
			t.Errorf("Expected API key validation error, got %v", err)
		}
	})

	t.Run("invalid log level", func(t *testing.T) {
		config := &Config{
			Server: ServerConfig{Port: 8080, Env: "development"},
			Redis:  RedisConfig{URL: "redis://localhost:6379", PoolSize: 10, DialTimeout: time.Second, ReadTimeout: time.Second, WriteTimeout: time.Second},
			App:    AppConfig{LogLevel: "invalid", CacheTTL: time.Second, RateLimit: RateLimitConfig{RequestsPerMinute: 100, BurstSize: 10}},
			Security: SecurityConfig{AllowedOrigins: []string{"http://localhost:3000"}},
		}

		err := config.Validate()
		if err == nil {
			t.Error("Expected validation error for invalid log level")
		}
		if !strings.Contains(err.Error(), "log level must be one of") {
			t.Errorf("Expected log level validation error, got %v", err)
		}
	})
}

func TestString(t *testing.T) {
	config := &Config{
		Server: ServerConfig{
			Host: "localhost",
			Port: 8080,
			Env:  "development",
		},
		ElectricityMaps: ElectricityMapsConfig{
			APIKey:  "secret_api_key_1234567890",
			BaseURL: "https://api.example.com",
		},
		Redis: RedisConfig{
			URL:      "redis://user:password@localhost:6379",
			PoolSize: 10,
		},
		App: AppConfig{
			LogLevel: "info",
			CacheTTL: 300 * time.Second,
			RateLimit: RateLimitConfig{
				RequestsPerMinute: 100,
			},
		},
		Security: SecurityConfig{
			AllowedOrigins: []string{"http://localhost:3000"},
		},
		Features: FeatureConfig{
			EnableDemoMode: true,
		},
	}

	result := config.String()

	// Check that sensitive data is masked
	if strings.Contains(result, "secret_api_key_1234567890") {
		t.Error("API key should be masked in string representation")
	}
	if strings.Contains(result, "password") {
		t.Error("Redis password should be masked in string representation")
	}

	// Check that non-sensitive data is present
	if !strings.Contains(result, "localhost") {
		t.Error("Host should be present in string representation")
	}
	if !strings.Contains(result, "8080") {
		t.Error("Port should be present in string representation")
	}
	if !strings.Contains(result, "development") {
		t.Error("Environment should be present in string representation")
	}
}

func TestHelperMethods(t *testing.T) {
	config := &Config{
		Server: ServerConfig{
			Host: "localhost",
			Port: 8080,
			Env:  "production",
		},
	}

	t.Run("IsProduction", func(t *testing.T) {
		if !config.IsProduction() {
			t.Error("Expected IsProduction() to return true")
		}
	})

	t.Run("IsDevelopment", func(t *testing.T) {
		if config.IsDevelopment() {
			t.Error("Expected IsDevelopment() to return false")
		}
	})

	t.Run("GetServerAddress", func(t *testing.T) {
		expected := "localhost:8080"
		if config.GetServerAddress() != expected {
			t.Errorf("Expected server address %s, got %s", expected, config.GetServerAddress())
		}
	})
}

func TestParseStringSlice(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{"", []string{}},
		{"single", []string{"single"}},
		{"one,two,three", []string{"one", "two", "three"}},
		{"one, two , three ", []string{"one", "two", "three"}},
		{"one,,three", []string{"one", "three"}},
		{" , , ", []string{}},
	}

	for _, test := range tests {
		result := parseStringSlice(test.input)
		if len(result) != len(test.expected) {
			t.Errorf("For input %q, expected length %d, got %d", test.input, len(test.expected), len(result))
			continue
		}
		for i, expected := range test.expected {
			if i < len(result) && result[i] != expected {
				t.Errorf("For input %q, expected element %d to be %q, got %q", test.input, i, expected, result[i])
			}
		}
	}
}