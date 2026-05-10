package application

import (
	"context"
	"place-search/config"
	"place-search/internal/domain"
	"testing"
)

// MockPlaceRepository is a mock implementation of PlaceRepository interface
type MockPlaceRepository struct {
	SearchFunc func(ctx context.Context, query, fuzziness string, maxResults int) ([]domain.Place, error)
}

func (m *MockPlaceRepository) Search(ctx context.Context, query, fuzziness string, maxResults int) ([]domain.Place, error) {
	if m.SearchFunc != nil {
		return m.SearchFunc(ctx, query, fuzziness, maxResults)
	}
	return nil, nil
}

func TestSearchService_Search_ValidQuery_ReturnsResults(t *testing.T) {
	// Arrange
	expectedPlaces := []domain.Place{
		{
			Name:        "Dhaka",
			Coordinate:  domain.Coordinate{Lat: 23.7104, Lon: 90.4074},
			OsmID:       123456,
			OsmType:     "relation",
			Type:        "city",
			Country:     "Bangladesh",
			CountryCode: "BD",
		},
	}

	mockRepo := &MockPlaceRepository{
		SearchFunc: func(ctx context.Context, query, fuzziness string, maxResults int) ([]domain.Place, error) {
			if query != "dhaka" {
				t.Errorf("Expected query 'dhaka', got '%s'", query)
			}
			if fuzziness != "AUTO" {
				t.Errorf("Expected fuzziness 'AUTO', got '%s'", fuzziness)
			}
			if maxResults != 15 {
				t.Errorf("Expected maxResults 15, got %d", maxResults)
			}
			return expectedPlaces, nil
		},
	}

	cfg := config.SearchConfig{
		MinChars:   2,
		Fuzziness:  "AUTO",
		MaxResults: 15,
	}

	service := NewSearchService(mockRepo, cfg)
	ctx := context.Background()

	// Act
	results, err := service.Search(ctx, SearchInput{Query: "dhaka"})

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}

	if results[0].Name != "Dhaka" {
		t.Errorf("Expected place name 'Dhaka', got '%s'", results[0].Name)
	}
}

func TestSearchService_Search_QueryTooShort_ReturnsError(t *testing.T) {
	// Arrange
	mockRepo := &MockPlaceRepository{
		SearchFunc: func(ctx context.Context, query, fuzziness string, maxResults int) ([]domain.Place, error) {
			t.Error("Repository should not be called for short queries")
			return nil, nil
		},
	}

	cfg := config.SearchConfig{
		MinChars:   2,
		Fuzziness:  "AUTO",
		MaxResults: 15,
	}

	service := NewSearchService(mockRepo, cfg)
	ctx := context.Background()

	// Act
	results, err := service.Search(ctx, SearchInput{Query: "d"})

	// Assert
	if err != ErrQueryTooShort {
		t.Errorf("Expected ErrQueryTooShort, got %v", err)
	}

	if results != nil {
		t.Errorf("Expected nil results, got %v", results)
	}
}

func TestSearchService_Search_EmptyQuery_ReturnsError(t *testing.T) {
	// Arrange
	mockRepo := &MockPlaceRepository{}
	cfg := config.SearchConfig{
		MinChars:   2,
		Fuzziness:  "AUTO",
		MaxResults: 15,
	}

	service := NewSearchService(mockRepo, cfg)
	ctx := context.Background()

	// Act
	results, err := service.Search(ctx, SearchInput{Query: ""})

	// Assert
	if err != ErrQueryTooShort {
		t.Errorf("Expected ErrQueryTooShort, got %v", err)
	}

	if results != nil {
		t.Errorf("Expected nil results, got %v", results)
	}
}

func TestSearchService_Search_QueryAtMinCharsThreshold_ReturnsResults(t *testing.T) {
	// Arrange
	expectedPlaces := []domain.Place{
		{Name: "Dhaka"},
		{Name: "Dinajpur"},
	}

	mockRepo := &MockPlaceRepository{
		SearchFunc: func(ctx context.Context, query, fuzziness string, maxResults int) ([]domain.Place, error) {
			return expectedPlaces, nil
		},
	}

	cfg := config.SearchConfig{
		MinChars:   2,
		Fuzziness:  "AUTO",
		MaxResults: 15,
	}

	service := NewSearchService(mockRepo, cfg)
	ctx := context.Background()

	// Act
	results, err := service.Search(ctx, SearchInput{Query: "dh"})

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}
}

func TestSearchService_Search_NoResults_ReturnsEmptySlice(t *testing.T) {
	// Arrange
	mockRepo := &MockPlaceRepository{
		SearchFunc: func(ctx context.Context, query, fuzziness string, maxResults int) ([]domain.Place, error) {
			return []domain.Place{}, nil
		},
	}

	cfg := config.SearchConfig{
		MinChars:   2,
		Fuzziness:  "AUTO",
		MaxResults: 15,
	}

	service := NewSearchService(mockRepo, cfg)
	ctx := context.Background()

	// Act
	results, err := service.Search(ctx, SearchInput{Query: "nonexistentplace"})

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(results) != 0 {
		t.Errorf("Expected 0 results, got %d", len(results))
	}
}

func TestSearchService_Search_UnicodeQuery_HandlesCorrectly(t *testing.T) {
	// Arrange
	expectedPlaces := []domain.Place{
		{Name: "ঢাকা", Country: "Bangladesh"},
	}

	mockRepo := &MockPlaceRepository{
		SearchFunc: func(ctx context.Context, query, fuzziness string, maxResults int) ([]domain.Place, error) {
			return expectedPlaces, nil
		},
	}

	cfg := config.SearchConfig{
		MinChars:   2,
		Fuzziness:  "AUTO",
		MaxResults: 15,
	}

	service := NewSearchService(mockRepo, cfg)
	ctx := context.Background()

	// Act
	results, err := service.Search(ctx, SearchInput{Query: "ঢাকা"})

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}

	if results[0].Name != "ঢাকা" {
		t.Errorf("Expected place name 'ঢাকা', got '%s'", results[0].Name)
	}
}
