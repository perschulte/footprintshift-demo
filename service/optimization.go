package service

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/perschulte/greenweb-api/pkg/carbon"
	"github.com/perschulte/greenweb-api/pkg/optimization"
)

// CarbonIntensityProvider defines the interface for carbon intensity services
type CarbonIntensityProvider interface {
	GetCarbonIntensity(ctx context.Context, location string) (*carbon.CarbonIntensity, error)
}

// OptimizationService handles website optimization recommendations
type OptimizationService struct {
	electricityMaps CarbonIntensityProvider
	logger          *slog.Logger
}

// OptimizationProfile is an alias for backward compatibility.
// New code should use github.com/perschulte/greenweb-api/pkg/optimization.OptimizationProfile
type OptimizationProfile = optimization.OptimizationProfile

// OptimizationRequest is an alias for backward compatibility.
// New code should use github.com/perschulte/greenweb-api/pkg/optimization.OptimizationRequest
type OptimizationRequest = optimization.OptimizationRequest

// OptimizationResponse is an alias for backward compatibility.
// New code should use github.com/perschulte/greenweb-api/pkg/optimization.OptimizationResponse
type OptimizationResponse = optimization.OptimizationResponse

// NewOptimizationService creates a new optimization service
func NewOptimizationService(electricityMaps CarbonIntensityProvider, logger *slog.Logger) *OptimizationService {
	return &OptimizationService{
		electricityMaps: electricityMaps,
		logger:          logger,
	}
}

// GetOptimizationProfile generates optimization recommendations based on carbon intensity
func (s *OptimizationService) GetOptimizationProfile(ctx context.Context, req optimization.OptimizationRequest) (*optimization.OptimizationResponse, error) {
	// Get current carbon intensity
	intensity, err := s.electricityMaps.GetCarbonIntensity(ctx, req.Location)
	if err != nil {
		s.logger.Error("Failed to get carbon intensity for optimization", "error", err, "location", req.Location)
		return nil, err
	}

	// Generate base optimization profile
	profile := s.generateProfile(intensity.CarbonIntensity)

	// Set profile metadata
	profile.GeneratedAt = time.Now()
	profile.ValidUntil = time.Now().Add(15 * time.Minute)
	profile.Metadata = optimization.OptimizationMetadata{
		CarbonIntensity: intensity.CarbonIntensity,
		Location:       req.Location,
		Source:         "greenweb-api",
		Version:        "1.0.0",
	}

	// Add URL-specific optimizations if provided
	if req.URL != "" {
		s.applyURLSpecificOptimizations(profile, req.URL)
	}

	s.logger.Info("Generated optimization profile",
		"location", req.Location,
		"url", req.URL,
		"carbon_intensity", intensity.CarbonIntensity,
		"mode", profile.Mode)

	return &optimization.OptimizationResponse{
		CarbonIntensity: intensity,
		Optimization:    profile,
		URL:             req.URL,
		GeneratedAt:     time.Now(),
	}, nil
}

// generateProfile creates an optimization profile based on carbon intensity
func (s *OptimizationService) generateProfile(intensity float64) *optimization.OptimizationProfile {
	profile := &optimization.OptimizationProfile{
		FeatureImpactScores: make(map[string]optimization.FeatureImpact),
	}

	// Calculate high-impact optimizations and CO2 savings
	highImpact := s.calculateHighImpactOptimizations(intensity)
	co2Savings := s.calculateCO2Savings(highImpact, intensity)

	if intensity < 150 {
		// Green mode - all features enabled, promote green behavior
		profile.Mode = optimization.ModeFull
		profile.DisableFeatures = []string{}
		profile.ImageQuality = optimization.ImageQualityHigh
		profile.VideoQuality = optimization.VideoQuality1080p
		profile.DeferAnalytics = false
		profile.EcoDiscount = 5 // Reward green behavior with discount
		profile.ShowGreenBanner = true
		profile.CachingStrategy = optimization.CachingNormal
		profile.UIOptimizations = optimization.UIOptimizations{
			ReduceAnimations: false,
			LazyLoadImages:   true,
		}
		profile.ContentOptimizations = optimization.ContentOptimizations{
			CompressImages:   false,
			MinifyCSS:        true,
			MinifyJavaScript: true,
		}
		// Even in green mode, promote 4K to 1080p reduction for significant savings
		highImpact.VideoStreamingOptimization = optimization.VideoStreamingOptimization{
			Enabled:                true,
			MaxQuality:             optimization.VideoQuality1080p,
			BitrateLimit:           8.0, // 1080p bitrate
			AutoplayDisabled:       false,
			PreloadStrategy:        "metadata",
			AdaptiveBitrateEnabled: true,
			EstimatedCO2Savings:    12.0, // 4K to 1080p saves ~12g CO2/hour
		}
	} else if intensity < 300 {
		// Yellow mode - focus on high-impact optimizations
		profile.Mode = optimization.ModeNormal
		profile.DisableFeatures = []string{"video_autoplay", "3d_models", "webgl_effects"}
		profile.ImageQuality = optimization.ImageQualityMedium
		profile.VideoQuality = optimization.VideoQuality720p
		profile.DeferAnalytics = false
		profile.EcoDiscount = 0
		profile.ShowGreenBanner = false
		profile.CachingStrategy = optimization.CachingAggressive
		profile.UIOptimizations = optimization.UIOptimizations{
			ReduceAnimations:   true,
			LazyLoadImages:     true,
			MinimizeJavaScript: true,
		}
		profile.ContentOptimizations = optimization.ContentOptimizations{
			CompressImages:   true,
			MinifyCSS:        true,
			MinifyJavaScript: true,
			MinifyHTML:       true,
		}
	} else if intensity < 500 {
		// Red mode - aggressive high-impact optimizations
		profile.Mode = optimization.ModeEco
		profile.DisableFeatures = []string{
			"video_autoplay", "3d_models", "webgl_features", "gpu_animations",
			"ai_recommendations", "ai_search", "live_chat_ai", "realtime_analytics",
		}
		profile.ImageQuality = optimization.ImageQualityLow
		profile.VideoQuality = optimization.VideoQuality480p
		profile.DeferAnalytics = true
		profile.EcoDiscount = 0
		profile.ShowGreenBanner = false
		profile.CachingStrategy = optimization.CachingAggressive
		profile.UIOptimizations = optimization.UIOptimizations{
			ReduceAnimations:   true,
			DisableTransitions: true,
			LazyLoadImages:     true,
			MinimizeJavaScript: true,
			PreferSystemFonts:  true,
		}
		profile.ContentOptimizations = optimization.ContentOptimizations{
			CompressImages:          true,
			CompressVideos:          true,
			MinifyCSS:               true,
			MinifyJavaScript:        true,
			MinifyHTML:              true,
			EnableGzipCompression:   true,
			EnableBrotliCompression: true,
			RemoveComments:          true,
		}
	} else {
		// Critical mode - maximum high-impact optimizations
		profile.Mode = optimization.ModeCritical
		profile.DisableFeatures = []string{
			"video_streaming", "3d_models", "webgl_features", "gpu_animations",
			"ai_features", "ml_inference", "live_updates", "realtime_features",
			"background_videos", "animated_backgrounds", "parallax_effects",
		}
		profile.ImageQuality = optimization.ImageQualityLow
		profile.VideoQuality = optimization.VideoQuality360p
		profile.DeferAnalytics = true
		profile.EcoDiscount = 0
		profile.ShowGreenBanner = false
		profile.CachingStrategy = optimization.CachingAggressive
		profile.UIOptimizations = optimization.UIOptimizations{
			ReduceAnimations:   true,
			SimplifyLayouts:    true,
			DisableTransitions: true,
			ReduceColors:       true,
			LazyLoadImages:     true,
			MinimizeJavaScript: true,
			PreferSystemFonts:  true,
		}
		profile.ContentOptimizations = optimization.ContentOptimizations{
			CompressImages:          true,
			CompressVideos:          true,
			MinifyCSS:               true,
			MinifyJavaScript:        true,
			MinifyHTML:              true,
			EnableGzipCompression:   true,
			EnableBrotliCompression: true,
			RemoveComments:          true,
			OptimizeCriticalCSS:     true,
		}
	}

	// Apply high-impact optimizations and calculate feature impact scores
	profile.HighImpactOptimizations = highImpact
	profile.EstimatedCO2Savings = co2Savings
	s.calculateFeatureImpactScores(profile, intensity)

	return profile
}

// applyURLSpecificOptimizations adds website-specific optimization recommendations
func (s *OptimizationService) applyURLSpecificOptimizations(profile *optimization.OptimizationProfile, url string) {
	url = strings.ToLower(url)

	// E-commerce specific optimizations
	if s.isEcommerceSite(url) {
		if profile.Mode == optimization.ModeEco || profile.Mode == optimization.ModeCritical {
			// High-impact e-commerce optimizations
			profile.DisableFeatures = append(profile.DisableFeatures, "product_360_view", "zoom_on_hover", "ai_recommendations")
			
			// Optimize product videos specifically
			if profile.HighImpactOptimizations.VideoStreamingOptimization.Enabled {
				profile.HighImpactOptimizations.VideoStreamingOptimization.MaxQuality = optimization.VideoQuality720p
				profile.HighImpactOptimizations.VideoStreamingOptimization.AutoplayDisabled = true
			}
			
			// Defer AI-powered product recommendations
			if profile.HighImpactOptimizations.AIInferenceOptimization.DeferToGreenWindows {
				profile.HighImpactOptimizations.AIInferenceOptimization.DisabledFeatures = append(
					profile.HighImpactOptimizations.AIInferenceOptimization.DisabledFeatures,
					"product_recommendations", "smart_search", "personalization_engine",
				)
			}
		}
	}

	// Media-heavy sites (streaming, video platforms)
	if s.isMediaSite(url) {
		// These sites have the highest video impact
		if profile.Mode != optimization.ModeFull {
			profile.DisableFeatures = append(profile.DisableFeatures, "auto_thumbnails", "preview_videos", "background_videos")
			
			// Aggressive video optimization for media sites
			if profile.HighImpactOptimizations.VideoStreamingOptimization.Enabled {
				if profile.Mode == optimization.ModeEco {
					profile.HighImpactOptimizations.VideoStreamingOptimization.MaxQuality = optimization.VideoQuality480p
					profile.HighImpactOptimizations.VideoStreamingOptimization.BitrateLimit = 1.0
				} else if profile.Mode == optimization.ModeCritical {
					profile.HighImpactOptimizations.VideoStreamingOptimization.MaxQuality = optimization.VideoQuality360p
					profile.HighImpactOptimizations.VideoStreamingOptimization.BitrateLimit = 0.5
					profile.HighImpactOptimizations.VideoStreamingOptimization.PreloadStrategy = "none"
				}
			}
		}
	}

	// Social media platforms
	if s.isSocialMediaSite(url) {
		if profile.Mode == optimization.ModeEco || profile.Mode == optimization.ModeCritical {
			profile.DisableFeatures = append(profile.DisableFeatures, "infinite_scroll", "story_previews", "auto_refresh", "live_notifications")
			
			// Social media often has heavy GPU usage for filters/effects
			if profile.HighImpactOptimizations.GPUFeatureOptimization.Disable3DModels {
				profile.DisableFeatures = append(profile.DisableFeatures, "camera_filters", "ar_effects", "3d_stickers")
			}
			
			// Defer AI content moderation and recommendations
			if profile.HighImpactOptimizations.AIInferenceOptimization.DeferToGreenWindows {
				profile.HighImpactOptimizations.AIInferenceOptimization.DisabledFeatures = append(
					profile.HighImpactOptimizations.AIInferenceOptimization.DisabledFeatures,
					"content_suggestions", "auto_tagging", "smart_feeds",
				)
			}
		}
	}

	// News sites
	if s.isNewsSite(url) {
		if profile.Mode != optimization.ModeFull {
			profile.DisableFeatures = append(profile.DisableFeatures, "breaking_news_animations", "comment_sections", "live_updates")
			
			// News sites often have heavy JavaScript for live updates
			if profile.HighImpactOptimizations.JavaScriptBundleOptimization.EnableCodeSplitting {
				profile.HighImpactOptimizations.JavaScriptBundleOptimization.MaxBundleSize = 150
				profile.HighImpactOptimizations.JavaScriptBundleOptimization.LazyLoadNonCritical = true
			}
		}
	}

	// Gaming or interactive sites
	if s.isGamingSite(url) {
		if profile.Mode != optimization.ModeFull {
			// Gaming sites have extremely high GPU usage
			profile.DisableFeatures = append(profile.DisableFeatures, "webgl_games", "3d_graphics", "particle_effects")
			
			if profile.HighImpactOptimizations.GPUFeatureOptimization.Disable3DModels {
				profile.HighImpactOptimizations.GPUFeatureOptimization.MaxFPS = 30
				profile.HighImpactOptimizations.GPUFeatureOptimization.SimplifyShaders = true
			}
		}
	}

	// Generic website optimizations
	if profile.Mode != optimization.ModeFull {
		profile.DisableFeatures = append(profile.DisableFeatures, "custom_fonts", "web_fonts")
		
		// Apply image optimization based on site type
		if s.isMediaSite(url) || s.isEcommerceSite(url) {
			// Heavy image sites get more aggressive optimization
			profile.HighImpactOptimizations.ImageOptimization.CompressionQuality = 65
			profile.HighImpactOptimizations.ImageOptimization.MaxImageDimensions.MaxWidth = 1280
			profile.HighImpactOptimizations.ImageOptimization.MaxImageDimensions.MaxHeight = 720
		}
	}

	s.logger.Debug("Applied URL-specific optimizations",
		"url", url,
		"mode", profile.Mode,
		"disabled_features", len(profile.DisableFeatures))
}

// isEcommerceSite detects if a URL belongs to an e-commerce site
func (s *OptimizationService) isEcommerceSite(url string) bool {
	ecommerceIndicators := []string{
		"shop", "store", "cart", "buy", "purchase", "checkout",
		"amazon", "ebay", "shopify", "woocommerce", "magento",
		"product", "catalog", "marketplace",
	}

	for _, indicator := range ecommerceIndicators {
		if strings.Contains(url, indicator) {
			return true
		}
	}
	return false
}

// isMediaSite detects if a URL belongs to a media-heavy site
func (s *OptimizationService) isMediaSite(url string) bool {
	mediaIndicators := []string{
		"youtube", "vimeo", "netflix", "twitch", "instagram",
		"video", "stream", "media", "photo", "gallery",
		"images", "pictures", "multimedia",
	}

	for _, indicator := range mediaIndicators {
		if strings.Contains(url, indicator) {
			return true
		}
	}
	return false
}

// isSocialMediaSite detects if a URL belongs to a social media platform
func (s *OptimizationService) isSocialMediaSite(url string) bool {
	socialIndicators := []string{
		"facebook", "twitter", "linkedin", "instagram", "tiktok",
		"snapchat", "pinterest", "reddit", "discord", "telegram",
		"social", "community", "forum", "chat",
	}

	for _, indicator := range socialIndicators {
		if strings.Contains(url, indicator) {
			return true
		}
	}
	return false
}

// isNewsSite detects if a URL belongs to a news site
func (s *OptimizationService) isNewsSite(url string) bool {
	newsIndicators := []string{
		"news", "newspaper", "journal", "times", "post",
		"guardian", "bbc", "cnn", "reuters", "ap",
		"breaking", "headlines", "articles", "press",
	}

	for _, indicator := range newsIndicators {
		if strings.Contains(url, indicator) {
			return true
		}
	}
	return false
}

// isGamingSite detects if a URL belongs to a gaming or interactive site
func (s *OptimizationService) isGamingSite(url string) bool {
	gamingIndicators := []string{
		"game", "gaming", "play", "steam", "epic",
		"twitch", "itch", "kongregate", "unity",
		"webgl", "three", "babylonjs", "arcade",
	}

	for _, indicator := range gamingIndicators {
		if strings.Contains(url, indicator) {
			return true
		}
	}
	return false
}

// GetOptimizationRecommendations provides detailed recommendations for a website
func (s *OptimizationService) GetOptimizationRecommendations(ctx context.Context, req optimization.OptimizationRequest) ([]string, error) {
	resp, err := s.GetOptimizationProfile(ctx, req)
	if err != nil {
		return nil, err
	}

	var recommendations []string
	profile := resp.Optimization
	intensity := resp.CarbonIntensity

	// Add carbon intensity context with realistic savings
	if intensity.CarbonIntensity > 400 {
		recommendations = append(recommendations,
			fmt.Sprintf("Critical carbon intensity (%.0fg CO2/kWh) - implement high-impact optimizations", intensity.CarbonIntensity),
			fmt.Sprintf("Potential CO2 savings: %.1fg per hour with video quality reduction", profile.EstimatedCO2Savings.TotalSavingsPerHour),
		)
	} else if intensity.CarbonIntensity > 300 {
		recommendations = append(recommendations,
			fmt.Sprintf("High carbon intensity (%.0fg CO2/kWh) - focus on energy-intensive features", intensity.CarbonIntensity),
			fmt.Sprintf("Video streaming optimization can save %.1fg CO2/hour", profile.EstimatedCO2Savings.VideoStreamingSavings),
		)
	} else if intensity.CarbonIntensity > 150 {
		recommendations = append(recommendations,
			fmt.Sprintf("Moderate carbon intensity (%.0fg CO2/kWh) - apply targeted optimizations", intensity.CarbonIntensity),
			"Consider deferring AI-powered features to green windows",
		)
	} else {
		recommendations = append(recommendations,
			fmt.Sprintf("Low carbon intensity (%.0fg CO2/kWh) - green window opportunity", intensity.CarbonIntensity),
			"Ideal time for AI inference, 4K video, and GPU-intensive features",
		)
	}

	// High-impact video optimization recommendations
	if profile.HighImpactOptimizations.VideoStreamingOptimization.Enabled {
		videoOpt := profile.HighImpactOptimizations.VideoStreamingOptimization
		recommendations = append(recommendations,
			fmt.Sprintf("Limit video quality to %s (saves %.1fg CO2/hour)", videoOpt.MaxQuality, videoOpt.EstimatedCO2Savings),
			fmt.Sprintf("Set bitrate limit to %.1f Mbps", videoOpt.BitrateLimit),
		)
		if videoOpt.AutoplayDisabled {
			recommendations = append(recommendations, "Disable video autoplay to reduce immediate bandwidth")
		}
	}

	// AI inference recommendations
	if profile.HighImpactOptimizations.AIInferenceOptimization.DeferToGreenWindows {
		aiOpt := profile.HighImpactOptimizations.AIInferenceOptimization
		recommendations = append(recommendations,
			fmt.Sprintf("Defer AI features to green windows (saves %.1fg CO2/session)", aiOpt.EstimatedCO2Savings),
			fmt.Sprintf("Limit AI inferences to %d per session", aiOpt.MaxInferencesPerSession),
		)
		if len(aiOpt.DisabledFeatures) > 0 {
			recommendations = append(recommendations,
				"Disable AI features: "+strings.Join(aiOpt.DisabledFeatures, ", "),
			)
		}
	}

	// GPU feature recommendations
	if profile.HighImpactOptimizations.GPUFeatureOptimization.Disable3DModels {
		gpuOpt := profile.HighImpactOptimizations.GPUFeatureOptimization
		recommendations = append(recommendations,
			fmt.Sprintf("Disable GPU-intensive features (saves %.1fg CO2/hour)", gpuOpt.EstimatedCO2Savings),
		)
		if gpuOpt.DisableWebGL {
			recommendations = append(recommendations, "Disable WebGL rendering to reduce GPU load")
		}
		if gpuOpt.MaxFPS < 60 {
			recommendations = append(recommendations,
				fmt.Sprintf("Limit animations to %d FPS", gpuOpt.MaxFPS),
			)
		}
	}

	// JavaScript bundle recommendations
	if profile.HighImpactOptimizations.JavaScriptBundleOptimization.EnableCodeSplitting {
		jsOpt := profile.HighImpactOptimizations.JavaScriptBundleOptimization
		recommendations = append(recommendations,
			fmt.Sprintf("Limit JavaScript bundles to %dKB maximum", jsOpt.MaxBundleSize),
		)
		if jsOpt.LazyLoadNonCritical {
			recommendations = append(recommendations, "Lazy load non-critical JavaScript components")
		}
	}

	// Image optimization recommendations
	if profile.HighImpactOptimizations.ImageOptimization.UseWebPFormat {
		imgOpt := profile.HighImpactOptimizations.ImageOptimization
		recommendations = append(recommendations,
			fmt.Sprintf("Convert images to WebP format with %d%% quality", imgOpt.CompressionQuality),
			fmt.Sprintf("Limit image dimensions to %dx%d", imgOpt.MaxImageDimensions.MaxWidth, imgOpt.MaxImageDimensions.MaxHeight),
		)
	}

	// Network transfer reduction
	if profile.EstimatedCO2Savings.NetworkTransferReduction > 0 {
		recommendations = append(recommendations,
			fmt.Sprintf("Reduce network transfer by %.0f MB/hour through optimization", profile.EstimatedCO2Savings.NetworkTransferReduction),
		)
	}

	// Feature-specific recommendations
	if len(profile.DisableFeatures) > 0 {
		recommendations = append(recommendations,
			"Temporarily disable features: "+strings.Join(profile.DisableFeatures, ", "),
		)
	}

	// Caching recommendations for reduced server load
	if profile.CachingStrategy == optimization.CachingAggressive {
		recommendations = append(recommendations,
			"Implement aggressive caching to reduce server requests and energy",
		)
	}

	// Analytics recommendations
	if profile.DeferAnalytics {
		recommendations = append(recommendations,
			"Defer analytics and tracking scripts to reduce processing overhead",
		)
	}

	// Green incentive recommendations
	if profile.EcoDiscount > 0 {
		recommendations = append(recommendations,
			fmt.Sprintf("Offer %d%% eco-discount during green energy periods", profile.EcoDiscount),
		)
	}

	// Show calculation methodology
	recommendations = append(recommendations,
		"Calculation method: "+profile.EstimatedCO2Savings.CalculationMethod,
	)

	s.logger.Info("Generated optimization recommendations",
		"location", req.Location,
		"url", req.URL,
		"recommendation_count", len(recommendations),
		"total_co2_savings", profile.EstimatedCO2Savings.TotalSavingsPerHour)

	return recommendations, nil
}

// calculateHighImpactOptimizations determines which high-impact features to optimize
func (s *OptimizationService) calculateHighImpactOptimizations(intensity float64) optimization.HighImpactOptimizations {
	highImpact := optimization.HighImpactOptimizations{}

	// Video streaming optimization - highest impact
	if intensity > 100 {
		videoOpt := optimization.VideoStreamingOptimization{
			Enabled:                true,
			AdaptiveBitrateEnabled: true,
			PreloadStrategy:        "metadata",
		}

		if intensity < 200 {
			// Mild optimization: 4K → 1080p
			videoOpt.MaxQuality = optimization.VideoQuality1080p
			videoOpt.BitrateLimit = 8.0
			videoOpt.AutoplayDisabled = false
			videoOpt.EstimatedCO2Savings = 12.0
		} else if intensity < 400 {
			// Moderate optimization: max 720p
			videoOpt.MaxQuality = optimization.VideoQuality720p
			videoOpt.BitrateLimit = 2.5
			videoOpt.AutoplayDisabled = true
			videoOpt.EstimatedCO2Savings = 24.0
		} else {
			// Aggressive optimization: max 480p
			videoOpt.MaxQuality = optimization.VideoQuality480p
			videoOpt.BitrateLimit = 1.0
			videoOpt.AutoplayDisabled = true
			videoOpt.PreloadStrategy = "none"
			videoOpt.EstimatedCO2Savings = 30.0
		}
		highImpact.VideoStreamingOptimization = videoOpt
	}

	// AI/LLM inference optimization
	if intensity > 200 {
		aiOpt := optimization.AIInferenceOptimization{
			UseLocalModels:  true,
			BatchInferences: true,
		}

		if intensity < 350 {
			// Defer non-critical AI features
			aiOpt.DeferToGreenWindows = true
			aiOpt.MaxInferencesPerSession = 10
			aiOpt.DisabledFeatures = []string{"ai_recommendations"}
			aiOpt.EstimatedCO2Savings = 1.5
		} else {
			// Disable most AI features
			aiOpt.DeferToGreenWindows = true
			aiOpt.MaxInferencesPerSession = 3
			aiOpt.DisabledFeatures = []string{"ai_recommendations", "smart_search", "ai_chat"}
			aiOpt.EstimatedCO2Savings = 3.0
		}
		highImpact.AIInferenceOptimization = aiOpt
	}

	// GPU feature optimization
	if intensity > 250 {
		gpuOpt := optimization.GPUFeatureOptimization{
			MaxFPS: 60,
		}

		if intensity < 400 {
			// Reduce GPU features
			gpuOpt.Disable3DModels = true
			gpuOpt.ReduceCanvasResolution = true
			gpuOpt.MaxFPS = 30
			gpuOpt.EstimatedCO2Savings = 8.0
		} else {
			// Aggressive GPU optimization
			gpuOpt.Disable3DModels = true
			gpuOpt.DisableWebGL = true
			gpuOpt.ReduceCanvasResolution = true
			gpuOpt.SimplifyShaders = true
			gpuOpt.MaxFPS = 24
			gpuOpt.EstimatedCO2Savings = 15.0
		}
		highImpact.GPUFeatureOptimization = gpuOpt
	}

	// JavaScript bundle optimization
	if intensity > 150 {
		jsOpt := optimization.JavaScriptBundleOptimization{
			EnableCodeSplitting: true,
			MaxBundleSize:       500,
		}

		if intensity < 300 {
			// Basic JS optimization
			jsOpt.LazyLoadNonCritical = true
			jsOpt.MaxBundleSize = 250
			jsOpt.EstimatedCO2Savings = 0.2
		} else {
			// Aggressive JS optimization
			jsOpt.LazyLoadNonCritical = true
			jsOpt.MaxBundleSize = 100
			jsOpt.DisablePolyfills = true
			jsOpt.TreeShakeAggressive = true
			jsOpt.EstimatedCO2Savings = 0.5
		}
		highImpact.JavaScriptBundleOptimization = jsOpt
	}

	// Server-side optimization
	if intensity > 200 {
		serverOpt := optimization.ServerSideOptimization{
			EnableStaticGeneration: true,
			EdgeCachingEnabled:     true,
		}

		if intensity > 350 {
			// Prefer server-side rendering to reduce client computation
			serverOpt.PreferServerSideRendering = true
			serverOpt.ReduceAPICallsToServer = true
			serverOpt.EstimatedCO2Savings = 0.2
		}
		highImpact.ServerSideOptimization = serverOpt
	}

	// Advanced image optimization
	imageOpt := optimization.AdvancedImageOptimization{
		ProgressiveLoading: true,
		MaxImageDimensions: optimization.ImageDimensions{
			MaxWidth:  3840,
			MaxHeight: 2160,
		},
		CompressionQuality: 85,
	}

	if intensity > 150 {
		imageOpt.UseWebPFormat = true
		imageOpt.CompressionQuality = 75
		imageOpt.MaxImageDimensions.MaxWidth = 1920
		imageOpt.MaxImageDimensions.MaxHeight = 1080
		imageOpt.EstimatedCO2Savings = 0.1
	}

	if intensity > 300 {
		imageOpt.UseAVIFFormat = true
		imageOpt.CompressionQuality = 60
		imageOpt.MaxImageDimensions.MaxWidth = 1280
		imageOpt.MaxImageDimensions.MaxHeight = 720
		imageOpt.EstimatedCO2Savings = 0.3
	}

	highImpact.ImageOptimization = imageOpt

	return highImpact
}

// calculateCO2Savings calculates total CO2 savings from optimizations
func (s *OptimizationService) calculateCO2Savings(highImpact optimization.HighImpactOptimizations, intensity float64) optimization.CO2SavingsBreakdown {
	savings := optimization.CO2SavingsBreakdown{
		VideoStreamingSavings:    highImpact.VideoStreamingOptimization.EstimatedCO2Savings,
		AIInferenceSavings:       highImpact.AIInferenceOptimization.EstimatedCO2Savings,
		GPUFeatureSavings:        highImpact.GPUFeatureOptimization.EstimatedCO2Savings,
		ImageOptimizationSavings: highImpact.ImageOptimization.EstimatedCO2Savings,
		JavaScriptSavings:        highImpact.JavaScriptBundleOptimization.EstimatedCO2Savings,
	}

	// Calculate total savings
	savings.TotalSavingsPerHour = savings.VideoStreamingSavings +
		savings.AIInferenceSavings +
		savings.GPUFeatureSavings +
		savings.ImageOptimizationSavings +
		savings.JavaScriptSavings

	// Estimate network transfer reduction (MB/hour)
	if highImpact.VideoStreamingOptimization.Enabled {
		// 4K: ~7GB/hour, 1080p: ~3GB/hour, 720p: ~1.5GB/hour
		switch highImpact.VideoStreamingOptimization.MaxQuality {
		case optimization.VideoQuality720p:
			savings.NetworkTransferReduction += 5500 // 7GB - 1.5GB
		case optimization.VideoQuality1080p:
			savings.NetworkTransferReduction += 4000 // 7GB - 3GB
		case optimization.VideoQuality480p:
			savings.NetworkTransferReduction += 6300 // 7GB - 0.7GB
		}
	}

	// Estimate endpoint energy reduction
	if highImpact.GPUFeatureOptimization.Disable3DModels || highImpact.GPUFeatureOptimization.DisableWebGL {
		savings.EndpointEnergyReduction += 3.5 // GPU features can use 3-5W
	}
	if highImpact.AIInferenceOptimization.DeferToGreenWindows {
		savings.EndpointEnergyReduction += 1.5 // AI inference energy
	}

	// Set calculation method based on intensity
	if intensity < 200 {
		savings.CalculationMethod = "EU energy mix (401g CO2/kWh) - moderate carbon intensity"
	} else if intensity < 400 {
		savings.CalculationMethod = fmt.Sprintf("Current grid intensity (%.0fg CO2/kWh) - high carbon", intensity)
	} else {
		savings.CalculationMethod = fmt.Sprintf("Current grid intensity (%.0fg CO2/kWh) - critical carbon", intensity)
	}

	return savings
}

// calculateFeatureImpactScores calculates impact scores for each optimization
func (s *OptimizationService) calculateFeatureImpactScores(profile *optimization.OptimizationProfile, intensity float64) {
	// Video streaming impact
	if profile.HighImpactOptimizations.VideoStreamingOptimization.Enabled {
		videoImpact := optimization.FeatureImpact{
			FeatureName:         "Video Streaming Quality",
			BaselineCO2PerHour:  36.0, // 4K streaming baseline
			OptimizedCO2PerHour: 36.0 - profile.HighImpactOptimizations.VideoStreamingOptimization.EstimatedCO2Savings,
			ImpactScore:         9.5,
			EffortScore:         3.0,
		}
		videoImpact.ReductionPercentage = (videoImpact.BaselineCO2PerHour - videoImpact.OptimizedCO2PerHour) / videoImpact.BaselineCO2PerHour * 100
		videoImpact.Priority = videoImpact.ImpactScore / videoImpact.EffortScore
		profile.FeatureImpactScores["video_streaming"] = videoImpact
	}

	// AI/LLM impact
	if profile.HighImpactOptimizations.AIInferenceOptimization.DeferToGreenWindows {
		aiImpact := optimization.FeatureImpact{
			FeatureName:         "AI/LLM Inference",
			BaselineCO2PerHour:  3.0, // Estimated for regular AI usage
			OptimizedCO2PerHour: 3.0 - profile.HighImpactOptimizations.AIInferenceOptimization.EstimatedCO2Savings,
			ImpactScore:         7.0,
			EffortScore:         4.0,
		}
		aiImpact.ReductionPercentage = (aiImpact.BaselineCO2PerHour - aiImpact.OptimizedCO2PerHour) / aiImpact.BaselineCO2PerHour * 100
		aiImpact.Priority = aiImpact.ImpactScore / aiImpact.EffortScore
		profile.FeatureImpactScores["ai_inference"] = aiImpact
	}

	// GPU features impact
	if profile.HighImpactOptimizations.GPUFeatureOptimization.Disable3DModels {
		gpuImpact := optimization.FeatureImpact{
			FeatureName:         "GPU-Intensive Features",
			BaselineCO2PerHour:  15.0, // GPU at 30-50W
			OptimizedCO2PerHour: 15.0 - profile.HighImpactOptimizations.GPUFeatureOptimization.EstimatedCO2Savings,
			ImpactScore:         8.0,
			EffortScore:         5.0,
		}
		gpuImpact.ReductionPercentage = (gpuImpact.BaselineCO2PerHour - gpuImpact.OptimizedCO2PerHour) / gpuImpact.BaselineCO2PerHour * 100
		gpuImpact.Priority = gpuImpact.ImpactScore / gpuImpact.EffortScore
		profile.FeatureImpactScores["gpu_features"] = gpuImpact
	}

	// JavaScript bundle impact
	if profile.HighImpactOptimizations.JavaScriptBundleOptimization.EnableCodeSplitting {
		jsImpact := optimization.FeatureImpact{
			FeatureName:         "JavaScript Bundle Size",
			BaselineCO2PerHour:  0.5, // Per page load impact
			OptimizedCO2PerHour: 0.5 - profile.HighImpactOptimizations.JavaScriptBundleOptimization.EstimatedCO2Savings,
			ImpactScore:         5.0,
			EffortScore:         2.0,
		}
		jsImpact.ReductionPercentage = (jsImpact.BaselineCO2PerHour - jsImpact.OptimizedCO2PerHour) / jsImpact.BaselineCO2PerHour * 100
		jsImpact.Priority = jsImpact.ImpactScore / jsImpact.EffortScore
		profile.FeatureImpactScores["javascript_bundles"] = jsImpact
	}

	// Image optimization impact
	if profile.HighImpactOptimizations.ImageOptimization.UseWebPFormat {
		imageImpact := optimization.FeatureImpact{
			FeatureName:         "Image Optimization",
			BaselineCO2PerHour:  0.5, // Based on average image transfer
			OptimizedCO2PerHour: 0.5 - profile.HighImpactOptimizations.ImageOptimization.EstimatedCO2Savings,
			ImpactScore:         6.0,
			EffortScore:         1.5,
		}
		imageImpact.ReductionPercentage = (imageImpact.BaselineCO2PerHour - imageImpact.OptimizedCO2PerHour) / imageImpact.BaselineCO2PerHour * 100
		imageImpact.Priority = imageImpact.ImpactScore / imageImpact.EffortScore
		profile.FeatureImpactScores["image_optimization"] = imageImpact
	}
}

// MockElectricityMapsClient provides a mock implementation for demos and testing
type MockElectricityMapsClient struct {
	mockIntensity float64
}

// SetMockIntensity sets the carbon intensity value returned by the mock client
func (m *MockElectricityMapsClient) SetMockIntensity(intensity float64) {
	m.mockIntensity = intensity
}

// GetCarbonIntensity returns mock carbon intensity data
func (m *MockElectricityMapsClient) GetCarbonIntensity(ctx context.Context, location string) (*carbon.CarbonIntensity, error) {
	if m.mockIntensity == 0 {
		m.mockIntensity = 250.0 // Default moderate intensity
	}

	// Determine mode based on intensity
	var mode string
	if m.mockIntensity < 150 {
		mode = "green"
	} else if m.mockIntensity < 300 {
		mode = "yellow"
	} else if m.mockIntensity < 500 {
		mode = "red"
	} else {
		mode = "critical"
	}

	// Calculate renewable percentage (inverse relationship with intensity)
	renewablePercent := 100.0 - (m.mockIntensity/10.0)
	if renewablePercent < 0 {
		renewablePercent = 0
	}

	return &carbon.CarbonIntensity{
		CarbonIntensity:  m.mockIntensity,
		Location:         location,
		Timestamp:        time.Now(),
		Mode:             mode,
		RenewablePercent: renewablePercent,
		Source:           "mock-demo",
	}, nil
}