package storage

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/wiredmatt/go-short/internal/config"
	"github.com/wiredmatt/go-short/internal/model"
)

func TestPostgresStore(t *testing.T) {
	cfg, err := config.LoadForTest()
	assert.NoError(t, err)

	ctx := context.Background()
	err = ResetPostgresStore(cfg.Database.ConnectionString)
	if err != nil {
		panic(err)
	}
	assert.NoError(t, err)

	store, err := NewPostgresStore(ctx, cfg.Database.ConnectionString)
	if err != nil {
		panic(err)
	}
	assert.NoError(t, err)

	cleanup := func() {
		store.Close()
		err := ResetPostgresStore(cfg.Database.ConnectionString)

		assert.NoError(t, err)
	}

	defer cleanup()

	t.Run("Save and Get", func(t *testing.T) {
		mapping := model.URLMapping{
			Code:      "test123",
			Original:  "https://example.com",
			UserID:    "user1",
			CreatedAt: time.Now(),
			Clicks:    0,
		}

		// Save the mapping
		err := store.Save(mapping)
		assert.NoError(t, err)

		// Get the mapping
		original, err := store.Get("test123")
		assert.NoError(t, err)
		assert.Equal(t, "https://example.com", *original)

		// Test non-existent code
		original, err = store.Get("nonexistent")
		assert.NoError(t, err)
		assert.Nil(t, original)
	})

	t.Run("IncrementClickCount", func(t *testing.T) {
		mapping := model.URLMapping{
			Code:      "clicktest",
			Original:  "https://clicktest.com",
			UserID:    "user1",
			CreatedAt: time.Now(),
			Clicks:    0,
		}

		// Save the mapping
		err := store.Save(mapping)
		assert.NoError(t, err)

		// Increment click count
		err = store.IncrementClickCount("clicktest")
		assert.NoError(t, err)

		// Verify click count was incremented by checking the mapping
		mappings, err := store.ListByUser("user1")
		assert.NoError(t, err)

		var foundMapping *model.URLMapping
		for _, m := range mappings {
			if m.Code == "clicktest" {
				foundMapping = &m
				break
			}
		}
		assert.NotNil(t, foundMapping)
		assert.Equal(t, 1, foundMapping.Clicks)
	})

	t.Run("ListByUser", func(t *testing.T) {
		// Create multiple mappings for the same user
		mappings := []model.URLMapping{
			{
				Code:      "user1_1",
				Original:  "https://user1_1.com",
				UserID:    "user1",
				CreatedAt: time.Now(),
				Clicks:    0,
			},
			{
				Code:      "user1_2",
				Original:  "https://user1_2.com",
				UserID:    "user1",
				CreatedAt: time.Now(),
				Clicks:    0,
			},
		}

		for _, mapping := range mappings {
			err := store.Save(mapping)
			assert.NoError(t, err)
		}

		// List mappings for user1
		userMappings, err := store.ListByUser("user1")
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(userMappings), 2)

		// Verify we can find our test mappings
		codes := make(map[string]bool)
		for _, m := range userMappings {
			codes[m.Code] = true
		}
		assert.True(t, codes["user1_1"])
		assert.True(t, codes["user1_2"])
	})

	t.Run("Delete", func(t *testing.T) {
		mapping := model.URLMapping{
			Code:      "deletetest",
			Original:  "https://deletetest.com",
			UserID:    "user1",
			CreatedAt: time.Now(),
			Clicks:    0,
		}

		// Save the mapping
		err := store.Save(mapping)
		assert.NoError(t, err)

		// Verify it exists
		original, err := store.Get("deletetest")
		assert.NoError(t, err)
		assert.Equal(t, "https://deletetest.com", *original)

		// Delete the mapping
		err = store.Delete("deletetest")
		assert.NoError(t, err)

		// Verify it's gone
		original, err = store.Get("deletetest")
		assert.NoError(t, err)
		assert.Nil(t, original)
	})

	t.Run("Expired URLs", func(t *testing.T) {
		expiresAt := time.Now().Add(-1 * time.Hour) // Expired 1 hour ago
		mapping := model.URLMapping{
			Code:      "expired",
			Original:  "https://expired.com",
			UserID:    "user1",
			CreatedAt: time.Now(),
			ExpiresAt: &expiresAt,
			Clicks:    0,
		}

		// Save the mapping
		err := store.Save(mapping)
		assert.NoError(t, err)

		// Try to get the expired URL
		original, err := store.Get("expired")
		assert.NoError(t, err)
		assert.Nil(t, original) // Should return nil for expired URLs
	})

	t.Run("CleanupExpired", func(t *testing.T) {
		expiresAt := time.Now().Add(-1 * time.Hour) // Expired 1 hour ago
		mapping := model.URLMapping{
			Code:      "cleanuptest",
			Original:  "https://cleanuptest.com",
			UserID:    "user1",
			CreatedAt: time.Now(),
			ExpiresAt: &expiresAt,
			Clicks:    0,
		}

		// Save the mapping
		err := store.Save(mapping)
		assert.NoError(t, err)

		// Run cleanup
		err = store.CleanupExpired()
		assert.NoError(t, err)

		// Verify the expired mapping was removed
		original, err := store.Get("cleanuptest")
		assert.NoError(t, err)
		assert.Nil(t, original)
	})
}
