package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/wiredmatt/go-backend-template/internal/shortener"
)

type shortenRequest struct {
	UserID string `json:"userId"`
	URL    string `json:"url"`
}

type shortenResponse struct {
	ShortURL string `json:"short_url"`
}

func ShortenURL(service *shortener.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req shortenRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}

		code, err := service.Shorten(req.UserID, req.URL)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		resp := shortenResponse{ShortURL: service.GetBaseURL() + "/" + code}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}

func ResolveURL(service *shortener.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		code := chi.URLParam(r, "code")
		originalURL, err := service.Resolve(code)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		http.Redirect(w, r, originalURL, http.StatusFound)
	}
}
