package cmd

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esutil"
	_ "github.com/lib/pq"
	"github.com/spf13/cobra"
)

var rankToLayer = map[int]string{
	4: "country", 8: "state", 12: "county",
	16: "city", 17: "city", 18: "city", 19: "city", 20: "city", 21: "city",
	22: "district", 23: "district", 24: "district", 25: "district",
	26: "locality", 27: "locality", 28: "locality", 29: "locality",
	30: "house",
}

var rankToField = map[int]string{
	8: "state", 12: "county",
	16: "city", 17: "city", 18: "city", 19: "city", 20: "city",
	22: "district", 23: "district", 24: "district", 25: "district",
	26: "locality", 27: "locality", 28: "locality",
}

type HStore map[string]string

func getName(h HStore) string {
	if v, ok := h["name:en"]; ok && v != "" {
		return v
	}
	if v, ok := h["name"]; ok && v != "" {
		return v
	}
	if v, ok := h["name:bn"]; ok && v != "" {
		return v
	}
	for _, v := range h {
		if v != "" {
			return v
		}
	}
	return ""
}

func parseHStore(raw string) HStore {
	result := HStore{}
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return result
	}
	var i int
	for i < len(raw) {
		if raw[i] != '"' {
			break
		}
		i++
		keyStart := i
		for i < len(raw) && raw[i] != '"' {
			i++
		}
		key := raw[keyStart:i]
		i++
		i += 2
		if i >= len(raw) {
			break
		}
		if raw[i] == '"' {
			i++
			valStart := i
			for i < len(raw) && raw[i] != '"' {
				if raw[i] == '\\' {
					i++
				}
				i++
			}
			result[key] = raw[valStart:i]
			i++
		} else {
			i += 4
		}
		for i < len(raw) && (raw[i] == ',' || raw[i] == ' ') {
			i++
		}
	}
	return result
}

type indexDocument struct {
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

func createIndex(es *elasticsearch.Client, index string) error {
	res, _ := es.Indices.Delete([]string{index})
	if res != nil {
		res.Body.Close()
	}
	mapping := `{
		"mappings": {
			"properties": {
				"name":        {"type": "search_as_you_type"},
				"coordinate":  {"type": "geo_point",  "index": false},
				"osm_id":      {"type": "long",        "index": false},
				"osm_type":    {"type": "keyword",     "index": false},
				"osm_key":     {"type": "keyword",     "index": false},
				"osm_value":   {"type": "keyword",     "index": false},
				"type":        {"type": "keyword",     "index": false},
				"country":     {"type": "keyword",     "index": false},
				"countrycode": {"type": "keyword",     "index": false},
				"state":       {"type": "keyword",     "index": false},
				"county":      {"type": "keyword",     "index": false},
				"city":        {"type": "keyword",     "index": false},
				"district":    {"type": "keyword",     "index": false},
				"locality":    {"type": "keyword",     "index": false},
				"street":      {"type": "keyword",     "index": false},
				"postcode":    {"type": "keyword",     "index": false},
				"extent":      {"type": "object",      "enabled": false}
			}
		}
	}`
	res, err := es.Indices.Create(index, es.Indices.Create.WithBody(strings.NewReader(mapping)))
	if err != nil {
		return err
	}
	defer res.Body.Close()
	fmt.Println("Index created.")
	return nil
}

func fetchAddressParts(db *sql.DB, placeID int64) (map[string]string, error) {
	rows, err := db.Query(`
		SELECT p.rank_address, p.name
		FROM place_addressline al
		JOIN placex p ON p.place_id = al.address_place_id
		WHERE al.place_id = $1
		  AND al.isaddress = true
		  AND p.name IS NOT NULL
		  AND p.rank_address BETWEEN 8 AND 28
		ORDER BY p.rank_address
	`, placeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	parts := map[string]string{}
	for rows.Next() {
		var rank int
		var rawName string
		if err := rows.Scan(&rank, &rawName); err != nil {
			continue
		}
		field, ok := rankToField[rank]
		if !ok {
			continue
		}
		if _, already := parts[field]; already {
			continue
		}
		h := parseHStore(rawName)
		name := getName(h)
		if name != "" {
			parts[field] = name
		}
	}
	return parts, nil
}

func generateDocs(db *sql.DB) <-chan indexDocument {
	ch := make(chan indexDocument, 100)
	go func() {
		defer close(ch)
		rows, err := db.Query(`
			SELECT
				place_id,
				osm_type,
				osm_id,
				class,
				type,
				name,
				country_code,
				postcode,
				rank_address,
				importance,
				ST_X(ST_Centroid(geometry)) as lon,
				ST_Y(ST_Centroid(geometry)) as lat,
				ST_XMin(geometry) as min_lon,
				ST_YMin(geometry) as min_lat,
				ST_XMax(geometry) as max_lon,
				ST_YMax(geometry) as max_lat
			FROM placex
			WHERE name IS NOT NULL
			  AND linked_place_id IS NULL
			  AND indexed_status = 0
		`)
		if err != nil {
			log.Println("Query error:", err)
			return
		}
		defer rows.Close()
		for rows.Next() {
			var (
				placeID                        int64
				osmType, class, typ            string
				osmID                          int64
				rawName                        string
				countryCode, postcode          sql.NullString
				rankAddress                    int
				importance                     sql.NullFloat64
				lon, lat                       float64
				minLon, minLat, maxLon, maxLat float64
			)
			err := rows.Scan(
				&placeID, &osmType, &osmID, &class, &typ,
				&rawName, &countryCode, &postcode,
				&rankAddress, &importance,
				&lon, &lat,
				&minLon, &minLat, &maxLon, &maxLat,
			)
			if err != nil {
				log.Println("Scan error:", err)
				continue
			}
			h := parseHStore(rawName)
			name := getName(h)
			if name == "" {
				continue
			}
			address, err := fetchAddressParts(db, placeID)
			if err != nil {
				log.Println("Address fetch error:", err)
			}
			var extent []float64
			if minLon != maxLon || minLat != maxLat {
				extent = []float64{minLon, maxLat, maxLon, minLat}
			}
			layer := rankToLayer[rankAddress]
			if layer == "" {
				layer = "locality"
			}
			doc := indexDocument{
				Name:        name,
				Coordinate:  map[string]float64{"lat": lat, "lon": lon},
				OsmID:       osmID,
				OsmType:     strings.TrimSpace(osmType),
				OsmKey:      class,
				OsmValue:    typ,
				Type:        layer,
				Country:     "Bangladesh",
				CountryCode: "BD",
				Extent:      extent,
			}
			if postcode.Valid {
				doc.Postcode = &postcode.String
			}
			doc.State = address["state"]
			doc.County = address["county"]
			doc.City = address["city"]
			doc.District = address["district"]
			doc.Locality = address["locality"]
			ch <- doc
		}
	}()
	return ch
}

var indexCmd = &cobra.Command{
	Use:   "index",
	Short: "Index places from Nominatim PostgreSQL into Elasticsearch",
	RunE: func(cmd *cobra.Command, args []string) error {
		db := cfg.Database
		dsn := fmt.Sprintf("host=%s port=%d dbname=%s user=%s password=%s sslmode=disable",
			db.Host, db.Port, db.Name, db.User, db.Password)

		pgDB, err := sql.Open("postgres", dsn)
		if err != nil {
			return err
		}
		defer pgDB.Close()
		if err := pgDB.Ping(); err != nil {
			return fmt.Errorf("cannot reach PostgreSQL: %w", err)
		}
		fmt.Println("Connected to PostgreSQL.")

		es, err := elasticsearch.NewClient(elasticsearch.Config{
			Addresses: []string{cfg.Elasticsearch.Host},
		})
		if err != nil {
			return err
		}
		fmt.Println("Connected to Elasticsearch.")

		if err := createIndex(es, cfg.Elasticsearch.Index); err != nil {
			return fmt.Errorf("failed to create index: %w", err)
		}

		bi, err := esutil.NewBulkIndexer(esutil.BulkIndexerConfig{
			Index:      cfg.Elasticsearch.Index,
			Client:     es,
			NumWorkers: 4,
			FlushBytes: 5_000_000,
		})
		if err != nil {
			return err
		}

		fmt.Println("Extracting and indexing...")
		for doc := range generateDocs(pgDB) {
			data, err := json.Marshal(doc)
			if err != nil {
				log.Println("Marshal error:", err)
				continue
			}
			bi.Add(context.Background(), esutil.BulkIndexerItem{
				Action: "index",
				Body:   strings.NewReader(string(data)),
				OnFailure: func(ctx context.Context, item esutil.BulkIndexerItem, res esutil.BulkIndexerResponseItem, err error) {
					if err != nil {
						log.Println("Bulk item error:", err)
					}
				},
			})
		}

		if err := bi.Close(context.Background()); err != nil {
			return err
		}
		stats := bi.Stats()
		fmt.Printf("Indexed: %d, Failed: %d\n", stats.NumIndexed, stats.NumFailed)
		fmt.Println("Done.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(indexCmd)
}
