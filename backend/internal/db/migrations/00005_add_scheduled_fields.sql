-- +goose Up
ALTER TABLE articles ADD COLUMN scheduled_at DATETIME;
ALTER TABLE pages ADD COLUMN scheduled_at DATETIME;

-- +goose Down
-- Best-effort: SQLite <3.35 doesn't support DROP COLUMN
