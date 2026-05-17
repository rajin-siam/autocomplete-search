package hstore

import "strings"

type HStore map[string]string

// Parse parses PostgreSQL HStore format into a map
func Parse(raw string) HStore {
	result := HStore{}
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return result
	}
	var i int
	for i < len(raw) {
		if raw[i] != '"' {
			break
		}
		i++
		keyStart := i
		for i < len(raw) && raw[i] != '"' {
			i++
		}
		key := raw[keyStart:i]
		i++
		i += 2
		if i >= len(raw) {
			break
		}
		if raw[i] == '"' {
			i++
			valStart := i
			for i < len(raw) && raw[i] != '"' {
				if raw[i] == '\\' {
					i++
				}
				i++
			}
			result[key] = raw[valStart:i]
			i++
		} else {
			i += 4
		}
		for i < len(raw) && (raw[i] == ',' || raw[i] == ' ') {
			i++
		}
	}
	return result
}

// GetName extracts name with priority: name:en → name → name:bn → any value
// This priority is intentional for the application's requirements
func GetName(h HStore) string {
	if v, ok := h["name:en"]; ok && v != "" {
		return v
	}
	if v, ok := h["name"]; ok && v != "" {
		return v
	}
	if v, ok := h["name:bn"]; ok && v != "" {
		return v
	}
	for _, v := range h {
		if v != "" {
			return v
		}
	}
	return ""
}
