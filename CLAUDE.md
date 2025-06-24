# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Development Commands

### Running the Application
```bash
# Start the Go API server
go run main.go

# Server runs on port 8090 by default
# Demo available at http://localhost:8090/demo
```

### Building and Testing
```bash
# Build the application
go build -o greenweb-api main.go

# Run tests
go test ./...

# Format code
go fmt ./...

# Vet code for potential issues  
go vet ./...

# Install dependencies
go mod tidy
```

## Architecture Overview

This is a Go-based API server that provides carbon intensity data and website optimization recommendations based on electricity grid carbon emissions.

### Core Components

**API Server (`main.go`)**
- Gin-based REST API with CORS middleware
- Mock carbon intensity data with time-based simulation
- Three main endpoints: carbon-intensity, optimization, green-hours
- Built-in demo dashboard at `/demo` with interactive features

**Data Models**
- `CarbonIntensity`: Real-time carbon data (g CO₂/kWh, renewable %, mode)
- `OptimizationProfile`: Website adaptation settings (image quality, disabled features, eco discounts)
- `GreenHoursForecast`: Predictions for optimal low-carbon periods

**JavaScript SDK (`sdk/greenweb.js`)**
- Client-side library for automatic website optimization
- React hook support for easy integration
- Real-time carbon monitoring with configurable update intervals
- Automatic feature disabling and eco-discount banners

### Key Patterns

**Carbon Intensity Thresholds**
- Green mode (< 150g CO₂/kWh): Full features, eco discounts
- Yellow mode (150-300g): Moderate optimization 
- Red mode (> 300g): Maximum optimization, defer heavy features

**Time-Based Mock Data**
- Night hours (22:00-06:00): Low carbon intensity (green mode)
- Peak hours (12:00-16:00): High carbon intensity (red mode)
- Other hours: Medium carbon intensity (yellow mode)

**SDK Auto-optimization**
- Applies CSS classes based on carbon mode
- Hides elements with `data-greenweb-feature` attributes
- Lazy loads images in low-carbon mode
- Shows eco-discount banners during green hours

## Project Structure

```
.
├── main.go                    # Go API server with all endpoints
├── go.mod                     # Go module dependencies
├── sdk/greenweb.js           # JavaScript client SDK
├── examples/                  # Integration examples
│   └── shopify-integration.liquid
├── README.md                  # Project overview and API docs
├── DEVELOPMENT_PLAN.md        # Detailed roadmap and technical plans
└── PROJECT_INFO.md           # Project separation notes
```

## Environment Configuration

The server reads from environment variables:
- `PORT`: Server port (default: 8090)
- API expects future integration with Electricity Maps API

## Integration Examples

The codebase includes Shopify Liquid template integration showing how e-commerce sites can implement dynamic pricing and feature management based on carbon intensity.