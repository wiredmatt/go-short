package internal

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wiredmatt/go_short/internal/api"
	"github.com/wiredmatt/go_short/internal/config"
	"github.com/wiredmatt/go_short/internal/shortener"
	"github.com/wiredmatt/go_short/internal/storage"
)

func TestShortenerIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	cfg, err := config.LoadForTest()
	assert.NoError(t, err)

	ctx := context.Background()

	store, err := storage.NewPostgresStore(ctx, cfg.Database.ConnectionString)
	assert.NoError(t, err)

	service := shortener.NewService(store, cfg.App.BaseURL, cfg.App.ShortCodeLength)
	router := api.NewRouter(service)

	t.Run("Shorten And Resolve", func(t *testing.T) {
		// Test shortening a URL
		reqBody := map[string]string{
			"userId": "user123",
			"url":    "https://example.com/very/long/url",
		}

		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/shorten", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// Verify shorten response
		assert.Equal(t, http.StatusOK, w.Code)

		var res struct {
			ShortURL string `json:"short_url"`
		}

		err := json.Unmarshal(w.Body.Bytes(), &res)
		assert.NoError(t, err)
		assert.Contains(t, res.ShortURL, cfg.App.BaseURL)

		// Extract the code from the short URL
		code := res.ShortURL[len(cfg.App.BaseURL)+1:] // +1 for the "/"
		assert.Len(t, code, 6)

		// Test resolving the shortened URL
		resolveReq := httptest.NewRequest("GET", "/"+code, nil)
		resolveW := httptest.NewRecorder()

		router.ServeHTTP(resolveW, resolveReq)

		// Verify redirect response
		assert.Equal(t, http.StatusFound, resolveW.Code)
		assert.Equal(t, "https://example.com/very/long/url", resolveW.Header().Get("Location"))
	})

	t.Run("Resolve Non Existent", func(t *testing.T) {
		// Test resolving a non-existent code
		req := httptest.NewRequest("GET", "/nonexistent", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("Multiple Shortens", func(t *testing.T) {
		urls := []string{
			"https://example.com/url1",
			"https://example.com/url2",
			"https://example.com/url3",
		}

		codes := make([]string, len(urls))

		// Shorten multiple URLs
		for i, url := range urls {
			reqBody := map[string]string{
				"userId": "user123",
				"url":    url,
			}

			jsonBody, _ := json.Marshal(reqBody)
			req := httptest.NewRequest("POST", "/shorten", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			var res struct {
				ShortURL string `json:"short_url"`
			}
			err := json.Unmarshal(w.Body.Bytes(), &res)
			assert.NoError(t, err)

			code := res.ShortURL[len(cfg.App.BaseURL)+1:]
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
	})

	t.Run("Invalid Shorten Request", func(t *testing.T) {
		// Test with invalid JSON
		req := httptest.NewRequest("POST", "/shorten", bytes.NewBufferString("invalid json"))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Empty Shorten Request", func(t *testing.T) {
		// Test with empty body
		req := httptest.NewRequest("POST", "/shorten", nil)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}
