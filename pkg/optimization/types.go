// Package optimization provides types and interfaces for website optimization based on carbon intensity.
//
// This package contains data structures for representing optimization profiles,
// rules, and recommendations that help websites adapt their behavior based on
// current carbon intensity levels. The goal is to reduce energy consumption
// during high-carbon periods and optimize user experience during low-carbon periods.
package optimization

import (
	"fmt"
	"strings"
	"time"

	"github.com/perschulte/greenweb-api/pkg/carbon"
)

// OptimizationMode represents the different optimization levels available.
type OptimizationMode string

const (
	// ModeFull enables all website features with no restrictions (used during green hours)
	ModeFull OptimizationMode = "full"

	// ModeNormal applies moderate optimizations to reduce energy consumption
	ModeNormal OptimizationMode = "normal"

	// ModeEco applies aggressive optimizations to minimize energy usage
	ModeEco OptimizationMode = "eco"

	// ModeCritical applies maximum optimizations for extremely high carbon intensity
	ModeCritical OptimizationMode = "critical"
)

// String returns the string representation of the optimization mode.
func (m OptimizationMode) String() string {
	return string(m)
}

// IsValid returns true if the optimization mode is valid.
func (m OptimizationMode) IsValid() bool {
	switch m {
	case ModeFull, ModeNormal, ModeEco, ModeCritical:
		return true
	default:
		return false
	}
}

// ImageQuality represents different image quality levels for optimization.
type ImageQuality string

const (
	// ImageQualityHigh provides maximum image quality with no compression
	ImageQualityHigh ImageQuality = "high"

	// ImageQualityMedium provides good image quality with moderate compression
	ImageQualityMedium ImageQuality = "medium"

	// ImageQualityLow provides basic image quality with aggressive compression
	ImageQualityLow ImageQuality = "low"
)

// String returns the string representation of the image quality.
func (iq ImageQuality) String() string {
	return string(iq)
}

// VideoQuality represents different video quality levels for optimization.
type VideoQuality string

const (
	// VideoQuality4K provides ultra-high definition video (3840x2160)
	VideoQuality4K VideoQuality = "4k"

	// VideoQuality1080p provides full high definition video (1920x1080)
	VideoQuality1080p VideoQuality = "1080p"

	// VideoQuality720p provides high definition video (1280x720)
	VideoQuality720p VideoQuality = "720p"

	// VideoQuality480p provides standard definition video (854x480)
	VideoQuality480p VideoQuality = "480p"

	// VideoQuality360p provides low definition video (640x360)
	VideoQuality360p VideoQuality = "360p"
)

// String returns the string representation of the video quality.
func (vq VideoQuality) String() string {
	return string(vq)
}

// CachingStrategy represents different caching strategies for optimization.
type CachingStrategy string

const (
	// CachingMinimal uses minimal caching to ensure fresh content
	CachingMinimal CachingStrategy = "minimal"

	// CachingNormal uses standard caching policies
	CachingNormal CachingStrategy = "normal"

	// CachingAggressive uses aggressive caching to minimize server requests
	CachingAggressive CachingStrategy = "aggressive"
)

// String returns the string representation of the caching strategy.
func (cs CachingStrategy) String() string {
	return string(cs)
}

// OptimizationProfile defines how a website should adapt based on carbon intensity.
//
// This profile contains recommendations for adjusting website behavior to minimize
// energy consumption during high-carbon periods and maximize user experience
// during low-carbon periods.
type OptimizationProfile struct {
	// Mode indicates the overall optimization level
	Mode OptimizationMode `json:"mode" validate:"required,oneof=full normal eco critical" example:"normal"`

	// DisableFeatures lists website features that should be disabled
	DisableFeatures []string `json:"disable_features" validate:"dive,required" example:"video_autoplay,3d_models"`

	// ImageQuality specifies the recommended image quality level
	ImageQuality ImageQuality `json:"image_quality" validate:"required,oneof=high medium low" example:"medium"`

	// VideoQuality specifies the recommended video quality level
	VideoQuality VideoQuality `json:"video_quality" validate:"required,oneof=4k 1080p 720p 480p 360p" example:"720p"`

	// DeferAnalytics indicates whether to defer non-essential analytics and tracking
	DeferAnalytics bool `json:"defer_analytics" example:"false"`

	// EcoDiscount is a percentage discount to offer during eco-friendly periods (0-100)
	EcoDiscount int `json:"eco_discount" validate:"min=0,max=100" example:"5"`

	// ShowGreenBanner indicates whether to display a green energy banner
	ShowGreenBanner bool `json:"show_green_banner" example:"true"`

	// CachingStrategy specifies the recommended caching approach
	CachingStrategy CachingStrategy `json:"caching_strategy" validate:"required,oneof=minimal normal aggressive" example:"normal"`

	// ResourceLimits defines limits on resource usage
	ResourceLimits ResourceLimits `json:"resource_limits,omitempty"`

	// UIOptimizations contains user interface optimization settings
	UIOptimizations UIOptimizations `json:"ui_optimizations,omitempty"`

	// ContentOptimizations contains content optimization settings
	ContentOptimizations ContentOptimizations `json:"content_optimizations,omitempty"`

	// HighImpactOptimizations contains optimizations with significant CO2 impact
	HighImpactOptimizations HighImpactOptimizations `json:"high_impact_optimizations,omitempty"`

	// FeatureImpactScores contains CO2 impact scores for each optimization
	FeatureImpactScores map[string]FeatureImpact `json:"feature_impact_scores,omitempty"`

	// EstimatedCO2Savings contains detailed CO2 savings calculations
	EstimatedCO2Savings CO2SavingsBreakdown `json:"estimated_co2_savings,omitempty"`

	// GeneratedAt indicates when this profile was generated
	GeneratedAt time.Time `json:"generated_at" validate:"required" example:"2024-01-15T14:30:00Z"`

	// ValidUntil indicates when this profile expires and should be refreshed
	ValidUntil time.Time `json:"valid_until" validate:"required" example:"2024-01-15T15:30:00Z"`

	// Metadata contains additional optimization metadata
	Metadata OptimizationMetadata `json:"metadata,omitempty"`
}

// ResourceLimits defines limits on resource usage during optimization.
type ResourceLimits struct {
	// MaxConcurrentRequests limits the number of simultaneous HTTP requests
	MaxConcurrentRequests int `json:"max_concurrent_requests,omitempty" validate:"min=1" example:"10"`

	// MaxImageSize limits the maximum size of images in bytes
	MaxImageSize int64 `json:"max_image_size,omitempty" validate:"min=0" example:"500000"`

	// MaxVideoSize limits the maximum size of videos in bytes
	MaxVideoSize int64 `json:"max_video_size,omitempty" validate:"min=0" example:"10000000"`

	// MaxScriptSize limits the maximum size of JavaScript files in bytes
	MaxScriptSize int64 `json:"max_script_size,omitempty" validate:"min=0" example:"100000"`

	// MaxStylesheetSize limits the maximum size of CSS files in bytes
	MaxStylesheetSize int64 `json:"max_stylesheet_size,omitempty" validate:"min=0" example:"50000"`

	// MaxFontSize limits the maximum size of font files in bytes
	MaxFontSize int64 `json:"max_font_size,omitempty" validate:"min=0" example:"200000"`
}

// UIOptimizations contains user interface optimization settings.
type UIOptimizations struct {
	// ReduceAnimations indicates whether to reduce or disable animations
	ReduceAnimations bool `json:"reduce_animations" example:"true"`

	// SimplifyLayouts indicates whether to use simplified page layouts
	SimplifyLayouts bool `json:"simplify_layouts" example:"false"`

	// DisableTransitions indicates whether to disable CSS transitions
	DisableTransitions bool `json:"disable_transitions" example:"true"`

	// ReduceColors indicates whether to use a reduced color palette
	ReduceColors bool `json:"reduce_colors" example:"false"`

	// DarkMode indicates whether to prefer dark mode to save energy on OLED displays
	DarkMode bool `json:"dark_mode" example:"false"`

	// MinimizeJavaScript indicates whether to minimize JavaScript execution
	MinimizeJavaScript bool `json:"minimize_javascript" example:"true"`

	// LazyLoadImages indicates whether to enable lazy loading for images
	LazyLoadImages bool `json:"lazy_load_images" example:"true"`

	// PreferSystemFonts indicates whether to use system fonts instead of web fonts
	PreferSystemFonts bool `json:"prefer_system_fonts" example:"true"`
}

// ContentOptimizations contains content optimization settings.
type ContentOptimizations struct {
	// CompressImages indicates whether to compress images
	CompressImages bool `json:"compress_images" example:"true"`

	// CompressVideos indicates whether to compress videos
	CompressVideos bool `json:"compress_videos" example:"true"`

	// MinifyCSS indicates whether to minify CSS files
	MinifyCSS bool `json:"minify_css" example:"true"`

	// MinifyJavaScript indicates whether to minify JavaScript files
	MinifyJavaScript bool `json:"minify_javascript" example:"true"`

	// MinifyHTML indicates whether to minify HTML content
	MinifyHTML bool `json:"minify_html" example:"true"`

	// EnableGzipCompression indicates whether to enable gzip compression
	EnableGzipCompression bool `json:"enable_gzip_compression" example:"true"`

	// EnableBrotliCompression indicates whether to enable Brotli compression
	EnableBrotliCompression bool `json:"enable_brotli_compression" example:"true"`

	// RemoveComments indicates whether to remove HTML/CSS/JS comments
	RemoveComments bool `json:"remove_comments" example:"true"`

	// OptimizeCriticalCSS indicates whether to optimize critical CSS
	OptimizeCriticalCSS bool `json:"optimize_critical_css" example:"true"`
}

// OptimizationMetadata contains additional metadata about the optimization profile.
type OptimizationMetadata struct {
	// CarbonIntensity is the carbon intensity that triggered this optimization
	CarbonIntensity float64 `json:"carbon_intensity" example:"250.5"`

	// Location is the location for which this optimization was generated
	Location string `json:"location" example:"Berlin"`

	// Source indicates the source of the optimization profile
	Source string `json:"source" example:"greenweb-api"`

	// Version is the version of the optimization algorithm used
	Version string `json:"version" example:"1.0.0"`

	// EstimatedEnergySavings is the estimated energy savings percentage
	EstimatedEnergySavings float64 `json:"estimated_energy_savings,omitempty" example:"25.5"`

	// PerformanceImpact is the estimated performance impact (negative = improvement)
	PerformanceImpact float64 `json:"performance_impact,omitempty" example:"-10.5"`
}

// IsExpired returns true if the optimization profile has expired.
func (p *OptimizationProfile) IsExpired() bool {
	return time.Now().After(p.ValidUntil)
}

// IsFeatureDisabled returns true if the specified feature is disabled in this profile.
func (p *OptimizationProfile) IsFeatureDisabled(feature string) bool {
	for _, disabled := range p.DisableFeatures {
		if disabled == feature {
			return true
		}
	}
	return false
}

// GetOptimizationLevel returns a numeric optimization level (0-100) where 100 is maximum optimization.
func (p *OptimizationProfile) GetOptimizationLevel() int {
	switch p.Mode {
	case ModeFull:
		return 0
	case ModeNormal:
		return 25
	case ModeEco:
		return 75
	case ModeCritical:
		return 100
	default:
		return 0
	}
}

// String provides a human-readable representation of the optimization profile.
func (p *OptimizationProfile) String() string {
	return fmt.Sprintf("Mode: %s, Features disabled: %d, Image quality: %s, Video quality: %s",
		p.Mode, len(p.DisableFeatures), p.ImageQuality, p.VideoQuality)
}

// OptimizationRequest represents a request for optimization recommendations.
type OptimizationRequest struct {
	// Location is the geographical location for carbon intensity lookup
	Location string `json:"location" validate:"required" example:"Berlin"`

	// URL is the website URL to optimize (optional, used for URL-specific optimizations)
	URL string `json:"url,omitempty" validate:"url" example:"https://example.com"`

	// UserAgent is the client's user agent string (optional, used for device-specific optimizations)
	UserAgent string `json:"user_agent,omitempty" example:"Mozilla/5.0..."`

	// DeviceType is the type of device (optional: desktop, mobile, tablet)
	DeviceType string `json:"device_type,omitempty" validate:"oneof=desktop mobile tablet" example:"desktop"`

	// ConnectionType is the connection type (optional: wifi, cellular, ethernet)
	ConnectionType string `json:"connection_type,omitempty" validate:"oneof=wifi cellular ethernet" example:"wifi"`

	// BandwidthLimit is the bandwidth limit in Mbps (optional)
	BandwidthLimit float64 `json:"bandwidth_limit,omitempty" validate:"min=0" example:"10.5"`

	// CustomThresholds allows overriding default carbon intensity thresholds
	CustomThresholds *carbon.CarbonIntensityThresholds `json:"custom_thresholds,omitempty"`

	// Features is a list of features available on the website
	Features []string `json:"features,omitempty" example:"video_autoplay,3d_models,ai_features"`

	// Preferences contains user preferences for optimization
	Preferences OptimizationPreferences `json:"preferences,omitempty"`
}

// OptimizationPreferences contains user preferences for optimization.
type OptimizationPreferences struct {
	// PreferPerformance indicates the user prefers performance over energy savings
	PreferPerformance bool `json:"prefer_performance" example:"false"`

	// PreferQuality indicates the user prefers quality over energy savings
	PreferQuality bool `json:"prefer_quality" example:"false"`

	// AllowReducedFunctionality indicates the user allows reduced functionality for energy savings
	AllowReducedFunctionality bool `json:"allow_reduced_functionality" example:"true"`

	// MaxPerformanceImpact is the maximum acceptable performance impact percentage
	MaxPerformanceImpact float64 `json:"max_performance_impact,omitempty" validate:"min=0,max=100" example:"20"`

	// DisallowedFeatures is a list of features that should never be disabled
	DisallowedFeatures []string `json:"disallowed_features,omitempty" example:"user_authentication,payment_processing"`
}

// OptimizationResponse includes both carbon intensity and optimization profile.
type OptimizationResponse struct {
	// CarbonIntensity is the current carbon intensity data
	CarbonIntensity *carbon.CarbonIntensity `json:"carbon_intensity" validate:"required"`

	// Optimization is the generated optimization profile
	Optimization *OptimizationProfile `json:"optimization" validate:"required"`

	// URL is the website URL that was optimized (if provided)
	URL string `json:"url,omitempty" example:"https://example.com"`

	// Recommendations contains human-readable optimization recommendations
	Recommendations []string `json:"recommendations,omitempty" example:"Reduce image quality to medium"`

	// GeneratedAt indicates when this response was generated
	GeneratedAt time.Time `json:"generated_at" validate:"required" example:"2024-01-15T14:30:00Z"`

	// RequestID is a unique identifier for this optimization request
	RequestID string `json:"request_id,omitempty" example:"req_123456789"`
}

// OptimizationRule represents a rule for generating optimization profiles.
type OptimizationRule struct {
	// ID is a unique identifier for this rule
	ID string `json:"id" validate:"required" example:"high_carbon_video"`

	// Name is a human-readable name for this rule
	Name string `json:"name" validate:"required" example:"Disable video autoplay during high carbon periods"`

	// Description provides more details about what this rule does
	Description string `json:"description" example:"Automatically disables video autoplay when carbon intensity exceeds 300 g CO2/kWh"`

	// Conditions defines when this rule should be applied
	Conditions []RuleCondition `json:"conditions" validate:"required,dive"`

	// Actions defines what optimizations to apply when conditions are met
	Actions []RuleAction `json:"actions" validate:"required,dive"`

	// Priority determines the order in which rules are evaluated (higher = earlier)
	Priority int `json:"priority" validate:"min=0" example:"100"`

	// Enabled indicates if this rule is currently active
	Enabled bool `json:"enabled" example:"true"`

	// Tags provide categorization for rules
	Tags []string `json:"tags,omitempty" example:"video,autoplay,energy"`

	// CreatedAt indicates when this rule was created
	CreatedAt time.Time `json:"created_at" example:"2024-01-15T14:30:00Z"`

	// UpdatedAt indicates when this rule was last modified
	UpdatedAt time.Time `json:"updated_at" example:"2024-01-15T14:30:00Z"`
}

// RuleCondition represents a condition that must be met for a rule to apply.
type RuleCondition struct {
	// Type is the type of condition (carbon_intensity, location, time, etc.)
	Type string `json:"type" validate:"required" example:"carbon_intensity"`

	// Operator is the comparison operator (gt, lt, eq, ne, in, not_in, etc.)
	Operator string `json:"operator" validate:"required" example:"gt"`

	// Value is the value to compare against
	Value interface{} `json:"value" validate:"required" example:"300"`

	// Field is the specific field to check (optional, depends on type)
	Field string `json:"field,omitempty" example:"carbon_intensity"`
}

// RuleAction represents an action to take when rule conditions are met.
type RuleAction struct {
	// Type is the type of action (disable_feature, set_quality, etc.)
	Type string `json:"type" validate:"required" example:"disable_feature"`

	// Target is what the action should affect
	Target string `json:"target" validate:"required" example:"video_autoplay"`

	// Value is the value to set (optional, depends on action type)
	Value interface{} `json:"value,omitempty" example:"false"`

	// Metadata contains additional action-specific data
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// Evaluate returns true if all conditions in the rule are met.
func (r *OptimizationRule) Evaluate(context *OptimizationContext) bool {
	if !r.Enabled {
		return false
	}

	for _, condition := range r.Conditions {
		if !condition.Evaluate(context) {
			return false
		}
	}

	return true
}

// OptimizationContext provides context for rule evaluation.
type OptimizationContext struct {
	// CarbonIntensity is the current carbon intensity
	CarbonIntensity *carbon.CarbonIntensity

	// Request is the optimization request
	Request *OptimizationRequest

	// Timestamp is when the optimization is being performed
	Timestamp time.Time

	// CustomData contains additional context data
	CustomData map[string]interface{}
}

// Evaluate returns true if this condition is met in the given context.
func (c *RuleCondition) Evaluate(context *OptimizationContext) bool {
	var actualValue interface{}

	// Extract the actual value based on condition type
	switch c.Type {
	case "carbon_intensity":
		if context.CarbonIntensity != nil {
			actualValue = context.CarbonIntensity.CarbonIntensity
		}
	case "renewable_percentage":
		if context.CarbonIntensity != nil {
			actualValue = context.CarbonIntensity.RenewablePercent
		}
	case "mode":
		if context.CarbonIntensity != nil {
			actualValue = context.CarbonIntensity.Mode
		}
	case "location":
		if context.Request != nil {
			actualValue = context.Request.Location
		}
	case "url":
		if context.Request != nil {
			actualValue = context.Request.URL
		}
	case "time":
		actualValue = context.Timestamp
	case "device_type":
		if context.Request != nil {
			actualValue = context.Request.DeviceType
		}
	case "connection_type":
		if context.Request != nil {
			actualValue = context.Request.ConnectionType
		}
	default:
		// Check custom data
		if context.CustomData != nil {
			actualValue = context.CustomData[c.Type]
		}
	}

	// Perform comparison based on operator
	return c.compareValues(actualValue, c.Value, c.Operator)
}

// compareValues performs comparison between actual and expected values.
func (c *RuleCondition) compareValues(actual, expected interface{}, operator string) bool {
	switch operator {
	case "eq":
		return actual == expected
	case "ne":
		return actual != expected
	case "gt":
		return compareNumeric(actual, expected, func(a, b float64) bool { return a > b })
	case "gte":
		return compareNumeric(actual, expected, func(a, b float64) bool { return a >= b })
	case "lt":
		return compareNumeric(actual, expected, func(a, b float64) bool { return a < b })
	case "lte":
		return compareNumeric(actual, expected, func(a, b float64) bool { return a <= b })
	case "in":
		return containsValue(actual, expected)
	case "not_in":
		return !containsValue(actual, expected)
	case "contains":
		return strings.Contains(fmt.Sprintf("%v", actual), fmt.Sprintf("%v", expected))
	case "not_contains":
		return !strings.Contains(fmt.Sprintf("%v", actual), fmt.Sprintf("%v", expected))
	default:
		return false
	}
}

// compareNumeric compares two values numerically.
func compareNumeric(actual, expected interface{}, compareFn func(float64, float64) bool) bool {
	a, ok1 := toFloat64(actual)
	b, ok2 := toFloat64(expected)
	if !ok1 || !ok2 {
		return false
	}
	return compareFn(a, b)
}

// toFloat64 converts an interface{} to float64.
func toFloat64(v interface{}) (float64, bool) {
	switch val := v.(type) {
	case float64:
		return val, true
	case float32:
		return float64(val), true
	case int:
		return float64(val), true
	case int32:
		return float64(val), true
	case int64:
		return float64(val), true
	default:
		return 0, false
	}
}

// containsValue checks if a value is contained in a slice or array.
func containsValue(actual, expected interface{}) bool {
	switch exp := expected.(type) {
	case []interface{}:
		for _, item := range exp {
			if actual == item {
				return true
			}
		}
	case []string:
		actualStr := fmt.Sprintf("%v", actual)
		for _, item := range exp {
			if actualStr == item {
				return true
			}
		}
	case []int:
		if actualInt, ok := actual.(int); ok {
			for _, item := range exp {
				if actualInt == item {
					return true
				}
			}
		}
	case []float64:
		if actualFloat, ok := toFloat64(actual); ok {
			for _, item := range exp {
				if actualFloat == item {
					return true
				}
			}
		}
	}
	return false
}

// HighImpactOptimizations contains features with significant CO2 reduction potential
type HighImpactOptimizations struct {
	// VideoStreamingOptimization contains video streaming quality controls
	VideoStreamingOptimization VideoStreamingOptimization `json:"video_streaming_optimization"`

	// AIInferenceOptimization contains AI/LLM inference deferral settings
	AIInferenceOptimization AIInferenceOptimization `json:"ai_inference_optimization"`

	// GPUFeatureOptimization contains GPU-intensive feature controls
	GPUFeatureOptimization GPUFeatureOptimization `json:"gpu_feature_optimization"`

	// JavaScriptBundleOptimization contains JS bundle optimization settings
	JavaScriptBundleOptimization JavaScriptBundleOptimization `json:"javascript_bundle_optimization"`

	// ServerSideOptimization contains server vs client rendering decisions
	ServerSideOptimization ServerSideOptimization `json:"server_side_optimization"`

	// ImageOptimization contains advanced image optimization settings
	ImageOptimization AdvancedImageOptimization `json:"image_optimization"`
}

// VideoStreamingOptimization controls video streaming quality and bandwidth
type VideoStreamingOptimization struct {
	// Enabled indicates if video optimization is active
	Enabled bool `json:"enabled" example:"true"`

	// MaxQuality is the maximum allowed video quality
	MaxQuality VideoQuality `json:"max_quality" example:"720p"`

	// BitrateLimit is the maximum bitrate in Mbps
	BitrateLimit float64 `json:"bitrate_limit" example:"2.5"`

	// AutoplayDisabled indicates if autoplay should be disabled
	AutoplayDisabled bool `json:"autoplay_disabled" example:"true"`

	// PreloadStrategy controls video preloading behavior
	PreloadStrategy string `json:"preload_strategy" example:"metadata"`

	// AdaptiveBitrateEnabled enables dynamic quality adjustment
	AdaptiveBitrateEnabled bool `json:"adaptive_bitrate_enabled" example:"true"`

	// EstimatedCO2Savings is the estimated CO2 savings in grams per hour
	EstimatedCO2Savings float64 `json:"estimated_co2_savings" example:"24.0"`
}

// AIInferenceOptimization controls AI/LLM feature deferral
type AIInferenceOptimization struct {
	// DeferToGreenWindows defers AI operations to low-carbon periods
	DeferToGreenWindows bool `json:"defer_to_green_windows" example:"true"`

	// MaxInferencesPerSession limits AI operations per user session
	MaxInferencesPerSession int `json:"max_inferences_per_session" example:"5"`

	// UseLocalModels prefers local models over cloud APIs when possible
	UseLocalModels bool `json:"use_local_models" example:"true"`

	// BatchInferences groups AI requests for efficiency
	BatchInferences bool `json:"batch_inferences" example:"true"`

	// DisabledFeatures lists AI features to disable
	DisabledFeatures []string `json:"disabled_features" example:"ai_recommendations,smart_search"`

	// EstimatedCO2Savings is the estimated CO2 savings per session
	EstimatedCO2Savings float64 `json:"estimated_co2_savings" example:"3.0"`
}

// GPUFeatureOptimization controls GPU-intensive features
type GPUFeatureOptimization struct {
	// Disable3DModels disables 3D model rendering
	Disable3DModels bool `json:"disable_3d_models" example:"true"`

	// DisableWebGL disables WebGL features
	DisableWebGL bool `json:"disable_webgl" example:"true"`

	// ReduceCanvasResolution reduces canvas rendering resolution
	ReduceCanvasResolution bool `json:"reduce_canvas_resolution" example:"true"`

	// SimplifyShaders uses simpler shader programs
	SimplifyShaders bool `json:"simplify_shaders" example:"true"`

	// MaxFPS limits maximum frames per second
	MaxFPS int `json:"max_fps" example:"30"`

	// EstimatedCO2Savings is the estimated CO2 savings per hour
	EstimatedCO2Savings float64 `json:"estimated_co2_savings" example:"15.0"`
}

// JavaScriptBundleOptimization controls JavaScript bundle loading
type JavaScriptBundleOptimization struct {
	// EnableCodeSplitting enables dynamic code splitting
	EnableCodeSplitting bool `json:"enable_code_splitting" example:"true"`

	// LazyLoadNonCritical defers non-critical JS loading
	LazyLoadNonCritical bool `json:"lazy_load_non_critical" example:"true"`

	// MaxBundleSize limits individual bundle size in KB
	MaxBundleSize int `json:"max_bundle_size" example:"250"`

	// DisablePolyfills removes unnecessary polyfills
	DisablePolyfills bool `json:"disable_polyfills" example:"true"`

	// TreeShakeAggressive enables aggressive tree shaking
	TreeShakeAggressive bool `json:"tree_shake_aggressive" example:"true"`

	// EstimatedCO2Savings is the estimated CO2 savings per page load
	EstimatedCO2Savings float64 `json:"estimated_co2_savings" example:"0.5"`
}

// ServerSideOptimization controls server vs client rendering
type ServerSideOptimization struct {
	// PreferServerSideRendering uses SSR over client-side rendering
	PreferServerSideRendering bool `json:"prefer_server_side_rendering" example:"true"`

	// EnableStaticGeneration uses static generation where possible
	EnableStaticGeneration bool `json:"enable_static_generation" example:"true"`

	// EdgeCachingEnabled enables edge caching
	EdgeCachingEnabled bool `json:"edge_caching_enabled" example:"true"`

	// ReduceAPICallsToServer minimizes API calls
	ReduceAPICallsToServer bool `json:"reduce_api_calls_to_server" example:"true"`

	// EstimatedCO2Savings is the estimated CO2 savings per request
	EstimatedCO2Savings float64 `json:"estimated_co2_savings" example:"0.2"`
}

// AdvancedImageOptimization provides detailed image optimization settings
type AdvancedImageOptimization struct {
	// UseWebPFormat converts images to WebP format
	UseWebPFormat bool `json:"use_webp_format" example:"true"`

	// UseAVIFFormat uses AVIF for better compression
	UseAVIFFormat bool `json:"use_avif_format" example:"true"`

	// ProgressiveLoading enables progressive image loading
	ProgressiveLoading bool `json:"progressive_loading" example:"true"`

	// MaxImageDimensions limits image dimensions
	MaxImageDimensions ImageDimensions `json:"max_image_dimensions"`

	// CompressionQuality sets JPEG/WebP quality (0-100)
	CompressionQuality int `json:"compression_quality" example:"75"`

	// EstimatedCO2Savings is the estimated CO2 savings per MB saved
	EstimatedCO2Savings float64 `json:"estimated_co2_savings" example:"0.1"`
}

// ImageDimensions represents maximum image dimensions
type ImageDimensions struct {
	// MaxWidth in pixels
	MaxWidth int `json:"max_width" example:"1920"`

	// MaxHeight in pixels
	MaxHeight int `json:"max_height" example:"1080"`
}

// FeatureImpact represents the CO2 impact of a specific feature
type FeatureImpact struct {
	// FeatureName is the name of the feature
	FeatureName string `json:"feature_name" example:"4K Video Streaming"`

	// BaselineCO2PerHour is the baseline CO2 emissions in grams per hour
	BaselineCO2PerHour float64 `json:"baseline_co2_per_hour" example:"36.0"`

	// OptimizedCO2PerHour is the optimized CO2 emissions in grams per hour
	OptimizedCO2PerHour float64 `json:"optimized_co2_per_hour" example:"12.0"`

	// ReductionPercentage is the percentage reduction in CO2
	ReductionPercentage float64 `json:"reduction_percentage" example:"66.7"`

	// ImpactScore rates the impact from 1-10 (10 being highest impact)
	ImpactScore float64 `json:"impact_score" example:"9.5"`

	// EffortScore rates the implementation effort from 1-10 (10 being highest effort)
	EffortScore float64 `json:"effort_score" example:"3.0"`

	// Priority is the impact/effort ratio (higher is better)
	Priority float64 `json:"priority" example:"3.17"`
}

// CO2SavingsBreakdown provides detailed CO2 savings calculations
type CO2SavingsBreakdown struct {
	// TotalSavingsPerHour is the total estimated CO2 savings in grams per hour
	TotalSavingsPerHour float64 `json:"total_savings_per_hour" example:"42.7"`

	// VideoStreamingSavings from video quality reduction
	VideoStreamingSavings float64 `json:"video_streaming_savings" example:"24.0"`

	// AIInferenceSavings from deferred AI operations
	AIInferenceSavings float64 `json:"ai_inference_savings" example:"3.0"`

	// GPUFeatureSavings from disabled GPU features
	GPUFeatureSavings float64 `json:"gpu_feature_savings" example:"15.0"`

	// ImageOptimizationSavings from image compression
	ImageOptimizationSavings float64 `json:"image_optimization_savings" example:"0.5"`

	// JavaScriptSavings from bundle optimization
	JavaScriptSavings float64 `json:"javascript_savings" example:"0.2"`

	// NetworkTransferReduction in MB per hour
	NetworkTransferReduction float64 `json:"network_transfer_reduction" example:"850"`

	// EndpointEnergyReduction in Wh per hour
	EndpointEnergyReduction float64 `json:"endpoint_energy_reduction" example:"5.2"`

	// CalculationMethod describes how savings were calculated
	CalculationMethod string `json:"calculation_method" example:"EU energy mix (401g CO2/kWh)"`
}