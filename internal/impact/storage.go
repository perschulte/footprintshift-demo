package impact

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// MemoryStorage provides an in-memory implementation of the Storage interface
// This is suitable for demonstration and testing purposes
type MemoryStorage struct {
	baselines map[string]*BaselineMeasurement
	sessions  map[string]*SessionMetrics
	reports   map[string]*ImpactReport
	mu        sync.RWMutex
}

// NewMemoryStorage creates a new in-memory storage instance
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		baselines: make(map[string]*BaselineMeasurement),
		sessions:  make(map[string]*SessionMetrics),
		reports:   make(map[string]*ImpactReport),
	}
}

// SaveBaseline saves a baseline measurement
func (ms *MemoryStorage) SaveBaseline(ctx context.Context, baseline *BaselineMeasurement) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	ms.baselines[baseline.ID] = baseline
	return nil
}

// GetBaseline retrieves a baseline measurement by ID
func (ms *MemoryStorage) GetBaseline(ctx context.Context, id string) (*BaselineMeasurement, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	baseline, exists := ms.baselines[id]
	if !exists {
		return nil, fmt.Errorf("baseline with ID %s not found", id)
	}

	return baseline, nil
}

// SaveSession saves session metrics
func (ms *MemoryStorage) SaveSession(ctx context.Context, session *SessionMetrics) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	ms.sessions[session.SessionID] = session
	return nil
}

// GetSession retrieves session metrics by session ID
func (ms *MemoryStorage) GetSession(ctx context.Context, sessionID string) (*SessionMetrics, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	session, exists := ms.sessions[sessionID]
	if !exists {
		return nil, fmt.Errorf("session with ID %s not found", sessionID)
	}

	return session, nil
}

// SaveReport saves an impact report
func (ms *MemoryStorage) SaveReport(ctx context.Context, report *ImpactReport) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	ms.reports[report.ID] = report
	return nil
}

// GetReport retrieves an impact report by ID
func (ms *MemoryStorage) GetReport(ctx context.Context, id string) (*ImpactReport, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	report, exists := ms.reports[id]
	if !exists {
		return nil, fmt.Errorf("report with ID %s not found", id)
	}

	return report, nil
}

// GetAggregatedMetrics calculates aggregated metrics for a period
func (ms *MemoryStorage) GetAggregatedMetrics(ctx context.Context, period ReportPeriod) (*AggregatedMetrics, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	metrics := &AggregatedMetrics{
		Period: period,
	}

	// Aggregate baseline data
	for _, baseline := range ms.baselines {
		if baseline.MeasuredAt.After(period.Start) && baseline.MeasuredAt.Before(period.End) {
			// Estimate total emissions for the period
			hoursInPeriod := period.End.Sub(period.Start).Hours()
			estimatedTotal := baseline.EstimatedHourlyEmissions * hoursInPeriod
			metrics.TotalBaseline += estimatedTotal
		}
	}

	// Aggregate session data
	for _, session := range ms.sessions {
		if session.StartTime.After(period.Start) && session.StartTime.Before(period.End) {
			metrics.SessionCount++
			metrics.TotalOptimized += session.TotalEmissions
			metrics.TotalSavings += session.EstimatedSavings
			metrics.OptimizationCount += session.OptimizationsApplied
		}
	}

	// If we have no baseline data, estimate based on typical patterns
	if metrics.TotalBaseline == 0 && metrics.SessionCount > 0 {
		// Conservative estimate: optimized emissions represent 70% of baseline
		metrics.TotalBaseline = metrics.TotalOptimized / 0.7
	}

	return metrics, nil
}

// PostgreSQLStorage provides a PostgreSQL implementation (skeleton for future use)
type PostgreSQLStorage struct {
	// db *sql.DB
	// In a real implementation, this would contain database connection
}

// NewPostgreSQLStorage creates a new PostgreSQL storage instance
func NewPostgreSQLStorage(connectionString string) (*PostgreSQLStorage, error) {
	// In a real implementation, you would:
	// 1. Connect to PostgreSQL
	// 2. Run migrations to create tables
	// 3. Return configured storage
	return &PostgreSQLStorage{}, fmt.Errorf("PostgreSQL storage not implemented in this demo")
}

// Example table schemas for PostgreSQL implementation:
const (
	createBaselinesTableSQL = `
	CREATE TABLE IF NOT EXISTS baselines (
		id VARCHAR(255) PRIMARY KEY,
		url TEXT NOT NULL,
		page_load_emissions DECIMAL(10,3) NOT NULL,
		data_transferred DECIMAL(10,3) NOT NULL,
		javascript_size DECIMAL(10,3) NOT NULL,
		image_size DECIMAL(10,3) NOT NULL,
		video_size DECIMAL(10,3) NOT NULL,
		load_time DECIMAL(10,3) NOT NULL,
		resource_count JSONB,
		estimated_hourly_emissions DECIMAL(10,3) NOT NULL,
		measured_at TIMESTAMP NOT NULL,
		device_type VARCHAR(50) NOT NULL,
		connection_type VARCHAR(50) NOT NULL,
		region VARCHAR(50) NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`

	createSessionsTableSQL = `
	CREATE TABLE IF NOT EXISTS sessions (
		session_id VARCHAR(255) PRIMARY KEY,
		start_time TIMESTAMP NOT NULL,
		duration DECIMAL(10,3) NOT NULL,
		total_emissions DECIMAL(10,3) NOT NULL,
		emissions_by_activity JSONB,
		optimizations_applied INTEGER NOT NULL DEFAULT 0,
		estimated_savings DECIMAL(10,3) NOT NULL DEFAULT 0,
		device_type VARCHAR(50) NOT NULL,
		region VARCHAR(50) NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`

	createReportsTableSQL = `
	CREATE TABLE IF NOT EXISTS reports (
		id VARCHAR(255) PRIMARY KEY,
		period_start TIMESTAMP NOT NULL,
		period_end TIMESTAMP NOT NULL,
		period_days INTEGER NOT NULL,
		total_savings DECIMAL(10,3) NOT NULL,
		baseline_total DECIMAL(10,3) NOT NULL,
		optimized_total DECIMAL(10,3) NOT NULL,
		savings_by_type JSONB,
		optimization_events INTEGER NOT NULL,
		average_optimization_level DECIMAL(5,2) NOT NULL,
		equivalent_to JSONB,
		confidence_score DECIMAL(5,2) NOT NULL,
		methodology TEXT,
		recommendations JSONB,
		generated_at TIMESTAMP NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`
)

// MockStorageWithData provides pre-populated mock data for testing
func NewMockStorageWithData() *MemoryStorage {
	storage := NewMemoryStorage()

	// Add sample baseline data
	baseline1 := &BaselineMeasurement{
		ID:                       "baseline_001",
		URL:                      "https://example.com",
		PageLoadEmissions:        245.7,
		DataTransferred:          2.1,
		JavaScriptSize:           0.6,
		ImageSize:                1.2,
		VideoSize:                0.0,
		LoadTime:                 3.2,
		ResourceCount:            map[string]int{"images": 15, "scripts": 8, "stylesheets": 3},
		EstimatedHourlyEmissions: 89.2,
		MeasuredAt:               time.Now().Add(-24 * time.Hour),
		DeviceType:               "laptop",
		ConnectionType:           "wifi",
		Region:                   "EU",
	}

	baseline2 := &BaselineMeasurement{
		ID:                       "baseline_002",
		URL:                      "https://shop.example.com",
		PageLoadEmissions:        567.3,
		DataTransferred:          4.8,
		JavaScriptSize:           1.2,
		ImageSize:                2.9,
		VideoSize:                0.5,
		LoadTime:                 5.7,
		ResourceCount:            map[string]int{"images": 32, "scripts": 12, "stylesheets": 5},
		EstimatedHourlyEmissions: 156.4,
		MeasuredAt:               time.Now().Add(-12 * time.Hour),
		DeviceType:               "smartphone",
		ConnectionType:           "mobile_4g",
		Region:                   "US",
	}

	// Add sample session data
	session1 := &SessionMetrics{
		SessionID:   "session_001",
		StartTime:   time.Now().Add(-6 * time.Hour),
		Duration:    1800, // 30 minutes
		TotalEmissions: 156.8,
		EmissionsByActivity: map[string]float64{
			"page_load":       67.2,
			"video_streaming": 45.3,
			"image_loading":   32.1,
			"javascript":      12.2,
		},
		OptimizationsApplied: 3,
		EstimatedSavings:     78.4,
		DeviceType:          "laptop",
		Region:              "EU",
	}

	session2 := &SessionMetrics{
		SessionID:   "session_002",
		StartTime:   time.Now().Add(-2 * time.Hour),
		Duration:    900, // 15 minutes
		TotalEmissions: 89.3,
		EmissionsByActivity: map[string]float64{
			"page_load":    34.5,
			"image_loading": 28.7,
			"javascript":   26.1,
		},
		OptimizationsApplied: 2,
		EstimatedSavings:     45.2,
		DeviceType:          "smartphone",
		Region:              "US",
	}

	// Save the mock data
	storage.SaveBaseline(context.Background(), baseline1)
	storage.SaveBaseline(context.Background(), baseline2)
	storage.SaveSession(context.Background(), session1)
	storage.SaveSession(context.Background(), session2)

	return storage
}