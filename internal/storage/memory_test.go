package storage

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/wiredmatt/go_short/internal/model"
)

func TestNewMemoryStore(t *testing.T) {
	store := NewMemoryStore()

	assert.NotNil(t, store)
	assert.NotNil(t, store.data)
	assert.Empty(t, store.data)
}

func TestMemoryStore_Save(t *testing.T) {
	store := NewMemoryStore()

	mapping := model.URLMapping{
		Code:      "abc123",
		Original:  "https://example.com/very/long/url",
		UserID:    "user123",
		CreatedAt: time.Now(),
	}

	err := store.Save(mapping)

	assert.NoError(t, err)
	assert.Len(t, store.data, 1)
	assert.Equal(t, mapping, store.data["abc123"])
}

func TestMemoryStore_Save_Overwrite(t *testing.T) {
	store := NewMemoryStore()

	originalMapping := model.URLMapping{
		Code:      "abc123",
		Original:  "https://example.com/old/url",
		UserID:    "user123",
		CreatedAt: time.Now(),
	}

	newMapping := model.URLMapping{
		Code:      "abc123",
		Original:  "https://example.com/new/url",
		UserID:    "user456",
		CreatedAt: time.Now(),
	}

	// Save original mapping
	err := store.Save(originalMapping)
	assert.NoError(t, err)

	// Overwrite with new mapping
	err = store.Save(newMapping)
	assert.NoError(t, err)

	assert.Len(t, store.data, 1)
	assert.Equal(t, newMapping, store.data["abc123"])
}

func TestMemoryStore_Get_Success(t *testing.T) {
	store := NewMemoryStore()

	expectedURL := "https://example.com/very/long/url"
	mapping := model.URLMapping{
		Code:      "abc123",
		Original:  expectedURL,
		UserID:    "user123",
		CreatedAt: time.Now(),
	}

	store.Save(mapping)

	url, err := store.Get("abc123")

	assert.NoError(t, err)
	assert.NotNil(t, url)
	assert.Equal(t, expectedURL, *url)
}

func TestMemoryStore_Get_NotFound(t *testing.T) {
	store := NewMemoryStore()

	url, err := store.Get("nonexistent")

	assert.Error(t, err)
	assert.Nil(t, url)
	assert.Equal(t, "code not found", err.Error())
}

func TestMemoryStore_IncrementClickCount_Success(t *testing.T) {
	store := NewMemoryStore()

	mapping := model.URLMapping{
		Code:      "abc123",
		Original:  "https://example.com/very/long/url",
		UserID:    "user123",
		CreatedAt: time.Now(),
		Clicks:    5,
	}

	store.Save(mapping)

	err := store.IncrementClickCount("abc123")

	assert.NoError(t, err)
	assert.Equal(t, 6, store.data["abc123"].Clicks)
}

func TestMemoryStore_IncrementClickCount_NotFound(t *testing.T) {
	store := NewMemoryStore()

	err := store.IncrementClickCount("nonexistent")

	assert.Error(t, err)
	assert.Equal(t, "code not found", err.Error())
}

func TestMemoryStore_ListByUser_Success(t *testing.T) {
	store := NewMemoryStore()

	user1Mapping1 := model.URLMapping{
		Code:      "abc123",
		Original:  "https://example.com/url1",
		UserID:    "user1",
		CreatedAt: time.Now(),
	}

	user1Mapping2 := model.URLMapping{
		Code:      "def456",
		Original:  "https://example.com/url2",
		UserID:    "user1",
		CreatedAt: time.Now(),
	}

	user2Mapping := model.URLMapping{
		Code:      "ghi789",
		Original:  "https://example.com/url3",
		UserID:    "user2",
		CreatedAt: time.Now(),
	}

	store.Save(user1Mapping1)
	store.Save(user1Mapping2)
	store.Save(user2Mapping)

	mappings, err := store.ListByUser("user1")

	assert.NoError(t, err)
	assert.Len(t, mappings, 2)

	// Check that both mappings belong to user1
	for _, mapping := range mappings {
		assert.Equal(t, "user1", mapping.UserID)
	}
}

func TestMemoryStore_ListByUser_Empty(t *testing.T) {
	store := NewMemoryStore()

	mappings, err := store.ListByUser("nonexistent")

	assert.NoError(t, err)
	assert.Empty(t, mappings)
}

func TestMemoryStore_Delete_Success(t *testing.T) {
	store := NewMemoryStore()

	mapping := model.URLMapping{
		Code:      "abc123",
		Original:  "https://example.com/very/long/url",
		UserID:    "user123",
		CreatedAt: time.Now(),
	}

	store.Save(mapping)
	assert.Len(t, store.data, 1)

	err := store.Delete("abc123")

	assert.NoError(t, err)
	assert.Empty(t, store.data)
}

func TestMemoryStore_Delete_NotFound(t *testing.T) {
	store := NewMemoryStore()

	err := store.Delete("nonexistent")

	assert.Error(t, err)
	assert.Equal(t, "code not found", err.Error())
}

func TestMemoryStore_ConcurrentAccess(t *testing.T) {
	store := NewMemoryStore()

	// Test concurrent saves
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(id int) {
			mapping := model.URLMapping{
				Code:      fmt.Sprintf("code%d", id),
				Original:  fmt.Sprintf("https://example.com/url%d", id),
				UserID:    fmt.Sprintf("user%d", id),
				CreatedAt: time.Now(),
			}
			store.Save(mapping)
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	assert.Len(t, store.data, 10)
}

func TestMemoryStore_ConcurrentReadWrite(t *testing.T) {
	store := NewMemoryStore()

	// Add initial data
	mapping := model.URLMapping{
		Code:      "abc123",
		Original:  "https://example.com/very/long/url",
		UserID:    "user123",
		CreatedAt: time.Now(),
	}
	store.Save(mapping)

	// Test concurrent read and write operations
	done := make(chan bool, 20)

	// Start 10 readers
	for i := 0; i < 10; i++ {
		go func() {
			url, err := store.Get("abc123")
			assert.NoError(t, err)
			assert.NotNil(t, url)
			done <- true
		}()
	}

	// Start 10 writers (incrementing click count)
	for i := 0; i < 10; i++ {
		go func() {
			err := store.IncrementClickCount("abc123")
			assert.NoError(t, err)
			done <- true
		}()
	}

	// Wait for all operations to complete
	for i := 0; i < 20; i++ {
		<-done
	}

	// Verify the final state
	assert.Equal(t, 10, store.data["abc123"].Clicks)
}
