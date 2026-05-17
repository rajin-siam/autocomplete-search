package application

import (
	"context"
	"errors"
	"place-search/internal/domain"
	"testing"
)

// MockPlaceExtractor is a mock implementation of PlaceExtractor interface
type MockPlaceExtractor struct {
	ExtractPlacesFunc func() (<-chan domain.Place, error)
}

func (m *MockPlaceExtractor) ExtractPlaces() (<-chan domain.Place, error) {
	if m.ExtractPlacesFunc != nil {
		return m.ExtractPlacesFunc()
	}
	ch := make(chan domain.Place)
	close(ch)
	return ch, nil
}

// MockIndexManager is a mock implementation of IndexManager interface
type MockIndexManager struct {
	CreateIndexFunc func() error
	DeleteIndexFunc func() error
}

func (m *MockIndexManager) CreateIndex() error {
	if m.CreateIndexFunc != nil {
		return m.CreateIndexFunc()
	}
	return nil
}

func (m *MockIndexManager) DeleteIndex() error {
	if m.DeleteIndexFunc != nil {
		return m.DeleteIndexFunc()
	}
	return nil
}

// MockDocumentIndexer is a mock implementation of DocumentIndexer interface
type MockDocumentIndexer struct {
	IndexDocumentFunc func(doc interface{}) error
	CloseFunc         func() error
	StatsFunc         func() IndexStats
	indexed           int64
	failed            int64
}

func (m *MockDocumentIndexer) IndexDocument(doc interface{}) error {
	if m.IndexDocumentFunc != nil {
		return m.IndexDocumentFunc(doc)
	}
	m.indexed++
	return nil
}

func (m *MockDocumentIndexer) Close() error {
	if m.CloseFunc != nil {
		return m.CloseFunc()
	}
	return nil
}

func (m *MockDocumentIndexer) Stats() IndexStats {
	if m.StatsFunc != nil {
		return m.StatsFunc()
	}
	return IndexStats{
		Indexed: m.indexed,
		Failed:  m.failed,
	}
}

func TestIndexingService_IndexPlaces_SuccessfulIndexing(t *testing.T) {
	// Arrange
	testPlaces := []domain.Place{
		{Name: "Dhaka", Coordinate: domain.Coordinate{Lat: 23.7104, Lon: 90.4074}},
		{Name: "Chittagong", Coordinate: domain.Coordinate{Lat: 22.3569, Lon: 91.7832}},
		{Name: "Sylhet", Coordinate: domain.Coordinate{Lat: 24.8949, Lon: 91.8687}},
	}

	mockExtractor := &MockPlaceExtractor{
		ExtractPlacesFunc: func() (<-chan domain.Place, error) {
			ch := make(chan domain.Place, len(testPlaces))
			for _, place := range testPlaces {
				ch <- place
			}
			close(ch)
			return ch, nil
		},
	}

	mockManager := &MockIndexManager{
		CreateIndexFunc: func() error {
			return nil
		},
	}

	mockIndexer := &MockDocumentIndexer{
		IndexDocumentFunc: func(doc interface{}) error {
			return nil
		},
		StatsFunc: func() IndexStats {
			return IndexStats{Indexed: 3, Failed: 0}
		},
	}

	service := NewIndexingService(mockExtractor, mockManager, mockIndexer)
	ctx := context.Background()

	// Act
	stats, err := service.IndexPlaces(ctx)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if stats.Indexed != 3 {
		t.Errorf("Expected 3 indexed, got %d", stats.Indexed)
	}

	if stats.Failed != 0 {
		t.Errorf("Expected 0 failed, got %d", stats.Failed)
	}
}

func TestIndexingService_IndexPlaces_IndexCreationFails_ReturnsError(t *testing.T) {
	// Arrange
	expectedError := errors.New("failed to create index")

	mockExtractor := &MockPlaceExtractor{}

	mockManager := &MockIndexManager{
		CreateIndexFunc: func() error {
			return expectedError
		},
	}

	mockIndexer := &MockDocumentIndexer{}

	service := NewIndexingService(mockExtractor, mockManager, mockIndexer)
	ctx := context.Background()

	// Act
	stats, err := service.IndexPlaces(ctx)

	// Assert
	if err != expectedError {
		t.Errorf("Expected error '%v', got '%v'", expectedError, err)
	}

	if stats.Indexed != 0 || stats.Failed != 0 {
		t.Errorf("Expected empty stats, got %+v", stats)
	}
}

func TestIndexingService_IndexPlaces_ExtractionFails_ReturnsError(t *testing.T) {
	// Arrange
	expectedError := errors.New("database connection failed")

	mockExtractor := &MockPlaceExtractor{
		ExtractPlacesFunc: func() (<-chan domain.Place, error) {
			return nil, expectedError
		},
	}

	mockManager := &MockIndexManager{
		CreateIndexFunc: func() error {
			return nil
		},
	}

	mockIndexer := &MockDocumentIndexer{}

	service := NewIndexingService(mockExtractor, mockManager, mockIndexer)
	ctx := context.Background()

	// Act
	stats, err := service.IndexPlaces(ctx)

	// Assert
	if err != expectedError {
		t.Errorf("Expected error '%v', got '%v'", expectedError, err)
	}

	if stats.Indexed != 0 || stats.Failed != 0 {
		t.Errorf("Expected empty stats, got %+v", stats)
	}
}
