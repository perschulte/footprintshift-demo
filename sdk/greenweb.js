/**
 * GreenWeb JavaScript SDK
 * Adaptive website optimization based on carbon intensity
 */

class GreenWeb {
  constructor(options = {}) {
    this.apiKey = options.apiKey || '';
    this.apiUrl = options.apiUrl || 'https://api.greenweb.dev';
    this.location = options.location || 'auto';
    this.updateInterval = options.updateInterval || 60000; // 1 minute
    this.callbacks = {
      onModeChange: options.onModeChange || (() => {}),
      onGreenHour: options.onGreenHour || (() => {}),
      onUpdate: options.onUpdate || (() => {})
    };
    
    this.currentData = null;
    this.currentMode = null;
    
    // Start monitoring
    if (options.autoStart !== false) {
      this.start();
    }
  }

  async start() {
    // Get initial data
    await this.update();
    
    // Set up periodic updates
    this.intervalId = setInterval(() => this.update(), this.updateInterval);
    
    // Listen for visibility changes to pause/resume updates
    document.addEventListener('visibilitychange', () => {
      if (document.hidden) {
        this.pause();
      } else {
        this.resume();
      }
    });
  }

  pause() {
    if (this.intervalId) {
      clearInterval(this.intervalId);
      this.intervalId = null;
    }
  }

  resume() {
    if (!this.intervalId) {
      this.start();
    }
  }

  async getCarbonIntensity(location = this.location) {
    try {
      const response = await fetch(`${this.apiUrl}/api/v1/carbon-intensity?location=${location}`, {
        headers: {
          'Authorization': `Bearer ${this.apiKey}`,
          'Content-Type': 'application/json'
        }
      });
      
      if (!response.ok) {
        throw new Error(`API error: ${response.status}`);
      }
      
      return await response.json();
    } catch (error) {
      console.error('GreenWeb: Failed to fetch carbon intensity', error);
      return null;
    }
  }

  async getOptimizationProfile(url = window.location.hostname) {
    try {
      const location = this.location === 'auto' ? await this.detectLocation() : this.location;
      const response = await fetch(`${this.apiUrl}/api/v1/optimization?location=${location}&url=${url}`, {
        headers: {
          'Authorization': `Bearer ${this.apiKey}`,
          'Content-Type': 'application/json'
        }
      });
      
      if (!response.ok) {
        throw new Error(`API error: ${response.status}`);
      }
      
      return await response.json();
    } catch (error) {
      console.error('GreenWeb: Failed to fetch optimization profile', error);
      return null;
    }
  }

  async getGreenHours(hours = 24) {
    try {
      const location = this.location === 'auto' ? await this.detectLocation() : this.location;
      const response = await fetch(`${this.apiUrl}/api/v1/green-hours?location=${location}&next=${hours}h`, {
        headers: {
          'Authorization': `Bearer ${this.apiKey}`,
          'Content-Type': 'application/json'
        }
      });
      
      if (!response.ok) {
        throw new Error(`API error: ${response.status}`);
      }
      
      return await response.json();
    } catch (error) {
      console.error('GreenWeb: Failed to fetch green hours', error);
      return null;
    }
  }

  async update() {
    const data = await this.getOptimizationProfile();
    if (!data) return;
    
    this.currentData = data;
    const newMode = data.optimization.mode;
    
    // Check if mode changed
    if (newMode !== this.currentMode) {
      this.currentMode = newMode;
      this.applyOptimizations(data.optimization);
      this.callbacks.onModeChange(newMode, data);
    }
    
    // Check if it's a green hour
    if (data.carbon_intensity.mode === 'green') {
      this.callbacks.onGreenHour(data);
    }
    
    // General update callback
    this.callbacks.onUpdate(data);
  }

  applyOptimizations(optimization) {
    // Apply CSS classes
    document.body.setAttribute('data-greenweb-mode', optimization.mode);
    
    // Disable features
    optimization.disable_features.forEach(feature => {
      document.querySelectorAll(`[data-greenweb-feature="${feature}"]`).forEach(el => {
        el.style.display = 'none';
      });
    });
    
    // Adjust image quality
    if (optimization.image_quality === 'low') {
      document.querySelectorAll('img').forEach(img => {
        img.loading = 'lazy';
        // Add low quality class for CSS handling
        img.classList.add('greenweb-low-quality');
      });
    }
    
    // Show eco discount if available
    if (optimization.eco_discount > 0) {
      this.showEcoDiscount(optimization.eco_discount);
    }
  }

  showEcoDiscount(discount) {
    // Create or update discount banner
    let banner = document.getElementById('greenweb-discount-banner');
    if (!banner) {
      banner = document.createElement('div');
      banner.id = 'greenweb-discount-banner';
      banner.style.cssText = `
        position: fixed;
        top: 0;
        left: 0;
        right: 0;
        background: #22c55e;
        color: white;
        padding: 15px;
        text-align: center;
        z-index: 9999;
        font-size: 16px;
        font-weight: bold;
      `;
      document.body.prepend(banner);
    }
    
    banner.textContent = `🌱 Green Hour Special: ${discount}% off all purchases!`;
    banner.style.display = 'block';
  }

  async detectLocation() {
    // Try to detect user location
    // In production, this would use geolocation or IP-based detection
    return 'Berlin';
  }

  // Utility methods for common use cases
  
  async waitForGreenEnergy(callback, maxWaitHours = 24) {
    const greenHours = await this.getGreenHours(maxWaitHours);
    if (!greenHours || !greenHours.best_window) {
      // No green window found, execute anyway
      callback();
      return;
    }
    
    const now = new Date();
    const greenStart = new Date(greenHours.best_window.start);
    const waitTime = greenStart - now;
    
    if (waitTime <= 0) {
      // Already in green window
      callback();
    } else {
      // Schedule for green window
      console.log(`GreenWeb: Deferring execution for ${(waitTime / 1000 / 60).toFixed(0)} minutes until green energy is available`);
      setTimeout(callback, waitTime);
    }
  }

  // React hook helper
  static useGreenWeb(options) {
    if (typeof React === 'undefined') {
      console.error('GreenWeb: React not found');
      return null;
    }
    
    const [carbonData, setCarbonData] = React.useState(null);
    const [mode, setMode] = React.useState(null);
    
    React.useEffect(() => {
      const gw = new GreenWeb({
        ...options,
        onUpdate: (data) => {
          setCarbonData(data);
          setMode(data.optimization.mode);
        }
      });
      
      return () => gw.pause();
    }, []);
    
    return { carbonData, mode, isGreen: mode === 'full' };
  }
}

// Auto-initialize if data attributes are present
if (typeof window !== 'undefined' && window.document) {
  document.addEventListener('DOMContentLoaded', () => {
    const script = document.querySelector('script[data-greenweb-api-key]');
    if (script) {
      window.greenWeb = new GreenWeb({
        apiKey: script.getAttribute('data-greenweb-api-key'),
        location: script.getAttribute('data-greenweb-location') || 'auto'
      });
    }
  });
}

// Export for various module systems
if (typeof module !== 'undefined' && module.exports) {
  module.exports = GreenWeb;
} else if (typeof define === 'function' && define.amd) {
  define([], () => GreenWeb);
} else {
  window.GreenWeb = GreenWeb;
}