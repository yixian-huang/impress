package repository

import (
	"context"

	"github.com/yixian-huang/inkless/backend/internal/model"
)

// MediaFolderRepository defines the interface for media folder data access
type MediaFolderRepository interface {
	// Create creates a new media folder
	Create(ctx context.Context, folder *model.MediaFolder) error

	// FindByID finds a media folder by ID
	FindByID(ctx context.Context, id uint) (*model.MediaFolder, error)

	// ListTree returns all folders as a flat list (caller builds tree)
	ListAll(ctx context.Context) ([]*model.MediaFolder, error)

	// Update updates a media folder
	Update(ctx context.Context, folder *model.MediaFolder) error

	// Delete deletes a media folder by ID
	Delete(ctx context.Context, id uint) error

	// FindChildren returns direct children of a folder
	FindChildren(ctx context.Context, parentID uint) ([]*model.MediaFolder, error)

	// UpdateMediaFolder updates the folder_id of a media item
	UpdateMediaFolder(ctx context.Context, mediaID uint, folderID *uint) error
}
