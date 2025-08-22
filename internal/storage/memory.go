package storage

import (
	"errors"
	"sync"

	"github.com/wiredmatt/go_short/internal/model"
)

type MemoryStore struct {
	data map[string]model.URLMapping
	mu   sync.RWMutex
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		data: make(map[string]model.URLMapping),
	}
}

func (m *MemoryStore) Save(mapping model.URLMapping) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[mapping.Code] = mapping
	return nil
}

func (m *MemoryStore) Get(code string) (*string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	mapping, exists := m.data[code]
	if !exists {
		return nil, errors.New("code not found")
	}
	return &mapping.Original, nil
}

func (m *MemoryStore) IncrementClickCount(code string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	mapping, exists := m.data[code]
	if !exists {
		return errors.New("code not found")
	}
	mapping.Clicks++
	m.data[code] = mapping
	return nil
}

func (m *MemoryStore) ListByUser(userID string) ([]model.URLMapping, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var results []model.URLMapping
	for _, mapping := range m.data {
		if mapping.UserID == userID {
			results = append(results, mapping)
		}
	}
	return results, nil
}

func (m *MemoryStore) Delete(code string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.data[code]; !exists {
		return errors.New("code not found")
	}
	delete(m.data, code)
	return nil
}

func (m *MemoryStore) Close() {}
