package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/perschulte/greenweb-api/pkg/carbon"
)

// ElectricityMapsClient handles integration with Electricity Maps API
type ElectricityMapsClient struct {
	apiKey     string
	httpClient *http.Client
	baseURL    string
	logger     *slog.Logger
}

// NewElectricityMapsClient creates a new Electricity Maps API client
func NewElectricityMapsClient(logger *slog.Logger) *ElectricityMapsClient {
	apiKey := os.Getenv("ELECTRICITY_MAPS_API_KEY")
	
	client := &ElectricityMapsClient{
		apiKey:  apiKey,
		baseURL: "https://api.co2signal.com/v1",
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		logger: logger,
	}
	
	if apiKey == "" {
		logger.Warn("ELECTRICITY_MAPS_API_KEY not set, will use mock data")
	}
	
	return client
}

// ElectricityMapsResponse represents the API response structure
type ElectricityMapsResponse struct {
	CountryCode string `json:"countryCode,omitempty"`
	Zone        string `json:"zone,omitempty"`
	Data        struct {
		CarbonIntensity      float64   `json:"carbonIntensity"`
		DateTime             time.Time `json:"datetime"`
		FossilFuelPercentage float64   `json:"fossilFuelPercentage"`
	} `json:"data"`
	Status string `json:"status"`
	Units  struct {
		CarbonIntensity string `json:"carbonIntensity"`
	} `json:"units"`
}

// CarbonIntensity is an alias for backward compatibility.
// New code should use github.com/perschulte/greenweb-api/pkg/carbon.CarbonIntensity
type CarbonIntensity = carbon.CarbonIntensity

// GreenHour is an alias for backward compatibility.
// New code should use github.com/perschulte/greenweb-api/pkg/carbon.GreenHour
type GreenHour = carbon.GreenHour

// GreenHoursForecast is an alias for backward compatibility.
// New code should use github.com/perschulte/greenweb-api/pkg/carbon.GreenHoursForecast
type GreenHoursForecast = carbon.GreenHoursForecast

// GetCarbonIntensity fetches current carbon intensity for a location
func (c *ElectricityMapsClient) GetCarbonIntensity(ctx context.Context, location string) (*carbon.CarbonIntensity, error) {
	// If no API key, fall back to mock data
	if c.apiKey == "" {
		c.logger.Info("Using mock data due to missing API key", "location", location)
		return c.getMockCarbonIntensity(location), nil
	}

	// Try to map location to country code or zone
	countryCode := c.mapLocationToCountryCode(location)
	
	url := fmt.Sprintf("%s/latest?countryCode=%s", c.baseURL, countryCode)
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		c.logger.Error("Failed to create request", "error", err, "location", location)
		return c.getMockCarbonIntensity(location), nil
	}
	
	req.Header.Set("auth-token", c.apiKey)
	req.Header.Set("User-Agent", "GreenWeb-API/1.0")
	
	c.logger.Info("Fetching carbon intensity from Electricity Maps", "url", url, "location", location)
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Error("API request failed", "error", err, "location", location)
		return c.getMockCarbonIntensity(location), nil
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		c.logger.Error("API returned non-200 status", "status", resp.StatusCode, "location", location)
		return c.getMockCarbonIntensity(location), nil
	}
	
	var apiResp ElectricityMapsResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		c.logger.Error("Failed to decode API response", "error", err, "location", location)
		return c.getMockCarbonIntensity(location), nil
	}
	
	if apiResp.Status != "ok" {
		c.logger.Error("API returned error status", "status", apiResp.Status, "location", location)
		return c.getMockCarbonIntensity(location), nil
	}
	
	c.logger.Info("Successfully fetched carbon intensity", 
		"location", location, 
		"intensity", apiResp.Data.CarbonIntensity,
		"fossil_fuel_percentage", apiResp.Data.FossilFuelPercentage)
	
	return c.convertAPIResponse(location, &apiResp), nil
}

// GetDualGridCarbonIntensity fetches carbon intensity for both user and edge locations
func (c *ElectricityMapsClient) GetDualGridCarbonIntensity(ctx context.Context, userLocation, edgeLocation, contentType string) (*carbon.DualGridCarbonIntensity, error) {
	// Fetch carbon intensity for both locations concurrently
	userChan := make(chan *carbon.CarbonIntensity, 1)
	edgeChan := make(chan *carbon.CarbonIntensity, 1)
	errChan := make(chan error, 2)

	// Fetch user location intensity
	go func() {
		intensity, err := c.GetCarbonIntensity(ctx, userLocation)
		if err != nil {
			errChan <- fmt.Errorf("user location: %w", err)
			return
		}
		userChan <- intensity
	}()

	// Fetch edge location intensity
	go func() {
		intensity, err := c.GetCarbonIntensity(ctx, edgeLocation)
		if err != nil {
			errChan <- fmt.Errorf("edge location: %w", err)
			return
		}
		edgeChan <- intensity
	}()

	// Wait for both results
	var userIntensity, edgeIntensity *carbon.CarbonIntensity
	
	for i := 0; i < 2; i++ {
		select {
		case user := <-userChan:
			userIntensity = user
		case edge := <-edgeChan:
			edgeIntensity = edge
		case err := <-errChan:
			c.logger.Error("Failed to fetch dual grid intensity", "error", err)
			// Return mock data on error
			return c.getMockDualGridIntensity(userLocation, edgeLocation, contentType), nil
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	// Calculate weighted intensity
	weightedIntensity, transmissionWeight, computationWeight := carbon.CalculateWeightedIntensity(
		userIntensity.CarbonIntensity, 
		edgeIntensity.CarbonIntensity, 
		contentType)

	// Create dual grid response
	dual := &carbon.DualGridCarbonIntensity{
		UserLocation:       userIntensity,
		EdgeLocation:       edgeIntensity,
		WeightedIntensity:  weightedIntensity,
		TransmissionWeight: transmissionWeight,
		ComputationWeight:  computationWeight,
		ContentType:        contentType,
		Timestamp:          time.Now(),
	}

	// Generate recommendations
	dual.Recommendation = carbon.GenerateDualGridRecommendation(dual, []carbon.EdgeAlternative{})

	c.logger.Info("Dual grid carbon intensity calculated", 
		"user_location", userLocation,
		"edge_location", edgeLocation,
		"user_intensity", userIntensity.CarbonIntensity,
		"edge_intensity", edgeIntensity.CarbonIntensity,
		"weighted_intensity", weightedIntensity,
		"content_type", contentType)

	return dual, nil
}

// GetOptimalEdgeLocation finds the best edge location from a CDN provider for minimal carbon impact
func (c *ElectricityMapsClient) GetOptimalEdgeLocation(ctx context.Context, userLocation, cdnProvider, contentType string) (*carbon.EdgeAlternative, error) {
	// Get CDN provider configuration
	provider, exists := carbon.GetCDNProvider(cdnProvider)
	if !exists {
		return nil, fmt.Errorf("CDN provider %s not supported", cdnProvider)
	}

	// Parse user location to get coordinates
	// For simplicity, use the default location coordinates if not found
	userLat, userLon := 52.5200, 13.4050 // Default to Berlin

	// Function to get intensity for a grid zone
	getIntensity := func(gridZone string) float64 {
		intensity, err := c.GetCarbonIntensity(ctx, gridZone)
		if err != nil {
			c.logger.Warn("Failed to get intensity for grid zone", "zone", gridZone, "error", err)
			return 300 // Default high value
		}
		return intensity.CarbonIntensity
	}

	// Convert edge locations to slice
	edges := make([]carbon.EdgeLocationInfo, 0, len(provider.EdgeLocations))
	for _, edge := range provider.EdgeLocations {
		edges = append(edges, edge)
	}

	// Get optimal edge location
	optimalEdge := carbon.GetOptimalEdgeLocation(userLat, userLon, edges, getIntensity)
	if optimalEdge == nil {
		return nil, fmt.Errorf("no suitable edge location found")
	}

	// Calculate distance and estimated latency
	distance := carbon.CalculateDistance(userLat, userLon, optimalEdge.Latitude, optimalEdge.Longitude)
	estimatedLatency := int(distance/20) + 10 // Rough estimate: 20km per ms + 10ms base

	return &carbon.EdgeAlternative{
		Location:          optimalEdge.City,
		Provider:          provider.Name,
		CarbonIntensity:   getIntensity(optimalEdge.GridZone),
		Distance:          distance,
		EstimatedLatency:  estimatedLatency,
		AvailabilityScore: 95.0, // Assume high availability for tier 1 locations
	}, nil
}

// GetCDNAlternatives returns alternative edge locations with lower carbon intensity
func (c *ElectricityMapsClient) GetCDNAlternatives(ctx context.Context, userLocation, currentEdgeLocation, cdnProvider, contentType string, maxAlternatives int) ([]carbon.EdgeAlternative, error) {
	provider, exists := carbon.GetCDNProvider(cdnProvider)
	if !exists {
		return nil, fmt.Errorf("CDN provider %s not supported", cdnProvider)
	}

	// Get current edge intensity for comparison
	currentIntensity, err := c.GetCarbonIntensity(ctx, currentEdgeLocation)
	if err != nil {
		currentIntensity = &carbon.CarbonIntensity{CarbonIntensity: 300} // Default high value
	}

	// Default user coordinates (Berlin)
	userLat, userLon := 52.5200, 13.4050

	var alternatives []carbon.EdgeAlternative

	// Function to get intensity for a grid zone
	getIntensity := func(gridZone string) float64 {
		intensity, err := c.GetCarbonIntensity(ctx, gridZone)
		if err != nil {
			return 300 // Default high value
		}
		return intensity.CarbonIntensity
	}

	// Evaluate all edge locations
	for edgeName, edge := range provider.EdgeLocations {
		// Skip current edge location
		if edgeName == currentEdgeLocation || edge.City == currentEdgeLocation {
			continue
		}

		intensity := getIntensity(edge.GridZone)
		
		// Only include if significantly better (at least 20% improvement)
		if intensity < currentIntensity.CarbonIntensity*0.8 {
			distance := carbon.CalculateDistance(userLat, userLon, edge.Latitude, edge.Longitude)
			estimatedLatency := int(distance/20) + 10

			alternatives = append(alternatives, carbon.EdgeAlternative{
				Location:          edge.City,
				Provider:          provider.Name,
				CarbonIntensity:   intensity,
				Distance:          distance,
				EstimatedLatency:  estimatedLatency,
				AvailabilityScore: float64(100 - edge.Tier*5), // Lower tier = higher score
			})
		}
	}

	// Sort by carbon intensity (best first)
	for i := 0; i < len(alternatives)-1; i++ {
		for j := 0; j < len(alternatives)-1-i; j++ {
			if alternatives[j].CarbonIntensity > alternatives[j+1].CarbonIntensity {
				alternatives[j], alternatives[j+1] = alternatives[j+1], alternatives[j]
			}
		}
	}

	// Limit to maxAlternatives
	if len(alternatives) > maxAlternatives {
		alternatives = alternatives[:maxAlternatives]
	}

	return alternatives, nil
}

// GetGreenHoursForecast generates a forecast of optimal low-carbon hours
func (c *ElectricityMapsClient) GetGreenHoursForecast(ctx context.Context, location string, hours int) (*carbon.GreenHoursForecast, error) {
	// Note: The basic Electricity Maps API doesn't provide forecast data
	// This would require their premium API. For now, we'll generate a smart mock
	// based on current carbon intensity and typical patterns
	
	current, err := c.GetCarbonIntensity(ctx, location)
	if err != nil {
		return c.getMockGreenHoursForecast(location), nil
	}
	
	c.logger.Info("Generating green hours forecast based on current data", 
		"location", location, 
		"current_intensity", current.CarbonIntensity)
	
	return c.generateSmartForecast(location, current, hours), nil
}

// convertAPIResponse converts Electricity Maps API response to our format
func (c *ElectricityMapsClient) convertAPIResponse(location string, resp *ElectricityMapsResponse) *carbon.CarbonIntensity {
	renewablePercent := 100.0 - resp.Data.FossilFuelPercentage
	if renewablePercent < 0 {
		renewablePercent = 0
	}
	
	mode, recommendation := c.calculateModeAndRecommendation(resp.Data.CarbonIntensity)
	
	return &carbon.CarbonIntensity{
		Location:         location,
		CarbonIntensity:  resp.Data.CarbonIntensity,
		RenewablePercent: renewablePercent,
		Mode:             mode,
		Recommendation:   recommendation,
		NextGreenWindow:  c.estimateNextGreenWindow(resp.Data.CarbonIntensity),
		Timestamp:        resp.Data.DateTime,
		Source:           "electricity_maps",
		GridZone:         resp.Zone,
	}
}

// calculateModeAndRecommendation determines the mode and recommendation based on carbon intensity
func (c *ElectricityMapsClient) calculateModeAndRecommendation(intensity float64) (string, string) {
	if intensity < 150 {
		return "green", "optimal"
	} else if intensity < 300 {
		return "yellow", "reduce"
	} else {
		return "red", "defer"
	}
}

// estimateNextGreenWindow estimates when the next green window might occur
func (c *ElectricityMapsClient) estimateNextGreenWindow(currentIntensity float64) time.Time {
	now := time.Now()
	
	// Simple heuristic: if current intensity is high, next green window is likely at night
	if currentIntensity > 300 {
		// Next night (10 PM)
		nextNight := time.Date(now.Year(), now.Month(), now.Day(), 22, 0, 0, 0, now.Location())
		if now.Hour() >= 22 {
			nextNight = nextNight.Add(24 * time.Hour)
		}
		return nextNight
	}
	
	// If already green, next window is soon
	if currentIntensity < 150 {
		return now.Add(2 * time.Hour)
	}
	
	// For yellow, next green window is in a few hours
	return now.Add(4 * time.Hour)
}

// generateSmartForecast creates an intelligent forecast based on current conditions and patterns
func (c *ElectricityMapsClient) generateSmartForecast(location string, current *carbon.CarbonIntensity, hours int) *carbon.GreenHoursForecast {
	var greenHours []carbon.GreenHour
	now := time.Now()
	
	// Generate forecast based on typical daily patterns
	for i := 1; i <= hours; i++ {
		futureTime := now.Add(time.Duration(i) * time.Hour)
		hour := futureTime.Hour()
		
		// Estimate carbon intensity based on hour (renewable sources are typically higher at night and midday)
		var estimatedIntensity float64
		var estimatedRenewable float64
		
		if hour >= 22 || hour <= 6 {
			// Night hours - typically more wind power
			estimatedIntensity = current.CarbonIntensity * 0.7
			estimatedRenewable = current.RenewablePercent * 1.3
		} else if hour >= 10 && hour <= 16 {
			// Midday - solar power peak
			estimatedIntensity = current.CarbonIntensity * 0.8
			estimatedRenewable = current.RenewablePercent * 1.2
		} else {
			// Peak hours - more demand, typically higher intensity
			estimatedIntensity = current.CarbonIntensity * 1.2
			estimatedRenewable = current.RenewablePercent * 0.9
		}
		
		// Cap renewable percentage at 100%
		if estimatedRenewable > 100 {
			estimatedRenewable = 100
		}
		
		// Only include as "green hours" if intensity is low enough
		if estimatedIntensity < 200 {
			greenHours = append(greenHours, carbon.GreenHour{
				Start:            futureTime,
				End:              futureTime.Add(time.Hour),
				CarbonIntensity:  estimatedIntensity,
				RenewablePercent: estimatedRenewable,
				Duration:         time.Hour,
			})
		}
	}
	
	forecast := &carbon.GreenHoursForecast{
		Location:   location,
		GreenHours: greenHours,
		ForecastPeriod: struct {
			Start time.Time `json:"start" example:"2024-01-15T14:00:00Z"`
			End   time.Time `json:"end" example:"2024-01-16T14:00:00Z"`
		}{
			Start: now,
			End:   now.Add(time.Duration(hours) * time.Hour),
		},
		GeneratedAt: now,
		Source:     "greenweb_forecast",
		Confidence: 70.0,
	}
	
	// Set best window to the first/lowest intensity green hour
	if len(greenHours) > 0 {
		bestWindow := greenHours[0]
		for _, hour := range greenHours {
			if hour.CarbonIntensity < bestWindow.CarbonIntensity {
				bestWindow = hour
			}
		}
		forecast.BestWindow = bestWindow
	}
	
	return forecast
}

// mapLocationToCountryCode maps common location names to ISO country codes
func (c *ElectricityMapsClient) mapLocationToCountryCode(location string) string {
	location = strings.ToLower(strings.TrimSpace(location))
	
	mapping := map[string]string{
		"berlin":     "DE",
		"germany":    "DE",
		"deutschland": "DE",
		"paris":      "FR",
		"france":     "FR",
		"london":     "GB",
		"uk":         "GB",
		"britain":    "GB",
		"england":    "GB",
		"madrid":     "ES",
		"spain":      "ES",
		"rome":       "IT",
		"italy":      "IT",
		"amsterdam":  "NL",
		"netherlands": "NL",
		"vienna":     "AT",
		"austria":    "AT",
		"stockholm":  "SE",
		"sweden":     "SE",
		"oslo":       "NO",
		"norway":     "NO",
		"copenhagen": "DK",
		"denmark":    "DK",
		"helsinki":   "FI",
		"finland":    "FI",
		"brussels":   "BE",
		"belgium":    "BE",
		"zurich":     "CH",
		"switzerland": "CH",
		"dublin":     "IE",
		"ireland":    "IE",
		"lisbon":     "PT",
		"portugal":   "PT",
		"warsaw":     "PL",
		"poland":     "PL",
		"prague":     "CZ",
		"czech":      "CZ",
		"budapest":   "HU",
		"hungary":    "HU",
		"bucharest":  "RO",
		"romania":    "RO",
		"sofia":      "BG",
		"bulgaria":   "BG",
		"athens":     "GR",
		"greece":     "GR",
		"new york":   "US-NY",
		"california": "US-CA",
		"texas":      "US-TEX",
		"florida":    "US-FLA",
		"toronto":    "CA-ON",
		"vancouver":  "CA-BC",
		"sydney":     "AU-NSW",
		"melbourne":  "AU-VIC",
		"tokyo":      "JP",
		"japan":      "JP",
	}
	
	if code, exists := mapping[location]; exists {
		return code
	}
	
	// Default to Germany if location is not found
	c.logger.Warn("Unknown location, defaulting to DE", "location", location)
	return "DE"
}

// getMockCarbonIntensity provides fallback mock data
func (c *ElectricityMapsClient) getMockCarbonIntensity(location string) *carbon.CarbonIntensity {
	// Simulate different intensities based on time of day
	hour := time.Now().Hour()
	var intensity float64
	var renewable float64
	var mode string
	var recommendation string

	// Night hours - more wind power
	if hour >= 22 || hour <= 6 {
		intensity = 120
		renewable = 75
		mode = "green"
		recommendation = "optimal"
	} else if hour >= 12 && hour <= 16 {
		// Peak hours - more fossil fuels
		intensity = 450
		renewable = 25
		mode = "red"
		recommendation = "defer"
	} else {
		// Normal hours
		intensity = 250
		renewable = 45
		mode = "yellow"
		recommendation = "reduce"
	}

	return &carbon.CarbonIntensity{
		Location:                location,
		CarbonIntensity:         intensity,
		RenewablePercent:        renewable,
		FossilFuelPercentage:    100 - renewable,
		Mode:                    mode,
		Recommendation:          recommendation,
		NextGreenWindow:         time.Now().Add(4 * time.Hour),
		Timestamp:               time.Now(),
		Source:                  "mock",
		GridZone:                c.mapLocationToCountryCode(location),
	}
}

// getMockDualGridIntensity provides fallback mock data for dual grid scenarios
func (c *ElectricityMapsClient) getMockDualGridIntensity(userLocation, edgeLocation, contentType string) *carbon.DualGridCarbonIntensity {
	userIntensity := c.getMockCarbonIntensity(userLocation)
	edgeIntensity := c.getMockCarbonIntensity(edgeLocation)

	// Calculate weighted intensity
	weightedIntensity, transmissionWeight, computationWeight := carbon.CalculateWeightedIntensity(
		userIntensity.CarbonIntensity,
		edgeIntensity.CarbonIntensity,
		contentType)

	dual := &carbon.DualGridCarbonIntensity{
		UserLocation:       userIntensity,
		EdgeLocation:       edgeIntensity,
		WeightedIntensity:  weightedIntensity,
		TransmissionWeight: transmissionWeight,
		ComputationWeight:  computationWeight,
		Distance:           850.5, // Mock distance
		NetworkHops:        carbon.EstimateNetworkHops(850.5),
		ContentType:        contentType,
		Timestamp:          time.Now(),
	}

	// Generate mock alternatives
	alternatives := []carbon.EdgeAlternative{
		{
			Location:          "Stockholm",
			Provider:          "CloudFlare",
			CarbonIntensity:   95.3,
			Distance:          650.2,
			EstimatedLatency:  45,
			AvailabilityScore: 95.5,
		},
	}

	dual.Recommendation = carbon.GenerateDualGridRecommendation(dual, alternatives)

	return dual
}

// getMockGreenHoursForecast provides fallback mock forecast
func (c *ElectricityMapsClient) getMockGreenHoursForecast(location string) *carbon.GreenHoursForecast {
	now := time.Now()
	forecast := &carbon.GreenHoursForecast{
		Location: location,
		GreenHours: []carbon.GreenHour{
			{
				Start:            now.Add(4 * time.Hour),
				End:              now.Add(8 * time.Hour),
				CarbonIntensity:  95,
				RenewablePercent: 85,
				Duration:         4 * time.Hour,
				Confidence:       75.0,
			},
			{
				Start:            now.Add(22 * time.Hour),
				End:              now.Add(26 * time.Hour),
				CarbonIntensity:  110,
				RenewablePercent: 80,
				Duration:         4 * time.Hour,
				Confidence:       70.0,
			},
		},
		ForecastPeriod: struct {
			Start time.Time `json:"start" example:"2024-01-15T14:00:00Z"`
			End   time.Time `json:"end" example:"2024-01-16T14:00:00Z"`
		}{
			Start: now,
			End:   now.Add(24 * time.Hour),
		},
		GeneratedAt: now,
		Source:     "mock",
		Confidence: 60.0,
	}
	
	forecast.BestWindow = forecast.GreenHours[0]
	return forecast
}

// IsHealthy checks if the Electricity Maps API is accessible
func (c *ElectricityMapsClient) IsHealthy(ctx context.Context) bool {
	if c.apiKey == "" {
		return false // API key not configured
	}
	
	// Try a simple request to check connectivity
	_, err := c.GetCarbonIntensity(ctx, "DE")
	return err == nil
}