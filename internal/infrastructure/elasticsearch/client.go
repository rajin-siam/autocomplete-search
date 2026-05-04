package elasticsearch

import (
	"github.com/elastic/go-elasticsearch/v8"
)

func NewClient(host string) (*elasticsearch.Client, error) {
	return elasticsearch.NewClient(elasticsearch.Config{
		Addresses: []string{host},
	})
}
