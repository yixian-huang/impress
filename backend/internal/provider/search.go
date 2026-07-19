package provider

import "context"

// SearchResult represents a single search hit.
type SearchResult struct {
	ID      uint    `json:"id"`
	Type    string  `json:"type"` // "article", "page"
	Title   string  `json:"title"`
	Snippet string  `json:"snippet"` // highlighted text excerpt
	URL     string  `json:"url"`
	Locale  string  `json:"locale"`
	Score   float64 `json:"score"`
}

// SearchResponse wraps paginated search results.
type SearchResponse struct {
	Results  []SearchResult `json:"results"`
	Total    int64          `json:"total"`
	Page     int            `json:"page"`
	PageSize int            `json:"pageSize"`
	Query    string         `json:"query"`
}

// SearchProvider defines the contract for full-text search backends.
// Default implementation uses SQLite FTS5 / PostgreSQL tsvector.
// Plugins can replace with Meilisearch, Elasticsearch, etc.
type SearchProvider interface {
	Search(ctx context.Context, query string, locale string, contentType string, page int, pageSize int) (*SearchResponse, error)
	Suggest(ctx context.Context, prefix string, locale string, limit int) ([]string, error)
	IndexArticle(ctx context.Context, id uint, locale string, title string, body string, slug string) error
	IndexPage(ctx context.Context, id uint, locale string, title string, body string, slug string) error
	RemoveFromIndex(ctx context.Context, contentType string, id uint) error
	RebuildIndex(ctx context.Context) error
}
