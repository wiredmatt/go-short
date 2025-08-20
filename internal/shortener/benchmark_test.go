package shortener

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/wiredmatt/go-short/internal/model"
	"github.com/wiredmatt/go-short/internal/storage"
)

type BenchmarkStore struct {
	mock.Mock
}

func (m *BenchmarkStore) Save(mapping model.URLMapping) error {
	args := m.Called(mapping)
	return args.Error(0)
}

func (m *BenchmarkStore) Get(code string) (*string, error) {
	args := m.Called(code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*string), args.Error(1)
}

func (m *BenchmarkStore) IncrementClickCount(code string) error {
	args := m.Called(code)
	return args.Error(0)
}

func (m *BenchmarkStore) ListByUser(userID string) ([]model.URLMapping, error) {
	args := m.Called(userID)
	return args.Get(0).([]model.URLMapping), args.Error(1)
}

func (m *BenchmarkStore) Delete(code string) error {
	args := m.Called(code)
	return args.Error(0)
}

func BenchmarkShorten(b *testing.B) {
	mockStore := &BenchmarkStore{}
	baseURL := "https://short.url"
	shortCodeLength := 6
	service := NewService(mockStore, baseURL, shortCodeLength)

	userID := "user123"
	originalURL := "https://example.com/very/long/url"

	mockStore.On("Save", mock.AnythingOfType("model.URLMapping")).Return(nil)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := service.Shorten(userID, originalURL)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkResolve(b *testing.B) {
	mockStore := &BenchmarkStore{}
	baseURL := "https://short.url"
	shortCodeLength := 6
	service := NewService(mockStore, baseURL, shortCodeLength)

	code := "abc123"
	expectedURL := "https://example.com/very/long/url"

	mockStore.On("Get", code).Return(&expectedURL, nil)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := service.Resolve(code)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGenerateCode(b *testing.B) {
	length := 6

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = generateCode(length)
	}
}

func BenchmarkGenerateCode_DifferentLengths(b *testing.B) {
	lengths := []int{4, 6, 8, 10}

	for _, length := range lengths {
		b.Run(fmt.Sprintf("Length_%d", length), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = generateCode(length)
			}
		})
	}
}

func BenchmarkMemoryStore_Save(b *testing.B) {
	store := storage.NewMemoryStore()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		mapping := model.URLMapping{
			Code:      fmt.Sprintf("code%d", i),
			Original:  fmt.Sprintf("https://example.com/url%d", i),
			UserID:    fmt.Sprintf("user%d", i),
			CreatedAt: time.Now(),
		}

		err := store.Save(mapping)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMemoryStore_Get(b *testing.B) {
	store := storage.NewMemoryStore()

	// Pre-populate with data
	for i := 0; i < 1000; i++ {
		mapping := model.URLMapping{
			Code:      fmt.Sprintf("code%d", i),
			Original:  fmt.Sprintf("https://example.com/url%d", i),
			UserID:    fmt.Sprintf("user%d", i),
			CreatedAt: time.Now(),
		}
		store.Save(mapping)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		code := fmt.Sprintf("code%d", i%1000)
		_, err := store.Get(code)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMemoryStore_ConcurrentAccess(b *testing.B) {
	store := storage.NewMemoryStore()

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			mapping := model.URLMapping{
				Code:      fmt.Sprintf("code%d", i),
				Original:  fmt.Sprintf("https://example.com/url%d", i),
				UserID:    fmt.Sprintf("user%d", i),
				CreatedAt: time.Now(),
			}

			err := store.Save(mapping)
			if err != nil {
				b.Fatal(err)
			}

			_, err = store.Get(mapping.Code)
			if err != nil {
				b.Fatal(err)
			}

			i++
		}
	})
}
