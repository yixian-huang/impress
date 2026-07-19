package repository

import (
	"context"
	"github.com/yixian-huang/inkless/backend/internal/model"
)

type SiteConfigRepository interface {
	FindByKey(ctx context.Context, key string) (*model.SiteConfig, error)
	Upsert(ctx context.Context, config *model.SiteConfig) error
	Update(ctx context.Context, config *model.SiteConfig) error
	UpdateDraft(ctx context.Context, key string, expectedVersion int, draftConfig model.JSONMap) (int, error)
	UpdatePublished(ctx context.Context, key string, publishedConfig model.JSONMap, publishedVersion int) error
}
