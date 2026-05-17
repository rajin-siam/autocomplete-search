package application

import (
	"context"
	"log"
)

type IndexingService struct {
	extractor PlaceExtractor
	manager   IndexManager
	indexer   DocumentIndexer
}

func NewIndexingService(extractor PlaceExtractor, manager IndexManager, indexer DocumentIndexer) *IndexingService {
	return &IndexingService{
		extractor: extractor,
		manager:   manager,
		indexer:   indexer,
	}
}

func (s *IndexingService) IndexPlaces(ctx context.Context) (IndexStats, error) {
	if err := s.manager.CreateIndex(); err != nil {
		return IndexStats{}, err
	}

	docChan, err := s.extractor.ExtractPlaces()
	if err != nil {
		return IndexStats{}, err
	}

	for doc := range docChan {
		select {
		case <-ctx.Done():
			return s.indexer.Stats(), ctx.Err()
		default:
			if err := s.indexer.IndexDocument(doc); err != nil {
				log.Println("Index document error:", err)
			}
		}
	}

	if err := s.indexer.Close(); err != nil {
		return IndexStats{}, err
	}

	return s.indexer.Stats(), nil
}
