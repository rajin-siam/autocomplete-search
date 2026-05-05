package elasticsearch

import (
	"bytes"
	"encoding/json"
)

func buildQuery(query string, fuzziness string, maxResults int) ([]byte, error) {
	body := map[string]any{
		"size": maxResults,
		"query": map[string]any{
			"bool": map[string]any{
				"should": []any{
					map[string]any{
						"multi_match": map[string]any{
							"query":  query,
							"type":   "bool_prefix",
							"fields": []string{"name", "name._2gram", "name._3gram", "name._index_prefix"},
						},
					},
					map[string]any{
						"match": map[string]any{
							"name": map[string]any{
								"query":     query,
								"fuzziness": fuzziness,
							},
						},
					},
				},
			},
		},
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(body); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
