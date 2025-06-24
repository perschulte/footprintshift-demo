package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/perschulte/greenweb-api/pkg/optimization"
	"github.com/perschulte/greenweb-api/service"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelWarn}))

	fmt.Println("🌍 GreenWeb Optimization Demo")
	fmt.Println("=============================")
	fmt.Println()

	electricityMaps := &service.MockElectricityMapsClient{}
	optimizationService := service.NewOptimizationService(electricityMaps, logger)

	scenarios := []struct {
		name      string
		url       string
		intensity float64
	}{
		{"Green E-commerce", "https://shop.example.com", 45.0},
		{"High-Carbon Video", "https://video.example.com", 350.0},
		{"Critical Social Media", "https://social.example.com", 650.0},
	}

	for i, scenario := range scenarios {
		fmt.Printf("%d. %s (%.0fg CO2/kWh)\n", i+1, scenario.name, scenario.intensity)
		electricityMaps.SetMockIntensity(scenario.intensity)

		req := optimization.OptimizationRequest{
			Location: "DE",
			URL:      scenario.url,
		}

		resp, err := optimizationService.GetOptimizationProfile(context.Background(), req)
		if err != nil {
			fmt.Printf("   Error: %v\n", err)
			continue
		}

		profile := resp.Optimization
		fmt.Printf("   Mode: %s, CO2 Savings: %.1fg/hour\n", profile.Mode, profile.EstimatedCO2Savings.TotalSavingsPerHour)
		
		if profile.EstimatedCO2Savings.VideoStreamingSavings > 0 {
			fmt.Printf("   Video: %s (saves %.1fg CO2/hour)\n", 
				profile.HighImpactOptimizations.VideoStreamingOptimization.MaxQuality,
				profile.EstimatedCO2Savings.VideoStreamingSavings)
		}
		fmt.Println()
	}

	fmt.Println("Key Insights:")
	fmt.Println("• Video quality reduction: 12-30g CO2/hour savings")
	fmt.Println("• GPU feature optimization: 8-15g CO2/hour savings")
	fmt.Println("• AI inference deferral: 1.5-3g CO2/session savings")
	fmt.Println("• Real-world focus on high-impact features only")
}