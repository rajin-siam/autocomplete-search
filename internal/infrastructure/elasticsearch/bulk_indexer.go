package elasticsearch

import (
	"context"
	"encoding/json"
	"log"
	"place-search/internal/application"
	"strings"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esutil"
)

type BulkDocumentIndexer struct {
	indexer esutil.BulkIndexer
}

func NewBulkDocumentIndexer(client *elasticsearch.Client, index string, config *BulkIndexerConfig) (*BulkDocumentIndexer, error) {
	bi, err := esutil.NewBulkIndexer(esutil.BulkIndexerConfig{
		Index:      index,
		Client:     client,
		NumWorkers: config.NumWorkers,
		FlushBytes: config.FlushBytes,
	})
	if err != nil {
		return nil, err
	}

	return &BulkDocumentIndexer{indexer: bi}, nil
}

func (b *BulkDocumentIndexer) IndexDocument(doc interface{}) error {
	data, err := json.Marshal(doc)
	if err != nil {
		return err
	}

	return b.indexer.Add(context.Background(), esutil.BulkIndexerItem{
		Action: "index",
		Body:   strings.NewReader(string(data)),
		OnFailure: func(ctx context.Context, item esutil.BulkIndexerItem, res esutil.BulkIndexerResponseItem, err error) {
			if err != nil {
				log.Println("Bulk item error:", err)
			}
		},
	})
}

func (b *BulkDocumentIndexer) Close() error {
	return b.indexer.Close(context.Background())
}

func (b *BulkDocumentIndexer) Stats() application.IndexStats {
	stats := b.indexer.Stats()
	return application.IndexStats{
		Indexed: int64(stats.NumIndexed),
		Failed:  int64(stats.NumFailed),
	}
}
