package elasticsearch

import (
	"bytes"
	"encoding/json"
	"place-search/internal/domain"
)

func buildQuery(q domain.SearchQuery, fuzziness string, maxResults int) ([]byte, error) {
	query := map[string]any{
		"size": maxResults,
		"query": map[string]any{
			"multi_match": map[string]any{
				"query":     q.Query,
				"type":      "bool_prefix",
				"fuzziness": fuzziness,
				"fields": []string{
					"name",
					"name._2gram",
					"name._3gram",
					"name._index_prefix",
				},
			},
		},
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
