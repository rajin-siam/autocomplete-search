package cmd

import (
	"context"
	"database/sql"
	"fmt"

	"place-search/internal/application"
	esinfra "place-search/internal/infrastructure/elasticsearch"
	"place-search/internal/infrastructure/nominatim"

	"github.com/elastic/go-elasticsearch/v8"
	_ "github.com/lib/pq"
	"github.com/spf13/cobra"
)

var indexCmd = &cobra.Command{
	Use:   "index",
	Short: "Index places from Nominatim PostgreSQL into Elasticsearch",
	RunE: func(cmd *cobra.Command, args []string) error {
		db := cfg.Database
		dsn := fmt.Sprintf("host=%s port=%d dbname=%s user=%s password=%s sslmode=disable",
			db.Host, db.Port, db.Name, db.User, db.Password)

		pgDB, err := sql.Open("postgres", dsn)
		if err != nil {
			return err
		}
		defer pgDB.Close()
		
		if err := pgDB.Ping(); err != nil {
			return fmt.Errorf("cannot reach PostgreSQL: %w", err)
		}
		fmt.Println("Connected to PostgreSQL.")

		es, err := elasticsearch.NewClient(elasticsearch.Config{
			Addresses: []string{cfg.Elasticsearch.Host},
		})
		if err != nil {
			return err
		}
		fmt.Println("Connected to Elasticsearch.")

		placeDB := nominatim.NewPostgresPlaceDB(pgDB)
		rankConfig := nominatim.DefaultRankConfig()
		countryConfig := nominatim.DefaultCountryConfig()
		extractor := nominatim.NewPlaceExtractor(placeDB, rankConfig, countryConfig)

		indexManager := esinfra.NewIndexManager(es, cfg.Elasticsearch.Index)
		
		bulkConfig := esinfra.DefaultBulkIndexerConfig()
		indexer, err := esinfra.NewBulkDocumentIndexer(es, cfg.Elasticsearch.Index, bulkConfig)
		if err != nil {
			return err
		}

		service := application.NewIndexingService(extractor, indexManager, indexer)

		fmt.Println("Index created.")
		fmt.Println("Extracting and indexing...")
		
		stats, err := service.IndexPlaces(context.Background())
		if err != nil {
			return err
		}

		fmt.Printf("Indexed: %d, Failed: %d\n", stats.Indexed, stats.Failed)
		fmt.Println("Done.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(indexCmd)
}
