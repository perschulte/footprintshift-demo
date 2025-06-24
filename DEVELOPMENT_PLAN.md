# GreenWeb API - Entwicklungsplan 🌍

## 🎯 **Vision**

Eine API die Websites hilft, sich intelligent an die aktuelle CO₂-Intensität der Stromnetze anzupassen. Wenn viel sauberer Strom verfügbar ist, können teure Features aktiviert werden. Bei schmutzigem Strom werden Features reduziert und Nutzer zu "grünen Stunden" umgeleitet.

## 💡 **Kern-Konzept**

### **Adaptive Website-Optimierung:**
- **Grüner Strom** (< 150g CO₂/kWh): Alle Features aktiv, Premium-Qualität, **5% Eco-Rabatt**
- **Gelber Strom** (150-300g): Mittlere Optimierung, reduzierte Features  
- **Roter Strom** (> 300g): Eco-Mode, minimale Features, "Green Hours" Empfehlungen

### **Business Cases:**
1. **E-Commerce**: Dynamische Rabatte bei grünem Strom ("Green Friday")
2. **Streaming**: Adaptive Videoqualität basierend auf CO₂-Intensität
3. **SaaS**: Deferring von Jobs zu grünen Stunden
4. **AI/LLM**: Token-Verarbeitung nur bei sauberem Strom

## 🏗️ **Aktueller Status (Implementiert)**

### ✅ **Go API Server** (`main.go`)
- `/health` - Health Check
- `/api/v1/carbon-intensity` - Live CO₂-Intensität für Location
- `/api/v1/optimization` - Website-Optimierungsprofile
- `/api/v1/green-hours` - Vorhersage der besten Zeiten
- `/demo` - Interaktive Demo-Website

### ✅ **JavaScript SDK** (`sdk/greenweb.js`)
- Auto-Optimierung basierend auf CO₂-Daten
- React Hook Support
- "Wait for Green Energy" Funktionalität
- Automatic feature disabling/enabling

### ✅ **Shopify Integration** (`examples/shopify-integration.liquid`)
- Liquid Template für Theme-Integration
- Automatische Preisanpassungen
- Green Hour Timer
- Feature-Management (Videos, 3D, AI)

## 🚀 **MVP Entwicklungsplan (4 Wochen)**

### **Phase 1: Core API Enhancement (Woche 1)**

**Ziele:**
- [ ] Echte Electricity Maps API Integration
- [ ] Location Detection (IP-basiert)
- [ ] Redis Caching für API-Calls
- [ ] Rate Limiting implementieren

**Tasks:**
1. **Electricity Maps Integration:**
   ```go
   // Ersetze Mock-Daten durch echte API
   func getElectricityMapsData(location string) (*CarbonIntensity, error) {
       // HTTP Client für Electricity Maps API
   }
   ```

2. **Geo-Location Service:**
   ```go
   // IP-to-Location Service
   func detectLocationFromIP(ip string) (string, error) {
       // MaxMind GeoIP oder ähnlich
   }
   ```

3. **Caching Layer:**
   ```go
   // Redis Cache für API-Responses (5min TTL)
   func getCachedCarbonIntensity(location string) (*CarbonIntensity, error)
   ```

### **Phase 2: Website-Analyse Engine (Woche 2)**

**Ziele:**
- [ ] Website Carbon Footprint Analyzer (wie Digital Beacon)
- [ ] Custom Optimization Rules Engine
- [ ] Performance Impact Measurement

**Tasks:**
1. **Website Analyzer:**
   ```go
   type WebsiteAnalysis struct {
       URL              string
       PageSize         int64
       ImageCount       int
       VideoCount       int
       JSFiles          int
       CSSFiles         int
       EstimatedCO2     float64
       OptimizationTips []string
   }
   ```

2. **Rules Engine:**
   ```go
   // Benutzerdefinierte Optimierungsregeln
   type OptimizationRule struct {
       CarbonThreshold  float64
       Actions          []Action
       Priority         int
   }
   ```

### **Phase 3: Client Libraries & Integrations (Woche 3)**

**Ziele:**
- [ ] React/Vue Komponenten
- [ ] WordPress Plugin
- [ ] WooCommerce Integration
- [ ] Webhook System

**Tasks:**
1. **React Komponenten:**
   ```javascript
   // React Hook für GreenWeb
   const { carbonData, mode, isGreen } = useGreenWeb({ apiKey });
   
   // Komponenten
   <GreenBanner />
   <EcoDiscountBadge />
   <CarbonIntensityDisplay />
   ```

2. **WordPress Plugin:**
   ```php
   // Plugin für automatische Optimierung
   class GreenWebWP {
       public function applyOptimizations($carbonIntensity) {}
   }
   ```

### **Phase 4: Demo & Marketing (Woche 4)**

**Ziele:**
- [ ] Vollständiger Demo-Shop
- [ ] A/B Testing Framework
- [ ] Analytics Dashboard
- [ ] Partnerschaften mit Eco-Brands

## 🔧 **Technische Architektur**

### **Backend (Go)**
```
┌─────────────────┐     ┌──────────────────┐     ┌─────────────────┐
│   API Gateway   │────▶│  Carbon Service  │────▶│ Electricity Maps│
│   (Gin Router)  │     │  (Caching)       │     │      API        │
└─────────────────┘     └──────────────────┘     └─────────────────┘
         │                        │
         ▼                        ▼
┌─────────────────┐     ┌──────────────────┐
│ Optimization    │     │   Analytics      │
│ Engine          │     │   Service        │
└─────────────────┘     └──────────────────┘
```

### **Frontend (JavaScript)**
```
┌─────────────────┐     ┌──────────────────┐
│   Website       │────▶│  GreenWeb SDK    │
│   (Client)      │     │  (Auto-optimize) │
└─────────────────┘     └──────────────────┘
         │                        │
         ▼                        ▼
┌─────────────────┐     ┌──────────────────┐
│   User Actions  │     │  Carbon-Aware    │
│   (Shopping)    │     │  Features        │
└─────────────────┘     └──────────────────┘
```

## 📊 **Key Performance Indicators (KPIs)**

### **Environmental Impact:**
- CO₂ gespart durch Feature-Deferrals
- % der Sessions während grüner Stunden
- Durchschnittliche Carbon Intensity bei Nutzung

### **Business Impact:**
- Conversion Rate bei Green Hour Pricing
- User Engagement mit Eco-Features
- Revenue durch Green Hour Sales

### **Technical Metrics:**
- API Response Time (< 100ms)
- Cache Hit Rate (> 90%)
- Error Rate (< 0.1%)

## 🚦 **Development Workflow**

### **Lokaler Start:**
```bash
cd /Users/perschulte/Documents/dev/greenweb/greenweb-api

# Go Server starten
go run main.go

# Demo öffnen
open http://localhost:8090/demo

# API testen
curl http://localhost:8090/api/v1/carbon-intensity?location=Berlin
```

### **Testing:**
```bash
# Unit Tests
go test ./...

# Integration Tests
go test -tags=integration ./...

# Load Tests
hey -n 1000 -c 10 http://localhost:8090/api/v1/carbon-intensity
```

## 🌱 **Beispiel Use Cases**

### **1. E-Commerce Green Pricing**
```javascript
// Automatische Preisanpassung
if (carbonIntensity < 150) {
    applyDiscount('GREEN5'); // 5% Eco-Rabatt
    showBanner('🌱 Grüne Stunde! Jetzt klimafreundlich shoppen!');
}
```

### **2. Streaming Service Optimization**
```javascript
// Video-Qualität anpassen
const quality = carbonIntensity < 200 ? '4K' : '720p';
player.setQuality(quality);
```

### **3. AI/LLM Token-Optimierung**
```javascript
// Heavy AI nur bei grünem Strom
if (mode === 'green') {
    enableAIRecommendations();
    enableAdvancedSearch();
} else {
    showMessage('AI-Features verfügbar ab 23:00 (Grüne Stunden)');
}
```

## 🎯 **Competitive Advantages**

1. **First Mover**: Erstes Tool für adaptive CO₂-Website-Optimierung
2. **Real-time**: Live-Anpassung basierend auf echten Stromdaten
3. **Business Impact**: Direkte Umsatzsteigerung durch Green Pricing
4. **Easy Integration**: Plug-and-Play für bestehende Websites
5. **Measurable**: Konkrete CO₂- und Business-Metriken

## 📈 **Go-to-Market Strategy**

### **Phase 1: Open Source Community**
- SDK open sourcen auf GitHub
- Freemium API (1000 calls/day free)
- Tech-Blogs und Conference Talks

### **Phase 2: E-Commerce Focus**
- Shopify App Store
- WooCommerce Plugin
- Partnerschaften mit nachhaltigen Brands

### **Phase 3: Enterprise**
- Custom Integrations für große Retailers
- White-label Lösungen
- Carbon Credit Integration

## 🔮 **Future Vision**

**"Green Friday" statt "Black Friday"** - Ein globaler Shopping-Tag an dem alle teilnehmenden Shops synchron Rabatte geben, wenn erneuerbarer Strom verfügbar ist. Das könnte Millionen von Menschen dazu bringen, bewusster zu konsumieren und dabei sogar Geld zu sparen!

---

**Start Command:** `go run main.go` → Demo unter http://localhost:8090/demo 🚀