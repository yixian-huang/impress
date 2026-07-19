package repository

import (
	"context"
	"github.com/yixian-huang/inkless/backend/internal/model"
	"gorm.io/gorm"
)

type GormPageTemplateRepository struct {
	db *gorm.DB
}

func NewGormPageTemplateRepository(db *gorm.DB) PageTemplateRepository {
	return &GormPageTemplateRepository{db: db}
}

func (r *GormPageTemplateRepository) Create(ctx context.Context, tmpl *model.PageTemplate) error {
	return r.db.WithContext(ctx).Create(tmpl).Error
}

func (r *GormPageTemplateRepository) Update(ctx context.Context, tmpl *model.PageTemplate) error {
	return r.db.WithContext(ctx).Save(tmpl).Error
}

func (r *GormPageTemplateRepository) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&model.PageTemplate{}, id)
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return result.Error
}

func (r *GormPageTemplateRepository) FindByID(ctx context.Context, id uint) (*model.PageTemplate, error) {
	var tmpl model.PageTemplate
	err := r.db.WithContext(ctx).First(&tmpl, id).Error
	return &tmpl, err
}

func (r *GormPageTemplateRepository) FindByKey(ctx context.Context, key string) (*model.PageTemplate, error) {
	var tmpl model.PageTemplate
	err := r.db.WithContext(ctx).Where("key = ?", key).First(&tmpl).Error
	return &tmpl, err
}

func (r *GormPageTemplateRepository) List(ctx context.Context, category string) ([]*model.PageTemplate, error) {
	q := r.db.WithContext(ctx).Model(&model.PageTemplate{})
	if category != "" {
		q = q.Where("category = ?", category)
	}
	var templates []*model.PageTemplate
	err := q.Order("category ASC, key ASC").Find(&templates).Error
	return templates, err
}
