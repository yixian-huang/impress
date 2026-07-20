package migration

import (
	"github.com/yixian-huang/inkless/backend/internal/repository"
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/yixian-huang/inkless/backend/internal/model"
	"github.com/yixian-huang/inkless/backend/internal/provider"
)

type migrationArticleRepoStub struct {
	mu        sync.Mutex
	failures  map[string]int
	created   []string
	createdID uint
}

func (r *migrationArticleRepoStub) Create(_ context.Context, article *model.Article) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.created = append(r.created, article.Slug)
	if r.failures[article.Slug] > 0 {
		r.failures[article.Slug]--
		return errors.New("create failed")
	}
	r.createdID++
	article.ID = r.createdID
	return nil
}

func (r *migrationArticleRepoStub) FindByID(context.Context, uint) (*model.Article, error) {
	return nil, nil
}

func (r *migrationArticleRepoStub) FindBySlug(context.Context, string) (*model.Article, error) {
	return nil, nil
}

func (r *migrationArticleRepoStub) Update(context.Context, *model.Article) error {
	return nil
}

func (r *migrationArticleRepoStub) UpdateScheduledPublication(context.Context, *model.Article, time.Time) error {
	return nil
}

func (r *migrationArticleRepoStub) Delete(context.Context, uint) error {
	return nil
}

func (r *migrationArticleRepoStub) Count(context.Context, string) (int64, error) {
	return 0, nil
}

func (r *migrationArticleRepoStub) UpdateIfMatch(context.Context, *model.Article, time.Time) error {
	return nil
}

func (r *migrationArticleRepoStub) List(context.Context, int, int, string, *uint, *uint) ([]*model.Article, int64, error) {
	return nil, 0, nil
}

func (r *migrationArticleRepoStub) ListPublished(context.Context, int, int, string, string) ([]*model.Article, int64, error) {
	return nil, 0, nil
}

func (r *migrationArticleRepoStub) ListPublishedSitemapMeta(context.Context, int) ([]repository.ArticleSitemapMeta, error) {
	return nil, nil
}

type migrationCategoryRepoStub struct {
	mu             sync.Mutex
	bySlug         map[string]*model.Category
	failCreateOnce bool
	conflictID     uint
}

func (r *migrationCategoryRepoStub) Create(_ context.Context, category *model.Category) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.failCreateOnce {
		r.failCreateOnce = false
		if r.bySlug == nil {
			r.bySlug = make(map[string]*model.Category)
		}
		r.bySlug[category.Slug] = &model.Category{ID: r.conflictID, Slug: category.Slug, ZhName: category.ZhName}
		return errors.New("unique constraint")
	}
	category.ID = 1
	if r.bySlug == nil {
		r.bySlug = make(map[string]*model.Category)
	}
	cp := *category
	r.bySlug[category.Slug] = &cp
	return nil
}

func (r *migrationCategoryRepoStub) FindByID(context.Context, uint) (*model.Category, error) {
	return nil, nil
}

func (r *migrationCategoryRepoStub) FindBySlug(_ context.Context, slug string) (*model.Category, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	category, ok := r.bySlug[slug]
	if !ok {
		return nil, errors.New("category not found")
	}
	cp := *category
	return &cp, nil
}

func (r *migrationCategoryRepoStub) Update(context.Context, *model.Category) error {
	return nil
}

func (r *migrationCategoryRepoStub) Delete(context.Context, uint) error {
	return nil
}

func (r *migrationCategoryRepoStub) List(context.Context) ([]*model.Category, error) {
	return nil, nil
}

func (r *migrationCategoryRepoStub) ListTree(context.Context) ([]*model.Category, error) {
	return nil, nil
}

func (r *migrationCategoryRepoStub) ListByParentID(context.Context, *uint) ([]*model.Category, error) {
	return nil, nil
}

func (r *migrationCategoryRepoStub) FindByIDs(context.Context, []uint) ([]model.Category, error) {
	return nil, nil
}

type migrationTagRepoStub struct {
	mu             sync.Mutex
	bySlug         map[string]*model.Tag
	failCreateOnce bool
	conflictID     uint
}

func (r *migrationTagRepoStub) Create(_ context.Context, tag *model.Tag) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.failCreateOnce {
		r.failCreateOnce = false
		if r.bySlug == nil {
			r.bySlug = make(map[string]*model.Tag)
		}
		r.bySlug[tag.Slug] = &model.Tag{ID: r.conflictID, Slug: tag.Slug, ZhName: tag.ZhName}
		return errors.New("unique constraint")
	}
	tag.ID = 1
	if r.bySlug == nil {
		r.bySlug = make(map[string]*model.Tag)
	}
	cp := *tag
	r.bySlug[tag.Slug] = &cp
	return nil
}

func (r *migrationTagRepoStub) FindByID(context.Context, uint) (*model.Tag, error) {
	return nil, nil
}

func (r *migrationTagRepoStub) FindBySlug(_ context.Context, slug string) (*model.Tag, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	tag, ok := r.bySlug[slug]
	if !ok {
		return nil, errors.New("tag not found")
	}
	cp := *tag
	return &cp, nil
}

func (r *migrationTagRepoStub) Update(context.Context, *model.Tag) error {
	return nil
}

func (r *migrationTagRepoStub) Delete(context.Context, uint) error {
	return nil
}

func (r *migrationTagRepoStub) List(context.Context) ([]*model.Tag, error) {
	return nil, nil
}

func (r *migrationTagRepoStub) FindByIDs(context.Context, []uint) ([]model.Tag, error) {
	return nil, nil
}

func TestServicePartialFailureIsRetryableAndRetryImportsOnlyFailedArticles(t *testing.T) {
	articleRepo := &migrationArticleRepoStub{failures: map[string]int{"bad": 1}}
	service := NewService(articleRepo, &migrationCategoryRepoStub{}, &migrationTagRepoStub{})

	jobID := service.StartImport(context.Background(), provider.SourceMarkdown, []*provider.MigrationArticle{
		{Slug: "good", Title: "Good", Status: model.ArticleStatusPublished},
		{Slug: "bad", Title: "Bad", Status: model.ArticleStatusPublished},
	}, nil)

	progress := waitForPhase(t, service, jobID, "failed")
	require.Equal(t, 2, progress.Processed)
	require.Equal(t, 1, progress.Succeeded)
	require.Equal(t, 1, progress.Failed)
	require.True(t, progress.Retryable)

	retryProgress, err := service.RetryJob(context.Background(), jobID)
	require.NoError(t, err)
	require.Equal(t, 2, retryProgress.Attempt)
	require.Equal(t, 2, retryProgress.Total)
	require.Equal(t, 1, retryProgress.Processed)
	require.Equal(t, 1, retryProgress.Succeeded)

	progress = waitForPhase(t, service, jobID, "done")
	require.Equal(t, 2, progress.Processed)
	require.Equal(t, 2, progress.Succeeded)
	require.Equal(t, 0, progress.Failed)
	require.False(t, progress.Retryable)

	articleRepo.mu.Lock()
	defer articleRepo.mu.Unlock()
	require.Equal(t, []string{"good", "bad", "bad"}, articleRepo.created)
}

func TestServiceListJobsNewestFirstAndProgressCopiesAreIsolated(t *testing.T) {
	service := NewService(&migrationArticleRepoStub{failures: map[string]int{}}, &migrationCategoryRepoStub{}, &migrationTagRepoStub{})

	first := service.StartImport(context.Background(), provider.SourceMarkdown, []*provider.MigrationArticle{
		{Slug: "first", Title: "First", Status: model.ArticleStatusPublished},
	}, []string{"first warning"})
	waitForPhase(t, service, first, "done")
	time.Sleep(time.Millisecond)
	second := service.StartImport(context.Background(), provider.SourceHalo, []*provider.MigrationArticle{
		{Slug: "second", Title: "Second", Status: model.ArticleStatusPublished},
	}, nil)
	waitForPhase(t, service, second, "done")

	require.NotEqual(t, first, second)
	jobs := service.ListJobs()
	require.Len(t, jobs, 2)
	require.Equal(t, second, jobs[0].JobID)
	require.Equal(t, first, jobs[1].JobID)

	progress, ok := service.GetProgress(first)
	require.True(t, ok)
	progress.Errors[0] = "mutated"
	require.NotNil(t, progress.FinishedAt)
	*progress.FinishedAt = time.Time{}
	progress, ok = service.GetProgress(first)
	require.True(t, ok)
	require.Equal(t, []string{"first warning"}, progress.Errors)
	require.False(t, progress.FinishedAt.IsZero())
}

func TestServiceResolvesTaxonomyBySlugAndRecoversCreateRace(t *testing.T) {
	categoryRepo := &migrationCategoryRepoStub{
		bySlug: map[string]*model.Category{
			"foo-bar": {ID: 7, Slug: "foo-bar", ZhName: "Existing"},
		},
	}
	tagRepo := &migrationTagRepoStub{
		failCreateOnce: true,
		conflictID:     9,
	}
	service := NewService(
		&migrationArticleRepoStub{failures: map[string]int{}},
		categoryRepo,
		tagRepo,
	)

	categoryID, err := service.resolveCategory(context.Background(), "foo_bar", map[string]uint{})
	require.NoError(t, err)
	require.Equal(t, uint(7), categoryID)

	tagID, err := service.resolveTag(context.Background(), "race tag", map[string]uint{})
	require.NoError(t, err)
	require.Equal(t, uint(9), tagID)

	cjkCache := map[string]uint{}
	firstCJKID, err := service.resolveCategory(context.Background(), "中文分类", cjkCache)
	require.NoError(t, err)
	secondCJKID, err := service.resolveCategory(context.Background(), " 中文分类 ", cjkCache)
	require.NoError(t, err)
	require.Equal(t, firstCJKID, secondCJKID)
	require.Len(t, categoryRepo.bySlug, 2)
}

func waitForPhase(t *testing.T, service *Service, jobID, phase string) *provider.MigrationProgress {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		progress, ok := service.GetProgress(jobID)
		require.True(t, ok)
		if progress.Phase == phase {
			return progress
		}
		time.Sleep(5 * time.Millisecond)
	}
	progress, _ := service.GetProgress(jobID)
	t.Fatalf("job %s did not reach phase %s; last progress: %+v", jobID, phase, progress)
	return nil
}

func (r *migrationArticleRepoStub) ListFilter(context.Context, repository.ArticleListFilter) ([]*model.Article, int64, error) {
	return nil, 0, nil
}

