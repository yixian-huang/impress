package repository

import (
	"context"
	"errors"

	"github.com/yixian-huang/inkless/backend/internal/model"
	"gorm.io/gorm"
)

// GormContentVersionRepository implements ContentVersionRepository using GORM
type GormContentVersionRepository struct {
	db *gorm.DB
}

// NewGormContentVersionRepository creates a new GormContentVersionRepository
func NewGormContentVersionRepository(db *gorm.DB) ContentVersionRepository {
	return &GormContentVersionRepository{db: db}
}

// Create creates a new content version
func (r *GormContentVersionRepository) Create(ctx context.Context, version *model.ContentVersion) error {
	if err := version.Validate(); err != nil {
		return err
	}
	return r.db.WithContext(ctx).Create(version).Error
}

// FindByID finds a content version by ID
func (r *GormContentVersionRepository) FindByID(ctx context.Context, id uint) (*model.ContentVersion, error) {
	var version model.ContentVersion
	err := r.db.WithContext(ctx).First(&version, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("content version not found")
		}
		return nil, err
	}
	return &version, nil
}

// FindByPageKeyAndVersion finds a specific version of a page
func (r *GormContentVersionRepository) FindByPageKeyAndVersion(ctx context.Context, pageKey model.PageKey, version int) (*model.ContentVersion, error) {
	var contentVersion model.ContentVersion
	err := r.db.WithContext(ctx).
		Where("page_key = ? AND version = ?", pageKey, version).
		First(&contentVersion).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("content version not found")
		}
		return nil, err
	}
	return &contentVersion, nil
}

// ListByPageKey returns all versions for a page with pagination
func (r *GormContentVersionRepository) ListByPageKey(ctx context.Context, pageKey model.PageKey, offset, limit int) ([]*model.ContentVersion, int64, error) {
	var versions []*model.ContentVersion
	var total int64

	// Get total count
	if err := r.db.WithContext(ctx).
		Model(&model.ContentVersion{}).
		Where("page_key = ?", pageKey).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results ordered by version descending (latest first)
	if err := r.db.WithContext(ctx).
		Where("page_key = ?", pageKey).
		Order("version DESC").
		Offset(offset).
		Limit(limit).
		Find(&versions).Error; err != nil {
		return nil, 0, err
	}

	return versions, total, nil
}

// GetLatestVersion returns the latest version number for a page
func (r *GormContentVersionRepository) GetLatestVersion(ctx context.Context, pageKey model.PageKey) (int, error) {
	var maxVersion int
	err := r.db.WithContext(ctx).
		Model(&model.ContentVersion{}).
		Where("page_key = ?", pageKey).
		Select("COALESCE(MAX(version), 0)").
		Scan(&maxVersion).Error
	if err != nil {
		return 0, err
	}
	return maxVersion, nil
}

// Delete deletes a content version by ID
func (r *GormContentVersionRepository) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&model.ContentVersion{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("content version not found")
	}
	return nil
}
