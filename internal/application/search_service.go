package application

import (
	"context"
	"place-search/internal/domain"
)

type PlaceRepository interface {
	Search(ctx context.Context, q domain.SearchQuery) ([]domain.Place, error)
}

type SearchService struct {
	repo PlaceRepository
}

func NewSearchService(repo PlaceRepository) *SearchService {
	return &SearchService{repo: repo}
}

func (s *SearchService) Search(ctx context.Context, q domain.SearchQuery) ([]domain.Place, error) {
	return s.repo.Search(ctx, q)
}
