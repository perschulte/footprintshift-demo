// Package carbon provides interfaces for the carbon intelligence service.
package carbon

import (
	"context"
	"time"

	"github.com/perschulte/greenweb-api/pkg/carbon"
)

// IntelligenceServiceInterface defines the interface for carbon intelligence operations.
type IntelligenceServiceInterface interface {
	// GetRelativeCarbonIntensity returns carbon intensity with relative metrics and regional context.
	GetRelativeCarbonIntensity(ctx context.Context, location string) (*RelativeCarbonIntensity, error)
	
	// GetDynamicGreenHours returns green hours forecast using dynamic thresholds.
	GetDynamicGreenHours(ctx context.Context, location string, hours int) (*carbon.GreenHoursForecast, error)
	
	// GetCarbonTrends returns historical carbon intensity trends and patterns.
	GetCarbonTrends(ctx context.Context, location string, period string, days int) (*CarbonTrend, error)
}

// PatternServiceInterface defines the interface for regional pattern management.
type PatternServiceInterface interface {
	// GetRegionalPattern retrieves the learned pattern for a region.
	GetRegionalPattern(location string) (*RegionPattern, error)
	
	// UpdatePattern forces an update of the regional pattern.
	UpdatePattern(ctx context.Context, location string) error
	
	// ClearPattern removes cached pattern data for a location.
	ClearPattern(location string) error
	
	// GetSupportedRegions returns regions with available pattern data.
	GetSupportedRegions() []string
}

// HighVariationRegionServiceInterface defines specialized operations for high-variation regions.
type HighVariationRegionServiceInterface interface {
	// GetOptimizedSchedule returns an optimized schedule for high-variation regions.
	GetOptimizedSchedule(ctx context.Context, location string, taskDuration int, flexibilityHours int) (*OptimizationSchedule, error)
	
	// GetRegionalStrategy returns specific optimization strategies for a region.
	GetRegionalStrategy(location string) (*RegionalStrategy, error)
	
	// IsHighVariationRegion checks if a region has high carbon intensity variation.
	IsHighVariationRegion(location string) bool
}

// OptimizationSchedule represents an optimal schedule for energy-intensive tasks.
type OptimizationSchedule struct {
	Location           string              `json:"location"`
	TaskDuration       int                 `json:"task_duration_hours"`
	RecommendedWindows []OptimalWindow     `json:"recommended_windows"`
	AvoidanceWindows   []AvoidanceWindow   `json:"avoidance_windows"`
	Savings            *EmissionsSavings   `json:"potential_savings"`
	Strategy           string              `json:"optimization_strategy"`
}

// AvoidanceWindow represents time periods to avoid due to high carbon intensity.
type AvoidanceWindow struct {
	Start             time.Time `json:"start"`
	End               time.Time `json:"end"`
	ExpectedIntensity float64   `json:"expected_intensity"`
	Reason            string    `json:"reason"`
}

// EmissionsSavings represents potential CO2 savings from optimization.
type EmissionsSavings struct {
	OptimalEmissions    float64 `json:"optimal_emissions_kg_co2"`
	WorstCaseEmissions  float64 `json:"worst_case_emissions_kg_co2"`
	PotentialSavings    float64 `json:"potential_savings_kg_co2"`
	PercentageSavings   float64 `json:"percentage_savings"`
	EquivalentTrees     float64 `json:"equivalent_trees_planted"`
}

// RegionalStrategy contains optimization strategies specific to a region.
type RegionalStrategy struct {
	Region                string                 `json:"region"`
	PrimaryEnergySource   string                 `json:"primary_energy_source"`
	OptimalHours          []int                  `json:"optimal_hours"`
	AvoidanceHours        []int                  `json:"avoidance_hours"`
	SeasonalFactors       map[string]float64     `json:"seasonal_factors"`
	WeekdayVsWeekend      bool                   `json:"weekday_weekend_difference"`
	StrategicRecommendations []string            `json:"strategic_recommendations"`
	CoalHeavyGrid         bool                   `json:"coal_heavy_grid"`
	VariationLevel        string                 `json:"variation_level"` // "low", "medium", "high"
}