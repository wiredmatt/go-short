package storage

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/wiredmatt/go-backend-template/internal/model"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

type PostgresStore struct {
	pool *pgxpool.Pool
}

// NewPostgresStore creates a new PostgreSQL storage instance
func NewPostgresStore(ctx context.Context, connString string) (*PostgresStore, error) {
	// Apply migrations before initializing the pool
	if err := runMigrations(connString); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	pool, err := pgxpool.New(ctx, connString)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	store := &PostgresStore{pool: pool}

	return store, nil
}

// runMigrations applies database migrations using Goose
func runMigrations(connString string) error {
	cfg, err := pgx.ParseConfig(connString)
	if err != nil {
		return fmt.Errorf("failed to parse connection string: %w", err)
	}
	db := stdlib.OpenDB(*cfg)
	defer db.Close()

	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}
	goose.SetBaseFS(migrationsFS)
	if err := goose.Up(db, "migrations"); err != nil {
		return err
	}
	return nil
}

// Save stores a new URL mapping
func (p *PostgresStore) Save(mapping model.URLMapping) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
		INSERT INTO url_mappings (code, original_url, user_id, created_at, expires_at, clicks)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := p.pool.Exec(ctx, query,
		mapping.Code,
		mapping.Original,
		mapping.UserID,
		mapping.CreatedAt,
		mapping.ExpiresAt,
		mapping.Clicks,
	)

	return err
}

// Get retrieves the original URL for a given code
func (p *PostgresStore) Get(code string) (*string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
		SELECT original_url FROM url_mappings 
		WHERE code = $1 AND (expires_at IS NULL OR expires_at > NOW())
	`

	var originalURL string
	err := p.pool.QueryRow(ctx, query, code).Scan(&originalURL)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &originalURL, nil
}

// IncrementClickCount increases the click count for a given code
func (p *PostgresStore) IncrementClickCount(code string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
		UPDATE url_mappings 
		SET clicks = clicks + 1 
		WHERE code = $1
	`

	result, err := p.pool.Exec(ctx, query, code)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("no URL mapping found for code: %s", code)
	}

	return nil
}

// ListByUser retrieves all URL mappings for a specific user
func (p *PostgresStore) ListByUser(userID string) ([]model.URLMapping, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	query := `
		SELECT code, original_url, user_id, created_at, expires_at, clicks
		FROM url_mappings 
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := p.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var mappings []model.URLMapping
	for rows.Next() {
		var mapping model.URLMapping
		var expiresAt sql.NullTime

		err := rows.Scan(
			&mapping.Code,
			&mapping.Original,
			&mapping.UserID,
			&mapping.CreatedAt,
			&expiresAt,
			&mapping.Clicks,
		)
		if err != nil {
			return nil, err
		}

		if expiresAt.Valid {
			mapping.ExpiresAt = &expiresAt.Time
		}

		mappings = append(mappings, mapping)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return mappings, nil
}

// Delete removes a URL mapping by code
func (p *PostgresStore) Delete(code string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `DELETE FROM url_mappings WHERE code = $1`

	result, err := p.pool.Exec(ctx, query, code)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("no URL mapping found for code: %s", code)
	}

	return nil
}

// Close closes the database connection pool
func (p *PostgresStore) Close() {
	if p.pool != nil {
		p.pool.Close()
	}
}

// CleanupExpired removes expired URL mappings
func (p *PostgresStore) CleanupExpired() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	query := `DELETE FROM url_mappings WHERE expires_at IS NOT NULL AND expires_at < NOW()`

	_, err := p.pool.Exec(ctx, query)
	return err
}
