package factory

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func NewHTTPServer(port string, r *chi.Mux) *http.Server {
	return &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}
}
