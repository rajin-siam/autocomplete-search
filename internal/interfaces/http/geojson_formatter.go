package http

import (
	"encoding/json"
	"net/http"
	"place-search/internal/domain"
)

type geometry struct {
	Type        string    `json:"type"`
	Coordinates []float64 `json:"coordinates"`
}

type feature struct {
	Type       string         `json:"type"`
	Geometry   geometry       `json:"geometry"`
	Properties map[string]any `json:"properties"`
}

type featureCollection struct {
	Type     string    `json:"type"`
	Features []feature `json:"features"`
}

func formatGeoJSON(places []domain.Place) []byte {
	features := make([]feature, 0, len(places))

	for _, p := range places {
		props := map[string]any{
			"osm_id":      p.OsmID,
			"osm_type":    p.OsmType,
			"osm_key":     p.OsmKey,
			"osm_value":   p.OsmValue,
			"name":        p.Name,
			"type":        p.Type,
			"country":     p.Country,
			"countrycode": p.CountryCode,
		}

		if p.State != "" {
			props["state"] = p.State
		}
		if p.County != "" {
			props["county"] = p.County
		}
		if p.City != "" {
			props["city"] = p.City
		}
		if p.District != "" {
			props["district"] = p.District
		}
		if p.Locality != "" {
			props["locality"] = p.Locality
		}
		if p.Street != "" {
			props["street"] = p.Street
		}
		if p.Postcode != "" {
			props["postcode"] = p.Postcode
		}
		if len(p.Extent) == 4 {
			props["extent"] = p.Extent
		}

		f := feature{
			Type: "Feature",
			Geometry: geometry{
				Type:        "Point",
				Coordinates: []float64{p.Coordinate.Lon, p.Coordinate.Lat},
			},
			Properties: props,
		}
		features = append(features, f)
	}

	fc := featureCollection{Type: "FeatureCollection", Features: features}
	out, _ := json.Marshal(fc)
	return out
}

func writeJSON(w http.ResponseWriter, status int, data []byte) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(status)
	w.Write(data)
}
