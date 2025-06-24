package handlers

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/perschulte/greenweb-api/internal/geolocation"
	"github.com/perschulte/greenweb-api/pkg/carbon"
)

// DualGridHandler handles dual-grid carbon intensity and optimization endpoints
type DualGridHandler struct {
	electricityService ElectricityMapsService
	geolocationService *geolocation.DualGridService
	cacheService       CacheService
	logger            *slog.Logger
	config            *Config
}

// NewDualGridHandler creates a new dual-grid handler with dependencies
func NewDualGridHandler(deps *Dependencies, dualGridGeoService *geolocation.DualGridService) *DualGridHandler {
	return &DualGridHandler{
		electricityService: deps.ElectricityMaps,
		geolocationService: dualGridGeoService,
		cacheService:       deps.Cache,
		logger:            deps.Logger,
		config:            deps.Config,
	}
}

// HandleGetDualGridCarbonIntensity retrieves carbon intensity for both user and edge locations
// @Summary Get dual-grid carbon intensity
// @Description Retrieves carbon intensity data for both user location (detected from IP) and edge/server location, providing weighted analysis
// @Tags dual-grid
// @Accept json
// @Produce json
// @Param user_location query string false "User location (auto-detected if not provided)"
// @Param edge_location query string true "Edge/server location"
// @Param content_type query string false "Content type (static, api, video, dynamic, ai, database)" default(static)
// @Param cdn_provider query string false "CDN provider (cloudflare, aws-cloudfront, google-cloud, azure)"
// @Success 200 {object} carbon.DualGridCarbonIntensity
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /v1/dual-grid/carbon-intensity [get]
func (h *DualGridHandler) HandleGetDualGridCarbonIntensity(c *gin.Context) {
	const operation = "get_dual_grid_carbon_intensity"

	// Extract and validate parameters
	userLocationParam := c.Query("user_location")
	edgeLocationParam := c.DefaultQuery("edge_location", "")
	contentTypeParam := c.DefaultQuery("content_type", "static")
	cdnProviderParam := c.Query("cdn_provider")

	// Validate parameters
	var allErrors []ValidationError

	// If user location not provided, detect from IP
	var userLocation string
	if userLocationParam == "" {
		// Detect user location from IP
		dualLocation, err := h.geolocationService.GetDualLocationFromRequest(c.Request.Context(), c.Request, cdnProviderParam)
		if err != nil {
			h.logger.Error("Failed to detect user location", "error", err)
			userLocation = "Berlin" // Default fallback
		} else {
			userLocation = dualLocation.UserLocation.Location.City
			if userLocation == "" {
				userLocation = dualLocation.UserLocation.Location.Country
			}
		}
	} else {
		var locationErrors []string
		userLocation, locationErrors = geolocation.ValidateLocation(userLocationParam)
		for _, err := range locationErrors {
			allErrors = append(allErrors, ValidationError{Field: "user_location", Message: err})
		}
	}

	// Validate edge location
	if edgeLocationParam == "" {
		allErrors = append(allErrors, ValidationError{Field: "edge_location", Message: "edge_location is required"})
	}
	var edgeLocationErrors []string
	edgeLocation, edgeLocationErrors := geolocation.ValidateLocation(edgeLocationParam)
	for _, err := range edgeLocationErrors {
		allErrors = append(allErrors, ValidationError{Field: "edge_location", Message: err})
	}

	// Validate content type
	var contentTypeErrors []string
	contentType, contentTypeErrors := geolocation.ValidateContentType(contentTypeParam)
	for _, err := range contentTypeErrors {
		allErrors = append(allErrors, ValidationError{Field: "content_type", Message: err})
	}

	// Validate CDN provider (optional)
	var cdnProviderErrors []string
	cdnProvider, cdnProviderErrors := geolocation.ValidateCDNProvider(cdnProviderParam)
	for _, err := range cdnProviderErrors {
		allErrors = append(allErrors, ValidationError{Field: "cdn_provider", Message: err})
	}

	if len(allErrors) > 0 {
		RespondWithValidationErrors(c, allErrors)
		LogResponse(h.logger, operation, http.StatusBadRequest, map[string]interface{}{
			"user_location":  userLocationParam,
			"edge_location":  edgeLocationParam,
			"content_type":   contentTypeParam,
			"cdn_provider":   cdnProviderParam,
			"errors":         allErrors,
		})
		return
	}

	// Log the incoming request
	LogRequest(h.logger, c, operation, map[string]interface{}{
		"user_location": userLocation,
		"edge_location": edgeLocation,
		"content_type":  contentType,
		"cdn_provider":  cdnProvider,
	})

	// Create context with timeout
	ctx, cancel := context.WithTimeout(c.Request.Context(), 15*time.Second)
	defer cancel()

	// Try to get from cache first
	var cacheKey string
	if h.cacheService != nil {
		cacheKey = "dual_grid:" + userLocation + ":" + edgeLocation + ":" + contentType
		if cached, found := h.cacheService.Get(cacheKey); found {
			h.logger.Info("serving dual grid data from cache",
				"user_location", userLocation,
				"edge_location", edgeLocation,
				"content_type", contentType)
			LogResponse(h.logger, operation, http.StatusOK, map[string]interface{}{
				"user_location": userLocation,
				"edge_location": edgeLocation,
				"source":        "cache",
			})
			c.JSON(http.StatusOK, cached)
			return
		}
	}

	// Get dual grid carbon intensity
	dualIntensity, err := h.electricityService.GetDualGridCarbonIntensity(ctx, userLocation, edgeLocation, contentType)
	if err != nil {
		h.logger.Error("failed to get dual grid carbon intensity",
			"error", err,
			"user_location", userLocation,
			"edge_location", edgeLocation,
			"content_type", contentType,
			"operation", operation)

		RespondWithError(c, http.StatusInternalServerError,
			"Failed to fetch dual grid carbon intensity",
			"DUAL_GRID_FETCH_ERROR",
			map[string]string{
				"user_location": userLocation,
				"edge_location": edgeLocation,
			})

		LogResponse(h.logger, operation, http.StatusInternalServerError, map[string]interface{}{
			"user_location": userLocation,
			"edge_location": edgeLocation,
			"error":         err.Error(),
		})
		return
	}

	// Get CDN alternatives if provider is specified
	if cdnProvider != "" {
		alternatives, err := h.electricityService.GetCDNAlternatives(ctx, userLocation, edgeLocation, cdnProvider, contentType, 3)
		if err != nil {
			h.logger.Warn("failed to get CDN alternatives", "error", err, "provider", cdnProvider)
		} else {
			dualIntensity.Recommendation.AlternativeEdges = alternatives
		}
	}

	// Cache the result
	if h.cacheService != nil && cacheKey != "" {
		h.cacheService.Set(cacheKey, dualIntensity, 600) // Cache for 10 minutes
	}

	// Log successful response
	h.logger.Info("dual grid carbon intensity retrieved successfully",
		"user_location", userLocation,
		"edge_location", edgeLocation,
		"content_type", contentType,
		"weighted_intensity", dualIntensity.WeightedIntensity,
		"recommendation", dualIntensity.Recommendation.Action,
		"operation", operation)

	LogResponse(h.logger, operation, http.StatusOK, map[string]interface{}{
		"user_location":      userLocation,
		"edge_location":      edgeLocation,
		"weighted_intensity": dualIntensity.WeightedIntensity,
		"recommendation":     dualIntensity.Recommendation.Action,
	})

	c.JSON(http.StatusOK, dualIntensity)
}

// HandleGetOptimalEdgeLocation finds the optimal edge location for a CDN provider
// @Summary Get optimal edge location
// @Description Finds the optimal edge location for a CDN provider based on carbon intensity and distance
// @Tags dual-grid
// @Accept json
// @Produce json
// @Param user_location query string false "User location (auto-detected if not provided)"
// @Param cdn_provider query string true "CDN provider (cloudflare, aws-cloudfront, google-cloud, azure)"
// @Param content_type query string false "Content type" default(static)
// @Success 200 {object} carbon.EdgeAlternative
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /v1/dual-grid/optimal-edge [get]
func (h *DualGridHandler) HandleGetOptimalEdgeLocation(c *gin.Context) {
	const operation = "get_optimal_edge_location"

	// Extract and validate parameters
	userLocationParam := c.Query("user_location")
	cdnProviderParam := c.DefaultQuery("cdn_provider", "")
	contentTypeParam := c.DefaultQuery("content_type", "static")

	// Validate parameters
	var allErrors []ValidationError

	// If user location not provided, detect from IP
	var userLocation string
	if userLocationParam == "" {
		dualLocation, err := h.geolocationService.GetDualLocationFromRequest(c.Request.Context(), c.Request, cdnProviderParam)
		if err != nil {
			h.logger.Error("Failed to detect user location", "error", err)
			userLocation = "Berlin"
		} else {
			userLocation = dualLocation.UserLocation.Location.City
		}
	} else {
		var locationErrors []string
		userLocation, locationErrors = geolocation.ValidateLocation(userLocationParam)
		for _, err := range locationErrors {
			allErrors = append(allErrors, ValidationError{Field: "user_location", Message: err})
		}
	}

	// Validate CDN provider (required)
	if cdnProviderParam == "" {
		allErrors = append(allErrors, ValidationError{Field: "cdn_provider", Message: "cdn_provider is required"})
	}
	var cdnProviderErrors []string
	cdnProvider, cdnProviderErrors := geolocation.ValidateCDNProvider(cdnProviderParam)
	for _, err := range cdnProviderErrors {
		allErrors = append(allErrors, ValidationError{Field: "cdn_provider", Message: err})
	}

	// Validate content type
	var contentTypeErrors []string
	contentType, contentTypeErrors := geolocation.ValidateContentType(contentTypeParam)
	for _, err := range contentTypeErrors {
		allErrors = append(allErrors, ValidationError{Field: "content_type", Message: err})
	}

	if len(allErrors) > 0 {
		RespondWithValidationErrors(c, allErrors)
		return
	}

	// Log the incoming request
	LogRequest(h.logger, c, operation, map[string]interface{}{
		"user_location": userLocation,
		"cdn_provider":  cdnProvider,
		"content_type":  contentType,
	})

	// Create context with timeout
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	// Get optimal edge location
	optimalEdge, err := h.electricityService.GetOptimalEdgeLocation(ctx, userLocation, cdnProvider, contentType)
	if err != nil {
		h.logger.Error("failed to get optimal edge location",
			"error", err,
			"user_location", userLocation,
			"cdn_provider", cdnProvider,
			"operation", operation)

		RespondWithError(c, http.StatusInternalServerError,
			"Failed to find optimal edge location",
			"OPTIMAL_EDGE_ERROR",
			map[string]string{
				"user_location": userLocation,
				"cdn_provider":  cdnProvider,
			})
		return
	}

	// Log successful response
	h.logger.Info("optimal edge location found",
		"user_location", userLocation,
		"cdn_provider", cdnProvider,
		"optimal_location", optimalEdge.Location,
		"carbon_intensity", optimalEdge.CarbonIntensity,
		"operation", operation)

	LogResponse(h.logger, operation, http.StatusOK, map[string]interface{}{
		"user_location":    userLocation,
		"optimal_location": optimalEdge.Location,
		"carbon_intensity": optimalEdge.CarbonIntensity,
	})

	c.JSON(http.StatusOK, optimalEdge)
}

// HandleGetCDNAlternatives returns alternative edge locations with better carbon characteristics
// @Summary Get CDN alternatives
// @Description Returns alternative edge locations for a CDN provider ranked by carbon efficiency
// @Tags dual-grid
// @Accept json
// @Produce json
// @Param user_location query string false "User location (auto-detected if not provided)"
// @Param current_edge query string true "Current edge location"
// @Param cdn_provider query string true "CDN provider"
// @Param content_type query string false "Content type" default(static)
// @Param max_results query int false "Maximum number of alternatives to return" default(5)
// @Success 200 {array} carbon.EdgeAlternative
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /v1/dual-grid/cdn-alternatives [get]
func (h *DualGridHandler) HandleGetCDNAlternatives(c *gin.Context) {
	const operation = "get_cdn_alternatives"

	// Extract and validate parameters
	userLocationParam := c.Query("user_location")
	currentEdgeParam := c.DefaultQuery("current_edge", "")
	cdnProviderParam := c.DefaultQuery("cdn_provider", "")
	contentTypeParam := c.DefaultQuery("content_type", "static")
	maxResultsParam := c.DefaultQuery("max_results", "5")

	// Validate parameters
	var allErrors []ValidationError

	// Parse max results
	maxResults, err := strconv.Atoi(maxResultsParam)
	if err != nil || maxResults < 1 || maxResults > 20 {
		allErrors = append(allErrors, ValidationError{Field: "max_results", Message: "max_results must be between 1 and 20"})
	}

	// Validate required parameters
	if currentEdgeParam == "" {
		allErrors = append(allErrors, ValidationError{Field: "current_edge", Message: "current_edge is required"})
	}
	if cdnProviderParam == "" {
		allErrors = append(allErrors, ValidationError{Field: "cdn_provider", Message: "cdn_provider is required"})
	}

	// If user location not provided, detect from IP
	var userLocation string
	if userLocationParam == "" {
		dualLocation, err := h.geolocationService.GetDualLocationFromRequest(c.Request.Context(), c.Request, cdnProviderParam)
		if err != nil {
			h.logger.Error("Failed to detect user location", "error", err)
			userLocation = "Berlin"
		} else {
			userLocation = dualLocation.UserLocation.Location.City
		}
	} else {
		var locationErrors []string
		userLocation, locationErrors = geolocation.ValidateLocation(userLocationParam)
		for _, err := range locationErrors {
			allErrors = append(allErrors, ValidationError{Field: "user_location", Message: err})
		}
	}

	// Validate CDN provider
	var cdnProviderErrors []string
	cdnProvider, cdnProviderErrors := geolocation.ValidateCDNProvider(cdnProviderParam)
	for _, err := range cdnProviderErrors {
		allErrors = append(allErrors, ValidationError{Field: "cdn_provider", Message: err})
	}

	// Validate content type
	var contentTypeErrors []string
	contentType, contentTypeErrors := geolocation.ValidateContentType(contentTypeParam)
	for _, err := range contentTypeErrors {
		allErrors = append(allErrors, ValidationError{Field: "content_type", Message: err})
	}

	if len(allErrors) > 0 {
		RespondWithValidationErrors(c, allErrors)
		return
	}

	// Log the incoming request
	LogRequest(h.logger, c, operation, map[string]interface{}{
		"user_location": userLocation,
		"current_edge":  currentEdgeParam,
		"cdn_provider":  cdnProvider,
		"content_type":  contentType,
		"max_results":   maxResults,
	})

	// Create context with timeout
	ctx, cancel := context.WithTimeout(c.Request.Context(), 15*time.Second)
	defer cancel()

	// Get CDN alternatives
	alternatives, err := h.electricityService.GetCDNAlternatives(ctx, userLocation, currentEdgeParam, cdnProvider, contentType, maxResults)
	if err != nil {
		h.logger.Error("failed to get CDN alternatives",
			"error", err,
			"user_location", userLocation,
			"current_edge", currentEdgeParam,
			"cdn_provider", cdnProvider,
			"operation", operation)

		RespondWithError(c, http.StatusInternalServerError,
			"Failed to get CDN alternatives",
			"CDN_ALTERNATIVES_ERROR",
			map[string]string{
				"user_location": userLocation,
				"current_edge":  currentEdgeParam,
				"cdn_provider":  cdnProvider,
			})
		return
	}

	// Log successful response
	h.logger.Info("CDN alternatives retrieved",
		"user_location", userLocation,
		"current_edge", currentEdgeParam,
		"cdn_provider", cdnProvider,
		"alternatives_count", len(alternatives),
		"operation", operation)

	LogResponse(h.logger, operation, http.StatusOK, map[string]interface{}{
		"user_location":      userLocation,
		"current_edge":       currentEdgeParam,
		"alternatives_count": len(alternatives),
	})

	c.JSON(http.StatusOK, alternatives)
}

// HandleGetSupportedCDNProviders returns the list of supported CDN providers
// @Summary Get supported CDN providers
// @Description Returns a list of all supported CDN providers and their configurations
// @Tags dual-grid
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /v1/dual-grid/cdn-providers [get]
func (h *DualGridHandler) HandleGetSupportedCDNProviders(c *gin.Context) {
	const operation = "get_supported_cdn_providers"

	LogRequest(h.logger, c, operation, map[string]interface{}{})

	providers := h.geolocationService.GetSupportedCDNProviders()
	
	// Get detailed information about each provider
	providerDetails := make(map[string]interface{})
	for _, providerName := range providers {
		if provider, exists := carbon.GetCDNProvider(providerName); exists {
			providerDetails[providerName] = map[string]interface{}{
				"name":                  provider.Name,
				"edge_locations_count":  len(provider.EdgeLocations),
				"default_edge_selection": provider.DefaultEdgeSelection,
				"carbon_aware_routing":  provider.CarbonAwareRouting,
			}
		}
	}

	response := map[string]interface{}{
		"supported_providers": providers,
		"provider_details":    providerDetails,
		"total_count":         len(providers),
	}

	h.logger.Info("supported CDN providers retrieved",
		"providers_count", len(providers),
		"operation", operation)

	LogResponse(h.logger, operation, http.StatusOK, map[string]interface{}{
		"providers_count": len(providers),
	})

	c.JSON(http.StatusOK, response)
}