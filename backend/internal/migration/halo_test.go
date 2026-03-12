package migration

import (
	"context"
	"strings"
	"testing"

	"blotting-consultancy/internal/model"
	"blotting-consultancy/internal/provider"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const sampleHaloJSON = `{
	"version": 1,
	"posts": [
		{
			"id": 1,
			"title": "Getting Started with Halo",
			"slug": "getting-started",
			"originalContent": "# Hello\n\nThis is **markdown** content.",
			"formatContent": "<h1>Hello</h1>\n<p>This is <strong>markdown</strong> content.</p>",
			"summary": "A quick start guide",
			"thumbnail": "https://example.com/thumb.jpg",
			"status": "PUBLISHED",
			"createTime": 1705305000000,
			"updateTime": 1705305000000
		},
		{
			"id": 2,
			"title": "Draft Ideas",
			"slug": "draft-ideas",
			"originalContent": "Some draft content",
			"formatContent": "",
			"summary": "",
			"thumbnail": "",
			"status": "DRAFT",
			"createTime": 1705400000000
		},
		{
			"id": 3,
			"title": "",
			"slug": "",
			"originalContent": "no title or slug",
			"status": "DRAFT",
			"createTime": 0
		}
	],
	"categories": [
		{"id": 10, "name": "Tech", "slug": "tech", "description": "Technology posts"},
		{"id": 20, "name": "Life", "slug": "life", "description": "Life posts"}
	],
	"tags": [
		{"id": 100, "name": "Java", "slug": "java"},
		{"id": 101, "name": "Spring", "slug": "spring"},
		{"id": 102, "name": "Docker", "slug": "docker"}
	],
	"post_categories": [
		{"postId": 1, "categoryId": 10},
		{"postId": 2, "categoryId": 20}
	],
	"post_tags": [
		{"postId": 1, "tagId": 100},
		{"postId": 1, "tagId": 101},
		{"postId": 2, "tagId": 102}
	]
}`

func TestHaloProvider_Source(t *testing.T) {
	p := NewHaloProvider()
	assert.Equal(t, provider.SourceHalo, p.Source())
}

func TestHaloProvider_Parse(t *testing.T) {
	p := NewHaloProvider()
	result, err := p.Parse(context.Background(), strings.NewReader(sampleHaloJSON))

	require.NoError(t, err)
	require.NotNil(t, result)
	require.Len(t, result.Articles, 3)

	// First: published with category and tags
	a1 := result.Articles[0]
	assert.Equal(t, "getting-started", a1.Slug)
	assert.Equal(t, "Getting Started with Halo", a1.Title)
	assert.Equal(t, model.ArticleStatusPublished, a1.Status)
	assert.Equal(t, "Tech", a1.CategoryName)
	assert.ElementsMatch(t, []string{"Java", "Spring"}, a1.TagNames)
	assert.Contains(t, a1.Body, "<h1>Hello</h1>") // prefers formatContent
	assert.Equal(t, "A quick start guide", a1.MetaDescription)
	assert.Equal(t, "https://example.com/thumb.jpg", a1.CoverImageURL)
	assert.NotNil(t, a1.PublishedAt)

	// Second: draft with different category and tag
	a2 := result.Articles[1]
	assert.Equal(t, "draft-ideas", a2.Slug)
	assert.Equal(t, model.ArticleStatusDraft, a2.Status)
	assert.Equal(t, "Life", a2.CategoryName)
	assert.ElementsMatch(t, []string{"Docker"}, a2.TagNames)
	// Falls back to originalContent when formatContent is empty
	assert.Equal(t, "Some draft content", a2.Body)

	// Third: post with no title/slug gets generated slug
	a3 := result.Articles[2]
	assert.NotEmpty(t, a3.Slug)
	assert.Empty(t, a3.CategoryName)
}

func TestHaloProvider_Parse_EmptyExport(t *testing.T) {
	p := NewHaloProvider()
	result, err := p.Parse(context.Background(), strings.NewReader(`{"posts":[],"categories":[],"tags":[],"post_tags":[],"post_categories":[]}`))

	require.NoError(t, err)
	assert.Empty(t, result.Articles)
}

func TestHaloProvider_Parse_InvalidJSON(t *testing.T) {
	p := NewHaloProvider()
	_, err := p.Parse(context.Background(), strings.NewReader("not json"))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "parsing Halo JSON")
}

func TestHaloProvider_Parse_CancelledContext(t *testing.T) {
	p := NewHaloProvider()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := p.Parse(ctx, strings.NewReader(sampleHaloJSON))
	assert.Error(t, err)
}

func TestHaloProvider_Parse_TimestampConversion(t *testing.T) {
	p := NewHaloProvider()
	result, err := p.Parse(context.Background(), strings.NewReader(sampleHaloJSON))
	require.NoError(t, err)

	a1 := result.Articles[0]
	require.NotNil(t, a1.PublishedAt)
	assert.Equal(t, 2024, a1.PublishedAt.Year())
	assert.Equal(t, 1, int(a1.PublishedAt.Month()))

	require.NotNil(t, a1.CreatedAt)
	assert.Equal(t, a1.PublishedAt.Unix(), a1.CreatedAt.Unix())
}
