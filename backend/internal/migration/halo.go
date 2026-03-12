package migration

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"blotting-consultancy/internal/model"
	"blotting-consultancy/internal/provider"
)

// Halo JSON export structures
// Halo is a popular Java-based blog platform. Its export format is a JSON
// document containing posts, categories, tags, and other metadata.

type haloExport struct {
	Version    int            `json:"version"`
	Posts      []haloPost     `json:"posts"`
	Categories []haloCategory `json:"categories"`
	Tags       []haloTag      `json:"tags"`
	PostTags   []haloPostTag  `json:"post_tags"`
	PostCategories []haloPostCategory `json:"post_categories"`
}

type haloPost struct {
	ID             int    `json:"id"`
	Title          string `json:"title"`
	Slug           string `json:"slug"`
	OriginalContent string `json:"originalContent"` // Markdown source
	FormatContent  string `json:"formatContent"`    // rendered HTML
	Summary        string `json:"summary"`
	Thumbnail      string `json:"thumbnail"`
	Status         string `json:"status"` // "PUBLISHED", "DRAFT", "RECYCLE"
	CreateTime     int64  `json:"createTime"`     // millis
	UpdateTime     int64  `json:"updateTime"`     // millis
	EditTime       int64  `json:"editTime"`
}

type haloCategory struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
}

type haloTag struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type haloPostTag struct {
	PostID int `json:"postId"`
	TagID  int `json:"tagId"`
}

type haloPostCategory struct {
	PostID     int `json:"postId"`
	CategoryID int `json:"categoryId"`
}

// HaloProvider parses Halo JSON export files.
type HaloProvider struct{}

// NewHaloProvider creates a new Halo migration provider.
func NewHaloProvider() *HaloProvider {
	return &HaloProvider{}
}

// Source returns the migration source identifier.
func (p *HaloProvider) Source() provider.MigrationSource {
	return provider.SourceHalo
}

// Parse reads a Halo JSON export and returns intermediate migration articles.
func (p *HaloProvider) Parse(ctx context.Context, r io.Reader) (*provider.MigrationResult, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("reading Halo export: %w", err)
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	var export haloExport
	if err := json.Unmarshal(data, &export); err != nil {
		return nil, fmt.Errorf("parsing Halo JSON: %w", err)
	}

	// Build lookup maps
	categoryMap := make(map[int]*haloCategory, len(export.Categories))
	for i := range export.Categories {
		categoryMap[export.Categories[i].ID] = &export.Categories[i]
	}

	tagMap := make(map[int]*haloTag, len(export.Tags))
	for i := range export.Tags {
		tagMap[export.Tags[i].ID] = &export.Tags[i]
	}

	// Build post -> category and post -> tags mappings
	postCategoryMap := make(map[int]int) // postID -> categoryID
	for _, pc := range export.PostCategories {
		postCategoryMap[pc.PostID] = pc.CategoryID
	}

	postTagsMap := make(map[int][]int) // postID -> []tagID
	for _, pt := range export.PostTags {
		postTagsMap[pt.PostID] = append(postTagsMap[pt.PostID], pt.TagID)
	}

	result := &provider.MigrationResult{}

	for _, post := range export.Posts {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		article, parseErrors := p.convertPost(post, categoryMap, tagMap, postCategoryMap, postTagsMap)
		result.Errors = append(result.Errors, parseErrors...)
		if article != nil {
			result.Articles = append(result.Articles, article)
		}
	}

	return result, nil
}

func (p *HaloProvider) convertPost(
	post haloPost,
	categoryMap map[int]*haloCategory,
	tagMap map[int]*haloTag,
	postCategoryMap map[int]int,
	postTagsMap map[int][]int,
) (*provider.MigrationArticle, []string) {
	var errs []string

	slug := post.Slug
	if slug == "" {
		slug = sanitizeSlug(post.Title)
	}
	if slug == "" && post.Title == "" {
		errs = append(errs, fmt.Sprintf("skipping post id=%d with empty title and slug", post.ID))
		return nil, errs
	}

	// Map Halo status
	status := model.ArticleStatusDraft
	if post.Status == "PUBLISHED" {
		status = model.ArticleStatusPublished
	}

	// Parse timestamps (Halo uses milliseconds since epoch)
	var publishedAt *time.Time
	var createdAt *time.Time
	if post.CreateTime > 0 {
		t := time.UnixMilli(post.CreateTime)
		createdAt = &t
		if status == model.ArticleStatusPublished {
			publishedAt = &t
		}
	}

	// Resolve category
	var categoryName string
	if catID, ok := postCategoryMap[post.ID]; ok {
		if cat, exists := categoryMap[catID]; exists {
			categoryName = cat.Name
		}
	}

	// Resolve tags
	var tagNames []string
	if tagIDs, ok := postTagsMap[post.ID]; ok {
		for _, tagID := range tagIDs {
			if tag, exists := tagMap[tagID]; exists {
				tagNames = append(tagNames, tag.Name)
			}
		}
	}

	// Prefer rendered HTML content; fall back to original Markdown
	body := post.FormatContent
	if body == "" {
		body = post.OriginalContent
	}

	// Extract media URLs from content
	mediaURLs := extractMediaURLs(body)

	article := &provider.MigrationArticle{
		Slug:            slug,
		Title:           post.Title,
		Body:            body,
		Status:          status,
		CategoryName:    categoryName,
		TagNames:        tagNames,
		CoverImageURL:   post.Thumbnail,
		MediaURLs:       mediaURLs,
		PublishedAt:     publishedAt,
		CreatedAt:       createdAt,
		MetaDescription: post.Summary,
	}

	return article, errs
}
