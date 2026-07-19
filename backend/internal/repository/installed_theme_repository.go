package repository

import (
	"context"

	"github.com/yixian-huang/inkless/backend/internal/model"
)

// InstalledThemeRepository defines the interface for installed theme data access
type InstalledThemeRepository interface {
	// List returns all installed themes
	List(ctx context.Context) ([]*model.InstalledTheme, error)

	// FindByThemeID finds an installed theme by its theme ID string
	FindByThemeID(ctx context.Context, themeID string) (*model.InstalledTheme, error)

	// FindActive returns the currently active theme
	FindActive(ctx context.Context) (*model.InstalledTheme, error)

	// SetActive activates a theme by its theme ID and deactivates all others
	SetActive(ctx context.Context, themeID string) error

	// Create creates a new installed theme
	Create(ctx context.Context, theme *model.InstalledTheme) error

	// Update updates an existing installed theme
	Update(ctx context.Context, theme *model.InstalledTheme) error

	// Delete soft-deletes an installed theme by ID
	Delete(ctx context.Context, id uint) error
}
