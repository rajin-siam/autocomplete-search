package domain

import "errors"

var ErrQueryTooShort = errors.New("query too short")

type SearchQuery struct {
	Query    string
	MinChars int
}

func NewSearchQuery(query string, minChars int) (SearchQuery, error) {
	if len([]rune(query)) < minChars {
		return SearchQuery{}, ErrQueryTooShort
	}
	return SearchQuery{Query: query, MinChars: minChars}, nil
}
