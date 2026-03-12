package repository

import (
	"context"
	"errors"

	"blotting-consultancy/internal/model"

	"gorm.io/gorm"
)

// GormChunkedUploadRepository implements ChunkedUploadRepository using GORM
type GormChunkedUploadRepository struct {
	db *gorm.DB
}

// NewGormChunkedUploadRepository creates a new GormChunkedUploadRepository
func NewGormChunkedUploadRepository(db *gorm.DB) ChunkedUploadRepository {
	return &GormChunkedUploadRepository{db: db}
}

// Create creates a new chunked upload record
func (r *GormChunkedUploadRepository) Create(ctx context.Context, upload *model.ChunkedUpload) error {
	return r.db.WithContext(ctx).Create(upload).Error
}

// FindByID finds a chunked upload by ID
func (r *GormChunkedUploadRepository) FindByID(ctx context.Context, id string) (*model.ChunkedUpload, error) {
	var upload model.ChunkedUpload
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&upload).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("chunked upload not found")
		}
		return nil, err
	}
	return &upload, nil
}

// Update updates a chunked upload record
func (r *GormChunkedUploadRepository) Update(ctx context.Context, upload *model.ChunkedUpload) error {
	return r.db.WithContext(ctx).Save(upload).Error
}

// Delete deletes a chunked upload record
func (r *GormChunkedUploadRepository) Delete(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Where("id = ?", id).Delete(&model.ChunkedUpload{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("chunked upload not found")
	}
	return nil
}
