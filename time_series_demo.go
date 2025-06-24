package main

import (
	"fmt"
	"log"
	"math"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

// Enhanced carbon intensity with 15-minute precision
type TimeSeriesCarbonIntensity struct {
	Location           string    `json:"location"`
	CarbonIntensity    float64   `json:"carbon_intensity"`
	RenewablePercent   float64   `json:"renewable_percentage"`
	Mode               string    `json:"mode"`
	Recommendation     string    `json:"recommendation"`
	NextGreenWindow    time.Time `json:"next_green_window"`
	Timestamp          time.Time `json:"timestamp"`
	LocalPercentile    float64   `json:"local_percentile"`
	DailyRank          string    `json:"daily_rank"`
	RelativeMode       string    `json:"relative_mode"`
	TrendDirection     string    `json:"trend_direction"`
	Hour               int       `json:"hour"`
	Minute             int       `json:"minute"`
	TimeIndex          int       `json:"time_index"` // 0-95 for 96 data points
}

type TimeSeriesOptimization struct {
	Mode                   string   `json:"mode"`
	DisableFeatures        []string `json:"disable_features"`
	ImageQuality           string   `json:"image_quality"`
	VideoQuality           string   `json:"video_quality"`
	DeferAnalytics         bool     `json:"defer_analytics"`
	EcoDiscount            int      `json:"eco_discount"`
	ShowGreenBanner        bool     `json:"show_green_banner"`
	CachingStrategy        string   `json:"caching_strategy"`
	VideoSavingsPerHour    float64  `json:"video_co2_savings_per_hour_g"`
	AISavingsPerSession    float64  `json:"ai_co2_savings_per_session_g"`
	GPUSavingsPerHour      float64  `json:"gpu_co2_savings_per_hour_g"`
	MaxVideoBitrate        int      `json:"max_video_bitrate_kbps"`
	AIDeferred             bool     `json:"ai_deferred_to_green_window"`
	GPUFeaturesDisabled    bool     `json:"gpu_features_disabled"`
	Hour                   int      `json:"hour"`
}

// Germany-specific realistic carbon intensity simulation with 15-minute precision
func getGermanyCarbonIntensityForTimeIndex(timeIndex int) TimeSeriesCarbonIntensity {
	// Convert time index to hour and minute (0-95 = 00:00 to 23:45)
	hour := timeIndex / 4
	minute := (timeIndex % 4) * 15
	// More precise time for smooth transitions
	timeFloat := float64(hour) + float64(minute)/60.0
	
	baseIntensity := 295.0 // Germany average g CO₂/kWh
	
	// Solar contribution with 15-minute granularity (peak at 13:00)
	solarFactor := math.Max(0, math.Sin(math.Pi*(timeFloat-6)/12))
	if timeFloat < 6 || timeFloat > 18 {
		solarFactor = 0 // No solar at night
	}
	
	// Wind contribution with micro-variations every 15 min
	windFactor := 0.6 + 0.4*math.Sin(math.Pi*(timeFloat+6)/12)
	// Add 15-minute wind gusts
	windFactor += 0.1 * math.Sin(math.Pi*timeFloat*4) // 4 cycles per hour
	
	// Demand curve with gradual transitions
	demandFactor := 0.7 + 0.3*math.Sin(math.Pi*(timeFloat-3)/12)
	if timeFloat >= 8 && timeFloat <= 18 {
		demandFactor += 0.2 // Business hours boost
	}
	
	// Add 15-minute demand variations (AC, production cycles)
	demandFactor += 0.05 * math.Sin(math.Pi*timeFloat*8) // 8 cycles per hour
	
	// Coal/gas backup (inversely related to renewables)
	renewablePercent := math.Min(80, math.Max(15, (solarFactor*35 + windFactor*45)))
	fossilBackup := (100 - renewablePercent) / 100
	
	// Final intensity calculation with smoother transitions
	intensity := baseIntensity * (0.3 + 0.7*fossilBackup*demandFactor)
	
	// Add realistic 15-minute variations
	microVariation := 5 * math.Sin(math.Pi*timeFloat*2) // 2 cycles per hour
	hourlyVariation := 15 * math.Sin(math.Pi*timeFloat/6) // ±15g variation
	intensity += microVariation + hourlyVariation
	
	// Calculate percentile based on all 96 data points
	dailyIntensities := make([]float64, 96)
	for i := 0; i < 96; i++ {
		dailyIntensities[i] = getIntensityForTimeIndex(i, baseIntensity)
	}
	
	// Calculate percentile
	lowerCount := 0
	for _, dayIntensity := range dailyIntensities {
		if dayIntensity < intensity {
			lowerCount++
		}
	}
	percentile := float64(lowerCount) / 96.0 * 100
	
	var mode, recommendation, dailyRank, relativeMode string
	
	if percentile <= 30 {
		mode = "green"
		recommendation = "optimal"
		dailyRank = "cleanest hours of the day"
		relativeMode = "clean"
	} else if percentile <= 70 {
		mode = "yellow"
		recommendation = "reduce"
		dailyRank = "average hours of the day"
		relativeMode = "normal"
	} else {
		mode = "red"
		recommendation = "defer"
		dailyRank = "dirtiest hours of the day"
		relativeMode = "dirty"
	}
	
	// Calculate next green window with 15-minute precision
	nextGreenIndex := (timeIndex + 1) % 96
	for i := 1; i < 96; i++ {
		checkIndex := (timeIndex + i) % 96
		checkIntensity := getIntensityForTimeIndex(checkIndex, baseIntensity)
		if checkIntensity < intensity*0.8 { // 20% lower than current
			nextGreenIndex = checkIndex
			break
		}
	}
	
	nextGreenHour := nextGreenIndex / 4
	nextGreenMinute := (nextGreenIndex % 4) * 15
	nextGreenWindow := time.Date(2024, 1, 1, nextGreenHour, nextGreenMinute, 0, 0, time.UTC)
	
	return TimeSeriesCarbonIntensity{
		Location:         "Germany",
		CarbonIntensity:  math.Round(intensity*10) / 10,
		RenewablePercent: math.Round(renewablePercent*10) / 10,
		Mode:             mode,
		Recommendation:   recommendation,
		NextGreenWindow:  nextGreenWindow,
		Timestamp:        time.Date(2024, 1, 1, hour, minute, 0, 0, time.UTC),
		LocalPercentile:  math.Round(percentile*10) / 10,
		DailyRank:        dailyRank,
		RelativeMode:     relativeMode,
		TrendDirection:   getTrendDirection(timeFloat),
		Hour:             hour,
		Minute:           minute,
		TimeIndex:        timeIndex,
	}
}

func getIntensityForTimeIndex(timeIndex int, baseIntensity float64) float64 {
	hour := timeIndex / 4
	minute := (timeIndex % 4) * 15
	timeFloat := float64(hour) + float64(minute)/60.0
	
	solarFactor := math.Max(0, math.Sin(math.Pi*(timeFloat-6)/12))
	if timeFloat < 6 || timeFloat > 18 {
		solarFactor = 0
	}
	
	windFactor := 0.6 + 0.4*math.Sin(math.Pi*(timeFloat+6)/12)
	windFactor += 0.1 * math.Sin(math.Pi*timeFloat*4)
	
	demandFactor := 0.7 + 0.3*math.Sin(math.Pi*(timeFloat-3)/12)
	if timeFloat >= 8 && timeFloat <= 18 {
		demandFactor += 0.2
	}
	demandFactor += 0.05 * math.Sin(math.Pi*timeFloat*8)
	
	renewablePercent := math.Min(80, math.Max(15, (solarFactor*35 + windFactor*45)))
	fossilBackup := (100 - renewablePercent) / 100
	
	intensity := baseIntensity * (0.3 + 0.7*fossilBackup*demandFactor)
	microVariation := 5 * math.Sin(math.Pi*timeFloat*2)
	hourlyVariation := 15 * math.Sin(math.Pi*timeFloat/6)
	
	return intensity + microVariation + hourlyVariation
}

func getTrendDirection(timeFloat float64) string {
	// Precise trend based on time with smooth transitions
	if timeFloat >= 6 && timeFloat <= 12 {
		return "worsening" // Morning ramp-up
	} else if timeFloat >= 13 && timeFloat <= 18 {
		return "improving" // Afternoon solar
	} else if timeFloat >= 19 && timeFloat <= 23 {
		return "improving" // Evening wind
	} else {
		return "stable" // Night hours
	}
}

func getOptimizationForHour(intensity float64, hour int) TimeSeriesOptimization {
	opt := TimeSeriesOptimization{Hour: hour}
	
	if intensity < 200 {
		// Green mode
		opt.Mode = "full"
		opt.DisableFeatures = []string{}
		opt.ImageQuality = "high"
		opt.VideoQuality = "4k"
		opt.DeferAnalytics = false
		opt.EcoDiscount = 5
		opt.ShowGreenBanner = true
		opt.CachingStrategy = "normal"
		opt.VideoSavingsPerHour = 0
		opt.AISavingsPerSession = 0
		opt.GPUSavingsPerHour = 0
		opt.MaxVideoBitrate = 25000
		opt.AIDeferred = false
		opt.GPUFeaturesDisabled = false
	} else if intensity < 350 {
		// Yellow mode
		opt.Mode = "normal"
		opt.DisableFeatures = []string{"video_autoplay", "3d_models"}
		opt.ImageQuality = "medium"
		opt.VideoQuality = "1080p"
		opt.VideoSavingsPerHour = 12
		opt.AISavingsPerSession = 1.5
		opt.GPUSavingsPerHour = 8
		opt.MaxVideoBitrate = 8000
		opt.AIDeferred = false
		opt.GPUFeaturesDisabled = false
	} else {
		// Red mode
		opt.Mode = "eco"
		opt.DisableFeatures = []string{"video_autoplay", "3d_models", "animations", "ai_features", "webgl"}
		opt.ImageQuality = "low"
		opt.VideoQuality = "720p"
		opt.DeferAnalytics = true
		opt.VideoSavingsPerHour = 24
		opt.AISavingsPerSession = 3
		opt.GPUSavingsPerHour = 15
		opt.MaxVideoBitrate = 5000
		opt.AIDeferred = true
		opt.GPUFeaturesDisabled = true
	}
	
	return opt
}

func main() {
	godotenv.Load()
	
	r := gin.Default()
	
	// CORS middleware
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// Root redirect to demo
	r.GET("/", func(c *gin.Context) {
		c.Redirect(301, "/demo")
	})

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":    "healthy",
			"service":   "footprintshift-api",
			"version":   "0.3.0",
			"features": []string{
				"24h_time_series_simulation",
				"germany_realistic_patterns",
				"interactive_time_slider",
				"renewable_energy_modeling",
			},
			"timestamp": time.Now(),
		})
	})

	// Time series endpoint - specific time index
	r.GET("/api/v1/carbon-intensity/:timeindex", func(c *gin.Context) {
		timeIndex := 0
		if t := c.Param("timeindex"); t != "" {
			if parsed, err := time.Parse("15", t); err == nil {
				timeIndex = parsed.Hour() * 4 // Convert hour to time index
			}
		}
		
		data := getGermanyCarbonIntensityForTimeIndex(timeIndex)
		c.JSON(200, data)
	})

	// Time series endpoint - full day with 15-minute precision
	r.GET("/api/v1/carbon-intensity/timeseries", func(c *gin.Context) {
		location := c.DefaultQuery("location", "Germany")
		
		var timeSeries []TimeSeriesCarbonIntensity
		for timeIndex := 0; timeIndex < 96; timeIndex++ { // 96 = 24h * 4 (15-min intervals)
			data := getGermanyCarbonIntensityForTimeIndex(timeIndex)
			timeSeries = append(timeSeries, data)
		}
		
		c.JSON(200, gin.H{
			"location":    location,
			"date":        "2024-01-01", // Example date
			"timezone":    "CET",
			"timeseries":  timeSeries,
			"resolution":  "15_minutes", // 96 data points
			"metadata": gin.H{
				"avg_intensity":      calculateAverage(timeSeries),
				"min_intensity":      findMin(timeSeries),
				"max_intensity":      findMax(timeSeries),
				"green_periods":      countGreenPeriods(timeSeries),
				"peak_renewable":     findPeakRenewable(timeSeries),
				"optimal_windows":    findOptimalWindows(timeSeries),
			},
		})
	})

	// Optimization for specific hour (kept for compatibility)
	r.GET("/api/v1/optimization/:hour", func(c *gin.Context) {
		hour := 0
		if h := c.Param("hour"); h != "" {
			if parsed, err := time.Parse("15", h); err == nil {
				hour = parsed.Hour()
			}
		}
		
		// Use middle of the hour (e.g., 12:30 for hour 12)
		timeIndex := hour*4 + 2 // +2 = 30 minutes
		carbonData := getGermanyCarbonIntensityForTimeIndex(timeIndex)
		optimization := getOptimizationForHour(carbonData.CarbonIntensity, hour)
		
		c.JSON(200, gin.H{
			"carbon_intensity": carbonData,
			"optimization":     optimization,
			"hour":            hour,
			"methodology":     "germany_realistic_simulation_15min",
		})
	})

	// Interest tracking endpoint
	r.POST("/api/v1/register-interest", func(c *gin.Context) {
		var request struct {
			Email       string `json:"email" binding:"required"`
			CompanyRole string `json:"company_role"`
			UseCase     string `json:"use_case"`
			Company     string `json:"company"`
		}
		
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(400, gin.H{"error": "Invalid request"})
			return
		}
		
		// Log interest (in production: save to database)
		log.Printf("🎯 Interest registered: %s (%s) - %s - %s", 
			request.Email, request.Company, request.CompanyRole, request.UseCase)
		
		c.JSON(200, gin.H{
			"status": "success",
			"message": "Thank you for your interest. We'll be in touch soon.",
		})
	})

	// Interactive time series demo page
	r.GET("/demo", func(c *gin.Context) {
		html := `<!DOCTYPE html>
<html>
<head>
    <title>FootprintShift Carbon API - Live Demo</title>
    <style>
        /* Dieter Rams inspired design principles */
        * { 
            box-sizing: border-box; 
            margin: 0; 
            padding: 0; 
        }
        body { 
            font-family: 'Helvetica Neue', Helvetica, Arial, sans-serif; 
            font-weight: 300;
            line-height: 1.6;
            color: #2c2c2c;
            background: #f8f8f6;
            margin: 0;
            padding: 0;
        }
        
        /* Unified green mission color - fresh leaf green */
        :root {
            --green-primary: #22c55e;
            --green-light: #4ade80;
            --green-bg: #f0fdf4;
            --green-subtle: #dcfce7;
        }
        
        /* Grid system - Rams loved systematic layouts */
        .page {
            max-width: 1200px;
            margin: 0 auto;
            padding: 40px 30px;
            display: grid;
            grid-template-columns: 1fr;
            gap: 40px;
        }
        
        /* Typography hierarchy */
        .title {
            font-size: 28px;
            font-weight: 200;
            letter-spacing: 0.5px;
            color: #1a1a1a;
            margin-bottom: 8px;
        }
        
        .subtitle {
            font-size: 14px;
            font-weight: 400;
            color: #666;
            text-transform: uppercase;
            letter-spacing: 1px;
            margin-bottom: 30px;
        }
        
        /* Value proposition - elegant product messaging */
        .value-proposition {
            max-width: 600px;
            margin: 0 auto;
            text-align: center;
            padding: 30px 0;
        }
        
        .value-text {
            font-size: 16px;
            font-weight: 300;
            line-height: 1.6;
            color: #4a4a4a;
            margin-bottom: 25px;
        }
        
        .demo-badge {
            display: inline-flex;
            align-items: center;
            background: #f8f8f6;
            border: 1px solid #e0e0e0;
            padding: 8px 16px;
            border-radius: 2px;
            gap: 10px;
        }
        
        .demo-label {
            font-size: 10px;
            font-weight: 600;
            color: #666;
            text-transform: uppercase;
            letter-spacing: 1px;
        }
        
        .demo-description {
            font-size: 12px;
            color: #999;
            font-weight: 300;
        }
        
        /* Control section - minimal, functional */
        .control-section {
            background: white;
            border: 1px solid #e0e0e0;
            padding: 30px;
            border-radius: 2px;
            box-shadow: 0 1px 3px rgba(0,0,0,0.05);
        }
        
        .time-display {
            text-align: center;
            margin-bottom: 30px;
        }
        
        .current-time {
            font-size: 48px;
            font-weight: 100;
            color: #1a1a1a;
            letter-spacing: -1px;
            margin-bottom: 5px;
        }
        
        .time-label {
            font-size: 12px;
            color: #999;
            text-transform: uppercase;
            letter-spacing: 1px;
        }
        
        /* Timeline Chart - Bar visualization as primary interface */
        .timeline-container {
            position: relative;
            width: 100%;
            height: 80px;
            margin: 20px 0;
            background: white;
            border: 1px solid #e8e8e8;
            border-radius: 1px;
            overflow: hidden;
            cursor: crosshair;
        }
        
        .timeline-chart {
            position: absolute;
            bottom: 0;
            left: 0;
            right: 0;
            height: 80px;
            background: linear-gradient(to bottom, #fafafa, #f5f5f5);
        }
        
        /* Carbon intensity bars (foreground layer) */
        .carbon-bar, .timeline-bar {
            position: absolute;
            bottom: 0;
            width: calc(100% / 96 - 1px);
            transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
            border-radius: 0;
            margin-right: 1px;
            min-height: 15px;
            background: #d1d5db; /* Default gray */
            z-index: 2;
        }
        
        /* Renewable energy bars (background layer) */
        .renewable-bar {
            position: absolute;
            bottom: 0;
            width: calc(100% / 96 - 1px);
            transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
            border-radius: 0;
            margin-right: 1px;
            min-height: 10px;
            opacity: 0.3;
            z-index: 1;
            background: #22c55e; /* Always green background */
        }
        
        /* Color scheme: only green for optimal phases */
        .renewable-bar.green { background: #22c55e; } /* Fresh leaf green */
        .renewable-bar.normal { background: transparent; } /* Invisible for non-optimal */
        
        /* Carbon intensity grayscale levels */
        .carbon-bar.low { background: #d1d5db; }    /* Light gray - best */
        .carbon-bar.medium { background: #9ca3af; } /* Medium gray */
        .carbon-bar.high { background: #6b7280; }   /* Dark gray - worst */
        
        /* Current time indicator - smooth pixel movement */
        .current-time-indicator {
            position: absolute;
            top: 0;
            bottom: 0;
            width: 2px;
            background: #22c55e; /* Darker green for visibility */
            z-index: 20;
            pointer-events: none;
            box-shadow: 0 0 8px rgba(5, 150, 105, 0.4);
            transition: left 0.1s linear; /* Smooth pixel movement */
        }
        
        .current-time-indicator::before {
            content: '';
            position: absolute;
            top: -4px;
            left: -3px;
            width: 8px;
            height: 8px;
            background: #22c55e;
            border-radius: 50%;
            box-shadow: 0 0 4px rgba(5, 150, 105, 0.6);
        }
        
        /* Timeline interaction overlay */
        .timeline-interaction {
            position: absolute;
            top: 0;
            left: 0;
            right: 0;
            bottom: 0;
            z-index: 10;
            cursor: crosshair;
            background: transparent;
        }
        
        .time-markers {
            display: flex;
            justify-content: space-between;
            font-size: 10px;
            color: #999;
            text-transform: uppercase;
            letter-spacing: 0.5px;
            margin-top: 8px;
        }
        
        /* Renewable energy line overlay */
        .renewable-line {
            position: absolute;
            top: 0;
            left: 0;
            right: 0;
            bottom: 0;
            width: 100%;
            height: 80px;
            pointer-events: none;
            z-index: 5;
        }
        
        .renewable-path {
            stroke: #22c55e;
            stroke-width: 1.5;
            stroke-dasharray: 3, 3;
            fill: none;
            opacity: 0.7;
        }
        
        /* Subtle legend */
        .timeline-legend {
            position: absolute;
            top: 8px;
            right: 8px;
            font-size: 9px;
            color: #999;
            text-transform: uppercase;
            letter-spacing: 0.5px;
            background: rgba(255, 255, 255, 0.9);
            padding: 4px 6px;
            border-radius: 1px;
        }
        
        /* Hover tooltip */
        .timeline-tooltip {
            position: absolute;
            background: rgba(26, 26, 26, 0.9);
            color: white;
            padding: 6px 8px;
            font-size: 11px;
            border-radius: 2px;
            pointer-events: none;
            z-index: 20;
            opacity: 0;
            transition: opacity 0.2s ease;
            white-space: nowrap;
        }
        
        /* Data display - clean, systematic */
        .data-grid {
            display: grid;
            grid-template-columns: 1fr;
            gap: 40px;
            margin: 40px 0;
        }
        
        .primary-data {
            background: white;
            border: 1px solid #e0e0e0;
            border-radius: 2px;
            overflow: hidden;
        }
        
        .data-header {
            background: #f5f5f5;
            padding: 20px 30px;
            border-bottom: 1px solid #e0e0e0;
        }
        
        .data-title {
            font-size: 14px;
            font-weight: 500;
            color: #333;
            text-transform: uppercase;
            letter-spacing: 1px;
        }
        
        .data-content {
            padding: 30px;
        }
        
        .carbon-value {
            font-size: 64px;
            font-weight: 100;
            letter-spacing: -2px;
            margin-bottom: 10px;
            transition: all 0.4s cubic-bezier(0.4, 0, 0.2, 1); /* Elegant easing */
        }
        
        .carbon-unit {
            font-size: 14px;
            color: #666;
            font-weight: 400;
        }
        
        .status-indicator {
            display: inline-block;
            width: 8px;
            height: 8px;
            border-radius: 50%;
            margin-right: 8px;
        }
        
        .status-text {
            font-size: 14px;
            font-weight: 400;
            margin: 15px 0;
        }
        
        /* Color coding - Technical precision */
        .green { color: #22c55e; } /* Fresh leaf green */
        .yellow { color: #6b7280; } /* Dark gray */
        .red { color: #374151; } /* Darker gray */
        
        .green .status-indicator { background: #22c55e; } /* Fresh leaf green */
        .yellow .status-indicator { background: #6b7280; } /* Dark gray */
        .red .status-indicator { background: #374151; } /* Darker gray */
        
        /* Chart - minimal data visualization */
        .chart-container {
            background: white;
            border: 1px solid #e0e0e0;
            border-radius: 2px;
        }
        
        .chart {
            height: 200px;
            padding: 20px;
            position: relative;
            background: #fafafa;
        }
        
        .chart-bar {
            position: absolute;
            bottom: 20px;
            width: 1px; /* Thinner for 96 bars */
            background: #ccc;
            border-radius: 0;
            transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
        }
        
        .chart-bar.active {
            background: #333;
            width: 3px;
            box-shadow: 0 0 8px rgba(51, 51, 51, 0.3);
        }
        
        .chart-bar.green { background: #22c55e; } /* Fresh leaf green */
        .chart-bar.yellow { background: #9ca3af; } /* Medium gray */
        .chart-bar.red { background: #6b7280; } /* Dark gray */
        
        /* Timeline bar colors - green mission: only green is colored */
        .timeline-bar.green { background: #22c55e; } /* Fresh leaf green */
        .timeline-bar.yellow { background: #9ca3af; } /* Medium gray */
        .timeline-bar.red { background: #6b7280; } /* Dark gray */
        
        /* Dieter Rams Controls - Minimal Precision */
        .controls {
            display: flex;
            justify-content: center;
            gap: 2px;
            margin: 20px 0;
        }
        
        .control-btn {
            background: #f5f5f5;
            color: #666;
            border: 1px solid #e0e0e0;
            padding: 6px 12px;
            font-size: 10px;
            text-transform: uppercase;
            letter-spacing: 1px;
            cursor: pointer;
            border-radius: 0;
            transition: all 0.15s ease;
            font-weight: 500;
            font-family: inherit;
            position: relative;
        }
        
        .control-btn:hover {
            background: white;
            color: #333;
            border-color: #ccc;
        }
        
        .control-btn:active,
        .control-btn.active {
            background: #22c55e;
            color: white;
            border-color: #22c55e;
        }
        
        .control-btn:first-child {
            border-top-left-radius: 1px;
            border-bottom-left-radius: 1px;
        }
        
        .control-btn:last-child {
            border-top-right-radius: 1px;
            border-bottom-right-radius: 1px;
        }
        
        /* Optimization details - systematic layout */
        .optimization-section {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
            gap: 15px;
            margin: 20px 0;
        }
        
        .optimization-card {
            background: white;
            border: 1px solid #e0e0e0;
            border-radius: 2px;
        }
        
        .card-header {
            background: #f9f9f9;
            padding: 15px 20px;
            border-bottom: 1px solid #e0e0e0;
            font-size: 11px;
            text-transform: uppercase;
            letter-spacing: 1px;
            color: #666;
            font-weight: 500;
        }
        
        .card-content {
            padding: 20px;
        }
        
        .metric {
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin: 12px 0;
            font-size: 13px;
        }
        
        .metric-label {
            color: #666;
            font-weight: 400;
        }
        
        .metric-value {
            color: #333;
            font-weight: 500;
        }
        
        .savings-indicator {
            background: var(--green-bg);
            color: var(--green-primary);
            padding: 8px 12px;
            border-radius: 1px;
            font-size: 11px;
            font-weight: 500;
            text-align: center;
            margin-top: 15px;
        }
        
        /* Minimal info display */
        .info-grid {
            display: grid;
            grid-template-columns: 1fr 1fr;
            gap: 20px;
            margin: 20px 0;
        }
        
        .info-item {
            display: flex;
            justify-content: space-between;
            padding: 8px 0;
            border-bottom: 1px solid #f0f0f0;
            font-size: 13px;
        }
        
        .info-label {
            color: #666;
            font-weight: 400;
        }
        
        .info-value {
            color: #333;
            font-weight: 500;
        }
        
        /* CTA Section - Rams-inspired conversion design */
        .cta-section {
            background: #f8f8f6;
            border-top: 1px solid #e0e0e0;
            margin-top: 60px;
            padding: 60px 0;
        }
        
        .cta-content {
            max-width: 600px;
            margin: 0 auto;
            text-align: center;
        }
        
        .cta-title {
            font-size: 24px;
            font-weight: 400;
            color: #1a1a1a;
            margin-bottom: 15px;
            letter-spacing: 0.5px;
        }
        
        .cta-description {
            font-size: 16px;
            font-weight: 300;
            color: #666;
            line-height: 1.6;
            margin-bottom: 40px;
        }
        
        /* Benefits - clean metrics display */
        .benefits-grid {
            display: grid;
            grid-template-columns: repeat(3, 1fr);
            gap: 30px;
            margin-bottom: 50px;
        }
        
        .benefit {
            text-align: center;
        }
        
        .benefit-metric {
            font-size: 20px;
            font-weight: 500;
            color: var(--green-primary);
            margin-bottom: 5px;
        }
        
        .benefit-label {
            font-size: 12px;
            color: #999;
            text-transform: uppercase;
            letter-spacing: 1px;
        }
        
        /* Form - minimal, functional */
        .interest-form {
            max-width: 500px;
            margin: 0 auto;
        }
        
        .form-row {
            display: grid;
            grid-template-columns: 1fr 1fr;
            gap: 15px;
            margin-bottom: 15px;
        }
        
        .interest-form input,
        .interest-form select {
            padding: 12px 16px;
            border: 1px solid #ddd;
            border-radius: 0;
            font-size: 14px;
            font-family: inherit;
            background: white;
            color: #333;
            outline: none;
            transition: border-color 0.2s ease;
        }
        
        .interest-form input:focus,
        .interest-form select:focus {
            border-color: #666;
        }
        
        .interest-form input::placeholder {
            color: #999;
            font-weight: 300;
        }
        
        .cta-button {
            width: 100%;
            background: #22c55e;
            color: white;
            border: none;
            padding: 16px 24px;
            font-size: 14px;
            font-weight: 400;
            text-transform: uppercase;
            letter-spacing: 1px;
            cursor: pointer;
            border-radius: 0;
            transition: all 0.2s ease;
            margin-top: 10px;
        }
        
        .cta-button:hover {
            background: var(--green-light);
        }
        
        .cta-button:disabled {
            background: #ccc;
            cursor: not-allowed;
        }
        
        /* Success state */
        .form-success {
            text-align: center;
            padding: 30px;
            background: #f0f8f0;
            border: 1px solid #d4edda;
            border-radius: 2px;
            margin-top: 20px;
        }
        
        .success-icon {
            font-size: 24px;
            color: var(--green-primary);
            margin-bottom: 10px;
        }
        
        .success-message {
            font-size: 14px;
            color: var(--green-primary);
            font-weight: 400;
        }
        
        /* Early access note */
        .early-access-note {
            display: flex;
            justify-content: center;
            align-items: center;
            gap: 10px;
            margin-top: 25px;
            font-size: 11px;
        }
        
        .early-label {
            color: #666;
            font-weight: 600;
            text-transform: uppercase;
            letter-spacing: 1px;
        }
        
        .early-description {
            color: #999;
            font-weight: 300;
        }
        
        /* Responsive design */
        @media (max-width: 768px) {
            .data-grid {
                grid-template-columns: 1fr;
                gap: 20px;
            }
            
            .page {
                padding: 20px 15px;
            }
            
            .carbon-value {
                font-size: 48px;
            }
            
            .benefits-grid {
                grid-template-columns: 1fr;
                gap: 20px;
            }
            
            .form-row {
                grid-template-columns: 1fr;
            }
            
            .cta-section {
                padding: 40px 20px;
            }
        }
    </style>
</head>
<body>
    <div class="page">
        <header>
            <h1 class="title">FootprintShift</h1>
            <p class="subtitle">Real-time environmental optimization for digital products</p>
            
            <div class="value-proposition">
                <p class="value-text">Automatically optimize your application's environmental footprint based on real-time electricity grid data. Reduce CO₂ emissions and resource consumption without compromising user experience.</p>
                
                <div class="demo-badge">
                    <span class="demo-label">LIVE DEMO</span>
                    <span class="demo-description">Germany • 15-minute precision</span>
                </div>
            </div>
        </header>

        <section class="control-section">
            <div class="time-display">
                <div class="current-time" id="currentTime">12:00</div>
                <div class="time-label" id="timeLabel">MITTAG</div>
            </div>
            
            <div class="timeline-container">
                <div class="timeline-progress" id="timelineProgress"></div>
                <div class="timeline-chart" id="timelineChart"></div>
                <div class="current-time-indicator" id="currentTimeIndicator"></div>
                <div class="timeline-legend">CO₂ Intensity</div>
                <div class="timeline-tooltip" id="timelineTooltip"></div>
                <input type="range" class="time-slider" id="timeSlider" min="0" max="95" value="48" step="1">
            </div>
            
            <div class="time-markers">
                <span>00:00</span>
                <span>06:00</span>
                <span>12:00</span>
                <span>18:00</span>
                <span>23:00</span>
            </div>
            
            <div class="controls">
                <button class="control-btn" onclick="playTimeSeries()">Start</button>
                <button class="control-btn" onclick="pauseTimeSeries()">Pause</button>
                <button class="control-btn" onclick="resetTimeSeries()">Reset</button>
            </div>
        </section>

        <section class="data-grid">
            <div class="primary-data">
                <div class="data-header">
                    <div class="data-title">Carbon Intensity</div>
                </div>
                <div class="data-content">
                    <div class="carbon-value" id="carbonValue">295</div>
                    <div class="carbon-unit">g CO₂/kWh</div>
                    
                    <div class="status-text" id="statusText">
                        <span class="status-indicator"></span>
                        <span id="statusLabel">Normal Operation</span>
                    </div>
                    
                    <div class="info-grid">
                        <div class="info-item">
                            <span class="info-label">Renewable</span>
                            <span class="info-value" id="renewableValue">45%</span>
                        </div>
                        <div class="info-item">
                            <span class="info-label">Trend</span>
                            <span class="info-value" id="trendValue">Stable</span>
                        </div>
                        <div class="info-item">
                            <span class="info-label">Percentile</span>
                            <span class="info-value" id="percentileValue">50%</span>
                        </div>
                        <div class="info-item">
                            <span class="info-label">Next Green</span>
                            <span class="info-value" id="nextGreenValue">22:00</span>
                        </div>
                    </div>
                </div>
            </div>
        </section>

        <section class="optimization-section">
            <div class="optimization-card">
                <div class="card-header">Video Optimization</div>
                <div class="card-content">
                    <div class="metric">
                        <span class="metric-label">Quality</span>
                        <span class="metric-value" id="videoQuality">4K</span>
                    </div>
                    <div class="metric">
                        <span class="metric-label">Bitrate</span>
                        <span class="metric-value" id="videoBitrate">25000 kbps</span>
                    </div>
                    <div class="savings-indicator" id="videoSavings">
                        0g CO₂/h saved
                    </div>
                </div>
            </div>

            <div class="optimization-card">
                <div class="card-header">AI Processing</div>
                <div class="card-content">
                    <div class="metric">
                        <span class="metric-label">Status</span>
                        <span class="metric-value" id="aiStatus">Active</span>
                    </div>
                    <div class="metric">
                        <span class="metric-label">Deferred</span>
                        <span class="metric-value" id="aiDeferred">No</span>
                    </div>
                    <div class="savings-indicator" id="aiSavings">
                        0g CO₂/session saved
                    </div>
                </div>
            </div>

            <div class="optimization-card">
                <div class="card-header">GPU Features</div>
                <div class="card-content">
                    <div class="metric">
                        <span class="metric-label">WebGL</span>
                        <span class="metric-value" id="webglStatus">Enabled</span>
                    </div>
                    <div class="metric">
                        <span class="metric-label">3D Models</span>
                        <span class="metric-value" id="modelsStatus">Enabled</span>
                    </div>
                    <div class="savings-indicator" id="gpuSavings">
                        0g CO₂/h saved
                    </div>
                </div>
            </div>

            <div class="optimization-card">
                <div class="card-header">System Status</div>
                <div class="card-content">
                    <div class="metric">
                        <span class="metric-label">Mode</span>
                        <span class="metric-value" id="systemMode">Full</span>
                    </div>
                    <div class="metric">
                        <span class="metric-label">Eco Discount</span>
                        <span class="metric-value" id="ecoDiscount">0%</span>
                    </div>
                    <div class="savings-indicator" id="totalSavings">
                        Total: 0g CO₂ saved
                    </div>
                </div>
            </div>
        </section>

        <section class="cta-section">
            <div class="cta-content">
                <h2 class="cta-title">Integrate Environmental-Aware Computing</h2>
                <p class="cta-description">Join forward-thinking companies building sustainable digital products. Get early access to our environmental footprint optimization API.</p>
                
                <div class="benefits-grid">
                    <div class="benefit">
                        <div class="benefit-metric">24g CO₂/h</div>
                        <div class="benefit-label">Video optimization</div>
                    </div>
                    <div class="benefit">
                        <div class="benefit-metric">3g CO₂/session</div>
                        <div class="benefit-label">AI deferral</div>
                    </div>
                    <div class="benefit">
                        <div class="benefit-metric">15g CO₂/h</div>
                        <div class="benefit-label">GPU optimization</div>
                    </div>
                </div>
                
                <form class="interest-form" id="interestForm">
                    <div class="form-row">
                        <input type="email" id="email" placeholder="your.email@company.com" required>
                        <select id="role" required>
                            <option value="">Your role</option>
                            <option value="developer">Developer</option>
                            <option value="cto">CTO/Tech Lead</option>
                            <option value="sustainability">Sustainability Manager</option>
                            <option value="product">Product Manager</option>
                            <option value="other">Other</option>
                        </select>
                    </div>
                    <div class="form-row">
                        <input type="text" id="company" placeholder="Company name (optional)">
                        <select id="usecase">
                            <option value="">Primary use case</option>
                            <option value="video">Video streaming platform</option>
                            <option value="ai">AI/ML applications</option>
                            <option value="ecommerce">E-commerce</option>
                            <option value="saas">SaaS platform</option>
                            <option value="gaming">Gaming</option>
                            <option value="other">Other</option>
                        </select>
                    </div>
                    <button type="submit" class="cta-button" id="ctaButton">
                        Request Early Access
                    </button>
                </form>
                
                <div class="form-success" id="formSuccess" style="display: none;">
                    <div class="success-icon">✓</div>
                    <div class="success-message">Thank you! We'll be in touch soon with early access details.</div>
                </div>
                
                <div class="early-access-note">
                    <span class="early-label">EARLY ACCESS</span>
                    <span class="early-description">Free during beta • No commitment required</span>
                </div>
            </div>
        </section>
    </div>

    <script>
        let timeSeriesData = [];
        let isPlaying = false;
        let playInterval;
        
        // Load 24h time series data with 15-minute precision
        async function loadTimeSeriesData() {
            try {
                const response = await fetch('/api/v1/carbon-intensity/timeseries');
                if (!response.ok) {
                    throw new Error('HTTP error! status: ' + response.status);
                }
                const data = await response.json();
                timeSeriesData = data.timeseries; // 96 data points
                
                // Initialize timeline with renewable overlay
                createTimelineChart();
                
                // Set initial time index (12:00 = index 48)
                updateForTimeIndex(48);
                updateCurrentTimeIndicator();
            } catch (error) {
                console.error('Error loading time series data:', error);
            }
        }
        
        function createTimelineChart() {
            const chart = document.getElementById('timelineChart');
            chart.innerHTML = '';
            
            const maxIntensity = Math.max(...timeSeriesData.map(d => d.carbon_intensity));
            const minIntensity = Math.min(...timeSeriesData.map(d => d.carbon_intensity));
            const chartWidth = 100;
            const barWidth = chartWidth / 96;
            const maxBarHeight = 70;
            
            timeSeriesData.forEach((data, index) => {
                const bar = document.createElement('div');
                bar.className = 'timeline-bar';
                bar.style.left = (index * barWidth) + '%';
                
                const normalizedHeight = (data.carbon_intensity - minIntensity) / (maxIntensity - minIntensity);
                const barHeight = Math.max(8, normalizedHeight * maxBarHeight);
                bar.style.height = barHeight + 'px';
                
                if (data.mode === 'green') bar.classList.add('green');
                else if (data.mode === 'yellow') bar.classList.add('yellow');
                else bar.classList.add('red');
                
                // Enhanced hover with tooltip
                bar.addEventListener('mouseenter', (e) => showTooltip(e, data, index));
                bar.addEventListener('mouseleave', hideTooltip);
                
                chart.appendChild(bar);
            });
        }
        
        function createRenewableLine() {
            const svg = document.getElementById('renewableLine');
            svg.innerHTML = '';
            svg.setAttribute('viewBox', '0 0 100 120');
            svg.setAttribute('preserveAspectRatio', 'none');
            
            // Create path for renewable percentage - full height usage
            let pathData = 'M';
            
            timeSeriesData.forEach((data, index) => {
                const x = (index / (timeSeriesData.length - 1)) * 100;
                const y = 120 - (data.renewable_percentage / 100 * 120); // Use full height
                
                if (index === 0) {
                    pathData += x + ',' + y;
                } else {
                    pathData += ' L' + x + ',' + y;
                }
            });
            
            const path = document.createElementNS('http://www.w3.org/2000/svg', 'path');
            path.setAttribute('d', pathData);
            path.setAttribute('class', 'renewable-path');
            svg.appendChild(path);
        }
        
        function showTooltip(event, data, index) {
            const tooltip = document.getElementById('timelineTooltip');
            const timeStr = String(data.hour).padStart(2, '0') + ':' + 
                           String(data.minute).padStart(2, '0');
            
            tooltip.innerHTML = 
                '<strong>' + timeStr + '</strong><br>' +
                'CO₂: ' + Math.round(data.carbon_intensity) + 'g/kWh<br>' +
                'Renewable: ' + Math.round(data.renewable_percentage) + '%';
            
            const rect = event.target.getBoundingClientRect();
            const containerRect = document.querySelector('.timeline-container').getBoundingClientRect();
            
            tooltip.style.left = (rect.left - containerRect.left) + 'px';
            tooltip.style.top = '5px';
            tooltip.style.opacity = '1';
        }
        
        function hideTooltip() {
            const tooltip = document.getElementById('timelineTooltip');
            tooltip.style.opacity = '0';
        }
        
        function updateCurrentTimeIndicator() {
            // Update the visual indicator position based on current time index
            const progressWidth = (currentTimeIndex / 95) * 100;
            document.getElementById('timelineProgress').style.width = progressWidth + '%';
            
            // Also update the current time indicator line position
            const container = document.getElementById('timelineChart');
            const containerWidth = container.offsetWidth;
            const pixelPosition = (currentTimeIndex / 95) * containerWidth;
            document.getElementById('currentTimeIndicator').style.left = pixelPosition + 'px';
        }
        
        async function updateForTimeIndex(timeIndex) {
            try {
                currentTimeIndex = timeIndex;
                const carbonData = timeSeriesData[timeIndex];
                const hour = carbonData.hour;
                const minute = carbonData.minute;
                
                // Update time display - Rams-style minimalism with precise time
                const timeLabels = {
                    0: 'MITTERNACHT', 1: 'NACHT', 2: 'NACHT', 3: 'NACHT', 4: 'NACHT', 5: 'NACHT',
                    6: 'FRÜH', 7: 'FRÜH', 8: 'FRÜH', 9: 'VORMITTAG', 10: 'VORMITTAG', 11: 'VORMITTAG',
                    12: 'MITTAG', 13: 'MITTAG', 14: 'NACHMITTAG', 15: 'NACHMITTAG', 16: 'NACHMITTAG', 17: 'NACHMITTAG',
                    18: 'ABEND', 19: 'ABEND', 20: 'ABEND', 21: 'NACHT', 22: 'NACHT', 23: 'NACHT'
                };
                
                // Elegant time display with smooth updates
                const timeStr = String(hour).padStart(2, '0') + ':' + String(minute).padStart(2, '0');
                document.getElementById('currentTime').textContent = timeStr;
                document.getElementById('timeLabel').textContent = timeLabels[hour] || 'TAG';
                
                // Get optimization data for this time
                const response = await fetch('/api/v1/optimization/' + hour);
                const optData = await response.json();
                
                // Update carbon display - systematic approach
                document.getElementById('carbonValue').textContent = Math.round(carbonData.carbon_intensity);
                
                // Status text with minimal indicator
                const statusText = document.getElementById('statusText');
                const statusLabel = document.getElementById('statusLabel');
                
                statusText.className = 'status-text ' + carbonData.mode;
                
                const statusMessages = {
                    'green': 'Optimal Operation',
                    'yellow': 'Normal Operation', 
                    'red': 'Reduced Operation'
                };
                statusLabel.textContent = statusMessages[carbonData.mode] || 'Normal Operation';
                
                // Update info grid - precise data
                document.getElementById('renewableValue').textContent = Math.round(carbonData.renewable_percentage) + '%';
                document.getElementById('trendValue').textContent = carbonData.trend_direction.charAt(0).toUpperCase() + carbonData.trend_direction.slice(1);
                document.getElementById('percentileValue').textContent = Math.round(carbonData.local_percentile) + '%';
                
                // Next green window
                const nextGreenHour = carbonData.next_green_window ? 
                    new Date(carbonData.next_green_window).getHours() : (hour + 4) % 24;
                document.getElementById('nextGreenValue').textContent = 
                    (nextGreenHour < 10 ? '0' : '') + nextGreenHour + ':00';
                
                // Update optimizations
                const opt = optData.optimization;
                updateOptimizationCards(opt);
                
                // Highlight current time index in timeline - smooth visual feedback
                document.querySelectorAll('.timeline-bar').forEach((bar, index) => {
                    bar.classList.toggle('active', index === timeIndex);
                });
                
                // Update progress indicator smoothly
                const progressWidth = (timeIndex / 95) * 100;
                document.getElementById('timelineProgress').style.width = progressWidth + '%';
                
                // Update slider position
                document.getElementById('timeSlider').value = timeIndex;
                
            } catch (error) {
                console.error('Error updating for hour:', hour, error);
            }
        }
        
        
        function updateOptimizationCards(opt) {
            // Video optimization - precise metrics
            document.getElementById('videoQuality').textContent = opt.video_quality || '4K';
            document.getElementById('videoBitrate').textContent = 
                (opt.max_video_bitrate_kbps || 25000).toLocaleString() + ' kbps';
            document.getElementById('videoSavings').textContent = 
                Math.round(opt.video_co2_savings_per_hour_g || 0) + 'g CO₂/h saved';
            
            // AI optimization - status focused
            document.getElementById('aiStatus').textContent = 
                opt.ai_deferred_to_green_window ? 'Deferred' : 'Active';
            document.getElementById('aiDeferred').textContent = 
                opt.ai_deferred_to_green_window ? 'Yes' : 'No';
            document.getElementById('aiSavings').textContent = 
                Math.round(opt.ai_co2_savings_per_session_g || 0) + 'g CO₂/session saved';
            
            // GPU optimization - binary states
            document.getElementById('webglStatus').textContent = 
                opt.gpu_features_disabled ? 'Disabled' : 'Enabled';
            document.getElementById('modelsStatus').textContent = 
                (opt.disable_features && opt.disable_features.includes('3d_models')) ? 'Disabled' : 'Enabled';
            document.getElementById('gpuSavings').textContent = 
                Math.round(opt.gpu_co2_savings_per_hour_g || 0) + 'g CO₂/h saved';
            
            // System status - minimal information
            document.getElementById('systemMode').textContent = 
                opt.mode.charAt(0).toUpperCase() + opt.mode.slice(1);
            document.getElementById('ecoDiscount').textContent = (opt.eco_discount || 0) + '%';
            
            // Total savings calculation
            const totalSavings = (opt.video_co2_savings_per_hour_g || 0) + 
                               (opt.ai_co2_savings_per_session_g || 0) + 
                               (opt.gpu_co2_savings_per_hour_g || 0);
            document.getElementById('totalSavings').textContent = 
                'Total: ' + Math.round(totalSavings) + 'g CO₂ saved';
        }
        
        
        function playTimeSeries() {
            if (isPlaying) return;
            
            isPlaying = true;
            const container = document.getElementById('timelineChart');
            const containerWidth = container.offsetWidth;
            const pixelsPerTimeIndex = containerWidth / 96;
            
            // Start smooth pixel-by-pixel animation
            let pixelPosition = (currentTimeIndex / 95) * containerWidth;
            let lastDataUpdateIndex = currentTimeIndex;
            
            playInterval = setInterval(() => {
                // Move indicator pixel by pixel
                pixelPosition += 2; // 2 pixels per frame for smooth motion
                
                if (pixelPosition >= containerWidth) {
                    pixelPosition = 0; // Loop back to start
                    currentTimeIndex = 0;
                    lastDataUpdateIndex = 0;
                    setTimeout(() => pauseTimeSeries(), 300);
                    return;
                }
                
                // Update indicator position smoothly
                const indicator = document.getElementById('currentTimeIndicator');
                indicator.style.left = pixelPosition + 'px';
                
                // Also update progress width
                const progressWidth = (pixelPosition / containerWidth) * 100;
                document.getElementById('timelineProgress').style.width = progressWidth + '%';
                
                // Calculate which time index we should be at
                const targetTimeIndex = Math.floor((pixelPosition / containerWidth) * 96);
                
                // Update data only when we cross into a new 15-minute period
                if (targetTimeIndex !== lastDataUpdateIndex && targetTimeIndex < 96) {
                    currentTimeIndex = targetTimeIndex;
                    updateForTimeIndex(currentTimeIndex);
                    lastDataUpdateIndex = targetTimeIndex;
                }
                
            }, 50); // 50ms intervals for smooth pixel movement (20fps)
        }
        
        function pauseTimeSeries() {
            isPlaying = false;
            if (playInterval) {
                clearInterval(playInterval);
                playInterval = null;
            }
        }
        
        function resetTimeSeries() {
            pauseTimeSeries();
            updateForTimeIndex(0); // Start at 00:00
        }
        
        // Event listeners - chart-based interaction
        document.querySelector('.timeline-container').addEventListener('click', (e) => {
            const container = e.currentTarget;
            const rect = container.getBoundingClientRect();
            const clickX = e.clientX - rect.left;
            const containerWidth = rect.width;
            
            // Calculate time index from click position
            const clickTimeIndex = Math.floor((clickX / containerWidth) * 96);
            const boundedTimeIndex = Math.max(0, Math.min(95, clickTimeIndex));
            
            // Update to clicked time
            pauseTimeSeries(); // Stop animation if running
            updateForTimeIndex(boundedTimeIndex);
        });
        
        // Add resize handler to keep indicator positioned correctly
        window.addEventListener('resize', () => {
            if (!isPlaying) {
                updateCurrentTimeIndicator();
            }
        });
        
        // Form submission handler
        document.getElementById('interestForm').addEventListener('submit', async (e) => {
            e.preventDefault();
            
            const button = document.getElementById('ctaButton');
            const form = document.getElementById('interestForm');
            const success = document.getElementById('formSuccess');
            
            // Disable button during submission
            button.disabled = true;
            button.textContent = 'Submitting...';
            
            try {
                const formData = {
                    email: document.getElementById('email').value,
                    company_role: document.getElementById('role').value,
                    use_case: document.getElementById('usecase').value,
                    company: document.getElementById('company').value || ''
                };
                
                const response = await fetch('/api/v1/register-interest', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json'
                    },
                    body: JSON.stringify(formData)
                });
                
                if (response.ok) {
                    // Success - show success state
                    form.style.display = 'none';
                    success.style.display = 'block';
                } else {
                    throw new Error('Submission failed');
                }
                
            } catch (error) {
                console.error('Error submitting form:', error);
                
                // Reset button on error
                button.disabled = false;
                button.textContent = 'Request Early Access';
                
                // Show error (in production: proper error handling)
                alert('Something went wrong. Please try again.');
            }
        });

        // Initialize
        loadTimeSeriesData();
    </script>
</body>
</html>`
		c.Data(200, "text/html; charset=utf-8", []byte(html))
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8090"
	}

	log.Printf("🌍 FootprintShift API Demo starting on port %s", port)
	log.Printf("⏰ 24h Germany carbon intensity simulation")
	log.Printf("🔬 Realistic solar/wind/demand modeling")
	log.Printf("🎮 Interactive time slider demo")
	
	// Show appropriate URL based on environment
	if os.Getenv("RAILWAY_ENVIRONMENT") != "" {
		log.Printf("📊 Demo: https://footprintshift-demo.railway.app/demo")
	} else if os.Getenv("RENDER") != "" {
		log.Printf("📊 Demo: https://footprintshift-demo.onrender.com/demo")
	} else {
		log.Printf("📊 Demo: http://localhost:%s/demo", port)
	}
	
	r.Run(":" + port)
}

// Helper functions for metadata calculation
func calculateAverage(data []TimeSeriesCarbonIntensity) float64 {
	sum := 0.0
	for _, d := range data {
		sum += d.CarbonIntensity
	}
	return math.Round(sum/float64(len(data))*10) / 10
}

func findMin(data []TimeSeriesCarbonIntensity) gin.H {
	min := data[0]
	for _, d := range data {
		if d.CarbonIntensity < min.CarbonIntensity {
			min = d
		}
	}
	return gin.H{"hour": min.Hour, "intensity": min.CarbonIntensity}
}

func findMax(data []TimeSeriesCarbonIntensity) gin.H {
	max := data[0]
	for _, d := range data {
		if d.CarbonIntensity > max.CarbonIntensity {
			max = d
		}
	}
	return gin.H{"hour": max.Hour, "intensity": max.CarbonIntensity}
}

func countGreenPeriods(data []TimeSeriesCarbonIntensity) []gin.H {
	var greenPeriods []gin.H
	for _, d := range data {
		if d.Mode == "green" {
			timeStr := fmt.Sprintf("%02d:%02d", d.Hour, d.Minute)
			greenPeriods = append(greenPeriods, gin.H{
				"time": timeStr,
				"time_index": d.TimeIndex,
				"hour": d.Hour,
				"minute": d.Minute,
			})
		}
	}
	return greenPeriods
}

func findPeakRenewable(data []TimeSeriesCarbonIntensity) gin.H {
	max := data[0]
	for _, d := range data {
		if d.RenewablePercent > max.RenewablePercent {
			max = d
		}
	}
	timeStr := fmt.Sprintf("%02d:%02d", max.Hour, max.Minute)
	return gin.H{
		"time": timeStr,
		"hour": max.Hour, 
		"minute": max.Minute,
		"renewable_percent": max.RenewablePercent,
	}
}

func findOptimalWindows(data []TimeSeriesCarbonIntensity) []gin.H {
	var windows []gin.H
	for _, d := range data {
		if d.Mode == "green" {
			timeStr := fmt.Sprintf("%02d:%02d", d.Hour, d.Minute)
			windows = append(windows, gin.H{
				"time": timeStr,
				"hour": d.Hour,
				"minute": d.Minute,
				"time_index": d.TimeIndex,
				"intensity": d.CarbonIntensity,
				"renewable": d.RenewablePercent,
			})
		}
	}
	return windows
}