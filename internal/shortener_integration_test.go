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

	cfg, err := config.Load()
	assert.NoError(t, err)

	ctx := context.Background()

	storage.ResetStore(ctx, cfg.Database)
	store, err := storage.NewStore(ctx, cfg.Database)
	assert.NoError(t, err)

	service := shortener.NewService(store, cfg.App.BaseURL, cfg.App.ShortCodeLength)
	router := api.NewRouter(service)

	cleanup := func() {
		store.Close()
	}

	defer cleanup()

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

	t.Run("List Mappings API", func(t *testing.T) {
		reqBody := map[string]string{
			"userId": "user456",
			"url":    "https://example.com/mapping-test",
		}

		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/shorten", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		mappingsReq := httptest.NewRequest("GET", "/mappings?userId=user456", nil)
		mappingsW := httptest.NewRecorder()

		router.ServeHTTP(mappingsW, mappingsReq)

		assert.Equal(t, http.StatusOK, mappingsW.Code)

		var mappingsRes struct {
			Mappings []struct {
				Code      string `json:"code"`
				Original  string `json:"original_url"`
				ShortURL  string `json:"short_url"`
				CreatedAt string `json:"created_at"`
				Clicks    int    `json:"clicks"`
			} `json:"mappings"`
		}

		err := json.Unmarshal(mappingsW.Body.Bytes(), &mappingsRes)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(mappingsRes.Mappings), 1)

		var foundMapping *struct {
			Code      string `json:"code"`
			Original  string `json:"original_url"`
			ShortURL  string `json:"short_url"`
			CreatedAt string `json:"created_at"`
			Clicks    int    `json:"clicks"`
		}

		for i := range mappingsRes.Mappings {
			if mappingsRes.Mappings[i].Original == "https://example.com/mapping-test" {
				foundMapping = &mappingsRes.Mappings[i]
				break
			}
		}

		assert.NotNil(t, foundMapping)
		assert.Equal(t, "https://example.com/mapping-test", foundMapping.Original)
		assert.Equal(t, 0, foundMapping.Clicks, "New mapping should have 0 clicks")
		assert.Contains(t, foundMapping.ShortURL, cfg.App.BaseURL)
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
