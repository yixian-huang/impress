package repository

import (
	"context"
	"errors"

	"blotting-consultancy/internal/model"

	"gorm.io/gorm"
)

// GormStorageConfigRepository implements StorageConfigRepository using GORM
type GormStorageConfigRepository struct {
	db *gorm.DB
}

// NewGormStorageConfigRepository creates a new GormStorageConfigRepository
func NewGormStorageConfigRepository(db *gorm.DB) StorageConfigRepository {
	return &GormStorageConfigRepository{db: db}
}

// Get returns the current storage config (singleton row with ID=1)
func (r *GormStorageConfigRepository) Get(ctx context.Context) (*model.StorageConfig, error) {
	var config model.StorageConfig
	err := r.db.WithContext(ctx).First(&config).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Return default local config
			return &model.StorageConfig{
				ID:       1,
				Strategy: model.StorageLocal,
			}, nil
		}
		return nil, err
	}
	return &config, nil
}

// Upsert creates or updates the storage config
func (r *GormStorageConfigRepository) Upsert(ctx context.Context, config *model.StorageConfig) error {
	if err := config.Validate(); err != nil {
		return err
	}
	config.ID = 1 // singleton
	return r.db.WithContext(ctx).Save(config).Error
}
