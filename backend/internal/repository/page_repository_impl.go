package repository

import (
	"context"
	"errors"

	"blotting-consultancy/internal/model"

	"gorm.io/gorm"
)

// GormPageRepository implements PageRepository using GORM
type GormPageRepository struct {
	db *gorm.DB
}

// NewGormPageRepository creates a new GormPageRepository
func NewGormPageRepository(db *gorm.DB) PageRepository {
	return &GormPageRepository{db: db}
}

// UnifiedPageRepository exposes the unified page repository that shares this
// repository's database handle. It supports compatibility constructors that
// still receive the legacy page repository at startup.
func (r *GormPageRepository) UnifiedPageRepository() UnifiedPageRepository {
	return NewGormUnifiedPageRepository(r.db)
}

// Create creates a new page
func (r *GormPageRepository) Create(ctx context.Context, page *model.Page) error {
	if err := page.Validate(); err != nil {
		return err
	}
	return r.db.WithContext(ctx).Create(page).Error
}

// Update updates an existing page
func (r *GormPageRepository) Update(ctx context.Context, page *model.Page) error {
	if err := page.Validate(); err != nil {
		return err
	}
	return r.db.WithContext(ctx).Save(page).Error
}

// Delete soft-deletes a page by ID
func (r *GormPageRepository) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&model.Page{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("page not found")
	}
	return nil
}

// FindByID finds a page by ID
func (r *GormPageRepository) FindByID(ctx context.Context, id uint) (*model.Page, error) {
	var page model.Page
	err := r.db.WithContext(ctx).
		Preload("Parent").
		First(&page, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("page not found")
		}
		return nil, err
	}
	return &page, nil
}

// FindBySlug finds a page by slug
func (r *GormPageRepository) FindBySlug(ctx context.Context, slug string) (*model.Page, error) {
	var page model.Page
	err := r.db.WithContext(ctx).
		Preload("Parent").
		Where("slug = ?", slug).
		First(&page).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("page not found")
		}
		return nil, err
	}
	return &page, nil
}

// List returns pages with optional status and parentID filters
func (r *GormPageRepository) List(ctx context.Context, status string, parentID *uint) ([]*model.Page, error) {
	var pages []*model.Page

	query := r.db.WithContext(ctx).Preload("Parent")

	if status != "" {
		query = query.Where("status = ?", status)
	}
	if parentID != nil {
		query = query.Where("parent_id = ?", *parentID)
	}

	err := query.Order("sort_order ASC, created_at DESC").Find(&pages).Error
	if err != nil {
		return nil, err
	}
	return pages, nil
}

// FindByThemeIDAndContentKey finds a page by theme ID and content key (for seed dedup)
func (r *GormPageRepository) FindByThemeIDAndContentKey(ctx context.Context, themeID string, contentKey string) (*model.Page, error) {
	var page model.Page
	err := r.db.WithContext(ctx).
		Where("theme_id = ? AND content_key = ?", themeID, contentKey).
		First(&page).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("page not found")
		}
		return nil, err
	}
	return &page, nil
}

// ListByThemeID returns pages filtered by themeID with optional status filter
func (r *GormPageRepository) ListByThemeID(ctx context.Context, themeID string, status string) ([]*model.Page, error) {
	var pages []*model.Page
	query := r.db.WithContext(ctx).Where("theme_id = ?", themeID)
	if status != "" {
		query = query.Where("status = ?", status)
	}
	err := query.Order("sort_order ASC, created_at DESC").Find(&pages).Error
	if err != nil {
		return nil, err
	}
	return pages, nil
}

// ListPublishedByThemeID returns published pages for a specific theme
func (r *GormPageRepository) ListPublishedByThemeID(ctx context.Context, themeID string) ([]*model.Page, error) {
	var pages []*model.Page
	err := r.db.WithContext(ctx).
		Where("theme_id = ? AND status = ?", themeID, model.PageStatusPublished).
		Order("sort_order ASC, created_at DESC").
		Find(&pages).Error
	if err != nil {
		return nil, err
	}
	return pages, nil
}

// ListPublished returns all published pages ordered by sort_order
func (r *GormPageRepository) ListPublished(ctx context.Context) ([]*model.Page, error) {
	var pages []*model.Page
	err := r.db.WithContext(ctx).
		Preload("Parent").
		Where("status = ?", model.PageStatusPublished).
		Order("sort_order ASC, created_at DESC").
		Find(&pages).Error
	if err != nil {
		return nil, err
	}
	return pages, nil
}

// UpdateSortOrder updates the sort order for a specific page
func (r *GormPageRepository) UpdateSortOrder(ctx context.Context, id uint, sortOrder int) error {
	result := r.db.WithContext(ctx).
		Model(&model.Page{}).
		Where("id = ?", id).
		Update("sort_order", sortOrder)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("page not found")
	}
	return nil
}
