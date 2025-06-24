package handlers

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// CarbonSimpleHandler handles basic carbon intensity and green hours endpoints
type CarbonSimpleHandler struct {
	electricityService ElectricityMapsService
	cacheService       CacheService
	logger            *slog.Logger
	config            *Config
}

// NewCarbonSimpleHandler creates a new carbon handler with dependencies
func NewCarbonSimpleHandler(deps *Dependencies) *CarbonSimpleHandler {
	return &CarbonSimpleHandler{
		electricityService: deps.ElectricityMaps,
		cacheService:       deps.Cache,
		logger:            deps.Logger,
		config:            deps.Config,
	}
}

// HandleGetCarbonIntensity retrieves current carbon intensity for a location
func (h *CarbonSimpleHandler) HandleGetCarbonIntensity(c *gin.Context) {
	const operation = "get_carbon_intensity"
	
	// Extract and validate location parameter
	locationParam := c.DefaultQuery("location", "Berlin")
	location, validationErrors := ValidateLocation(locationParam)
	if len(validationErrors) > 0 {
		RespondWithValidationErrors(c, validationErrors)
		LogResponse(h.logger, operation, http.StatusBadRequest, map[string]interface{}{
			"location": locationParam,
			"errors":   validationErrors,
		})
		return
	}
	
	// Log the incoming request
	LogRequest(h.logger, c, operation, map[string]interface{}{
		"location": location,
	})
	
	// Create context with timeout
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()
	
	// Try to get from cache first if cache service is available
	var cacheKey string
	if h.cacheService != nil {
		cacheKey = "carbon_intensity:" + location
		if cached, found := h.cacheService.Get(cacheKey); found {
			h.logger.Info("serving carbon intensity from cache", "location", location)
			LogResponse(h.logger, operation, http.StatusOK, map[string]interface{}{
				"location": location,
				"source":   "cache",
			})
			c.JSON(http.StatusOK, cached)
			return
		}
	}
	
	// Fetch carbon intensity from service
	intensity, err := h.electricityService.GetCarbonIntensity(ctx, location)
	if err != nil {
		h.logger.Error("failed to get carbon intensity", 
			"error", err, 
			"location", location,
			"operation", operation)
		
		RespondWithError(c, http.StatusInternalServerError, 
			"Failed to fetch carbon intensity", 
			"FETCH_ERROR", 
			map[string]string{
				"location": location,
			})
		
		LogResponse(h.logger, operation, http.StatusInternalServerError, map[string]interface{}{
			"location": location,
			"error":    err.Error(),
		})
		return
	}
	
	// Cache the result if cache service is available
	if h.cacheService != nil && cacheKey != "" {
		h.cacheService.Set(cacheKey, intensity, 300) // Cache for 5 minutes
	}
	
	// Log successful response
	h.logger.Info("carbon intensity retrieved successfully", 
		"location", location, 
		"intensity", intensity.CarbonIntensity,
		"mode", intensity.Mode,
		"operation", operation)
	
	LogResponse(h.logger, operation, http.StatusOK, map[string]interface{}{
		"location":  location,
		"intensity": intensity.CarbonIntensity,
		"mode":      intensity.Mode,
	})
	
	c.JSON(http.StatusOK, intensity)
}

// HandleGetGreenHours retrieves green hours forecast for a location
func (h *CarbonSimpleHandler) HandleGetGreenHours(c *gin.Context) {
	const operation = "get_green_hours"
	
	// Extract and validate parameters
	locationParam := c.DefaultQuery("location", "Berlin")
	hoursParam := c.DefaultQuery("next", "24")
	
	location, locationErrors := ValidateLocation(locationParam)
	hours, hoursErrors := ValidateHours(hoursParam, 24, 168) // Max 1 week
	
	// Combine validation errors
	var allErrors []ValidationError
	allErrors = append(allErrors, locationErrors...)
	allErrors = append(allErrors, hoursErrors...)
	
	if len(allErrors) > 0 {
		RespondWithValidationErrors(c, allErrors)
		LogResponse(h.logger, operation, http.StatusBadRequest, map[string]interface{}{
			"location": locationParam,
			"hours":    hoursParam,
			"errors":   allErrors,
		})
		return
	}
	
	// Log the incoming request
	LogRequest(h.logger, c, operation, map[string]interface{}{
		"location": location,
		"hours":    hours,
	})
	
	// Create context with timeout
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()
	
	// Try to get from cache first if cache service is available
	var cacheKey string
	if h.cacheService != nil {
		cacheKey = "green_hours:" + location + ":" + hoursParam
		if cached, found := h.cacheService.Get(cacheKey); found {
			h.logger.Info("serving green hours forecast from cache", 
				"location", location, 
				"hours", hours)
			LogResponse(h.logger, operation, http.StatusOK, map[string]interface{}{
				"location": location,
				"hours":    hours,
				"source":   "cache",
			})
			c.JSON(http.StatusOK, cached)
			return
		}
	}
	
	// Fetch green hours forecast from service
	forecast, err := h.electricityService.GetGreenHoursForecast(ctx, location, hours)
	if err != nil {
		h.logger.Error("failed to get green hours forecast", 
			"error", err, 
			"location", location,
			"hours", hours,
			"operation", operation)
		
		RespondWithError(c, http.StatusInternalServerError, 
			"Failed to generate green hours forecast", 
			"FORECAST_ERROR", 
			map[string]string{
				"location": location,
				"hours":    hoursParam,
			})
		
		LogResponse(h.logger, operation, http.StatusInternalServerError, map[string]interface{}{
			"location": location,
			"hours":    hours,
			"error":    err.Error(),
		})
		return
	}
	
	// Cache the result if cache service is available
	if h.cacheService != nil && cacheKey != "" {
		h.cacheService.Set(cacheKey, forecast, 1800) // Cache for 30 minutes
	}
	
	// Log successful response
	h.logger.Info("green hours forecast generated successfully", 
		"location", location, 
		"hours", hours,
		"green_hours_count", len(forecast.GreenHours),
		"operation", operation)
	
	LogResponse(h.logger, operation, http.StatusOK, map[string]interface{}{
		"location":          location,
		"hours":             hours,
		"green_hours_count": len(forecast.GreenHours),
	})
	
	c.JSON(http.StatusOK, forecast)
}