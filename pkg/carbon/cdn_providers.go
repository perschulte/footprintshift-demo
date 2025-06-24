// Package carbon provides CDN provider configurations for carbon-aware routing.
//
// This file contains edge location mappings for major CDN providers including
// CloudFlare, AWS CloudFront, Google Cloud CDN, and Azure CDN.
package carbon

// MajorCDNProviders contains configuration for major CDN providers.
var MajorCDNProviders = map[string]CDNProvider{
	"cloudflare": {
		Name: "CloudFlare",
		EdgeLocations: map[string]EdgeLocationInfo{
			// North America
			"atlanta": {
				City: "Atlanta", Country: "USA", GridZone: "US-SE",
				Latitude: 33.7490, Longitude: -84.3880, Tier: 1, Capacity: "high",
			},
			"chicago": {
				City: "Chicago", Country: "USA", GridZone: "US-MIDA",
				Latitude: 41.8781, Longitude: -87.6298, Tier: 1, Capacity: "high",
			},
			"dallas": {
				City: "Dallas", Country: "USA", GridZone: "US-TEX",
				Latitude: 32.7767, Longitude: -96.7970, Tier: 1, Capacity: "high",
			},
			"los-angeles": {
				City: "Los Angeles", Country: "USA", GridZone: "US-CAL-CISO",
				Latitude: 34.0522, Longitude: -118.2437, Tier: 1, Capacity: "high",
			},
			"miami": {
				City: "Miami", Country: "USA", GridZone: "US-FLA",
				Latitude: 25.7617, Longitude: -80.1918, Tier: 1, Capacity: "high",
			},
			"new-york": {
				City: "New York", Country: "USA", GridZone: "US-NY-NYIS",
				Latitude: 40.7128, Longitude: -74.0060, Tier: 1, Capacity: "high",
			},
			"san-francisco": {
				City: "San Francisco", Country: "USA", GridZone: "US-CAL-CISO",
				Latitude: 37.7749, Longitude: -122.4194, Tier: 1, Capacity: "high",
			},
			"seattle": {
				City: "Seattle", Country: "USA", GridZone: "US-NW-PACW",
				Latitude: 47.6062, Longitude: -122.3321, Tier: 1, Capacity: "high",
			},
			"toronto": {
				City: "Toronto", Country: "Canada", GridZone: "CA-ON",
				Latitude: 43.6532, Longitude: -79.3832, Tier: 1, Capacity: "high",
			},

			// Europe
			"amsterdam": {
				City: "Amsterdam", Country: "Netherlands", GridZone: "NL",
				Latitude: 52.3676, Longitude: 4.9041, Tier: 1, Capacity: "high",
				RenewableCommitment: true,
			},
			"frankfurt": {
				City: "Frankfurt", Country: "Germany", GridZone: "DE",
				Latitude: 50.1109, Longitude: 8.6821, Tier: 1, Capacity: "high",
				RenewableCommitment: true,
			},
			"london": {
				City: "London", Country: "UK", GridZone: "GB",
				Latitude: 51.5074, Longitude: -0.1278, Tier: 1, Capacity: "high",
			},
			"paris": {
				City: "Paris", Country: "France", GridZone: "FR",
				Latitude: 48.8566, Longitude: 2.3522, Tier: 1, Capacity: "high",
			},
			"stockholm": {
				City: "Stockholm", Country: "Sweden", GridZone: "SE",
				Latitude: 59.3293, Longitude: 18.0686, Tier: 1, Capacity: "high",
				RenewableCommitment: true,
			},

			// Asia Pacific
			"singapore": {
				City: "Singapore", Country: "Singapore", GridZone: "SG",
				Latitude: 1.3521, Longitude: 103.8198, Tier: 1, Capacity: "high",
			},
			"tokyo": {
				City: "Tokyo", Country: "Japan", GridZone: "JP",
				Latitude: 35.6762, Longitude: 139.6503, Tier: 1, Capacity: "high",
			},
			"sydney": {
				City: "Sydney", Country: "Australia", GridZone: "AU-NSW",
				Latitude: -33.8688, Longitude: 151.2093, Tier: 1, Capacity: "high",
			},
			"hong-kong": {
				City: "Hong Kong", Country: "China", GridZone: "HK",
				Latitude: 22.3193, Longitude: 114.1694, Tier: 1, Capacity: "high",
			},
		},
		DefaultEdgeSelection: "geo_nearest",
		CarbonAwareRouting:   false,
	},

	"aws-cloudfront": {
		Name: "AWS CloudFront",
		EdgeLocations: map[string]EdgeLocationInfo{
			// North America
			"us-east-1": {
				City: "N. Virginia", Country: "USA", GridZone: "US-MIDA",
				Latitude: 38.7464, Longitude: -77.4735, Tier: 1, Capacity: "high",
			},
			"us-east-2": {
				City: "Ohio", Country: "USA", GridZone: "US-MIDW-MISO",
				Latitude: 40.4173, Longitude: -82.9071, Tier: 1, Capacity: "high",
			},
			"us-west-1": {
				City: "N. California", Country: "USA", GridZone: "US-CAL-CISO",
				Latitude: 37.3541, Longitude: -121.9552, Tier: 1, Capacity: "high",
			},
			"us-west-2": {
				City: "Oregon", Country: "USA", GridZone: "US-NW-PACW",
				Latitude: 45.5152, Longitude: -122.6784, Tier: 1, Capacity: "high",
				RenewableCommitment: true,
			},
			"ca-central-1": {
				City: "Montreal", Country: "Canada", GridZone: "CA-QC",
				Latitude: 45.5017, Longitude: -73.5673, Tier: 1, Capacity: "high",
				RenewableCommitment: true,
			},

			// Europe
			"eu-west-1": {
				City: "Dublin", Country: "Ireland", GridZone: "IE",
				Latitude: 53.3498, Longitude: -6.2603, Tier: 1, Capacity: "high",
				RenewableCommitment: true,
			},
			"eu-central-1": {
				City: "Frankfurt", Country: "Germany", GridZone: "DE",
				Latitude: 50.1109, Longitude: 8.6821, Tier: 1, Capacity: "high",
				RenewableCommitment: true,
			},
			"eu-west-2": {
				City: "London", Country: "UK", GridZone: "GB",
				Latitude: 51.5074, Longitude: -0.1278, Tier: 1, Capacity: "high",
			},
			"eu-west-3": {
				City: "Paris", Country: "France", GridZone: "FR",
				Latitude: 48.8566, Longitude: 2.3522, Tier: 1, Capacity: "high",
			},
			"eu-north-1": {
				City: "Stockholm", Country: "Sweden", GridZone: "SE",
				Latitude: 59.3293, Longitude: 18.0686, Tier: 1, Capacity: "high",
				RenewableCommitment: true,
			},

			// Asia Pacific
			"ap-southeast-1": {
				City: "Singapore", Country: "Singapore", GridZone: "SG",
				Latitude: 1.3521, Longitude: 103.8198, Tier: 1, Capacity: "high",
			},
			"ap-northeast-1": {
				City: "Tokyo", Country: "Japan", GridZone: "JP",
				Latitude: 35.6762, Longitude: 139.6503, Tier: 1, Capacity: "high",
			},
			"ap-southeast-2": {
				City: "Sydney", Country: "Australia", GridZone: "AU-NSW",
				Latitude: -33.8688, Longitude: 151.2093, Tier: 1, Capacity: "high",
			},
			"ap-south-1": {
				City: "Mumbai", Country: "India", GridZone: "IN-WR",
				Latitude: 19.0760, Longitude: 72.8777, Tier: 1, Capacity: "high",
			},
		},
		DefaultEdgeSelection: "latency_based",
		CarbonAwareRouting:   false,
	},

	"google-cloud": {
		Name: "Google Cloud CDN",
		EdgeLocations: map[string]EdgeLocationInfo{
			// Americas
			"us-central1": {
				City: "Iowa", Country: "USA", GridZone: "US-MIDW-MISO",
				Latitude: 41.8780, Longitude: -93.0977, Tier: 1, Capacity: "high",
				RenewableCommitment: true,
			},
			"us-east1": {
				City: "South Carolina", Country: "USA", GridZone: "US-SE",
				Latitude: 33.8361, Longitude: -81.1637, Tier: 1, Capacity: "high",
			},
			"us-east4": {
				City: "Northern Virginia", Country: "USA", GridZone: "US-MIDA",
				Latitude: 38.7464, Longitude: -77.4735, Tier: 1, Capacity: "high",
			},
			"us-west1": {
				City: "Oregon", Country: "USA", GridZone: "US-NW-PACW",
				Latitude: 45.5152, Longitude: -122.6784, Tier: 1, Capacity: "high",
				RenewableCommitment: true,
			},
			"us-west2": {
				City: "Los Angeles", Country: "USA", GridZone: "US-CAL-CISO",
				Latitude: 34.0522, Longitude: -118.2437, Tier: 1, Capacity: "high",
			},
			"northamerica-northeast1": {
				City: "Montreal", Country: "Canada", GridZone: "CA-QC",
				Latitude: 45.5017, Longitude: -73.5673, Tier: 1, Capacity: "high",
				RenewableCommitment: true,
			},
			"southamerica-east1": {
				City: "Sao Paulo", Country: "Brazil", GridZone: "BR-SP",
				Latitude: -23.5505, Longitude: -46.6333, Tier: 1, Capacity: "high",
			},

			// Europe
			"europe-west1": {
				City: "Belgium", Country: "Belgium", GridZone: "BE",
				Latitude: 50.8503, Longitude: 4.3517, Tier: 1, Capacity: "high",
				RenewableCommitment: true,
			},
			"europe-west2": {
				City: "London", Country: "UK", GridZone: "GB",
				Latitude: 51.5074, Longitude: -0.1278, Tier: 1, Capacity: "high",
			},
			"europe-west3": {
				City: "Frankfurt", Country: "Germany", GridZone: "DE",
				Latitude: 50.1109, Longitude: 8.6821, Tier: 1, Capacity: "high",
			},
			"europe-west4": {
				City: "Netherlands", Country: "Netherlands", GridZone: "NL",
				Latitude: 52.3676, Longitude: 4.9041, Tier: 1, Capacity: "high",
				RenewableCommitment: true,
			},
			"europe-west6": {
				City: "Zurich", Country: "Switzerland", GridZone: "CH",
				Latitude: 47.3769, Longitude: 8.5417, Tier: 1, Capacity: "high",
				RenewableCommitment: true,
			},
			"europe-north1": {
				City: "Finland", Country: "Finland", GridZone: "FI",
				Latitude: 60.1699, Longitude: 24.9384, Tier: 1, Capacity: "high",
				RenewableCommitment: true,
			},

			// Asia Pacific
			"asia-east1": {
				City: "Taiwan", Country: "Taiwan", GridZone: "TW",
				Latitude: 25.0330, Longitude: 121.5654, Tier: 1, Capacity: "high",
			},
			"asia-east2": {
				City: "Hong Kong", Country: "China", GridZone: "HK",
				Latitude: 22.3193, Longitude: 114.1694, Tier: 1, Capacity: "high",
			},
			"asia-northeast1": {
				City: "Tokyo", Country: "Japan", GridZone: "JP",
				Latitude: 35.6762, Longitude: 139.6503, Tier: 1, Capacity: "high",
			},
			"asia-northeast2": {
				City: "Osaka", Country: "Japan", GridZone: "JP",
				Latitude: 34.6937, Longitude: 135.5023, Tier: 1, Capacity: "high",
			},
			"asia-south1": {
				City: "Mumbai", Country: "India", GridZone: "IN-WR",
				Latitude: 19.0760, Longitude: 72.8777, Tier: 1, Capacity: "high",
			},
			"asia-southeast1": {
				City: "Singapore", Country: "Singapore", GridZone: "SG",
				Latitude: 1.3521, Longitude: 103.8198, Tier: 1, Capacity: "high",
			},
			"australia-southeast1": {
				City: "Sydney", Country: "Australia", GridZone: "AU-NSW",
				Latitude: -33.8688, Longitude: 151.2093, Tier: 1, Capacity: "high",
			},
		},
		DefaultEdgeSelection: "geo_distributed",
		CarbonAwareRouting:   true, // Google has carbon-aware features
	},

	"azure": {
		Name: "Azure CDN",
		EdgeLocations: map[string]EdgeLocationInfo{
			// Americas
			"eastus": {
				City: "Virginia", Country: "USA", GridZone: "US-MIDA",
				Latitude: 37.3719, Longitude: -79.8164, Tier: 1, Capacity: "high",
			},
			"eastus2": {
				City: "Virginia", Country: "USA", GridZone: "US-MIDA",
				Latitude: 36.6681, Longitude: -78.3889, Tier: 1, Capacity: "high",
			},
			"westus": {
				City: "California", Country: "USA", GridZone: "US-CAL-CISO",
				Latitude: 37.7749, Longitude: -122.4194, Tier: 1, Capacity: "high",
			},
			"westus2": {
				City: "Washington", Country: "USA", GridZone: "US-NW-PACW",
				Latitude: 47.2330, Longitude: -119.8524, Tier: 1, Capacity: "high",
				RenewableCommitment: true,
			},
			"centralus": {
				City: "Iowa", Country: "USA", GridZone: "US-MIDW-MISO",
				Latitude: 41.8780, Longitude: -93.0977, Tier: 1, Capacity: "high",
			},
			"canadaeast": {
				City: "Quebec", Country: "Canada", GridZone: "CA-QC",
				Latitude: 46.8139, Longitude: -71.2080, Tier: 1, Capacity: "high",
				RenewableCommitment: true,
			},
			"brazilsouth": {
				City: "Sao Paulo", Country: "Brazil", GridZone: "BR-SP",
				Latitude: -23.5505, Longitude: -46.6333, Tier: 1, Capacity: "high",
			},

			// Europe
			"northeurope": {
				City: "Dublin", Country: "Ireland", GridZone: "IE",
				Latitude: 53.3498, Longitude: -6.2603, Tier: 1, Capacity: "high",
				RenewableCommitment: true,
			},
			"westeurope": {
				City: "Amsterdam", Country: "Netherlands", GridZone: "NL",
				Latitude: 52.3676, Longitude: 4.9041, Tier: 1, Capacity: "high",
				RenewableCommitment: true,
			},
			"uksouth": {
				City: "London", Country: "UK", GridZone: "GB",
				Latitude: 50.9410, Longitude: -0.7992, Tier: 1, Capacity: "high",
			},
			"francecentral": {
				City: "Paris", Country: "France", GridZone: "FR",
				Latitude: 46.3630, Longitude: 2.7072, Tier: 1, Capacity: "high",
			},
			"germanywestcentral": {
				City: "Frankfurt", Country: "Germany", GridZone: "DE",
				Latitude: 50.1109, Longitude: 8.6821, Tier: 1, Capacity: "high",
			},
			"switzerlandnorth": {
				City: "Zurich", Country: "Switzerland", GridZone: "CH",
				Latitude: 47.3769, Longitude: 8.5417, Tier: 1, Capacity: "high",
				RenewableCommitment: true,
			},
			"norwayeast": {
				City: "Oslo", Country: "Norway", GridZone: "NO",
				Latitude: 59.9139, Longitude: 10.7522, Tier: 1, Capacity: "high",
				RenewableCommitment: true,
			},

			// Asia Pacific
			"eastasia": {
				City: "Hong Kong", Country: "China", GridZone: "HK",
				Latitude: 22.3193, Longitude: 114.1694, Tier: 1, Capacity: "high",
			},
			"southeastasia": {
				City: "Singapore", Country: "Singapore", GridZone: "SG",
				Latitude: 1.3521, Longitude: 103.8198, Tier: 1, Capacity: "high",
			},
			"japaneast": {
				City: "Tokyo", Country: "Japan", GridZone: "JP",
				Latitude: 35.6762, Longitude: 139.6503, Tier: 1, Capacity: "high",
			},
			"koreacentral": {
				City: "Seoul", Country: "South Korea", GridZone: "KR",
				Latitude: 37.5665, Longitude: 126.9780, Tier: 1, Capacity: "high",
			},
			"centralindia": {
				City: "Pune", Country: "India", GridZone: "IN-WR",
				Latitude: 18.5204, Longitude: 73.8567, Tier: 1, Capacity: "high",
			},
			"australiaeast": {
				City: "Sydney", Country: "Australia", GridZone: "AU-NSW",
				Latitude: -33.8688, Longitude: 151.2093, Tier: 1, Capacity: "high",
			},
		},
		DefaultEdgeSelection: "performance_based",
		CarbonAwareRouting:   false,
	},
}

// GetCDNProvider returns the configuration for a specific CDN provider.
func GetCDNProvider(providerName string) (*CDNProvider, bool) {
	provider, exists := MajorCDNProviders[providerName]
	if !exists {
		return nil, false
	}
	return &provider, true
}

// GetAllCDNProviders returns a list of all configured CDN provider names.
func GetAllCDNProviders() []string {
	providers := make([]string, 0, len(MajorCDNProviders))
	for name := range MajorCDNProviders {
		providers = append(providers, name)
	}
	return providers
}

// FindNearestEdgeLocation finds the nearest edge location for a given provider and user location.
func FindNearestEdgeLocation(providerName string, userLat, userLon float64) (*EdgeLocationInfo, float64) {
	provider, exists := MajorCDNProviders[providerName]
	if !exists {
		return nil, 0
	}

	var nearestEdge *EdgeLocationInfo
	minDistance := float64(999999)

	for _, edge := range provider.EdgeLocations {
		distance := CalculateDistance(userLat, userLon, edge.Latitude, edge.Longitude)
		if distance < minDistance {
			minDistance = distance
			edgeCopy := edge
			nearestEdge = &edgeCopy
		}
	}

	return nearestEdge, minDistance
}

// GetRenewableEdgeLocations returns all edge locations with renewable energy commitments for a provider.
func GetRenewableEdgeLocations(providerName string) []EdgeLocationInfo {
	provider, exists := MajorCDNProviders[providerName]
	if !exists {
		return nil
	}

	var renewableEdges []EdgeLocationInfo
	for _, edge := range provider.EdgeLocations {
		if edge.RenewableCommitment {
			renewableEdges = append(renewableEdges, edge)
		}
	}

	return renewableEdges
}