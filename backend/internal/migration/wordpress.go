package migration

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"strings"
	"time"

	"blotting-consultancy/internal/model"
	"blotting-consultancy/internal/provider"
)

// WordPress WXR XML structures
// WXR (WordPress eXtended RSS) is the standard export format.

type wxrRSS struct {
	XMLName xml.Name   `xml:"rss"`
	Channel wxrChannel `xml:"channel"`
}

type wxrChannel struct {
	Items []wxrItem `xml:"item"`
}

type wxrItem struct {
	Title      string        `xml:"title"`
	Link       string        `xml:"link"`
	PubDate    string        `xml:"pubDate"`
	Creator    string        `xml:"http://purl.org/dc/elements/1.1/ creator"`
	Content    string        `xml:"http://purl.org/rss/1.0/modules/content/ encoded"`
	Excerpt    string        `xml:"http://wordpress.org/export/1.2/excerpt/ encoded"`
	PostName   string        `xml:"http://wordpress.org/export/1.2/ post_name"`
	Status     string        `xml:"http://wordpress.org/export/1.2/ status"`
	PostType   string        `xml:"http://wordpress.org/export/1.2/ post_type"`
	Categories []wxrCategory `xml:"category"`
	PostMeta   []wxrPostMeta `xml:"http://wordpress.org/export/1.2/ postmeta"`
}

type wxrCategory struct {
	Domain string `xml:"domain,attr"`
	Slug   string `xml:"nicename,attr"`
	Name   string `xml:",chardata"`
}

type wxrPostMeta struct {
	Key   string `xml:"http://wordpress.org/export/1.2/ meta_key"`
	Value string `xml:"http://wordpress.org/export/1.2/ meta_value"`
}

// WordPressProvider parses WordPress WXR (XML) export files.
type WordPressProvider struct{}

// NewWordPressProvider creates a new WordPress migration provider.
func NewWordPressProvider() *WordPressProvider {
	return &WordPressProvider{}
}

// Source returns the migration source identifier.
func (p *WordPressProvider) Source() provider.MigrationSource {
	return provider.SourceWordPress
}

// Parse reads a WXR XML file and returns intermediate migration articles.
func (p *WordPressProvider) Parse(ctx context.Context, r io.Reader) (*provider.MigrationResult, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("reading WXR data: %w", err)
	}

	// Check for cancellation
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	var rss wxrRSS
	if err := xml.Unmarshal(data, &rss); err != nil {
		return nil, fmt.Errorf("parsing WXR XML: %w", err)
	}

	result := &provider.MigrationResult{}

	// Build attachment map (post_type == "attachment") for media resolution
	attachmentURLs := make(map[string]string) // meta_key "_wp_attached_file" -> URL
	for _, item := range rss.Channel.Items {
		if item.PostType == "attachment" && item.Link != "" {
			attachmentURLs[item.PostName] = item.Link
		}
	}

	for _, item := range rss.Channel.Items {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		// Only import posts (skip pages, attachments, nav_menu_item, etc.)
		if item.PostType != "" && item.PostType != "post" {
			continue
		}

		article, parseErrors := p.convertItem(item)
		result.Errors = append(result.Errors, parseErrors...)

		if article != nil {
			result.Articles = append(result.Articles, article)
		}
	}

	return result, nil
}

func (p *WordPressProvider) convertItem(item wxrItem) (*provider.MigrationArticle, []string) {
	var errs []string

	slug := item.PostName
	if slug == "" {
		slug = sanitizeSlug(item.Title)
	}
	if slug == "" {
		errs = append(errs, fmt.Sprintf("skipping item with empty title and slug"))
		return nil, errs
	}

	// Map WordPress status to our status
	status := model.ArticleStatusDraft
	if item.Status == "publish" {
		status = model.ArticleStatusPublished
	}

	// Parse publication date
	var publishedAt *time.Time
	if item.PubDate != "" {
		// WXR uses RFC1123 format typically
		for _, layout := range []string{
			time.RFC1123Z,
			time.RFC1123,
			"Mon, 02 Jan 2006 15:04:05 -0700",
			"Mon, 02 Jan 2006 15:04:05 +0000",
			"2006-01-02 15:04:05",
		} {
			if t, err := time.Parse(layout, item.PubDate); err == nil {
				publishedAt = &t
				break
			}
		}
		if publishedAt == nil {
			errs = append(errs, fmt.Sprintf("could not parse date %q for %q", item.PubDate, slug))
		}
	}

	// Extract category and tags
	var categoryName string
	var tagNames []string
	for _, cat := range item.Categories {
		switch cat.Domain {
		case "category":
			if categoryName == "" {
				categoryName = cat.Name
			}
		case "post_tag":
			tagNames = append(tagNames, cat.Name)
		}
	}

	// Extract featured image from post meta
	var coverImageURL string
	for _, meta := range item.PostMeta {
		if meta.Key == "_thumbnail_id" || meta.Key == "_wp_attached_file" {
			// _thumbnail_id references an attachment ID; we can't fully resolve
			// without the full WXR, but record the meta value as a hint.
			if meta.Key == "_wp_attached_file" {
				coverImageURL = meta.Value
			}
		}
	}

	// Extract embedded media URLs from content
	mediaURLs := extractMediaURLs(item.Content)

	// Convert WordPress HTML content (already HTML)
	body := item.Content

	// Use excerpt as meta description if available
	metaDesc := strings.TrimSpace(item.Excerpt)

	article := &provider.MigrationArticle{
		Slug:            slug,
		Title:           item.Title,
		Body:            body,
		Status:          status,
		CategoryName:    categoryName,
		TagNames:        tagNames,
		CoverImageURL:   coverImageURL,
		MediaURLs:       mediaURLs,
		PublishedAt:     publishedAt,
		MetaDescription: metaDesc,
	}

	return article, errs
}

// extractMediaURLs finds image/media URLs in HTML content.
func extractMediaURLs(html string) []string {
	var urls []string
	// Simple extraction: find src="..." in img tags
	remaining := html
	for {
		idx := strings.Index(remaining, `src="`)
		if idx == -1 {
			break
		}
		remaining = remaining[idx+5:]
		end := strings.Index(remaining, `"`)
		if end == -1 {
			break
		}
		url := remaining[:end]
		if strings.HasPrefix(url, "http") {
			urls = append(urls, url)
		}
		remaining = remaining[end:]
	}
	return urls
}

// sanitizeSlug creates a URL-safe slug from a title.
func sanitizeSlug(title string) string {
	slug := strings.ToLower(title)
	var result strings.Builder
	for _, r := range slug {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			result.WriteRune(r)
		} else if r == ' ' || r == '_' {
			result.WriteRune('-')
		}
		// skip other characters (including CJK — those get a hash-based slug)
	}
	s := result.String()
	// Collapse multiple dashes
	for strings.Contains(s, "--") {
		s = strings.ReplaceAll(s, "--", "-")
	}
	s = strings.Trim(s, "-")
	if s == "" {
		s = fmt.Sprintf("post-%d", time.Now().UnixNano())
	}
	if len(s) > 200 {
		s = s[:200]
	}
	return s
}
