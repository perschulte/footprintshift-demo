package geolocation

import (
	"context"
	"net/http/httptest"
	"testing"
	"time"
)

func TestExtractClientIP(t *testing.T) {
	tests := []struct {
		name           string
		headers        map[string]string
		remoteAddr     string
		expectedResult string
	}{
		{
			name:           "X-Forwarded-For private IP falls back",
			headers:        map[string]string{"X-Forwarded-For": "192.168.1.1"},
			remoteAddr:     "127.0.0.1:8080",
			expectedResult: "127.0.0.1", // Private IP should fall back to cleaned RemoteAddr
		},
		{
			name:           "X-Real-IP public",
			headers:        map[string]string{"X-Real-IP": "8.8.8.8"},
			remoteAddr:     "127.0.0.1:8080",
			expectedResult: "8.8.8.8",
		},
		{
			name:           "CF-Connecting-IP public",
			headers:        map[string]string{"CF-Connecting-IP": "1.1.1.1"},
			remoteAddr:     "127.0.0.1:8080",
			expectedResult: "1.1.1.1",
		},
		{
			name:           "No headers",
			headers:        map[string]string{},
			remoteAddr:     "127.0.0.1:8080",
			expectedResult: "127.0.0.1",
		},
		{
			name:           "X-Forwarded-For public IP",
			headers:        map[string]string{"X-Forwarded-For": "8.8.8.8"},
			remoteAddr:     "127.0.0.1:8080",
			expectedResult: "8.8.8.8",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			req.RemoteAddr = tt.remoteAddr
			
			for key, value := range tt.headers {
				req.Header.Set(key, value)
			}

			result := ExtractClientIP(req)
			if result != tt.expectedResult {
				t.Errorf("ExtractClientIP() = %v, want %v", result, tt.expectedResult)
			}
		})
	}
}

func TestIsValidIP(t *testing.T) {
	tests := []struct {
		ip       string
		expected bool
	}{
		{"8.8.8.8", true},
		{"1.1.1.1", true},
		{"192.168.1.1", false}, // Private IP
		{"127.0.0.1", false},   // Loopback IP
		{"10.0.0.1", false},    // Private IP
		{"invalid", false},     // Invalid IP
		{"", false},            // Empty string
	}

	for _, tt := range tests {
		t.Run(tt.ip, func(t *testing.T) {
			result := isValidIP(tt.ip)
			if result != tt.expected {
				t.Errorf("isValidIP(%v) = %v, want %v", tt.ip, result, tt.expected)
			}
		})
	}
}

func TestGridZoneMapper(t *testing.T) {
	mapper := NewGridZoneMapper()

	// Test known country codes
	location := Location{
		Country:     "Germany",
		CountryCode: "DE",
		Region:      "Berlin",
	}

	zone := mapper.MapToGridZone(location)
	if zone.Zone != "DE" {
		t.Errorf("Expected zone DE, got %s", zone.Zone)
	}
	if zone.Country != "Germany" {
		t.Errorf("Expected country Germany, got %s", zone.Country)
	}

	// Test unknown country code
	unknownLocation := Location{
		Country:     "Unknown Country",
		CountryCode: "XX",
		Region:      "Unknown Region",
	}

	unknownZone := mapper.MapToGridZone(unknownLocation)
	if unknownZone.Zone != "DE" { // Should default to DE
		t.Errorf("Expected default zone DE, got %s", unknownZone.Zone)
	}
}

func TestGeolocationService(t *testing.T) {
	// Create service with short timeout for testing
	config := ServiceConfig{
		APIBaseURL:   "https://httpbin.org", // Use httpbin for testing
		Timeout:      2 * time.Second,
		RateLimitRPS: 5,
	}
	service := NewService(config)

	// Test with private IP (should return default)
	ctx := context.Background()
	result, err := service.GetLocationByIP(ctx, "192.168.1.1")
	if err != nil {
		t.Errorf("GetLocationByIP should not return error for private IP: %v", err)
	}
	if result.Location.Country != DefaultLocation.Country {
		t.Errorf("Expected default location for private IP, got %s", result.Location.Country)
	}

	// Test grid zone mapper access
	mapper := service.GetGridZoneMapper()
	if mapper == nil {
		t.Error("GetGridZoneMapper should not return nil")
	}

	zones := mapper.GetAvailableZones()
	if len(zones) == 0 {
		t.Error("Should have available zones")
	}
}

func TestDefaultLocationWithZone(t *testing.T) {
	config := ServiceConfig{}
	service := NewService(config)

	defaultResult := service.getDefaultLocationWithZone()
	
	if defaultResult.Location.Country != DefaultLocation.Country {
		t.Errorf("Expected default country %s, got %s", DefaultLocation.Country, defaultResult.Location.Country)
	}
	
	if defaultResult.GridZone.Zone != DefaultGridZone.Zone {
		t.Errorf("Expected default grid zone %s, got %s", DefaultGridZone.Zone, defaultResult.GridZone.Zone)
	}
}