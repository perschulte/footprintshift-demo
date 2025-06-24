package handlers

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// HealthHandler handles health check endpoints
type HealthHandler struct {
	electricityService ElectricityMapsService
	logger            *slog.Logger
	config            *Config
}

// NewHealthHandler creates a new health handler with dependencies
func NewHealthHandler(deps *Dependencies) *HealthHandler {
	return &HealthHandler{
		electricityService: deps.ElectricityMaps,
		logger:            deps.Logger,
		config:            deps.Config,
	}
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status              string    `json:"status"`
	Service             string    `json:"service"`
	Version             string    `json:"version"`
	Timestamp           time.Time `json:"timestamp"`
	ElectricityMapsAPI  bool      `json:"electricity_maps_api"`
	APIKeyConfigured    bool      `json:"api_key_configured"`
}

// HandleHealthCheck performs a comprehensive health check
func (h *HealthHandler) HandleHealthCheck(c *gin.Context) {
	const operation = "health_check"
	
	// Log the incoming request
	LogRequest(h.logger, c, operation, nil)
	
	// Create context with timeout for health checks
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	
	// Check if Electricity Maps API is healthy
	apiHealthy := h.electricityService.IsHealthy(ctx)
	apiKeyConfigured := h.config.ElectricityAPIKey != ""
	
	// Prepare health response
	health := HealthResponse{
		Status:              "healthy",
		Service:             h.config.ServiceName,
		Version:             h.config.Version,
		Timestamp:           time.Now(),
		ElectricityMapsAPI:  apiHealthy,
		APIKeyConfigured:    apiKeyConfigured,
	}
	
	// Log warning if API is not available
	if !apiHealthy {
		h.logger.Warn("health check warning", 
			"message", "Electricity Maps API not available, using mock data",
			"api_key_configured", apiKeyConfigured)
	}
	
	// Log successful response
	LogResponse(h.logger, operation, http.StatusOK, map[string]interface{}{
		"electricity_maps_healthy": apiHealthy,
		"api_key_configured":      apiKeyConfigured,
	})
	
	c.JSON(http.StatusOK, health)
}