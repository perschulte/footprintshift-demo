// Package handlers provides HTTP request handlers for the GreenWeb API.
// It implements a modular, dependency-injected architecture with clean separation of concerns.
package handlers

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/perschulte/greenweb-api/internal/geolocation"
	"github.com/perschulte/greenweb-api/pkg/carbon"
)

// ElectricityMapsService defines the interface for electricity maps operations
type ElectricityMapsService interface {
	IsHealthy(ctx context.Context) bool
	GetCarbonIntensity(ctx context.Context, location string) (*carbon.CarbonIntensity, error)
	GetGreenHoursForecast(ctx context.Context, location string, hours int) (*carbon.GreenHoursForecast, error)
	// Dual-grid methods
	GetDualGridCarbonIntensity(ctx context.Context, userLocation, edgeLocation, contentType string) (*carbon.DualGridCarbonIntensity, error)
	GetOptimalEdgeLocation(ctx context.Context, userLocation, cdnProvider, contentType string) (*carbon.EdgeAlternative, error)
	GetCDNAlternatives(ctx context.Context, userLocation, currentEdgeLocation, cdnProvider, contentType string, maxAlternatives int) ([]carbon.EdgeAlternative, error)
}

// OptimizationService defines the interface for optimization operations
type OptimizationService interface {
	// TODO: Define optimization methods when implemented
}

// CarbonIntelligenceService defines the interface for carbon intelligence operations
type CarbonIntelligenceService interface {
	// TODO: Define carbon intelligence methods when implemented
}

// CacheService defines the interface for cache operations (for future implementation)
type CacheService interface {
	Get(key string) (interface{}, bool)
	Set(key string, value interface{}, expiration int64)
	Delete(key string)
	Clear()
	Stats() map[string]interface{}
}

// Config holds application configuration
type Config struct {
	Version          string
	ServiceName      string
	ElectricityAPIKey string
}

// Dependencies holds all services that handlers depend on
type Dependencies struct {
	ElectricityMaps       ElectricityMapsService
	Optimization          OptimizationService
	CarbonIntelligence    CarbonIntelligenceService // Optional, can be nil for fallback
	Cache                 CacheService // Optional, can be nil
	Logger                *slog.Logger
	Config                *Config
}

// ErrorResponse represents a standardized error response
type ErrorResponse struct {
	Error   string            `json:"error"`
	Code    string            `json:"code,omitempty"`
	Details map[string]string `json:"details,omitempty"`
}

// ValidationError represents input validation errors
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Value   string `json:"value,omitempty"`
}

// ValidationErrors holds multiple validation errors
type ValidationErrors struct {
	Errors []ValidationError `json:"errors"`
}

// RespondWithError sends a standardized error response
func RespondWithError(c *gin.Context, statusCode int, message string, code string, details map[string]string) {
	response := ErrorResponse{
		Error:   message,
		Code:    code,
		Details: details,
	}
	c.JSON(statusCode, response)
}

// RespondWithValidationErrors sends validation error response
func RespondWithValidationErrors(c *gin.Context, errors []ValidationError) {
	response := ValidationErrors{
		Errors: errors,
	}
	c.JSON(http.StatusBadRequest, response)
}

// ValidateLocation validates and normalizes location parameter
func ValidateLocation(location string) (string, []ValidationError) {
	var errors []ValidationError
	
	if location == "" {
		errors = append(errors, ValidationError{
			Field:   "location",
			Message: "location parameter is required",
		})
		return "", errors
	}
	
	// Normalize location (trim spaces, capitalize first letter)
	location = strings.TrimSpace(location)
	if len(location) == 0 {
		errors = append(errors, ValidationError{
			Field:   "location",
			Message: "location parameter cannot be empty",
		})
		return "", errors
	}
	
	// Basic length validation
	if len(location) > 100 {
		errors = append(errors, ValidationError{
			Field:   "location",
			Message: "location parameter too long (max 100 characters)",
			Value:   location,
		})
		return "", errors
	}
	
	// Sanitize - remove potentially dangerous characters
	if strings.ContainsAny(location, "<>\"'&") {
		errors = append(errors, ValidationError{
			Field:   "location",
			Message: "location parameter contains invalid characters",
			Value:   location,
		})
		return "", errors
	}
	
	return location, errors
}

// ValidateURL validates URL parameter
func ValidateURL(url string) (string, []ValidationError) {
	var errors []ValidationError
	
	if url == "" {
		return "", errors // URL is optional
	}
	
	url = strings.TrimSpace(url)
	if len(url) > 2048 {
		errors = append(errors, ValidationError{
			Field:   "url",
			Message: "URL parameter too long (max 2048 characters)",
			Value:   url,
		})
		return "", errors
	}
	
	// Basic URL sanitization
	if strings.ContainsAny(url, "<>\"'") {
		errors = append(errors, ValidationError{
			Field:   "url",
			Message: "URL parameter contains invalid characters",
			Value:   url,
		})
		return "", errors
	}
	
	return url, errors
}

// ValidateHours validates hours parameter for forecasts
func ValidateHours(hoursParam string, defaultHours int, maxHours int) (int, []ValidationError) {
	var errors []ValidationError
	
	if hoursParam == "" {
		return defaultHours, errors
	}
	
	hours, err := strconv.Atoi(hoursParam)
	if err != nil {
		errors = append(errors, ValidationError{
			Field:   "hours",
			Message: "hours parameter must be a valid integer",
			Value:   hoursParam,
		})
		return defaultHours, errors
	}
	
	if hours < 1 {
		errors = append(errors, ValidationError{
			Field:   "hours",
			Message: "hours parameter must be at least 1",
			Value:   hoursParam,
		})
		return defaultHours, errors
	}
	
	if hours > maxHours {
		errors = append(errors, ValidationError{
			Field:   "hours",
			Message: "hours parameter exceeds maximum allowed value",
			Value:   hoursParam,
		})
		return maxHours, errors
	}
	
	return hours, errors
}

// ValidatePeriod validates period parameter for trends analysis
func ValidatePeriod(periodParam string) (string, []ValidationError) {
	var errors []ValidationError
	
	if periodParam == "" {
		return "daily", errors // Default to daily
	}
	
	period := strings.ToLower(strings.TrimSpace(periodParam))
	
	validPeriods := map[string]bool{
		"daily":   true,
		"weekly":  true,
		"monthly": true,
	}
	
	if !validPeriods[period] {
		errors = append(errors, ValidationError{
			Field:   "period",
			Message: "period must be one of: daily, weekly, monthly",
			Value:   periodParam,
		})
		return "daily", errors
	}
	
	return period, errors
}

// ValidateDays validates days parameter for historical data requests
func ValidateDays(daysParam string, defaultDays int, maxDays int) (int, []ValidationError) {
	var errors []ValidationError
	
	if daysParam == "" {
		return defaultDays, errors
	}
	
	days, err := strconv.Atoi(daysParam)
	if err != nil {
		errors = append(errors, ValidationError{
			Field:   "days",
			Message: "days parameter must be a valid integer",
			Value:   daysParam,
		})
		return defaultDays, errors
	}
	
	if days < 1 {
		errors = append(errors, ValidationError{
			Field:   "days",
			Message: "days parameter must be at least 1",
			Value:   daysParam,
		})
		return defaultDays, errors
	}
	
	if days > maxDays {
		errors = append(errors, ValidationError{
			Field:   "days",
			Message: "days parameter exceeds maximum allowed value",
			Value:   daysParam,
		})
		return maxDays, errors
	}
	
	return days, errors
}

// LogRequest logs incoming requests with relevant details
func LogRequest(logger *slog.Logger, c *gin.Context, operation string, params map[string]interface{}) {
	logger.Info("handling request",
		"operation", operation,
		"method", c.Request.Method,
		"path", c.Request.URL.Path,
		"params", params,
		"user_agent", c.GetHeader("User-Agent"),
		"remote_addr", c.ClientIP(),
	)
}

// LogResponse logs response details
func LogResponse(logger *slog.Logger, operation string, statusCode int, params map[string]interface{}) {
	level := slog.LevelInfo
	if statusCode >= 400 {
		level = slog.LevelError
	} else if statusCode >= 300 {
		level = slog.LevelWarn
	}
	
	logger.Log(context.Background(), level, "request completed",
		"operation", operation,
		"status_code", statusCode,
		"params", params,
	)
}

// RegisterHandlers registers all handlers with the Gin router
func RegisterHandlers(r *gin.Engine, deps *Dependencies, dualGridGeoService *geolocation.DualGridService) {
	// Create handler instances
	healthHandler := NewHealthHandler(deps)
	carbonHandler := NewCarbonSimpleHandler(deps)
	demoHandler := NewDemoHandler(deps)
	
	// Create dual-grid handler if geolocation service is provided
	var dualGridHandler *DualGridHandler
	if dualGridGeoService != nil {
		dualGridHandler = NewDualGridHandler(deps, dualGridGeoService)
	}
	
	// Register routes
	r.GET("/health", healthHandler.HandleHealthCheck)
	
	// API v1 routes
	v1 := r.Group("/api/v1")
	{
		// Standard carbon intensity endpoints
		v1.GET("/carbon-intensity", carbonHandler.HandleGetCarbonIntensity)
		v1.GET("/green-hours", carbonHandler.HandleGetGreenHours)
		
		
		// Dual-grid endpoints (if available)
		if dualGridHandler != nil {
			dualGrid := v1.Group("/dual-grid")
			{
				dualGrid.GET("/carbon-intensity", dualGridHandler.HandleGetDualGridCarbonIntensity)
				dualGrid.GET("/optimal-edge", dualGridHandler.HandleGetOptimalEdgeLocation)
				dualGrid.GET("/cdn-alternatives", dualGridHandler.HandleGetCDNAlternatives)
				dualGrid.GET("/cdn-providers", dualGridHandler.HandleGetSupportedCDNProviders)
			}
		}
	}
	
	// Demo routes
	r.GET("/demo", demoHandler.HandleDemo)
}

// CORSMiddleware provides CORS support
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		
		c.Next()
	}
}