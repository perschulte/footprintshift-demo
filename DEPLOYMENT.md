# 🚀 FootprintShift Demo Deployment

Live Demo der FootprintShift API mit 15-Minuten Carbon Intensity Simulation für Deutschland.

## 🌐 **Live Demo**

**URL**: [Wird nach Deployment hier eingefügt]

**Features:**
- 24h Germany Carbon Intensity Simulation (96 Datenpunkte)
- Interactive Time Slider (Dieter Rams Design)
- Real-time CO₂ Optimization Recommendations  
- Interest Registration für Early Access

## 🛠 **Deployment Optionen**

### **Option 1: Railway (Empfohlen)**

1. **Repository Setup:**
   ```bash
   git init
   git add .
   git commit -m "FootprintShift Demo for Railway"
   git push origin main
   ```

2. **Railway Deployment:**
   - Gehe zu [railway.app](https://railway.app)
   - "New Project" → "Deploy from GitHub"
   - Select Repository
   - Environment Variables: keine nötig (Demo uses mock data)
   - Deploy!

3. **URL Access:**
   - Railway generiert automatisch: `https://footprintshift-demo.railway.app`
   - Demo verfügbar unter: `/demo`

### **Option 2: Render**

1. **Render Setup:**
   - Gehe zu [render.com](https://render.com)
   - "New Web Service" → GitHub Repository
   - Build Command: `go build time_series_demo.go`
   - Start Command: `./time_series_demo`

### **Option 3: Fly.io**

1. **Dockerfile erstellen:**
   ```dockerfile
   FROM golang:1.21-alpine
   WORKDIR /app
   COPY . .
   RUN go build -o main time_series_demo.go
   EXPOSE 8090
   CMD ["./main"]
   ```

2. **Fly Deploy:**
   ```bash
   fly launch
   fly deploy
   ```

## ⚙️ **Configuration**

### **Environment Variables**
```bash
PORT=8090                    # Automatisch von Platform gesetzt
ENVIRONMENT=production
```

### **Health Check**
- Endpoint: `/health`
- Response: `{"status": "healthy", "service": "footprintshift-api"}`

## 🔧 **Local Development**

```bash
# Clone repository
git clone [repository-url]
cd greenweb

# Install dependencies  
go mod tidy

# Run demo
go run time_series_demo.go

# Access demo
open http://localhost:8090/demo
```

## 📊 **Demo Features**

### **Time Series Simulation:**
- **96 Datenpunkte** (15-Minuten-Intervalle)
- **Realistische Patterns**: Solar, Wind, Demand modeling
- **Interactive Controls**: Play, Pause, Reset
- **Smooth Animation**: 250ms pro Step (24s für ganzen Tag)

### **Carbon Optimization:**
- **Video**: 4K→720p = 24g CO₂/h Einsparung
- **AI**: Green Window Deferral = 3g CO₂/session
- **GPU**: WebGL Disable = 15g CO₂/h

### **Interest Registration:**
- Email capture für Early Access
- Role selection (Developer, CTO, etc.)
- Use case tracking (Video, AI, E-commerce, etc.)
- Form submissions logged in server console

## 🎨 **Design**

**Dieter Rams Principles:**
- Weniger ist mehr
- Funktionalität vor Ästhetik  
- Systematisches Grid-Layout
- Präzise Typographie (Helvetica)
- Minimale Farbpalette

## 🔍 **Monitoring**

### **Logs (Railway/Render):**
```bash
# Interest registrations erscheinen als:
🎯 Interest registered: user@company.com (TechCorp) - developer - video

# Server startup:
🌍 FootprintShift API Demo starting on port 8090
📊 Demo: http://localhost:8090/demo
```

### **Analytics:**
- Form submissions tracked in logs
- Visitor analytics über Platform verfügbar
- Custom analytics später hinzufügbar

## 🚀 **Next Steps nach Deployment**

1. **Share URL** mit Partnern/Friends
2. **Monitor Interest** via Server Logs  
3. **Collect Feedback** 
4. **Iterate** based on User Input
5. **Custom Domain** (footprintshift.dev) später

---

**Deployment Ready!** 🌍 Bereit für Railway/Render/Fly.io deployment.