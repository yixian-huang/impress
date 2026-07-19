package repository

import (
	"context"
	"errors"

	"github.com/yixian-huang/inkless/backend/internal/model"

	"gorm.io/gorm"
)

// GormGlossaryRepository implements GlossaryRepository using GORM
type GormGlossaryRepository struct {
	db *gorm.DB
}

// NewGormGlossaryRepository creates a new GormGlossaryRepository
func NewGormGlossaryRepository(db *gorm.DB) GlossaryRepository {
	return &GormGlossaryRepository{db: db}
}

// Create creates a new glossary term
func (r *GormGlossaryRepository) Create(ctx context.Context, glossary *model.Glossary) error {
	if err := glossary.Validate(); err != nil {
		return err
	}
	return r.db.WithContext(ctx).Create(glossary).Error
}

// FindByID finds a glossary term by ID
func (r *GormGlossaryRepository) FindByID(ctx context.Context, id uint) (*model.Glossary, error) {
	var glossary model.Glossary
	err := r.db.WithContext(ctx).First(&glossary, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("glossary term not found")
		}
		return nil, err
	}
	return &glossary, nil
}

// Update updates a glossary term
func (r *GormGlossaryRepository) Update(ctx context.Context, glossary *model.Glossary) error {
	if err := glossary.Validate(); err != nil {
		return err
	}
	return r.db.WithContext(ctx).Save(glossary).Error
}

// Delete deletes a glossary term by ID
func (r *GormGlossaryRepository) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&model.Glossary{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("glossary term not found")
	}
	return nil
}

// List returns a paginated list of glossary terms with optional language filter
func (r *GormGlossaryRepository) List(ctx context.Context, offset, limit int, sourceLang, targetLang string) ([]*model.Glossary, int64, error) {
	var items []*model.Glossary
	var total int64

	query := r.db.WithContext(ctx).Model(&model.Glossary{})

	if sourceLang != "" {
		query = query.Where("source_lang = ?", sourceLang)
	}
	if targetLang != "" {
		query = query.Where("target_lang = ?", targetLang)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	dataQuery := r.db.WithContext(ctx).Model(&model.Glossary{})
	if sourceLang != "" {
		dataQuery = dataQuery.Where("source_lang = ?", sourceLang)
	}
	if targetLang != "" {
		dataQuery = dataQuery.Where("target_lang = ?", targetLang)
	}

	if err := dataQuery.
		Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&items).Error; err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

// FindByLangs returns all glossary terms for a given language pair
func (r *GormGlossaryRepository) FindByLangs(ctx context.Context, sourceLang, targetLang string) ([]*model.Glossary, error) {
	var items []*model.Glossary
	err := r.db.WithContext(ctx).
		Where("source_lang = ? AND target_lang = ?", sourceLang, targetLang).
		Find(&items).Error
	if err != nil {
		return nil, err
	}
	return items, nil
}
