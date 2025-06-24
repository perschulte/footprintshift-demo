# Impact Calculation Service

A science-based CO₂ impact calculator designed to provide conservative, transparent estimates while preventing greenwashing.

## Overview

This service calculates carbon emissions from digital activities using real-world data and conservative methodologies. It includes comprehensive anti-greenwashing features to ensure claimed savings are realistic and verifiable.

## Key Features

### 🔬 Science-Based Calculations
- **Real emission factors** from IEA, Carbon Trust, EPA
- **Regional grid variations** (EU: 295g CO₂/kWh, US: 420g, etc.)
- **Device-specific consumption** patterns (smartphone: 2Wh/hour browsing)
- **Network transmission** costs by connection type
- **Data center PUE** values by region

### 🛡️ Anti-Greenwashing Protection
- **Conservative estimates** with ±25% confidence intervals
- **Rebound effect calculations** (10-40% depending on optimization type)
- **Device energy inclusion** (often 50%+ of total footprint)
- **Methodology transparency** with data source attribution
- **Validation API** to check claimed savings

### 📊 Comprehensive Tracking
- **Baseline measurement** for before/after comparisons
- **Session tracking** for real-time monitoring
- **Impact reporting** with equivalencies (miles driven, trees planted)
- **Real-time dashboard** with live metrics

## API Endpoints

### Core Calculation
```http
POST /api/v1/impact/calculate
```
Calculate carbon impact for various activities:
- Video streaming (36g CO₂/hour HD in EU)
- Image optimization (WebP/AVIF savings)
- AI inference (6Wh per GPT-3 scale inference)
- JavaScript execution
- Page loading
- Data transfer

### Baseline & Validation
```http
POST /api/v1/impact/baseline    # Measure current footprint
POST /api/v1/impact/validate    # Validate claimed savings
```

### Reporting & Dashboard
```http
GET /api/v1/impact/report       # Generate impact reports
GET /api/v1/impact/dashboard    # Real-time dashboard
GET /api/v1/impact/metrics/realtime  # Live metrics
```

### Session Tracking
```http
POST /api/v1/impact/session/{id}/start     # Start tracking
POST /api/v1/impact/session/{id}/activity  # Record activity
POST /api/v1/impact/session/{id}/end       # End session
```

## Calculation Methodology

### Video Streaming
Based on Shift Project (2023 revised) and Carbon Trust research:
- **Bitrates**: 360p: 1Mbps, 720p: 5Mbps, 1080p: 8Mbps, 4K: 25Mbps
- **Device consumption**: 2-40Wh/hour depending on device type
- **Network transmission**: 3.5-11g CO₂/GB depending on connection
- **Data center**: 0.012 kWh/GB including CDN costs

### Image Optimization
- **Baseline**: 0.5MB average per image
- **WebP savings**: Up to 35% size reduction
- **AVIF savings**: Up to 50% size reduction
- **Network cost**: Connection-dependent transmission emissions
- **Device cost**: Loading and rendering energy

### AI Inference
Based on Strubell et al. (2019) and Patterson et al. (2021):
- **Base consumption**: 0.006 kWh per GPT-3 scale inference
- **Data center PUE**: 1.1 for modern AI facilities
- **Optimization potential**: 80% through caching/batching

### JavaScript Execution
- **Device consumption**: 2.8-75Wh/hour depending on device
- **Bundle optimization**: 40-60% reduction through tree shaking
- **Execution optimization**: 30-50% reduction through code splitting

## Emission Factors (2023 Data)

### Grid Carbon Intensity (g CO₂/kWh)
- **EU Average**: 295
- **United States**: 420
- **China**: 580
- **India**: 720
- **United Kingdom**: 233
- **France**: 85 (nuclear-heavy)
- **Germany**: 380
- **Global Average**: 475

### Network Transmission (g CO₂/GB)
- **3G Mobile**: 11.0
- **4G Mobile**: 7.0
- **5G Mobile**: 5.0
- **WiFi**: 3.5
- **Ethernet**: 2.9
- **Fixed Broadband**: 3.2

### Device Consumption by Activity
#### Smartphone (Wh/hour)
- Idle: 0.5, Browsing: 2.0, Video: 3.5, Heavy JS: 2.8

#### Laptop (Wh/hour)
- Idle: 15.0, Browsing: 25.0, Video: 40.0, Heavy JS: 35.0

#### Desktop (Wh/hour)
- Idle: 40.0, Browsing: 60.0, Video: 85.0, Heavy JS: 75.0

## Rebound Effects

The service accounts for rebound effects - increased consumption due to efficiency gains:

- **Video Streaming**: 30% (better quality → more watching)
- **AI Inference**: 40% (cheaper AI → more usage)
- **Page Loading**: 20% (faster pages → more browsing)
- **Image Loading**: 10% (minimal rebound)
- **JavaScript**: 5% (technical optimization)

## Confidence Intervals

All calculations include ±25% confidence intervals to account for:
- Regional grid variations
- Device efficiency differences
- Usage pattern variations
- Network condition changes
- Measurement uncertainties

## Data Sources

- **IEA (2023)**: Electricity grid carbon intensity by region
- **Carbon Trust (2021)**: Digital service emissions methodology
- **EPA (2023)**: Data center PUE values and efficiency metrics
- **HTTP Archive (2023)**: Web page size and performance statistics
- **Shift Project (2023)**: Revised video streaming impact analysis
- **Strubell et al. (2019)**: Energy and policy considerations for deep learning
- **Patterson et al. (2021)**: Carbon emissions and large neural network training

## Usage Examples

### Calculate Video Streaming Impact
```json
POST /api/v1/impact/calculate
{
  "type": "video_streaming",
  "duration": 3600,
  "video_quality": "1080p",
  "device_type": "laptop",
  "connection_type": "wifi",
  "region": "EU",
  "optimization_level": 30,
  "include_rebound_effects": true
}
```

Response:
```json
{
  "baseline_emissions": 295.7,
  "optimized_emissions": 206.9,
  "savings": 88.8,
  "savings_percentage": 30.0,
  "confidence_interval": 25.0,
  "net_savings": 62.2,
  "rebound_effect": 26.6,
  "components": {
    "device_emissions": 190.4,
    "network_emissions": 67.8,
    "datacenter_emissions": 37.5,
    "device_percentage": 64.4,
    "network_percentage": 22.9,
    "datacenter_percentage": 12.7
  },
  "methodology": "Conservative calculation based on device energy consumption, network transmission, and data center operations. Includes ±25% confidence interval.",
  "warnings": [
    "Rebound effect estimated at 30% - improved efficiency may lead to increased consumption",
    "Device emissions account for 64% of total - user device efficiency is critical"
  ]
}
```

### Validate Claimed Savings
```json
POST /api/v1/impact/validate
{
  "claimed_savings": 150.5,
  "optimization_type": "image_loading",
  "parameters": {
    "image_count": 30,
    "data_size": 15.0
  }
}
```

Response:
```json
{
  "is_valid": false,
  "validated_savings": 89.2,
  "variance": 68.7,
  "rating": "optimistic",
  "explanation": "Your claimed savings of 150.50 g CO2 compared to our calculated 89.20 g CO2 (68.7% variance)",
  "suggestions": [
    "Use conservative estimates to avoid greenwashing",
    "Include device energy consumption in calculations",
    "Account for rebound effects"
  ]
}
```

## Anti-Greenwashing Features

1. **Conservative Methodology**: All estimates err on the side of caution
2. **Transparency**: Full methodology and data source disclosure
3. **Confidence Intervals**: ±25% uncertainty acknowledgment
4. **Rebound Effects**: Account for increased consumption from efficiency
5. **Device Inclusion**: Include often-ignored client-side energy consumption
6. **Validation API**: Check claimed savings against scientific baselines
7. **Warning System**: Alert users to limitations and assumptions

## Implementation Notes

- Uses in-memory storage for demonstration (production should use PostgreSQL)
- Conservative emission factors from latest research
- Regional variations properly accounted for
- Device consumption based on measured data
- Network costs vary by connection type and provider
- Data center emissions include PUE and cooling costs

## Contributing

When adding new calculation types:
1. Use conservative emission factors
2. Include confidence intervals
3. Account for rebound effects
4. Provide methodology transparency
5. Add validation warnings
6. Update data sources documentation