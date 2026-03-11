-- +goose Up
-- Add keywords field to pages table
ALTER TABLE pages ADD COLUMN keywords TEXT DEFAULT '{}';

-- +goose Down
-- SQLite doesn't support DROP COLUMN before 3.35.0, so this is best-effort
