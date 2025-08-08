package api

import (
	"github.com/go-chi/chi/v5"
	"github.com/wiredmatt/go-backend-template/internal/shortener"
)

func NewRouter(service *shortener.Service) *chi.Mux {
	r := chi.NewRouter()
	r.Post("/shorten", ShortenURL(service))
	r.Get("/{code}", ResolveURL(service))
	return r
}
