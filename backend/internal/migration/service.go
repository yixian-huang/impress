package migration

import (
	"context"
	"fmt"
	"sync"
	"time"

	"blotting-consultancy/internal/model"
	"blotting-consultancy/internal/provider"
	"blotting-consultancy/internal/repository"
)

// Service orchestrates the import of parsed MigrationArticles into the
// database using the existing repository layer.
type Service struct {
	articleRepo  repository.ArticleRepository
	categoryRepo repository.CategoryRepository
	tagRepo      repository.TagRepository

	// In-flight job tracking
	mu   sync.RWMutex
	jobs map[string]*provider.MigrationProgress
}

// NewService creates a new migration service.
func NewService(
	articleRepo repository.ArticleRepository,
	categoryRepo repository.CategoryRepository,
	tagRepo repository.TagRepository,
) *Service {
	return &Service{
		articleRepo:  articleRepo,
		categoryRepo: categoryRepo,
		tagRepo:      tagRepo,
		jobs:         make(map[string]*provider.MigrationProgress),
	}
}

// GetProgress returns the current progress for a migration job.
func (s *Service) GetProgress(jobID string) (*provider.MigrationProgress, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	p, ok := s.jobs[jobID]
	if !ok {
		return nil, false
	}
	// Return a copy to avoid races
	cp := *p
	cp.Errors = make([]string, len(p.Errors))
	copy(cp.Errors, p.Errors)
	return &cp, true
}

// ListJobs returns all known migration jobs.
func (s *Service) ListJobs() []*provider.MigrationProgress {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]*provider.MigrationProgress, 0, len(s.jobs))
	for _, p := range s.jobs {
		cp := *p
		cp.Errors = make([]string, len(p.Errors))
		copy(cp.Errors, p.Errors)
		result = append(result, &cp)
	}
	return result
}

// ImportArticles writes parsed MigrationArticles to the database.
// It runs asynchronously: call GetProgress to monitor status.
// The returned jobID can be used to poll for progress.
func (s *Service) ImportArticles(
	ctx context.Context,
	jobID string,
	source provider.MigrationSource,
	articles []*provider.MigrationArticle,
	parseErrors []string,
) {
	progress := &provider.MigrationProgress{
		JobID:     jobID,
		Source:    source,
		Phase:     "importing",
		Total:     len(articles),
		Errors:    append([]string{}, parseErrors...),
		StartedAt: time.Now(),
	}

	s.mu.Lock()
	s.jobs[jobID] = progress
	s.mu.Unlock()

	go func() {
		defer func() {
			s.mu.Lock()
			now := time.Now()
			progress.FinishedAt = &now
			if progress.Failed > 0 && progress.Succeeded == 0 {
				progress.Phase = "failed"
			} else {
				progress.Phase = "done"
			}
			s.mu.Unlock()
		}()

		// Build/cache categories and tags
		categoryCache := make(map[string]uint)
		tagCache := make(map[string]uint)

		for _, article := range articles {
			select {
			case <-ctx.Done():
				s.mu.Lock()
				progress.Phase = "failed"
				progress.Errors = append(progress.Errors, "import cancelled")
				s.mu.Unlock()
				return
			default:
			}

			err := s.importOne(ctx, article, categoryCache, tagCache)

			s.mu.Lock()
			progress.Processed++
			if err != nil {
				progress.Failed++
				progress.Errors = append(progress.Errors, fmt.Sprintf("slug=%s: %v", article.Slug, err))
			} else {
				progress.Succeeded++
			}
			s.mu.Unlock()
		}
	}()
}

func (s *Service) importOne(
	ctx context.Context,
	ma *provider.MigrationArticle,
	categoryCache map[string]uint,
	tagCache map[string]uint,
) error {
	// Resolve or create category
	var categoryID *uint
	if ma.CategoryName != "" {
		id, err := s.resolveCategory(ctx, ma.CategoryName, categoryCache)
		if err != nil {
			return fmt.Errorf("resolving category %q: %w", ma.CategoryName, err)
		}
		categoryID = &id
	}

	// Resolve or create tags
	var tags []model.Tag
	for _, tagName := range ma.TagNames {
		id, err := s.resolveTag(ctx, tagName, tagCache)
		if err != nil {
			return fmt.Errorf("resolving tag %q: %w", tagName, err)
		}
		tags = append(tags, model.Tag{ID: id})
	}

	article := &model.Article{
		Slug:              ma.Slug,
		Status:            ma.Status,
		ZhTitle:           ma.Title,
		EnTitle:           ma.EnTitle,
		ZhBody:            ma.Body,
		EnBody:            ma.EnBody,
		CoverImage:        ma.CoverImageURL,
		ZhMetaDescription: ma.MetaDescription,
		ZhSeoTitle:        ma.SeoTitle,
		CategoryID:        categoryID,
		Tags:              tags,
		PublishedAt:       ma.PublishedAt,
	}

	// Check for slug collision: append suffix if needed
	if existing, _ := s.articleRepo.FindBySlug(ctx, article.Slug); existing != nil {
		article.Slug = fmt.Sprintf("%s-%d", article.Slug, time.Now().UnixMilli())
	}

	if err := s.articleRepo.Create(ctx, article); err != nil {
		return fmt.Errorf("creating article: %w", err)
	}

	return nil
}

func (s *Service) resolveCategory(ctx context.Context, name string, cache map[string]uint) (uint, error) {
	if id, ok := cache[name]; ok {
		return id, nil
	}

	// List all categories and find by name
	categories, err := s.categoryRepo.List(ctx)
	if err != nil {
		return 0, err
	}
	for _, c := range categories {
		if c.ZhName == name || c.EnName == name {
			cache[name] = c.ID
			return c.ID, nil
		}
	}

	// Create new category
	cat := &model.Category{
		Slug:   sanitizeSlug(name),
		ZhName: name,
	}
	if err := s.categoryRepo.Create(ctx, cat); err != nil {
		return 0, err
	}
	cache[name] = cat.ID
	return cat.ID, nil
}

func (s *Service) resolveTag(ctx context.Context, name string, cache map[string]uint) (uint, error) {
	if id, ok := cache[name]; ok {
		return id, nil
	}

	// List all tags and find by name
	tags, err := s.tagRepo.List(ctx)
	if err != nil {
		return 0, err
	}
	for _, t := range tags {
		if t.ZhName == name || t.EnName == name {
			cache[name] = t.ID
			return t.ID, nil
		}
	}

	// Create new tag
	tag := &model.Tag{
		Slug:   sanitizeSlug(name),
		ZhName: name,
	}
	if err := s.tagRepo.Create(ctx, tag); err != nil {
		return 0, err
	}
	cache[name] = tag.ID
	return tag.ID, nil
}
