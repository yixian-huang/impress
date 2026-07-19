package repository

import (
	"context"
	"errors"

	"github.com/yixian-huang/inkless/backend/internal/model"
	"gorm.io/gorm"
)

// GormContentDocumentRepository implements ContentDocumentRepository using GORM
type GormContentDocumentRepository struct {
	db *gorm.DB
}

// NewGormContentDocumentRepository creates a new GormContentDocumentRepository
func NewGormContentDocumentRepository(db *gorm.DB) ContentDocumentRepository {
	return &GormContentDocumentRepository{db: db}
}

// Create creates a new content document
func (r *GormContentDocumentRepository) Create(ctx context.Context, doc *model.ContentDocument) error {
	if err := doc.Validate(); err != nil {
		return err
	}
	return r.db.WithContext(ctx).Create(doc).Error
}

// FindByPageKey finds a content document by page key
func (r *GormContentDocumentRepository) FindByPageKey(ctx context.Context, pageKey model.PageKey) (*model.ContentDocument, error) {
	var doc model.ContentDocument
	err := r.db.WithContext(ctx).Where("page_key = ?", pageKey).First(&doc).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("content document not found")
		}
		return nil, err
	}
	return &doc, nil
}

// Update updates an existing content document with optimistic locking
func (r *GormContentDocumentRepository) Update(ctx context.Context, doc *model.ContentDocument) error {
	if err := doc.Validate(); err != nil {
		return err
	}
	result := r.db.WithContext(ctx).Save(doc)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("content document not found")
	}
	return nil
}

// UpdateDraft updates only the draft fields with optimistic locking
func (r *GormContentDocumentRepository) UpdateDraft(ctx context.Context, pageKey model.PageKey, expectedDraftVersion int, draftConfig model.JSONMap) (int, error) {
	// Use optimistic locking: only update if draft_version matches expected
	result := r.db.WithContext(ctx).
		Model(&model.ContentDocument{}).
		Where("page_key = ? AND draft_version = ?", pageKey, expectedDraftVersion).
		Updates(map[string]interface{}{
			"draft_config":  draftConfig,
			"draft_version": gorm.Expr("draft_version + 1"),
		})

	if result.Error != nil {
		return 0, result.Error
	}

	if result.RowsAffected == 0 {
		return 0, errors.New("draft version conflict or document not found")
	}

	// Retrieve the new version
	var doc model.ContentDocument
	if err := r.db.WithContext(ctx).Where("page_key = ?", pageKey).First(&doc).Error; err != nil {
		return 0, err
	}

	return doc.DraftVersion, nil
}

// UpdatePublished updates only the published fields atomically
func (r *GormContentDocumentRepository) UpdatePublished(ctx context.Context, pageKey model.PageKey, publishedConfig model.JSONMap, publishedVersion int) error {
	result := r.db.WithContext(ctx).
		Model(&model.ContentDocument{}).
		Where("page_key = ?", pageKey).
		Updates(map[string]interface{}{
			"published_config":  publishedConfig,
			"published_version": publishedVersion,
		})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("content document not found")
	}

	return nil
}

// List returns all content documents
func (r *GormContentDocumentRepository) List(ctx context.Context) ([]*model.ContentDocument, error) {
	var docs []*model.ContentDocument
	err := r.db.WithContext(ctx).Order("page_key ASC").Find(&docs).Error
	if err != nil {
		return nil, err
	}
	return docs, nil
}

// Delete deletes a content document by page key
func (r *GormContentDocumentRepository) Delete(ctx context.Context, pageKey model.PageKey) error {
	result := r.db.WithContext(ctx).Where("page_key = ?", pageKey).Delete(&model.ContentDocument{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("content document not found")
	}
	return nil
}
