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

func TestMediaListFilter_QueryTypeAndFolder(t *testing.T) {
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", strings.ReplaceAll(t.Name(), "/", "_"))
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.Media{}))
	repo := repository.NewGormMediaRepository(db)
	ctx := context.Background()

	folder := uint(3)
	require.NoError(t, repo.Create(ctx, &model.Media{
		URL: "/u/a.png", Filename: "hero.png", MimeType: "image/png", Size: 10, FolderID: &folder,
	}))
	require.NoError(t, repo.Create(ctx, &model.Media{
		URL: "/u/b.mp4", Filename: "clip.mp4", MimeType: "video/mp4", Size: 20,
	}))

	images, total, err := repo.ListFilter(ctx, repository.MediaListFilter{
		Offset: 0, Limit: 10, MimePrefix: "image/", Sort: "filename ASC",
	})
	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	require.Equal(t, "hero.png", images[0].Filename)

	byName, total, err := repo.ListFilter(ctx, repository.MediaListFilter{
		Offset: 0, Limit: 10, Query: "clip", Sort: "created_at DESC",
	})
	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	require.Equal(t, "clip.mp4", byName[0].Filename)

	root, total, err := repo.ListFilter(ctx, repository.MediaListFilter{
		Offset: 0, Limit: 10, FolderRoot: true, Sort: "created_at DESC",
	})
	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	require.Equal(t, "clip.mp4", root[0].Filename)

	inFolder, total, err := repo.ListFilter(ctx, repository.MediaListFilter{
		Offset: 0, Limit: 10, FolderID: &folder, Sort: "created_at DESC",
	})
	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	require.Equal(t, "hero.png", inFolder[0].Filename)
}
