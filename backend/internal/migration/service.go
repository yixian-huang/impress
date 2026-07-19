package migration

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"github.com/yixian-huang/inkless/backend/internal/model"
	"github.com/yixian-huang/inkless/backend/internal/provider"
	"github.com/yixian-huang/inkless/backend/internal/repository"
)

var (
	ErrJobNotFound     = errors.New("migration job not found")
	ErrJobRunning      = errors.New("migration job is still running")
	ErrJobNotFailed    = errors.New("migration job has not failed")
	ErrJobNotRetryable = errors.New("migration job is not retryable")
)

type importJob struct {
	progress       provider.MigrationProgress
	failedArticles []*provider.MigrationArticle
	parseErrors    []string
}

// Service orchestrates the import of parsed MigrationArticles into the
// database using the existing repository layer.
type Service struct {
	articleRepo  repository.ArticleRepository
	categoryRepo repository.CategoryRepository
	tagRepo      repository.TagRepository

	// In-flight job tracking
	mu      sync.RWMutex
	counter uint64
	jobs    map[string]*importJob
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
		jobs:         make(map[string]*importJob),
	}
}

// GetProgress returns the current progress for a migration job.
func (s *Service) GetProgress(jobID string) (*provider.MigrationProgress, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	job, ok := s.jobs[jobID]
	if !ok {
		return nil, false
	}
	return copyProgress(job.progress), true
}

// ListJobs returns all known migration jobs.
func (s *Service) ListJobs() []*provider.MigrationProgress {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]*provider.MigrationProgress, 0, len(s.jobs))
	for _, job := range s.jobs {
		result = append(result, copyProgress(job.progress))
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].StartedAt.Equal(result[j].StartedAt) {
			return result[i].JobID > result[j].JobID
		}
		return result[i].StartedAt.After(result[j].StartedAt)
	})
	return result
}

// StartImport writes parsed MigrationArticles to the database asynchronously.
// The returned jobID can be used to poll or stream progress.
func (s *Service) StartImport(
	ctx context.Context,
	source provider.MigrationSource,
	articles []*provider.MigrationArticle,
	parseErrors []string,
) string {
	jobID := s.nextJobID(source)
	job := &importJob{
		parseErrors: append([]string(nil), parseErrors...),
	}
	job.progress = provider.MigrationProgress{
		JobID:     jobID,
		Source:    source,
		Phase:     "importing",
		Total:     len(articles),
		Errors:    append([]string{}, parseErrors...),
		Attempt:   1,
		StartedAt: time.Now(),
	}

	s.mu.Lock()
	s.jobs[jobID] = job
	s.mu.Unlock()

	go s.runImport(ctx, jobID, cloneArticles(articles))
	return jobID
}

// ImportArticles preserves the previous call shape for existing callers.
func (s *Service) ImportArticles(
	ctx context.Context,
	jobID string,
	source provider.MigrationSource,
	articles []*provider.MigrationArticle,
	parseErrors []string,
) {
	job := &importJob{
		parseErrors: append([]string(nil), parseErrors...),
	}
	job.progress = provider.MigrationProgress{
		JobID:     jobID,
		Source:    source,
		Phase:     "importing",
		Total:     len(articles),
		Errors:    append([]string{}, parseErrors...),
		Attempt:   1,
		StartedAt: time.Now(),
	}

	s.mu.Lock()
	s.jobs[jobID] = job
	s.mu.Unlock()

	go s.runImport(ctx, jobID, cloneArticles(articles))
}

// RetryJob retries only the articles that failed in the previous attempt.
func (s *Service) RetryJob(ctx context.Context, jobID string) (*provider.MigrationProgress, error) {
	s.mu.Lock()
	job, ok := s.jobs[jobID]
	if !ok {
		s.mu.Unlock()
		return nil, ErrJobNotFound
	}
	if isRunning(job.progress.Phase) {
		s.mu.Unlock()
		return nil, ErrJobRunning
	}
	if job.progress.Phase != "failed" {
		s.mu.Unlock()
		return nil, ErrJobNotFailed
	}
	if len(job.failedArticles) == 0 {
		s.mu.Unlock()
		return nil, ErrJobNotRetryable
	}

	articles := cloneArticles(job.failedArticles)
	job.failedArticles = nil
	job.progress.Phase = "importing"
	job.progress.Processed = job.progress.Succeeded
	job.progress.Failed = 0
	job.progress.Errors = append([]string(nil), job.parseErrors...)
	job.progress.Attempt++
	job.progress.Retryable = false
	job.progress.FinishedAt = nil
	progress := copyProgress(job.progress)
	s.mu.Unlock()

	go s.runImport(ctx, jobID, articles)
	return progress, nil
}

func (s *Service) runImport(ctx context.Context, jobID string, articles []*provider.MigrationArticle) {
	categoryCache := make(map[string]uint)
	tagCache := make(map[string]uint)

	defer s.finishJob(jobID)

	for i, article := range articles {
		select {
		case <-ctx.Done():
			s.mu.Lock()
			if job, ok := s.jobs[jobID]; ok {
				remaining := cloneArticles(articles[i:])
				job.progress.Processed += len(remaining)
				job.progress.Failed += len(remaining)
				job.progress.Errors = append(job.progress.Errors, "import cancelled")
				job.failedArticles = append(job.failedArticles, remaining...)
			}
			s.mu.Unlock()
			return
		default:
		}

		err := s.importOne(ctx, article, categoryCache, tagCache)

		s.mu.Lock()
		if job, ok := s.jobs[jobID]; ok {
			job.progress.Processed++
			if err != nil {
				job.progress.Failed++
				job.progress.Errors = append(job.progress.Errors, fmt.Sprintf("slug=%s: %v", article.Slug, err))
				job.failedArticles = append(job.failedArticles, cloneArticle(article))
			} else {
				job.progress.Succeeded++
			}
		}
		s.mu.Unlock()
	}
}

func (s *Service) finishJob(jobID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	job, ok := s.jobs[jobID]
	if !ok {
		return
	}
	now := time.Now()
	job.progress.FinishedAt = &now
	if job.progress.Failed > 0 {
		job.progress.Phase = "failed"
		job.progress.Retryable = len(job.failedArticles) > 0
		return
	}
	job.progress.Phase = "done"
	job.progress.Retryable = false
}

func (s *Service) nextJobID(source provider.MigrationSource) string {
	sequence := atomic.AddUint64(&s.counter, 1)
	return fmt.Sprintf("mig-%s-%d-%d", source, time.Now().UnixNano(), sequence)
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

func copyProgress(p provider.MigrationProgress) *provider.MigrationProgress {
	cp := p
	cp.Errors = append([]string(nil), p.Errors...)
	if p.FinishedAt != nil {
		finishedAt := *p.FinishedAt
		cp.FinishedAt = &finishedAt
	}
	return &cp
}

func cloneArticles(articles []*provider.MigrationArticle) []*provider.MigrationArticle {
	cloned := make([]*provider.MigrationArticle, 0, len(articles))
	for _, article := range articles {
		cloned = append(cloned, cloneArticle(article))
	}
	return cloned
}

func cloneArticle(article *provider.MigrationArticle) *provider.MigrationArticle {
	if article == nil {
		return nil
	}
	cp := *article
	cp.TagNames = append([]string(nil), article.TagNames...)
	cp.MediaURLs = append([]string(nil), article.MediaURLs...)
	return &cp
}

func isRunning(phase string) bool {
	return phase == "parsing" || phase == "importing"
}

func (s *Service) resolveCategory(ctx context.Context, name string, cache map[string]uint) (uint, error) {
	slug := sanitizeSlug(name)
	if id, ok := cache[slug]; ok {
		return id, nil
	}

	if category, err := s.categoryRepo.FindBySlug(ctx, slug); err == nil {
		cache[slug] = category.ID
		return category.ID, nil
	}

	cat := &model.Category{
		Slug:   slug,
		ZhName: name,
	}
	if err := s.categoryRepo.Create(ctx, cat); err != nil {
		if existing, findErr := s.categoryRepo.FindBySlug(ctx, slug); findErr == nil {
			cache[slug] = existing.ID
			return existing.ID, nil
		}
		return 0, err
	}
	cache[slug] = cat.ID
	return cat.ID, nil
}

func (s *Service) resolveTag(ctx context.Context, name string, cache map[string]uint) (uint, error) {
	slug := sanitizeSlug(name)
	if id, ok := cache[slug]; ok {
		return id, nil
	}

	if tag, err := s.tagRepo.FindBySlug(ctx, slug); err == nil {
		cache[slug] = tag.ID
		return tag.ID, nil
	}

	tag := &model.Tag{
		Slug:   slug,
		ZhName: name,
	}
	if err := s.tagRepo.Create(ctx, tag); err != nil {
		if existing, findErr := s.tagRepo.FindBySlug(ctx, slug); findErr == nil {
			cache[slug] = existing.ID
			return existing.ID, nil
		}
		return 0, err
	}
	cache[slug] = tag.ID
	return tag.ID, nil
}
