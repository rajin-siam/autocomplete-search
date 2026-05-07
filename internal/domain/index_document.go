package domain

type IndexDocument struct {
	Name        string             `json:"name"`
	Coordinate  map[string]float64 `json:"coordinate"`
	OsmID       int64              `json:"osm_id"`
	OsmType     string             `json:"osm_type"`
	OsmKey      string             `json:"osm_key"`
	OsmValue    string             `json:"osm_value"`
	Type        string             `json:"type"`
	Country     string             `json:"country"`
	CountryCode string             `json:"countrycode"`
	Postcode    *string            `json:"postcode,omitempty"`
	Extent      []float64          `json:"extent,omitempty"`
	State       string             `json:"state,omitempty"`
	County      string             `json:"county,omitempty"`
	City        string             `json:"city,omitempty"`
	District    string             `json:"district,omitempty"`
	Locality    string             `json:"locality,omitempty"`
}
