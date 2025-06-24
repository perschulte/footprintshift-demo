// Package carbon provides types and interfaces for carbon intensity monitoring and green energy forecasting.
//
// This package contains the core data structures used for representing carbon intensity data,
// green hour forecasts, and related carbon monitoring information. These types are designed
// to be stable and backward-compatible for external SDK consumers.
package carbon

import (
	"fmt"
	"time"
)

// CarbonIntensity represents the current carbon intensity data for a specific location.
//
// Carbon intensity is measured in grams of CO2 equivalent per kilowatt-hour (g CO2/kWh)
// and indicates how much carbon dioxide is emitted to produce electricity at a given time.
// Lower values indicate cleaner energy sources (more renewable energy).
//
// The Mode field provides a simple categorization:
//   - "green": Low carbon intensity (< 150 g CO2/kWh) - optimal time for energy consumption
//   - "yellow": Medium carbon intensity (150-300 g CO2/kWh) - moderate energy usage recommended
//   - "red": High carbon intensity (> 300 g CO2/kWh) - defer non-essential energy usage
type CarbonIntensity struct {
	// Location is the geographical location for this carbon intensity reading
	Location string `json:"location" validate:"required" example:"Berlin"`

	// CarbonIntensity is the carbon intensity in grams CO2 per kWh
	CarbonIntensity float64 `json:"carbon_intensity" validate:"min=0" example:"120.5"`

	// RenewablePercent is the percentage of renewable energy in the grid (0-100)
	RenewablePercent float64 `json:"renewable_percentage" validate:"min=0,max=100" example:"75.2"`

	// Mode provides a simple classification: "green", "yellow", or "red"
	Mode string `json:"mode" validate:"required,oneof=green yellow red" example:"green"`

	// Recommendation provides human-readable guidance based on the carbon intensity
	Recommendation string `json:"recommendation" validate:"required" example:"optimal"`

	// NextGreenWindow estimates when the next low-carbon period will occur
	NextGreenWindow time.Time `json:"next_green_window" example:"2024-01-15T22:00:00Z"`

	// Timestamp indicates when this carbon intensity reading was taken
	Timestamp time.Time `json:"timestamp" validate:"required" example:"2024-01-15T14:30:00Z"`

	// FossilFuelPercentage is the percentage of energy from fossil fuels (0-100)
	// This is the inverse of RenewablePercent and is included for compatibility
	FossilFuelPercentage float64 `json:"fossil_fuel_percentage,omitempty" validate:"min=0,max=100" example:"24.8"`

	// Source indicates the data source for this reading (e.g., "electricity_maps", "mock")
	Source string `json:"source,omitempty" example:"electricity_maps"`

	// GridZone identifies the specific electricity grid zone if available
	GridZone string `json:"grid_zone,omitempty" example:"DE"`
}

// IsGreen returns true if the carbon intensity is in the "green" range (< 150 g CO2/kWh).
func (c *CarbonIntensity) IsGreen() bool {
	return c.CarbonIntensity < 150
}

// IsOptimal returns true if this is an optimal time for energy consumption.
// This is equivalent to IsGreen() but provides a more semantic method name.
func (c *CarbonIntensity) IsOptimal() bool {
	return c.IsGreen()
}

// ShouldDefer returns true if energy consumption should be deferred due to high carbon intensity.
func (c *CarbonIntensity) ShouldDefer() bool {
	return c.CarbonIntensity > 300
}

// GetEfficiencyRating returns a rating from 1-10 where 10 is most efficient (lowest carbon).
func (c *CarbonIntensity) GetEfficiencyRating() int {
	switch {
	case c.CarbonIntensity < 100:
		return 10
	case c.CarbonIntensity < 150:
		return 9
	case c.CarbonIntensity < 200:
		return 8
	case c.CarbonIntensity < 250:
		return 7
	case c.CarbonIntensity < 300:
		return 6
	case c.CarbonIntensity < 350:
		return 5
	case c.CarbonIntensity < 400:
		return 4
	case c.CarbonIntensity < 450:
		return 3
	case c.CarbonIntensity < 500:
		return 2
	default:
		return 1
	}
}

// String provides a human-readable representation of the carbon intensity.
func (c *CarbonIntensity) String() string {
	return fmt.Sprintf("%s: %.1f g CO2/kWh (%s mode, %.1f%% renewable)",
		c.Location, c.CarbonIntensity, c.Mode, c.RenewablePercent)
}

// GreenHour represents a forecasted time window with low carbon intensity.
//
// Green hours are periods when the electricity grid has a high proportion of
// renewable energy sources, resulting in lower carbon emissions per unit of
// energy consumed. These periods are optimal for energy-intensive activities.
type GreenHour struct {
	// Start is the beginning of the green hour window
	Start time.Time `json:"start" validate:"required" example:"2024-01-15T22:00:00Z"`

	// End is the end of the green hour window
	End time.Time `json:"end" validate:"required" example:"2024-01-15T23:00:00Z"`

	// CarbonIntensity is the predicted carbon intensity during this window (g CO2/kWh)
	CarbonIntensity float64 `json:"carbon_intensity" validate:"min=0" example:"95.3"`

	// RenewablePercent is the predicted percentage of renewable energy (0-100)
	RenewablePercent float64 `json:"renewable_percentage" validate:"min=0,max=100" example:"85.7"`

	// Confidence indicates the prediction confidence level (0-100)
	Confidence float64 `json:"confidence,omitempty" validate:"min=0,max=100" example:"82.5"`

	// Duration provides the duration of this green hour window as a convenience
	Duration time.Duration `json:"duration,omitempty" example:"1h0m0s"`
}

// GetDuration returns the duration of this green hour window.
func (g *GreenHour) GetDuration() time.Duration {
	if g.Duration != 0 {
		return g.Duration
	}
	return g.End.Sub(g.Start)
}

// IsActive returns true if the current time falls within this green hour window.
func (g *GreenHour) IsActive() bool {
	now := time.Now()
	return now.After(g.Start) && now.Before(g.End)
}

// IsUpcoming returns true if this green hour window starts in the future.
func (g *GreenHour) IsUpcoming() bool {
	return time.Now().Before(g.Start)
}

// TimeUntilStart returns the duration until this green hour window starts.
// Returns 0 if the window has already started.
func (g *GreenHour) TimeUntilStart() time.Duration {
	now := time.Now()
	if now.After(g.Start) {
		return 0
	}
	return g.Start.Sub(now)
}

// String provides a human-readable representation of the green hour.
func (g *GreenHour) String() string {
	return fmt.Sprintf("%.1f g CO2/kWh from %s to %s (%.1f%% renewable)",
		g.CarbonIntensity, g.Start.Format("15:04"), g.End.Format("15:04"), g.RenewablePercent)
}

// GreenHoursForecast predicts the best times for low-carbon energy consumption.
//
// This forecast identifies upcoming time windows when the electricity grid
// will have lower carbon intensity, allowing applications to schedule
// energy-intensive operations during cleaner energy periods.
type GreenHoursForecast struct {
	// Location is the geographical location for this forecast
	Location string `json:"location" validate:"required" example:"Berlin"`

	// GreenHours is a list of predicted low-carbon time windows, sorted by start time
	GreenHours []GreenHour `json:"green_hours" validate:"dive"`

	// BestWindow identifies the single best green hour window in the forecast period
	BestWindow GreenHour `json:"best_window"`

	// ForecastPeriod indicates the time range covered by this forecast
	ForecastPeriod struct {
		Start time.Time `json:"start" example:"2024-01-15T14:00:00Z"`
		End   time.Time `json:"end" example:"2024-01-16T14:00:00Z"`
	} `json:"forecast_period"`

	// GeneratedAt indicates when this forecast was created
	GeneratedAt time.Time `json:"generated_at" validate:"required" example:"2024-01-15T14:30:00Z"`

	// Source indicates the data source for this forecast (e.g., "electricity_maps", "mock")
	Source string `json:"source,omitempty" example:"electricity_maps"`

	// Confidence indicates the overall forecast confidence level (0-100)
	Confidence float64 `json:"confidence,omitempty" validate:"min=0,max=100" example:"75.5"`

	// AverageIntensity is the average carbon intensity across all green hours
	AverageIntensity float64 `json:"average_intensity,omitempty" example:"105.2"`
}

// GetActiveGreenHours returns green hours that are currently active.
func (f *GreenHoursForecast) GetActiveGreenHours() []GreenHour {
	var active []GreenHour
	for _, hour := range f.GreenHours {
		if hour.IsActive() {
			active = append(active, hour)
		}
	}
	return active
}

// GetUpcomingGreenHours returns green hours that haven't started yet.
func (f *GreenHoursForecast) GetUpcomingGreenHours() []GreenHour {
	var upcoming []GreenHour
	for _, hour := range f.GreenHours {
		if hour.IsUpcoming() {
			upcoming = append(upcoming, hour)
		}
	}
	return upcoming
}

// GetNextGreenHour returns the next upcoming green hour, or nil if none found.
func (f *GreenHoursForecast) GetNextGreenHour() *GreenHour {
	upcoming := f.GetUpcomingGreenHours()
	if len(upcoming) == 0 {
		return nil
	}
	return &upcoming[0] // GreenHours should be sorted by start time
}

// HasActiveGreenHour returns true if there's currently an active green hour.
func (f *GreenHoursForecast) HasActiveGreenHour() bool {
	return len(f.GetActiveGreenHours()) > 0
}

// GetTotalGreenDuration returns the total duration of all green hours in the forecast.
func (f *GreenHoursForecast) GetTotalGreenDuration() time.Duration {
	var total time.Duration
	for _, hour := range f.GreenHours {
		total += hour.GetDuration()
	}
	return total
}

// String provides a human-readable representation of the forecast.
func (f *GreenHoursForecast) String() string {
	return fmt.Sprintf("%s: %d green hours (best: %.1f g CO2/kWh at %s)",
		f.Location, len(f.GreenHours), f.BestWindow.CarbonIntensity,
		f.BestWindow.Start.Format("15:04"))
}

// CarbonIntensityThresholds defines the thresholds used for carbon intensity classification.
type CarbonIntensityThresholds struct {
	// GreenThreshold is the upper limit for "green" classification (g CO2/kWh)
	GreenThreshold float64 `json:"green_threshold" validate:"min=0" example:"150"`

	// YellowThreshold is the upper limit for "yellow" classification (g CO2/kWh)
	YellowThreshold float64 `json:"yellow_threshold" validate:"min=0" example:"300"`

	// RedThreshold is the lower limit for "red" classification (g CO2/kWh)
	// Values above this threshold are considered high carbon intensity
	RedThreshold float64 `json:"red_threshold" validate:"min=0" example:"300"`
}

// DefaultThresholds provides the standard carbon intensity thresholds.
var DefaultThresholds = CarbonIntensityThresholds{
	GreenThreshold:  150,
	YellowThreshold: 300,
	RedThreshold:    300,
}

// ClassifyIntensity returns the mode classification for a given carbon intensity value.
func (t *CarbonIntensityThresholds) ClassifyIntensity(intensity float64) string {
	switch {
	case intensity < t.GreenThreshold:
		return "green"
	case intensity < t.YellowThreshold:
		return "yellow"
	default:
		return "red"
	}
}

// GetRecommendation returns a recommendation string based on carbon intensity.
func (t *CarbonIntensityThresholds) GetRecommendation(intensity float64) string {
	switch t.ClassifyIntensity(intensity) {
	case "green":
		return "optimal"
	case "yellow":
		return "reduce"
	default:
		return "defer"
	}
}

// DynamicCarbonIntensityThresholds defines thresholds that adapt to regional patterns.
type DynamicCarbonIntensityThresholds struct {
	// Region is the geographical region these thresholds apply to
	Region string `json:"region"`
	
	// GreenPercentile is the percentile threshold for "green" classification (typically 20th percentile)
	GreenPercentile float64 `json:"green_percentile" validate:"min=0,max=100" example:"20"`
	
	// RedPercentile is the percentile threshold for "red" classification (typically 80th percentile)
	RedPercentile float64 `json:"red_percentile" validate:"min=0,max=100" example:"80"`
	
	// AbsoluteGreenThreshold is the absolute upper limit for green (safety threshold)
	AbsoluteGreenThreshold float64 `json:"absolute_green_threshold" validate:"min=0" example:"200"`
	
	// AbsoluteRedThreshold is the absolute lower limit for red (safety threshold)
	AbsoluteRedThreshold float64 `json:"absolute_red_threshold" validate:"min=0" example:"500"`
	
	// RegionalBaseline is the typical carbon intensity for this region
	RegionalBaseline float64 `json:"regional_baseline" validate:"min=0" example:"300"`
	
	// LastUpdated indicates when these thresholds were last calculated
	LastUpdated time.Time `json:"last_updated"`
	
	// Confidence indicates the reliability of these thresholds (0-100)
	Confidence float64 `json:"confidence" validate:"min=0,max=100" example:"85"`
}

// ClassifyIntensityDynamic returns the mode classification using dynamic thresholds.
func (t *DynamicCarbonIntensityThresholds) ClassifyIntensityDynamic(intensity float64, currentPercentile float64) string {
	// Use percentile-based classification with absolute safety limits
	if intensity <= t.AbsoluteGreenThreshold && currentPercentile <= t.GreenPercentile {
		return "green"
	} else if intensity >= t.AbsoluteRedThreshold || currentPercentile >= t.RedPercentile {
		return "red"
	} else {
		return "yellow"
	}
}

// GetDynamicRecommendation returns a recommendation based on dynamic thresholds.
func (t *DynamicCarbonIntensityThresholds) GetDynamicRecommendation(intensity float64, currentPercentile float64) string {
	switch t.ClassifyIntensityDynamic(intensity, currentPercentile) {
	case "green":
		return "optimal"
	case "yellow":
		return "reduce"
	default:
		return "defer"
	}
}

// RelativeMetrics contains relative carbon intensity metrics.
type RelativeMetrics struct {
	// LocalPercentile indicates where this reading falls in the regional distribution (0-100)
	LocalPercentile float64 `json:"local_percentile" validate:"min=0,max=100" example:"25.5"`
	
	// DailyRank provides a human-readable rank within today's readings
	DailyRank string `json:"daily_rank" example:"top 25% cleanest hour today"`
	
	// RelativeMode is the mode based on regional patterns ("clean", "average", "dirty")
	RelativeMode string `json:"relative_mode" validate:"oneof=clean average dirty" example:"clean"`
	
	// TrendDirection indicates if carbon intensity is improving, worsening, or stable
	TrendDirection string `json:"trend_direction" validate:"oneof=improving worsening stable" example:"improving"`
	
	// TrendMagnitude shows the percentage change from regional baseline
	TrendMagnitude float64 `json:"trend_magnitude" example:"-15.2"`
	
	// RegionalBaseline is the typical carbon intensity for this region
	RegionalBaseline float64 `json:"regional_baseline" validate:"min=0" example:"280.5"`
	
	// ConfidenceScore indicates reliability of relative metrics (0-100)
	ConfidenceScore float64 `json:"confidence_score" validate:"min=0,max=100" example:"78.5"`
	
	// IsHighVariation indicates if this is a high-variation region
	IsHighVariation bool `json:"is_high_variation" example:"true"`
}

// EnhancedCarbonIntensity extends CarbonIntensity with relative metrics and predictions.
type EnhancedCarbonIntensity struct {
	CarbonIntensity
	
	// RelativeMetrics provides regional context and percentile rankings
	RelativeMetrics RelativeMetrics `json:"relative_metrics"`
	
	// NextOptimalWindow predicts the next optimal time for energy consumption
	NextOptimalWindow *OptimalWindow `json:"next_optimal_window,omitempty"`
	
	// RegionalStrategy contains region-specific optimization advice
	RegionalStrategy *RegionalOptimization `json:"regional_strategy,omitempty"`
}

// OptimalWindow represents a predicted optimal time window.
type OptimalWindow struct {
	// Start is the beginning of the optimal window
	Start time.Time `json:"start" validate:"required" example:"2024-01-15T22:00:00Z"`
	
	// End is the end of the optimal window
	End time.Time `json:"end" validate:"required" example:"2024-01-15T23:00:00Z"`
	
	// ExpectedIntensity is the predicted carbon intensity during this window
	ExpectedIntensity float64 `json:"expected_intensity" validate:"min=0" example:"95.3"`
	
	// Confidence indicates prediction confidence (0-100)
	Confidence float64 `json:"confidence" validate:"min=0,max=100" example:"82.5"`
	
	// Reason explains why this window is optimal
	Reason string `json:"reason" example:"Night wind patterns"`
	
	// Duration provides the window duration as a convenience
	Duration time.Duration `json:"duration,omitempty" example:"1h0m0s"`
}

// GetDuration returns the duration of this optimal window.
func (w *OptimalWindow) GetDuration() time.Duration {
	if w.Duration != 0 {
		return w.Duration
	}
	return w.End.Sub(w.Start)
}

// RegionalOptimization contains region-specific optimization strategies.
type RegionalOptimization struct {
	// Region identifier
	Region string `json:"region" example:"PL"`
	
	// PrimaryEnergySource describes the main energy source
	PrimaryEnergySource string `json:"primary_energy_source" example:"coal"`
	
	// OptimalHours lists the generally best hours for this region
	OptimalHours []int `json:"optimal_hours" example:"[22,23,0,1,2,3]"`
	
	// AvoidanceHours lists hours to generally avoid
	AvoidanceHours []int `json:"avoidance_hours" example:"[17,18,19,20]"`
	
	// VariationLevel indicates how much carbon intensity varies
	VariationLevel string `json:"variation_level" validate:"oneof=low medium high" example:"high"`
	
	// Recommendations provides strategic advice for this region
	Recommendations []string `json:"recommendations"`
}