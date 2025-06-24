package handlers

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// CarbonHandler handles carbon intensity and green hours endpoints
type CarbonHandler struct {
	electricityService ElectricityMapsService
	cacheService       CacheService
	logger            *slog.Logger
	config            *Config
}

// NewCarbonHandler creates a new carbon handler with dependencies
func NewCarbonHandler(deps *Dependencies) *CarbonHandler {
	return &CarbonHandler{
		electricityService: deps.ElectricityMaps,
		cacheService:       deps.Cache,
		logger:            deps.Logger,
		config:            deps.Config,
	}
}

// HandleGetCarbonIntensity retrieves current carbon intensity for a location
func (h *CarbonHandler) HandleGetCarbonIntensity(c *gin.Context) {
	const operation = "get_carbon_intensity"
	
	// Extract and validate location parameter
	locationParam := c.DefaultQuery("location", "Berlin")
	includeRelative := c.DefaultQuery("relative", "false") == "true"
	
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
		"relative": includeRelative,
	})
	
	// Create context with timeout
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()
	
	// Check if relative metrics are requested and intelligence service is available
	if includeRelative && h.intelligenceService != nil {
		// Try cache for relative data first
		var cacheKey string
		if h.cacheService != nil {
			cacheKey = "carbon_intensity_relative:" + location
			if cached, found := h.cacheService.Get(cacheKey); found {
				h.logger.Info("serving relative carbon intensity from cache", "location", location)
				LogResponse(h.logger, operation, http.StatusOK, map[string]interface{}{
					"location": location,
					"source":   "cache",
					"relative": true,
				})
				c.JSON(http.StatusOK, cached)
				return
			}
		}
		
		// Get relative carbon intensity with intelligence
		relativeIntensity, err := h.intelligenceService.GetRelativeCarbonIntensity(ctx, location)
		if err != nil {
			h.logger.Warn("failed to get relative carbon intensity, falling back to absolute", 
				"error", err, 
				"location", location)
			// Fall back to absolute intensity
		} else {
			// Cache relative result
			if h.cacheService != nil && cacheKey != "" {
				h.cacheService.Set(cacheKey, relativeIntensity, 300) // Cache for 5 minutes
			}
			
			h.logger.Info("relative carbon intensity retrieved successfully", 
				"location", location, 
				"intensity", relativeIntensity.CarbonIntensity,
				"percentile", relativeIntensity.LocalPercentile,
				"mode", relativeIntensity.RelativeMode)
			
			LogResponse(h.logger, operation, http.StatusOK, map[string]interface{}{
				"location":   location,
				"intensity":  relativeIntensity.CarbonIntensity,
				"percentile": relativeIntensity.LocalPercentile,
				"mode":       relativeIntensity.RelativeMode,
			})
			
			c.JSON(http.StatusOK, relativeIntensity)
			return
		}
	}
	
	// Fall back to standard absolute carbon intensity
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
func (h *CarbonHandler) HandleGetGreenHours(c *gin.Context) {
	const operation = "get_green_hours"
	
	// Extract and validate parameters
	locationParam := c.DefaultQuery("location", "Berlin")
	hoursParam := c.DefaultQuery("next", "24")
	useDynamic := c.DefaultQuery("dynamic", "true") == "true"
	
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
		"dynamic":  useDynamic,
	})
	
	// Create context with timeout
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()
	
	// Use dynamic thresholds if intelligence service is available and requested
	if useDynamic && h.intelligenceService != nil {
		// Try cache for dynamic forecast first
		var cacheKey string
		if h.cacheService != nil {
			cacheKey = "green_hours_dynamic:" + location + ":" + hoursParam
			if cached, found := h.cacheService.Get(cacheKey); found {
				h.logger.Info("serving dynamic green hours forecast from cache", 
					"location", location, 
					"hours", hours)
				LogResponse(h.logger, operation, http.StatusOK, map[string]interface{}{
					"location": location,
					"hours":    hours,
					"source":   "cache",
					"dynamic":  true,
				})
				c.JSON(http.StatusOK, cached)
				return
			}
		}
		
		// Get dynamic green hours forecast
		forecast, err := h.intelligenceService.GetDynamicGreenHours(ctx, location, hours)
		if err != nil {
			h.logger.Warn("failed to get dynamic green hours forecast, falling back to standard", 
				"error", err, 
				"location", location)
			// Fall back to standard forecast
		} else {
			// Cache dynamic result
			if h.cacheService != nil && cacheKey != "" {
				h.cacheService.Set(cacheKey, forecast, 1800) // Cache for 30 minutes
			}
			
			h.logger.Info("dynamic green hours forecast generated successfully", 
				"location", location, 
				"hours", hours,
				"green_hours_count", len(forecast.GreenHours))
			
			LogResponse(h.logger, operation, http.StatusOK, map[string]interface{}{
				"location":          location,
				"hours":             hours,
				"green_hours_count": len(forecast.GreenHours),
				"dynamic":           true,
			})
			
			c.JSON(http.StatusOK, forecast)
			return
		}
	}
	
	// Fall back to standard green hours forecast
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

// HandleGetCarbonTrends retrieves historical carbon intensity trends for a location
func (h *CarbonHandler) HandleGetCarbonTrends(c *gin.Context) {
	const operation = "get_carbon_trends"
	
	// Extract and validate parameters
	locationParam := c.DefaultQuery("location", "Berlin")
	periodParam := c.DefaultQuery("period", "daily")
	daysParam := c.DefaultQuery("days", "7")
	
	location, locationErrors := ValidateLocation(locationParam)
	period, periodErrors := ValidatePeriod(periodParam)
	days, daysErrors := ValidateDays(daysParam, 7, 90) // Max 90 days
	
	// Combine validation errors
	var allErrors []ValidationError
	allErrors = append(allErrors, locationErrors...)
	allErrors = append(allErrors, periodErrors...)
	allErrors = append(allErrors, daysErrors...)
	
	if len(allErrors) > 0 {
		RespondWithValidationErrors(c, allErrors)
		LogResponse(h.logger, operation, http.StatusBadRequest, map[string]interface{}{
			"location": locationParam,
			"period":   periodParam,
			"days":     daysParam,
			"errors":   allErrors,
		})
		return
	}
	
	// Log the incoming request
	LogRequest(h.logger, c, operation, map[string]interface{}{
		"location": location,
		"period":   period,
		"days":     days,
	})
	
	// Check if intelligence service is available
	if h.intelligenceService == nil {
		RespondWithError(c, http.StatusServiceUnavailable, 
			"Carbon trends analysis not available", 
			"SERVICE_UNAVAILABLE", 
			map[string]string{
				"feature": "carbon_trends",
			})
		
		LogResponse(h.logger, operation, http.StatusServiceUnavailable, map[string]interface{}{
			"location": location,
			"error":    "intelligence service not available",
		})
		return
	}
	
	// Create context with timeout
	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second) // Longer timeout for historical data
	defer cancel()
	
	// Try to get from cache first if cache service is available
	var cacheKey string
	if h.cacheService != nil {
		cacheKey = "carbon_trends:" + location + ":" + period + ":" + daysParam
		if cached, found := h.cacheService.Get(cacheKey); found {
			h.logger.Info("serving carbon trends from cache", 
				"location", location, 
				"period", period,
				"days", days)
			LogResponse(h.logger, operation, http.StatusOK, map[string]interface{}{
				"location": location,
				"period":   period,
				"days":     days,
				"source":   "cache",
			})
			c.JSON(http.StatusOK, cached)
			return
		}
	}
	
	// Get carbon trends from intelligence service
	trends, err := h.intelligenceService.GetCarbonTrends(ctx, location, period, days)
	if err != nil {
		h.logger.Error("failed to get carbon trends", 
			"error", err, 
			"location", location,
			"period", period,
			"days", days,
			"operation", operation)
		
		RespondWithError(c, http.StatusInternalServerError, 
			"Failed to analyze carbon trends", 
			"TRENDS_ERROR", 
			map[string]string{
				"location": location,
				"period":   period,
				"days":     daysParam,
			})
		
		LogResponse(h.logger, operation, http.StatusInternalServerError, map[string]interface{}{
			"location": location,
			"period":   period,
			"days":     days,
			"error":    err.Error(),
		})
		return
	}
	
	// Cache the result if cache service is available
	if h.cacheService != nil && cacheKey != "" {
		h.cacheService.Set(cacheKey, trends, 3600) // Cache for 1 hour
	}
	
	// Log successful response
	h.logger.Info("carbon trends retrieved successfully", 
		"location", location, 
		"period", period,
		"days", days,
		"average_intensity", trends.AverageIntensity,
		"cleanest_hours", trends.CleanestHours,
		"operation", operation)
	
	LogResponse(h.logger, operation, http.StatusOK, map[string]interface{}{
		"location":         location,
		"period":           period,
		"days":             days,
		"average_intensity": trends.AverageIntensity,
		"data_points":      len(trends.DataPoints),
	})
	
	c.JSON(http.StatusOK, trends)
}