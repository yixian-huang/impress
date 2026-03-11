package repository

import (
	"context"

	"blotting-consultancy/internal/model"
)

// CommentRepository defines the interface for comment data access
type CommentRepository interface {
	// Create creates a new comment
	Create(ctx context.Context, comment *model.Comment) error

	// FindByID finds a comment by ID with Children preloaded
	FindByID(ctx context.Context, id uint) (*model.Comment, error)

	// Update updates a comment
	Update(ctx context.Context, comment *model.Comment) error

	// Delete deletes a comment by ID
	Delete(ctx context.Context, id uint) error

	// ListByContent returns a paginated list of top-level comments for a given content target
	ListByContent(ctx context.Context, contentType string, contentID uint, status model.CommentStatus, page, pageSize int) ([]*model.Comment, int64, error)

	// ListAll returns a paginated list of all comments with optional status filter
	ListAll(ctx context.Context, status string, page, pageSize int) ([]*model.Comment, int64, error)

	// CountByContent returns the count of approved comments for a given content target
	CountByContent(ctx context.Context, contentType string, contentID uint) (int64, error)

	// UpdateStatus updates the moderation status of a comment
	UpdateStatus(ctx context.Context, id uint, status model.CommentStatus) error

	// SetPinned sets the pinned flag on a comment
	SetPinned(ctx context.Context, id uint, pinned bool) error
}
