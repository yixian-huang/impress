package repository

import (
	"context"

	"github.com/yixian-huang/inkless/backend/internal/model"
)

// MediaUsage represents a reference to a media item from another entity
type MediaUsage struct {
	Type  string `json:"type"` // "article", "page", "content_document"
	ID    string `json:"id"`
	Title string `json:"title"`
	Field string `json:"field"`
}

// MediaRepository defines the interface for media data access
type MediaRepository interface {
	// Create creates a new media record
	Create(ctx context.Context, media *model.Media) error

	// FindByID finds a media record by ID
	FindByID(ctx context.Context, id uint) (*model.Media, error)

	// List returns a paginated list of media records, optionally filtered by MIME type prefix
	List(ctx context.Context, offset, limit int, mimePrefix string) ([]*model.Media, int64, error)

	// Count returns total media records.
	Count(ctx context.Context) (int64, error)

	// Delete deletes a media record by ID
	Delete(ctx context.Context, id uint) error

	// Update updates an existing media record
	Update(ctx context.Context, media *model.Media) error

	// FindUsages searches for references to a media URL across articles, pages, and content documents
	FindUsages(ctx context.Context, mediaURL string) ([]MediaUsage, error)
}
