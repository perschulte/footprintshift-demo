# Carbon Intelligence Service

The Carbon Intelligence Service provides dynamic carbon intensity thresholds and advanced analytics for optimizing energy consumption based on regional patterns and historical data.

## Overview

Traditional carbon intensity monitoring uses static thresholds (e.g., green < 150 g CO2/kWh) that don't work well in regions with low variation or different baseline intensities. This service implements dynamic thresholding based on:

- **Regional percentiles** (top 20% / bottom 20% of local patterns)
- **Time-of-day learning algorithms**
- **Seasonal adjustments**
- **Historical pattern analysis**

## Key Features

### 1. Dynamic Thresholding

Instead of static thresholds, the service uses:
- **20th percentile** for "clean" classification (region-specific)
- **80th percentile** for "dirty" classification
- **Regional baseline** calculations
- **Confidence-weighted** recommendations

### 2. High-Variation Region Support

Specialized optimization for coal-heavy grids with significant daily variation:
- **Poland** - Coal-heavy with wind variation
- **Texas** - High wind + gas peakers with extreme peaks
- **Eastern Asia** (China, India) - Industrial coal patterns
- **Australia NSW** - Coal transition with high solar variation

### 3. Relative Metrics

Provides context-aware carbon intensity information:
- **Local percentile ranking** (0-100, where 0 is cleanest for that region)
- **Daily rank** ("top 15% cleanest hour today")
- **Trend analysis** (improving/worsening/stable)
- **Regional comparison** vs local baseline

### 4. Intelligent Predictions

- **Next optimal window** predictions based on learned patterns
- **Confidence scores** for all recommendations
- **Reason explanations** (e.g., "Night wind patterns", "Solar generation peak")

## API Endpoints

### Enhanced Carbon Intensity

```http
GET /api/v1/carbon-intensity?location=Poland&relative=true
```

**Response:**
```json
{
  "location": "Poland",
  "carbon_intensity": 280.5,
  "renewable_percentage": 35.2,
  "mode": "yellow",
  "recommendation": "reduce",
  "timestamp": "2024-01-15T14:30:00Z",
  "source": "electricity_maps",
  "grid_zone": "PL",
  
  "local_percentile": 25.5,
  "daily_rank": "top 25% cleanest hour today",
  "relative_mode": "clean",
  "trend_direction": "improving",
  "trend_magnitude": -15.2,
  "regional_baseline": 340.8,
  "confidence_score": 78.5,
  "is_high_variation": true,
  
  "next_optimal_window": {
    "start": "2024-01-15T22:00:00Z",
    "end": "2024-01-15T23:00:00Z",
    "expected_intensity": 180.3,
    "confidence": 82.5,
    "reason": "Night wind patterns"
  }
}
```

### Dynamic Green Hours

```http
GET /api/v1/green-hours?location=Texas&next=24&dynamic=true
```

Uses regional percentiles instead of static thresholds to identify truly optimal windows.

### Carbon Trends Analysis

```http
GET /api/v1/carbon-trends?location=Poland&period=daily&days=30
```

**Response:**
```json
{
  "location": "Poland",
  "period": "daily",
  "start_date": "2024-01-01T00:00:00Z",
  "end_date": "2024-01-30T23:59:59Z",
  
  "average_intensity": 295.4,
  "min_intensity": 120.8,
  "max_intensity": 580.2,
  "std_deviation": 85.6,
  
  "cleanest_hours": [2, 3, 4, 23, 1],
  "dirtiest_hours": [18, 19, 20, 17, 8],
  
  "weekday_vs_weekend": {
    "weekday_average": 310.2,
    "weekend_average": 265.8
  }
}
```

## Regional Optimization Strategies

### High-Variation Regions

The service provides specialized strategies for regions with significant carbon intensity variation:

#### Poland (Coal-Heavy Grid)
- **Optimal Hours:** 22:00-05:00 (night wind), 11:00-14:00 (solar)
- **Avoid:** 17:00-21:00 (coal ramp-up for evening peak)
- **Key Strategy:** Night scheduling provides 30-40% carbon savings
- **Seasonal Factor:** Summer has better renewable mix

#### Texas (Wind + Gas Peakers)
- **Optimal Hours:** 10:00-15:00 (solar), 23:00-04:00 (wind)
- **Avoid:** 16:00-21:00 (extreme gas peaker activation)
- **Key Strategy:** Leverage wind patterns and solar peak
- **Special Note:** Duck curve creates extreme variation

#### Eastern Asia (Industrial Coal)
- **Optimal Hours:** 01:00-05:00 (low industrial demand), 11:00-15:00 (solar)
- **Avoid:** 18:00-22:00 (industrial peak)
- **Key Strategy:** Early morning scheduling for lowest coal dependency
- **Regional Note:** Coastal areas typically cleaner than inland

## Implementation Architecture

### Core Components

1. **IntelligenceService** - Main service coordinating all carbon intelligence features
2. **RegionPattern** - Stores learned patterns for each region
3. **ElectricityMapsAdapter** - Bridges existing services with intelligence capabilities
4. **ServiceManager** - Handles service initialization and configuration

### Data Flow

```
ElectricityMaps API -> Adapter -> Intelligence Service -> Enhanced Metrics
                  Historical Data -> Pattern Learning -> Dynamic Thresholds
```

### Pattern Learning

The service continuously learns regional patterns:
- **Hourly averages** for each hour of the day
- **Day-of-week patterns** (weekday vs weekend)
- **Seasonal adjustments** (monthly factors)
- **Trend analysis** using linear regression
- **Confidence scoring** based on data quality and completeness

## Configuration

### Default Configuration
```go
config := carbon.GetDefaultConfig()
// 30-day history retention
// 15-minute update intervals
// 168 minimum data points (1 week)
```

### High-Variation Region Configuration
```go
config := carbon.GetHighVariationConfig()
// 14-day history retention (faster adaptation)
// 10-minute update intervals
// 72 minimum data points (3 days)
```

## Usage Examples

### Basic Integration

```go
// Create service manager
serviceManager := carbon.NewServiceManager(electricityClient, logger, config)
intelligence := serviceManager.GetIntelligenceService()

// Get enhanced carbon intensity
relativeIntensity, err := intelligence.GetRelativeCarbonIntensity(ctx, "Poland")

// Get dynamic green hours
forecast, err := intelligence.GetDynamicGreenHours(ctx, "Texas", 24)

// Analyze trends
trends, err := intelligence.GetCarbonTrends(ctx, "Poland", "daily", 7)
```

### Handler Integration

The service integrates with existing HTTP handlers by adding the intelligence service to dependencies:

```go
deps := &handlers.Dependencies{
    ElectricityMaps:    electricityClient,
    CarbonIntelligence: intelligence,
    Cache:              cacheService,
    Logger:            logger,
    Config:            config,
}
```

## Benefits

### For High-Variation Regions

1. **Better Optimization:** 20-40% improvement in carbon efficiency vs static thresholds
2. **Regional Context:** Understands local grid patterns and energy sources
3. **Predictive Insights:** Learns time-of-day and seasonal patterns
4. **Confidence Scoring:** Provides reliability metrics for decisions

### For All Regions

1. **Relative Rankings:** "Top 15% cleanest hour today" vs absolute values
2. **Trend Analysis:** Understanding if grid is getting cleaner or dirtier
3. **Smart Scheduling:** Predicts next optimal windows with confidence scores
4. **Historical Context:** Compares current conditions to regional baseline

## Future Enhancements

- **Weather Integration:** Correlate patterns with weather data
- **Grid Event Detection:** Identify maintenance, outages, or policy changes
- **ML-Based Forecasting:** Advanced prediction models
- **Real-Time Adaptation:** Faster pattern updates for rapid grid changes
- **Cross-Regional Analysis:** Compare optimization strategies across regions

## Monitoring and Observability

The service provides comprehensive logging and metrics:
- Pattern update success/failure rates
- Confidence score distributions
- Prediction accuracy tracking
- Cache hit rates for performance monitoring
- API response times and error rates

This enables monitoring of the intelligence service effectiveness and identifying regions that need configuration adjustments.