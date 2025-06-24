// This file demonstrates how to use the config package in your application.
// This is an example file and should not be included in production builds.

//go:build example

package config

import (
	"fmt"
	"log"

	"github.com/redis/go-redis/v9"
)

// ExampleUsage demonstrates how to use the configuration module in your application.
func ExampleUsage() {
	// Load configuration from environment variables
	cfg, err := Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Log the configuration (sensitive data is automatically masked)
	fmt.Printf("Loaded configuration:\n%s\n", cfg)

	// Use configuration values throughout your application
	
	// 1. Server setup
	serverAddr := cfg.GetServerAddress()
	fmt.Printf("Starting server on %s\n", serverAddr)
	
	// 2. Redis client setup
	redisClient := redis.NewClient(&redis.Options{
		Addr:            cfg.Redis.URL[8:], // Remove redis:// prefix
		MaxRetries:      cfg.Redis.MaxRetries,
		MinRetryBackoff: cfg.Redis.MinRetryBackoff,
		MaxRetryBackoff: cfg.Redis.MaxRetryBackoff,
		DialTimeout:     cfg.Redis.DialTimeout,
		ReadTimeout:     cfg.Redis.ReadTimeout,
		WriteTimeout:    cfg.Redis.WriteTimeout,
		PoolSize:        cfg.Redis.PoolSize,
		MinIdleConns:    cfg.Redis.MinIdleConns,
		MaxConnAge:      cfg.Redis.MaxConnAge,
		PoolTimeout:     cfg.Redis.PoolTimeout,
		IdleTimeout:     cfg.Redis.IdleTimeout,
		ConnMaxIdleTime: cfg.Redis.IdleCheckFrequency,
	})
	
	// 3. Environment-specific behavior
	if cfg.IsProduction() {
		fmt.Println("Running in production mode")
		// Enable production-specific features
		// Disable debug endpoints
		// Use production logging levels
	} else if cfg.IsDevelopment() {
		fmt.Println("Running in development mode")
		// Enable development features
		// Allow debug endpoints
		// Use verbose logging
	}
	
	// 4. Feature flags
	if cfg.Features.EnableDemoMode {
		fmt.Println("Demo mode is enabled - using mock data")
		// Use mock data instead of real APIs
	}
	
	// 5. API client setup
	if cfg.ElectricityMaps.APIKey != "" {
		fmt.Println("Electricity Maps API key is configured")
		// Initialize real API client
	} else {
		fmt.Println("No API key configured, using mock data")
		// Use mock implementation
	}
	
	// 6. CORS setup
	fmt.Printf("Allowed CORS origins: %v\n", cfg.Security.AllowedOrigins)
	
	// 7. Cache configuration
	fmt.Printf("Cache TTL: %s\n", cfg.App.CacheTTL)
	
	// 8. Rate limiting
	fmt.Printf("Rate limit: %d requests/minute, burst: %d\n", 
		cfg.App.RateLimit.RequestsPerMinute, 
		cfg.App.RateLimit.BurstSize)

	// Clean up resources
	redisClient.Close()
}

// ExampleIntegrationWithGin shows how to integrate the config with a Gin web server.
func ExampleIntegrationWithGin() {
	// This example shows how you might integrate the config with your existing Gin setup
	
	cfg, err := Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Create Redis client using config
	redisClient := redis.NewClient(&redis.Options{
		Addr:         cfg.Redis.URL[8:], // Remove redis:// prefix
		PoolSize:     cfg.Redis.PoolSize,
		MinIdleConns: cfg.Redis.MinIdleConns,
		// ... other Redis options from config
	})
	defer redisClient.Close()

	// Example middleware that uses config
	fmt.Printf("Setting up CORS middleware with origins: %v\n", cfg.Security.AllowedOrigins)
	
	// Example of conditional behavior based on environment
	if cfg.IsProduction() {
		// Production-specific setup
		fmt.Println("Production mode: enabling security features")
	} else {
		// Development-specific setup
		fmt.Println("Development mode: enabling debug features")
	}
	
	// Start server
	fmt.Printf("Server will start on %s\n", cfg.GetServerAddress())
}

// ExampleConfigValidation shows how to handle configuration errors gracefully.
func ExampleConfigValidation() {
	cfg, err := Load()
	if err != nil {
		// Handle configuration errors gracefully
		fmt.Printf("Configuration error: %v\n", err)
		
		// You might want to:
		// 1. Log the error
		// 2. Show helpful error messages
		// 3. Exit gracefully
		// 4. Fall back to safe defaults (if appropriate)
		
		return
	}
	
	// Configuration is valid, proceed with application startup
	fmt.Println("Configuration loaded successfully")
	fmt.Printf("Environment: %s\n", cfg.Server.Env)
	fmt.Printf("Server: %s\n", cfg.GetServerAddress())
}