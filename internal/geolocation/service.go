package geolocation

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

// Service provides IP-based geolocation functionality
type Service struct {
	client      *http.Client
	gridMapper  *GridZoneMapper
	apiBaseURL  string
	rateLimiter chan struct{}
}

// ServiceConfig holds configuration for the geolocation service
type ServiceConfig struct {
	APIBaseURL    string
	Timeout       time.Duration
	RateLimitRPS  int // Requests per second
}

// NewService creates a new geolocation service
func NewService(config ServiceConfig) *Service {
	if config.APIBaseURL == "" {
		config.APIBaseURL = "https://ipapi.co"
	}
	if config.Timeout == 0 {
		config.Timeout = 5 * time.Second
	}
	if config.RateLimitRPS == 0 {
		config.RateLimitRPS = 2 // Conservative rate limit for free tier
	}

	// Create rate limiter channel
	rateLimiter := make(chan struct{}, config.RateLimitRPS)
	go func() {
		ticker := time.NewTicker(time.Second / time.Duration(config.RateLimitRPS))
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				select {
				case rateLimiter <- struct{}{}:
				default:
				}
			}
		}
	}()

	return &Service{
		client: &http.Client{
			Timeout: config.Timeout,
		},
		gridMapper:  NewGridZoneMapper(),
		apiBaseURL:  config.APIBaseURL,
		rateLimiter: rateLimiter,
	}
}

// GetLocationByIP retrieves location information for the given IP address
func (s *Service) GetLocationByIP(ctx context.Context, ip string) (LocationWithZone, error) {
	// Check for private/local IPs
	if IsPrivateOrLocalIP(ip) {
		log.Printf("Private or local IP detected (%s), using default location", ip)
		return LocationWithZone{
			Location: DefaultLocation,
			GridZone: DefaultGridZone,
		}, nil
	}

	// Rate limiting
	select {
	case <-s.rateLimiter:
		// Proceed with request
	case <-ctx.Done():
		return s.getDefaultLocationWithZone(), ctx.Err()
	case <-time.After(2 * time.Second):
		log.Printf("Rate limit timeout for IP %s, using default location", ip)
		return s.getDefaultLocationWithZone(), nil
	}

	// Make API request
	location, err := s.fetchLocationFromAPI(ctx, ip)
	if err != nil {
		log.Printf("Failed to fetch location for IP %s: %v, using default", ip, err)
		return s.getDefaultLocationWithZone(), nil
	}

	// Map to grid zone
	gridZone := s.gridMapper.MapToGridZone(location)

	return LocationWithZone{
		Location: location,
		GridZone: gridZone,
	}, nil
}

// GetLocationFromRequest extracts IP from HTTP request and returns location
func (s *Service) GetLocationFromRequest(ctx context.Context, r *http.Request) (LocationWithZone, error) {
	ip := ExtractClientIP(r)
	return s.GetLocationByIP(ctx, ip)
}

// fetchLocationFromAPI makes the actual API call to ipapi.co
func (s *Service) fetchLocationFromAPI(ctx context.Context, ip string) (Location, error) {
	url := fmt.Sprintf("%s/%s/json/", s.apiBaseURL, ip)
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return Location{}, fmt.Errorf("failed to create request: %w", err)
	}

	// Set user agent
	req.Header.Set("User-Agent", "GreenWeb-Service/1.0")

	resp, err := s.client.Do(req)
	if err != nil {
		return Location{}, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Location{}, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var apiResp IpapiResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return Location{}, fmt.Errorf("failed to decode response: %w", err)
	}

	// Convert to our internal format
	location := Location{
		IP:          apiResp.IP,
		Country:     apiResp.CountryName,
		CountryCode: apiResp.CountryCode,
		Region:      apiResp.Region,
		RegionCode:  apiResp.RegionCode,
		City:        apiResp.City,
		Latitude:    apiResp.Latitude,
		Longitude:   apiResp.Longitude,
		Timezone:    apiResp.Timezone,
	}

	// Validate required fields
	if location.CountryCode == "" || location.Country == "" {
		return Location{}, fmt.Errorf("incomplete location data received")
	}

	return location, nil
}

// getDefaultLocationWithZone returns the default Berlin location with grid zone
func (s *Service) getDefaultLocationWithZone() LocationWithZone {
	return LocationWithZone{
		Location: DefaultLocation,
		GridZone: DefaultGridZone,
	}
}

// GetGridZoneMapper returns the grid zone mapper for external use
func (s *Service) GetGridZoneMapper() *GridZoneMapper {
	return s.gridMapper
}

// HealthCheck verifies the service is working by testing with a known IP
func (s *Service) HealthCheck(ctx context.Context) error {
	// Test with Google's public DNS IP
	testIP := "8.8.8.8"
	
	_, err := s.fetchLocationFromAPI(ctx, testIP)
	if err != nil {
		return fmt.Errorf("geolocation service health check failed: %w", err)
	}
	
	return nil
}