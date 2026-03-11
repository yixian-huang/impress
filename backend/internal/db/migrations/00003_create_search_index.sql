-- +goose Up

-- SQLite FTS5 virtual table for full-text search
-- Note: This migration only runs on SQLite. For PostgreSQL, use tsvector columns.
-- The application code detects the DB driver and uses the appropriate search strategy.

-- +goose StatementBegin
CREATE VIRTUAL TABLE IF NOT EXISTS search_index_fts USING fts5(
    content_type,
    content_id UNINDEXED,
    locale,
    title,
    body,
    slug UNINDEXED,
    tokenize='unicode61'
);
-- +goose StatementEnd

-- +goose Down
DROP TABLE IF EXISTS search_index_fts;
