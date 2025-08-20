package api

import (
	"net/http"

	"github.com/wiredmatt/go-backend-template/internal/shortener"
)

func NewRouter(service shortener.Shortener) *http.ServeMux {
	r := http.NewServeMux()
	r.HandleFunc("POST /shorten", ShortenURL(service))
	r.HandleFunc("GET /{code}", ResolveURL(service))
	return r
}
