# 🔌 FootprintShift Service Integration Guide

So integrieren Sie den FootprintShift Service in Ihre Website oder App:

## 🚀 **Schnelle Integration (JavaScript)**

### **1. Script Tag (Einfachste Methode)**
```html
<!-- In Ihrem HTML Head -->
<script src="https://api.footprintshift.dev/sdk/footprintshift.js" 
        data-footprintshift-api-key="your-api-key" 
        data-footprintshift-location="auto">
</script>

<!-- Automatische Optimierung läuft im Hintergrund -->
```

### **2. NPM Package**
```bash
npm install footprintshift-sdk
```

```javascript
import { FootprintShift } from 'footprintshift-sdk';

const fps = new FootprintShift({
    apiKey: 'your-api-key',
    location: 'auto', // oder 'Berlin', 'Poland', etc.
    updateInterval: 60000 // 1 Minute
});

// Event Listener für Optimierungen
fps.onModeChange((mode, data) => {
    console.log('Footprint mode changed to:', mode);
    if (mode === 'red') {
        // High environmental impact - apply optimizations
        disableHeavyFeatures();
    }
});
```

## 🎯 **Content-Type Spezifische Integration**

### **E-Commerce (Shopify, WooCommerce)**
```javascript
// Produktseite Optimierung
const { optimization } = await fps.getOptimization({
    url: window.location.href,
    contentType: 'ecommerce'
});

if (optimization.mode === 'eco') {
    // Video auf 720p reduzieren
    document.querySelectorAll('video').forEach(video => {
        video.style.maxHeight = '720px';
    });
    
    // 3D Produktviewer deaktivieren
    document.querySelectorAll('[data-3d-viewer]').forEach(viewer => {
        viewer.style.display = 'none';
    });
    
    // Eco-Discount anzeigen
    if (optimization.eco_discount > 0) {
        showEcoDiscount(optimization.eco_discount);
    }
}
```

### **Video Platforms (YouTube, Netflix-like)**
```javascript
// Video Quality Anpassung
const carbonData = await fps.getCarbonIntensity();

let maxQuality = '4K';
if (carbonData.carbon_intensity > 300) {
    maxQuality = '720p'; // 24g CO₂/h Einsparung
} else if (carbonData.carbon_intensity > 150) {
    maxQuality = '1080p'; // 12g CO₂/h Einsparung
}

videoPlayer.setMaxQuality(maxQuality);
```

### **AI/LLM Platforms (ChatGPT-like)**
```javascript
// AI Inference Deferral
const { optimization } = await fps.getOptimization({
    contentType: 'ai_platform'
});

if (optimization.ai_deferred_to_green_window) {
    // Zeige Nutzer wann AI wieder verfügbar ist
    const nextGreenWindow = optimization.next_green_window;
    showMessage(`AI features available again at ${nextGreenWindow} when energy is cleaner. Save 3g CO₂ per session!`);
    
    // Disable heavy AI features
    disableAIFeatures();
}
```

### **Gaming Platforms**
```javascript
// GPU Feature Management
if (optimization.gpu_features_disabled) {
    // WebGL/3D Features deaktivieren (15g CO₂/h Einsparung)
    disableWebGL();
    disable3DModels();
    setFrameRateLimit(30); // statt 60fps
    
    showNotification('Graphics optimized for cleaner energy. Full quality available during green hours.');
}
```

## 🔧 **Backend Integration (APIs)**

### **Node.js/Express**
```javascript
const express = require('express');
const axios = require('axios');

app.get('/api/content', async (req, res) => {
    // Carbon intensity für User-Location abrufen
    const carbonData = await axios.get('https://api.footprintshift.dev/v1/carbon-intensity', {
        params: { location: req.headers['cf-ipcountry'] || 'auto' }
    });
    
    // Content basierend auf Carbon intensity anpassen
    let content = getBaseContent();
    
    if (carbonData.data.mode === 'red') {
        // High carbon - optimized content
        content.videos = content.videos.map(video => ({
            ...video,
            quality: '720p',
            preload: 'none' // Lazy loading
        }));
        
        content.images = content.images.map(img => ({
            ...img,
            format: 'webp',
            quality: 70
        }));
    }
    
    res.json(content);
});
```

### **Python/Django**
```python
import requests
from django.http import JsonResponse

def get_optimized_content(request):
    # Carbon intensity abrufen
    response = requests.get('https://api.footprintshift.dev/v1/optimization', {
        'params': {
            'location': get_user_location(request),
            'url': request.get_host()
        }
    })
    
    optimization = response.json()['optimization']
    
    # Content anpassen
    if optimization['mode'] == 'eco':
        # Reduzierte Qualität für hohen Carbon footprint
        return serve_optimized_content(optimization)
    else:
        return serve_full_content()
```

## 🎨 **CSS-Based Optimization (Progressive Enhancement)**

```css
/* Automatische Anpassung via CSS Custom Properties */
:root {
    --footprintshift-mode: normal; /* wird vom SDK gesetzt */
}

/* Video Optimierung */
video {
    max-height: var(--video-max-height, 1080px);
    preload: var(--video-preload, metadata);
}

/* Bei high footprint: reduzierte Features */
[data-footprintshift-mode="eco"] video {
    --video-max-height: 720px;
    --video-preload: none;
}

[data-footprintshift-mode="eco"] .gpu-intensive {
    display: none !important;
}

[data-footprintshift-mode="eco"] .ai-features {
    opacity: 0.5;
    pointer-events: none;
}

/* Green mode: premium features */
[data-footprintshift-mode="green"] .eco-discount {
    display: block;
}
```

## 📱 **Mobile App Integration (React Native)**

```javascript
import { FootprintShiftSDK } from 'footprintshift-react-native';

const App = () => {
    const { carbonData, optimization } = useFootprintShift({
        apiKey: 'your-key',
        location: 'auto'
    });
    
    return (
        <View>
            {carbonData.mode === 'green' && (
                <EcoDiscountBanner discount={optimization.eco_discount} />
            )}
            
            <VideoPlayer 
                maxQuality={getVideoQuality(carbonData.carbon_intensity)}
                autoPlay={!optimization.disable_features.includes('video_autoplay')}
            />
            
            {!optimization.ai_deferred_to_green_window && (
                <AIFeatures />
            )}
        </View>
    );
};
```

## 🏪 **E-Commerce Platform Plugins**

### **Shopify App**
```liquid
<!-- Shopify Liquid Template -->
{% assign footprintshift_mode = 'normal' %}
{% if footprintshift_carbon_intensity > 300 %}
    {% assign footprintshift_mode = 'eco' %}
{% endif %}

<!-- Product Videos -->
{% if footprintshift_mode == 'eco' %}
    <video poster="{{ product.featured_image | img_url: 'master' }}" preload="none">
        <source src="{{ product.video_720p }}" type="video/mp4">
    </video>
    <div class="eco-savings">Video optimized - saving 24g CO₂/hour</div>
{% else %}
    <video autoplay loop>
        <source src="{{ product.video_4k }}" type="video/mp4">
    </video>
{% endif %}

<!-- Eco Discount -->
{% if footprintshift_mode == 'green' and footprintshift_eco_discount > 0 %}
    <div class="eco-discount">
        🌱 Green Hour Special: {{ footprintshift_eco_discount }}% off!
    </div>
{% endif %}
```

### **WordPress Plugin**
```php
<?php
class FootprintShiftWP {
    public function __construct() {
        add_action('wp_enqueue_scripts', array($this, 'enqueue_footprintshift'));
        add_filter('the_content', array($this, 'optimize_content'));
    }
    
    public function optimize_content($content) {
        $carbon_data = $this->get_carbon_intensity();
        
        if ($carbon_data['mode'] === 'red') {
            // Video tags auf 720p begrenzen
            $content = preg_replace_callback('/<video[^>]*>.*?<\/video>/s', 
                array($this, 'optimize_video'), $content);
                
            // Lazy loading für Bilder
            $content = str_replace('<img ', '<img loading="lazy" ', $content);
        }
        
        return $content;
    }
}

new FootprintShiftWP();
?>
```

## 📊 **Analytics & Monitoring**

### **Impact Tracking**
```javascript
// CO₂ Einsparungen tracken
fps.onOptimizationApplied((optimization) => {
    const savings = optimization.total_co2_savings_per_hour;
    
    // An Analytics senden
    analytics.track('footprintshift_optimization_applied', {
        savings_g_co2_per_hour: savings,
        mode: optimization.mode,
        features_disabled: optimization.disable_features,
        location: optimization.location
    });
});

// Business Impact messen
if (optimization.eco_discount > 0) {
    analytics.track('eco_discount_shown', {
        discount_percent: optimization.eco_discount,
        expected_conversion_uplift: '12%' // Green marketing effect
    });
}
```

## 🚦 **Rate Limiting & Caching**

### **Empfohlene Implementierung**
```javascript
// Client-side Caching (5 Minuten)
class FootprintShiftClient {
    constructor() {
        this.cache = new Map();
        this.cacheTTL = 5 * 60 * 1000; // 5 Minuten
    }
    
    async getCarbonIntensity(location) {
        const cacheKey = `carbon_${location}`;
        const cached = this.cache.get(cacheKey);
        
        if (cached && Date.now() - cached.timestamp < this.cacheTTL) {
            return cached.data;
        }
        
        const data = await this.fetchFromAPI(location);
        this.cache.set(cacheKey, {
            data: data,
            timestamp: Date.now()
        });
        
        return data;
    }
}
```

## 💰 **Pricing Integration**

### **API Usage Tracking**
```javascript
// Free Tier: 1000 calls/day
// Pro Tier: 10,000 calls/day
// Enterprise: Unlimited

const client = new FootprintShift({
    apiKey: 'your-key',
    plan: 'pro', // free, pro, enterprise
    onQuotaExceeded: () => {
        // Fallback zu cached data oder default behavior
        console.warn('FootprintShift quota exceeded, using fallback optimization');
    }
});
```

## 🔒 **Security & Privacy**

### **API Key Management**
```javascript
// Umgebungsvariablen verwenden
const client = new FootprintShift({
    apiKey: process.env.FOOTPRINTSHIFT_API_KEY, // Server-side
    // Oder für client-side: Public key mit Domain-Beschränkung
    publicKey: process.env.FOOTPRINTSHIFT_PUBLIC_KEY
});
```

## 📈 **A/B Testing**

### **Optimization Impact messen**
```javascript
// Kontrollgruppe ohne FootprintShift
// Test-Gruppe mit FootprintShift

if (experimentGroup === 'footprintshift_enabled') {
    await fps.applyOptimizations();
    
    // Metriken tracken
    trackMetrics({
        page_load_time: performance.now(),
        co2_savings: optimization.total_savings,
        user_engagement: measureEngagement(),
        conversion_rate: trackConversions()
    });
}
```

Diese Integration ermöglicht es Entwicklern, echte Umwelt-Einsparungen zu erzielen, während sie die User Experience beibehalten und sogar verbessern können! 🌍