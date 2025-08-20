package shortener

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/wiredmatt/go-backend-template/internal/model"
)

// MockStore is a mock implementation of the storage.Store interface
type MockStore struct {
	mock.Mock
}

func (m *MockStore) Save(mapping model.URLMapping) error {
	args := m.Called(mapping)
	return args.Error(0)
}

func (m *MockStore) Get(code string) (*string, error) {
	args := m.Called(code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*string), args.Error(1)
}

func (m *MockStore) IncrementClickCount(code string) error {
	args := m.Called(code)
	return args.Error(0)
}

func (m *MockStore) ListByUser(userID string) ([]model.URLMapping, error) {
	args := m.Called(userID)
	return args.Get(0).([]model.URLMapping), args.Error(1)
}

func (m *MockStore) Delete(code string) error {
	args := m.Called(code)
	return args.Error(0)
}

func TestNewService(t *testing.T) {
	mockStore := &MockStore{}
	baseURL := "https://short.url"
	shortCodeLength := 6

	service := NewService(mockStore, baseURL, shortCodeLength)

	assert.NotNil(t, service)
	assert.Equal(t, mockStore, service.store)
	assert.Equal(t, baseURL, service.baseURL)
	assert.Equal(t, shortCodeLength, service.shortCodeLength)
}

func TestGetBaseURL(t *testing.T) {
	mockStore := &MockStore{}
	baseURL := "https://short.url"
	shortCodeLength := 6

	service := NewService(mockStore, baseURL, shortCodeLength)

	assert.Equal(t, baseURL, service.GetBaseURL())
}

func TestShorten_Success(t *testing.T) {
	mockStore := &MockStore{}
	baseURL := "https://short.url"
	shortCodeLength := 6
	service := NewService(mockStore, baseURL, shortCodeLength)

	userID := "user123"
	originalURL := "https://example.com/very/long/url"

	mockStore.On("Save", mock.AnythingOfType("model.URLMapping")).Return(nil)

	code, err := service.Shorten(userID, originalURL)

	assert.NoError(t, err)
	assert.NotEmpty(t, code)
	assert.Len(t, code, shortCodeLength) // Should be configured length

	mockStore.AssertExpectations(t)
}

func TestShorten_StoreError(t *testing.T) {
	mockStore := &MockStore{}
	baseURL := "https://short.url"
	shortCodeLength := 6
	service := NewService(mockStore, baseURL, shortCodeLength)

	userID := "user123"
	originalURL := "https://example.com/very/long/url"
	expectedError := errors.New("storage error")

	mockStore.On("Save", mock.AnythingOfType("model.URLMapping")).Return(expectedError)

	code, err := service.Shorten(userID, originalURL)

	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	assert.Empty(t, code)

	mockStore.AssertExpectations(t)
}

func TestResolve_Success(t *testing.T) {
	mockStore := &MockStore{}
	baseURL := "https://short.url"
	shortCodeLength := 6
	service := NewService(mockStore, baseURL, shortCodeLength)

	code := "abc123"
	expectedURL := "https://example.com/very/long/url"

	mockStore.On("Get", code).Return(&expectedURL, nil)

	originalURL, err := service.Resolve(code)

	assert.NoError(t, err)
	assert.Equal(t, expectedURL, originalURL)

	mockStore.AssertExpectations(t)
}

func TestResolve_NotFound(t *testing.T) {
	mockStore := &MockStore{}
	baseURL := "https://short.url"
	shortCodeLength := 6
	service := NewService(mockStore, baseURL, shortCodeLength)

	code := "nonexistent"
	expectedError := errors.New("code not found")

	mockStore.On("Get", code).Return(nil, expectedError)

	originalURL, err := service.Resolve(code)

	assert.Error(t, err)
	assert.Equal(t, "code not found", err.Error())
	assert.Empty(t, originalURL)

	mockStore.AssertExpectations(t)
}

func TestResolve_NilURL(t *testing.T) {
	mockStore := &MockStore{}
	baseURL := "https://short.url"
	shortCodeLength := 6
	service := NewService(mockStore, baseURL, shortCodeLength)

	code := "abc123"

	// Expect the store to return nil URL
	mockStore.On("Get", code).Return(nil, nil)

	originalURL, err := service.Resolve(code)

	assert.Error(t, err)
	assert.Equal(t, "code not found", err.Error())
	assert.Empty(t, originalURL)

	mockStore.AssertExpectations(t)
}

func TestGenerateCode(t *testing.T) {
	length := 6
	code := generateCode(length)

	assert.Len(t, code, length)

	// Test that it generates different codes on subsequent calls
	code2 := generateCode(length)
	assert.NotEqual(t, code, code2)

	// Test that it only contains valid characters
	validChars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	for _, char := range code {
		assert.Contains(t, validChars, string(char))
	}
}

func TestShorten_GeneratesUniqueCodes(t *testing.T) {
	mockStore := &MockStore{}
	baseURL := "https://short.url"
	shortCodeLength := 6
	service := NewService(mockStore, baseURL, shortCodeLength)

	userID := "user123"
	originalURL := "https://example.com/very/long/url"

	mockStore.On("Save", mock.AnythingOfType("model.URLMapping")).Return(nil)

	codes := make(map[string]bool)
	for i := 0; i < 100; i++ {
		code, err := service.Shorten(userID, originalURL)
		assert.NoError(t, err)
		assert.NotEmpty(t, code)

		// Check for uniqueness
		assert.False(t, codes[code], "Generated duplicate code: %s", code)
		codes[code] = true
	}

	mockStore.AssertExpectations(t)
}
