package seed

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"blotting-consultancy/internal/model"
	"blotting-consultancy/internal/repository"
)

const sampleArticleCount = 48

type blogSampleCategory struct {
	Slug        string
	ZhName      string
	EnName      string
	ZhDesc      string
}

type blogSampleTag struct {
	Slug   string
	ZhName string
	EnName string
}

var blogSampleCategories = []blogSampleCategory{
	{Slug: "tech-notes", ZhName: "技术笔记", EnName: "Tech Notes", ZhDesc: "开发与实践记录"},
	{Slug: "life", ZhName: "生活随想", EnName: "Life", ZhDesc: "日常观察与随笔"},
	{Slug: "reading", ZhName: "读书笔记", EnName: "Reading", ZhDesc: "阅读摘录与感想"},
	{Slug: "projects", ZhName: "项目实践", EnName: "Projects", ZhDesc: "项目复盘与方案"},
}

var blogSampleTags = []blogSampleTag{
	{Slug: "go", ZhName: "Go", EnName: "Go"},
	{Slug: "react", ZhName: "React", EnName: "React"},
	{Slug: "cms", ZhName: "CMS", EnName: "CMS"},
	{Slug: "design", ZhName: "设计", EnName: "Design"},
	{Slug: "travel", ZhName: "旅行", EnName: "Travel"},
	{Slug: "book", ZhName: "读书", EnName: "Books"},
	{Slug: "writing", ZhName: "写作", EnName: "Writing"},
	{Slug: "ops", ZhName: "运维", EnName: "Ops"},
}

// SeedBlogSamples inserts idempotent sample categories, tags, and published articles for UI testing.
func SeedBlogSamples(
	ctx context.Context,
	articleRepo repository.ArticleRepository,
	categoryRepo repository.CategoryRepository,
	tagRepo repository.TagRepository,
) error {
	log.Println("Seeding blog sample data...")

	categories, err := ensureBlogSampleCategories(ctx, categoryRepo)
	if err != nil {
		return err
	}
	tags, err := ensureBlogSampleTags(ctx, tagRepo)
	if err != nil {
		return err
	}

	created := 0
	skipped := 0
	now := time.Now()

	for i := 1; i <= sampleArticleCount; i++ {
		slug := fmt.Sprintf("sample-post-%02d", i)
		if _, err := articleRepo.FindBySlug(ctx, slug); err == nil {
			skipped++
			continue
		} else if err != nil && !strings.Contains(err.Error(), "not found") {
			return fmt.Errorf("lookup article %s: %w", slug, err)
		}

		cat := categories[i%len(categories)]
		tagA := tags[i%len(tags)]
		tagB := tags[(i+3)%len(tags)]
		published := now.AddDate(0, 0, -i)
		titleZh := fmt.Sprintf("示例文章 %d：%s 中的一则思考", i, cat.ZhName)
		titleEn := fmt.Sprintf("Sample Post %d: Thoughts on %s", i, cat.EnName)
		bodyZh := sampleBodyZh(i, cat.ZhName)
		bodyEn := sampleBodyEn(i, cat.EnName)

		article := &model.Article{
			Slug:          slug,
			Status:        model.ArticleStatusPublished,
			ZhTitle:       titleZh,
			EnTitle:       titleEn,
			ZhBody:        bodyZh,
			EnBody:        bodyEn,
			CategoryID:    &cat.ID,
			Category:      cat,
			Tags:          []model.Tag{tagA, tagB},
			AllowComments: true,
			Visibility:    "public",
			PublishedAt:   &published,
		}

		if err := articleRepo.Create(ctx, article); err != nil {
			return fmt.Errorf("create article %s: %w", slug, err)
		}
		created++
	}

	log.Printf("Blog sample seed done: created=%d skipped=%d (total target=%d)", created, skipped, sampleArticleCount)
	return nil
}

func ensureBlogSampleCategories(ctx context.Context, repo repository.CategoryRepository) ([]*model.Category, error) {
	out := make([]*model.Category, 0, len(blogSampleCategories))
	for _, cfg := range blogSampleCategories {
		existing, err := repo.FindBySlug(ctx, cfg.Slug)
		if err == nil {
			out = append(out, existing)
			continue
		}
		if err != nil && !strings.Contains(err.Error(), "not found") {
			return nil, err
		}
		c := &model.Category{
			Slug:          cfg.Slug,
			ZhName:        cfg.ZhName,
			EnName:        cfg.EnName,
			ZhDescription: cfg.ZhDesc,
		}
		if err := repo.Create(ctx, c); err != nil {
			return nil, fmt.Errorf("create category %s: %w", cfg.Slug, err)
		}
		out = append(out, c)
	}
	return out, nil
}

func ensureBlogSampleTags(ctx context.Context, repo repository.TagRepository) ([]model.Tag, error) {
	out := make([]model.Tag, 0, len(blogSampleTags))
	for _, cfg := range blogSampleTags {
		existing, err := repo.FindBySlug(ctx, cfg.Slug)
		if err == nil {
			out = append(out, *existing)
			continue
		}
		if err != nil && !strings.Contains(err.Error(), "not found") {
			return nil, err
		}
		t := &model.Tag{
			Slug:   cfg.Slug,
			ZhName: cfg.ZhName,
			EnName: cfg.EnName,
		}
		if err := repo.Create(ctx, t); err != nil {
			return nil, fmt.Errorf("create tag %s: %w", cfg.Slug, err)
		}
		out = append(out, *t)
	}
	return out, nil
}

func sampleBodyZh(n int, category string) string {
	return fmt.Sprintf(`<p>这是第 %d 篇示例文章，分类为「%s」。用于测试博客列表分页、筛选与阅读排版。</p>
<h2>背景</h2>
<p>当文章数量增多时，列表行的标题与日期对齐、归档筛选区的扁平样式，以及长文目录行为都需要在真实数据下验证。</p>
<h2>要点</h2>
<p>本文包含若干段落，便于检查行高、字号与段间距。若主题启用了目录，多级标题也会参与 TOC 模式切换。</p>
<p>继续补充一段文字，让字数更接近日常发布的随笔长度，方便观察评论区的排版与层级缩进效果。</p>`, n, category)
}

func sampleBodyEn(n int, category string) string {
	return fmt.Sprintf(`<p>This is sample article #%d in category "%s", for testing archive pagination and reading layout.</p>
<h2>Context</h2>
<p>With more posts, list density, filters, and TOC behavior are easier to evaluate under realistic content volume.</p>
<h2>Notes</h2>
<p>Extra paragraphs help verify typography rhythm and comment thread styling on article pages.</p>`, n, category)
}
