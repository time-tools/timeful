package models

// {"country_code":"US","country_name":"United States","city":"Los Angeles","postal":"90007","latitude":34.0294,"longitude":-118.2871,"IPv4":"207.151.52.114","state":"California"}

type Location struct {
	CountryCode string  `json:"country_code"`
	CountryName string  `json:"country_name"`
	City        string  `json:"city"`
	Postal      string  `json:"postal"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	State       string  `json:"state"`
}
