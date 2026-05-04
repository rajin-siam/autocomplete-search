package http

import (
	"context"
	"encoding/json"
	"net/http"
	"place-search/internal/domain"
)

type SearchService interface {
	Search(ctx context.Context, q domain.SearchQuery) ([]domain.Place, error)
}

type SearchHandler struct {
	service  SearchService
	minChars int
}

func NewSearchHandler(service SearchService, minChars int) *SearchHandler {
	return &SearchHandler{service: service, minChars: minChars}
}

func (h *SearchHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	if q == "" {
		writeJSON(w, http.StatusBadRequest, errorJSON("missing query parameter 'q'"))
		return
	}

	query, err := domain.NewSearchQuery(q, h.minChars)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorJSON(err.Error()))
		return
	}

	places, err := h.service.Search(context.Background(), query)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorJSON("search failed"))
		return
	}

	writeJSON(w, http.StatusOK, formatGeoJSON(places))
}

func errorJSON(msg string) []byte {
	out, _ := json.Marshal(map[string]string{"message": msg})
	return out
}
