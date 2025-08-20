-- +goose Up
CREATE TABLE IF NOT EXISTS url_mappings (
    code VARCHAR(255) PRIMARY KEY,
    original_url TEXT NOT NULL,
    user_id VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    expires_at TIMESTAMPTZ,
    clicks INTEGER NOT NULL DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_url_mappings_user_id ON url_mappings(user_id);
CREATE INDEX IF NOT EXISTS idx_url_mappings_expires_at ON url_mappings(expires_at);

-- +goose Down
DROP INDEX IF EXISTS idx_url_mappings_expires_at;
DROP INDEX IF EXISTS idx_url_mappings_user_id;
DROP TABLE IF EXISTS url_mappings;
