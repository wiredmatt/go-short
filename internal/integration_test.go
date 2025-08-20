package internal

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wiredmatt/go-backend-template/internal/api"
	"github.com/wiredmatt/go-backend-template/internal/shortener"
	"github.com/wiredmatt/go-backend-template/internal/storage"
)

func TestIntegration_ShortenAndResolve(t *testing.T) {
	// Setup real components
	store := storage.NewMemoryStore()
	baseURL := "https://short.url"
	shortCodeLength := 6
	service := shortener.NewService(store, baseURL, shortCodeLength)
	router := api.NewRouter(service)

	// Test shortening a URL
	reqBody := api.ShortenRequest{
		UserID: "user123",
		URL:    "https://example.com/very/long/url",
	}

	jsonBody, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/shorten", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Verify shorten response
	assert.Equal(t, http.StatusOK, w.Code)

	var res api.ShortenResponse

	err := json.Unmarshal(w.Body.Bytes(), &res)
	assert.NoError(t, err)
	assert.Contains(t, res.ShortURL, baseURL)

	// Extract the code from the short URL
	code := res.ShortURL[len(baseURL)+1:] // +1 for the "/"
	assert.Len(t, code, 6)

	// Test resolving the shortened URL
	resolveReq := httptest.NewRequest("GET", "/"+code, nil)
	resolveW := httptest.NewRecorder()

	router.ServeHTTP(resolveW, resolveReq)

	// Verify redirect response
	assert.Equal(t, http.StatusFound, resolveW.Code)
	assert.Equal(t, "https://example.com/very/long/url", resolveW.Header().Get("Location"))
}

func TestIntegration_ResolveNonExistent(t *testing.T) {
	// Setup real components
	store := storage.NewMemoryStore()
	baseURL := "https://short.url"
	shortCodeLength := 6
	service := shortener.NewService(store, baseURL, shortCodeLength)
	router := api.NewRouter(service)

	// Test resolving a non-existent code
	req := httptest.NewRequest("GET", "/nonexistent", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Should return 404
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestIntegration_MultipleShortens(t *testing.T) {
	// Setup real components
	store := storage.NewMemoryStore()
	baseURL := "https://short.url"
	shortCodeLength := 6
	service := shortener.NewService(store, baseURL, shortCodeLength)
	router := api.NewRouter(service)

	urls := []string{
		"https://example.com/url1",
		"https://example.com/url2",
		"https://example.com/url3",
	}

	codes := make([]string, len(urls))

	// Shorten multiple URLs
	for i, url := range urls {
		reqBody := api.ShortenRequest{
			UserID: "user123",
			URL:    url,
		}

		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/shorten", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var res api.ShortenResponse
		err := json.Unmarshal(w.Body.Bytes(), &res)
		assert.NoError(t, err)

		code := res.ShortURL[len(baseURL)+1:]
		codes[i] = code
	}

	// Verify all codes are unique
	codeMap := make(map[string]bool)
	for _, code := range codes {
		assert.False(t, codeMap[code], "Duplicate code generated: %s", code)
		codeMap[code] = true
	}

	// Verify all codes can be resolved
	for i, code := range codes {
		req := httptest.NewRequest("GET", "/"+code, nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusFound, w.Code)
		assert.Equal(t, urls[i], w.Header().Get("Location"))
	}
}

func TestIntegration_InvalidShortenRequest(t *testing.T) {
	// Setup real components
	store := storage.NewMemoryStore()
	baseURL := "https://short.url"
	shortCodeLength := 6
	service := shortener.NewService(store, baseURL, shortCodeLength)
	router := api.NewRouter(service)

	// Test with invalid JSON
	req := httptest.NewRequest("POST", "/shorten", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "invalid request")
}

func TestIntegration_EmptyShortenRequest(t *testing.T) {
	// Setup real components
	store := storage.NewMemoryStore()
	baseURL := "https://short.url"
	shortCodeLength := 6
	service := shortener.NewService(store, baseURL, shortCodeLength)
	router := api.NewRouter(service)

	// Test with empty body
	req := httptest.NewRequest("POST", "/shorten", nil)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "invalid request")
}
