package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"place-search/internal/domain"

	"github.com/elastic/go-elasticsearch/v8"
)

type PlaceRepository struct {
	client *elasticsearch.Client
	index  string
}

func NewPlaceRepository(client *elasticsearch.Client, index string) *PlaceRepository {
	return &PlaceRepository{client: client, index: index}
}

type esResponse struct {
	Hits struct {
		Hits []struct {
			Source esSource `json:"_source"`
		} `json:"hits"`
	} `json:"hits"`
}

type esCoordinate struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

type esSource struct {
	Name        string       `json:"name"`
	OsmID       int64        `json:"osm_id"`
	OsmType     string       `json:"osm_type"`
	OsmKey      string       `json:"osm_key"`
	OsmValue    string       `json:"osm_value"`
	Type        string       `json:"type"`
	Country     string       `json:"country"`
	CountryCode string       `json:"countrycode"`
	State       string       `json:"state"`
	County      string       `json:"county"`
	City        string       `json:"city"`
	District    string       `json:"district"`
	Locality    string       `json:"locality"`
	Street      string       `json:"street"`
	Postcode    string       `json:"postcode"`
	Coordinate  esCoordinate `json:"coordinate"`
	Extent      []float64    `json:"extent"`
}

func (r *PlaceRepository) Search(ctx context.Context, query, fuzziness string, maxResults int) ([]domain.Place, error) {
	body, err := buildQuery(query, fuzziness, maxResults)
	if err != nil {
		return nil, err
	}

	res, err := r.client.Search(
		r.client.Search.WithContext(ctx),
		r.client.Search.WithIndex(r.index),
		r.client.Search.WithBody(bytes.NewReader(body)),
	)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("elasticsearch error: %s", res.String())
	}

	var esRes esResponse
	if err := json.NewDecoder(res.Body).Decode(&esRes); err != nil {
		return nil, err
	}

	places := make([]domain.Place, 0, len(esRes.Hits.Hits))
	for _, hit := range esRes.Hits.Hits {
		s := hit.Source
		places = append(places, domain.Place{
			Name:        s.Name,
			OsmID:       s.OsmID,
			OsmType:     s.OsmType,
			OsmKey:      s.OsmKey,
			OsmValue:    s.OsmValue,
			Type:        s.Type,
			Country:     s.Country,
			CountryCode: s.CountryCode,
			State:       s.State,
			County:      s.County,
			City:        s.City,
			District:    s.District,
			Locality:    s.Locality,
			Street:      s.Street,
			Postcode:    s.Postcode,
			Coordinate: domain.Coordinate{
				Lat: s.Coordinate.Lat,
				Lon: s.Coordinate.Lon,
			},
			Extent: s.Extent,
		})
	}

	return places, nil
}
