package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/perschulte/greenweb-api/pkg/carbon"
)

// DualGridOptimizationService provides CDN-aware optimization recommendations
type DualGridOptimizationService struct {
	electricityService *ElectricityMapsClient
	logger            *slog.Logger
}

// NewDualGridOptimizationService creates a new dual-grid optimization service
func NewDualGridOptimizationService(electricityService *ElectricityMapsClient, logger *slog.Logger) *DualGridOptimizationService {
	return &DualGridOptimizationService{
		electricityService: electricityService,
		logger:            logger,
	}
}

// OptimizationRecommendation represents content delivery optimization recommendations
type OptimizationRecommendation struct {
	Strategy         string                       `json:"strategy"`
	Description      string                       `json:"description"`
	EstimatedSavings float64                      `json:"estimated_savings_percent"`
	Implementation   []string                     `json:"implementation_steps"`
	EdgeSuggestions  []carbon.EdgeAlternative     `json:"edge_suggestions,omitempty"`
	TimingStrategy   *carbon.TimeBasedStrategy    `json:"timing_strategy,omitempty"`
}

// GetOptimizationForContentType returns optimization recommendations for specific content types
func (s *DualGridOptimizationService) GetOptimizationForContentType(ctx context.Context, 
	userLocation, edgeLocation, contentType, cdnProvider string) (*OptimizationRecommendation, error) {

	// Get dual grid analysis
	dualGrid, err := s.electricityService.GetDualGridCarbonIntensity(ctx, userLocation, edgeLocation, contentType)
	if err != nil {
		s.logger.Error("Failed to get dual grid analysis", "error", err)
		return s.getDefaultOptimization(contentType), nil
	}

	// Generate content-specific recommendations
	optimization := s.generateContentOptimization(dualGrid, cdnProvider)

	// Add edge alternatives if available
	if cdnProvider != "" {
		alternatives, err := s.electricityService.GetCDNAlternatives(ctx, userLocation, edgeLocation, cdnProvider, contentType, 3)
		if err == nil && len(alternatives) > 0 {
			optimization.EdgeSuggestions = alternatives
			optimization.EstimatedSavings += 15 // Additional savings from edge optimization
		}
	}

	s.logger.Info("Generated optimization recommendation",
		"user_location", userLocation,
		"edge_location", edgeLocation,
		"content_type", contentType,
		"strategy", optimization.Strategy,
		"estimated_savings", optimization.EstimatedSavings)

	return optimization, nil
}

// generateContentOptimization creates optimization recommendations based on dual-grid analysis
func (s *DualGridOptimizationService) generateContentOptimization(dualGrid *carbon.DualGridCarbonIntensity, cdnProvider string) *OptimizationRecommendation {
	optimization := &OptimizationRecommendation{
		Implementation: []string{},
	}

	// Determine strategy based on weighted intensity and content type
	switch {
	case dualGrid.WeightedIntensity < 150:
		optimization.Strategy = "green_acceleration"
		optimization.Description = "Both locations have low carbon intensity - optimal for content acceleration"
		optimization.EstimatedSavings = 5
		optimization.Implementation = append(optimization.Implementation,
			"Use aggressive caching strategies",
			"Pre-fetch related content",
			"Enable content compression")

	case dualGrid.WeightedIntensity < 300:
		optimization.Strategy = "balanced_optimization"
		optimization.Description = "Moderate carbon intensity detected - balance performance with efficiency"
		optimization.EstimatedSavings = 20
		optimization.Implementation = append(optimization.Implementation,
			"Implement smart caching with TTL optimization",
			"Use adaptive quality based on network conditions",
			"Batch non-critical requests")

	default:
		optimization.Strategy = "carbon_minimization"
		optimization.Description = "High carbon intensity - prioritize carbon reduction over performance"
		optimization.EstimatedSavings = 40
		optimization.Implementation = append(optimization.Implementation,
			"Defer non-essential operations",
			"Reduce content quality/resolution",
			"Implement aggressive local caching")
	}

	// Add content-specific optimizations
	switch dualGrid.ContentType {
	case "video":
		optimization.Implementation = append(optimization.Implementation,
			"Use adaptive bitrate streaming",
			"Pre-generate multiple quality levels",
			"Implement smart thumbnail generation")
		if dualGrid.WeightedIntensity > 250 {
			optimization.Implementation = append(optimization.Implementation,
				"Default to lower resolution during high-carbon periods",
				"Use AV1 codec for better compression")
		}

	case "api":
		optimization.Implementation = append(optimization.Implementation,
			"Implement response caching with intelligent TTL",
			"Use GraphQL to reduce over-fetching",
			"Batch API requests where possible")
		if dualGrid.WeightedIntensity > 250 {
			optimization.Implementation = append(optimization.Implementation,
				"Defer analytics and logging calls",
				"Use local storage for temporary data")
		}

	case "static":
		optimization.Implementation = append(optimization.Implementation,
			"Use aggressive CDN caching (1 year+ TTL)",
			"Implement service workers for offline capability",
			"Use WebP/AVIF formats for images")

	case "dynamic":
		optimization.Implementation = append(optimization.Implementation,
			"Cache database query results",
			"Use edge-side rendering where possible",
			"Implement progressive loading")

	case "ai":
		optimization.Implementation = append(optimization.Implementation,
			"Cache model inference results",
			"Use quantized models during high-carbon periods",
			"Implement request batching for AI operations")

	case "database":
		optimization.Implementation = append(optimization.Implementation,
			"Optimize database queries and indexing",
			"Use read replicas for query distribution",
			"Cache frequently accessed data at the edge")
	}

	// Add timing strategy if beneficial
	if dualGrid.Recommendation.TimeBasedStrategy != nil {
		optimization.TimingStrategy = dualGrid.Recommendation.TimeBasedStrategy
		optimization.Implementation = append(optimization.Implementation,
			fmt.Sprintf("Schedule operations during next green window at %s",
				dualGrid.Recommendation.TimeBasedStrategy.NextOptimalWindow.Start.Format("15:04")))
	}

	return optimization
}

// getDefaultOptimization provides fallback optimization recommendations
func (s *DualGridOptimizationService) getDefaultOptimization(contentType string) *OptimizationRecommendation {
	optimization := &OptimizationRecommendation{
		Strategy:         "standard_optimization",
		Description:      "Standard optimization practices (fallback mode)",
		EstimatedSavings: 15,
		Implementation: []string{
			"Enable gzip/brotli compression",
			"Use appropriate cache headers",
			"Optimize content delivery",
		},
	}

	// Add basic content-specific recommendations
	switch contentType {
	case "video":
		optimization.Implementation = append(optimization.Implementation,
			"Use adaptive bitrate streaming",
			"Implement progressive video loading")
	case "api":
		optimization.Implementation = append(optimization.Implementation,
			"Cache API responses appropriately",
			"Use request batching")
	case "static":
		optimization.Implementation = append(optimization.Implementation,
			"Use long-term caching for static assets",
			"Implement image optimization")
	}

	return optimization
}

// GetCDNOptimization provides CDN-specific optimization recommendations
func (s *DualGridOptimizationService) GetCDNOptimization(ctx context.Context, userLocation, cdnProvider string) (*OptimizationRecommendation, error) {
	// Get optimal edge location
	optimalEdge, err := s.electricityService.GetOptimalEdgeLocation(ctx, userLocation, cdnProvider, "static")
	if err != nil {
		s.logger.Warn("Failed to get optimal edge location", "error", err)
		return s.getDefaultCDNOptimization(cdnProvider), nil
	}

	optimization := &OptimizationRecommendation{
		Strategy:         "cdn_optimization",
		Description:      fmt.Sprintf("Optimize CDN configuration for %s", cdnProvider),
		EstimatedSavings: 25,
		Implementation: []string{
			fmt.Sprintf("Use %s edge location for optimal carbon efficiency", optimalEdge.Location),
			"Configure intelligent cache purging",
			"Implement origin shield configuration",
		},
		EdgeSuggestions: []carbon.EdgeAlternative{*optimalEdge},
	}

	// Add CDN-specific recommendations
	switch cdnProvider {
	case "cloudflare":
		optimization.Implementation = append(optimization.Implementation,
			"Enable Cloudflare's carbon reduction features",
			"Use Argo Smart Routing for optimal path selection",
			"Enable Polish for automatic image optimization")

	case "aws-cloudfront":
		optimization.Implementation = append(optimization.Implementation,
			"Use AWS's renewable energy regions when possible",
			"Configure Origin Request Policies for cache optimization",
			"Implement Lambda@Edge for edge processing")

	case "google-cloud":
		optimization.Implementation = append(optimization.Implementation,
			"Leverage Google's carbon-neutral infrastructure",
			"Use Cloud CDN with Load Balancer integration",
			"Implement Smart Resize for images")

	case "azure":
		optimization.Implementation = append(optimization.Implementation,
			"Use Azure's sustainability features",
			"Configure Dynamic Site Acceleration",
			"Implement Azure Front Door optimization")
	}

	return optimization, nil
}

// getDefaultCDNOptimization provides fallback CDN optimization
func (s *DualGridOptimizationService) getDefaultCDNOptimization(cdnProvider string) *OptimizationRecommendation {
	return &OptimizationRecommendation{
		Strategy:         "basic_cdn_optimization",
		Description:      "Basic CDN optimization practices",
		EstimatedSavings: 15,
		Implementation: []string{
			"Use nearest edge location",
			"Configure appropriate cache TTLs",
			"Enable compression at the edge",
		},
	}
}