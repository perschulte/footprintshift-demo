// Package carbon provides interfaces for carbon intensity monitoring services.
package carbon

import (
	"context"
	"time"
)

// CarbonService defines the interface for carbon intensity monitoring services.
//
// This interface provides methods for retrieving real-time carbon intensity data
// and generating forecasts for optimal energy consumption timing. Implementations
// may use various data sources such as Electricity Maps, WattTime, or other
// grid monitoring services.
//
// The interface is designed to be stable and backward-compatible for external
// SDK consumers, abstracting away the underlying data source implementation.
type CarbonService interface {
	// GetCarbonIntensity retrieves the current carbon intensity for a specific location.
	//
	// Parameters:
	//   - ctx: Context for request timeout and cancellation
	//   - location: Location identifier (city name, country code, or grid zone)
	//
	// Returns the current carbon intensity data or an error if the request fails.
	// Implementations should provide fallback data when the primary data source
	// is unavailable.
	GetCarbonIntensity(ctx context.Context, location string) (*CarbonIntensity, error)

	// GetGreenHoursForecast generates a forecast of optimal low-carbon time windows.
	//
	// Parameters:
	//   - ctx: Context for request timeout and cancellation
	//   - location: Location identifier (city name, country code, or grid zone)
	//   - hours: Number of hours to forecast (typically 24-168, max 1 week)
	//
	// Returns a forecast containing predicted green hours or an error if the
	// request fails. The forecast includes the best time windows for scheduling
	// energy-intensive operations.
	GetGreenHoursForecast(ctx context.Context, location string, hours int) (*GreenHoursForecast, error)

	// IsHealthy checks if the carbon intensity service is operational.
	//
	// Parameters:
	//   - ctx: Context for request timeout and cancellation
	//
	// Returns true if the service can successfully retrieve carbon intensity data,
	// false otherwise. This is useful for health checks and monitoring.
	IsHealthy(ctx context.Context) bool

	// GetSupportedLocations returns a list of locations supported by this service.
	//
	// Parameters:
	//   - ctx: Context for request timeout and cancellation
	//
	// Returns a list of location identifiers that can be used with other methods.
	// This is optional and may return an empty slice if the service supports
	// arbitrary location queries.
	GetSupportedLocations(ctx context.Context) ([]string, error)
}

// CarbonServiceWithCache extends CarbonService with caching capabilities.
//
// This interface adds methods for managing cached carbon intensity data,
// which can improve performance and reduce API calls to external services.
type CarbonServiceWithCache interface {
	CarbonService

	// GetCachedCarbonIntensity retrieves carbon intensity from cache if available.
	//
	// Parameters:
	//   - ctx: Context for request timeout and cancellation
	//   - location: Location identifier
	//   - maxAge: Maximum age of cached data to accept
	//
	// Returns cached carbon intensity data or nil if not available or expired.
	GetCachedCarbonIntensity(ctx context.Context, location string, maxAge time.Duration) (*CarbonIntensity, error)

	// ClearCache removes cached data for a specific location or all locations.
	//
	// Parameters:
	//   - ctx: Context for request timeout and cancellation
	//   - location: Location identifier, or empty string to clear all cache
	//
	// Returns an error if the cache operation fails.
	ClearCache(ctx context.Context, location string) error
}

// CarbonServiceWithHistory extends CarbonService with historical data capabilities.
//
// This interface adds methods for retrieving historical carbon intensity data,
// which can be useful for analysis and trend identification.
type CarbonServiceWithHistory interface {
	CarbonService

	// GetHistoricalCarbonIntensity retrieves historical carbon intensity data.
	//
	// Parameters:
	//   - ctx: Context for request timeout and cancellation
	//   - location: Location identifier
	//   - start: Start time for historical data
	//   - end: End time for historical data
	//
	// Returns a slice of historical carbon intensity readings or an error.
	GetHistoricalCarbonIntensity(ctx context.Context, location string, start, end time.Time) ([]CarbonIntensity, error)

	// GetAverageCarbonIntensity calculates average carbon intensity over a period.
	//
	// Parameters:
	//   - ctx: Context for request timeout and cancellation
	//   - location: Location identifier
	//   - start: Start time for averaging period
	//   - end: End time for averaging period
	//
	// Returns the average carbon intensity or an error.
	GetAverageCarbonIntensity(ctx context.Context, location string, start, end time.Time) (float64, error)
}

// CarbonWebhookService defines the interface for carbon intensity webhook notifications.
//
// This interface allows services to register for notifications when carbon
// intensity changes significantly or when green hours begin/end.
type CarbonWebhookService interface {
	// RegisterWebhook registers a webhook URL for carbon intensity notifications.
	//
	// Parameters:
	//   - ctx: Context for request timeout and cancellation
	//   - webhookURL: URL to receive webhook notifications
	//   - location: Location to monitor
	//   - events: Types of events to notify about (e.g., "green_hour_start", "intensity_change")
	//
	// Returns a webhook ID for managing the registration or an error.
	RegisterWebhook(ctx context.Context, webhookURL, location string, events []string) (string, error)

	// UnregisterWebhook removes a webhook registration.
	//
	// Parameters:
	//   - ctx: Context for request timeout and cancellation
	//   - webhookID: ID of the webhook to remove
	//
	// Returns an error if the operation fails.
	UnregisterWebhook(ctx context.Context, webhookID string) error

	// ListWebhooks returns all registered webhooks for a location.
	//
	// Parameters:
	//   - ctx: Context for request timeout and cancellation
	//   - location: Location to query webhooks for
	//
	// Returns a list of webhook registrations or an error.
	ListWebhooks(ctx context.Context, location string) ([]WebhookRegistration, error)
}

// WebhookRegistration represents a registered webhook for carbon intensity notifications.
type WebhookRegistration struct {
	// ID is the unique identifier for this webhook registration
	ID string `json:"id"`

	// URL is the webhook endpoint URL
	URL string `json:"url"`

	// Location is the location being monitored
	Location string `json:"location"`

	// Events is the list of event types that trigger notifications
	Events []string `json:"events"`

	// CreatedAt is when the webhook was registered
	CreatedAt time.Time `json:"created_at"`

	// LastTriggered is when the webhook was last triggered (optional)
	LastTriggered *time.Time `json:"last_triggered,omitempty"`

	// Active indicates if the webhook is currently active
	Active bool `json:"active"`
}

// CarbonServiceConfig represents configuration options for carbon intensity services.
type CarbonServiceConfig struct {
	// APIKey is the authentication key for the carbon intensity data provider
	APIKey string `json:"api_key,omitempty"`

	// BaseURL is the base URL for the carbon intensity API
	BaseURL string `json:"base_url,omitempty"`

	// Timeout is the request timeout for API calls
	Timeout time.Duration `json:"timeout,omitempty"`

	// DefaultLocation is the fallback location when none is specified
	DefaultLocation string `json:"default_location,omitempty"`

	// CacheEnabled indicates if caching should be enabled
	CacheEnabled bool `json:"cache_enabled,omitempty"`

	// CacheTTL is the time-to-live for cached data
	CacheTTL time.Duration `json:"cache_ttl,omitempty"`

	// FallbackEnabled indicates if fallback/mock data should be used when API is unavailable
	FallbackEnabled bool `json:"fallback_enabled,omitempty"`

	// Thresholds defines custom carbon intensity thresholds
	Thresholds *CarbonIntensityThresholds `json:"thresholds,omitempty"`
}

// DefaultCarbonServiceConfig provides sensible defaults for carbon service configuration.
var DefaultCarbonServiceConfig = CarbonServiceConfig{
	Timeout:         30 * time.Second,
	DefaultLocation: "DE",
	CacheEnabled:    true,
	CacheTTL:        15 * time.Minute,
	FallbackEnabled: true,
	Thresholds:      &DefaultThresholds,
}