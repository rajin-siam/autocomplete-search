package main

import (
	"log"
	"place-search/config"
	"place-search/internal/application"
	esinfra "place-search/internal/infrastructure/elasticsearch"
	httphandler "place-search/internal/interfaces/http"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	esClient, err := esinfra.NewClient(cfg.Elasticsearch.Host)
	if err != nil {
		log.Fatalf("failed to create elasticsearch client: %v", err)
	}

	repo := esinfra.NewPlaceRepository(esClient, cfg.Elasticsearch.Index, cfg.Search.Fuzziness, cfg.Search.MaxResults)
	service := application.NewSearchService(repo)
	searchHandler := httphandler.NewSearchHandler(service, cfg.Search.MinChars)
	server := httphandler.NewServer(cfg.Server.Port, searchHandler)

	log.Printf("server starting on port %d", cfg.Server.Port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
