package migration

import (
	"archive/zip"
	"bytes"
	"context"
	"testing"

	"github.com/yixian-huang/inkless/backend/internal/model"
	"github.com/yixian-huang/inkless/backend/internal/provider"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createTestZip builds an in-memory ZIP archive from a map of filename -> content.
func createTestZip(t *testing.T, files map[string]string) *bytes.Reader {
	t.Helper()
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)
	for name, content := range files {
		f, err := w.Create(name)
		require.NoError(t, err)
		_, err = f.Write([]byte(content))
		require.NoError(t, err)
	}
	require.NoError(t, w.Close())
	return bytes.NewReader(buf.Bytes())
}

func TestMarkdownProvider_Source(t *testing.T) {
	p := NewMarkdownProvider()
	assert.Equal(t, provider.SourceMarkdown, p.Source())
}

func TestMarkdownProvider_Parse_WithFrontMatter(t *testing.T) {
	p := NewMarkdownProvider()

	files := map[string]string{
		"posts/hello-world.md": `---
title: "Hello World"
slug: hello-world
date: 2024-01-15
status: published
category: Tech
tags: [go, web]
cover: https://example.com/cover.jpg
description: "A short summary"
---
# Hello World

This is my **first** post.
`,
	}

	zr := createTestZip(t, files)
	result, err := p.Parse(context.Background(), zr)

	require.NoError(t, err)
	require.Len(t, result.Articles, 1)

	a := result.Articles[0]
	assert.Equal(t, "hello-world", a.Slug)
	assert.Equal(t, "Hello World", a.Title)
	assert.Equal(t, model.ArticleStatusPublished, a.Status)
	assert.Equal(t, "Tech", a.CategoryName)
	assert.ElementsMatch(t, []string{"go", "web"}, a.TagNames)
	assert.Equal(t, "https://example.com/cover.jpg", a.CoverImageURL)
	assert.Equal(t, "A short summary", a.MetaDescription)
	assert.Contains(t, a.Body, "# Hello World")
	assert.NotNil(t, a.PublishedAt)
	assert.Equal(t, 2024, a.PublishedAt.Year())
	assert.Equal(t, 1, int(a.PublishedAt.Month()))
	assert.Equal(t, 15, a.PublishedAt.Day())
}

func TestMarkdownProvider_Parse_WithoutFrontMatter(t *testing.T) {
	p := NewMarkdownProvider()

	files := map[string]string{
		"my-post.md": `# Just Content

No front matter here.
`,
	}

	zr := createTestZip(t, files)
	result, err := p.Parse(context.Background(), zr)

	require.NoError(t, err)
	require.Len(t, result.Articles, 1)

	a := result.Articles[0]
	assert.Equal(t, "my-post", a.Slug)  // derived from filename
	assert.Equal(t, "my-post", a.Title) // derived from filename
	assert.Equal(t, model.ArticleStatusDraft, a.Status)
	assert.Contains(t, a.Body, "# Just Content")
}

func TestMarkdownProvider_Parse_MultipleFiles(t *testing.T) {
	p := NewMarkdownProvider()

	files := map[string]string{
		"post1.md": `---
title: First
---
Content 1`,
		"post2.md": `---
title: Second
status: published
---
Content 2`,
		"readme.txt":   "This is not markdown",
		"images/a.png": "binary data",
	}

	zr := createTestZip(t, files)
	result, err := p.Parse(context.Background(), zr)

	require.NoError(t, err)
	// Should only import .md files
	assert.Len(t, result.Articles, 2)
}

func TestMarkdownProvider_Parse_YAMLListTags(t *testing.T) {
	p := NewMarkdownProvider()

	files := map[string]string{
		"post.md": `---
title: Tagged Post
tags:
  - alpha
  - beta
  - gamma
---
Body`,
	}

	zr := createTestZip(t, files)
	result, err := p.Parse(context.Background(), zr)

	require.NoError(t, err)
	require.Len(t, result.Articles, 1)
	assert.ElementsMatch(t, []string{"alpha", "beta", "gamma"}, result.Articles[0].TagNames)
}

func TestMarkdownProvider_Parse_EmptyZip(t *testing.T) {
	p := NewMarkdownProvider()
	zr := createTestZip(t, map[string]string{})
	result, err := p.Parse(context.Background(), zr)

	require.NoError(t, err)
	assert.Empty(t, result.Articles)
}

func TestMarkdownProvider_Parse_InvalidZip(t *testing.T) {
	p := NewMarkdownProvider()
	_, err := p.Parse(context.Background(), bytes.NewReader([]byte("not a zip")))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "opening zip archive")
}

func TestMarkdownProvider_Parse_CancelledContext(t *testing.T) {
	p := NewMarkdownProvider()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	zr := createTestZip(t, map[string]string{"a.md": "content"})
	_, err := p.Parse(ctx, zr)
	assert.Error(t, err)
}

func TestSplitFrontMatter(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantFM   bool
		wantBody string
	}{
		{
			name:     "with front matter",
			input:    "---\ntitle: Hello\n---\nBody content",
			wantFM:   true,
			wantBody: "\nBody content",
		},
		{
			name:     "without front matter",
			input:    "Just body content",
			wantFM:   false,
			wantBody: "Just body content",
		},
		{
			name:     "empty",
			input:    "",
			wantFM:   false,
			wantBody: "",
		},
		{
			name:     "no closing delimiter",
			input:    "---\ntitle: Hello\nno closing",
			wantFM:   false,
			wantBody: "---\ntitle: Hello\nno closing",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fm, body := splitFrontMatter([]byte(tt.input))
			if tt.wantFM {
				assert.NotEmpty(t, fm)
			} else {
				assert.Empty(t, fm)
			}
			assert.Equal(t, tt.wantBody, string(body))
		})
	}
}

func TestParseFrontMatterMap(t *testing.T) {
	input := `title: "My Post"
slug: my-post
tags: [go, web, api]
category: Tech`

	m := parseFrontMatterMap([]byte(input))

	assert.Equal(t, "My Post", m["title"])
	assert.Equal(t, "my-post", m["slug"])
	assert.Equal(t, "Tech", m["category"])

	tags, ok := m["tags"].([]string)
	require.True(t, ok)
	assert.ElementsMatch(t, []string{"go", "web", "api"}, tags)
}

func TestParseFrontMatterMap_ListStyle(t *testing.T) {
	input := `title: Test
tags:
  - alpha
  - beta`

	m := parseFrontMatterMap([]byte(input))
	assert.Equal(t, "Test", m["title"])

	tags, ok := m["tags"].([]string)
	require.True(t, ok)
	assert.ElementsMatch(t, []string{"alpha", "beta"}, tags)
}

func TestFmString(t *testing.T) {
	m := map[string]interface{}{
		"title": "Hello",
		"tags":  []string{"a", "b"},
	}
	assert.Equal(t, "Hello", fmString(m, "title"))
	assert.Equal(t, "", fmString(m, "missing"))
	assert.Equal(t, "", fmString(m, "tags")) // not a string
}

func TestFmStringSlice(t *testing.T) {
	m := map[string]interface{}{
		"tags":     []string{"a", "b"},
		"category": "Tech",
	}
	assert.ElementsMatch(t, []string{"a", "b"}, fmStringSlice(m, "tags"))
	assert.ElementsMatch(t, []string{"Tech"}, fmStringSlice(m, "category"))
	assert.Nil(t, fmStringSlice(m, "missing"))
}
