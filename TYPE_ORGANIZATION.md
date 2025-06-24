# GreenWeb API Type Organization

This document describes the comprehensive type organization implemented in the GreenWeb API codebase. The types have been restructured to provide clear separation between public and internal APIs, maintain backward compatibility, and follow Go best practices.

## Directory Structure

```
greenweb/
├── pkg/                          # Public API types (external consumers)
│   ├── carbon/                   # Carbon intensity types and interfaces
│   │   ├── types.go             # CarbonIntensity, GreenHour, GreenHoursForecast
│   │   └── interfaces.go        # CarbonService interface
│   └── optimization/            # Optimization types and interfaces
│       ├── types.go            # OptimizationProfile, rules, requests/responses
│       └── interfaces.go       # OptimizationService interface
├── internal/                     # Internal implementation types
│   └── types/                   # Internal shared types
│       ├── errors.go           # Custom error types with proper wrapping
│       └── responses.go        # HTTP response wrappers and pagination
└── service/                     # Service implementations (backward compatible)
    ├── electricity_maps.go     # Uses pkg/carbon types with aliases
    └── optimization.go         # Uses pkg/optimization types with aliases
```

## Public Types (`pkg/`)

### Carbon Package (`pkg/carbon/`)

**Purpose**: Provides stable, public types for carbon intensity monitoring and green energy forecasting.

#### Core Types

1. **`CarbonIntensity`** - Represents current carbon intensity data
   - Carbon intensity in g CO2/kWh
   - Renewable energy percentage
   - Mode classification (green/yellow/red)
   - Location and timestamp information
   - Helper methods: `IsGreen()`, `IsOptimal()`, `ShouldDefer()`, `GetEfficiencyRating()`

2. **`GreenHour`** - Represents a forecasted green hour window
   - Start and end times
   - Predicted carbon intensity and renewable percentage
   - Confidence level and duration
   - Helper methods: `IsActive()`, `IsUpcoming()`, `TimeUntilStart()`

3. **`GreenHoursForecast`** - Collection of green hours with metadata
   - Array of green hours sorted by time
   - Best window identification
   - Forecast period and generation metadata
   - Helper methods: `GetActiveGreenHours()`, `GetNextGreenHour()`, `GetTotalGreenDuration()`

4. **`CarbonIntensityThresholds`** - Configurable thresholds for classification
   - Green, yellow, and red thresholds
   - Methods for classification and recommendations

#### Interfaces

1. **`CarbonService`** - Core carbon intensity service interface
   - `GetCarbonIntensity(ctx, location) (*CarbonIntensity, error)`
   - `GetGreenHoursForecast(ctx, location, hours) (*GreenHoursForecast, error)`
   - `IsHealthy(ctx) bool`
   - `GetSupportedLocations(ctx) ([]string, error)`

2. **`CarbonServiceWithCache`** - Extended interface with caching
   - All CarbonService methods
   - Cache management methods

3. **`CarbonServiceWithHistory`** - Extended interface with historical data
   - All CarbonService methods
   - Historical data retrieval methods

4. **`CarbonWebhookService`** - Interface for webhook notifications
   - Webhook registration and management

### Optimization Package (`pkg/optimization/`)

**Purpose**: Provides stable, public types for website optimization based on carbon intensity.

#### Core Types

1. **`OptimizationProfile`** - Defines how a website should adapt
   - Mode (full/normal/eco/critical)
   - Features to disable
   - Quality settings (image/video)
   - Caching strategy
   - UI and content optimizations
   - Resource limits and metadata

2. **`OptimizationRequest`** - Request for optimization recommendations
   - Location and URL
   - Device and connection information
   - Custom thresholds and preferences
   - Available features list

3. **`OptimizationResponse`** - Complete optimization response
   - Carbon intensity data
   - Generated optimization profile
   - Human-readable recommendations
   - Request metadata

4. **`OptimizationRule`** - Rule-based optimization logic
   - Conditions and actions
   - Priority and enablement status
   - Rule evaluation methods

#### Enums and Constants

- **`OptimizationMode`**: full, normal, eco, critical
- **`ImageQuality`**: high, medium, low
- **`VideoQuality`**: 4k, 1080p, 720p, 480p, 360p
- **`CachingStrategy`**: minimal, normal, aggressive

#### Interfaces

1. **`OptimizationService`** - Core optimization service interface
   - `GetOptimizationProfile(ctx, request) (*OptimizationResponse, error)`
   - `GetOptimizationRecommendations(ctx, request) ([]string, error)`
   - `ValidateOptimizationProfile(ctx, profile) (bool, error)`

2. **`OptimizationServiceWithRules`** - Extended interface with rule management
   - All OptimizationService methods
   - Rule CRUD operations
   - Rule evaluation methods

3. **`OptimizationServiceWithAnalytics`** - Extended interface with analytics
   - All OptimizationService methods
   - Optimization tracking and statistics
   - Energy savings reporting

## Internal Types (`internal/types/`)

### Error Types (`internal/types/errors.go`)

**Purpose**: Provides structured error handling with proper context and wrapping.

#### Core Types

1. **`GreenWebError`** - Structured error with context
   - Error code for categorization
   - Human-readable message and details
   - Underlying cause error
   - Request ID and location context
   - HTTP status code and retry information
   - Metadata for additional context

2. **`ErrorResponse`** - Standardized API error response
   - Wrapped GreenWebError
   - Request tracking information
   - Timestamp and path information

3. **`MultiError`** - Collection of multiple errors
   - Array of GreenWebError instances
   - Error counting and access methods

#### Error Codes

Comprehensive error codes for different categories:
- Carbon intensity errors: `CARBON_INTENSITY_FETCH_ERROR`, `CARBON_INTENSITY_TIMEOUT`
- Optimization errors: `OPTIMIZATION_GENERATION_ERROR`, `OPTIMIZATION_RULE_ERROR`
- Location errors: `LOCATION_INVALID`, `GEOLOCATION_FAILED`
- External API errors: `EXTERNAL_API_ERROR`, `EXTERNAL_API_TIMEOUT`, `EXTERNAL_API_RATE_LIMIT`
- Configuration errors: `CONFIGURATION_ERROR`, `SERVICE_UNAVAILABLE`
- Validation errors: `VALIDATION_ERROR`, `INVALID_REQUEST`

#### Helper Functions

Factory functions for common error types:
- `NewCarbonIntensityError()`, `NewCarbonIntensityTimeoutError()`
- `NewOptimizationError()`, `NewLocationError()`
- `NewExternalAPIError()`, `NewValidationError()`
- Error collectors for batch operations

### Response Types (`internal/types/responses.go`)

**Purpose**: Provides standardized HTTP response wrappers with metadata.

#### Core Types

1. **`APIResponse[T]`** - Generic response wrapper
   - Typed data payload
   - Success indicator and message
   - Request ID and timestamp
   - API version and metadata

2. **`ResponseMetadata`** - Additional response information
   - Processing time and cache information
   - Data source and location
   - Rate limiting information
   - Non-fatal warnings

3. **`HealthResponse`** - Health check response
   - Overall status and service information
   - Individual component health checks
   - Uptime and version information

4. **`PaginatedResponse[T]`** - Paginated data response
   - Typed data array
   - Pagination metadata
   - Total item count

#### Specialized Responses

- **`CarbonIntensityResponse`** - Carbon intensity with API metadata
- **`GreenHoursForecastResponse`** - Forecast with API metadata
- **`OptimizationResponse`** - Optimization with API metadata
- **`BatchResponse[T]`** - Batch operation results
- **`StreamResponse`** - Server-sent events

## Backward Compatibility

### Service Layer Aliases

The service layer maintains full backward compatibility through type aliases:

```go
// In service/electricity_maps.go
type CarbonIntensity = carbon.CarbonIntensity
type GreenHour = carbon.GreenHour
type GreenHoursForecast = carbon.GreenHoursForecast

// In service/optimization.go
type OptimizationProfile = optimization.OptimizationProfile
type OptimizationRequest = optimization.OptimizationRequest
type OptimizationResponse = optimization.OptimizationResponse
```

### Enhanced Implementations

While maintaining compatibility, the implementations have been enhanced:

1. **Additional Fields**: New optional fields added to existing types
2. **Helper Methods**: Convenience methods added to all types
3. **Better Validation**: Comprehensive validation tags
4. **Rich Metadata**: Extended metadata for tracking and debugging

## Migration Guide

### For External SDK Users

**No breaking changes** - All existing code continues to work:

```go
// This continues to work
client := service.NewElectricityMapsClient(logger)
intensity, err := client.GetCarbonIntensity(ctx, "Berlin")
```

**Recommended migration** to use new stable types:

```go
// New recommended approach
import "github.com/perschulte/greenweb-api/pkg/carbon"

var carbonService carbon.CarbonService = client
intensity, err := carbonService.GetCarbonIntensity(ctx, "Berlin")
```

### For Internal Development

**Use new type organization** for all new code:

```go
import (
    "github.com/perschulte/greenweb-api/pkg/carbon"
    "github.com/perschulte/greenweb-api/pkg/optimization"
    "github.com/perschulte/greenweb-api/internal/types"
)
```

### Error Handling Migration

**Old approach**:
```go
if err != nil {
    c.JSON(500, gin.H{"error": "Something went wrong"})
}
```

**New approach**:
```go
if err != nil {
    gwErr := types.NewCarbonIntensityError("Failed to fetch data", err)
    response := types.NewErrorResponse(gwErr, requestID)
    c.JSON(gwErr.HTTPStatus, response)
}
```

## Go 1.21+ Features Used

1. **Type Parameters (Generics)**:
   - `APIResponse[T]` for typed responses
   - `PaginatedResponse[T]` for typed pagination
   - `BatchResponse[T]` for typed batch operations

2. **Type Constraints**:
   - Interface constraints for service implementations
   - Validation constraints using struct tags

3. **Enhanced Error Handling**:
   - Error wrapping with context preservation
   - Structured error hierarchies

## Validation and JSON Support

All public types include:

1. **JSON Tags**: Proper serialization with examples
2. **Validation Tags**: Comprehensive validation rules
3. **Documentation**: Extensive godoc comments
4. **Examples**: Usage examples in struct tags

Example:
```go
type CarbonIntensity struct {
    Location string `json:"location" validate:"required" example:"Berlin"`
    CarbonIntensity float64 `json:"carbon_intensity" validate:"min=0" example:"120.5"`
    // ...
}
```

## Testing Considerations

The new type organization supports:

1. **Interface-Based Testing**: Mock implementations of all interfaces
2. **Comprehensive Validation**: Test all validation scenarios
3. **Error Scenarios**: Test all error types and codes
4. **Backward Compatibility**: Ensure no breaking changes

## Performance Implications

1. **Zero-Cost Abstractions**: Type aliases have no runtime overhead
2. **Efficient Serialization**: Optimized JSON marshaling/unmarshaling
3. **Memory Efficiency**: Struct packing and minimal allocations
4. **Caching Support**: Built-in cache key generation and TTL support

## Security Considerations

1. **Input Validation**: Comprehensive validation for all external inputs
2. **Error Information**: Careful error message design to avoid information leakage
3. **Rate Limiting**: Built-in rate limiting types and interfaces
4. **Audit Trail**: Request ID tracking throughout the system

This type organization provides a solid foundation for the GreenWeb API that can scale with future requirements while maintaining stability for external consumers.