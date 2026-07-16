package repository

import (
	"context"
	"errors"

	"blotting-consultancy/internal/model"

	"gorm.io/gorm"
)

// GormAIConfigRepository implements AIConfigRepository using GORM.
type GormAIConfigRepository struct {
	db *gorm.DB
}

func NewGormAIConfigRepository(db *gorm.DB) AIConfigRepository {
	return &GormAIConfigRepository{db: db}
}

// Get returns the singleton AI config, defaulting to disabled when absent.
func (r *GormAIConfigRepository) Get(ctx context.Context) (*model.AIConfig, error) {
	var config model.AIConfig
	err := r.db.WithContext(ctx).First(&config, model.AIConfigSingletonID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &model.AIConfig{
				ID:       model.AIConfigSingletonID,
				Provider: model.AIProviderDisabled,
			}, nil
		}
		return nil, err
	}
	return &config, nil
}

// Upsert creates or updates the singleton AI config.
func (r *GormAIConfigRepository) Upsert(ctx context.Context, config *model.AIConfig) error {
	if err := config.Validate(); err != nil {
		return err
	}
	config.ID = model.AIConfigSingletonID
	return r.db.WithContext(ctx).Save(config).Error
}
