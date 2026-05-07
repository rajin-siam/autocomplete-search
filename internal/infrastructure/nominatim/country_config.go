package nominatim

type CountryConfig struct {
	Country     string
	CountryCode string
}

func DefaultCountryConfig() *CountryConfig {
	return &CountryConfig{
		Country:     "Bangladesh",
		CountryCode: "BD",
	}
}
