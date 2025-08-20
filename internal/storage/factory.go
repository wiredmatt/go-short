package storage

import (
	"context"
	"fmt"

	"github.com/wiredmatt/go-short/internal/config"
)

func NewStore(ctx context.Context, cfg config.DatabaseConfig) (Store, error) {
	switch cfg.Type {
	case "memory":
		return NewMemoryStore(), nil
	case "postgres":
		return NewPostgresStore(ctx, cfg.ConnectionString)
	case "redis":
		return nil, fmt.Errorf("redis storage not yet implemented")
	default:
		return nil, fmt.Errorf("unknown database type: %s", cfg.Type)
	}
}
