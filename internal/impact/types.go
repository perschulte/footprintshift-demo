// Package impact provides science-based CO2 impact calculation with conservative estimates
// to avoid greenwashing. All calculations include confidence intervals and methodology transparency.
package impact

import (
	"time"
)

// EmissionFactors contains region-specific emission factors from reputable sources
// Sources: IEA (2023), Carbon Trust, EPA, EU Grid Mix Data
type EmissionFactors struct {
	// GridCarbonIntensity in g CO2/kWh by region
	GridCarbonIntensity map[string]float64

	// NetworkTransmission in g CO2/GB by provider type
	NetworkTransmission map[string]float64

	// DeviceConsumption in Wh by device type and activity
	DeviceConsumption map[string]map[string]float64

	// DataCenterPUE (Power Usage Effectiveness) by region
	DataCenterPUE map[string]float64
}

// DefaultEmissionFactors provides conservative emission factors based on 2023 data
var DefaultEmissionFactors = EmissionFactors{
	GridCarbonIntensity: map[string]float64{
		"EU":     295.0, // EU average 2023
		"US":     420.0, // US average 2023
		"CN":     580.0, // China average 2023
		"IN":     720.0, // India average 2023
		"UK":     233.0, // UK average 2023
		"FR":     85.0,  // France (nuclear heavy)
		"DE":     380.0, // Germany average 2023
		"global": 475.0, // Global average 2023
	},
	NetworkTransmission: map[string]float64{
		"mobile_3g":   11.0, // g CO2/GB
		"mobile_4g":   7.0,  // g CO2/GB
		"mobile_5g":   5.0,  // g CO2/GB
		"wifi":        3.5,  // g CO2/GB
		"ethernet":    2.9,  // g CO2/GB
		"fixed_broad": 3.2,  // g CO2/GB
	},
	DeviceConsumption: map[string]map[string]float64{
		"smartphone": {
			"idle":             0.5,  // Wh/hour
			"browsing":         2.0,  // Wh/hour
			"video_streaming":  3.5,  // Wh/hour
			"javascript_heavy": 2.8,  // Wh/hour
			"image_loading":    2.2,  // Wh/hour
		},
		"laptop": {
			"idle":             15.0, // Wh/hour
			"browsing":         25.0, // Wh/hour
			"video_streaming":  40.0, // Wh/hour
			"javascript_heavy": 35.0, // Wh/hour
			"image_loading":    28.0, // Wh/hour
		},
		"desktop": {
			"idle":             40.0, // Wh/hour
			"browsing":         60.0, // Wh/hour
			"video_streaming":  85.0, // Wh/hour
			"javascript_heavy": 75.0, // Wh/hour
			"image_loading":    65.0, // Wh/hour
		},
		"tablet": {
			"idle":             3.0,  // Wh/hour
			"browsing":         8.0,  // Wh/hour
			"video_streaming":  12.0, // Wh/hour
			"javascript_heavy": 10.0, // Wh/hour
			"image_loading":    9.0,  // Wh/hour
		},
	},
	DataCenterPUE: map[string]float64{
		"EU":     1.6,  // European average
		"US":     1.8,  // US average
		"global": 1.67, // Global average
	},
}

// ImpactType represents different types of carbon impact measurements
type ImpactType string

const (
	ImpactTypeVideoStreaming ImpactType = "video_streaming"
	ImpactTypeImageLoading   ImpactType = "image_loading"
	ImpactTypeJavaScript     ImpactType = "javascript_execution"
	ImpactTypeAIInference    ImpactType = "ai_inference"
	ImpactTypePageLoad       ImpactType = "page_load"
	ImpactTypeDataTransfer   ImpactType = "data_transfer"
)

// CalculationRequest represents a request to calculate carbon impact
type CalculationRequest struct {
	// Type of impact to calculate
	Type ImpactType `json:"type" validate:"required"`

	// Duration in seconds (for streaming, JS execution)
	Duration float64 `json:"duration,omitempty" validate:"min=0"`

	// DataSize in MB (for images, data transfer)
	DataSize float64 `json:"data_size,omitempty" validate:"min=0"`

	// VideoQuality for video streaming calculations
	VideoQuality string `json:"video_quality,omitempty" validate:"oneof=360p 480p 720p 1080p 4k"`

	// ImageCount for image optimization calculations
	ImageCount int `json:"image_count,omitempty" validate:"min=0"`

	// Device type (smartphone, laptop, desktop, tablet)
	DeviceType string `json:"device_type,omitempty" validate:"oneof=smartphone laptop desktop tablet"`

	// Connection type (mobile_3g, mobile_4g, mobile_5g, wifi, ethernet)
	ConnectionType string `json:"connection_type,omitempty"`

	// Region for grid carbon intensity
	Region string `json:"region,omitempty"`

	// OptimizationLevel (0-100) for calculating savings
	OptimizationLevel float64 `json:"optimization_level,omitempty" validate:"min=0,max=100"`

	// Include rebound effects in calculation
	IncludeReboundEffects bool `json:"include_rebound_effects,omitempty"`
}

// ImpactResult represents the calculated carbon impact with confidence intervals
type ImpactResult struct {
	// BaselineEmissions in grams CO2e
	BaselineEmissions float64 `json:"baseline_emissions"`

	// OptimizedEmissions in grams CO2e
	OptimizedEmissions float64 `json:"optimized_emissions"`

	// Savings in grams CO2e
	Savings float64 `json:"savings"`

	// SavingsPercentage as a percentage
	SavingsPercentage float64 `json:"savings_percentage"`

	// ConfidenceInterval represents uncertainty (±%)
	ConfidenceInterval float64 `json:"confidence_interval"`

	// LowerBound of emissions estimate
	LowerBound float64 `json:"lower_bound"`

	// UpperBound of emissions estimate
	UpperBound float64 `json:"upper_bound"`

	// Components breakdown of emissions
	Components EmissionComponents `json:"components"`

	// ReboundEffect estimated additional consumption
	ReboundEffect float64 `json:"rebound_effect,omitempty"`

	// NetSavings after accounting for rebound effects
	NetSavings float64 `json:"net_savings"`

	// Methodology explanation
	Methodology string `json:"methodology"`

	// DataSources used for calculation
	DataSources []string `json:"data_sources"`

	// Warnings about limitations or assumptions
	Warnings []string `json:"warnings,omitempty"`

	// CalculatedAt timestamp
	CalculatedAt time.Time `json:"calculated_at"`
}

// EmissionComponents breaks down emissions by source
type EmissionComponents struct {
	// DeviceEmissions from client device energy use
	DeviceEmissions float64 `json:"device_emissions"`

	// NetworkEmissions from data transmission
	NetworkEmissions float64 `json:"network_emissions"`

	// DataCenterEmissions from server processing
	DataCenterEmissions float64 `json:"datacenter_emissions"`

	// Percentages for each component
	DevicePercentage     float64 `json:"device_percentage"`
	NetworkPercentage    float64 `json:"network_percentage"`
	DataCenterPercentage float64 `json:"datacenter_percentage"`
}

// BaselineMeasurement represents measured baseline carbon footprint
type BaselineMeasurement struct {
	// ID unique identifier
	ID string `json:"id"`

	// URL being measured
	URL string `json:"url" validate:"required,url"`

	// PageLoadEmissions in g CO2e
	PageLoadEmissions float64 `json:"page_load_emissions"`

	// DataTransferred in MB
	DataTransferred float64 `json:"data_transferred"`

	// JavaScriptSize in KB
	JavaScriptSize float64 `json:"javascript_size"`

	// ImageSize in MB
	ImageSize float64 `json:"image_size"`

	// VideoSize in MB
	VideoSize float64 `json:"video_size"`

	// LoadTime in seconds
	LoadTime float64 `json:"load_time"`

	// ResourceCount by type
	ResourceCount map[string]int `json:"resource_count"`

	// EstimatedHourlyEmissions based on typical usage
	EstimatedHourlyEmissions float64 `json:"estimated_hourly_emissions"`

	// MeasuredAt timestamp
	MeasuredAt time.Time `json:"measured_at"`

	// DeviceType used for measurement
	DeviceType string `json:"device_type"`

	// ConnectionType used for measurement
	ConnectionType string `json:"connection_type"`

	// Region of measurement
	Region string `json:"region"`
}

// ImpactReport represents a comprehensive impact report
type ImpactReport struct {
	// ID unique identifier
	ID string `json:"id"`

	// Period covered by the report
	Period ReportPeriod `json:"period"`

	// TotalSavings in kg CO2e
	TotalSavings float64 `json:"total_savings"`

	// BaselineTotal in kg CO2e
	BaselineTotal float64 `json:"baseline_total"`

	// OptimizedTotal in kg CO2e
	OptimizedTotal float64 `json:"optimized_total"`

	// SavingsByType breakdown
	SavingsByType map[ImpactType]float64 `json:"savings_by_type"`

	// OptimizationEvents count
	OptimizationEvents int `json:"optimization_events"`

	// AverageOptimizationLevel
	AverageOptimizationLevel float64 `json:"average_optimization_level"`

	// EquivalentTo comparisons (e.g., "X miles driven")
	EquivalentTo []Equivalence `json:"equivalent_to"`

	// ConfidenceScore (0-100)
	ConfidenceScore float64 `json:"confidence_score"`

	// Methodology used
	Methodology string `json:"methodology"`

	// Recommendations for further improvements
	Recommendations []string `json:"recommendations"`

	// GeneratedAt timestamp
	GeneratedAt time.Time `json:"generated_at"`
}

// ReportPeriod defines the time period for a report
type ReportPeriod struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
	Days  int       `json:"days"`
}

// Equivalence provides relatable comparisons for CO2 savings
type Equivalence struct {
	// Type of equivalence (e.g., "driving", "trees", "flights")
	Type string `json:"type"`

	// Value numerical value
	Value float64 `json:"value"`

	// Unit of measurement
	Unit string `json:"unit"`

	// Description human-readable description
	Description string `json:"description"`
}

// ValidationRequest represents a request to validate claimed savings
type ValidationRequest struct {
	// ClaimedSavings in g CO2e
	ClaimedSavings float64 `json:"claimed_savings" validate:"required,min=0"`

	// OptimizationType what was optimized
	OptimizationType ImpactType `json:"optimization_type" validate:"required"`

	// Parameters used for optimization
	Parameters map[string]interface{} `json:"parameters"`

	// BaselineData if available
	BaselineData *BaselineMeasurement `json:"baseline_data,omitempty"`
}

// ValidationResult represents the result of validating claimed savings
type ValidationResult struct {
	// IsValid whether the claimed savings are reasonable
	IsValid bool `json:"is_valid"`

	// ValidatedSavings our calculated savings
	ValidatedSavings float64 `json:"validated_savings"`

	// Variance percentage difference from claimed
	Variance float64 `json:"variance"`

	// Rating (conservative, reasonable, optimistic, unrealistic)
	Rating string `json:"rating"`

	// Explanation of the validation
	Explanation string `json:"explanation"`

	// Suggestions for more accurate calculations
	Suggestions []string `json:"suggestions,omitempty"`
}

// SessionMetrics tracks carbon impact for a user session
type SessionMetrics struct {
	// SessionID unique identifier
	SessionID string `json:"session_id"`

	// StartTime of the session
	StartTime time.Time `json:"start_time"`

	// Duration in seconds
	Duration float64 `json:"duration"`

	// TotalEmissions in g CO2e
	TotalEmissions float64 `json:"total_emissions"`

	// EmissionsByActivity breakdown
	EmissionsByActivity map[string]float64 `json:"emissions_by_activity"`

	// OptimizationsApplied count
	OptimizationsApplied int `json:"optimizations_applied"`

	// EstimatedSavings from optimizations
	EstimatedSavings float64 `json:"estimated_savings"`

	// DeviceType used
	DeviceType string `json:"device_type"`

	// Region of the user
	Region string `json:"region"`
}

// RealTimeMetrics for dashboard display
type RealTimeMetrics struct {
	// CurrentCO2Rate in g CO2/hour
	CurrentCO2Rate float64 `json:"current_co2_rate"`

	// OptimizationActive whether optimizations are active
	OptimizationActive bool `json:"optimization_active"`

	// InstantSavingsRate in g CO2/hour
	InstantSavingsRate float64 `json:"instant_savings_rate"`

	// CumulativeSavingsToday in kg CO2
	CumulativeSavingsToday float64 `json:"cumulative_savings_today"`

	// ActiveUsers currently being tracked
	ActiveUsers int `json:"active_users"`

	// LastUpdated timestamp
	LastUpdated time.Time `json:"last_updated"`
}