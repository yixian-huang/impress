package repository

import (
	"context"
	"errors"

	"github.com/yixian-huang/inkless/backend/internal/model"

	"gorm.io/gorm"
)

// GormMediaFolderRepository implements MediaFolderRepository using GORM
type GormMediaFolderRepository struct {
	db *gorm.DB
}

// NewGormMediaFolderRepository creates a new GormMediaFolderRepository
func NewGormMediaFolderRepository(db *gorm.DB) MediaFolderRepository {
	return &GormMediaFolderRepository{db: db}
}

// Create creates a new media folder
func (r *GormMediaFolderRepository) Create(ctx context.Context, folder *model.MediaFolder) error {
	if err := folder.Validate(); err != nil {
		return err
	}
	return r.db.WithContext(ctx).Create(folder).Error
}

// FindByID finds a media folder by ID
func (r *GormMediaFolderRepository) FindByID(ctx context.Context, id uint) (*model.MediaFolder, error) {
	var folder model.MediaFolder
	err := r.db.WithContext(ctx).First(&folder, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("media folder not found")
		}
		return nil, err
	}
	return &folder, nil
}

// ListAll returns all media folders
func (r *GormMediaFolderRepository) ListAll(ctx context.Context) ([]*model.MediaFolder, error) {
	var folders []*model.MediaFolder
	err := r.db.WithContext(ctx).Order("path ASC").Find(&folders).Error
	return folders, err
}

// Update updates a media folder
func (r *GormMediaFolderRepository) Update(ctx context.Context, folder *model.MediaFolder) error {
	if err := folder.Validate(); err != nil {
		return err
	}
	return r.db.WithContext(ctx).Save(folder).Error
}

// Delete deletes a media folder by ID
func (r *GormMediaFolderRepository) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&model.MediaFolder{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("media folder not found")
	}
	return nil
}

// FindChildren returns direct children of a folder
func (r *GormMediaFolderRepository) FindChildren(ctx context.Context, parentID uint) ([]*model.MediaFolder, error) {
	var folders []*model.MediaFolder
	err := r.db.WithContext(ctx).Where("parent_id = ?", parentID).Find(&folders).Error
	return folders, err
}

// UpdateMediaFolder updates the folder_id of a media item
func (r *GormMediaFolderRepository) UpdateMediaFolder(ctx context.Context, mediaID uint, folderID *uint) error {
	return r.db.WithContext(ctx).Model(&model.Media{}).Where("id = ?", mediaID).Update("folder_id", folderID).Error
}
