package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockShortenerService struct {
	mock.Mock
}

func (m *MockShortenerService) GetBaseURL() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockShortenerService) Shorten(userID, originalURL string) (string, error) {
	args := m.Called(userID, originalURL)
	return args.String(0), args.Error(1)
}

func (m *MockShortenerService) Resolve(code string) (string, error) {
	args := m.Called(code)
	return args.String(0), args.Error(1)
}

func TestRouter_ShortenEndpoint(t *testing.T) {
	mockService := &MockShortenerService{}
	baseURL := "https://short.url"

	// Setup mock expectations
	mockService.On("Shorten", "user123", "https://example.com/very/long/url").Return("abc123", nil)
	mockService.On("GetBaseURL").Return(baseURL)

	router := NewRouter(mockService)

	// Create request body as plain JSON
	body := map[string]string{
		"userId": "user123",
		"url":    "https://example.com/very/long/url",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/shorten", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Call router
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var response struct {
		ShortURL string `json:"short_url"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, baseURL+"/abc123", response.ShortURL)

	mockService.AssertExpectations(t)
}

func TestRouter_ResolveEndpoint(t *testing.T) {
	mockService := &MockShortenerService{}
	expectedURL := "https://example.com/very/long/url"

	// Setup mock expectations
	mockService.On("Resolve", "abc123").Return(expectedURL, nil)

	router := NewRouter(mockService)

	// Create request
	req := httptest.NewRequest("GET", "/abc123", nil)
	w := httptest.NewRecorder()

	// Call router
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusFound, w.Code)
	assert.Equal(t, expectedURL, w.Header().Get("Location"))

	mockService.AssertExpectations(t)
}

func TestRouter_NotFound(t *testing.T) {
	mockService := &MockShortenerService{}
	router := NewRouter(mockService)

	// Test non-existent endpoint
	req := httptest.NewRequest("GET", "/nonexistent/endpoint", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Should return 404
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestRouter_MethodNotAllowed(t *testing.T) {
	mockService := &MockShortenerService{}
	router := NewRouter(mockService)

	// Setup mock to return error for "shorten" as a code
	mockService.On("Resolve", "shorten").Return("", assert.AnError)

	// Test wrong method for shorten endpoint
	req := httptest.NewRequest("GET", "/shorten", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Should return 404
	assert.Equal(t, http.StatusNotFound, w.Code)

	mockService.AssertExpectations(t)
}

func TestRouter_ResolveNotFound(t *testing.T) {
	mockService := &MockShortenerService{}

	// Setup mock to return error
	mockService.On("Resolve", "nonexistent").Return("", assert.AnError)

	router := NewRouter(mockService)

	// Create request
	req := httptest.NewRequest("GET", "/nonexistent", nil)
	w := httptest.NewRecorder()

	// Call router
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusNotFound, w.Code)

	mockService.AssertExpectations(t)
}

func TestRouter_ShortenInvalidJSON(t *testing.T) {
	mockService := &MockShortenerService{}
	router := NewRouter(mockService)

	// Create request with invalid JSON
	req := httptest.NewRequest("POST", "/shorten", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Call router
	router.ServeHTTP(w, req)

	// Should return 400 with Huma error body
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRouter_ShortenServiceError(t *testing.T) {
	mockService := &MockShortenerService{}

	// Setup mock to return error
	mockService.On("Shorten", "user123", "https://example.com/very/long/url").Return("", assert.AnError)
	mockService.On("GetBaseURL").Return("https://short.url")

	router := NewRouter(mockService)

	body := map[string]string{
		"userId": "user123",
		"url":    "https://example.com/very/long/url",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/shorten", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Call router
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
