package geolocation

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/perschulte/greenweb-api/pkg/carbon"
)

// DualGridService provides enhanced geolocation functionality for dual-grid carbon detection
type DualGridService struct {
	*Service                     // Embed base Service
	cdnProviders map[string]bool // Track available CDN providers
}

// NewDualGridService creates a new dual-grid geolocation service
func NewDualGridService(config ServiceConfig) *DualGridService {
	baseService := NewService(config)
	
	// Initialize with available CDN providers
	cdnProviders := make(map[string]bool)
	for _, provider := range carbon.GetAllCDNProviders() {
		cdnProviders[provider] = true
	}

	return &DualGridService{
		Service:      baseService,
		cdnProviders: cdnProviders,
	}
}

// DualLocationResult contains both user and edge location information
type DualLocationResult struct {
	UserLocation LocationWithZone     `json:"user_location"`
	EdgeLocation *EdgeLocationDetails `json:"edge_location,omitempty"`
	Distance     float64              `json:"distance_km,omitempty"`
	NetworkHops  int                  `json:"network_hops,omitempty"`
}

// EdgeLocationDetails provides detailed information about an edge location
type EdgeLocationDetails struct {
	LocationInfo carbon.EdgeLocationInfo `json:"location_info"`
	GridZone     GridZone                `json:"grid_zone"`
	Provider     string                  `json:"provider"`
	IsOptimal    bool                    `json:"is_optimal"`
}

// GetDualLocationFromRequest extracts user location and finds optimal edge location
func (s *DualGridService) GetDualLocationFromRequest(ctx context.Context, r *http.Request, cdnProvider string) (*DualLocationResult, error) {
	// Get user location from IP
	userLocation, err := s.GetLocationFromRequest(ctx, r)
	if err != nil {
		log.Printf("Failed to get user location: %v", err)
		// Use default location on error
		userLocation = s.getDefaultLocationWithZone()
	}

	result := &DualLocationResult{
		UserLocation: userLocation,
	}

	// If CDN provider is specified, find optimal edge location
	if cdnProvider != "" && s.cdnProviders[cdnProvider] {
		edgeDetails, err := s.getOptimalEdgeLocation(userLocation.Location, cdnProvider)
		if err != nil {
			log.Printf("Failed to get optimal edge location: %v", err)
			return result, nil // Return user location only
		}

		result.EdgeLocation = edgeDetails
		
		// Calculate distance and network hops
		if edgeDetails != nil {
			result.Distance = carbon.CalculateDistance(
				userLocation.Location.Latitude,
				userLocation.Location.Longitude,
				edgeDetails.LocationInfo.Latitude,
				edgeDetails.LocationInfo.Longitude,
			)
			result.NetworkHops = carbon.EstimateNetworkHops(result.Distance)
		}
	}

	return result, nil
}

// GetNearestCDNEdge finds the nearest edge location for a given CDN provider
func (s *DualGridService) GetNearestCDNEdge(userLocation Location, cdnProvider string) (*EdgeLocationDetails, error) {
	if !s.cdnProviders[cdnProvider] {
		return nil, fmt.Errorf("CDN provider %s not supported", cdnProvider)
	}

	edgeInfo, distance := carbon.FindNearestEdgeLocation(
		cdnProvider,
		userLocation.Latitude,
		userLocation.Longitude,
	)

	if edgeInfo == nil {
		return nil, fmt.Errorf("no edge locations found for provider %s", cdnProvider)
	}

	// Map edge location to grid zone
	gridZone := s.gridMapper.MapToGridZone(Location{
		Country:     edgeInfo.Country,
		CountryCode: s.getCountryCodeFromGridZone(edgeInfo.GridZone),
		Region:      edgeInfo.City,
		City:        edgeInfo.City,
		Latitude:    edgeInfo.Latitude,
		Longitude:   edgeInfo.Longitude,
	})

	return &EdgeLocationDetails{
		LocationInfo: *edgeInfo,
		GridZone:     gridZone,
		Provider:     cdnProvider,
		IsOptimal:    distance < 500, // Consider optimal if within 500km
	}, nil
}

// GetOptimalCDNEdgeForCarbon finds the edge location with the lowest carbon footprint
func (s *DualGridService) GetOptimalCDNEdgeForCarbon(userLocation Location, cdnProvider string, getCarbonIntensity func(gridZone string) float64) (*EdgeLocationDetails, error) {
	if !s.cdnProviders[cdnProvider] {
		return nil, fmt.Errorf("CDN provider %s not supported", cdnProvider)
	}

	provider, exists := carbon.GetCDNProvider(cdnProvider)
	if !exists {
		return nil, fmt.Errorf("CDN provider configuration not found")
	}

	// Convert edge locations to slice
	edges := make([]carbon.EdgeLocationInfo, 0, len(provider.EdgeLocations))
	for _, edge := range provider.EdgeLocations {
		edges = append(edges, edge)
	}

	// Get optimal edge based on carbon intensity
	optimalEdge := carbon.GetOptimalEdgeLocation(
		userLocation.Latitude,
		userLocation.Longitude,
		edges,
		getCarbonIntensity,
	)

	if optimalEdge == nil {
		return nil, fmt.Errorf("no optimal edge location found")
	}

	// Map to grid zone
	gridZone := s.gridMapper.MapToGridZone(Location{
		Country:     optimalEdge.Country,
		CountryCode: s.getCountryCodeFromGridZone(optimalEdge.GridZone),
		Region:      optimalEdge.City,
		City:        optimalEdge.City,
		Latitude:    optimalEdge.Latitude,
		Longitude:   optimalEdge.Longitude,
	})

	return &EdgeLocationDetails{
		LocationInfo: *optimalEdge,
		GridZone:     gridZone,
		Provider:     cdnProvider,
		IsOptimal:    true, // This is the optimal choice
	}, nil
}

// GetAllCDNAlternatives returns all available edge locations for a CDN provider, ranked by carbon efficiency
func (s *DualGridService) GetAllCDNAlternatives(userLocation Location, cdnProvider string, getCarbonIntensity func(gridZone string) float64, maxResults int) ([]EdgeLocationDetails, error) {
	if !s.cdnProviders[cdnProvider] {
		return nil, fmt.Errorf("CDN provider %s not supported", cdnProvider)
	}

	provider, exists := carbon.GetCDNProvider(cdnProvider)
	if !exists {
		return nil, fmt.Errorf("CDN provider configuration not found")
	}

	type edgeScore struct {
		edge  carbon.EdgeLocationInfo
		score float64
	}

	var scoredEdges []edgeScore

	// Score all edge locations
	for _, edge := range provider.EdgeLocations {
		// Calculate distance factor
		distance := carbon.CalculateDistance(
			userLocation.Latitude,
			userLocation.Longitude,
			edge.Latitude,
			edge.Longitude,
		)

		// Get carbon intensity
		intensity := getCarbonIntensity(edge.GridZone)

		// Calculate composite score (lower is better)
		// 70% carbon intensity, 20% distance, 10% tier priority
		distanceFactor := distance / 1000.0 // Normalize
		tierFactor := float64(edge.Tier)
		score := (intensity * 0.7) + (distanceFactor * 100 * 0.2) + (tierFactor * 10 * 0.1)

		// Apply renewable commitment bonus
		if edge.RenewableCommitment {
			score *= 0.9
		}

		scoredEdges = append(scoredEdges, edgeScore{
			edge:  edge,
			score: score,
		})
	}

	// Sort by score (ascending - lower is better)
	for i := 0; i < len(scoredEdges)-1; i++ {
		for j := 0; j < len(scoredEdges)-1-i; j++ {
			if scoredEdges[j].score > scoredEdges[j+1].score {
				scoredEdges[j], scoredEdges[j+1] = scoredEdges[j+1], scoredEdges[j]
			}
		}
	}

	// Convert to EdgeLocationDetails
	var results []EdgeLocationDetails
	limit := maxResults
	if limit > len(scoredEdges) {
		limit = len(scoredEdges)
	}

	for i := 0; i < limit; i++ {
		edge := scoredEdges[i].edge
		
		gridZone := s.gridMapper.MapToGridZone(Location{
			Country:     edge.Country,
			CountryCode: s.getCountryCodeFromGridZone(edge.GridZone),
			Region:      edge.City,
			City:        edge.City,
			Latitude:    edge.Latitude,
			Longitude:   edge.Longitude,
		})

		results = append(results, EdgeLocationDetails{
			LocationInfo: edge,
			GridZone:     gridZone,
			Provider:     cdnProvider,
			IsOptimal:    i == 0, // First one is optimal
		})
	}

	return results, nil
}

// GetSupportedCDNProviders returns a list of supported CDN providers
func (s *DualGridService) GetSupportedCDNProviders() []string {
	providers := make([]string, 0, len(s.cdnProviders))
	for provider := range s.cdnProviders {
		providers = append(providers, provider)
	}
	return providers
}

// getOptimalEdgeLocation is a helper method to find the optimal edge location
func (s *DualGridService) getOptimalEdgeLocation(userLocation Location, cdnProvider string) (*EdgeLocationDetails, error) {
	// For now, return the nearest edge location
	// In production, this would integrate with carbon intensity data
	return s.GetNearestCDNEdge(userLocation, cdnProvider)
}

// getCountryCodeFromGridZone attempts to extract country code from grid zone
func (s *DualGridService) getCountryCodeFromGridZone(gridZone string) string {
	// Simple mapping - in production this would be more sophisticated
	if len(gridZone) >= 2 {
		return gridZone[:2]
	}
	return "DE" // Default fallback
}

// ValidateLocation validates and normalizes location parameters for dual-grid scenarios
func ValidateLocation(location string) (string, []string) {
	var errors []string
	
	if location == "" {
		errors = append(errors, "location cannot be empty")
		return "", errors
	}
	
	// Add any additional location validation logic here
	
	return location, errors
}

// ValidateCDNProvider validates the CDN provider parameter
func ValidateCDNProvider(provider string) (string, []string) {
	var errors []string
	
	if provider == "" {
		return provider, errors // CDN provider is optional
	}
	
	supportedProviders := carbon.GetAllCDNProviders()
	providerValid := false
	for _, supported := range supportedProviders {
		if provider == supported {
			providerValid = true
			break
		}
	}
	
	if !providerValid {
		errors = append(errors, fmt.Sprintf("unsupported CDN provider: %s. Supported providers: %v", provider, supportedProviders))
	}
	
	return provider, errors
}

// ValidateContentType validates the content type parameter
func ValidateContentType(contentType string) (string, []string) {
	var errors []string
	
	if contentType == "" {
		return "static", errors // Default to static content
	}
	
	validTypes := []string{"static", "api", "video", "dynamic", "ai", "database"}
	typeValid := false
	for _, valid := range validTypes {
		if contentType == valid {
			typeValid = true
			break
		}
	}
	
	if !typeValid {
		errors = append(errors, fmt.Sprintf("invalid content type: %s. Valid types: %v", contentType, validTypes))
	}
	
	return contentType, errors
}