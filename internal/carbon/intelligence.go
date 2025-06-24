// Package carbon provides intelligent carbon intensity monitoring with dynamic thresholds.
package carbon

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"sort"
	"sync"
	"time"

	"github.com/perschulte/greenweb-api/pkg/carbon"
)

// IntelligenceService provides dynamic carbon intensity analysis with region-specific thresholds.
type IntelligenceService struct {
	// Dependencies
	dataSource carbon.CarbonServiceWithHistory
	logger     *slog.Logger
	
	// Historical data storage
	mu             sync.RWMutex
	regionPatterns map[string]*RegionPattern
	
	// Configuration
	config IntelligenceConfig
}

// IntelligenceConfig contains configuration for the intelligence service.
type IntelligenceConfig struct {
	// HistoryRetentionDays controls how long historical data is kept
	HistoryRetentionDays int
	
	// MinDataPointsForAnalysis is the minimum number of data points needed for reliable analysis
	MinDataPointsForAnalysis int
	
	// UpdateInterval controls how often patterns are recalculated
	UpdateInterval time.Duration
	
	// SeasonalAdjustment enables seasonal pattern recognition
	SeasonalAdjustment bool
	
	// HighVariationRegions lists regions with known high carbon variation
	HighVariationRegions []string
}

// RegionPattern stores learned patterns for a specific region.
type RegionPattern struct {
	Region           string
	LastUpdated      time.Time
	DataPoints       []HistoricalDataPoint
	
	// Statistical measures
	Mean             float64
	StdDev           float64
	P20              float64  // 20th percentile (clean threshold)
	P80              float64  // 80th percentile (dirty threshold)
	
	// Time-based patterns
	HourlyAverages   [24]float64
	DayOfWeekPattern [7]DayPattern
	SeasonalFactors  map[string]float64
	
	// Trends
	TrendDirection   string  // "improving", "worsening", "stable"
	TrendConfidence  float64
}

// HistoricalDataPoint represents a single carbon intensity measurement.
type HistoricalDataPoint struct {
	Timestamp        time.Time
	CarbonIntensity  float64
	RenewablePercent float64
}

// DayPattern represents carbon intensity patterns for a specific day of the week.
type DayPattern struct {
	HourlyIntensity [24]float64
	PeakHours       []int
	CleanHours      []int
}

// RelativeCarbonIntensity extends carbon.CarbonIntensity with relative metrics.
type RelativeCarbonIntensity struct {
	carbon.CarbonIntensity
	
	// Relative metrics
	LocalPercentile  float64 `json:"local_percentile"`  // 0-100, where 0 is cleanest
	DailyRank        string  `json:"daily_rank"`        // e.g., "top 15% cleanest hour today"
	RelativeMode     string  `json:"relative_mode"`     // "clean", "average", "dirty" based on local patterns
	
	// Trend information
	TrendDirection   string  `json:"trend_direction"`   // "improving", "worsening", "stable"
	TrendMagnitude   float64 `json:"trend_magnitude"`   // Percentage change
	
	// Predictions
	NextOptimalWindow *OptimalWindow `json:"next_optimal_window,omitempty"`
	ConfidenceScore   float64        `json:"confidence_score"`
	
	// Regional context
	RegionalBaseline  float64 `json:"regional_baseline"`
	IsHighVariation   bool    `json:"is_high_variation"`
}

// OptimalWindow represents a predicted optimal time window.
type OptimalWindow struct {
	Start            time.Time `json:"start"`
	End              time.Time `json:"end"`
	ExpectedIntensity float64   `json:"expected_intensity"`
	Confidence       float64   `json:"confidence"`
	Reason           string    `json:"reason"` // e.g., "Night wind patterns", "Weekend low demand"
}

// CarbonTrend represents historical carbon intensity trends.
type CarbonTrend struct {
	Location         string                  `json:"location"`
	Period           string                  `json:"period"` // "daily", "weekly", "monthly"
	StartDate        time.Time               `json:"start_date"`
	EndDate          time.Time               `json:"end_date"`
	
	// Statistical summary
	AverageIntensity float64                 `json:"average_intensity"`
	MinIntensity     float64                 `json:"min_intensity"`
	MaxIntensity     float64                 `json:"max_intensity"`
	StdDeviation     float64                 `json:"std_deviation"`
	
	// Patterns
	CleanestHours    []int                   `json:"cleanest_hours"`    // Hours of day with lowest average
	DirtiestHours    []int                   `json:"dirtiest_hours"`    // Hours of day with highest average
	WeekdayVsWeekend map[string]float64      `json:"weekday_vs_weekend"` // Average intensity comparison
	
	// Time series data
	DataPoints       []HistoricalDataPoint   `json:"data_points,omitempty"`
}

// DefaultIntelligenceConfig provides sensible defaults.
var DefaultIntelligenceConfig = IntelligenceConfig{
	HistoryRetentionDays:     30,
	MinDataPointsForAnalysis: 168, // 1 week of hourly data
	UpdateInterval:           15 * time.Minute,
	SeasonalAdjustment:       true,
	HighVariationRegions: []string{
		"PL", "Poland",
		"US-TEX", "Texas",
		"CN", "China",
		"IN", "India",
		"AU-NSW", "Australia-NSW",
		"ZA", "South Africa",
	},
}

// NewIntelligenceService creates a new carbon intelligence service.
func NewIntelligenceService(dataSource carbon.CarbonServiceWithHistory, logger *slog.Logger, config *IntelligenceConfig) *IntelligenceService {
	if config == nil {
		config = &DefaultIntelligenceConfig
	}
	
	service := &IntelligenceService{
		dataSource:     dataSource,
		logger:         logger,
		config:         *config,
		regionPatterns: make(map[string]*RegionPattern),
	}
	
	// Start background pattern update
	go service.startPatternUpdater()
	
	return service
}

// GetRelativeCarbonIntensity returns carbon intensity with relative metrics.
func (s *IntelligenceService) GetRelativeCarbonIntensity(ctx context.Context, location string) (*RelativeCarbonIntensity, error) {
	// Get current absolute intensity
	current, err := s.dataSource.GetCarbonIntensity(ctx, location)
	if err != nil {
		return nil, fmt.Errorf("failed to get current intensity: %w", err)
	}
	
	// Get or update regional pattern
	pattern, err := s.getOrUpdatePattern(ctx, location)
	if err != nil {
		s.logger.Warn("Failed to get regional pattern, using absolute values only", 
			"location", location, 
			"error", err)
		// Return basic response without relative metrics
		return &RelativeCarbonIntensity{
			CarbonIntensity: *current,
			ConfidenceScore: 0.5,
		}, nil
	}
	
	// Calculate relative metrics
	relative := s.calculateRelativeMetrics(current, pattern)
	
	// Add predictions
	relative.NextOptimalWindow = s.predictNextOptimalWindow(pattern, time.Now())
	
	// Check if this is a high variation region
	relative.IsHighVariation = s.isHighVariationRegion(location)
	
	return relative, nil
}

// GetDynamicGreenHours returns green hours based on dynamic thresholds.
func (s *IntelligenceService) GetDynamicGreenHours(ctx context.Context, location string, hours int) (*carbon.GreenHoursForecast, error) {
	// Get regional pattern
	pattern, err := s.getOrUpdatePattern(ctx, location)
	if err != nil {
		// Fall back to standard service if pattern unavailable
		return s.dataSource.GetGreenHoursForecast(ctx, location, hours)
	}
	
	// Get forecast from base service
	forecast, err := s.dataSource.GetGreenHoursForecast(ctx, location, hours)
	if err != nil {
		return nil, err
	}
	
	// Apply dynamic thresholds to identify truly green hours
	dynamicGreenHours := s.applyDynamicThresholds(forecast.GreenHours, pattern)
	
	// Update forecast with dynamic green hours
	forecast.GreenHours = dynamicGreenHours
	if len(dynamicGreenHours) > 0 {
		// Find best window based on dynamic analysis
		bestWindow := dynamicGreenHours[0]
		for _, hour := range dynamicGreenHours {
			if hour.CarbonIntensity < bestWindow.CarbonIntensity {
				bestWindow = hour
			}
		}
		forecast.BestWindow = bestWindow
	}
	
	return forecast, nil
}

// GetCarbonTrends returns historical carbon intensity trends.
func (s *IntelligenceService) GetCarbonTrends(ctx context.Context, location string, period string, days int) (*CarbonTrend, error) {
	endTime := time.Now()
	startTime := endTime.AddDate(0, 0, -days)
	
	// Fetch historical data
	historical, err := s.dataSource.GetHistoricalCarbonIntensity(ctx, location, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to get historical data: %w", err)
	}
	
	if len(historical) < s.config.MinDataPointsForAnalysis {
		return nil, fmt.Errorf("insufficient data for trend analysis: got %d points, need %d", 
			len(historical), s.config.MinDataPointsForAnalysis)
	}
	
	// Convert to data points
	dataPoints := make([]HistoricalDataPoint, len(historical))
	for i, h := range historical {
		dataPoints[i] = HistoricalDataPoint{
			Timestamp:        h.Timestamp,
			CarbonIntensity:  h.CarbonIntensity,
			RenewablePercent: h.RenewablePercent,
		}
	}
	
	// Calculate trends
	trend := s.analyzeTrends(location, period, dataPoints)
	trend.StartDate = startTime
	trend.EndDate = endTime
	
	return trend, nil
}

// calculateRelativeMetrics calculates relative carbon intensity metrics.
func (s *IntelligenceService) calculateRelativeMetrics(current *carbon.CarbonIntensity, pattern *RegionPattern) *RelativeCarbonIntensity {
	relative := &RelativeCarbonIntensity{
		CarbonIntensity:  *current,
		RegionalBaseline: pattern.Mean,
	}
	
	// Calculate local percentile
	relative.LocalPercentile = s.calculatePercentile(current.CarbonIntensity, pattern)
	
	// Determine daily rank
	if relative.LocalPercentile <= 20 {
		relative.DailyRank = fmt.Sprintf("top %d%% cleanest", int(relative.LocalPercentile))
	} else if relative.LocalPercentile >= 80 {
		relative.DailyRank = fmt.Sprintf("top %d%% dirtiest", int(100-relative.LocalPercentile))
	} else {
		relative.DailyRank = "average for this region"
	}
	
	// Set relative mode based on regional patterns
	if current.CarbonIntensity <= pattern.P20 {
		relative.RelativeMode = "clean"
	} else if current.CarbonIntensity >= pattern.P80 {
		relative.RelativeMode = "dirty"
	} else {
		relative.RelativeMode = "average"
	}
	
	// Calculate trend
	relative.TrendDirection = pattern.TrendDirection
	if pattern.Mean > 0 {
		relative.TrendMagnitude = ((current.CarbonIntensity - pattern.Mean) / pattern.Mean) * 100
	}
	
	// Set confidence based on data quality
	relative.ConfidenceScore = s.calculateConfidence(pattern)
	
	return relative
}

// calculatePercentile calculates the percentile rank of a value within the regional pattern.
func (s *IntelligenceService) calculatePercentile(value float64, pattern *RegionPattern) float64 {
	if len(pattern.DataPoints) == 0 {
		return 50.0 // Default to median if no data
	}
	
	// Sort intensities
	intensities := make([]float64, len(pattern.DataPoints))
	for i, dp := range pattern.DataPoints {
		intensities[i] = dp.CarbonIntensity
	}
	sort.Float64s(intensities)
	
	// Find position
	position := 0
	for _, intensity := range intensities {
		if value > intensity {
			position++
		} else {
			break
		}
	}
	
	percentile := (float64(position) / float64(len(intensities))) * 100
	return math.Round(percentile*10) / 10 // Round to 1 decimal place
}

// applyDynamicThresholds filters green hours based on regional patterns.
func (s *IntelligenceService) applyDynamicThresholds(hours []carbon.GreenHour, pattern *RegionPattern) []carbon.GreenHour {
	var dynamicGreenHours []carbon.GreenHour
	
	for _, hour := range hours {
		// Use dynamic threshold (20th percentile) instead of static 150
		if hour.CarbonIntensity <= pattern.P20 {
			// This is truly green for this region
			dynamicGreenHours = append(dynamicGreenHours, hour)
		} else if hour.CarbonIntensity < pattern.Mean && pattern.StdDev > 50 {
			// For high-variation regions, also include below-average hours
			hour.Confidence = hour.Confidence * 0.8 // Lower confidence
			dynamicGreenHours = append(dynamicGreenHours, hour)
		}
	}
	
	return dynamicGreenHours
}

// predictNextOptimalWindow predicts the next optimal time window.
func (s *IntelligenceService) predictNextOptimalWindow(pattern *RegionPattern, from time.Time) *OptimalWindow {
	if pattern == nil || len(pattern.HourlyAverages) == 0 {
		return nil
	}
	
	// Find next hour with historically low intensity
	currentHour := from.Hour()
	var bestHour int
	var lowestIntensity float64 = math.MaxFloat64
	
	// Look ahead 24 hours
	for i := 1; i <= 24; i++ {
		hour := (currentHour + i) % 24
		if pattern.HourlyAverages[hour] < lowestIntensity {
			lowestIntensity = pattern.HourlyAverages[hour]
			bestHour = hour
		}
	}
	
	// Calculate start time for best hour
	nextTime := from.Add(time.Duration(bestHour-currentHour) * time.Hour)
	if bestHour <= currentHour {
		nextTime = nextTime.Add(24 * time.Hour)
	}
	nextTime = time.Date(nextTime.Year(), nextTime.Month(), nextTime.Day(), bestHour, 0, 0, 0, nextTime.Location())
	
	// Determine reason based on time
	reason := "Historical low-carbon period"
	if bestHour >= 22 || bestHour <= 6 {
		reason = "Night wind patterns"
	} else if bestHour >= 10 && bestHour <= 16 {
		reason = "Solar generation peak"
	}
	
	return &OptimalWindow{
		Start:             nextTime,
		End:               nextTime.Add(time.Hour),
		ExpectedIntensity: lowestIntensity,
		Confidence:        s.calculateConfidence(pattern),
		Reason:            reason,
	}
}

// analyzeTrends analyzes historical data to identify trends.
func (s *IntelligenceService) analyzeTrends(location, period string, dataPoints []HistoricalDataPoint) *CarbonTrend {
	trend := &CarbonTrend{
		Location: location,
		Period:   period,
	}
	
	if len(dataPoints) == 0 {
		return trend
	}
	
	// Calculate basic statistics
	var sum, min, max float64
	min = math.MaxFloat64
	
	hourlyTotals := make(map[int][]float64)
	weekdayTotals := []float64{}
	weekendTotals := []float64{}
	
	for _, dp := range dataPoints {
		sum += dp.CarbonIntensity
		if dp.CarbonIntensity < min {
			min = dp.CarbonIntensity
		}
		if dp.CarbonIntensity > max {
			max = dp.CarbonIntensity
		}
		
		// Collect hourly data
		hour := dp.Timestamp.Hour()
		hourlyTotals[hour] = append(hourlyTotals[hour], dp.CarbonIntensity)
		
		// Weekday vs weekend
		if dp.Timestamp.Weekday() == time.Saturday || dp.Timestamp.Weekday() == time.Sunday {
			weekendTotals = append(weekendTotals, dp.CarbonIntensity)
		} else {
			weekdayTotals = append(weekdayTotals, dp.CarbonIntensity)
		}
	}
	
	trend.AverageIntensity = sum / float64(len(dataPoints))
	trend.MinIntensity = min
	trend.MaxIntensity = max
	
	// Calculate standard deviation
	var varianceSum float64
	for _, dp := range dataPoints {
		diff := dp.CarbonIntensity - trend.AverageIntensity
		varianceSum += diff * diff
	}
	trend.StdDeviation = math.Sqrt(varianceSum / float64(len(dataPoints)))
	
	// Find cleanest and dirtiest hours
	hourlyAverages := make(map[int]float64)
	for hour, values := range hourlyTotals {
		if len(values) > 0 {
			hourSum := 0.0
			for _, v := range values {
				hourSum += v
			}
			hourlyAverages[hour] = hourSum / float64(len(values))
		}
	}
	
	// Sort hours by average intensity
	type hourAvg struct {
		hour int
		avg  float64
	}
	var hourList []hourAvg
	for h, a := range hourlyAverages {
		hourList = append(hourList, hourAvg{h, a})
	}
	sort.Slice(hourList, func(i, j int) bool {
		return hourList[i].avg < hourList[j].avg
	})
	
	// Get cleanest and dirtiest hours
	if len(hourList) >= 3 {
		for i := 0; i < 3 && i < len(hourList); i++ {
			trend.CleanestHours = append(trend.CleanestHours, hourList[i].hour)
			trend.DirtiestHours = append(trend.DirtiestHours, hourList[len(hourList)-1-i].hour)
		}
	}
	
	// Calculate weekday vs weekend averages
	trend.WeekdayVsWeekend = make(map[string]float64)
	if len(weekdayTotals) > 0 {
		weekdaySum := 0.0
		for _, v := range weekdayTotals {
			weekdaySum += v
		}
		trend.WeekdayVsWeekend["weekday_average"] = weekdaySum / float64(len(weekdayTotals))
	}
	if len(weekendTotals) > 0 {
		weekendSum := 0.0
		for _, v := range weekendTotals {
			weekendSum += v
		}
		trend.WeekdayVsWeekend["weekend_average"] = weekendSum / float64(len(weekendTotals))
	}
	
	// Include limited data points to avoid response bloat
	if period == "daily" && len(dataPoints) <= 24 {
		trend.DataPoints = dataPoints
	}
	
	return trend
}

// getOrUpdatePattern retrieves or updates the regional pattern.
func (s *IntelligenceService) getOrUpdatePattern(ctx context.Context, location string) (*RegionPattern, error) {
	s.mu.RLock()
	pattern, exists := s.regionPatterns[location]
	s.mu.RUnlock()
	
	// Check if pattern needs update
	needsUpdate := !exists || time.Since(pattern.LastUpdated) > s.config.UpdateInterval
	
	if needsUpdate {
		newPattern, err := s.updateRegionalPattern(ctx, location)
		if err != nil {
			return pattern, err // Return existing pattern if update fails
		}
		pattern = newPattern
	}
	
	return pattern, nil
}

// updateRegionalPattern fetches historical data and updates the pattern.
func (s *IntelligenceService) updateRegionalPattern(ctx context.Context, location string) (*RegionPattern, error) {
	// Fetch historical data
	endTime := time.Now()
	startTime := endTime.AddDate(0, 0, -s.config.HistoryRetentionDays)
	
	historical, err := s.dataSource.GetHistoricalCarbonIntensity(ctx, location, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch historical data: %w", err)
	}
	
	if len(historical) < s.config.MinDataPointsForAnalysis {
		return nil, fmt.Errorf("insufficient historical data: %d points", len(historical))
	}
	
	// Convert to data points
	dataPoints := make([]HistoricalDataPoint, len(historical))
	intensities := make([]float64, len(historical))
	
	for i, h := range historical {
		dataPoints[i] = HistoricalDataPoint{
			Timestamp:        h.Timestamp,
			CarbonIntensity:  h.CarbonIntensity,
			RenewablePercent: h.RenewablePercent,
		}
		intensities[i] = h.CarbonIntensity
	}
	
	// Calculate statistics
	pattern := &RegionPattern{
		Region:      location,
		LastUpdated: time.Now(),
		DataPoints:  dataPoints,
	}
	
	// Calculate mean
	sum := 0.0
	for _, intensity := range intensities {
		sum += intensity
	}
	pattern.Mean = sum / float64(len(intensities))
	
	// Calculate standard deviation
	varianceSum := 0.0
	for _, intensity := range intensities {
		diff := intensity - pattern.Mean
		varianceSum += diff * diff
	}
	pattern.StdDev = math.Sqrt(varianceSum / float64(len(intensities)))
	
	// Calculate percentiles
	sort.Float64s(intensities)
	pattern.P20 = intensities[int(float64(len(intensities))*0.2)]
	pattern.P80 = intensities[int(float64(len(intensities))*0.8)]
	
	// Calculate hourly averages
	hourlyData := make(map[int][]float64)
	for _, dp := range dataPoints {
		hour := dp.Timestamp.Hour()
		hourlyData[hour] = append(hourlyData[hour], dp.CarbonIntensity)
	}
	
	for hour := 0; hour < 24; hour++ {
		if data, exists := hourlyData[hour]; exists && len(data) > 0 {
			hourSum := 0.0
			for _, v := range data {
				hourSum += v
			}
			pattern.HourlyAverages[hour] = hourSum / float64(len(data))
		} else {
			pattern.HourlyAverages[hour] = pattern.Mean // Default to mean if no data
		}
	}
	
	// Analyze trend
	pattern.TrendDirection = s.analyzeTrendDirection(dataPoints)
	pattern.TrendConfidence = s.calculateTrendConfidence(dataPoints)
	
	// Store updated pattern
	s.mu.Lock()
	s.regionPatterns[location] = pattern
	s.mu.Unlock()
	
	s.logger.Info("Updated regional pattern", 
		"location", location,
		"mean", pattern.Mean,
		"stddev", pattern.StdDev,
		"p20", pattern.P20,
		"p80", pattern.P80,
		"trend", pattern.TrendDirection)
	
	return pattern, nil
}

// analyzeTrendDirection determines if carbon intensity is improving, worsening, or stable.
func (s *IntelligenceService) analyzeTrendDirection(dataPoints []HistoricalDataPoint) string {
	if len(dataPoints) < 2 {
		return "stable"
	}
	
	// Simple linear regression to determine trend
	n := float64(len(dataPoints))
	var sumX, sumY, sumXY, sumX2 float64
	
	for i, dp := range dataPoints {
		x := float64(i)
		y := dp.CarbonIntensity
		sumX += x
		sumY += y
		sumXY += x * y
		sumX2 += x * x
	}
	
	// Calculate slope
	slope := (n*sumXY - sumX*sumY) / (n*sumX2 - sumX*sumX)
	
	// Determine trend based on slope
	if math.Abs(slope) < 0.1 {
		return "stable"
	} else if slope < 0 {
		return "improving"
	} else {
		return "worsening"
	}
}

// calculateConfidence calculates confidence score based on data quality.
func (s *IntelligenceService) calculateConfidence(pattern *RegionPattern) float64 {
	if pattern == nil || len(pattern.DataPoints) == 0 {
		return 0.0
	}
	
	// Base confidence on data completeness
	expectedPoints := s.config.HistoryRetentionDays * 24
	dataCompleteness := math.Min(float64(len(pattern.DataPoints))/float64(expectedPoints), 1.0)
	
	// Factor in data recency
	recencyFactor := 1.0
	if len(pattern.DataPoints) > 0 {
		lastDataPoint := pattern.DataPoints[len(pattern.DataPoints)-1].Timestamp
		hoursSinceLastData := time.Since(lastDataPoint).Hours()
		if hoursSinceLastData > 24 {
			recencyFactor = math.Max(0.5, 1.0-hoursSinceLastData/168) // Decay over a week
		}
	}
	
	// Factor in variation (lower variation = higher confidence)
	variationFactor := 1.0
	if pattern.Mean > 0 {
		coefficientOfVariation := pattern.StdDev / pattern.Mean
		variationFactor = math.Max(0.5, 1.0-coefficientOfVariation/2)
	}
	
	confidence := dataCompleteness * recencyFactor * variationFactor
	return math.Round(confidence*100) / 100 // Round to 2 decimal places
}

// calculateTrendConfidence calculates confidence in the trend analysis.
func (s *IntelligenceService) calculateTrendConfidence(dataPoints []HistoricalDataPoint) float64 {
	if len(dataPoints) < s.config.MinDataPointsForAnalysis {
		return 0.0
	}
	
	// More data points = higher confidence
	dataFactor := math.Min(float64(len(dataPoints))/float64(s.config.MinDataPointsForAnalysis*4), 1.0)
	
	// Check for consistency in the trend
	// This is simplified - a real implementation would use R-squared or similar
	return dataFactor * 0.8 // Max 80% confidence for now
}

// isHighVariationRegion checks if a location is known for high carbon variation.
func (s *IntelligenceService) isHighVariationRegion(location string) bool {
	for _, region := range s.config.HighVariationRegions {
		if location == region {
			return true
		}
	}
	return false
}

// startPatternUpdater runs periodic pattern updates in the background.
func (s *IntelligenceService) startPatternUpdater() {
	ticker := time.NewTicker(s.config.UpdateInterval)
	defer ticker.Stop()
	
	for range ticker.C {
		s.mu.RLock()
		locations := make([]string, 0, len(s.regionPatterns))
		for location := range s.regionPatterns {
			locations = append(locations, location)
		}
		s.mu.RUnlock()
		
		// Update patterns for all tracked locations
		for _, location := range locations {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			_, err := s.updateRegionalPattern(ctx, location)
			if err != nil {
				s.logger.Error("Failed to update regional pattern", 
					"location", location, 
					"error", err)
			}
			cancel()
		}
	}
}