package repository

import (
	"context"

	"github.com/yixian-huang/inkless/backend/internal/model"
	"gorm.io/gorm"
)

// GormArticleVersionRepository implements ArticleVersionRepository using GORM.
type GormArticleVersionRepository struct {
	db *gorm.DB
}

// NewGormArticleVersionRepository creates a new repository.
func NewGormArticleVersionRepository(db *gorm.DB) ArticleVersionRepository {
	return &GormArticleVersionRepository{db: db}
}

func (r *GormArticleVersionRepository) Create(ctx context.Context, v *model.ArticleVersion) error {
	if err := v.Validate(); err != nil {
		return err
	}
	return r.db.WithContext(ctx).Create(v).Error
}

func (r *GormArticleVersionRepository) FindByArticleIDAndVersion(ctx context.Context, articleID uint, version int) (*model.ArticleVersion, error) {
	var v model.ArticleVersion
	err := r.db.WithContext(ctx).
		Where("article_id = ? AND version = ?", articleID, version).
		First(&v).Error
	if err != nil {
		return nil, err
	}
	return &v, nil
}

func (r *GormArticleVersionRepository) ListByArticleID(ctx context.Context, articleID uint, offset, limit int) ([]*model.ArticleVersion, int64, error) {
	var versions []*model.ArticleVersion
	var count int64
	q := r.db.WithContext(ctx).Model(&model.ArticleVersion{}).Where("article_id = ?", articleID)
	if err := q.Count(&count).Error; err != nil {
		return nil, 0, err
	}
	// List without full body-heavy snapshot for list performance — still return snapshot
	// but order by version desc. Bodies live inside JSON; keep full for compare UX simplicity.
	err := q.Order("version DESC").Offset(offset).Limit(limit).Find(&versions).Error
	return versions, count, err
}

func (r *GormArticleVersionRepository) GetLatestVersion(ctx context.Context, articleID uint) (int, error) {
	var result struct{ Max int }
	err := r.db.WithContext(ctx).
		Model(&model.ArticleVersion{}).
		Select("COALESCE(MAX(version), 0) as max").
		Where("article_id = ?", articleID).
		Scan(&result).Error
	return result.Max, err
}
