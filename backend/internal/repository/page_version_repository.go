package repository

import (
	"context"
	"github.com/yixian-huang/inkless/backend/internal/model"
)

type PageVersionRepository interface {
	Create(ctx context.Context, version *model.PageVersion) error
	FindByPageIDAndVersion(ctx context.Context, pageID uint, version int) (*model.PageVersion, error)
	ListByPageID(ctx context.Context, pageID uint, offset, limit int) ([]*model.PageVersion, int64, error)
	GetLatestVersion(ctx context.Context, pageID uint) (int, error)
	Delete(ctx context.Context, id uint) error
}
