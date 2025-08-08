package model

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestURLMapping_Fields(t *testing.T) {
	now := time.Now()
	expiresAt := now.Add(24 * time.Hour)

	mapping := URLMapping{
		Code:      "abc123",
		Original:  "https://example.com/very/long/url",
		UserID:    "user123",
		CreatedAt: now,
		ExpiresAt: &expiresAt,
		Clicks:    5,
	}

	assert.Equal(t, "abc123", mapping.Code)
	assert.Equal(t, "https://example.com/very/long/url", mapping.Original)
	assert.Equal(t, "user123", mapping.UserID)
	assert.Equal(t, now, mapping.CreatedAt)
	assert.Equal(t, &expiresAt, mapping.ExpiresAt)
	assert.Equal(t, 5, mapping.Clicks)
}

func TestURLMapping_ZeroValue(t *testing.T) {
	var mapping URLMapping

	assert.Equal(t, "", mapping.Code)
	assert.Equal(t, "", mapping.Original)
	assert.Equal(t, "", mapping.UserID)
	assert.Equal(t, time.Time{}, mapping.CreatedAt)
	assert.Nil(t, mapping.ExpiresAt)
	assert.Equal(t, 0, mapping.Clicks)
}

func TestURLMapping_WithoutExpiration(t *testing.T) {
	now := time.Now()

	mapping := URLMapping{
		Code:      "abc123",
		Original:  "https://example.com/very/long/url",
		UserID:    "user123",
		CreatedAt: now,
		ExpiresAt: nil,
		Clicks:    0,
	}

	assert.Equal(t, "abc123", mapping.Code)
	assert.Equal(t, "https://example.com/very/long/url", mapping.Original)
	assert.Equal(t, "user123", mapping.UserID)
	assert.Equal(t, now, mapping.CreatedAt)
	assert.Nil(t, mapping.ExpiresAt)
	assert.Equal(t, 0, mapping.Clicks)
}

func TestURLMapping_IsExpired(t *testing.T) {
	now := time.Now()
	pastTime := now.Add(-1 * time.Hour)
	futureTime := now.Add(1 * time.Hour)

	tests := []struct {
		name     string
		mapping  URLMapping
		expected bool
	}{
		{
			name: "expired mapping",
			mapping: URLMapping{
				Code:      "abc123",
				Original:  "https://example.com/url",
				UserID:    "user123",
				CreatedAt: now,
				ExpiresAt: &pastTime,
				Clicks:    0,
			},
			expected: true,
		},
		{
			name: "not expired mapping",
			mapping: URLMapping{
				Code:      "def456",
				Original:  "https://example.com/url",
				UserID:    "user123",
				CreatedAt: now,
				ExpiresAt: &futureTime,
				Clicks:    0,
			},
			expected: false,
		},
		{
			name: "no expiration set",
			mapping: URLMapping{
				Code:      "ghi789",
				Original:  "https://example.com/url",
				UserID:    "user123",
				CreatedAt: now,
				ExpiresAt: nil,
				Clicks:    0,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isExpired := tt.mapping.ExpiresAt != nil && time.Now().After(*tt.mapping.ExpiresAt)
			assert.Equal(t, tt.expected, isExpired)
		})
	}
}
