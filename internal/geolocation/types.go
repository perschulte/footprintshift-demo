package geolocation

// Location represents geographical location data
type Location struct {
	IP          string  `json:"ip"`
	Country     string  `json:"country"`
	CountryCode string  `json:"country_code"`
	Region      string  `json:"region"`
	RegionCode  string  `json:"region_code"`
	City        string  `json:"city"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	Timezone    string  `json:"timezone"`
}

// GridZone represents an electricity grid zone mapping
type GridZone struct {
	Zone        string `json:"zone"`
	Country     string `json:"country"`
	Region      string `json:"region"`
	Description string `json:"description"`
}

// LocationWithZone combines location data with grid zone information
type LocationWithZone struct {
	Location Location `json:"location"`
	GridZone GridZone `json:"grid_zone"`
}

// Default locations for fallback
var (
	DefaultLocation = Location{
		IP:          "0.0.0.0",
		Country:     "Germany",
		CountryCode: "DE",
		Region:      "Berlin",
		RegionCode:  "BE",
		City:        "Berlin",
		Latitude:    52.5200,
		Longitude:   13.4050,
		Timezone:    "Europe/Berlin",
	}

	DefaultGridZone = GridZone{
		Zone:        "DE",
		Country:     "Germany", 
		Region:      "Berlin",
		Description: "German electricity grid",
	}
)

// ipapi.co API response structure
type IpapiResponse struct {
	IP                 string  `json:"ip"`
	Network            string  `json:"network"`
	Version            string  `json:"version"`
	City               string  `json:"city"`
	Region             string  `json:"region"`
	RegionCode         string  `json:"region_code"`
	Country            string  `json:"country"`
	CountryName        string  `json:"country_name"`
	CountryCode        string  `json:"country_code"`
	CountryCodeISO3    string  `json:"country_code_iso3"`
	CountryCapital     string  `json:"country_capital"`
	CountryTLD         string  `json:"country_tld"`
	ContinentCode      string  `json:"continent_code"`
	InEU               bool    `json:"in_eu"`
	Postal             string  `json:"postal"`
	Latitude           float64 `json:"latitude"`
	Longitude          float64 `json:"longitude"`
	Timezone           string  `json:"timezone"`
	UTCOffset          string  `json:"utc_offset"`
	CountryCallingCode string  `json:"country_calling_code"`
	Currency           string  `json:"currency"`
	CurrencyName       string  `json:"currency_name"`
	Languages          string  `json:"languages"`
	CountryArea        float64 `json:"country_area"`
	CountryPopulation  int64   `json:"country_population"`
	ASN                string  `json:"asn"`
	ORG                string  `json:"org"`
}