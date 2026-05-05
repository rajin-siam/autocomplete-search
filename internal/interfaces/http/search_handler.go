package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"place-search/internal/application"
	"place-search/internal/domain"
)

type SearchService interface {
	Search(ctx context.Context, input application.SearchInput) ([]domain.Place, error)
}

type SearchHandler struct {
	service SearchService
}

func NewSearchHandler(service SearchService) *SearchHandler {
	return &SearchHandler{service: service}
}

func (h *SearchHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	if q == "" {
		writeJSON(w, http.StatusBadRequest, errorJSON("missing query parameter 'q'"))
		return
	}

	places, err := h.service.Search(context.Background(), application.SearchInput{Query: q})
	if err != nil {
		if errors.Is(err, application.ErrQueryTooShort) {
			writeJSON(w, http.StatusBadRequest, errorJSON(err.Error()))
			return
		}
		writeJSON(w, http.StatusInternalServerError, errorJSON("search failed"))
		return
	}

	writeJSON(w, http.StatusOK, formatGeoJSON(places))
}

func errorJSON(msg string) []byte {
	out, _ := json.Marshal(map[string]string{"message": msg})
	return out
}
