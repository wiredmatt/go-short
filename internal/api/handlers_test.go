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

// MockShortenerService is a mock implementation of the shortener.Shortener
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

func TestShortenURL_Success(t *testing.T) {
	mockService := &MockShortenerService{}
	baseURL := "https://short.url"

	// Setup mock expectations
	mockService.On("Shorten", "user123", "https://example.com/very/long/url").Return("abc123", nil)
	mockService.On("GetBaseURL").Return(baseURL)

	// Create request
	requestBody := ShortenRequest{
		UserID: "user123",
		URL:    "https://example.com/very/long/url",
	}
	jsonBody, _ := json.Marshal(requestBody)

	req := httptest.NewRequest("POST", "/shorten", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Call handler
	handler := ShortenURL(mockService)
	handler.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var response ShortenResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, baseURL+"/abc123", response.ShortURL)

	mockService.AssertExpectations(t)
}

func TestShortenURL_InvalidJSON(t *testing.T) {
	mockService := &MockShortenerService{}

	// Create request with invalid JSON
	req := httptest.NewRequest("POST", "/shorten", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Call handler
	handler := ShortenURL(mockService)
	handler.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "invalid request")
}

func TestShortenURL_ServiceError(t *testing.T) {
	mockService := &MockShortenerService{}

	// Setup mock to return error
	mockService.On("Shorten", "user123", "https://example.com/very/long/url").Return("", assert.AnError)

	// Create request
	requestBody := ShortenRequest{
		UserID: "user123",
		URL:    "https://example.com/very/long/url",
	}
	jsonBody, _ := json.Marshal(requestBody)

	req := httptest.NewRequest("POST", "/shorten", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Call handler
	handler := ShortenURL(mockService)
	handler.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), assert.AnError.Error())

	mockService.AssertExpectations(t)
}

func TestShortenURL_EmptyBody(t *testing.T) {
	mockService := &MockShortenerService{}

	// Create request with empty body
	req := httptest.NewRequest("POST", "/shorten", nil)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Call handler
	handler := ShortenURL(mockService)
	handler.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "invalid request")
}

func TestResolveURL_Success(t *testing.T) {
	mockService := &MockShortenerService{}
	expectedURL := "https://example.com/very/long/url"

	// Setup mock expectations
	mockService.On("Resolve", "abc123").Return(expectedURL, nil)

	// Create router and request
	router := NewRouter(mockService)
	req := httptest.NewRequest("GET", "/abc123", nil)
	w := httptest.NewRecorder()

	// Call router
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusFound, w.Code)
	assert.Equal(t, expectedURL, w.Header().Get("Location"))

	mockService.AssertExpectations(t)
}

func TestResolveURL_NotFound(t *testing.T) {
	mockService := &MockShortenerService{}

	// Setup mock to return error
	mockService.On("Resolve", "nonexistent").Return("", assert.AnError)

	// Create router and request
	router := NewRouter(mockService)
	req := httptest.NewRequest("GET", "/nonexistent", nil)
	w := httptest.NewRecorder()

	// Call router
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusNotFound, w.Code)

	mockService.AssertExpectations(t)
}

func TestResolveURL_EmptyCode(t *testing.T) {
	mockService := &MockShortenerService{}

	// Create router and request with empty code
	router := NewRouter(mockService)
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	// Call router
	router.ServeHTTP(w, req)

	// Assertions - should return 404 since "/" doesn't match "/{code}" pattern
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestShortenRequest_Validation(t *testing.T) {
	tests := []struct {
		name        string
		requestBody ShortenRequest
		expectError bool
	}{
		{
			name: "valid request",
			requestBody: ShortenRequest{
				UserID: "user123",
				URL:    "https://example.com/very/long/url",
			},
			expectError: false,
		},
		{
			name: "empty user ID",
			requestBody: ShortenRequest{
				UserID: "",
				URL:    "https://example.com/very/long/url",
			},
			expectError: false, // Currently no validation, but could be added
		},
		{
			name: "empty URL",
			requestBody: ShortenRequest{
				UserID: "user123",
				URL:    "",
			},
			expectError: false, // Currently no validation, but could be added
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockShortenerService{}
			baseURL := "https://short.url"

			if !tt.expectError {
				mockService.On("Shorten", tt.requestBody.UserID, tt.requestBody.URL).Return("abc123", nil)
				mockService.On("GetBaseURL").Return(baseURL)
			}

			jsonBody, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest("POST", "/shorten", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler := ShortenURL(mockService)
			handler.ServeHTTP(w, req)

			if tt.expectError {
				assert.Equal(t, http.StatusBadRequest, w.Code)
			} else {
				assert.Equal(t, http.StatusOK, w.Code)
			}
		})
	}
}
