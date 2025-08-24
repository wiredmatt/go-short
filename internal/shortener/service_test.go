package shortener

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/wiredmatt/go_short/internal/model"
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

func (m *MockStore) Close() {}

type AsyncMockStore struct {
	mock.Mock
	clickCountCalls chan string
}

func NewAsyncMockStore() *AsyncMockStore {
	return &AsyncMockStore{
		clickCountCalls: make(chan string, 10),
	}
}

func (m *AsyncMockStore) Save(mapping model.URLMapping) error {
	args := m.Called(mapping)
	return args.Error(0)
}

func (m *AsyncMockStore) Get(code string) (*string, error) {
	args := m.Called(code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*string), args.Error(1)
}

func (m *AsyncMockStore) IncrementClickCount(code string) error {
	// Signal that this method was called
	select {
	case m.clickCountCalls <- code:
	default:
	}

	args := m.Called(code)
	return args.Error(0)
}

func (m *AsyncMockStore) ListByUser(userID string) ([]model.URLMapping, error) {
	args := m.Called(userID)
	return args.Get(0).([]model.URLMapping), args.Error(1)
}

func (m *AsyncMockStore) Delete(code string) error {
	args := m.Called(code)
	return args.Error(0)
}

func (m *AsyncMockStore) Close() {}

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
	mockStore := NewAsyncMockStore()
	baseURL := "https://short.url"
	shortCodeLength := 6
	service := NewService(mockStore, baseURL, shortCodeLength)

	code := "abc123"
	expectedURL := "https://example.com/very/long/url"

	mockStore.On("Get", code).Return(&expectedURL, nil)
	mockStore.On("IncrementClickCount", code).Return(nil)

	// Test that Resolve returns immediately
	originalURL, err := service.Resolve(code)

	assert.NoError(t, err)
	assert.Equal(t, expectedURL, originalURL)

	// Wait for the async click counting to happen
	select {
	case calledCode := <-mockStore.clickCountCalls:
		assert.Equal(t, code, calledCode, "Click count should be called with the correct code")
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Click counting should happen asynchronously")
	}

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

func TestResolve_ClickCountFailure(t *testing.T) {
	mockStore := NewAsyncMockStore()
	baseURL := "https://short.url"
	shortCodeLength := 6
	service := NewService(mockStore, baseURL, shortCodeLength)

	code := "abc123"
	expectedURL := "https://example.com/very/long/url"
	expectedError := errors.New("click count error")

	mockStore.On("Get", code).Return(&expectedURL, nil)
	mockStore.On("IncrementClickCount", code).Return(expectedError)

	// Test that Resolve returns immediately even when click counting will fail
	originalURL, err := service.Resolve(code)

	// Should still succeed even if click counting fails
	assert.NoError(t, err)
	assert.Equal(t, expectedURL, originalURL)

	// Wait for the async click counting to happen (even though it will fail)
	select {
	case calledCode := <-mockStore.clickCountCalls:
		assert.Equal(t, code, calledCode, "Click count should be called with the correct code")
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Click counting should happen asynchronously")
	}

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

func TestListMappings_Success(t *testing.T) {
	mockStore := &MockStore{}
	baseURL := "https://short.url"
	shortCodeLength := 6
	service := NewService(mockStore, baseURL, shortCodeLength)

	userID := "user123"
	expectedMappings := []model.URLMapping{
		{
			Code:      "abc123",
			Original:  "https://example.com/url1",
			UserID:    userID,
			CreatedAt: time.Now(),
			Clicks:    5,
		},
		{
			Code:      "def456",
			Original:  "https://example.com/url2",
			UserID:    userID,
			CreatedAt: time.Now(),
			Clicks:    10,
		},
	}

	mockStore.On("ListByUser", userID).Return(expectedMappings, nil)

	mappings, err := service.ListMappings(userID)

	assert.NoError(t, err)
	assert.Equal(t, expectedMappings, mappings)
	assert.Len(t, mappings, 2)

	mockStore.AssertExpectations(t)
}

func TestListMappings_StoreError(t *testing.T) {
	mockStore := &MockStore{}
	baseURL := "https://short.url"
	shortCodeLength := 6
	service := NewService(mockStore, baseURL, shortCodeLength)

	userID := "user123"
	expectedError := errors.New("storage error")

	mockStore.On("ListByUser", userID).Return([]model.URLMapping{}, expectedError)

	mappings, err := service.ListMappings(userID)

	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	assert.Empty(t, mappings)

	mockStore.AssertExpectations(t)
}

func TestResolve_AsyncClickCounting(t *testing.T) {
	mockStore := NewAsyncMockStore()
	baseURL := "https://short.url"
	shortCodeLength := 6
	service := NewService(mockStore, baseURL, shortCodeLength)

	code := "abc123"
	expectedURL := "https://example.com/very/long/url"

	mockStore.On("Get", code).Return(&expectedURL, nil)
	mockStore.On("IncrementClickCount", code).Return(nil)

	// Test that Resolve returns immediately
	originalURL, err := service.Resolve(code)

	assert.NoError(t, err)
	assert.Equal(t, expectedURL, originalURL)

	// Wait for the async click counting to happen
	select {
	case calledCode := <-mockStore.clickCountCalls:
		assert.Equal(t, code, calledCode, "Click count should be called with the correct code")
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Click counting should happen asynchronously")
	}

	mockStore.AssertExpectations(t)
}
