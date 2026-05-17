package nominatim

type PlaceRow struct {
	PlaceID    int64
	OsmType    string
	OsmID      int64
	Class      string
	Type       string
	RawName    string
	CountryCode string
	Postcode    *string
	RankAddress int
	Lon         float64
	Lat         float64
	MinLon      float64
	MinLat      float64
	MaxLon      float64
	MaxLat      float64
}

type AddressPart struct {
	Rank int
	Name string
}

type PlaceDatabase interface {
	QueryPlaces() ([]PlaceRow, error)
	QueryAddressParts(placeID int64) ([]AddressPart, error)
}
