package repository

import (
	"context"
	"errors"
	"fmt"
	"github.com/yixian-huang/inkless/backend/internal/model"
	"gorm.io/gorm"
)

type GormSiteConfigRepository struct {
	db *gorm.DB
}

func NewGormSiteConfigRepository(db *gorm.DB) SiteConfigRepository {
	return &GormSiteConfigRepository{db: db}
}

func (r *GormSiteConfigRepository) FindByKey(ctx context.Context, key string) (*model.SiteConfig, error) {
	var sc model.SiteConfig
	err := r.db.WithContext(ctx).Where("key = ?", key).First(&sc).Error
	return &sc, err
}

func (r *GormSiteConfigRepository) Update(ctx context.Context, config *model.SiteConfig) error {
	return r.db.WithContext(ctx).Save(config).Error
}

func (r *GormSiteConfigRepository) Upsert(ctx context.Context, config *model.SiteConfig) error {
	var existing model.SiteConfig
	err := r.db.WithContext(ctx).Where("key = ?", config.Key).First(&existing).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return r.db.WithContext(ctx).Create(config).Error
	}
	if err != nil {
		return err
	}
	config.ID = existing.ID
	return r.db.WithContext(ctx).Save(config).Error
}

func (r *GormSiteConfigRepository) UpdateDraft(ctx context.Context, key string, expectedVersion int, draftConfig model.JSONMap) (int, error) {
	result := r.db.WithContext(ctx).Table("site_configs").Where("key = ? AND draft_version = ?", key, expectedVersion).Updates(map[string]interface{}{
		"draft_config":  draftConfig,
		"draft_version": gorm.Expr("draft_version + 1"),
	})
	if result.Error != nil {
		return 0, result.Error
	}
	if result.RowsAffected == 0 {
		return 0, errors.New("draft version conflict or config not found")
	}
	var sc model.SiteConfig
	if err := r.db.WithContext(ctx).Select("draft_version").Where("key = ?", key).First(&sc).Error; err != nil {
		return 0, fmt.Errorf("fetch new version: %w", err)
	}
	return sc.DraftVersion, nil
}

func (r *GormSiteConfigRepository) UpdatePublished(ctx context.Context, key string, publishedConfig model.JSONMap, publishedVersion int) error {
	return r.db.WithContext(ctx).Table("site_configs").Where("key = ?", key).Updates(map[string]interface{}{
		"published_config":  publishedConfig,
		"published_version": publishedVersion,
	}).Error
}
