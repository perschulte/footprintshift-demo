package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/perschulte/greenweb-api/internal/impact"
)

func main() {
	fmt.Println("🌱 GreenWeb Impact Calculator Demo")
	fmt.Println("==================================")
	fmt.Println()

	// Create the calculator
	calculator := impact.NewCalculator()

	// Create storage and service
	storage := impact.NewMockStorageWithData()
	service := impact.NewService(storage)

	// Demo 1: Video Streaming Impact
	fmt.Println("📺 Video Streaming Impact (1080p, 1 hour, laptop, EU)")
	videoReq := &impact.CalculationRequest{
		Type:                   impact.ImpactTypeVideoStreaming,
		Duration:               3600, // 1 hour
		VideoQuality:           "1080p",
		DeviceType:             "laptop",
		ConnectionType:         "wifi",
		Region:                 "EU",
		OptimizationLevel:      30,
		IncludeReboundEffects:  true,
	}

	videoResult, err := calculator.Calculate(videoReq)
	if err != nil {
		log.Fatalf("Video calculation failed: %v", err)
	}

	fmt.Printf("   Baseline emissions: %.1f g CO2\n", videoResult.BaselineEmissions)
	fmt.Printf("   Optimized emissions: %.1f g CO2\n", videoResult.OptimizedEmissions)
	fmt.Printf("   Gross savings: %.1f g CO2 (%.1f%%)\n", videoResult.Savings, videoResult.SavingsPercentage)
	fmt.Printf("   Rebound effect: %.1f g CO2\n", videoResult.ReboundEffect)
	fmt.Printf("   Net savings: %.1f g CO2\n", videoResult.NetSavings)
	fmt.Printf("   Confidence: ±%.0f%%\n", videoResult.ConfidenceInterval)
	
	fmt.Printf("   Components: Device %.1f%%, Network %.1f%%, Data Center %.1f%%\n",
		videoResult.Components.DevicePercentage,
		videoResult.Components.NetworkPercentage,
		videoResult.Components.DataCenterPercentage)
	fmt.Println()

	// Demo 2: Image Optimization Impact
	fmt.Println("🖼️  Image Optimization Impact (50 images, 25MB, smartphone, 4G, aggressive optimization)")
	imageReq := &impact.CalculationRequest{
		Type:              impact.ImpactTypeImageLoading,
		ImageCount:        50,
		DataSize:          25.0,
		DeviceType:        "smartphone",
		ConnectionType:    "mobile_4g",
		Region:            "US",
		OptimizationLevel: 70, // Aggressive WebP/AVIF optimization
	}

	imageResult, err := calculator.Calculate(imageReq)
	if err != nil {
		log.Fatalf("Image calculation failed: %v", err)
	}

	fmt.Printf("   Baseline emissions: %.1f g CO2\n", imageResult.BaselineEmissions)
	fmt.Printf("   Optimized emissions: %.1f g CO2\n", imageResult.OptimizedEmissions)
	fmt.Printf("   Savings: %.1f g CO2 (%.1f%%)\n", imageResult.Savings, imageResult.SavingsPercentage)
	fmt.Printf("   Mobile 4G network: %.1f%% of total impact\n", imageResult.Components.NetworkPercentage)
	fmt.Println()

	// Demo 3: AI Inference Impact
	fmt.Println("🤖 AI Inference Impact (5 inferences, GPT-3 scale, with caching)")
	aiReq := &impact.CalculationRequest{
		Type:              impact.ImpactTypeAIInference,
		Duration:          10.0, // 10 seconds (5 inferences at 2s each)
		DeviceType:        "laptop",
		ConnectionType:    "wifi",
		Region:            "US",
		OptimizationLevel: 50, // Caching optimization
	}

	aiResult, err := calculator.Calculate(aiReq)
	if err != nil {
		log.Fatalf("AI calculation failed: %v", err)
	}

	fmt.Printf("   Baseline emissions: %.1f g CO2\n", aiResult.BaselineEmissions)
	fmt.Printf("   Optimized emissions: %.1f g CO2\n", aiResult.OptimizedEmissions)
	fmt.Printf("   Savings from caching: %.1f g CO2 (%.1f%%)\n", aiResult.Savings, aiResult.SavingsPercentage)
	fmt.Printf("   Data center intensive: %.1f%% of total impact\n", aiResult.Components.DataCenterPercentage)
	fmt.Println()

	// Demo 4: Validation (Anti-Greenwashing)
	fmt.Println("🛡️  Anti-Greenwashing Validation")
	validationReq := &impact.ValidationRequest{
		ClaimedSavings:   250.0, // Claiming unrealistic savings
		OptimizationType: impact.ImpactTypeImageLoading,
		Parameters: map[string]interface{}{
			"image_count": 30,
			"data_size":   15.0,
		},
	}

	validationResult, err := service.ValidateSavings(context.Background(), validationReq)
	if err != nil {
		log.Fatalf("Validation failed: %v", err)
	}

	fmt.Printf("   Claimed savings: %.1f g CO2\n", validationReq.ClaimedSavings)
	fmt.Printf("   Validated savings: %.1f g CO2\n", validationResult.ValidatedSavings)
	fmt.Printf("   Variance: %.1f%%\n", validationResult.Variance)
	fmt.Printf("   Rating: %s\n", validationResult.Rating)
	fmt.Printf("   Valid: %t\n", validationResult.IsValid)
	if len(validationResult.Suggestions) > 0 {
		fmt.Println("   Suggestions:")
		for _, suggestion := range validationResult.Suggestions {
			fmt.Printf("   • %s\n", suggestion)
		}
	}
	fmt.Println()

	// Demo 5: Regional Differences
	fmt.Println("🌍 Regional Carbon Intensity Impact (same page load)")
	basePageReq := &impact.CalculationRequest{
		Type:           impact.ImpactTypePageLoad,
		Duration:       3.0,
		DataSize:       2.5,
		DeviceType:     "laptop",
		ConnectionType: "wifi",
	}

	regions := []string{"FR", "DE", "US", "CN", "IN"}
	regionNames := map[string]string{
		"FR": "France (Nuclear)",
		"DE": "Germany",
		"US": "United States", 
		"CN": "China",
		"IN": "India",
	}

	fmt.Println("   Same page load across different grids:")
	for _, region := range regions {
		req := *basePageReq
		req.Region = region
		result, err := calculator.Calculate(&req)
		if err != nil {
			continue
		}
		fmt.Printf("   %s: %.1f g CO2\n", regionNames[region], result.BaselineEmissions)
	}
	fmt.Println()

	// Demo 6: Impact Report
	fmt.Println("📊 Sample Impact Report")
	period := impact.ReportPeriod{
		Days: 7,
	}
	
	report, err := service.GenerateReport(context.Background(), period)
	if err != nil {
		log.Fatalf("Report generation failed: %v", err)
	}

	fmt.Printf("   Period: Last %d days\n", period.Days)
	fmt.Printf("   Total savings: %.2f kg CO2\n", report.TotalSavings)
	fmt.Printf("   Confidence score: %.0f%%\n", report.ConfidenceScore)
	
	if len(report.EquivalentTo) > 0 {
		fmt.Println("   Equivalent to:")
		for _, equiv := range report.EquivalentTo {
			fmt.Printf("   • %s\n", equiv.Description)
		}
	}
	fmt.Println()

	// Demo 7: Methodology Transparency
	fmt.Println("📖 Methodology & Data Sources")
	fmt.Printf("   Approach: %s\n", report.Methodology)
	fmt.Println("   Data Sources:")
	for _, source := range videoResult.DataSources {
		fmt.Printf("   • %s\n", source)
	}
	if len(videoResult.Warnings) > 0 {
		fmt.Println("   Warnings:")
		for _, warning := range videoResult.Warnings {
			fmt.Printf("   • %s\n", warning)
		}
	}
	fmt.Println()

	fmt.Println("✅ Demo completed successfully!")
	fmt.Println()
	fmt.Println("Key Features Demonstrated:")
	fmt.Println("• Science-based calculations with real emission factors")
	fmt.Println("• Regional carbon intensity variations (85-720 g CO2/kWh)")
	fmt.Println("• Conservative estimates with ±25% confidence intervals")
	fmt.Println("• Rebound effect calculations (10-40% depending on type)")
	fmt.Println("• Anti-greenwashing validation with rating system")
	fmt.Println("• Device energy consumption (often 50%+ of total footprint)")
	fmt.Println("• Methodology transparency with data source attribution")
	fmt.Println("• Component breakdown (device/network/datacenter)")
}