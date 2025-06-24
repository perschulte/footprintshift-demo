# 🚀 FootprintShift Quick Deployment Guide

## 📋 **Prerequisites:**
- GitHub Account  
- Railway/Render Account (kostenlos)

## ⚡ **Railway Deployment (5 Minuten):**

### **1. GitHub Setup:**
```bash
# In /Users/perschulte/Documents/dev/greenweb
git add .
git commit -m "FootprintShift Demo - Ready for Railway"
git push origin main
```

### **2. Railway Deployment:**
1. Gehe zu **https://railway.app**
2. "New Project" → "Deploy from GitHub"
3. Select: **greenweb** repository
4. Railway erkennt automatisch Go
5. **Environment Variables**: keine nötig
6. Deploy! ✅

### **3. Access Demo:**
- URL: `https://[project-name].railway.app/demo`
- Automatisches HTTPS
- Global CDN

## 🎯 **Alternative: Render**

1. **https://render.com** → "New Web Service"
2. GitHub Repository: greenweb
3. **Build Command:** `go build time_series_demo.go`
4. **Start Command:** `./time_series_demo`
5. Deploy!

## 📊 **Demo Features Available:**

✅ **15-Minute Carbon Simulation** (96 data points)  
✅ **Interactive Time Slider** (Dieter Rams Design)  
✅ **Real-time Optimization Cards**  
✅ **Interest Registration Form**  
✅ **Mobile Responsive**  

## 🔍 **Monitoring Interest:**

Nach Deployment schauen Sie die **Railway/Render Logs**:
```
🎯 Interest registered: user@company.com (TechCorp) - developer - video
```

## 🌐 **Share URLs:**

**Demo Page:** `https://[your-app].railway.app/demo`  
**API Health:** `https://[your-app].railway.app/health`  
**Time Series:** `https://[your-app].railway.app/api/v1/carbon-intensity/timeseries`

---

**Ready to deploy!** 🚀 Railway macht es super einfach.