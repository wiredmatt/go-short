package storage

import "github.com/wiredmatt/go_short/internal/model"

type Store interface {
	Save(mapping model.URLMapping) error
	Get(code string) (*string, error)
	IncrementClickCount(code string) error
	ListByUser(userID string) ([]model.URLMapping, error)
	Delete(code string) error
	Close()
}
