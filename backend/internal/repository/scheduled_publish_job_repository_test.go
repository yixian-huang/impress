package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/yixian-huang/inkless/backend/internal/model"
	"github.com/yixian-huang/inkless/backend/internal/repository"
)

func newScheduledJobRepository(t *testing.T) (repository.ScheduledPublishJobRepository, *gorm.DB) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(
		&model.Article{},
		&model.UnifiedPage{},
		&model.ScheduledPublishJob{},
	))
	return repository.NewGormScheduledPublishJobRepository(db), db
}

func TestScheduledJobRepositoryReplacesPendingButNotRunningJob(t *testing.T) {
	repo, db := newScheduledJobRepository(t)
	ctx := context.Background()
	now := time.Date(2026, 7, 16, 12, 0, 0, 0, time.UTC)
	require.NoError(t, db.Create(&model.Article{
		ID:      7,
		Slug:    "article-7",
		ZhTitle: "Article 7",
		Status:  model.ArticleStatusDraft,
	}).Error)

	first := &model.ScheduledPublishJob{
		ContentType: model.ScheduledContentArticle,
		ContentID:   7,
		Status:      model.ScheduledJobPending,
		ScheduledAt: now,
	}
	require.NoError(t, repo.Schedule(ctx, first))
	var scheduledArticle model.Article
	require.NoError(t, db.First(&scheduledArticle, 7).Error)
	require.Equal(t, model.ArticleStatusScheduled, scheduledArticle.Status)
	require.NotNil(t, scheduledArticle.ScheduledAt)

	second := &model.ScheduledPublishJob{
		ContentType: model.ScheduledContentArticle,
		ContentID:   7,
		Status:      model.ScheduledJobPending,
		ScheduledAt: now.Add(time.Hour),
	}
	require.NoError(t, repo.Schedule(ctx, second))

	replaced, err := repo.FindByID(ctx, first.ID)
	require.NoError(t, err)
	require.Equal(t, model.ScheduledJobCancelled, replaced.Status)
	require.Nil(t, replaced.ActiveKey)

	claimed, err := repo.ClaimDue(ctx, now.Add(2*time.Hour), 1, time.Minute)
	require.NoError(t, err)
	require.Len(t, claimed, 1)
	require.Equal(t, second.ID, claimed[0].ID)

	third := &model.ScheduledPublishJob{
		ContentType: model.ScheduledContentArticle,
		ContentID:   7,
		Status:      model.ScheduledJobPending,
		ScheduledAt: now.Add(3 * time.Hour),
	}
	require.Error(t, repo.Schedule(ctx, third))

	running, err := repo.FindByID(ctx, second.ID)
	require.NoError(t, err)
	require.Equal(t, model.ScheduledJobRunning, running.Status)
}

func TestScheduledJobRepositoryDoesNotCancelRunningJob(t *testing.T) {
	repo, db := newScheduledJobRepository(t)
	ctx := context.Background()
	now := time.Date(2026, 7, 16, 12, 0, 0, 0, time.UTC)
	require.NoError(t, db.Create(&model.UnifiedPage{
		ID:           9,
		Slug:         "page-9",
		ZhTitle:      "Page 9",
		Mode:         model.PageModeComposable,
		DraftConfig:  model.JSONMap{"sections": []interface{}{}},
		DraftVersion: 1,
		Status:       "draft",
	}).Error)
	job := &model.ScheduledPublishJob{
		ContentType: model.ScheduledContentPage,
		ContentID:   9,
		Status:      model.ScheduledJobPending,
		ScheduledAt: now,
	}
	require.NoError(t, repo.Schedule(ctx, job))
	claimed, err := repo.ClaimDue(ctx, now, 1, time.Minute)
	require.NoError(t, err)
	require.Len(t, claimed, 1)

	_, err = repo.Cancel(ctx, job.ID, 3, now)
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)

	current, err := repo.FindByID(ctx, job.ID)
	require.NoError(t, err)
	require.Equal(t, model.ScheduledJobRunning, current.Status)
}

func TestScheduledJobRepositoryCancelUpdatesContentInSameTransaction(t *testing.T) {
	repo, db := newScheduledJobRepository(t)
	ctx := context.Background()
	now := time.Date(2026, 7, 16, 12, 0, 0, 0, time.UTC)
	article := &model.Article{
		Slug:    "atomic-cancel",
		ZhTitle: "Atomic Cancel",
		Status:  model.ArticleStatusDraft,
	}
	require.NoError(t, db.Create(article).Error)
	job := &model.ScheduledPublishJob{
		ContentType: model.ScheduledContentArticle,
		ContentID:   article.ID,
		Status:      model.ScheduledJobPending,
		ScheduledAt: now.Add(time.Hour),
	}
	require.NoError(t, repo.Schedule(ctx, job))

	_, err := repo.Cancel(ctx, job.ID, 3, now)
	require.NoError(t, err)

	var updated model.Article
	require.NoError(t, db.First(&updated, article.ID).Error)
	require.Equal(t, model.ArticleStatusDraft, updated.Status)
	require.Nil(t, updated.ScheduledAt)
}

func TestScheduledJobRepositoryRollsBackJobWhenContentProjectionIsMissing(t *testing.T) {
	repo, _ := newScheduledJobRepository(t)
	ctx := context.Background()
	job := &model.ScheduledPublishJob{
		ContentType: model.ScheduledContentArticle,
		ContentID:   404,
		Status:      model.ScheduledJobPending,
		ScheduledAt: time.Date(2026, 7, 16, 12, 0, 0, 0, time.UTC),
	}

	require.ErrorIs(t, repo.Schedule(ctx, job), gorm.ErrRecordNotFound)
	_, err := repo.FindByID(ctx, job.ID)
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestScheduledJobRepositoryRecoversExpiredFinalLeaseAndRejectsStaleWorker(t *testing.T) {
	repo, db := newScheduledJobRepository(t)
	ctx := context.Background()
	now := time.Date(2026, 7, 16, 12, 0, 0, 0, time.UTC)
	article := &model.Article{
		Slug:    "lease-recovery",
		ZhTitle: "Lease Recovery",
		Status:  model.ArticleStatusDraft,
	}
	require.NoError(t, db.Create(article).Error)
	job := &model.ScheduledPublishJob{
		ContentType: model.ScheduledContentArticle,
		ContentID:   article.ID,
		Status:      model.ScheduledJobPending,
		ScheduledAt: now,
		MaxAttempts: 1,
	}
	require.NoError(t, repo.Schedule(ctx, job))

	firstClaim, err := repo.ClaimDue(ctx, now, 1, time.Minute)
	require.NoError(t, err)
	require.Len(t, firstClaim, 1)
	require.Equal(t, 1, firstClaim[0].Attempts)
	require.NotEmpty(t, firstClaim[0].LeaseToken)

	secondClaim, err := repo.ClaimDue(ctx, now.Add(2*time.Minute), 1, time.Minute)
	require.NoError(t, err)
	require.Len(t, secondClaim, 1)
	require.Equal(t, 2, secondClaim[0].Attempts)
	require.NotEqual(t, firstClaim[0].LeaseToken, secondClaim[0].LeaseToken)

	require.ErrorIs(
		t,
		repo.MarkSucceeded(ctx, firstClaim[0], now.Add(2*time.Minute)),
		gorm.ErrRecordNotFound,
	)
	current, err := repo.FindByID(ctx, job.ID)
	require.NoError(t, err)
	require.Equal(t, model.ScheduledJobRunning, current.Status)
	require.Equal(t, secondClaim[0].LeaseToken, current.LeaseToken)

	require.NoError(t, repo.MarkSucceeded(ctx, secondClaim[0], now.Add(2*time.Minute)))
	completed, err := repo.FindByID(ctx, job.ID)
	require.NoError(t, err)
	require.Equal(t, model.ScheduledJobSucceeded, completed.Status)
	require.Empty(t, completed.LeaseToken)
}
