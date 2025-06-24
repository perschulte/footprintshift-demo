// Package carbon provides adapters for bridging existing services with the intelligence service.
package carbon

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/perschulte/greenweb-api/pkg/carbon"
	"github.com/perschulte/greenweb-api/service"
)

// ElectricityMapsAdapter adapts the ElectricityMapsClient to provide historical data capabilities.
type ElectricityMapsAdapter struct {
	client ElectricityMapsService
}

// ElectricityMapsService defines the interface for electricity maps operations
type ElectricityMapsService interface {
	GetCarbonIntensity(ctx context.Context, location string) (*service.CarbonIntensity, error)
	GetGreenHoursForecast(ctx context.Context, location string, hours int) (*service.GreenHoursForecast, error)
	IsHealthy(ctx context.Context) bool
}

// NewElectricityMapsAdapter creates a new adapter for electricity maps service.
func NewElectricityMapsAdapter(client ElectricityMapsService) *ElectricityMapsAdapter {
	return &ElectricityMapsAdapter{
		client: client,
	}
}

// GetCarbonIntensity retrieves current carbon intensity.
func (a *ElectricityMapsAdapter) GetCarbonIntensity(ctx context.Context, location string) (*carbon.CarbonIntensity, error) {
	intensity, err := a.client.GetCarbonIntensity(ctx, location)
	if err != nil {
		return nil, err
	}
	
	// Convert service.CarbonIntensity to carbon.CarbonIntensity
	return &carbon.CarbonIntensity{
		Location:             intensity.Location,
		CarbonIntensity:      intensity.CarbonIntensity,
		RenewablePercent:     intensity.RenewablePercent,
		Mode:                 intensity.Mode,
		Recommendation:       intensity.Recommendation,
		NextGreenWindow:      intensity.NextGreenWindow,
		Timestamp:            intensity.Timestamp,
		FossilFuelPercentage: intensity.FossilFuelPercentage,
		Source:               intensity.Source,
		GridZone:             intensity.GridZone,
	}, nil
}

// GetGreenHoursForecast retrieves green hours forecast.
func (a *ElectricityMapsAdapter) GetGreenHoursForecast(ctx context.Context, location string, hours int) (*carbon.GreenHoursForecast, error) {
	forecast, err := a.client.GetGreenHoursForecast(ctx, location, hours)
	if err != nil {
		return nil, err
	}
	
	// Convert service.GreenHoursForecast to carbon.GreenHoursForecast
	greenHours := make([]carbon.GreenHour, len(forecast.GreenHours))
	for i, hour := range forecast.GreenHours {
		greenHours[i] = carbon.GreenHour{
			Start:            hour.Start,
			End:              hour.End,
			CarbonIntensity:  hour.CarbonIntensity,
			RenewablePercent: hour.RenewablePercent,
			Confidence:       hour.Confidence,
			Duration:         hour.Duration,
		}
	}
	
	return &carbon.GreenHoursForecast{
		Location:   forecast.Location,
		GreenHours: greenHours,
		BestWindow: carbon.GreenHour{
			Start:            forecast.BestWindow.Start,
			End:              forecast.BestWindow.End,
			CarbonIntensity:  forecast.BestWindow.CarbonIntensity,
			RenewablePercent: forecast.BestWindow.RenewablePercent,
			Confidence:       forecast.BestWindow.Confidence,
			Duration:         forecast.BestWindow.Duration,
		},
		ForecastPeriod: struct {
			Start time.Time `json:"start" example:"2024-01-15T14:00:00Z"`
			End   time.Time `json:"end" example:"2024-01-16T14:00:00Z"`
		}{
			Start: forecast.ForecastPeriod.Start,
			End:   forecast.ForecastPeriod.End,
		},
		GeneratedAt:      forecast.GeneratedAt,
		Source:           forecast.Source,
		Confidence:       forecast.Confidence,
		AverageIntensity: forecast.AverageIntensity,
	}, nil
}

// IsHealthy checks service health.
func (a *ElectricityMapsAdapter) IsHealthy(ctx context.Context) bool {
	return a.client.IsHealthy(ctx)
}

// GetSupportedLocations returns supported locations.
func (a *ElectricityMapsAdapter) GetSupportedLocations(ctx context.Context) ([]string, error) {
	// This would need to be implemented in the underlying service
	// For now, return a reasonable set of supported locations
	return []string{
		"DE", "FR", "GB", "ES", "IT", "NL", "AT", "SE", "NO", "DK", "FI", "BE", "CH", "IE", "PT",
		"PL", "CZ", "HU", "RO", "BG", "GR", "US-NY", "US-CA", "US-TEX", "US-FLA", "CA-ON", "CA-BC",
		"AU-NSW", "AU-VIC", "JP",
	}, nil
}

// GetHistoricalCarbonIntensity generates mock historical data.
// Note: This is a placeholder implementation. In production, you would integrate with
// a service that provides actual historical carbon intensity data.
func (a *ElectricityMapsAdapter) GetHistoricalCarbonIntensity(ctx context.Context, location string, start, end time.Time) ([]carbon.CarbonIntensity, error) {
	// Generate realistic mock historical data
	var historical []carbon.CarbonIntensity
	
	// Get current intensity as baseline
	current, err := a.GetCarbonIntensity(ctx, location)
	if err != nil {
		return nil, fmt.Errorf("failed to get current intensity for baseline: %w", err)
	}
	
	baseline := current.CarbonIntensity
	
	// Generate hourly data points
	duration := end.Sub(start)
	hours := int(duration.Hours())
	
	if hours > 2160 { // Limit to 90 days for performance
		hours = 2160
		start = end.Add(-90 * 24 * time.Hour)
	}
	
	// Seed random number generator for consistent results
	rng := rand.New(rand.NewSource(int64(len(location)) + start.Unix()))
	
	for i := 0; i < hours; i++ {
		timestamp := start.Add(time.Duration(i) * time.Hour)
		hour := timestamp.Hour()
		dayOfWeek := timestamp.Weekday()
		
		// Create realistic patterns based on time
		var intensity float64
		var renewable float64
		
		// Daily patterns - lower at night and midday, higher during peak hours
		hourFactor := 1.0
		if hour >= 22 || hour <= 6 {
			// Night hours - more wind power, less demand
			hourFactor = 0.7 + rng.Float64()*0.2 // 0.7-0.9
		} else if hour >= 10 && hour <= 16 {
			// Midday - solar power peak
			hourFactor = 0.8 + rng.Float64()*0.2 // 0.8-1.0
		} else {
			// Peak hours - more demand
			hourFactor = 1.1 + rng.Float64()*0.3 // 1.1-1.4
		}
		
		// Weekly patterns - weekends typically lower
		weekFactor := 1.0
		if dayOfWeek == time.Saturday || dayOfWeek == time.Sunday {
			weekFactor = 0.85 + rng.Float64()*0.15 // 0.85-1.0
		}
		
		// Add seasonal variation (simplified)
		monthFactor := 1.0
		month := timestamp.Month()
		if month >= 6 && month <= 8 {
			// Summer - more cooling demand, more solar
			monthFactor = 1.05 + rng.Float64()*0.1
		} else if month >= 12 || month <= 2 {
			// Winter - more heating demand
			monthFactor = 1.15 + rng.Float64()*0.1
		}
		
		// Add random variation
		randomFactor := 0.9 + rng.Float64()*0.2 // 0.9-1.1
		
		intensity = baseline * hourFactor * weekFactor * monthFactor * randomFactor
		
		// Ensure reasonable bounds
		intensity = math.Max(50, math.Min(800, intensity))
		
		// Calculate renewable percentage (inverse relationship with intensity)
		maxRenewable := 95.0
		minRenewable := 15.0
		// Normalize intensity for renewable calculation
		normalizedIntensity := (intensity - 50) / (800 - 50)
		renewable = maxRenewable - (normalizedIntensity * (maxRenewable - minRenewable))
		renewable = math.Max(minRenewable, math.Min(maxRenewable, renewable))
		
		// Determine mode based on intensity
		var mode string
		var recommendation string
		if intensity < 150 {
			mode = "green"
			recommendation = "optimal"
		} else if intensity < 300 {
			mode = "yellow"
			recommendation = "reduce"
		} else {
			mode = "red"
			recommendation = "defer"
		}
		
		historical = append(historical, carbon.CarbonIntensity{
			Location:             location,
			CarbonIntensity:      intensity,
			RenewablePercent:     renewable,
			Mode:                 mode,
			Recommendation:       recommendation,
			NextGreenWindow:      timestamp.Add(time.Hour * time.Duration(rng.Intn(8)+1)),
			Timestamp:            timestamp,
			FossilFuelPercentage: 100 - renewable,
			Source:               "mock_historical",
			GridZone:             current.GridZone,
		})
	}
	
	return historical, nil
}

// GetAverageCarbonIntensity calculates average over a period.
func (a *ElectricityMapsAdapter) GetAverageCarbonIntensity(ctx context.Context, location string, start, end time.Time) (float64, error) {
	historical, err := a.GetHistoricalCarbonIntensity(ctx, location, start, end)
	if err != nil {
		return 0, err
	}
	
	if len(historical) == 0 {
		return 0, fmt.Errorf("no historical data available for the specified period")
	}
	
	var sum float64
	for _, h := range historical {
		sum += h.CarbonIntensity
	}
	
	return sum / float64(len(historical)), nil
}

// GetCachedCarbonIntensity returns cached data (not implemented in this adapter).
func (a *ElectricityMapsAdapter) GetCachedCarbonIntensity(ctx context.Context, location string, maxAge time.Duration) (*carbon.CarbonIntensity, error) {
	return nil, fmt.Errorf("caching not implemented in adapter")
}

// ClearCache clears cache (not implemented in this adapter).
func (a *ElectricityMapsAdapter) ClearCache(ctx context.Context, location string) error {
	return fmt.Errorf("caching not implemented in adapter")
}