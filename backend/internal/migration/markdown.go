package migration

import (
	"archive/zip"
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/yixian-huang/inkless/backend/internal/model"
	"github.com/yixian-huang/inkless/backend/internal/provider"
)

// MarkdownProvider parses a ZIP archive of Markdown files with YAML
// front-matter. Each .md file becomes one article.
//
// Expected front-matter keys (all optional):
//
//	---
//	title: "My Post"
//	slug: my-post
//	date: 2024-01-15
//	status: published
//	category: Tech
//	tags: [go, web]
//	cover: https://example.com/cover.jpg
//	description: "A short summary"
//	---
type MarkdownProvider struct{}

// NewMarkdownProvider creates a new Markdown batch migration provider.
func NewMarkdownProvider() *MarkdownProvider {
	return &MarkdownProvider{}
}

// Source returns the migration source identifier.
func (p *MarkdownProvider) Source() provider.MigrationSource {
	return provider.SourceMarkdown
}

// Parse reads a ZIP archive of .md files from r and returns articles.
func (p *MarkdownProvider) Parse(ctx context.Context, r io.Reader) (*provider.MigrationResult, error) {
	// Read all data into memory so we can use zip.NewReader
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("reading markdown archive: %w", err)
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	zipReader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, fmt.Errorf("opening zip archive: %w", err)
	}

	result := &provider.MigrationResult{}

	for _, f := range zipReader.File {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		// Skip directories and non-markdown files
		if f.FileInfo().IsDir() {
			continue
		}
		ext := strings.ToLower(filepath.Ext(f.Name))
		if ext != ".md" && ext != ".markdown" {
			continue
		}

		rc, err := f.Open()
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("failed to open %s: %v", f.Name, err))
			continue
		}
		content, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("failed to read %s: %v", f.Name, err))
			continue
		}

		article, parseErrors := p.parseMarkdownFile(f.Name, content)
		result.Errors = append(result.Errors, parseErrors...)
		if article != nil {
			result.Articles = append(result.Articles, article)
		}
	}

	return result, nil
}

func (p *MarkdownProvider) parseMarkdownFile(filename string, content []byte) (*provider.MigrationArticle, []string) {
	var errs []string

	frontMatter, body := splitFrontMatter(content)

	// Derive defaults from filename
	baseName := strings.TrimSuffix(filepath.Base(filename), filepath.Ext(filename))
	defaultSlug := sanitizeSlug(baseName)

	// Parse front-matter
	fm := parseFrontMatterMap(frontMatter)

	title := fmString(fm, "title")
	if title == "" {
		title = baseName
	}

	slug := fmString(fm, "slug")
	if slug == "" {
		slug = defaultSlug
	}

	status := model.ArticleStatusDraft
	if s := fmString(fm, "status"); s == "published" || s == "publish" {
		status = model.ArticleStatusPublished
	}

	var publishedAt *time.Time
	if dateStr := fmString(fm, "date"); dateStr != "" {
		for _, layout := range []string{
			"2006-01-02",
			"2006-01-02 15:04:05",
			time.RFC3339,
		} {
			if t, err := time.Parse(layout, dateStr); err == nil {
				publishedAt = &t
				break
			}
		}
		if publishedAt == nil {
			errs = append(errs, fmt.Sprintf("could not parse date %q in %s", dateStr, filename))
		}
	}

	categoryName := fmString(fm, "category")
	tagNames := fmStringSlice(fm, "tags")
	coverImage := fmString(fm, "cover")
	description := fmString(fm, "description")

	// Extract media URLs from body
	mediaURLs := extractMediaURLs(string(body))

	article := &provider.MigrationArticle{
		Slug:            slug,
		Title:           title,
		Body:            strings.TrimSpace(string(body)),
		Status:          status,
		CategoryName:    categoryName,
		TagNames:        tagNames,
		CoverImageURL:   coverImage,
		MediaURLs:       mediaURLs,
		PublishedAt:     publishedAt,
		MetaDescription: description,
	}

	return article, errs
}

// splitFrontMatter splits YAML front-matter (delimited by ---) from body.
func splitFrontMatter(content []byte) (frontMatter []byte, body []byte) {
	s := string(content)
	trimmed := strings.TrimSpace(s)
	if !strings.HasPrefix(trimmed, "---") {
		return nil, content
	}

	// Find the closing ---
	rest := trimmed[3:]
	idx := strings.Index(rest, "\n---")
	if idx == -1 {
		return nil, content
	}

	fm := rest[:idx]
	bodyStart := rest[idx+4:]

	return []byte(fm), []byte(bodyStart)
}

// parseFrontMatterMap does a simple line-by-line YAML parse for flat key:value
// pairs and simple lists. This avoids importing a full YAML library.
func parseFrontMatterMap(data []byte) map[string]interface{} {
	m := make(map[string]interface{})
	if len(data) == 0 {
		return m
	}

	scanner := bufio.NewScanner(bytes.NewReader(data))
	var currentKey string

	for scanner.Scan() {
		line := scanner.Text()

		// List item continuation (e.g., "  - item")
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "- ") && currentKey != "" {
			item := strings.TrimSpace(strings.TrimPrefix(trimmed, "- "))
			item = strings.Trim(item, `"'`)
			if existing, ok := m[currentKey]; ok {
				if arr, ok := existing.([]string); ok {
					m[currentKey] = append(arr, item)
				}
			} else {
				m[currentKey] = []string{item}
			}
			continue
		}

		// Key: value pair
		colonIdx := strings.Index(line, ":")
		if colonIdx == -1 {
			continue
		}
		key := strings.TrimSpace(line[:colonIdx])
		value := strings.TrimSpace(line[colonIdx+1:])

		if key == "" {
			continue
		}
		currentKey = key

		// Handle inline array: [item1, item2]
		if strings.HasPrefix(value, "[") && strings.HasSuffix(value, "]") {
			inner := value[1 : len(value)-1]
			var items []string
			for _, part := range strings.Split(inner, ",") {
				item := strings.TrimSpace(part)
				item = strings.Trim(item, `"'`)
				if item != "" {
					items = append(items, item)
				}
			}
			m[key] = items
			continue
		}

		// Strip surrounding quotes
		value = strings.Trim(value, `"'`)

		if value == "" {
			// Could be followed by list items
			continue
		}
		m[key] = value
	}

	return m
}

// fmString extracts a string value from the front-matter map.
func fmString(m map[string]interface{}, key string) string {
	v, ok := m[key]
	if !ok {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

// fmStringSlice extracts a string slice from the front-matter map.
func fmStringSlice(m map[string]interface{}, key string) []string {
	v, ok := m[key]
	if !ok {
		return nil
	}
	if arr, ok := v.([]string); ok {
		return arr
	}
	if s, ok := v.(string); ok && s != "" {
		return []string{s}
	}
	return nil
}
