// Package carbon provides service initialization for the carbon intelligence system.
package carbon

import (
	"log/slog"
	"time"

	"github.com/perschulte/greenweb-api/pkg/carbon"
)

// ServiceManager manages carbon intelligence services and their dependencies.
type ServiceManager struct {
	intelligenceService *IntelligenceService
	adapter            *ElectricityMapsAdapter
	logger             *slog.Logger
}

// NewServiceManager creates a new carbon service manager.
func NewServiceManager(electricityService ElectricityMapsService, logger *slog.Logger, config *IntelligenceConfig) *ServiceManager {
	// Create adapter to bridge electricity maps service with carbon interfaces
	adapter := NewElectricityMapsAdapter(electricityService)
	
	// Create intelligence service
	intelligence := NewIntelligenceService(adapter, logger, config)
	
	return &ServiceManager{
		intelligenceService: intelligence,
		adapter:            adapter,
		logger:             logger,
	}
}

// GetIntelligenceService returns the carbon intelligence service.
func (sm *ServiceManager) GetIntelligenceService() *IntelligenceService {
	return sm.intelligenceService
}

// GetAdapter returns the electricity maps adapter.
func (sm *ServiceManager) GetAdapter() *ElectricityMapsAdapter {
	return sm.adapter
}

// GetRegionalOptimization returns region-specific optimization strategies.
func (sm *ServiceManager) GetRegionalOptimization(region string) *carbon.RegionalOptimization {
	// Define region-specific strategies for high-variation regions
	strategies := map[string]*carbon.RegionalOptimization{
		"PL": {
			Region:              "PL",
			PrimaryEnergySource: "coal",
			OptimalHours:        []int{22, 23, 0, 1, 2, 3, 4, 5, 11, 12, 13, 14},
			AvoidanceHours:      []int{17, 18, 19, 20, 21, 7, 8, 9},
			VariationLevel:      "high",
			Recommendations: []string{
				"Schedule energy-intensive tasks during night hours (22:00-05:00)",
				"Avoid peak evening hours (17:00-21:00) when coal plants ramp up",
				"Take advantage of midday solar generation (11:00-14:00)",
				"Weekend scheduling provides 15-20% better carbon efficiency",
				"Consider seasonal patterns - summer has more renewable generation",
			},
		},
		"US-TEX": {
			Region:              "US-TEX",
			PrimaryEnergySource: "mixed",
			OptimalHours:        []int{10, 11, 12, 13, 14, 15, 23, 0, 1, 2, 3, 4},
			AvoidanceHours:      []int{16, 17, 18, 19, 20, 21, 6, 7, 8, 9},
			VariationLevel:      "high",
			Recommendations: []string{
				"Maximize solar window utilization (10:00-15:00)",
				"Avoid extreme peak hours (16:00-21:00) when gas peakers activate",
				"Night hours (23:00-04:00) often have good wind generation",
				"Summer cooling loads create high variation - plan accordingly",
				"West Texas wind patterns favor overnight scheduling",
			},
		},
		"CN": {
			Region:              "CN",
			PrimaryEnergySource: "coal",
			OptimalHours:        []int{1, 2, 3, 4, 5, 11, 12, 13, 14, 15},
			AvoidanceHours:      []int{18, 19, 20, 21, 22, 7, 8, 9, 10},
			VariationLevel:      "high",
			Recommendations: []string{
				"Schedule during early morning hours (01:00-05:00) for lowest grid load",
				"Midday solar generation window (11:00-15:00) increasingly reliable",
				"Avoid industrial peak hours (18:00-22:00)",
				"Regional differences significant - eastern coastal areas cleaner",
				"Seasonal coal heating creates winter optimization challenges",
			},
		},
		"IN": {
			Region:              "IN",
			PrimaryEnergySource: "coal",
			OptimalHours:        []int{2, 3, 4, 5, 11, 12, 13, 14, 15, 16},
			AvoidanceHours:      []int{18, 19, 20, 21, 22, 23, 6, 7, 8, 9},
			VariationLevel:      "high",
			Recommendations: []string{
				"Early morning hours (02:00-05:00) have lowest coal dependency",
				"Solar generation peak (11:00-16:00) offers best carbon efficiency",
				"Avoid evening industrial peak (18:00-23:00)",
				"Monsoon season affects renewable generation patterns",
				"Southern states typically have better renewable mix",
			},
		},
		"AU-NSW": {
			Region:              "AU-NSW",
			PrimaryEnergySource: "mixed",
			OptimalHours:        []int{10, 11, 12, 13, 14, 15, 23, 0, 1, 2, 3},
			AvoidanceHours:      []int{17, 18, 19, 20, 21, 7, 8, 9},
			VariationLevel:      "high",
			Recommendations: []string{
				"Solar generation window (10:00-15:00) provides cleanest energy",
				"Night hours (23:00-03:00) benefit from lower demand",
				"Avoid evening air conditioning peak (17:00-21:00)",
				"Seasonal patterns significant - summer has high variation",
				"Coal retirement schedule improving long-term trends",
			},
		},
		"ZA": {
			Region:              "ZA",
			PrimaryEnergySource: "coal",
			OptimalHours:        []int{11, 12, 13, 14, 15, 1, 2, 3, 4, 5},
			AvoidanceHours:      []int{17, 18, 19, 20, 21, 22, 6, 7, 8},
			VariationLevel:      "high",
			Recommendations: []string{
				"Midday solar generation (11:00-15:00) offers best opportunities",
				"Early morning hours (01:00-05:00) have reduced coal load",
				"Avoid evening peak (17:00-22:00) when load shedding risk highest",
				"Grid stability affects optimization - flexible scheduling essential",
				"Industrial demand patterns create predictable carbon peaks",
			},
		},
	}
	
	// Return strategy for region, or default strategy if not found
	if strategy, exists := strategies[region]; exists {
		return strategy
	}
	
	// Default strategy for regions without specific optimization
	return &carbon.RegionalOptimization{
		Region:              region,
		PrimaryEnergySource: "mixed",
		OptimalHours:        []int{22, 23, 0, 1, 2, 3, 11, 12, 13, 14},
		AvoidanceHours:      []int{17, 18, 19, 20},
		VariationLevel:      "medium",
		Recommendations: []string{
			"Schedule during typical low-demand hours (22:00-03:00)",
			"Take advantage of midday renewable generation when available",
			"Avoid evening peak hours (17:00-20:00)",
			"Monitor local grid patterns for region-specific optimization",
		},
	}
}

// IsHighVariationRegion checks if a region is known for high carbon variation.
func (sm *ServiceManager) IsHighVariationRegion(region string) bool {
	highVariationRegions := map[string]bool{
		"PL":     true, // Poland - coal heavy with wind variation
		"US-TEX": true, // Texas - high wind + gas peakers
		"CN":     true, // China - coal heavy industrial patterns
		"IN":     true, // India - coal heavy with growing renewables
		"AU-NSW": true, // Australia NSW - coal transition + high solar
		"ZA":     true, // South Africa - coal heavy with grid constraints
		"CZ":     true, // Czech Republic - coal + industrial patterns
		"RO":     true, // Romania - mixed sources with high variation
		"BG":     true, // Bulgaria - coal + imports variation
		"GR":     true, // Greece - lignite + renewables + islands
		"US-CA":  true, // California - high renewables with duck curve
		"US-NY":  true, // New York - mixed sources with demand variation
	}
	
	return highVariationRegions[region]
}

// GetDefaultConfig returns the default intelligence service configuration.
func GetDefaultConfig() *IntelligenceConfig {
	return &DefaultIntelligenceConfig
}

// GetHighVariationConfig returns configuration optimized for high-variation regions.
func GetHighVariationConfig() *IntelligenceConfig {
	config := DefaultIntelligenceConfig
	config.UpdateInterval = 10 * time.Minute     // More frequent updates
	config.MinDataPointsForAnalysis = 72         // Require less data for faster adaptation
	config.HistoryRetentionDays = 14             // Shorter retention for faster adaptation
	return &config
}