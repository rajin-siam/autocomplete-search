package elasticsearch

type BulkIndexerConfig struct {
	NumWorkers int
	FlushBytes int
}

func DefaultBulkIndexerConfig() *BulkIndexerConfig {
	return &BulkIndexerConfig{
		NumWorkers: 4,
		FlushBytes: 5_000_000,
	}
}
