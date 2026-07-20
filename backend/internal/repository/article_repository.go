package repository

import (
	"context"
	"errors"
	"time"

	"github.com/yixian-huang/inkless/backend/internal/model"
)

var ErrArticleVersionConflict = errors.New("article version conflict")

// ArticleListFilter is the admin list query for articles.
type ArticleListFilter struct {
	Offset     int
	Limit      int
	Status     string
	CategoryID *uint
	TagID      *uint
	// Query matches slug / zh_title / en_title (case-insensitive LIKE).
	Query string
	// Sort is a pre-validated SQL ORDER BY clause (never pass raw client input).
	Sort string
}

// ArticleRepository defines the interface for article data access
type ArticleRepository interface {
	// Create creates a new article
	Create(ctx context.Context, article *model.Article) error

	// FindByID finds an article by ID with Category and Tags preloaded
	FindByID(ctx context.Context, id uint) (*model.Article, error)

	// FindBySlug finds an article by slug with Category and Tags preloaded
	FindBySlug(ctx context.Context, slug string) (*model.Article, error)

	// Update updates an article (unconditional save)
	Update(ctx context.Context, article *model.Article) error
	// UpdateIfMatch updates only when updated_at still matches expectedUpdatedAt (optimistic lock).
	// Returns ErrArticleVersionConflict when another writer changed the row.
	UpdateIfMatch(ctx context.Context, article *model.Article, expectedUpdatedAt time.Time) error
	UpdateScheduledPublication(ctx context.Context, article *model.Article, expectedUpdatedAt time.Time) error

	// Delete deletes an article by ID
	Delete(ctx context.Context, id uint) error

	// List returns a paginated admin list (bodies omitted). Prefer ListFilter for new code.
	List(ctx context.Context, offset, limit int, status string, categoryID *uint, tagID *uint) ([]*model.Article, int64, error)

	// ListFilter is the admin article list with q/sort/filters.
	ListFilter(ctx context.Context, f ArticleListFilter) ([]*model.Article, int64, error)

	// ListPublished returns a paginated list of published articles with optional filters.
	// Includes body fields so public list can derive short excerpts; full HTML
	// content for reading still comes from FindBySlug.
	ListPublished(ctx context.Context, offset, limit int, categorySlug string, tagSlug string) ([]*model.Article, int64, error)

	// Count returns the number of articles, optionally filtered by status (empty = all).
	Count(ctx context.Context, status string) (int64, error)
}
