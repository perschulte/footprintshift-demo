package geolocation

import (
	"strings"
)

// GridZoneMapper maps locations to electricity grid zones
type GridZoneMapper struct {
	zones map[string]GridZone
}

// NewGridZoneMapper creates a new grid zone mapper with predefined mappings
func NewGridZoneMapper() *GridZoneMapper {
	mapper := &GridZoneMapper{
		zones: make(map[string]GridZone),
	}
	
	// Initialize with common grid zones
	mapper.initializeZones()
	
	return mapper
}

// initializeZones sets up the mapping between countries/regions and grid zones
func (g *GridZoneMapper) initializeZones() {
	// European countries
	g.zones["DE"] = GridZone{Zone: "DE", Country: "Germany", Region: "", Description: "German electricity grid"}
	g.zones["FR"] = GridZone{Zone: "FR", Country: "France", Region: "", Description: "French electricity grid"}
	g.zones["NL"] = GridZone{Zone: "NL", Country: "Netherlands", Region: "", Description: "Dutch electricity grid"}
	g.zones["BE"] = GridZone{Zone: "BE", Country: "Belgium", Region: "", Description: "Belgian electricity grid"}
	g.zones["DK"] = GridZone{Zone: "DK", Country: "Denmark", Region: "", Description: "Danish electricity grid"}
	g.zones["NO"] = GridZone{Zone: "NO", Country: "Norway", Region: "", Description: "Norwegian electricity grid"}
	g.zones["SE"] = GridZone{Zone: "SE", Country: "Sweden", Region: "", Description: "Swedish electricity grid"}
	g.zones["FI"] = GridZone{Zone: "FI", Country: "Finland", Region: "", Description: "Finnish electricity grid"}
	g.zones["AT"] = GridZone{Zone: "AT", Country: "Austria", Region: "", Description: "Austrian electricity grid"}
	g.zones["CH"] = GridZone{Zone: "CH", Country: "Switzerland", Region: "", Description: "Swiss electricity grid"}
	g.zones["IT"] = GridZone{Zone: "IT", Country: "Italy", Region: "", Description: "Italian electricity grid"}
	g.zones["ES"] = GridZone{Zone: "ES", Country: "Spain", Region: "", Description: "Spanish electricity grid"}
	g.zones["PT"] = GridZone{Zone: "PT", Country: "Portugal", Region: "", Description: "Portuguese electricity grid"}
	g.zones["PL"] = GridZone{Zone: "PL", Country: "Poland", Region: "", Description: "Polish electricity grid"}
	g.zones["CZ"] = GridZone{Zone: "CZ", Country: "Czech Republic", Region: "", Description: "Czech electricity grid"}
	g.zones["HU"] = GridZone{Zone: "HU", Country: "Hungary", Region: "", Description: "Hungarian electricity grid"}
	g.zones["SK"] = GridZone{Zone: "SK", Country: "Slovakia", Region: "", Description: "Slovak electricity grid"}
	g.zones["SI"] = GridZone{Zone: "SI", Country: "Slovenia", Region: "", Description: "Slovenian electricity grid"}
	g.zones["HR"] = GridZone{Zone: "HR", Country: "Croatia", Region: "", Description: "Croatian electricity grid"}
	g.zones["BG"] = GridZone{Zone: "BG", Country: "Bulgaria", Region: "", Description: "Bulgarian electricity grid"}
	g.zones["RO"] = GridZone{Zone: "RO", Country: "Romania", Region: "", Description: "Romanian electricity grid"}
	g.zones["GR"] = GridZone{Zone: "GR", Country: "Greece", Region: "", Description: "Greek electricity grid"}
	g.zones["IE"] = GridZone{Zone: "IE", Country: "Ireland", Region: "", Description: "Irish electricity grid"}
	
	// UK regions
	g.zones["GB"] = GridZone{Zone: "GB", Country: "United Kingdom", Region: "", Description: "British electricity grid"}
	g.zones["UK"] = GridZone{Zone: "GB", Country: "United Kingdom", Region: "", Description: "British electricity grid"}
	
	// North America - simplified zones
	g.zones["US"] = GridZone{Zone: "US", Country: "United States", Region: "", Description: "US electricity grid"}
	g.zones["CA"] = GridZone{Zone: "CA", Country: "Canada", Region: "", Description: "Canadian electricity grid"}
	
	// Other major regions
	g.zones["JP"] = GridZone{Zone: "JP", Country: "Japan", Region: "", Description: "Japanese electricity grid"}
	g.zones["AU"] = GridZone{Zone: "AU", Country: "Australia", Region: "", Description: "Australian electricity grid"}
	g.zones["NZ"] = GridZone{Zone: "NZ", Country: "New Zealand", Region: "", Description: "New Zealand electricity grid"}
	g.zones["SG"] = GridZone{Zone: "SG", Country: "Singapore", Region: "", Description: "Singapore electricity grid"}
	g.zones["HK"] = GridZone{Zone: "HK", Country: "Hong Kong", Region: "", Description: "Hong Kong electricity grid"}
	g.zones["TW"] = GridZone{Zone: "TW", Country: "Taiwan", Region: "", Description: "Taiwan electricity grid"}
	g.zones["KR"] = GridZone{Zone: "KR", Country: "South Korea", Region: "", Description: "South Korean electricity grid"}
}

// MapToGridZone maps a location to its corresponding electricity grid zone
func (g *GridZoneMapper) MapToGridZone(location Location) GridZone {
	// First try country code
	if zone, exists := g.zones[strings.ToUpper(location.CountryCode)]; exists {
		zone.Region = location.Region
		return zone
	}
	
	// Try alternative country code mappings
	countryCode := g.normalizeCountryCode(location.CountryCode, location.Country)
	if zone, exists := g.zones[countryCode]; exists {
		zone.Region = location.Region
		return zone
	}
	
	// Default to German grid (Berlin fallback)
	return DefaultGridZone
}

// normalizeCountryCode handles alternative country code formats
func (g *GridZoneMapper) normalizeCountryCode(code, country string) string {
	code = strings.ToUpper(code)
	
	// Handle common alternatives
	switch code {
	case "UK":
		return "GB"
	case "EN":
		return "GB"
	}
	
	// Handle by country name if code is unclear
	countryUpper := strings.ToUpper(country)
	switch {
	case strings.Contains(countryUpper, "UNITED KINGDOM") || strings.Contains(countryUpper, "BRITAIN"):
		return "GB"
	case strings.Contains(countryUpper, "UNITED STATES") || strings.Contains(countryUpper, "USA"):
		return "US"
	}
	
	return code
}

// GetAvailableZones returns all available grid zones
func (g *GridZoneMapper) GetAvailableZones() []GridZone {
	zones := make([]GridZone, 0, len(g.zones))
	for _, zone := range g.zones {
		zones = append(zones, zone)
	}
	return zones
}

// AddCustomZone allows adding custom grid zone mappings
func (g *GridZoneMapper) AddCustomZone(countryCode string, zone GridZone) {
	g.zones[strings.ToUpper(countryCode)] = zone
}