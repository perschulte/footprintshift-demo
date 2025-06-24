// Package optimization provides interfaces for website optimization services.
package optimization

import (
	"context"
	"time"
)

// OptimizationService defines the interface for website optimization services.
//
// This interface provides methods for generating optimization profiles and
// recommendations based on carbon intensity data. The service helps websites
// adapt their behavior to minimize energy consumption during high-carbon periods
// and maximize user experience during low-carbon periods.
type OptimizationService interface {
	// GetOptimizationProfile generates optimization recommendations based on carbon intensity.
	//
	// Parameters:
	//   - ctx: Context for request timeout and cancellation
	//   - request: Optimization request containing location, URL, and preferences
	//
	// Returns an optimization response with profile and recommendations, or an error.
	GetOptimizationProfile(ctx context.Context, request OptimizationRequest) (*OptimizationResponse, error)

	// GetOptimizationRecommendations provides detailed human-readable recommendations.
	//
	// Parameters:
	//   - ctx: Context for request timeout and cancellation
	//   - request: Optimization request containing location, URL, and preferences
	//
	// Returns a list of optimization recommendations or an error.
	GetOptimizationRecommendations(ctx context.Context, request OptimizationRequest) ([]string, error)

	// ValidateOptimizationProfile checks if an optimization profile is valid and safe to apply.
	//
	// Parameters:
	//   - ctx: Context for request timeout and cancellation
	//   - profile: Optimization profile to validate
	//
	// Returns true if the profile is valid, false otherwise.
	ValidateOptimizationProfile(ctx context.Context, profile *OptimizationProfile) (bool, error)

	// GetSupportedFeatures returns a list of features that can be optimized.
	//
	// Parameters:
	//   - ctx: Context for request timeout and cancellation
	//
	// Returns a list of feature names that can be disabled or optimized.
	GetSupportedFeatures(ctx context.Context) ([]string, error)
}

// OptimizationServiceWithRules extends OptimizationService with rule-based optimization.
//
// This interface adds support for custom optimization rules that can be
// configured and managed dynamically.
type OptimizationServiceWithRules interface {
	OptimizationService

	// CreateRule creates a new optimization rule.
	//
	// Parameters:
	//   - ctx: Context for request timeout and cancellation
	//   - rule: Rule definition to create
	//
	// Returns the created rule with assigned ID or an error.
	CreateRule(ctx context.Context, rule *OptimizationRule) (*OptimizationRule, error)

	// UpdateRule updates an existing optimization rule.
	//
	// Parameters:
	//   - ctx: Context for request timeout and cancellation
	//   - ruleID: ID of the rule to update
	//   - rule: Updated rule definition
	//
	// Returns the updated rule or an error.
	UpdateRule(ctx context.Context, ruleID string, rule *OptimizationRule) (*OptimizationRule, error)

	// DeleteRule removes an optimization rule.
	//
	// Parameters:
	//   - ctx: Context for request timeout and cancellation
	//   - ruleID: ID of the rule to delete
	//
	// Returns an error if the deletion fails.
	DeleteRule(ctx context.Context, ruleID string) error

	// GetRule retrieves a specific optimization rule.
	//
	// Parameters:
	//   - ctx: Context for request timeout and cancellation
	//   - ruleID: ID of the rule to retrieve
	//
	// Returns the rule or an error if not found.
	GetRule(ctx context.Context, ruleID string) (*OptimizationRule, error)

	// ListRules returns all optimization rules, optionally filtered by tags.
	//
	// Parameters:
	//   - ctx: Context for request timeout and cancellation
	//   - tags: Optional list of tags to filter by
	//
	// Returns a list of rules or an error.
	ListRules(ctx context.Context, tags []string) ([]*OptimizationRule, error)

	// EvaluateRules evaluates all rules against the given context.
	//
	// Parameters:
	//   - ctx: Context for request timeout and cancellation
	//   - evalContext: Context for rule evaluation
	//
	// Returns a list of rules that match the context or an error.
	EvaluateRules(ctx context.Context, evalContext *OptimizationContext) ([]*OptimizationRule, error)
}

// OptimizationServiceWithAnalytics extends OptimizationService with analytics capabilities.
//
// This interface adds methods for tracking optimization effectiveness and
// gathering insights about energy savings and performance impact.
type OptimizationServiceWithAnalytics interface {
	OptimizationService

	// TrackOptimizationApplication records when an optimization profile is applied.
	//
	// Parameters:
	//   - ctx: Context for request timeout and cancellation
	//   - profile: Optimization profile that was applied
	//   - metadata: Additional metadata about the application
	//
	// Returns an error if tracking fails.
	TrackOptimizationApplication(ctx context.Context, profile *OptimizationProfile, metadata map[string]interface{}) error

	// GetOptimizationStats returns statistics about optimization effectiveness.
	//
	// Parameters:
	//   - ctx: Context for request timeout and cancellation
	//   - location: Location to get stats for (optional)
	//   - timeRange: Time range for statistics
	//
	// Returns optimization statistics or an error.
	GetOptimizationStats(ctx context.Context, location string, timeRange TimeRange) (*OptimizationStats, error)

	// GetEnergySavingsReport generates a report on estimated energy savings.
	//
	// Parameters:
	//   - ctx: Context for request timeout and cancellation
	//   - location: Location to generate report for (optional)
	//   - timeRange: Time range for the report
	//
	// Returns an energy savings report or an error.
	GetEnergySavingsReport(ctx context.Context, location string, timeRange TimeRange) (*EnergySavingsReport, error)
}

// OptimizationServiceWithCache extends OptimizationService with caching capabilities.
//
// This interface adds methods for managing cached optimization profiles
// to improve performance and reduce computation overhead.
type OptimizationServiceWithCache interface {
	OptimizationService

	// GetCachedOptimizationProfile retrieves an optimization profile from cache.
	//
	// Parameters:
	//   - ctx: Context for request timeout and cancellation
	//   - cacheKey: Key to identify the cached profile
	//   - maxAge: Maximum age of cached data to accept
	//
	// Returns cached optimization profile or nil if not available/expired.
	GetCachedOptimizationProfile(ctx context.Context, cacheKey string, maxAge time.Duration) (*OptimizationProfile, error)

	// CacheOptimizationProfile stores an optimization profile in cache.
	//
	// Parameters:
	//   - ctx: Context for request timeout and cancellation
	//   - cacheKey: Key to store the profile under
	//   - profile: Optimization profile to cache
	//   - ttl: Time-to-live for the cached data
	//
	// Returns an error if caching fails.
	CacheOptimizationProfile(ctx context.Context, cacheKey string, profile *OptimizationProfile, ttl time.Duration) error

	// ClearOptimizationCache removes cached optimization profiles.
	//
	// Parameters:
	//   - ctx: Context for request timeout and cancellation
	//   - pattern: Pattern to match cache keys (empty = clear all)
	//
	// Returns an error if cache clearing fails.
	ClearOptimizationCache(ctx context.Context, pattern string) error
}

// TimeRange represents a time range for analytics queries.
type TimeRange struct {
	// Start is the beginning of the time range
	Start time.Time `json:"start"`

	// End is the end of the time range
	End time.Time `json:"end"`
}

// OptimizationStats contains statistics about optimization effectiveness.
type OptimizationStats struct {
	// TotalOptimizations is the total number of optimizations applied
	TotalOptimizations int64 `json:"total_optimizations"`

	// OptimizationsByMode breaks down optimizations by mode
	OptimizationsByMode map[OptimizationMode]int64 `json:"optimizations_by_mode"`

	// MostDisabledFeatures lists the most frequently disabled features
	MostDisabledFeatures []FeatureDisableCount `json:"most_disabled_features"`

	// AverageEnergySavings is the average estimated energy savings percentage
	AverageEnergySavings float64 `json:"average_energy_savings"`

	// AveragePerformanceImpact is the average performance impact percentage
	AveragePerformanceImpact float64 `json:"average_performance_impact"`

	// OptimizationsByLocation breaks down optimizations by location
	OptimizationsByLocation map[string]int64 `json:"optimizations_by_location"`

	// TimeRange is the time range covered by these statistics
	TimeRange TimeRange `json:"time_range"`

	// GeneratedAt indicates when these statistics were generated
	GeneratedAt time.Time `json:"generated_at"`
}

// FeatureDisableCount represents how often a feature has been disabled.
type FeatureDisableCount struct {
	Feature string `json:"feature"`
	Count   int64  `json:"count"`
}

// EnergySavingsReport contains detailed information about energy savings.
type EnergySavingsReport struct {
	// TotalEnergySaved is the total estimated energy saved in kWh
	TotalEnergySaved float64 `json:"total_energy_saved"`

	// TotalCarbonSaved is the total estimated carbon emissions avoided in kg CO2
	TotalCarbonSaved float64 `json:"total_carbon_saved"`

	// SavingsByMode breaks down savings by optimization mode
	SavingsByMode map[OptimizationMode]EnergySavingsData `json:"savings_by_mode"`

	// SavingsByFeature breaks down savings by disabled features
	SavingsByFeature map[string]EnergySavingsData `json:"savings_by_feature"`

	// DailySavings shows daily energy savings over the time period
	DailySavings []DailySavings `json:"daily_savings"`

	// TimeRange is the time range covered by this report
	TimeRange TimeRange `json:"time_range"`

	// GeneratedAt indicates when this report was generated
	GeneratedAt time.Time `json:"generated_at"`

	// Methodology describes how the savings were calculated
	Methodology string `json:"methodology"`
}

// EnergySavingsData contains energy and carbon savings data.
type EnergySavingsData struct {
	// EnergySaved is the energy saved in kWh
	EnergySaved float64 `json:"energy_saved"`

	// CarbonSaved is the carbon emissions avoided in kg CO2
	CarbonSaved float64 `json:"carbon_saved"`

	// Count is the number of times this optimization was applied
	Count int64 `json:"count"`
}

// DailySavings represents energy savings for a single day.
type DailySavings struct {
	// Date is the date for these savings
	Date time.Time `json:"date"`

	// EnergySaved is the energy saved on this date in kWh
	EnergySaved float64 `json:"energy_saved"`

	// CarbonSaved is the carbon emissions avoided on this date in kg CO2
	CarbonSaved float64 `json:"carbon_saved"`

	// OptimizationsApplied is the number of optimizations applied on this date
	OptimizationsApplied int64 `json:"optimizations_applied"`
}

// OptimizationServiceConfig represents configuration options for optimization services.
type OptimizationServiceConfig struct {
	// DefaultMode is the default optimization mode when none is specified
	DefaultMode OptimizationMode `json:"default_mode"`

	// CustomRulesEnabled indicates if custom rules are supported
	CustomRulesEnabled bool `json:"custom_rules_enabled"`

	// AnalyticsEnabled indicates if analytics tracking is enabled
	AnalyticsEnabled bool `json:"analytics_enabled"`

	// CacheEnabled indicates if optimization profile caching is enabled
	CacheEnabled bool `json:"cache_enabled"`

	// CacheTTL is the default time-to-live for cached optimization profiles
	CacheTTL time.Duration `json:"cache_ttl"`

	// MaxCacheSize is the maximum number of profiles to cache
	MaxCacheSize int `json:"max_cache_size"`

	// RuleEvaluationTimeout is the timeout for rule evaluation
	RuleEvaluationTimeout time.Duration `json:"rule_evaluation_timeout"`

	// SupportedFeatures is the list of features that can be optimized
	SupportedFeatures []string `json:"supported_features"`

	// FeatureCategories groups features by category for better organization
	FeatureCategories map[string][]string `json:"feature_categories"`

	// DefaultThresholds defines default carbon intensity thresholds for optimization
	DefaultThresholds map[OptimizationMode]float64 `json:"default_thresholds"`
}

// DefaultOptimizationServiceConfig provides sensible defaults for optimization service configuration.
var DefaultOptimizationServiceConfig = OptimizationServiceConfig{
	DefaultMode:           ModeNormal,
	CustomRulesEnabled:    true,
	AnalyticsEnabled:      true,
	CacheEnabled:          true,
	CacheTTL:              15 * time.Minute,
	MaxCacheSize:          1000,
	RuleEvaluationTimeout: 5 * time.Second,
	SupportedFeatures: []string{
		"video_autoplay", "3d_models", "animations", "ai_features", "live_chat",
		"background_videos", "carousels", "parallax_effects", "product_zoom",
		"360_view", "recommendation_widgets", "auto_thumbnails", "preview_videos",
		"infinite_scroll", "story_previews", "auto_refresh", "breaking_news_animations",
		"comment_sections", "custom_fonts", "web_fonts", "heavy_scripts",
		"tracking_scripts", "social_widgets", "embedded_maps", "video_backgrounds",
	},
	FeatureCategories: map[string][]string{
		"video": {"video_autoplay", "background_videos", "preview_videos", "video_backgrounds"},
		"interactive": {"3d_models", "product_zoom", "360_view", "parallax_effects"},
		"ai": {"ai_features", "recommendation_widgets"},
		"social": {"live_chat", "social_widgets", "comment_sections"},
		"animations": {"animations", "carousels", "heavy_animations"},
		"fonts": {"custom_fonts", "web_fonts"},
		"scripts": {"heavy_scripts", "tracking_scripts"},
		"auto_features": {"auto_thumbnails", "auto_refresh", "infinite_scroll"},
	},
	DefaultThresholds: map[OptimizationMode]float64{
		ModeFull:     0,
		ModeNormal:   150,
		ModeEco:      300,
		ModeCritical: 500,
	},
}