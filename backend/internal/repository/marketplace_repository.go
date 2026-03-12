package repository

import (
	"context"

	"blotting-consultancy/internal/model"
)

// MarketplaceFilter defines optional filters for listing marketplace items
type MarketplaceFilter struct {
	Type     string
	Category string
	Status   string
	Search   string
	Page     int
	PageSize int
}

// MarketplaceRepository defines the interface for marketplace data access
type MarketplaceRepository interface {
	// List returns marketplace items with optional filters (paginated)
	List(ctx context.Context, filter MarketplaceFilter) ([]*model.MarketplaceItem, int64, error)

	// GetBySlug returns a marketplace item by its slug (including versions)
	GetBySlug(ctx context.Context, slug string) (*model.MarketplaceItem, error)

	// GetByID returns a marketplace item by its ID
	GetByID(ctx context.Context, id uint) (*model.MarketplaceItem, error)

	// Create inserts a new marketplace item
	Create(ctx context.Context, item *model.MarketplaceItem) error

	// Update saves changes to an existing marketplace item
	Update(ctx context.Context, item *model.MarketplaceItem) error

	// Delete soft-deletes a marketplace item by ID
	Delete(ctx context.Context, id uint) error

	// IncrementDownloads atomically increments the download counter for an item
	IncrementDownloads(ctx context.Context, id uint) error

	// CreateVersion inserts a new version record for an item
	CreateVersion(ctx context.Context, version *model.MarketplaceVersion) error

	// ListVersions returns all versions for a given item ID ordered by created_at desc
	ListVersions(ctx context.Context, itemID uint) ([]*model.MarketplaceVersion, error)
}
