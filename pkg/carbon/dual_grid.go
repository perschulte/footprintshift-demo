// Package carbon provides types and interfaces for carbon intensity monitoring and green energy forecasting.
//
// This file implements dual-grid carbon detection functionality to handle scenarios where
// CDN edge locations are in different carbon zones than users.
package carbon

import (
	"fmt"
	"math"
	"time"
)

// DualGridCarbonIntensity represents carbon intensity data for both user and edge locations.
//
// This structure captures the carbon footprint of content delivery by considering both:
// - The user's location (where energy is consumed for receiving/displaying content)
// - The edge/server location (where energy is consumed for processing/serving content)
//
// The weighted intensity provides a combined view based on the relative carbon costs
// of transmission vs computation.
type DualGridCarbonIntensity struct {
	// UserLocation contains carbon intensity data for the user's location
	UserLocation *CarbonIntensity `json:"user_location" validate:"required"`

	// EdgeLocation contains carbon intensity data for the CDN edge/server location
	EdgeLocation *CarbonIntensity `json:"edge_location" validate:"required"`

	// WeightedIntensity is the combined carbon intensity considering both locations
	WeightedIntensity float64 `json:"weighted_intensity" validate:"min=0" example:"125.5"`

	// TransmissionWeight represents the percentage of carbon footprint from transmission (0-100)
	TransmissionWeight float64 `json:"transmission_weight" validate:"min=0,max=100" example:"30"`

	// ComputationWeight represents the percentage of carbon footprint from computation (0-100)
	ComputationWeight float64 `json:"computation_weight" validate:"min=0,max=100" example:"70"`

	// Distance is the approximate distance between user and edge locations in kilometers
	Distance float64 `json:"distance_km,omitempty" example:"850.5"`

	// NetworkHops estimates the number of network hops between locations
	NetworkHops int `json:"network_hops,omitempty" example:"8"`

	// Recommendation provides optimization guidance based on dual-grid analysis
	Recommendation DualGridRecommendation `json:"recommendation"`

	// ContentType indicates the type of content being served (affects weight calculation)
	ContentType string `json:"content_type,omitempty" example:"video"`

	// Timestamp indicates when this analysis was performed
	Timestamp time.Time `json:"timestamp" validate:"required" example:"2024-01-15T14:30:00Z"`
}

// DualGridRecommendation provides specific recommendations for dual-grid scenarios.
type DualGridRecommendation struct {
	// Action is the recommended action: "proceed", "optimize", "defer", "relocate"
	Action string `json:"action" validate:"required,oneof=proceed optimize defer relocate" example:"optimize"`

	// Reason explains why this action is recommended
	Reason string `json:"reason" validate:"required" example:"High carbon intensity at edge location"`

	// AlternativeEdges suggests alternative edge locations with lower carbon intensity
	AlternativeEdges []EdgeAlternative `json:"alternative_edges,omitempty"`

	// OptimizationTips provides specific optimization suggestions
	OptimizationTips []string `json:"optimization_tips,omitempty"`

	// EstimatedSavings estimates potential carbon savings in grams CO2
	EstimatedSavings float64 `json:"estimated_savings_g_co2,omitempty" example:"45.2"`

	// TimeBasedStrategy suggests time-based optimization if applicable
	TimeBasedStrategy *TimeBasedStrategy `json:"time_based_strategy,omitempty"`
}

// EdgeAlternative represents an alternative edge location with better carbon characteristics.
type EdgeAlternative struct {
	// Location is the alternative edge location
	Location string `json:"location" example:"Frankfurt"`

	// Provider is the CDN provider for this edge
	Provider string `json:"provider" example:"CloudFlare"`

	// CarbonIntensity at this alternative location
	CarbonIntensity float64 `json:"carbon_intensity" example:"95.3"`

	// Distance from user in kilometers
	Distance float64 `json:"distance_km" example:"650.2"`

	// EstimatedLatency in milliseconds
	EstimatedLatency int `json:"estimated_latency_ms" example:"45"`

	// AvailabilityScore indicates the likelihood this edge can serve the content (0-100)
	AvailabilityScore float64 `json:"availability_score" example:"95.5"`
}

// TimeBasedStrategy provides time-based optimization recommendations.
type TimeBasedStrategy struct {
	// CurrentOptimal indicates if current time is optimal
	CurrentOptimal bool `json:"current_optimal" example:"false"`

	// NextOptimalWindow suggests the next optimal time window
	NextOptimalWindow *GreenHour `json:"next_optimal_window,omitempty"`

	// DeferralBenefit estimates carbon savings by deferring (g CO2)
	DeferralBenefit float64 `json:"deferral_benefit_g_co2,omitempty" example:"120.5"`
}

// CDNProvider represents a CDN provider's edge network configuration.
type CDNProvider struct {
	// Name is the provider name
	Name string `json:"name" example:"CloudFlare"`

	// EdgeLocations maps edge location names to their grid zones
	EdgeLocations map[string]EdgeLocationInfo `json:"edge_locations"`

	// DefaultEdgeSelection is the default edge selection strategy
	DefaultEdgeSelection string `json:"default_edge_selection" example:"geo_nearest"`

	// CarbonAwareRouting indicates if the provider supports carbon-aware routing
	CarbonAwareRouting bool `json:"carbon_aware_routing" example:"false"`
}

// EdgeLocationInfo contains information about a specific edge location.
type EdgeLocationInfo struct {
	// City is the city name
	City string `json:"city" example:"Frankfurt"`

	// Country is the country name
	Country string `json:"country" example:"Germany"`

	// GridZone is the electricity grid zone
	GridZone string `json:"grid_zone" example:"DE"`

	// Latitude of the edge location
	Latitude float64 `json:"latitude" example:"50.1109"`

	// Longitude of the edge location
	Longitude float64 `json:"longitude" example:"8.6821"`

	// Tier indicates the edge tier (1=primary, 2=secondary, 3=tertiary)
	Tier int `json:"tier" example:"1"`

	// Capacity indicates relative capacity (high, medium, low)
	Capacity string `json:"capacity" example:"high"`

	// RenewableCommitment indicates if this location has renewable energy commitments
	RenewableCommitment bool `json:"renewable_commitment" example:"true"`
}

// ContentTypeWeights defines carbon weight distributions for different content types.
type ContentTypeWeights struct {
	// TransmissionWeight is the percentage of carbon footprint from transmission
	TransmissionWeight float64

	// ComputationWeight is the percentage of carbon footprint from computation
	ComputationWeight float64
}

// PredefinedContentWeights provides standard weight distributions for common content types.
var PredefinedContentWeights = map[string]ContentTypeWeights{
	"static": {
		TransmissionWeight: 80, // Static content: mostly transmission
		ComputationWeight:  20,
	},
	"api": {
		TransmissionWeight: 40, // API calls: balanced
		ComputationWeight:  60,
	},
	"video": {
		TransmissionWeight: 60, // Video: significant transmission, some transcoding
		ComputationWeight:  40,
	},
	"dynamic": {
		TransmissionWeight: 30, // Dynamic content: mostly computation
		ComputationWeight:  70,
	},
	"ai": {
		TransmissionWeight: 20, // AI/ML: heavy computation
		ComputationWeight:  80,
	},
	"database": {
		TransmissionWeight: 25, // Database queries: computation heavy
		ComputationWeight:  75,
	},
}

// CalculateWeightedIntensity computes the weighted carbon intensity based on location data and content type.
func CalculateWeightedIntensity(userIntensity, edgeIntensity float64, contentType string) (float64, float64, float64) {
	weights, exists := PredefinedContentWeights[contentType]
	if !exists {
		// Default to balanced weights
		weights = ContentTypeWeights{
			TransmissionWeight: 50,
			ComputationWeight:  50,
		}
	}

	// Calculate weighted intensity
	weightedIntensity := (userIntensity * weights.TransmissionWeight / 100) +
		(edgeIntensity * weights.ComputationWeight / 100)

	return weightedIntensity, weights.TransmissionWeight, weights.ComputationWeight
}

// CalculateDistance estimates the distance between two locations using the Haversine formula.
func CalculateDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const earthRadius = 6371 // km

	lat1Rad := lat1 * math.Pi / 180
	lat2Rad := lat2 * math.Pi / 180
	deltaLat := (lat2 - lat1) * math.Pi / 180
	deltaLon := (lon2 - lon1) * math.Pi / 180

	a := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
			math.Sin(deltaLon/2)*math.Sin(deltaLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadius * c
}

// EstimateNetworkHops estimates the number of network hops based on distance.
func EstimateNetworkHops(distance float64) int {
	// Rough estimation: 1 hop per 150km on average
	hops := int(distance/150) + 3 // Minimum 3 hops
	if hops > 30 {
		hops = 30 // Cap at 30 hops
	}
	return hops
}

// GenerateDualGridRecommendation creates optimization recommendations based on dual-grid analysis.
func GenerateDualGridRecommendation(dual *DualGridCarbonIntensity, alternatives []EdgeAlternative) DualGridRecommendation {
	rec := DualGridRecommendation{
		AlternativeEdges: alternatives,
		OptimizationTips: []string{},
	}

	// Determine action based on weighted intensity
	switch {
	case dual.WeightedIntensity < 150:
		rec.Action = "proceed"
		rec.Reason = "Both user and edge locations have low carbon intensity"
		rec.OptimizationTips = append(rec.OptimizationTips, "Current conditions are optimal for content delivery")

	case dual.WeightedIntensity < 300:
		rec.Action = "optimize"
		rec.Reason = "Moderate carbon intensity detected"
		
		// Add specific optimization tips based on which location has higher intensity
		if dual.UserLocation.CarbonIntensity > dual.EdgeLocation.CarbonIntensity {
			rec.OptimizationTips = append(rec.OptimizationTips, 
				"Consider caching content locally to reduce repeated transmissions",
				"Enable aggressive browser caching for static assets")
		} else {
			rec.OptimizationTips = append(rec.OptimizationTips,
				"Consider using pre-computed/cached responses",
				"Optimize server-side processing to reduce computation time")
		}

		// Check if alternative edges are significantly better
		if len(alternatives) > 0 && alternatives[0].CarbonIntensity < dual.EdgeLocation.CarbonIntensity*0.7 {
			rec.Action = "relocate"
			rec.Reason = fmt.Sprintf("Alternative edge location has %.0f%% lower carbon intensity",
				(1-alternatives[0].CarbonIntensity/dual.EdgeLocation.CarbonIntensity)*100)
			rec.EstimatedSavings = (dual.EdgeLocation.CarbonIntensity - alternatives[0].CarbonIntensity) * 
				dual.ComputationWeight / 100
		}

	default: // > 300
		rec.Action = "defer"
		rec.Reason = "High carbon intensity at one or both locations"
		rec.OptimizationTips = append(rec.OptimizationTips,
			"Defer non-essential operations to off-peak hours",
			"Consider scheduling batch operations during green windows")

		// Add time-based strategy if applicable
		if dual.UserLocation.NextGreenWindow.After(time.Now()) || 
		   dual.EdgeLocation.NextGreenWindow.After(time.Now()) {
			nextWindow := dual.UserLocation.NextGreenWindow
			if dual.EdgeLocation.NextGreenWindow.Before(nextWindow) {
				nextWindow = dual.EdgeLocation.NextGreenWindow
			}

			rec.TimeBasedStrategy = &TimeBasedStrategy{
				CurrentOptimal: false,
				NextOptimalWindow: &GreenHour{
					Start:            nextWindow,
					End:              nextWindow.Add(2 * time.Hour),
					CarbonIntensity:  dual.WeightedIntensity * 0.6, // Estimate 40% reduction
					RenewablePercent: 70,
				},
				DeferralBenefit: dual.WeightedIntensity * 0.4,
			}
		}
	}

	// Add content-specific tips
	switch dual.ContentType {
	case "video":
		rec.OptimizationTips = append(rec.OptimizationTips,
			"Use adaptive bitrate streaming to reduce bandwidth",
			"Consider lower resolution options during high-carbon periods")
	case "api":
		rec.OptimizationTips = append(rec.OptimizationTips,
			"Implement response caching with appropriate TTLs",
			"Batch API requests to reduce overhead")
	case "static":
		rec.OptimizationTips = append(rec.OptimizationTips,
			"Use CDN caching aggressively",
			"Implement service workers for offline capability")
	}

	return rec
}

// GetOptimalEdgeLocation returns the most carbon-efficient edge location from a list.
func GetOptimalEdgeLocation(userLat, userLon float64, edges []EdgeLocationInfo, 
	getIntensity func(gridZone string) float64) *EdgeLocationInfo {
	
	if len(edges) == 0 {
		return nil
	}

	var optimalEdge *EdgeLocationInfo
	var lowestScore float64 = math.MaxFloat64

	for i := range edges {
		edge := &edges[i]
		
		// Calculate distance factor
		distance := CalculateDistance(userLat, userLon, edge.Latitude, edge.Longitude)
		distanceFactor := math.Min(distance/1000, 2.0) // Normalize to 0-2 range

		// Get carbon intensity for this edge
		intensity := getIntensity(edge.GridZone)
		
		// Calculate composite score (lower is better)
		// Weight: 70% carbon intensity, 30% distance
		score := (intensity * 0.7) + (distanceFactor * 100 * 0.3)
		
		// Apply tier bonus (lower tier = higher priority)
		score = score * (1 + float64(edge.Tier-1)*0.1)
		
		// Apply renewable commitment bonus
		if edge.RenewableCommitment {
			score = score * 0.9
		}

		if score < lowestScore {
			lowestScore = score
			optimalEdge = edge
		}
	}

	return optimalEdge
}

// String provides a human-readable representation of dual-grid carbon intensity.
func (d *DualGridCarbonIntensity) String() string {
	return fmt.Sprintf("Dual-Grid Carbon: User(%s: %.1f) <-> Edge(%s: %.1f) = Weighted: %.1f g CO2/kWh",
		d.UserLocation.Location, d.UserLocation.CarbonIntensity,
		d.EdgeLocation.Location, d.EdgeLocation.CarbonIntensity,
		d.WeightedIntensity)
}

// IsOptimal returns true if both locations have low carbon intensity.
func (d *DualGridCarbonIntensity) IsOptimal() bool {
	return d.UserLocation.IsGreen() && d.EdgeLocation.IsGreen()
}

// RequiresOptimization returns true if optimization is recommended.
func (d *DualGridCarbonIntensity) RequiresOptimization() bool {
	return d.WeightedIntensity > 150 || 
		d.UserLocation.CarbonIntensity > 300 || 
		d.EdgeLocation.CarbonIntensity > 300
}