package impact

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// Service provides impact calculation and tracking functionality
type Service struct {
	calculator *Calculator
	storage    Storage
	metrics    *MetricsCollector
	mu         sync.RWMutex
}

// Storage interface for persisting impact data
type Storage interface {
	SaveBaseline(ctx context.Context, baseline *BaselineMeasurement) error
	GetBaseline(ctx context.Context, id string) (*BaselineMeasurement, error)
	SaveSession(ctx context.Context, session *SessionMetrics) error
	GetSession(ctx context.Context, sessionID string) (*SessionMetrics, error)
	SaveReport(ctx context.Context, report *ImpactReport) error
	GetReport(ctx context.Context, id string) (*ImpactReport, error)
	GetAggregatedMetrics(ctx context.Context, period ReportPeriod) (*AggregatedMetrics, error)
}

// AggregatedMetrics contains aggregated impact metrics
type AggregatedMetrics struct {
	TotalSavings      float64
	TotalBaseline     float64
	TotalOptimized    float64
	SessionCount      int
	OptimizationCount int
	Period            ReportPeriod
}

// NewService creates a new impact calculation service
func NewService(storage Storage) *Service {
	return &Service{
		calculator: NewCalculator(),
		storage:    storage,
		metrics:    NewMetricsCollector(),
	}
}

// CalculateImpact calculates the carbon impact for a given request
func (s *Service) CalculateImpact(ctx context.Context, req *CalculationRequest) (*ImpactResult, error) {
	result, err := s.calculator.Calculate(req)
	if err != nil {
		return nil, fmt.Errorf("calculation failed: %w", err)
	}

	// Track metrics
	s.metrics.RecordCalculation(req.Type, result.Savings)

	return result, nil
}

// MeasureBaseline measures and stores baseline carbon footprint for a URL
func (s *Service) MeasureBaseline(ctx context.Context, url string, params BaselineParams) (*BaselineMeasurement, error) {
	// Simulate measurement (in real implementation, would use browser automation)
	baseline := &BaselineMeasurement{
		ID:               generateID(),
		URL:              url,
		PageLoadEmissions: s.calculatePageLoadBaseline(params),
		DataTransferred:  params.DataTransferred,
		JavaScriptSize:   params.JavaScriptSize,
		ImageSize:        params.ImageSize,
		VideoSize:        params.VideoSize,
		LoadTime:         params.LoadTime,
		ResourceCount:    params.ResourceCount,
		MeasuredAt:       time.Now(),
		DeviceType:       params.DeviceType,
		ConnectionType:   params.ConnectionType,
		Region:           params.Region,
	}

	// Calculate hourly emissions estimate
	baseline.EstimatedHourlyEmissions = baseline.PageLoadEmissions * (3600.0 / baseline.LoadTime) * 0.1 // Assume 10% active time

	if err := s.storage.SaveBaseline(ctx, baseline); err != nil {
		return nil, fmt.Errorf("failed to save baseline: %w", err)
	}

	return baseline, nil
}

// BaselineParams contains parameters for baseline measurement
type BaselineParams struct {
	DataTransferred float64
	JavaScriptSize  float64
	ImageSize       float64
	VideoSize       float64
	LoadTime        float64
	ResourceCount   map[string]int
	DeviceType      string
	ConnectionType  string
	Region          string
}

// calculatePageLoadBaseline calculates baseline emissions for page load
func (s *Service) calculatePageLoadBaseline(params BaselineParams) float64 {
	req := &CalculationRequest{
		Type:           ImpactTypePageLoad,
		Duration:       params.LoadTime,
		DataSize:       params.DataTransferred,
		DeviceType:     params.DeviceType,
		ConnectionType: params.ConnectionType,
		Region:         params.Region,
	}

	result, err := s.calculator.Calculate(req)
	if err != nil {
		return 0
	}

	return result.BaselineEmissions
}

// ValidateSavings validates claimed carbon savings
func (s *Service) ValidateSavings(ctx context.Context, req *ValidationRequest) (*ValidationResult, error) {
	// Calculate what the savings should be
	calcReq := &CalculationRequest{
		Type:              req.OptimizationType,
		OptimizationLevel: 50, // Assume moderate optimization
	}

	// Extract parameters from the validation request
	if params, ok := req.Parameters["duration"].(float64); ok {
		calcReq.Duration = params
	}
	if params, ok := req.Parameters["data_size"].(float64); ok {
		calcReq.DataSize = params
	}
	if params, ok := req.Parameters["device_type"].(string); ok {
		calcReq.DeviceType = params
	}

	result, err := s.calculator.Calculate(calcReq)
	if err != nil {
		return nil, fmt.Errorf("validation calculation failed: %w", err)
	}

	// Compare with claimed savings
	variance := ((req.ClaimedSavings - result.Savings) / result.Savings) * 100
	
	// Determine rating
	rating := "reasonable"
	isValid := true
	
	if variance < -50 {
		rating = "conservative"
	} else if variance > 100 {
		rating = "unrealistic"
		isValid = false
	} else if variance > 50 {
		rating = "optimistic"
	}

	validationResult := &ValidationResult{
		IsValid:          isValid,
		ValidatedSavings: result.Savings,
		Variance:         variance,
		Rating:           rating,
		Explanation:      fmt.Sprintf("Your claimed savings of %.2f g CO2 compared to our calculated %.2f g CO2 (%.1f%% variance)", req.ClaimedSavings, result.Savings, variance),
	}

	// Add suggestions
	if !isValid {
		validationResult.Suggestions = []string{
			"Use conservative estimates to avoid greenwashing",
			"Include device energy consumption in calculations",
			"Account for rebound effects",
			"Use regional grid carbon intensity data",
		}
	}

	return validationResult, nil
}

// GenerateReport generates a comprehensive impact report for a time period
func (s *Service) GenerateReport(ctx context.Context, period ReportPeriod) (*ImpactReport, error) {
	// Get aggregated metrics from storage
	metrics, err := s.storage.GetAggregatedMetrics(ctx, period)
	if err != nil {
		return nil, fmt.Errorf("failed to get aggregated metrics: %w", err)
	}

	report := &ImpactReport{
		ID: generateID(),
		Period: period,
		TotalSavings: metrics.TotalSavings / 1000.0, // Convert to kg
		BaselineTotal: metrics.TotalBaseline / 1000.0,
		OptimizedTotal: metrics.TotalOptimized / 1000.0,
		OptimizationEvents: metrics.OptimizationCount,
		AverageOptimizationLevel: 50.0, // Placeholder
		ConfidenceScore: 75.0, // Conservative confidence
		GeneratedAt: time.Now(),
	}

	// Calculate equivalences
	report.EquivalentTo = s.calculateEquivalences(report.TotalSavings)

	// Add methodology
	report.Methodology = "Impact calculated using conservative emission factors from IEA, Carbon Trust, and EPA. Includes device, network, and data center emissions with ±25% confidence intervals."

	// Add recommendations
	report.Recommendations = s.generateRecommendations(metrics)

	if err := s.storage.SaveReport(ctx, report); err != nil {
		return nil, fmt.Errorf("failed to save report: %w", err)
	}

	return report, nil
}

// calculateEquivalences calculates relatable comparisons for CO2 savings
func (s *Service) calculateEquivalences(savingsKg float64) []Equivalence {
	return []Equivalence{
		{
			Type:        "driving",
			Value:       savingsKg * 2.3, // 1 kg CO2 = 2.3 miles in average car
			Unit:        "miles",
			Description: fmt.Sprintf("Equivalent to avoiding %.1f miles of driving", savingsKg*2.3),
		},
		{
			Type:        "trees",
			Value:       savingsKg / 21.0, // 1 tree absorbs ~21 kg CO2/year
			Unit:        "tree-years",
			Description: fmt.Sprintf("Equivalent to %.2f trees absorbing CO2 for a year", savingsKg/21.0),
		},
		{
			Type:        "smartphones",
			Value:       savingsKg / 0.008, // Charging smartphone = 8g CO2
			Unit:        "charges",
			Description: fmt.Sprintf("Equivalent to %.0f smartphone charges", savingsKg/0.008),
		},
	}
}

// generateRecommendations generates improvement recommendations
func (s *Service) generateRecommendations(metrics *AggregatedMetrics) []string {
	recommendations := []string{}

	savingsRate := (metrics.TotalSavings / metrics.TotalBaseline) * 100
	
	if savingsRate < 20 {
		recommendations = append(recommendations, 
			"Consider implementing more aggressive optimizations during high carbon periods",
			"Enable adaptive video quality based on grid carbon intensity",
			"Implement lazy loading for below-the-fold images")
	}

	if metrics.OptimizationCount < metrics.SessionCount/2 {
		recommendations = append(recommendations,
			"Increase optimization adoption - less than 50% of sessions are optimized",
			"Consider automatic optimization triggers based on carbon thresholds")
	}

	recommendations = append(recommendations,
		"Monitor peak usage hours and encourage shifting to low-carbon periods",
		"Implement caching strategies to reduce repeated data transfers")

	return recommendations
}

// TrackSession tracks carbon impact for a user session
func (s *Service) TrackSession(ctx context.Context, sessionID string) (*SessionTracker, error) {
	return &SessionTracker{
		service:   s,
		sessionID: sessionID,
		startTime: time.Now(),
		metrics: &SessionMetrics{
			SessionID:           sessionID,
			StartTime:           time.Now(),
			EmissionsByActivity: make(map[string]float64),
		},
	}, nil
}

// SessionTracker tracks emissions for an active session
type SessionTracker struct {
	service   *Service
	sessionID string
	startTime time.Time
	metrics   *SessionMetrics
	mu        sync.Mutex
}

// RecordActivity records an activity's emissions in the session
func (st *SessionTracker) RecordActivity(activityType string, emissions float64) {
	st.mu.Lock()
	defer st.mu.Unlock()

	st.metrics.EmissionsByActivity[activityType] += emissions
	st.metrics.TotalEmissions += emissions
}

// RecordOptimization records that an optimization was applied
func (st *SessionTracker) RecordOptimization(savings float64) {
	st.mu.Lock()
	defer st.mu.Unlock()

	st.metrics.OptimizationsApplied++
	st.metrics.EstimatedSavings += savings
}

// End ends the session and saves metrics
func (st *SessionTracker) End(ctx context.Context) error {
	st.mu.Lock()
	defer st.mu.Unlock()

	st.metrics.Duration = time.Since(st.startTime).Seconds()
	
	return st.service.storage.SaveSession(ctx, st.metrics)
}

// GetStartTime returns the session start time
func (st *SessionTracker) GetStartTime() time.Time {
	return st.startTime
}

// GetRealTimeMetrics returns current real-time metrics for dashboard
func (s *Service) GetRealTimeMetrics(ctx context.Context) (*RealTimeMetrics, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	metrics := s.metrics.GetSnapshot()

	return &RealTimeMetrics{
		CurrentCO2Rate:         metrics.CurrentRate,
		OptimizationActive:     metrics.OptimizationsActive > 0,
		InstantSavingsRate:     metrics.SavingsRate,
		CumulativeSavingsToday: metrics.TodaySavings / 1000.0, // Convert to kg
		ActiveUsers:            metrics.ActiveSessions,
		LastUpdated:            time.Now(),
	}, nil
}

// MetricsCollector collects real-time metrics
type MetricsCollector struct {
	mu                  sync.RWMutex
	currentRate         float64
	savingsRate         float64
	todaySavings        float64
	activeSessions      int
	optimizationsActive int
	lastReset           time.Time
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		lastReset: time.Now(),
	}
}

// RecordCalculation records a calculation
func (mc *MetricsCollector) RecordCalculation(impactType ImpactType, savings float64) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.todaySavings += savings
	mc.savingsRate = savings * 3600 // Convert to hourly rate
	
	// Reset daily metrics at midnight
	if time.Since(mc.lastReset) > 24*time.Hour {
		mc.todaySavings = 0
		mc.lastReset = time.Now()
	}
}

// GetSnapshot returns current metrics snapshot
func (mc *MetricsCollector) GetSnapshot() struct {
	CurrentRate         float64
	SavingsRate         float64
	TodaySavings        float64
	ActiveSessions      int
	OptimizationsActive int
} {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	return struct {
		CurrentRate         float64
		SavingsRate         float64
		TodaySavings        float64
		ActiveSessions      int
		OptimizationsActive int
	}{
		CurrentRate:         mc.currentRate,
		SavingsRate:         mc.savingsRate,
		TodaySavings:        mc.todaySavings,
		ActiveSessions:      mc.activeSessions,
		OptimizationsActive: mc.optimizationsActive,
	}
}

// generateID generates a unique ID
func generateID() string {
	return fmt.Sprintf("imp_%d_%d", time.Now().Unix(), rand.Intn(10000))
}