package repository

import (
	"context"

	"github.com/yixian-huang/inkless/backend/internal/model"
)

// ChunkedUploadRepository defines the interface for chunked upload data access
type ChunkedUploadRepository interface {
	// Create creates a new chunked upload record
	Create(ctx context.Context, upload *model.ChunkedUpload) error

	// FindByID finds a chunked upload by ID
	FindByID(ctx context.Context, id string) (*model.ChunkedUpload, error)

	// Update updates a chunked upload record
	Update(ctx context.Context, upload *model.ChunkedUpload) error

	// Delete deletes a chunked upload record
	Delete(ctx context.Context, id string) error
}
