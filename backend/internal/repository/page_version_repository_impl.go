package repository

import (
	"context"
	"github.com/yixian-huang/inkless/backend/internal/model"
	"gorm.io/gorm"
)

type GormPageVersionRepository struct {
	db *gorm.DB
}

func NewGormPageVersionRepository(db *gorm.DB) PageVersionRepository {
	return &GormPageVersionRepository{db: db}
}

func (r *GormPageVersionRepository) Create(ctx context.Context, v *model.PageVersion) error {
	return r.db.WithContext(ctx).Create(v).Error
}

func (r *GormPageVersionRepository) FindByPageIDAndVersion(ctx context.Context, pageID uint, version int) (*model.PageVersion, error) {
	var v model.PageVersion
	err := r.db.WithContext(ctx).Where("page_id = ? AND version = ?", pageID, version).First(&v).Error
	return &v, err
}

func (r *GormPageVersionRepository) ListByPageID(ctx context.Context, pageID uint, offset, limit int) ([]*model.PageVersion, int64, error) {
	var versions []*model.PageVersion
	var count int64
	q := r.db.WithContext(ctx).Model(&model.PageVersion{}).Where("page_id = ?", pageID)
	if err := q.Count(&count).Error; err != nil {
		return nil, 0, err
	}
	err := q.Order("version DESC").Offset(offset).Limit(limit).Find(&versions).Error
	return versions, count, err
}

func (r *GormPageVersionRepository) GetLatestVersion(ctx context.Context, pageID uint) (int, error) {
	var result struct{ Max int }
	err := r.db.WithContext(ctx).Model(&model.PageVersion{}).Select("COALESCE(MAX(version), 0) as max").Where("page_id = ?", pageID).Scan(&result).Error
	return result.Max, err
}

func (r *GormPageVersionRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&model.PageVersion{}, id).Error
}
