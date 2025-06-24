# 🧪 GreenWeb API Testing Guide - Enhanced Version

Nach dem Feedback-basierten Update können Sie das Ergebnis jetzt testen:

## 🚀 **Schneller Start**

```bash
# Im GreenWeb Verzeichnis
cd /Users/perschulte/Documents/dev/greenweb

# Server mit einfacher Version starten
go run main_simple.go
```

Der Server läuft dann auf **http://localhost:8090**

## 📊 **1. Feedback-Enhanced Features testen**

### **Dynamic Carbon Thresholds (statt statische 150/300g)**
```bash
# Berlin (EU Durchschnitt)
curl "http://localhost:8090/api/v1/carbon-intensity?location=Berlin&relative=true"

# Polen (Coal-heavy, high variation - Feedback focus region)
curl "http://localhost:8090/api/v1/carbon-intensity?location=Poland&relative=true"

# Texas (Wind + Gas peaks - Feedback focus region) 
curl "http://localhost:8090/api/v1/carbon-intensity?location=Texas&relative=true"

# China (Industrial coal patterns - Feedback focus region)
curl "http://localhost:8090/api/v1/carbon-intensity?location=China&relative=true"
```

**Was Sie sehen sollten:**
- `local_percentile`: 0-100 relative Ranking
- `daily_rank`: "top 15% cleanest hour today"
- `relative_mode`: "clean/normal/dirty" 
- `trend_direction`: "improving/stable/worsening"

### **High-Impact Optimizations (schwere Inhalte fokussiert)**
```bash
# Video Platform (4K→720p spart 24g CO₂/Stunde)
curl "http://localhost:8090/api/v1/optimization?location=Poland&url=youtube.com"

# AI Platform (3g CO₂/Session durch Green Window Deferral)
curl "http://localhost:8090/api/v1/optimization?location=Texas&url=openai.com"

# Gaming Site (15g CO₂/Stunde durch GPU-Feature Disable)
curl "http://localhost:8090/api/v1/optimization?location=China&url=gaming.com"
```

**Was Sie sehen sollten:**
- `high_impact_optimizations` Objekt mit realistischen CO₂-Einsparungen
- Video: `max_bitrate_kbps`, `co2_savings_per_hour_g`
- AI: `defer_to_green_window`, `co2_savings_per_session_g`
- GPU: `disable_webgl`, `co2_savings_per_hour_g`

### **Carbon Trends (Historical Analysis)**
```bash
# Dynamic thresholds für verschiedene Regionen
curl "http://localhost:8090/api/v1/carbon-trends?location=Poland&period=daily"
curl "http://localhost:8090/api/v1/carbon-trends?location=Texas&period=weekly"
```

**Was Sie sehen sollten:**
- `cleanest_hours`: [2, 3, 4, 23, 1] (nachts)
- `dirtiest_hours`: [18, 19, 20, 17, 8] (peak hours)
- `dynamic_thresholds` mit regionalen Percentile-Cutoffs

### **High-Impact Demo API**
```bash
# Anti-Greenwashing wissenschaftliche Basis
curl "http://localhost:8090/api/v1/high-impact-demo"
```

**Was Sie sehen sollten:**
- Realistic CO₂ savings mit wissenschaftlicher Evidenz
- Anti-Greenwashing Methodologie
- Conservative estimates mit Rebound-Effekten

## 🎮 **2. Interactive Demo Dashboard**

Öffnen Sie **http://localhost:8090/demo** im Browser:

### **Enhanced Features:**
- **Regional Rotation**: Wechselt alle 30s zwischen Polen, Texas, China, Berlin
- **High-Impact Visualisierung**: Zeigt nur Features mit echten CO₂-Einsparungen
- **Realistic Savings Calculator**: Live CO₂-Einsparungen basierend auf wissenschaftlichen Daten
- **Dynamic Thresholds**: Percentile-basierte Klassifizierung statt statisch
- **Anti-Greenwashing Info**: Methodologie-Transparenz

### **Was passiert bei verschiedenen Carbon-Levels:**

**Grüner Strom (< 150g CO₂/kWh - z.B. nachts):**
- ✅ Alle Features verfügbar
- ✅ 4K Video, AI aktiv, WebGL/3D enabled
- ✅ 5% Eco-Discount angezeigt
- ✅ "Grüne Stunde" Banner

**Hoher Carbon (> 300g CO₂/kWh - z.B. Peak hours):**
- 🔴 Video auf 720p reduziert (24g CO₂/h Einsparung)
- 🔴 AI-Features deferred (3g CO₂/Session Einsparung)
- 🔴 GPU Features disabled (15g CO₂/h Einsparung)
- 🔴 Realistische Savings-Anzeige

## 🧪 **3. Command Line Testing**

### **Verschiedene Zeiten simulieren:**
```bash
# Die Mock-Daten variieren basierend auf Tageszeit:
# 22:00-06:00: Niedrig (Green Mode)
# 12:00-16:00: Hoch (Red Mode) 
# Andere: Mittel (Yellow Mode)

# Verschiedene Regionen haben verschiedene Basis-Intensitäten:
# Polen: 340g (Coal-heavy)
# Texas: 420g (Gas + Wind)
# China: 580g (Coal dominant)
# Deutschland: 295g (EU average)
```

### **High-Impact Content Detection:**
```bash
# Video platforms
curl "http://localhost:8090/api/v1/optimization?url=netflix.com&location=Poland"

# AI platforms  
curl "http://localhost:8090/api/v1/optimization?url=chatgpt.com&location=Texas"

# Gaming sites
curl "http://localhost:8090/api/v1/optimization?url=game.com&location=China"
```

## 📊 **4. Expected Results (Feedback-basiert)**

### **Polen während Peak Hours (18:00):**
```json
{
  "carbon_intensity": 442,
  "local_percentile": 85,
  "daily_rank": "top 15% dirtiest hour today",
  "relative_mode": "dirty",
  "mode": "red",
  "optimization": {
    "mode": "eco",
    "high_impact_optimizations": {
      "video_streaming": {
        "max_bitrate_kbps": 5000,
        "max_resolution": "720p", 
        "co2_savings_per_hour_g": 24
      },
      "ai_inference": {
        "defer_to_green_window": true,
        "co2_savings_per_session_g": 3
      },
      "gpu_features": {
        "disable_webgl": true,
        "co2_savings_per_hour_g": 15
      }
    }
  }
}
```

### **Deutschland nachts (02:00):**
```json
{
  "carbon_intensity": 103,
  "local_percentile": 15,
  "daily_rank": "top 15% cleanest hour today",
  "relative_mode": "clean",
  "mode": "green",
  "optimization": {
    "mode": "full",
    "eco_discount": 5,
    "high_impact_optimizations": {
      "video_streaming": {
        "max_resolution": "4K",
        "co2_savings_per_hour_g": 0
      }
    }
  }
}
```

## 🔍 **5. Feedback-Verbesserungen validieren**

### **✅ High-Impact Focus (statt oberflächliche Optimierungen):**
- Video: 24g CO₂/h Einsparung (4K→720p) 
- AI: 3g CO₂/Session (Green Window Deferral)
- GPU: 15g CO₂/h (WebGL/3D Disable)
- **NICHT**: 200kB CSS Optimierung (minimal impact)

### **✅ Dynamic Thresholds (statt statische 150/300g):**
- Percentile-basiert: "top 20% cleanest"
- Regional angepasst: Polen vs Deutschland
- Trend-bewusst: "improving/worsening"

### **✅ Realistic CO₂ Calculations:**
- Wissenschaftlich fundiert (IEA, Carbon Trust)
- Conservative estimates
- Rebound-Effekte berücksichtigt
- Device energy included

### **✅ Graceful Degradation:**
- Standard-Qualität bleibt verfügbar
- Keine Funktionalität komplett gebrochen
- Progressive Enhancement

## 🚨 **6. Troubleshooting**

**Server startet nicht:**
```bash
# Dependencies checken
go mod tidy

# Einfache Version verwenden
go run main_simple.go
```

**Keine Daten:**
- API nutzt Mock-Daten (kein API-Key nötig)
- Carbon intensity variiert nach Tageszeit
- Verschiedene Regionen haben verschiedene Baseline

## 🎯 **Was zu erwarten ist:**

Die wichtigsten Unterschiede zum ursprünglichen System:

1. **Fokus auf High-Impact Features**: Nur noch Optimierungen die echte CO₂-Einsparungen bringen
2. **Dynamic Thresholds**: Regional angepasste Percentile statt statische Werte
3. **Realistic Calculations**: Wissenschaftlich fundierte CO₂-Berechnungen
4. **Anti-Greenwashing**: Conservative estimates mit Transparenz

Das System zeigt jetzt ehrlich was wirklich CO₂ spart und vermeidet oberflächliche Optimierungen!