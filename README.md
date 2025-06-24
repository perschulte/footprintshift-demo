# GreenWeb API 🌍

Adaptive website optimization based on real-time carbon intensity. Make the web greener by automatically adjusting features, quality, and even pricing based on renewable energy availability.

## 🎯 Mission

Enable websites to dynamically adapt their content and features based on the current carbon intensity of electricity in the user's location. When renewable energy is abundant, unlock premium features. When the grid is dirty, reduce computational load and incentivize users to come back during "green hours."

## 🌟 Key Features

- **Real-time Carbon Intensity** - Location-based CO₂ emissions data
- **Smart Optimization Profiles** - Automatic recommendations for websites
- **Green Hours Pricing** - Dynamic discounts during renewable energy peaks
- **Performance Adaptation** - Adjust image quality, disable features based on carbon intensity
- **Analytics Dashboard** - Track your green impact and user behavior

## 🚀 Quick Start

```javascript
// Install the SDK
npm install greenweb-sdk

// Basic usage
import { GreenWeb } from 'greenweb-sdk';

const gw = new GreenWeb({ apiKey: 'your-api-key' });

// Get current carbon intensity
const { intensity, mode } = await gw.getCarbonIntensity();

// Apply optimizations
if (mode === 'high-carbon') {
  // Reduce features, offer green hours discount
  showBanner('🌱 Come back tonight for 10% off when energy is cleaner!');
  document.body.classList.add('eco-mode');
}
```

## ⚙️ Configuration

### Environment Variables

Create a `.env` file in your project root:

```bash
# Server Configuration
PORT=8090

# Electricity Maps API Integration
# Get your free API key from: https://api-portal.electricitymaps.com/
ELECTRICITY_MAPS_API_KEY=your_api_key_here

# Optional: Leave empty to use mock data for development
# ELECTRICITY_MAPS_API_KEY=
```

### Electricity Maps API Integration

This API now integrates with **Electricity Maps** to provide real-time carbon intensity data from over 200 regions worldwide. The integration includes:

- **Real-time Data**: Live carbon intensity in gCO₂eq/kWh
- **Global Coverage**: 200+ regions and 50+ countries
- **Graceful Fallback**: Automatically falls back to intelligent mock data if API is unavailable
- **Rate Limiting**: Built-in HTTP client with proper timeouts and error handling

#### Getting an API Key

1. Visit the [Electricity Maps API Portal](https://api-portal.electricitymaps.com/)
2. Sign up for a free account
3. Choose the **Free Tier** for development (includes carbon intensity data)
4. Copy your API key to the `ELECTRICITY_MAPS_API_KEY` environment variable

#### API Features Used

- **Carbon Intensity**: Real-time gCO₂eq/kWh for any supported region
- **Renewable Percentage**: Calculated from fossil fuel percentage
- **Smart Fallback**: Mock data with realistic patterns when API unavailable
- **Location Mapping**: Intelligent mapping of location names to country/zone codes

### Running the API

```bash
# Install dependencies
go mod tidy

# Build the application
go build -o greenweb .

# Run with environment variables
./greenweb

# Or run directly with go
go run .
```

The API will start on port 8090 by default. Visit `http://localhost:8090/demo` to see the interactive demo.

## 📊 Example Use Cases

### E-Commerce: Green Hours Pricing
```javascript
// Shopify integration
if (carbonIntensity < 150) {
  applyDiscount('GREEN5'); // 5% off during green hours
  enablePremiumFeatures();
}
```

### Media: Adaptive Streaming
```javascript
// Adjust video quality based on carbon intensity
const quality = carbonIntensity < 200 ? '1080p' : '480p';
player.setQuality(quality);
```

### SaaS: Carbon-Aware Computing
```javascript
// Defer non-critical tasks
if (carbonIntensity > 300) {
  scheduleJob('data-processing', nextGreenWindow);
}
```

## 🏗️ Architecture

```
┌─────────────────┐     ┌──────────────────┐     ┌─────────────────┐
│   Website/App   │────▶│  GreenWeb API    │────▶│ Electricity Maps│
└─────────────────┘     └──────────────────┘     └─────────────────┘
                               │
                               ▼
                        ┌──────────────────┐
                        │ Optimization     │
                        │ Engine           │
                        └──────────────────┘
```

## 🔧 API Endpoints

### Health Check
```
GET /health
```
Returns API status and Electricity Maps API connectivity.

### Get Carbon Intensity
```
GET /api/v1/carbon-intensity?location=Berlin
```
Returns real-time carbon intensity data for the specified location.

**Response Example:**
```json
{
  "location": "Berlin",
  "carbon_intensity": 106,
  "renewable_percentage": 91.7,
  "mode": "green",
  "recommendation": "optimal",
  "next_green_window": "2025-06-24T17:46:29Z",
  "timestamp": "2025-06-24T13:00:00Z"
}
```

### Get Optimization Profile
```
GET /api/v1/optimization?location=Berlin&url=shop.example.com
```
Returns both carbon intensity and optimization recommendations for the website.

**Response Example:**
```json
{
  "carbon_intensity": {
    "location": "Berlin",
    "carbon_intensity": 106,
    "renewable_percentage": 91.7,
    "mode": "green",
    "recommendation": "optimal"
  },
  "optimization": {
    "mode": "full",
    "disable_features": [],
    "image_quality": "high",
    "video_quality": "1080p",
    "defer_analytics": false,
    "eco_discount": 5,
    "show_green_banner": true,
    "caching_strategy": "normal"
  },
  "url": "shop.example.com"
}
```

### Get Green Hours Forecast
```
GET /api/v1/green-hours?location=Berlin&next=24
```
Returns forecast of optimal low-carbon hours (next 1-168 hours).

### Demo Dashboard
```
GET /demo
```
Interactive demo showing real-time adaptation based on carbon intensity.

## 🌱 Impact Metrics

Track your environmental impact:
- CO₂ saved by deferrals
- User sessions during green hours
- Features adapted based on carbon intensity
- Revenue impact of green pricing

## 🤝 Contributing

Join us in making the web more sustainable! See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## 📜 License

MIT License - See [LICENSE](LICENSE) for details.

---

**Together, we can make every click count for the climate.** 🌍