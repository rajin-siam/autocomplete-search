package http

import (
	"fmt"
	"net/http"
)

func NewServer(port int, searchHandler *SearchHandler) *http.Server {
	mux := http.NewServeMux()
	mux.Handle("/api", searchHandler)
	mux.Handle("/api/", searchHandler)

	return &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}
}
