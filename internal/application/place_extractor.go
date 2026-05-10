package application

import "place-search/internal/domain"

type PlaceExtractor interface {
	ExtractPlaces() (<-chan domain.Place, error)
}
