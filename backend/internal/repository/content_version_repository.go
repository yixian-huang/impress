package repository

import (
	"context"

	"github.com/yixian-huang/inkless/backend/internal/model"
)

// ContentVersionRepository defines the interface for content version data access
type ContentVersionRepository interface {
	// Create creates a new content version
	Create(ctx context.Context, version *model.ContentVersion) error

	// FindByID finds a content version by ID
	FindByID(ctx context.Context, id uint) (*model.ContentVersion, error)

	// FindByPageKeyAndVersion finds a specific version of a page
	FindByPageKeyAndVersion(ctx context.Context, pageKey model.PageKey, version int) (*model.ContentVersion, error)

	// ListByPageKey returns all versions for a page with pagination
	ListByPageKey(ctx context.Context, pageKey model.PageKey, offset, limit int) ([]*model.ContentVersion, int64, error)

	// GetLatestVersion returns the latest version number for a page
	GetLatestVersion(ctx context.Context, pageKey model.PageKey) (int, error)

	// Delete deletes a content version by ID
	Delete(ctx context.Context, id uint) error
}
