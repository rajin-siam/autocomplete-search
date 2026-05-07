package elasticsearch

import (
	"fmt"
	"strings"

	"github.com/elastic/go-elasticsearch/v8"
)

type IndexManagerImpl struct {
	client *elasticsearch.Client
	index  string
}

func NewIndexManager(client *elasticsearch.Client, index string) *IndexManagerImpl {
	return &IndexManagerImpl{
		client: client,
		index:  index,
	}
}

func (m *IndexManagerImpl) DeleteIndex() error {
	res, err := m.client.Indices.Delete([]string{m.index})
	if res != nil {
		defer res.Body.Close()
	}
	return err
}

func (m *IndexManagerImpl) CreateIndex() error {
	m.DeleteIndex()
	
	mapping := `{
		"mappings": {
			"properties": {
				"name":        {"type": "search_as_you_type"},
				"coordinate":  {"type": "geo_point",  "index": false},
				"osm_id":      {"type": "long",        "index": false},
				"osm_type":    {"type": "keyword",     "index": false},
				"osm_key":     {"type": "keyword",     "index": false},
				"osm_value":   {"type": "keyword",     "index": false},
				"type":        {"type": "keyword",     "index": false},
				"country":     {"type": "keyword",     "index": false},
				"countrycode": {"type": "keyword",     "index": false},
				"state":       {"type": "keyword",     "index": false},
				"county":      {"type": "keyword",     "index": false},
				"city":        {"type": "keyword",     "index": false},
				"district":    {"type": "keyword",     "index": false},
				"locality":    {"type": "keyword",     "index": false},
				"street":      {"type": "keyword",     "index": false},
				"postcode":    {"type": "keyword",     "index": false},
				"extent":      {"type": "object",      "enabled": false}
			}
		}
	}`
	
	res, err := m.client.Indices.Create(
		m.index,
		m.client.Indices.Create.WithBody(strings.NewReader(mapping)),
	)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	
	if res.IsError() {
		return fmt.Errorf("failed to create index: %s", res.String())
	}
	
	return nil
}
