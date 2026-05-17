package application

type IndexManager interface {
	CreateIndex() error
	DeleteIndex() error
}

type DocumentIndexer interface {
	IndexDocument(doc interface{}) error
	Close() error
	Stats() IndexStats
}

type IndexStats struct {
	Indexed int64
	Failed  int64
}
