package repository

import (
	"context"

	"github.com/yixian-huang/inkless/backend/internal/model"
)

// StorageConfigRepository defines the interface for storage config data access
type StorageConfigRepository interface {
	// Get returns the current storage config (singleton row)
	Get(ctx context.Context) (*model.StorageConfig, error)

	// Upsert creates or updates the storage config
	Upsert(ctx context.Context, config *model.StorageConfig) error
}
