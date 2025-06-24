package impact

import (
	"fmt"
	"math"
	"time"
)

// Calculator provides science-based CO2 impact calculations
type Calculator struct {
	factors EmissionFactors
}

// NewCalculator creates a new impact calculator with default emission factors
func NewCalculator() *Calculator {
	return &Calculator{
		factors: DefaultEmissionFactors,
	}
}

// NewCalculatorWithFactors creates a calculator with custom emission factors
func NewCalculatorWithFactors(factors EmissionFactors) *Calculator {
	return &Calculator{
		factors: factors,
	}
}

// Calculate performs impact calculation based on the request
func (c *Calculator) Calculate(req *CalculationRequest) (*ImpactResult, error) {
	if err := c.validateRequest(req); err != nil {
		return nil, err
	}

	// Set defaults
	c.setDefaults(req)

	var result *ImpactResult
	var err error

	switch req.Type {
	case ImpactTypeVideoStreaming:
		result, err = c.calculateVideoStreaming(req)
	case ImpactTypeImageLoading:
		result, err = c.calculateImageLoading(req)
	case ImpactTypeJavaScript:
		result, err = c.calculateJavaScript(req)
	case ImpactTypeAIInference:
		result, err = c.calculateAIInference(req)
	case ImpactTypePageLoad:
		result, err = c.calculatePageLoad(req)
	case ImpactTypeDataTransfer:
		result, err = c.calculateDataTransfer(req)
	default:
		return nil, fmt.Errorf("unsupported impact type: %s", req.Type)
	}

	if err != nil {
		return nil, err
	}

	// Apply confidence intervals (conservative approach)
	c.applyConfidenceIntervals(result)

	// Calculate rebound effects if requested
	if req.IncludeReboundEffects {
		c.calculateReboundEffects(result, req)
	}

	// Add methodology and warnings
	c.addMethodologyInfo(result, req)

	result.CalculatedAt = time.Now()

	return result, nil
}

// calculateVideoStreaming calculates impact of video streaming
// Based on: IEA (2020), Carbon Trust (2021), Shift Project (2023 revised)
func (c *Calculator) calculateVideoStreaming(req *CalculationRequest) (*ImpactResult, error) {
	// Video bitrates in Mbps
	bitrates := map[string]float64{
		"360p":  1.0,
		"480p":  2.5,
		"720p":  5.0,
		"1080p": 8.0,
		"4k":    25.0,
	}

	bitrate := bitrates[req.VideoQuality]
	hours := req.Duration / 3600.0

	// Data transferred in GB
	dataGB := (bitrate * req.Duration) / 8.0 / 1000.0

	// Device consumption
	devicePower := c.factors.DeviceConsumption[req.DeviceType]["video_streaming"]
	deviceEnergy := devicePower * hours // Wh
	gridIntensity := c.factors.GridCarbonIntensity[req.Region]
	deviceEmissions := (deviceEnergy / 1000.0) * gridIntensity // g CO2

	// Network emissions
	networkEmissions := dataGB * c.factors.NetworkTransmission[req.ConnectionType]

	// Data center emissions (including CDN)
	dcPower := 0.012 // kWh/GB for video streaming (includes CDN)
	dcPUE := c.factors.DataCenterPUE["global"]
	dcEnergy := dataGB * dcPower * dcPUE
	dcEmissions := dcEnergy * gridIntensity

	// Store baseline components before optimization
	baselineDeviceEmissions := deviceEmissions
	baselineNetworkEmissions := networkEmissions
	baselineDcEmissions := dcEmissions
	
	// Total baseline
	baseline := deviceEmissions + networkEmissions + dcEmissions

	// Calculate optimized emissions based on optimization level
	optimizationFactor := 1.0 - (req.OptimizationLevel / 100.0)
	
	// Video optimization can reduce bitrate significantly
	optimizedNetworkEmissions := networkEmissions
	optimizedDcEmissions := dcEmissions
	
	if req.OptimizationLevel > 0 {
		// Adaptive bitrate can reduce data by up to 40%
		dataReduction := math.Min(req.OptimizationLevel*0.4/100.0, 0.4)
		optimizedData := dataGB * (1 - dataReduction)
		
		// Recalculate with optimized data
		optimizedNetworkEmissions = optimizedData * c.factors.NetworkTransmission[req.ConnectionType]
		dcEnergy = optimizedData * dcPower * dcPUE
		optimizedDcEmissions = dcEnergy * gridIntensity
	}

	optimized := (deviceEmissions + optimizedNetworkEmissions + optimizedDcEmissions) * optimizationFactor

	return &ImpactResult{
		BaselineEmissions:  baseline,
		OptimizedEmissions: optimized,
		Savings:            baseline - optimized,
		SavingsPercentage:  ((baseline - optimized) / baseline) * 100,
		Components: EmissionComponents{
			DeviceEmissions:      baselineDeviceEmissions,
			NetworkEmissions:     baselineNetworkEmissions,
			DataCenterEmissions:  baselineDcEmissions,
			DevicePercentage:     (baselineDeviceEmissions / baseline) * 100,
			NetworkPercentage:    (baselineNetworkEmissions / baseline) * 100,
			DataCenterPercentage: (baselineDcEmissions / baseline) * 100,
		},
		NetSavings: baseline - optimized,
	}, nil
}

// calculateImageLoading calculates impact of image loading and optimization
func (c *Calculator) calculateImageLoading(req *CalculationRequest) (*ImpactResult, error) {
	// Average image sizes by optimization level
	avgImageSize := 0.5 // MB per image (unoptimized)
	
	if req.DataSize > 0 {
		avgImageSize = req.DataSize / float64(max(req.ImageCount, 1))
	}

	totalDataMB := avgImageSize * float64(req.ImageCount)
	dataGB := totalDataMB / 1000.0

	// Device consumption for image loading
	loadTime := totalDataMB * 2.0 // Rough estimate: 2 seconds per MB
	devicePower := c.factors.DeviceConsumption[req.DeviceType]["image_loading"]
	deviceEnergy := (devicePower * loadTime) / 3600.0 // Wh
	gridIntensity := c.factors.GridCarbonIntensity[req.Region]
	deviceEmissions := (deviceEnergy / 1000.0) * gridIntensity

	// Network emissions
	networkEmissions := dataGB * c.factors.NetworkTransmission[req.ConnectionType]

	// Data center emissions (serving images)
	dcPower := 0.008 // kWh/GB for static content
	dcPUE := c.factors.DataCenterPUE["global"]
	dcEnergy := dataGB * dcPower * dcPUE
	dcEmissions := dcEnergy * gridIntensity

	baseline := deviceEmissions + networkEmissions + dcEmissions

	// Image optimization can achieve 60-80% size reduction
	optimizationFactor := 1.0
	if req.OptimizationLevel > 0 {
		// WebP/AVIF can reduce size by up to 70%
		sizeReduction := math.Min(req.OptimizationLevel*0.7/100.0, 0.7)
		optimizationFactor = 1.0 - sizeReduction
	}

	optimized := baseline * optimizationFactor

	return &ImpactResult{
		BaselineEmissions:  baseline,
		OptimizedEmissions: optimized,
		Savings:            baseline - optimized,
		SavingsPercentage:  ((baseline - optimized) / baseline) * 100,
		Components: EmissionComponents{
			DeviceEmissions:      deviceEmissions,
			NetworkEmissions:     networkEmissions,
			DataCenterEmissions:  dcEmissions,
			DevicePercentage:     (deviceEmissions / baseline) * 100,
			NetworkPercentage:    (networkEmissions / baseline) * 100,
			DataCenterPercentage: (dcEmissions / baseline) * 100,
		},
		NetSavings: baseline - optimized,
	}, nil
}

// calculateJavaScript calculates impact of JavaScript execution
func (c *Calculator) calculateJavaScript(req *CalculationRequest) (*ImpactResult, error) {
	hours := req.Duration / 3600.0

	// Device consumption for JS execution
	devicePower := c.factors.DeviceConsumption[req.DeviceType]["javascript_heavy"]
	deviceEnergy := devicePower * hours // Wh
	gridIntensity := c.factors.GridCarbonIntensity[req.Region]
	deviceEmissions := (deviceEnergy / 1000.0) * gridIntensity

	// JS bundle size impact (download)
	bundleSize := req.DataSize / 1000.0 // Convert MB to GB
	if bundleSize == 0 {
		bundleSize = 0.002 // Default 2MB bundle
	}
	networkEmissions := bundleSize * c.factors.NetworkTransmission[req.ConnectionType]

	// No significant data center emissions for client-side JS
	dcEmissions := 0.0

	baseline := deviceEmissions + networkEmissions + dcEmissions

	// JS optimization can reduce execution time and bundle size
	optimizationFactor := 1.0
	if req.OptimizationLevel > 0 {
		// Code splitting, tree shaking can reduce by 40-60%
		reduction := math.Min(req.OptimizationLevel*0.5/100.0, 0.5)
		optimizationFactor = 1.0 - reduction
	}

	optimized := baseline * optimizationFactor

	return &ImpactResult{
		BaselineEmissions:  baseline,
		OptimizedEmissions: optimized,
		Savings:            baseline - optimized,
		SavingsPercentage:  ((baseline - optimized) / baseline) * 100,
		Components: EmissionComponents{
			DeviceEmissions:      deviceEmissions,
			NetworkEmissions:     networkEmissions,
			DataCenterEmissions:  dcEmissions,
			DevicePercentage:     (deviceEmissions / baseline) * 100,
			NetworkPercentage:    (networkEmissions / baseline) * 100,
			DataCenterPercentage: 0,
		},
		NetSavings: baseline - optimized,
	}, nil
}

// calculateAIInference calculates impact of AI inference
// Based on: Strubell et al. (2019), Patterson et al. (2021)
func (c *Calculator) calculateAIInference(req *CalculationRequest) (*ImpactResult, error) {
	// AI inference energy consumption
	inferenceEnergy := 0.006 // kWh per inference (GPT-3 scale model)
	if req.Duration > 0 {
		// Multiple inferences based on duration
		inferences := req.Duration / 2.0 // Assume 2 seconds per inference
		inferenceEnergy *= inferences
	}

	// Data center emissions (GPU intensive)
	gridIntensity := c.factors.GridCarbonIntensity["US"] // Most AI runs in US data centers
	dcPUE := 1.1                                          // Modern AI data centers are efficient
	dcEmissions := inferenceEnergy * gridIntensity * dcPUE

	// Network emissions (API calls)
	apiDataGB := 0.001 // ~1MB per inference
	networkEmissions := apiDataGB * c.factors.NetworkTransmission[req.ConnectionType]

	// Device emissions (minimal for API calls)
	deviceEmissions := 0.5 // g CO2 per inference

	baseline := deviceEmissions + networkEmissions + dcEmissions

	// AI optimization through caching, batching
	optimizationFactor := 1.0
	if req.OptimizationLevel > 0 {
		// Caching can reduce by up to 80% for repeated queries
		reduction := math.Min(req.OptimizationLevel*0.8/100.0, 0.8)
		optimizationFactor = 1.0 - reduction
	}

	optimized := baseline * optimizationFactor

	return &ImpactResult{
		BaselineEmissions:  baseline,
		OptimizedEmissions: optimized,
		Savings:            baseline - optimized,
		SavingsPercentage:  ((baseline - optimized) / baseline) * 100,
		Components: EmissionComponents{
			DeviceEmissions:      deviceEmissions,
			NetworkEmissions:     networkEmissions,
			DataCenterEmissions:  dcEmissions,
			DevicePercentage:     (deviceEmissions / baseline) * 100,
			NetworkPercentage:    (networkEmissions / baseline) * 100,
			DataCenterPercentage: (dcEmissions / baseline) * 100,
		},
		NetSavings: baseline - optimized,
	}, nil
}

// calculatePageLoad calculates full page load impact
func (c *Calculator) calculatePageLoad(req *CalculationRequest) (*ImpactResult, error) {
	// Average page size: 2.2MB (HTTP Archive 2023)
	pageSize := req.DataSize
	if pageSize == 0 {
		pageSize = 2.2
	}
	dataGB := pageSize / 1000.0

	// Page load time
	loadTime := req.Duration
	if loadTime == 0 {
		loadTime = 5.0 // Average 5 seconds
	}

	// Device consumption
	devicePower := c.factors.DeviceConsumption[req.DeviceType]["browsing"]
	deviceEnergy := (devicePower * loadTime) / 3600.0 // Wh
	gridIntensity := c.factors.GridCarbonIntensity[req.Region]
	deviceEmissions := (deviceEnergy / 1000.0) * gridIntensity

	// Network emissions
	networkEmissions := dataGB * c.factors.NetworkTransmission[req.ConnectionType]

	// Data center emissions
	dcPower := 0.01 // kWh/GB for web pages
	dcPUE := c.factors.DataCenterPUE["global"]
	dcEnergy := dataGB * dcPower * dcPUE
	dcEmissions := dcEnergy * gridIntensity

	baseline := deviceEmissions + networkEmissions + dcEmissions

	// Page optimization can reduce size by 50-70%
	optimizationFactor := 1.0
	if req.OptimizationLevel > 0 {
		reduction := math.Min(req.OptimizationLevel*0.6/100.0, 0.6)
		optimizationFactor = 1.0 - reduction
	}

	optimized := baseline * optimizationFactor

	return &ImpactResult{
		BaselineEmissions:  baseline,
		OptimizedEmissions: optimized,
		Savings:            baseline - optimized,
		SavingsPercentage:  ((baseline - optimized) / baseline) * 100,
		Components: EmissionComponents{
			DeviceEmissions:      deviceEmissions,
			NetworkEmissions:     networkEmissions,
			DataCenterEmissions:  dcEmissions,
			DevicePercentage:     (deviceEmissions / baseline) * 100,
			NetworkPercentage:    (networkEmissions / baseline) * 100,
			DataCenterPercentage: (dcEmissions / baseline) * 100,
		},
		NetSavings: baseline - optimized,
	}, nil
}

// calculateDataTransfer calculates generic data transfer impact
func (c *Calculator) calculateDataTransfer(req *CalculationRequest) (*ImpactResult, error) {
	dataGB := req.DataSize / 1000.0

	// Network emissions
	networkEmissions := dataGB * c.factors.NetworkTransmission[req.ConnectionType]

	// Minimal device emissions for data transfer
	transferTime := dataGB * 8.0 // seconds (assuming 1 Gbps)
	devicePower := c.factors.DeviceConsumption[req.DeviceType]["idle"]
	deviceEnergy := (devicePower * transferTime) / 3600.0
	gridIntensity := c.factors.GridCarbonIntensity[req.Region]
	deviceEmissions := (deviceEnergy / 1000.0) * gridIntensity

	// Data center emissions
	dcPower := 0.008 // kWh/GB
	dcPUE := c.factors.DataCenterPUE["global"]
	dcEnergy := dataGB * dcPower * dcPUE
	dcEmissions := dcEnergy * gridIntensity

	baseline := deviceEmissions + networkEmissions + dcEmissions

	// Data compression can reduce by 20-80%
	optimizationFactor := 1.0
	if req.OptimizationLevel > 0 {
		reduction := math.Min(req.OptimizationLevel*0.5/100.0, 0.5)
		optimizationFactor = 1.0 - reduction
	}

	optimized := baseline * optimizationFactor

	return &ImpactResult{
		BaselineEmissions:  baseline,
		OptimizedEmissions: optimized,
		Savings:            baseline - optimized,
		SavingsPercentage:  ((baseline - optimized) / baseline) * 100,
		Components: EmissionComponents{
			DeviceEmissions:      deviceEmissions,
			NetworkEmissions:     networkEmissions,
			DataCenterEmissions:  dcEmissions,
			DevicePercentage:     (deviceEmissions / baseline) * 100,
			NetworkPercentage:    (networkEmissions / baseline) * 100,
			DataCenterPercentage: (dcEmissions / baseline) * 100,
		},
		NetSavings: baseline - optimized,
	}, nil
}

// applyConfidenceIntervals adds conservative confidence intervals
func (c *Calculator) applyConfidenceIntervals(result *ImpactResult) {
	// Conservative approach: ±25% uncertainty
	result.ConfidenceInterval = 25.0
	
	// Calculate bounds
	result.LowerBound = result.OptimizedEmissions * 0.75
	result.UpperBound = result.OptimizedEmissions * 1.25
}

// calculateReboundEffects estimates additional consumption from efficiency gains
// Based on: Sorrell (2007), Greening et al. (2000)
func (c *Calculator) calculateReboundEffects(result *ImpactResult, req *CalculationRequest) {
	// Rebound effect varies by optimization type
	reboundFactors := map[ImpactType]float64{
		ImpactTypeVideoStreaming: 0.3,  // 30% - people watch more when quality improves
		ImpactTypeImageLoading:   0.1,  // 10% - minimal rebound
		ImpactTypeJavaScript:     0.05, // 5% - technical optimization
		ImpactTypeAIInference:    0.4,  // 40% - cheaper AI leads to more use
		ImpactTypePageLoad:       0.2,  // 20% - faster pages increase browsing
		ImpactTypeDataTransfer:   0.15, // 15% - general data transfer
	}

	reboundFactor := reboundFactors[req.Type]
	result.ReboundEffect = result.Savings * reboundFactor
	result.NetSavings = result.Savings - result.ReboundEffect

	// Add warning about rebound effects
	result.Warnings = append(result.Warnings,
		fmt.Sprintf("Rebound effect estimated at %.0f%% - improved efficiency may lead to increased consumption",
			reboundFactor*100))
}

// addMethodologyInfo adds methodology explanation and data sources
func (c *Calculator) addMethodologyInfo(result *ImpactResult, req *CalculationRequest) {
	result.Methodology = "Conservative calculation based on device energy consumption, network transmission, and data center operations. Includes ±25% confidence interval."
	
	result.DataSources = []string{
		"IEA (2023) - Electricity grid carbon intensity",
		"Carbon Trust (2021) - Digital service emissions",
		"EPA (2023) - Data center PUE values",
		"HTTP Archive (2023) - Web page statistics",
		"Shift Project (2023) - Video streaming analysis",
	}

	// Add specific warnings
	if req.Type == ImpactTypeVideoStreaming {
		result.Warnings = append(result.Warnings,
			"Video streaming estimates assume average CDN efficiency and may vary by provider")
	}

	if req.Type == ImpactTypeAIInference {
		result.Warnings = append(result.Warnings,
			"AI inference estimates based on GPT-3 scale models; actual emissions vary by model size")
	}

	// Device consumption warning
	if result.Components.DevicePercentage > 50 {
		result.Warnings = append(result.Warnings,
			fmt.Sprintf("Device emissions account for %.0f%% of total - user device efficiency is critical",
				result.Components.DevicePercentage))
	}
}

// validateRequest validates the calculation request
func (c *Calculator) validateRequest(req *CalculationRequest) error {
	if req.Type == "" {
		return fmt.Errorf("impact type is required")
	}

	switch req.Type {
	case ImpactTypeVideoStreaming:
		if req.Duration <= 0 {
			return fmt.Errorf("duration is required for video streaming calculations")
		}
		if req.VideoQuality == "" {
			req.VideoQuality = "720p" // Default
		}
	case ImpactTypeImageLoading:
		if req.ImageCount <= 0 && req.DataSize <= 0 {
			return fmt.Errorf("either image count or data size is required for image calculations")
		}
	case ImpactTypeJavaScript:
		if req.Duration <= 0 && req.DataSize <= 0 {
			return fmt.Errorf("either duration or data size is required for JavaScript calculations")
		}
	case ImpactTypeDataTransfer:
		if req.DataSize <= 0 {
			return fmt.Errorf("data size is required for data transfer calculations")
		}
	}

	return nil
}

// setDefaults sets default values for missing fields
func (c *Calculator) setDefaults(req *CalculationRequest) {
	if req.DeviceType == "" {
		req.DeviceType = "laptop"
	}
	if req.ConnectionType == "" {
		req.ConnectionType = "wifi"
	}
	if req.Region == "" {
		req.Region = "global"
	}
}

// Helper function for max
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}