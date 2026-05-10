//go:build integration
// +build integration

package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"place-search/config"
	"place-search/internal/application"
	"place-search/internal/domain"
	"place-search/internal/infrastructure/elasticsearch"
	httpHandler "place-search/internal/interfaces/http"

	es "github.com/elastic/go-elasticsearch/v8"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const testIndexName = "places_test"

// Shared container for all tests
var sharedSetup *TestSetup

// GeoJSON response structures
type GeoJSONResponse struct {
	Type     string    `json:"type"`
	Features []Feature `json:"features"`
}

type Feature struct {
	Type       string                 `json:"type"`
	Geometry   Geometry               `json:"geometry"`
	Properties map[string]interface{} `json:"properties"`
}

type Geometry struct {
	Type        string    `json:"type"`
	Coordinates []float64 `json:"coordinates"`
}

type ErrorResponse struct {
	Message string `json:"message"`
}

// Test helper to setup Elasticsearch container and test data
type TestSetup struct {
	container testcontainers.Container
	client    *es.Client
	handler   *httpHandler.SearchHandler
	ctx       context.Context
}

// TestMain sets up shared container for all tests
func TestMain(m *testing.M) {
	log.Println("Setting up shared Elasticsearch container...")
	sharedSetup = setupElasticsearchContainer()
	log.Println("Container ready. Running tests...")

	code := m.Run()

	log.Println("Tearing down container...")
	if err := sharedSetup.container.Terminate(sharedSetup.ctx); err != nil {
		log.Printf("Failed to terminate container: %v", err)
	}

	os.Exit(code)
}

func setupElasticsearchContainer() *TestSetup {
	ctx := context.Background()

	// Start Elasticsearch container
	req := testcontainers.ContainerRequest{
		Image:        "docker.elastic.co/elasticsearch/elasticsearch:8.11.0",
		ExposedPorts: []string{"9200/tcp"},
		Env: map[string]string{
			"discovery.type":         "single-node",
			"xpack.security.enabled": "false",
			"ES_JAVA_OPTS":           "-Xms512m -Xmx512m",
		},
		WaitingFor: wait.ForHTTP("/").WithPort("9200").WithStartupTimeout(120 * time.Second),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		log.Fatalf("Failed to start Elasticsearch container: %v", err)
	}

	// Get container host and port
	host, err := container.Host(ctx)
	if err != nil {
		log.Fatalf("Failed to get container host: %v", err)
	}

	port, err := container.MappedPort(ctx, "9200")
	if err != nil {
		log.Fatalf("Failed to get container port: %v", err)
	}

	esURL := fmt.Sprintf("http://%s:%s", host, port.Port())

	// Create Elasticsearch client
	cfg := es.Config{
		Addresses: []string{esURL},
	}
	client, err := es.NewClient(cfg)
	if err != nil {
		log.Fatalf("Failed to create Elasticsearch client: %v", err)
	}

	// Wait for Elasticsearch to be ready
	time.Sleep(2 * time.Second)

	// Create index with mapping
	indexManager := elasticsearch.NewIndexManager(client, testIndexName)
	if err := indexManager.CreateIndex(); err != nil {
		log.Fatalf("Failed to create index: %v", err)
	}

	// Create repository and service
	repo := elasticsearch.NewPlaceRepository(client, testIndexName)
	searchConfig := config.SearchConfig{
		MinChars:   2,
		Fuzziness:  "AUTO",
		MaxResults: 15,
	}
	service := application.NewSearchService(repo, searchConfig)

	// Create HTTP handler
	handler := httpHandler.NewSearchHandler(service)

	return &TestSetup{
		container: container,
		client:    client,
		handler:   handler,
		ctx:       ctx,
	}
}

func cleanIndex(t *testing.T) {
	indexManager := elasticsearch.NewIndexManager(sharedSetup.client, testIndexName)
	if err := indexManager.CreateIndex(); err != nil {
		t.Fatalf("Failed to recreate index: %v", err)
	}
	time.Sleep(500 * time.Millisecond)
}

func indexTestData(t *testing.T, places []domain.Place) {
	bulkConfig := &elasticsearch.BulkIndexerConfig{
		NumWorkers: 2,
		FlushBytes: 1024 * 1024,
	}
	indexer, err := elasticsearch.NewBulkDocumentIndexer(sharedSetup.client, testIndexName, bulkConfig)
	if err != nil {
		t.Fatalf("Failed to create bulk indexer: %v", err)
	}

	for _, place := range places {
		if err := indexer.IndexDocument(place); err != nil {
			t.Fatalf("Failed to index document: %v", err)
		}
	}

	if err := indexer.Close(); err != nil {
		t.Fatalf("Failed to close indexer: %v", err)
	}

	// Wait for documents to be searchable
	time.Sleep(1 * time.Second)
}

// Test 1: Successful search with exact match
func TestAPI_Search_ExactMatch_ReturnsResults(t *testing.T) {
	cleanIndex(t)

	// Index test data
	testPlaces := []domain.Place{
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
	indexTestData(t, testPlaces)

	// Make request
	req := httptest.NewRequest(http.MethodGet, "/api?q=Dhaka", nil)
	w := httptest.NewRecorder()

	sharedSetup.handler.ServeHTTP(w, req)

	// Assert response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response GeoJSONResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.Type != "FeatureCollection" {
		t.Errorf("Expected type 'FeatureCollection', got '%s'", response.Type)
	}

	if len(response.Features) != 1 {
		t.Fatalf("Expected 1 feature, got %d", len(response.Features))
	}

	feature := response.Features[0]
	if feature.Properties["name"] != "Dhaka" {
		t.Errorf("Expected name 'Dhaka', got '%v'", feature.Properties["name"])
	}

	if feature.Geometry.Type != "Point" {
		t.Errorf("Expected geometry type 'Point', got '%s'", feature.Geometry.Type)
	}

	// Verify coordinates are [lon, lat]
	if len(feature.Geometry.Coordinates) != 2 {
		t.Fatalf("Expected 2 coordinates, got %d", len(feature.Geometry.Coordinates))
	}
	if feature.Geometry.Coordinates[0] != 90.4074 {
		t.Errorf("Expected longitude 90.4074, got %f", feature.Geometry.Coordinates[0])
	}
	if feature.Geometry.Coordinates[1] != 23.7104 {
		t.Errorf("Expected latitude 23.7104, got %f", feature.Geometry.Coordinates[1])
	}
}

// Test 2: Fuzzy search with typo
func TestAPI_Search_FuzzyMatch_ReturnsResults(t *testing.T) {
	cleanIndex(t)

	// Index test data
	testPlaces := []domain.Place{
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
	indexTestData(t, testPlaces)

	// Make request with typo
	req := httptest.NewRequest(http.MethodGet, "/api?q=dhka", nil)
	w := httptest.NewRecorder()

	sharedSetup.handler.ServeHTTP(w, req)

	// Assert response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response GeoJSONResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(response.Features) == 0 {
		t.Fatal("Expected fuzzy match to return results")
	}

	// Should find "Dhaka" despite typo
	feature := response.Features[0]
	if feature.Properties["name"] != "Dhaka" {
		t.Errorf("Expected fuzzy match to find 'Dhaka', got '%v'", feature.Properties["name"])
	}
}

// Test 3: Prefix search (autocomplete)
func TestAPI_Search_PrefixMatch_ReturnsResults(t *testing.T) {
	cleanIndex(t)

	// Index test data
	testPlaces := []domain.Place{
		{
			Name:        "Dhaka",
			Coordinate:  domain.Coordinate{Lat: 23.7104, Lon: 90.4074},
			OsmID:       1,
			OsmType:     "relation",
			Type:        "city",
			Country:     "Bangladesh",
			CountryCode: "BD",
		},
		{
			Name:        "Dinajpur",
			Coordinate:  domain.Coordinate{Lat: 25.6217, Lon: 88.6354},
			OsmID:       2,
			OsmType:     "relation",
			Type:        "city",
			Country:     "Bangladesh",
			CountryCode: "BD",
		},
	}
	indexTestData(t, testPlaces)

	// Make request with prefix
	req := httptest.NewRequest(http.MethodGet, "/api?q=dh", nil)
	w := httptest.NewRecorder()

	sharedSetup.handler.ServeHTTP(w, req)

	// Assert response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response GeoJSONResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Should find "Dhaka" (and possibly "Dinajpur" depending on fuzzy matching)
	if len(response.Features) == 0 {
		t.Fatal("Expected prefix match to return results")
	}

	foundDhaka := false
	for _, feature := range response.Features {
		if feature.Properties["name"] == "Dhaka" {
			foundDhaka = true
			break
		}
	}

	if !foundDhaka {
		t.Error("Expected to find 'Dhaka' in prefix search results")
	}
}

// Test 4: Query too short returns 400
func TestAPI_Search_QueryTooShort_Returns400(t *testing.T) {
	// Make request with single character (min_chars = 2)
	req := httptest.NewRequest(http.MethodGet, "/api?q=d", nil)
	w := httptest.NewRecorder()

	sharedSetup.handler.ServeHTTP(w, req)

	// Assert response
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}

	var response ErrorResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.Message != "query too short" {
		t.Errorf("Expected error message 'query too short', got '%s'", response.Message)
	}
}

// Test 5: Missing query parameter returns 400
func TestAPI_Search_MissingQueryParam_Returns400(t *testing.T) {
	// Make request without query parameter
	req := httptest.NewRequest(http.MethodGet, "/api", nil)
	w := httptest.NewRecorder()

	sharedSetup.handler.ServeHTTP(w, req)

	// Assert response
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}

	var response ErrorResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.Message != "missing query parameter 'q'" {
		t.Errorf("Expected error message about missing parameter, got '%s'", response.Message)
	}
}

// Test 6: No results returns empty features array
func TestAPI_Search_NoResults_ReturnsEmptyFeatures(t *testing.T) {
	cleanIndex(t)

	// Don't index any data

	// Make request
	req := httptest.NewRequest(http.MethodGet, "/api?q=nonexistent", nil)
	w := httptest.NewRecorder()

	sharedSetup.handler.ServeHTTP(w, req)

	// Assert response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response GeoJSONResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.Type != "FeatureCollection" {
		t.Errorf("Expected type 'FeatureCollection', got '%s'", response.Type)
	}

	if len(response.Features) != 0 {
		t.Errorf("Expected 0 features, got %d", len(response.Features))
	}
}

// Test 7: Unicode query (Bengali characters)
func TestAPI_Search_UnicodeQuery_ReturnsResults(t *testing.T) {
	cleanIndex(t)

	// Index test data with Bengali name
	testPlaces := []domain.Place{
		{
			Name:        "ঢাকা",
			Coordinate:  domain.Coordinate{Lat: 23.7104, Lon: 90.4074},
			OsmID:       123456,
			OsmType:     "relation",
			Type:        "city",
			Country:     "Bangladesh",
			CountryCode: "BD",
		},
	}
	indexTestData(t, testPlaces)

	// Make request with Bengali characters
	req := httptest.NewRequest(http.MethodGet, "/api?q=ঢাকা", nil)
	w := httptest.NewRecorder()

	sharedSetup.handler.ServeHTTP(w, req)

	// Assert response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response GeoJSONResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(response.Features) == 0 {
		t.Fatal("Expected to find results for Bengali query")
	}

	feature := response.Features[0]
	if feature.Properties["name"] != "ঢাকা" {
		t.Errorf("Expected name 'ঢাকা', got '%v'", feature.Properties["name"])
	}
}

// Test 10: Special characters in query
func TestAPI_Search_SpecialCharacters_HandlesCorrectly(t *testing.T) {
	cleanIndex(t)

	// Index test data
	testPlaces := []domain.Place{
		{
			Name:        "Cox's Bazar",
			Coordinate:  domain.Coordinate{Lat: 21.4272, Lon: 92.0058},
			OsmID:       123456,
			OsmType:     "relation",
			Type:        "city",
			Country:     "Bangladesh",
			CountryCode: "BD",
		},
	}
	indexTestData(t, testPlaces)

	// Make request with special characters
	req := httptest.NewRequest(http.MethodGet, "/api?q=Cox's", nil)
	w := httptest.NewRecorder()

	sharedSetup.handler.ServeHTTP(w, req)

	// Assert response - should not crash
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}
