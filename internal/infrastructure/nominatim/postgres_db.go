package nominatim

import (
	"database/sql"
)

type PostgresPlaceDB struct {
	db *sql.DB
}

func NewPostgresPlaceDB(db *sql.DB) *PostgresPlaceDB {
	return &PostgresPlaceDB{db: db}
}

func (p *PostgresPlaceDB) QueryPlaces() ([]PlaceRow, error) {
	rows, err := p.db.Query(`
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
		return nil, err
	}
	defer rows.Close()

	var places []PlaceRow
	for rows.Next() {
		var p PlaceRow
		var countryCode, postcode sql.NullString
		
		err := rows.Scan(
			&p.PlaceID, &p.OsmType, &p.OsmID, &p.Class, &p.Type,
			&p.RawName, &countryCode, &postcode,
			&p.RankAddress,
			&p.Lon, &p.Lat,
			&p.MinLon, &p.MinLat, &p.MaxLon, &p.MaxLat,
		)
		if err != nil {
			continue
		}
		
		if countryCode.Valid {
			p.CountryCode = countryCode.String
		}
		if postcode.Valid {
			p.Postcode = &postcode.String
		}
		
		places = append(places, p)
	}
	
	return places, rows.Err()
}

func (p *PostgresPlaceDB) QueryAddressParts(placeID int64) ([]AddressPart, error) {
	rows, err := p.db.Query(`
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

	var parts []AddressPart
	for rows.Next() {
		var part AddressPart
		if err := rows.Scan(&part.Rank, &part.Name); err != nil {
			continue
		}
		parts = append(parts, part)
	}
	
	return parts, rows.Err()
}
