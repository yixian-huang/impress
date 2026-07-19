package repository

import (
	"context"

	"github.com/yixian-huang/inkless/backend/internal/model"
)

// PageRepository defines the interface for page data access
type PageRepository interface {
	// Create creates a new page
	Create(ctx context.Context, page *model.Page) error

	// Update updates an existing page
	Update(ctx context.Context, page *model.Page) error

	// Delete soft-deletes a page by ID
	Delete(ctx context.Context, id uint) error

	// FindByID finds a page by ID
	FindByID(ctx context.Context, id uint) (*model.Page, error)

	// FindBySlug finds a page by slug
	FindBySlug(ctx context.Context, slug string) (*model.Page, error)

	// FindByThemeIDAndContentKey finds a page by theme ID and content key (for seed dedup)
	FindByThemeIDAndContentKey(ctx context.Context, themeID string, contentKey string) (*model.Page, error)

	// List returns pages with optional status and parentID filters
	List(ctx context.Context, status string, parentID *uint) ([]*model.Page, error)

	// ListByThemeID returns pages filtered by themeID with optional status filter
	ListByThemeID(ctx context.Context, themeID string, status string) ([]*model.Page, error)

	// ListPublished returns all published pages ordered by sort_order
	ListPublished(ctx context.Context) ([]*model.Page, error)

	// ListPublishedByThemeID returns published pages for a specific theme
	ListPublishedByThemeID(ctx context.Context, themeID string) ([]*model.Page, error)

	// UpdateSortOrder updates the sort order for a specific page
	UpdateSortOrder(ctx context.Context, id uint, sortOrder int) error
}
