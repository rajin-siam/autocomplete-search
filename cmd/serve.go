package cmd

import (
	"log"
	"place-search/internal/application"
	esinfra "place-search/internal/infrastructure/elasticsearch"
	httphandler "place-search/internal/interfaces/http"

	"github.com/spf13/cobra"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the HTTP server",
	RunE: func(cmd *cobra.Command, args []string) error {
		esClient, err := esinfra.NewClient(cfg.Elasticsearch.Host)
		if err != nil {
			return err
		}

		repo := esinfra.NewPlaceRepository(esClient, cfg.Elasticsearch.Index)
		service := application.NewSearchService(repo, cfg.Search)
		searchHandler := httphandler.NewSearchHandler(service)
		server := httphandler.NewServer(cfg.Server.Port, searchHandler)

		log.Printf("server starting on port %d", cfg.Server.Port)
		return server.ListenAndServe()
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
}
