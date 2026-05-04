package domain

type Coordinate struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

type Place struct {
	OsmID       int64      `json:"osm_id"`
	OsmType     string     `json:"osm_type"`
	OsmKey      string     `json:"osm_key"`
	OsmValue    string     `json:"osm_value"`
	Name        string     `json:"name"`
	Type        string     `json:"type"`
	Country     string     `json:"country"`
	CountryCode string     `json:"countrycode"`
	State       string     `json:"state"`
	County      string     `json:"county"`
	City        string     `json:"city"`
	District    string     `json:"district"`
	Locality    string     `json:"locality"`
	Street      string     `json:"street"`
	Postcode    string     `json:"postcode"`
	Coordinate  Coordinate `json:"coordinate"`
	Extent      []float64  `json:"extent,omitempty"`
}
