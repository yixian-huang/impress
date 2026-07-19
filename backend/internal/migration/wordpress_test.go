package migration

import (
	"context"
	"strings"
	"testing"

	"github.com/yixian-huang/inkless/backend/internal/model"
	"github.com/yixian-huang/inkless/backend/internal/provider"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const sampleWXR = `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0"
	xmlns:content="http://purl.org/rss/1.0/modules/content/"
	xmlns:dc="http://purl.org/dc/elements/1.1/"
	xmlns:wp="http://wordpress.org/export/1.2/"
	xmlns:excerpt="http://wordpress.org/export/1.2/excerpt/"
>
<channel>
	<item>
		<title>Hello World</title>
		<wp:post_name>hello-world</wp:post_name>
		<wp:status>publish</wp:status>
		<wp:post_type>post</wp:post_type>
		<pubDate>Mon, 15 Jan 2024 10:30:00 +0000</pubDate>
		<content:encoded><![CDATA[<p>This is my first post with an <img src="https://example.com/img.jpg" /> image.</p>]]></content:encoded>
		<excerpt:encoded><![CDATA[A short excerpt]]></excerpt:encoded>
		<category domain="category" nicename="tech">Technology</category>
		<category domain="post_tag" nicename="go">Go</category>
		<category domain="post_tag" nicename="web">Web</category>
	</item>
	<item>
		<title>Draft Post</title>
		<wp:post_name>draft-post</wp:post_name>
		<wp:status>draft</wp:status>
		<wp:post_type>post</wp:post_type>
		<content:encoded><![CDATA[<p>Still working on this.</p>]]></content:encoded>
	</item>
	<item>
		<title>About Page</title>
		<wp:post_name>about</wp:post_name>
		<wp:status>publish</wp:status>
		<wp:post_type>page</wp:post_type>
		<content:encoded><![CDATA[<p>About us.</p>]]></content:encoded>
	</item>
	<item>
		<title>photo.jpg</title>
		<wp:post_name>photo-jpg</wp:post_name>
		<wp:post_type>attachment</wp:post_type>
		<link>https://example.com/wp-content/uploads/photo.jpg</link>
	</item>
</channel>
</rss>`

func TestWordPressProvider_Source(t *testing.T) {
	p := NewWordPressProvider()
	assert.Equal(t, provider.SourceWordPress, p.Source())
}

func TestWordPressProvider_Parse(t *testing.T) {
	p := NewWordPressProvider()
	result, err := p.Parse(context.Background(), strings.NewReader(sampleWXR))

	require.NoError(t, err)
	require.NotNil(t, result)

	// Should import 2 posts (skip page and attachment)
	require.Len(t, result.Articles, 2)

	// First article: published post
	a1 := result.Articles[0]
	assert.Equal(t, "hello-world", a1.Slug)
	assert.Equal(t, "Hello World", a1.Title)
	assert.Equal(t, model.ArticleStatusPublished, a1.Status)
	assert.Equal(t, "Technology", a1.CategoryName)
	assert.ElementsMatch(t, []string{"Go", "Web"}, a1.TagNames)
	assert.Contains(t, a1.Body, "first post")
	assert.Equal(t, "A short excerpt", a1.MetaDescription)
	assert.NotNil(t, a1.PublishedAt)
	assert.Equal(t, 2024, a1.PublishedAt.Year())

	// Media URLs extracted from content
	assert.Contains(t, a1.MediaURLs, "https://example.com/img.jpg")

	// Second article: draft post
	a2 := result.Articles[1]
	assert.Equal(t, "draft-post", a2.Slug)
	assert.Equal(t, model.ArticleStatusDraft, a2.Status)
	assert.Empty(t, a2.CategoryName)
	assert.Empty(t, a2.TagNames)
}

func TestWordPressProvider_Parse_EmptyDocument(t *testing.T) {
	p := NewWordPressProvider()
	wxr := `<?xml version="1.0"?><rss><channel></channel></rss>`
	result, err := p.Parse(context.Background(), strings.NewReader(wxr))

	require.NoError(t, err)
	assert.Empty(t, result.Articles)
}

func TestWordPressProvider_Parse_InvalidXML(t *testing.T) {
	p := NewWordPressProvider()
	_, err := p.Parse(context.Background(), strings.NewReader("not xml at all"))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "parsing WXR XML")
}

func TestWordPressProvider_Parse_CancelledContext(t *testing.T) {
	p := NewWordPressProvider()
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	_, err := p.Parse(ctx, strings.NewReader(sampleWXR))
	assert.Error(t, err)
}

func TestSanitizeSlug(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"Hello World", "hello-world"},
		{"My  Post!", "my-post"},
		{"a-b-c", "a-b-c"},
		{"CamelCase", "camelcase"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.want, sanitizeSlug(tt.input))
		})
	}

	assert.Equal(t, sanitizeSlug("中文分类"), sanitizeSlug(" 中文分类 "))
	assert.NotEqual(t, sanitizeSlug("中文分类"), sanitizeSlug("另一个分类"))
	assert.Regexp(t, `^post-[0-9a-f]{12}$`, sanitizeSlug("中文分类"))
	assert.Equal(t, sanitizeSlug("!!!"), sanitizeSlug("!!!"))
}

func TestExtractMediaURLs(t *testing.T) {
	html := `<p>Hello <img src="https://example.com/a.jpg" /> world <img src="https://cdn.test/b.png" /></p>`
	urls := extractMediaURLs(html)
	assert.Len(t, urls, 2)
	assert.Contains(t, urls, "https://example.com/a.jpg")
	assert.Contains(t, urls, "https://cdn.test/b.png")
}

func TestExtractMediaURLs_NoImages(t *testing.T) {
	urls := extractMediaURLs("<p>No images here</p>")
	assert.Empty(t, urls)
}

func TestExtractMediaURLs_RelativeURLsIgnored(t *testing.T) {
	urls := extractMediaURLs(`<img src="/local/img.jpg" />`)
	assert.Empty(t, urls)
}
