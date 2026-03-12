package repository

import (
	"context"
	"errors"
	"strings"

	"blotting-consultancy/internal/model"

	"gorm.io/gorm"
)

// GormMarketplaceRepository implements MarketplaceRepository using GORM
type GormMarketplaceRepository struct {
	db *gorm.DB
}

// NewGormMarketplaceRepository creates a new GormMarketplaceRepository
func NewGormMarketplaceRepository(db *gorm.DB) MarketplaceRepository {
	return &GormMarketplaceRepository{db: db}
}

// List returns marketplace items with optional filters and pagination
func (r *GormMarketplaceRepository) List(ctx context.Context, filter MarketplaceFilter) ([]*model.MarketplaceItem, int64, error) {
	q := r.db.WithContext(ctx).Model(&model.MarketplaceItem{})

	if filter.Type != "" {
		q = q.Where("type = ?", filter.Type)
	}
	if filter.Category != "" {
		q = q.Where("category = ?", filter.Category)
	}
	if filter.Status != "" {
		q = q.Where("status = ?", filter.Status)
	} else {
		// Default: only show active items
		q = q.Where("status = ?", model.MarketplaceItemStatusActive)
	}
	if filter.Search != "" {
		search := "%" + strings.ToLower(filter.Search) + "%"
		q = q.Where("LOWER(name) LIKE ? OR LOWER(name_zh) LIKE ? OR LOWER(description) LIKE ?", search, search, search)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	page := filter.Page
	if page < 1 {
		page = 1
	}
	pageSize := filter.PageSize
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	var items []*model.MarketplaceItem
	err := q.Order("downloads DESC, created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&items).Error
	if err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

// GetBySlug returns a marketplace item by slug including its versions
func (r *GormMarketplaceRepository) GetBySlug(ctx context.Context, slug string) (*model.MarketplaceItem, error) {
	var item model.MarketplaceItem
	err := r.db.WithContext(ctx).
		Preload("Versions", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at DESC")
		}).
		Where("slug = ?", slug).
		First(&item).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("marketplace item not found")
		}
		return nil, err
	}
	return &item, nil
}

// GetByID returns a marketplace item by ID
func (r *GormMarketplaceRepository) GetByID(ctx context.Context, id uint) (*model.MarketplaceItem, error) {
	var item model.MarketplaceItem
	err := r.db.WithContext(ctx).
		Preload("Versions", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at DESC")
		}).
		First(&item, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("marketplace item not found")
		}
		return nil, err
	}
	return &item, nil
}

// Create inserts a new marketplace item
func (r *GormMarketplaceRepository) Create(ctx context.Context, item *model.MarketplaceItem) error {
	return r.db.WithContext(ctx).Create(item).Error
}

// Update saves changes to an existing marketplace item
func (r *GormMarketplaceRepository) Update(ctx context.Context, item *model.MarketplaceItem) error {
	return r.db.WithContext(ctx).Save(item).Error
}

// Delete soft-deletes a marketplace item by ID
func (r *GormMarketplaceRepository) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&model.MarketplaceItem{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("marketplace item not found")
	}
	return nil
}

// IncrementDownloads atomically increments the download counter for an item
func (r *GormMarketplaceRepository) IncrementDownloads(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).
		Model(&model.MarketplaceItem{}).
		Where("id = ?", id).
		UpdateColumn("downloads", gorm.Expr("downloads + 1"))
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("marketplace item not found")
	}
	return nil
}

// CreateVersion inserts a new version record for an item
func (r *GormMarketplaceRepository) CreateVersion(ctx context.Context, version *model.MarketplaceVersion) error {
	return r.db.WithContext(ctx).Create(version).Error
}

// ListVersions returns all versions for a given item ID ordered by created_at desc
func (r *GormMarketplaceRepository) ListVersions(ctx context.Context, itemID uint) ([]*model.MarketplaceVersion, error) {
	var versions []*model.MarketplaceVersion
	err := r.db.WithContext(ctx).
		Where("item_id = ?", itemID).
		Order("created_at DESC").
		Find(&versions).Error
	if err != nil {
		return nil, err
	}
	return versions, nil
}
