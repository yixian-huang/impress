package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/yixian-huang/inkless/backend/internal/model"
	"github.com/yixian-huang/inkless/backend/internal/repository"
)

func TestListPublishedSitemapMetaOmitsBodies(t *testing.T) {
	db := openArticleListDB(t)
	repo := repository.NewGormArticleRepository(db)
	ctx := context.Background()

	big := make([]byte, 8000)
	for i := range big {
		big[i] = 'x'
	}
	zhBody := "<p>" + string(big) + "</p>"

	now := time.Now().UTC()
	a := &model.Article{
		Slug:        "sitemap-lean-post",
		Status:      model.ArticleStatusPublished,
		ZhTitle:     "站点地图文章",
		ZhBody:      zhBody,
		EnBody:      zhBody,
		Visibility:  "public",
		PublishedAt: &now,
	}
	if err := repo.Create(ctx, a); err != nil {
		t.Fatalf("create: %v", err)
	}

	meta, err := repo.ListPublishedSitemapMeta(ctx, 100)
	if err != nil {
		t.Fatalf("ListPublishedSitemapMeta: %v", err)
	}
	if len(meta) == 0 {
		t.Fatal("expected at least one meta row")
	}
	found := false
	for _, m := range meta {
		if m.Slug == "sitemap-lean-post" {
			found = true
			if m.UpdatedAt.IsZero() {
				t.Fatal("expected UpdatedAt set")
			}
		}
	}
	if !found {
		t.Fatal("expected sitemap-lean-post in meta results")
	}

	// Contrast: full ListPublished materializes bodies for the same slug.
	items, _, err := repo.ListPublished(ctx, 0, 10, "", "")
	if err != nil {
		t.Fatalf("ListPublished: %v", err)
	}
	var full *model.Article
	for _, it := range items {
		if it.Slug == "sitemap-lean-post" {
			full = it
			break
		}
	}
	if full == nil {
		t.Fatal("ListPublished missing article")
	}
	if len(full.ZhBody) < 1000 {
		t.Fatalf("expected full list to load body, got len=%d", len(full.ZhBody))
	}
}
