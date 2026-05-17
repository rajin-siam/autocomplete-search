package application

import (
	"context"
	"errors"
	"place-search/config"
	"place-search/internal/domain"
)

var ErrQueryTooShort = errors.New("query too short")

type SearchInput struct {
	Query string
}

type PlaceRepository interface {
	Search(ctx context.Context, query, fuzziness string, maxResults int) ([]domain.Place, error)
}

type SearchService struct {
	repo PlaceRepository
	cfg  config.SearchConfig
}

func NewSearchService(repo PlaceRepository, cfg config.SearchConfig) *SearchService {
	return &SearchService{repo: repo, cfg: cfg}
}

func (s *SearchService) Search(ctx context.Context, input SearchInput) ([]domain.Place, error) {
	if len([]rune(input.Query)) < s.cfg.MinChars {
		return nil, ErrQueryTooShort
	}
	return s.repo.Search(ctx, input.Query, s.cfg.Fuzziness, s.cfg.MaxResults)
}
