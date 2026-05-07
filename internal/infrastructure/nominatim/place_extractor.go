package nominatim

import (
	"log"
	"place-search/internal/domain"
	"place-search/internal/infrastructure/hstore"
	"strings"
)

type PlaceExtractorImpl struct {
	db            PlaceDatabase
	rankConfig    *RankConfig
	countryConfig *CountryConfig
}

func NewPlaceExtractor(db PlaceDatabase, rankConfig *RankConfig, countryConfig *CountryConfig) *PlaceExtractorImpl {
	return &PlaceExtractorImpl{
		db:            db,
		rankConfig:    rankConfig,
		countryConfig: countryConfig,
	}
}

func (e *PlaceExtractorImpl) ExtractPlaces() (<-chan domain.IndexDocument, error) {
	ch := make(chan domain.IndexDocument, 100)
	
	go func() {
		defer close(ch)
		
		places, err := e.db.QueryPlaces()
		if err != nil {
			log.Println("Query places error:", err)
			return
		}
		
		for _, row := range places {
			doc, err := e.transformPlace(row)
			if err != nil {
				log.Println("Transform error:", err)
				continue
			}
			ch <- doc
		}
	}()
	
	return ch, nil
}

func (e *PlaceExtractorImpl) transformPlace(row PlaceRow) (domain.IndexDocument, error) {
	h := hstore.Parse(row.RawName)
	name := hstore.GetName(h)
	if name == "" {
		return domain.IndexDocument{}, nil
	}
	
	addressParts, err := e.fetchAddressParts(row.PlaceID)
	if err != nil {
		log.Println("Address fetch error:", err)
	}
	
	var extent []float64
	if row.MinLon != row.MaxLon || row.MinLat != row.MaxLat {
		extent = []float64{row.MinLon, row.MaxLat, row.MaxLon, row.MinLat}
	}
	
	layer := e.rankConfig.RankToLayer[row.RankAddress]
	if layer == "" {
		layer = "locality"
	}
	
	doc := domain.IndexDocument{
		Name:        name,
		Coordinate:  map[string]float64{"lat": row.Lat, "lon": row.Lon},
		OsmID:       row.OsmID,
		OsmType:     strings.TrimSpace(row.OsmType),
		OsmKey:      row.Class,
		OsmValue:    row.Type,
		Type:        layer,
		Country:     e.countryConfig.Country,
		CountryCode: e.countryConfig.CountryCode,
		Extent:      extent,
		Postcode:    row.Postcode,
		State:       addressParts["state"],
		County:      addressParts["county"],
		City:        addressParts["city"],
		District:    addressParts["district"],
		Locality:    addressParts["locality"],
	}
	
	return doc, nil
}

func (e *PlaceExtractorImpl) fetchAddressParts(placeID int64) (map[string]string, error) {
	parts, err := e.db.QueryAddressParts(placeID)
	if err != nil {
		return nil, err
	}
	
	result := map[string]string{}
	for _, part := range parts {
		field, ok := e.rankConfig.RankToField[part.Rank]
		if !ok {
			continue
		}
		if _, already := result[field]; already {
			continue
		}
		h := hstore.Parse(part.Name)
		name := hstore.GetName(h)
		if name != "" {
			result[field] = name
		}
	}
	
	return result, nil
}
