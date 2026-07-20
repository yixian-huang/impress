package repository_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/yixian-huang/inkless/backend/internal/model"
	"github.com/yixian-huang/inkless/backend/internal/repository"
)

func openArticleListDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", strings.ReplaceAll(t.Name(), "/", "_"))
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.Article{}, &model.Category{}, &model.Tag{}))
	return db
}

func TestArticleListFilter_QueryAndSort(t *testing.T) {
	db := openArticleListDB(t)
	repo := repository.NewGormArticleRepository(db)
	ctx := context.Background()

	require.NoError(t, repo.Create(ctx, &model.Article{
		Slug: "alpha-post", ZhTitle: "Alpha 标题", Status: model.ArticleStatusDraft, Author: "alice",
	}))
	require.NoError(t, repo.Create(ctx, &model.Article{
		Slug: "beta-post", ZhTitle: "Beta 文章", Status: model.ArticleStatusPublished, Author: "bob",
	}))

	items, total, err := repo.ListFilter(ctx, repository.ArticleListFilter{
		Offset: 0, Limit: 10, Query: "alpha", Sort: "zh_title ASC",
	})
	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	require.Len(t, items, 1)
	require.Equal(t, "alpha-post", items[0].Slug)

	published, total, err := repo.ListFilter(ctx, repository.ArticleListFilter{
		Offset: 0, Limit: 10, Status: string(model.ArticleStatusPublished), Sort: "created_at DESC",
	})
	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	require.Equal(t, "beta-post", published[0].Slug)
}
