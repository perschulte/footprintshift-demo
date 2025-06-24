package impact

import (
	"testing"
)

func TestCalculator_VideoStreaming(t *testing.T) {
	calculator := NewCalculator()

	req := &CalculationRequest{
		Type:                    ImpactTypeVideoStreaming,
		Duration:               3600, // 1 hour
		VideoQuality:           "1080p",
		DeviceType:             "laptop",
		ConnectionType:         "wifi",
		Region:                 "EU",
		OptimizationLevel:      30,
		IncludeReboundEffects:  true,
	}

	result, err := calculator.Calculate(req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Basic validation
	if result.BaselineEmissions <= 0 {
		t.Errorf("Expected positive baseline emissions, got %f", result.BaselineEmissions)
	}

	if result.OptimizedEmissions >= result.BaselineEmissions {
		t.Errorf("Expected optimized emissions to be less than baseline")
	}

	if result.Savings <= 0 {
		t.Errorf("Expected positive savings, got %f", result.Savings)
	}

	if result.SavingsPercentage <= 0 || result.SavingsPercentage > 100 {
		t.Errorf("Expected savings percentage between 0-100, got %f", result.SavingsPercentage)
	}

	// Test rebound effects
	if result.ReboundEffect <= 0 {
		t.Errorf("Expected positive rebound effect, got %f", result.ReboundEffect)
	}

	if result.NetSavings >= result.Savings {
		t.Errorf("Expected net savings to be less than gross savings due to rebound effect")
	}

	// Test confidence intervals
	if result.ConfidenceInterval != 25.0 {
		t.Errorf("Expected 25%% confidence interval, got %f", result.ConfidenceInterval)
	}

	// Test component breakdown
	components := result.Components
	totalComponents := components.DeviceEmissions + components.NetworkEmissions + components.DataCenterEmissions
	tolerance := 1.0 // Allow 1g CO2 tolerance for floating point precision

	if abs(totalComponents-result.BaselineEmissions) > tolerance {
		t.Errorf("Components don't add up to baseline: %f + %f + %f = %f, expected %f",
			components.DeviceEmissions, components.NetworkEmissions, components.DataCenterEmissions,
			totalComponents, result.BaselineEmissions)
	}

	// For video streaming on laptop, device should be a reasonable portion (but not necessarily majority)
	if components.DevicePercentage < 20 {
		t.Errorf("Expected device emissions to be at least 20%% for laptop video streaming, got %f%%", components.DevicePercentage)
	}
	
	// Network should also be significant for video streaming
	if components.NetworkPercentage < 15 {
		t.Errorf("Expected network emissions to be significant for video streaming, got %f%%", components.NetworkPercentage)
	}
}

func TestCalculator_ImageOptimization(t *testing.T) {
	calculator := NewCalculator()

	req := &CalculationRequest{
		Type:              ImpactTypeImageLoading,
		ImageCount:        50,
		DataSize:          25.0, // 25MB total
		DeviceType:        "smartphone",
		ConnectionType:    "mobile_4g",
		Region:            "US",
		OptimizationLevel: 70, // Aggressive optimization
	}

	result, err := calculator.Calculate(req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// With 70% optimization, should see significant savings
	if result.SavingsPercentage < 40 {
		t.Errorf("Expected significant savings with 70%% optimization, got %f%%", result.SavingsPercentage)
	}

	// Network should be significant component for mobile
	if result.Components.NetworkPercentage < 20 {
		t.Errorf("Expected significant network component for mobile 4G, got %f%%", result.Components.NetworkPercentage)
	}
}

func TestCalculator_AIInference(t *testing.T) {
	calculator := NewCalculator()

	req := &CalculationRequest{
		Type:              ImpactTypeAIInference,
		Duration:          10.0, // 10 seconds (5 inferences)
		DeviceType:        "laptop",
		ConnectionType:    "wifi",
		Region:            "US",
		OptimizationLevel: 50,
	}

	result, err := calculator.Calculate(req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// AI inference should have high data center component
	if result.Components.DataCenterPercentage < 70 {
		t.Errorf("Expected AI inference to be data center heavy, got %f%%", result.Components.DataCenterPercentage)
	}
}

func TestCalculator_JavaScript(t *testing.T) {
	calculator := NewCalculator()

	req := &CalculationRequest{
		Type:              ImpactTypeJavaScript,
		Duration:          60.0, // 1 minute execution
		DataSize:          2.0,  // 2MB bundle
		DeviceType:        "desktop",
		ConnectionType:    "ethernet",
		Region:            "EU",
		OptimizationLevel: 40,
	}

	result, err := calculator.Calculate(req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// JavaScript should be device-heavy (client-side execution)
	if result.Components.DevicePercentage < 80 {
		t.Errorf("Expected JavaScript to be device-heavy, got %f%%", result.Components.DevicePercentage)
	}

	// Minimal data center emissions for client-side JS
	if result.Components.DataCenterPercentage > 5 {
		t.Errorf("Expected minimal data center emissions for client-side JS, got %f%%", result.Components.DataCenterPercentage)
	}
}

func TestCalculator_ValidationErrors(t *testing.T) {
	calculator := NewCalculator()

	// Test missing type
	req := &CalculationRequest{}
	_, err := calculator.Calculate(req)
	if err == nil {
		t.Error("Expected error for missing type")
	}

	// Test video streaming without duration
	req = &CalculationRequest{
		Type: ImpactTypeVideoStreaming,
	}
	_, err = calculator.Calculate(req)
	if err == nil {
		t.Error("Expected error for video streaming without duration")
	}

	// Test image loading without count or size
	req = &CalculationRequest{
		Type: ImpactTypeImageLoading,
	}
	_, err = calculator.Calculate(req)
	if err == nil {
		t.Error("Expected error for image loading without count or size")
	}
}

func TestCalculator_DefaultValues(t *testing.T) {
	calculator := NewCalculator()

	req := &CalculationRequest{
		Type:     ImpactTypePageLoad,
		Duration: 3.0,
		DataSize: 2.0,
		// No device type, connection, or region specified
	}

	result, err := calculator.Calculate(req)
	if err != nil {
		t.Fatalf("Expected no error with defaults, got %v", err)
	}

	if result.BaselineEmissions <= 0 {
		t.Errorf("Expected positive emissions with default values")
	}
}

func TestCalculator_RegionalVariations(t *testing.T) {
	calculator := NewCalculator()

	// Test same calculation in different regions
	baseReq := &CalculationRequest{
		Type:           ImpactTypePageLoad,
		Duration:       3.0,
		DataSize:       2.0,
		DeviceType:     "laptop",
		ConnectionType: "wifi",
	}

	// Test EU (lower carbon intensity)
	euReq := *baseReq
	euReq.Region = "EU"
	euResult, err := calculator.Calculate(&euReq)
	if err != nil {
		t.Fatalf("EU calculation failed: %v", err)
	}

	// Test China (higher carbon intensity)
	cnReq := *baseReq
	cnReq.Region = "CN"
	cnResult, err := calculator.Calculate(&cnReq)
	if err != nil {
		t.Fatalf("China calculation failed: %v", err)
	}

	// China should have higher emissions due to higher grid carbon intensity
	if cnResult.BaselineEmissions <= euResult.BaselineEmissions {
		t.Errorf("Expected higher emissions in China (%.2f) than EU (%.2f)",
			cnResult.BaselineEmissions, euResult.BaselineEmissions)
	}
}

func TestCalculator_OptimizationLevels(t *testing.T) {
	calculator := NewCalculator()

	baseReq := &CalculationRequest{
		Type:           ImpactTypeImageLoading,
		ImageCount:     20,
		DataSize:       10.0,
		DeviceType:     "laptop",
		ConnectionType: "wifi",
		Region:         "EU",
	}

	// Test no optimization
	noOptReq := *baseReq
	noOptReq.OptimizationLevel = 0
	noOptResult, err := calculator.Calculate(&noOptReq)
	if err != nil {
		t.Fatalf("No optimization calculation failed: %v", err)
	}

	// Test high optimization
	highOptReq := *baseReq
	highOptReq.OptimizationLevel = 80
	highOptResult, err := calculator.Calculate(&highOptReq)
	if err != nil {
		t.Fatalf("High optimization calculation failed: %v", err)
	}

	// High optimization should show more savings
	if highOptResult.SavingsPercentage <= noOptResult.SavingsPercentage {
		t.Errorf("Expected higher optimization to show more savings: %f%% vs %f%%",
			highOptResult.SavingsPercentage, noOptResult.SavingsPercentage)
	}
}

// Helper function for absolute value
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}