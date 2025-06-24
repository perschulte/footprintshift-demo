package handlers

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

// DemoHandler handles demo dashboard endpoints
type DemoHandler struct {
	logger *slog.Logger
	config *Config
}

// NewDemoHandler creates a new demo handler with dependencies
func NewDemoHandler(deps *Dependencies) *DemoHandler {
	return &DemoHandler{
		logger: deps.Logger,
		config: deps.Config,
	}
}

// HandleDemo serves the interactive demo dashboard
func (h *DemoHandler) HandleDemo(c *gin.Context) {
	const operation = "serve_demo"
	
	// Log the incoming request
	LogRequest(h.logger, c, operation, nil)
	
	// Generate the HTML content for the demo dashboard
	html := h.generateDemoHTML()
	
	// Log successful response
	LogResponse(h.logger, operation, http.StatusOK, nil)
	
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
}

// generateDemoHTML creates the demo dashboard HTML content
func (h *DemoHandler) generateDemoHTML() string {
	return `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>GreenWeb Demo - Adaptive Website</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            margin: 0;
            padding: 20px;
            transition: all 0.3s ease;
        }
        .container {
            max-width: 1200px;
            margin: 0 auto;
        }
        .header {
            display: flex;
            justify-content: space-between;
            align-items: center;
            padding: 20px;
            background: #f5f5f5;
            border-radius: 10px;
            margin-bottom: 30px;
        }
        .carbon-display {
            font-size: 24px;
            font-weight: bold;
            padding: 10px 20px;
            border-radius: 5px;
            color: white;
        }
        .carbon-green { background: #22c55e; }
        .carbon-yellow { background: #eab308; }
        .carbon-red { background: #ef4444; }
        .features {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
            gap: 20px;
            margin-top: 30px;
        }
        .feature {
            border: 2px solid #e5e5e5;
            padding: 20px;
            border-radius: 10px;
            transition: all 0.3s ease;
        }
        .feature.disabled {
            opacity: 0.5;
            filter: grayscale(100%);
        }
        .eco-banner {
            background: #22c55e;
            color: white;
            padding: 15px;
            text-align: center;
            border-radius: 10px;
            margin: 20px 0;
            font-size: 18px;
            animation: pulse 2s infinite;
        }
        @keyframes pulse {
            0% { opacity: 0.8; }
            50% { opacity: 1; }
            100% { opacity: 0.8; }
        }
        .product {
            border: 1px solid #ddd;
            padding: 20px;
            border-radius: 10px;
            text-align: center;
        }
        .price {
            font-size: 24px;
            font-weight: bold;
            margin: 10px 0;
        }
        .discount {
            color: #22c55e;
            text-decoration: line-through;
        }
        body.eco-mode {
            background: #f3f4f6;
        }
        body.eco-mode img {
            filter: contrast(0.8);
        }
        .api-info {
            background: #f9fafb;
            padding: 20px;
            border-radius: 10px;
            margin-top: 30px;
        }
        .status-indicator {
            display: inline-block;
            width: 12px;
            height: 12px;
            border-radius: 50%;
            margin-right: 8px;
        }
        .status-ok { background: #22c55e; }
        .status-warning { background: #eab308; }
        .status-error { background: #ef4444; }
        .error-message {
            background: #fee2e2;
            color: #dc2626;
            padding: 15px;
            border-radius: 10px;
            margin: 20px 0;
            border: 1px solid #fecaca;
        }
        .loading {
            opacity: 0.6;
            pointer-events: none;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🌍 GreenWeb Demo Shop</h1>
            <div class="carbon-display" id="carbonDisplay">
                Loading...
            </div>
        </div>

        <div class="api-info">
            <h3>🔌 API Status</h3>
            <div id="apiStatus">
                <span class="status-indicator status-warning"></span>
                <span>Checking API status...</span>
            </div>
        </div>

        <div id="errorMessage" class="error-message" style="display: none;"></div>

        <div id="ecoBanner" class="eco-banner" style="display: none;">
            🌱 Green Hour Active! Optimal time for full website functionality.
        </div>

        <div class="features">
            <div class="feature" id="videoFeature">
                <h3>📹 Product Videos</h3>
                <p>High-quality product demonstrations</p>
                <video width="100%" controls id="productVideo">
                    <source src="demo.mp4" type="video/mp4">
                    Your browser does not support the video tag.
                </video>
            </div>

            <div class="feature" id="aiFeature">
                <h3>🤖 AI Recommendations</h3>
                <p>Personalized product suggestions based on your preferences</p>
                <button onclick="showAIRecommendations()">Get Recommendations</button>
            </div>

            <div class="feature" id="3dFeature">
                <h3>🎮 3D Product View</h3>
                <p>Interactive 3D models of products</p>
                <div style="height: 200px; background: #e5e5e5; display: flex; align-items: center; justify-content: center;">
                    <span id="3dModelText">3D Model Viewer</span>
                </div>
            </div>
        </div>

        <h2>Featured Products</h2>
        <div class="features">
            <div class="product">
                <h3>Eco Laptop</h3>
                <img src="https://via.placeholder.com/200x150/22c55e/ffffff?text=Eco+Laptop" alt="Laptop" id="productImage1">
                <p class="price" id="price1">€999</p>
                <button onclick="addToCart('Eco Laptop')">Add to Cart</button>
            </div>
            <div class="product">
                <h3>Solar Charger</h3>
                <img src="https://via.placeholder.com/200x150/eab308/ffffff?text=Solar+Charger" alt="Charger" id="productImage2">
                <p class="price" id="price2">€49</p>
                <button onclick="addToCart('Solar Charger')">Add to Cart</button>
            </div>
        </div>

        <div style="margin-top: 40px; padding: 20px; background: #f9fafb; border-radius: 10px;">
            <h3>Current Optimization Settings</h3>
            <pre id="optimizationData" style="background: white; padding: 15px; border-radius: 5px; font-size: 12px; overflow-x: auto;"></pre>
        </div>
    </div>

    <script>
        let isLoading = false;

        function showError(message) {
            const errorEl = document.getElementById('errorMessage');
            errorEl.textContent = message;
            errorEl.style.display = 'block';
            setTimeout(() => {
                errorEl.style.display = 'none';
            }, 5000);
        }

        function setLoading(loading) {
            isLoading = loading;
            document.body.classList.toggle('loading', loading);
        }

        function showAIRecommendations() {
            if (document.getElementById('aiFeature').classList.contains('disabled')) {
                alert('🌱 AI Features are disabled during high carbon periods to save energy!');
            } else {
                alert('🤖 AI analyzing your preferences... Recommended: Eco-friendly products!');
            }
        }

        function addToCart(product) {
            alert('🛒 Added ' + product + ' to cart!');
        }

        async function checkApiStatus() {
            try {
                const response = await fetch('/health');
                if (!response.ok) {
                    throw new Error('Health check failed');
                }
                
                const health = await response.json();
                const statusEl = document.getElementById('apiStatus');
                
                if (health.electricity_maps_api) {
                    statusEl.innerHTML = '<span class="status-indicator status-ok"></span>Electricity Maps API: Connected';
                } else if (health.api_key_configured) {
                    statusEl.innerHTML = '<span class="status-indicator status-warning"></span>Electricity Maps API: Key configured, but API unreachable (using fallback)';
                } else {
                    statusEl.innerHTML = '<span class="status-indicator status-warning"></span>Electricity Maps API: No API key configured (using mock data)';
                }
            } catch (error) {
                console.error('API status check failed:', error);
                document.getElementById('apiStatus').innerHTML = '<span class="status-indicator status-error"></span>API Status: Unavailable';
            }
        }

        async function updateCarbonIntensity() {
            if (isLoading) return;
            
            try {
                setLoading(true);
                
                // Fetch current carbon intensity
                const carbonResponse = await fetch('/api/v1/carbon-intensity?location=Berlin');
                if (!carbonResponse.ok) {
                    throw new Error('Failed to fetch carbon intensity');
                }
                const carbonData = await carbonResponse.json();
                
                // Fetch optimization profile
                const optResponse = await fetch('/api/v1/optimization?location=Berlin&url=demo-shop.com');
                if (!optResponse.ok) {
                    throw new Error('Failed to fetch optimization profile');
                }
                const optData = await optResponse.json();
                
                // Update carbon display
                const display = document.getElementById('carbonDisplay');
                display.textContent = Math.round(carbonData.carbon_intensity) + ' g CO₂/kWh';
                display.className = 'carbon-display carbon-' + carbonData.mode;
                
                // Apply optimizations
                const opt = optData.optimization;
                document.getElementById('optimizationData').textContent = JSON.stringify(opt, null, 2);
                
                // Update body class
                document.body.className = opt.mode === 'eco' ? 'eco-mode' : '';
                
                // Show/hide eco banner
                const showBanner = opt.show_green_banner || carbonData.mode === 'green';
                document.getElementById('ecoBanner').style.display = showBanner ? 'block' : 'none';
                
                // Reset all features first
                document.querySelectorAll('.feature').forEach(f => f.classList.remove('disabled'));
                
                // Disable features based on optimization
                opt.disable_features.forEach(feature => {
                    if (feature === 'video_autoplay') {
                        const video = document.getElementById('productVideo');
                        video.removeAttribute('autoplay');
                        video.muted = true;
                    }
                    if (feature === '3d_models') {
                        const feature3d = document.getElementById('3dFeature');
                        feature3d.classList.add('disabled');
                        document.getElementById('3dModelText').textContent = '3D Models Disabled (Eco Mode)';
                    }
                    if (feature === 'ai_features') {
                        document.getElementById('aiFeature').classList.add('disabled');
                    }
                });
                
                // Update image quality
                document.querySelectorAll('img').forEach(img => {
                    img.style.filter = '';
                    if (opt.image_quality === 'low') {
                        img.style.filter = 'blur(0.5px) contrast(0.9)';
                        img.loading = 'lazy';
                    }
                });
                
                // Apply eco discount
                if (opt.eco_discount > 0) {
                    const newPrice1 = Math.round(999 * (1 - opt.eco_discount/100));
                    const newPrice2 = Math.round(49 * (1 - opt.eco_discount/100));
                    document.getElementById('price1').innerHTML = '<span class="discount">€999</span> €' + newPrice1;
                    document.getElementById('price2').innerHTML = '<span class="discount">€49</span> €' + newPrice2;
                } else {
                    document.getElementById('price1').textContent = '€999';
                    document.getElementById('price2').textContent = '€49';
                }
                
            } catch (error) {
                console.error('Error updating carbon intensity:', error);
                showError('Failed to update carbon data: ' + error.message);
                document.getElementById('carbonDisplay').textContent = 'Error';
            } finally {
                setLoading(false);
            }
        }

        // Initialize
        document.addEventListener('DOMContentLoaded', function() {
            checkApiStatus();
            updateCarbonIntensity();
            
            // Update every 30 seconds
            setInterval(updateCarbonIntensity, 30000);
            setInterval(checkApiStatus, 60000);
        });
    </script>
</body>
</html>
	`
}