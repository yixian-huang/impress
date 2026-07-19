package repository

import (
	"context"
	"errors"

	"github.com/yixian-huang/inkless/backend/internal/model"

	"gorm.io/gorm"
)

// GormInstalledThemeRepository implements InstalledThemeRepository using GORM
type GormInstalledThemeRepository struct {
	db *gorm.DB
}

// NewGormInstalledThemeRepository creates a new GormInstalledThemeRepository
func NewGormInstalledThemeRepository(db *gorm.DB) InstalledThemeRepository {
	return &GormInstalledThemeRepository{db: db}
}

// List returns all installed themes
func (r *GormInstalledThemeRepository) List(ctx context.Context) ([]*model.InstalledTheme, error) {
	var themes []*model.InstalledTheme
	err := r.db.WithContext(ctx).
		Order("created_at DESC").
		Find(&themes).Error
	if err != nil {
		return nil, err
	}
	return themes, nil
}

// FindByThemeID finds an installed theme by its theme ID string
func (r *GormInstalledThemeRepository) FindByThemeID(ctx context.Context, themeID string) (*model.InstalledTheme, error) {
	var theme model.InstalledTheme
	err := r.db.WithContext(ctx).
		Where("theme_id = ?", themeID).
		First(&theme).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("theme not found")
		}
		return nil, err
	}
	return &theme, nil
}

// FindActive returns the currently active theme
func (r *GormInstalledThemeRepository) FindActive(ctx context.Context) (*model.InstalledTheme, error) {
	var theme model.InstalledTheme
	err := r.db.WithContext(ctx).
		Where("is_active = ?", true).
		First(&theme).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("no active theme found")
		}
		return nil, err
	}
	return &theme, nil
}

// SetActive activates a theme by its theme ID and deactivates all others.
// Uses a transaction to ensure atomicity.
func (r *GormInstalledThemeRepository) SetActive(ctx context.Context, themeID string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Deactivate all themes
		if err := tx.Model(&model.InstalledTheme{}).
			Where("is_active = ?", true).
			Update("is_active", false).Error; err != nil {
			return err
		}

		// Activate the target theme
		result := tx.Model(&model.InstalledTheme{}).
			Where("theme_id = ?", themeID).
			Update("is_active", true)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return errors.New("theme not found")
		}

		return nil
	})
}

// Create creates a new installed theme
func (r *GormInstalledThemeRepository) Create(ctx context.Context, theme *model.InstalledTheme) error {
	return r.db.WithContext(ctx).Create(theme).Error
}

// Update updates an existing installed theme
func (r *GormInstalledThemeRepository) Update(ctx context.Context, theme *model.InstalledTheme) error {
	return r.db.WithContext(ctx).Save(theme).Error
}

// Delete soft-deletes an installed theme by ID
func (r *GormInstalledThemeRepository) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&model.InstalledTheme{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("theme not found")
	}
	return nil
}
