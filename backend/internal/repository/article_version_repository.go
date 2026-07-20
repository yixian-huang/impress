package repository

import (
	"context"

	"github.com/yixian-huang/inkless/backend/internal/model"
)

// ArticleVersionRepository defines data access for article version history.
type ArticleVersionRepository interface {
	Create(ctx context.Context, version *model.ArticleVersion) error
	FindByArticleIDAndVersion(ctx context.Context, articleID uint, version int) (*model.ArticleVersion, error)
	ListByArticleID(ctx context.Context, articleID uint, offset, limit int) ([]*model.ArticleVersion, int64, error)
	GetLatestVersion(ctx context.Context, articleID uint) (int, error)
}
