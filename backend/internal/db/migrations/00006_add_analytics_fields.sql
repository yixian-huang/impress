-- +goose Up
ALTER TABLE page_views ADD COLUMN IF NOT EXISTS visitor_id VARCHAR(64);
ALTER TABLE page_views ADD COLUMN IF NOT EXISTS referer VARCHAR(500);
CREATE INDEX IF NOT EXISTS idx_pageview_visitor ON page_views(visitor_id);

-- +goose Down
-- Best-effort: SQLite <3.35 doesn't support DROP COLUMN
