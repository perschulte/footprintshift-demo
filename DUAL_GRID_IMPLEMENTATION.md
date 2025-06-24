# Dual-Grid Carbon Detection Implementation

This document describes the comprehensive dual-grid carbon detection system implemented in the GreenWeb API to address the critical issue of CDN edge locations being in different carbon zones than users.

## Overview

Traditional carbon intensity monitoring only considers a single location (usually the user's location). However, in modern web architectures with CDNs, content is served from edge locations that may be in completely different electricity grids with different carbon intensities. This implementation provides:

1. **Dual-location carbon monitoring** - Track both user and edge carbon intensity
2. **Weighted carbon calculations** - Combine transmission and computation carbon costs
3. **CDN-aware optimization** - Provide recommendations specific to CDN providers
4. **Content-type specific analysis** - Different weights for different content types
5. **Real-time edge optimization** - Find optimal edge locations for minimal carbon impact

## Architecture

### Core Components

#### 1. Carbon Types (`pkg/carbon/`)

- **`dual_grid.go`** - Core dual-grid types and calculations
- **`cdn_providers.go`** - CDN provider configurations and edge location mappings
- **`types.go`** - Base carbon intensity types (existing)

#### 2. Enhanced Services (`service/`)

- **`electricity_maps.go`** - Extended with dual-grid methods
- **`dual_grid_optimization.go`** - CDN-aware optimization service

#### 3. Geolocation Enhancement (`internal/geolocation/`)

- **`dual_grid_service.go`** - Enhanced geolocation with CDN mapping
- **`service.go`** - Base geolocation service (existing)
- **`types.go`** - Location and grid zone types (existing)

#### 4. New API Handlers (`internal/handlers/`)

- **`dual_grid.go`** - Dual-grid specific endpoints
- **`handlers.go`** - Updated with dual-grid integration

## Key Features

### 1. Dual-Grid Carbon Intensity Analysis

```go
type DualGridCarbonIntensity struct {
    UserLocation       *CarbonIntensity `json:"user_location"`
    EdgeLocation       *CarbonIntensity `json:"edge_location"`
    WeightedIntensity  float64          `json:"weighted_intensity"`
    TransmissionWeight float64          `json:"transmission_weight"`
    ComputationWeight  float64          `json:"computation_weight"`
    // ... additional fields
}
```

**Weighted Calculation Logic:**
```
WeightedIntensity = (UserIntensity × TransmissionWeight) + (EdgeIntensity × ComputationWeight)
```

### 2. Content-Type Specific Weights

Different content types have different carbon footprint distributions:

| Content Type | Transmission Weight | Computation Weight | Use Case |
|--------------|--------------------|--------------------|-----------|
| Static       | 80%                | 20%                | Static assets, images |
| API          | 40%                | 60%                | REST/GraphQL APIs |
| Video        | 60%                | 40%                | Video streaming |
| Dynamic      | 30%                | 70%                | Server-rendered pages |
| AI           | 20%                | 80%                | ML inference |
| Database     | 25%                | 75%                | Database queries |

### 3. CDN Provider Support

Currently supports major CDN providers with 100+ edge locations:

- **CloudFlare** - 35+ global edge locations
- **AWS CloudFront** - 40+ edge locations across AWS regions  
- **Google Cloud CDN** - 30+ locations with carbon-aware features
- **Azure CDN** - 25+ locations across Azure regions

Each provider includes:
- Exact coordinates for edge locations
- Grid zone mappings
- Tier information (1=primary, 2=secondary, 3=tertiary)
- Renewable energy commitment flags
- Capacity indicators

### 4. Intelligent Optimization Recommendations

The system provides actionable recommendations:

```go
type DualGridRecommendation struct {
    Action              string               `json:"action"` // proceed, optimize, defer, relocate
    Reason              string               `json:"reason"`
    AlternativeEdges    []EdgeAlternative    `json:"alternative_edges"`
    OptimizationTips    []string             `json:"optimization_tips"`
    EstimatedSavings    float64              `json:"estimated_savings_g_co2"`
    TimeBasedStrategy   *TimeBasedStrategy   `json:"time_based_strategy"`
}
```

## API Endpoints

### 1. Dual-Grid Carbon Intensity

```http
GET /api/v1/dual-grid/carbon-intensity
```

**Parameters:**
- `user_location` (optional) - User location (auto-detected if not provided)
- `edge_location` (required) - Edge/server location
- `content_type` (optional) - Content type (default: static)
- `cdn_provider` (optional) - CDN provider for alternatives

**Response:**
```json
{
  "user_location": {
    "location": "Berlin",
    "carbon_intensity": 120.5,
    "renewable_percentage": 75.2,
    "mode": "green"
  },
  "edge_location": {
    "location": "Frankfurt", 
    "carbon_intensity": 95.3,
    "renewable_percentage": 85.7,
    "mode": "green"
  },
  "weighted_intensity": 125.8,
  "transmission_weight": 80,
  "computation_weight": 20,
  "content_type": "static",
  "recommendation": {
    "action": "proceed",
    "reason": "Both locations have low carbon intensity",
    "optimization_tips": [
      "Current conditions are optimal for content delivery"
    ]
  }
}
```

### 2. Optimal Edge Location

```http
GET /api/v1/dual-grid/optimal-edge
```

**Parameters:**
- `user_location` (optional) - Auto-detected if not provided
- `cdn_provider` (required) - CDN provider
- `content_type` (optional) - Content type

**Response:**
```json
{
  "location": "Stockholm",
  "provider": "CloudFlare",
  "carbon_intensity": 45.2,
  "distance": 834.5,
  "estimated_latency": 52,
  "availability_score": 95.5
}
```

### 3. CDN Alternatives

```http
GET /api/v1/dual-grid/cdn-alternatives
```

**Parameters:**
- `user_location` (optional) - Auto-detected if not provided
- `current_edge` (required) - Current edge location
- `cdn_provider` (required) - CDN provider
- `content_type` (optional) - Content type
- `max_results` (optional) - Maximum alternatives (default: 5)

**Response:**
```json
[
  {
    "location": "Stockholm",
    "provider": "CloudFlare", 
    "carbon_intensity": 45.2,
    "distance": 834.5,
    "estimated_latency": 52,
    "availability_score": 95.5
  }
]
```

### 4. Supported CDN Providers

```http
GET /api/v1/dual-grid/cdn-providers
```

**Response:**
```json
{
  "supported_providers": ["cloudflare", "aws-cloudfront", "google-cloud", "azure"],
  "provider_details": {
    "cloudflare": {
      "name": "CloudFlare",
      "edge_locations_count": 35,
      "default_edge_selection": "geo_nearest",
      "carbon_aware_routing": false
    }
  },
  "total_count": 4
}
```

## Implementation Examples

### 1. Basic Dual-Grid Check

```javascript
const response = await fetch('/api/v1/dual-grid/carbon-intensity?' + 
  new URLSearchParams({
    edge_location: 'Frankfurt',
    content_type: 'video',
    cdn_provider: 'cloudflare'
  }));
const data = await response.json();

if (data.recommendation.action === 'proceed') {
  // Optimal conditions - use full quality
  enableHighQualityStreaming();
} else if (data.recommendation.action === 'optimize') {
  // Apply optimization suggestions
  data.recommendation.optimization_tips.forEach(tip => applyOptimization(tip));
} else if (data.recommendation.action === 'defer') {
  // High carbon intensity - defer or reduce quality
  scheduleForGreenWindow(data.recommendation.time_based_strategy);
}
```

### 2. CDN Edge Optimization

```javascript
// Find optimal edge location
const optimalEdge = await fetch('/api/v1/dual-grid/optimal-edge?' +
  new URLSearchParams({
    cdn_provider: 'aws-cloudfront',
    content_type: 'static'
  })).then(r => r.json());

// Reconfigure CDN to use optimal edge
await configureCDN({
  edgeLocation: optimalEdge.location,
  reason: `Carbon savings: ${optimalEdge.carbon_intensity}g CO2/kWh`
});
```

### 3. Content-Type Specific Optimization

```javascript
// Different strategies for different content types
const strategies = {
  video: async () => {
    const dual = await getDualGridIntensity('video');
    if (dual.weighted_intensity > 300) {
      return { quality: 'low', codec: 'av1', preload: 'none' };
    }
    return { quality: 'high', codec: 'h264', preload: 'metadata' };
  },
  
  api: async () => {
    const dual = await getDualGridIntensity('api');
    if (dual.weighted_intensity > 250) {
      return { caching: 'aggressive', batching: true, defer_analytics: true };
    }
    return { caching: 'normal', batching: false, defer_analytics: false };
  },
  
  static: async () => {
    const dual = await getDualGridIntensity('static');
    return {
      compression: dual.weighted_intensity > 200 ? 'brotli' : 'gzip',
      image_format: dual.weighted_intensity > 200 ? 'avif' : 'webp',
      ttl: dual.weighted_intensity < 150 ? '1year' : '1month'
    };
  }
};
```

## Configuration

### Environment Variables

```bash
# Required for real carbon intensity data
ELECTRICITY_MAPS_API_KEY=your_api_key_here

# Optional geolocation service configuration
IPAPI_BASE_URL=https://ipapi.co
GEOLOCATION_TIMEOUT=5s
GEOLOCATION_RATE_LIMIT=2
```

### CDN Provider Configuration

Each CDN provider can be extended with additional edge locations:

```go
"new-edge-location": {
    City: "New City", 
    Country: "Country", 
    GridZone: "GRID_ZONE",
    Latitude: 00.0000, 
    Longitude: 00.0000, 
    Tier: 1, 
    Capacity: "high",
    RenewableCommitment: true,
},
```

## Technical Details

### Distance Calculation

Uses the Haversine formula for accurate distance calculation between user and edge locations:

```go
func CalculateDistance(lat1, lon1, lat2, lon2 float64) float64 {
    const earthRadius = 6371 // km
    // Haversine formula implementation
    return earthRadius * c
}
```

### Network Hop Estimation

Estimates network hops based on distance:
- Base: 3 hops minimum
- Additional: 1 hop per 150km
- Maximum: 30 hops

### Carbon Intensity Weighting

The weighted calculation considers:
1. **Content Type** - Different content has different carbon profiles
2. **Transmission vs Computation** - Where the energy is consumed
3. **Distance Factor** - Longer transmission = higher carbon cost
4. **Edge Tier** - Primary edges are more efficient

### Edge Selection Algorithm

1. **Calculate Distance** - Haversine distance from user to edge
2. **Get Carbon Intensity** - Current grid intensity at edge location
3. **Compute Score** - `(intensity × 0.7) + (distance_factor × 0.3) × tier_penalty`
4. **Apply Bonuses** - Renewable commitment reduces score by 10%
5. **Sort and Return** - Lowest score = optimal edge

## Performance Considerations

- **Concurrent API Calls** - User and edge carbon intensity fetched in parallel
- **Caching** - Dual-grid results cached for 10 minutes
- **Rate Limiting** - Geolocation service rate limited to 2 RPS
- **Fallback Data** - Mock data when APIs unavailable
- **Timeout Handling** - 15-second timeout for dual-grid operations

## Future Enhancements

1. **Machine Learning** - Predict optimal switching times
2. **Real-time CDN API Integration** - Direct integration with CDN provider APIs
3. **Extended Grid Data** - More granular grid zone mappings
4. **Carbon Forecasting** - Predict future carbon intensity
5. **Custom Provider Support** - Framework for adding new CDN providers
6. **Advanced Routing** - Multi-hop routing optimization

## Testing

The implementation includes comprehensive test coverage:

```bash
# Run all tests
go test ./...

# Test specific dual-grid functionality
go test ./pkg/carbon -v
go test ./internal/geolocation -v 
go test ./service -v

# Test API endpoints
go test ./internal/handlers -v
```

## Monitoring and Analytics

Key metrics to monitor:

- **Carbon Savings** - Total CO2 saved through optimization
- **Edge Efficiency** - Usage of optimal vs suboptimal edges  
- **API Performance** - Response times for dual-grid endpoints
- **Cache Hit Rates** - Effectiveness of caching strategy
- **Provider Coverage** - Edge location utilization across CDNs

This implementation provides a comprehensive foundation for carbon-aware content delivery that can be extended and customized based on specific requirements and CDN provider capabilities.