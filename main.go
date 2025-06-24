package main

import (
	"flag"
	"log/slog"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/perschulte/greenweb-api/examples"
	"github.com/perschulte/greenweb-api/internal/carbon"
	"github.com/perschulte/greenweb-api/internal/config"
	"github.com/perschulte/greenweb-api/internal/geolocation"
	"github.com/perschulte/greenweb-api/internal/handlers"
	"github.com/perschulte/greenweb-api/internal/impact"
	"github.com/perschulte/greenweb-api/service"
)

func main() {
	// Parse command line flags
	demo := flag.String("demo", "", "Run demo mode: 'high-impact' for optimization demo")
	flag.Parse()

	// Check if running in demo mode
	if *demo == "high-impact" {
		examples.RunHighImpactOptimizationDemo()
		return
	}

	// Load environment variables and configuration
	godotenv.Load()
	
	// Load centralized configuration
	cfg, err := config.Load()
	if err != nil {
		slog.Error("Failed to load configuration", "error", err)
		os.Exit(1)
	}

	// Initialize structured logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	// Initialize core services with enhanced feedback-based features
	electricityMapsClient := service.NewElectricityMapsClient(logger)
	
	// Initialize carbon intelligence service for dynamic thresholds
	carbonIntelligence := carbon.NewIntelligenceService(electricityMapsClient, logger)
	
	// Initialize dual-grid geolocation service
	geolocationConfig := geolocation.ServiceConfig{
		APIBaseURL:   cfg.App.GeolocationAPIURL,
		Timeout:      5 * time.Second,
		RateLimitRPS: 2,
	}
	dualGridGeoService := geolocation.NewDualGridService(geolocationConfig)
	
	// Initialize impact calculation service for realistic CO2 tracking
	impactStorage := impact.NewMockStorage()
	impactCalculator := impact.NewCalculator()
	impactService := impact.NewService(impactCalculator, impactStorage, logger)

	logger.Info("GreenWeb enhanced services initialized", 
		"electricity_maps_configured", cfg.ElectricityMaps.APIKey != "",
		"carbon_intelligence", "enabled",
		"dual_grid_geolocation", "enabled",
		"impact_calculator", "enabled",
		"environment", cfg.Server.Environment)

	// Create configuration
	config := &handlers.Config{
		Version:           "0.2.0",
		ServiceName:       "greenweb-api",
		ElectricityAPIKey: os.Getenv("ELECTRICITY_MAPS_API_KEY"),
	}

	// Create dependencies container
	deps := &handlers.Dependencies{
		ElectricityMaps: electricityMapsClient,
		Optimization:    nil, // Optimization service disabled for now
		Cache:           nil, // No cache service for now
		Logger:          logger,
		Config:          config,
	}

	// Setup Gin router
	r := gin.Default()

	// Middleware
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(handlers.CORSMiddleware())

	// Register all handlers
	handlers.RegisterHandlers(r, deps, dualGridGeoService)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8090"
	}

	logger.Info("Starting GreenWeb API server", "port", port)
	if err := r.Run(":" + port); err != nil {
		logger.Error("Failed to start server", "error", err)
		os.Exit(1)
	}
}