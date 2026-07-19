package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/yixian-huang/inkless/backend/internal/model"
	"github.com/yixian-huang/inkless/backend/internal/repository"
	"github.com/yixian-huang/inkless/backend/pkg/audit"
)

const (
	defaultClaimLimit    = 10
	defaultLeaseDuration = 2 * time.Minute
)

type SchedulerService struct {
	jobRepo    repository.ScheduledPublishJobRepository
	articleSvc *ArticlePublicationService
	pageSvc    *UnifiedPageService
	logger     *slog.Logger
	done       chan struct{}
	startOnce  sync.Once
	stopOnce   sync.Once
	wg         sync.WaitGroup
}

func NewSchedulerService(
	jobRepo repository.ScheduledPublishJobRepository,
	articleSvc *ArticlePublicationService,
	pageSvc *UnifiedPageService,
) *SchedulerService {
	return &SchedulerService{
		jobRepo:    jobRepo,
		articleSvc: articleSvc,
		pageSvc:    pageSvc,
		logger:     slog.Default(),
		done:       make(chan struct{}),
	}
}

func (s *SchedulerService) Schedule(
	ctx context.Context,
	contentType model.ScheduledContentType,
	contentID uint,
	scheduledAt time.Time,
	expectedVersion *int,
	publishPayload model.JSONMap,
	actorID uint,
) (*model.ScheduledPublishJob, error) {
	resolvedVersion, expectedUpdatedAt, err := s.prepareSchedule(ctx, contentType, contentID, expectedVersion, nil, publishPayload)
	if err != nil {
		return nil, err
	}
	job := &model.ScheduledPublishJob{
		ContentType:       contentType,
		ContentID:         contentID,
		Status:            model.ScheduledJobPending,
		ScheduledAt:       scheduledAt,
		ExpectedVersion:   resolvedVersion,
		ExpectedUpdatedAt: expectedUpdatedAt,
		PublishPayload:    publishPayload,
		MaxAttempts:       3,
		CreatedBy:         actorID,
		UpdatedBy:         actorID,
	}
	if err := s.jobRepo.Schedule(ctx, job); err != nil {
		return nil, err
	}
	return job, nil
}

func (s *SchedulerService) Reschedule(
	ctx context.Context,
	jobID uint,
	scheduledAt time.Time,
	expectedVersion *int,
	publishPayload model.JSONMap,
	actorID uint,
) (*model.ScheduledPublishJob, error) {
	current, err := s.jobRepo.FindByID(ctx, jobID)
	if err != nil {
		return nil, err
	}
	resolvedInputVersion := expectedVersion
	resolvedPayload := publishPayload
	var expectedUpdatedAt *time.Time
	switch current.ContentType {
	case model.ScheduledContentArticle:
		if publishPayload == nil {
			resolvedPayload = current.PublishPayload
			expectedUpdatedAt = current.ExpectedUpdatedAt
		}
	case model.ScheduledContentPage:
		if expectedVersion == nil {
			resolvedInputVersion = current.ExpectedVersion
		}
	}
	resolvedVersion, expectedUpdatedAt, err := s.prepareSchedule(
		ctx,
		current.ContentType,
		current.ContentID,
		resolvedInputVersion,
		expectedUpdatedAt,
		resolvedPayload,
	)
	if err != nil {
		return nil, err
	}
	job, err := s.jobRepo.UpdateSchedule(
		ctx,
		jobID,
		scheduledAt,
		resolvedVersion,
		expectedUpdatedAt,
		resolvedPayload,
		actorID,
	)
	if err != nil {
		return nil, err
	}
	return job, nil
}

func (s *SchedulerService) Cancel(ctx context.Context, jobID uint, actorID uint, cancelledAt time.Time) (*model.ScheduledPublishJob, error) {
	job, err := s.jobRepo.Cancel(ctx, jobID, actorID, cancelledAt)
	if err != nil {
		return nil, err
	}
	return job, nil
}

func (s *SchedulerService) Retry(
	ctx context.Context,
	jobID uint,
	actorID uint,
	retryAt time.Time,
) (*model.ScheduledPublishJob, error) {
	current, err := s.jobRepo.FindByID(ctx, jobID)
	if err != nil {
		return nil, err
	}
	resolvedVersion, _, err := s.prepareSchedule(
		ctx,
		current.ContentType,
		current.ContentID,
		current.ExpectedVersion,
		current.ExpectedUpdatedAt,
		current.PublishPayload,
	)
	if err != nil {
		return nil, err
	}
	job, err := s.jobRepo.Retry(ctx, jobID, retryAt, resolvedVersion, actorID)
	if err != nil {
		return nil, err
	}
	return job, nil
}

func (s *SchedulerService) Get(ctx context.Context, jobID uint) (*model.ScheduledPublishJob, error) {
	return s.jobRepo.FindByID(ctx, jobID)
}

func (s *SchedulerService) List(
	ctx context.Context,
	contentTypes []model.ScheduledContentType,
	contentID uint,
	status model.ScheduledJobStatus,
	offset,
	limit int,
) ([]*model.ScheduledPublishJob, int64, error) {
	return s.jobRepo.List(ctx, contentTypes, contentID, status, offset, limit)
}

func (s *SchedulerService) DescribeResource(
	ctx context.Context,
	job *model.ScheduledPublishJob,
) (title string, slug string) {
	if job == nil {
		return "", ""
	}
	switch job.ContentType {
	case model.ScheduledContentArticle:
		if payload, err := decodeScheduledArticlePayload(job.PublishPayload); err == nil && payload != nil {
			if payload.ZhTitle != nil {
				title = *payload.ZhTitle
			}
			if payload.Slug != nil {
				slug = *payload.Slug
			}
		}
		if title != "" || slug != "" || s.articleSvc == nil {
			return title, slug
		}
		if article, err := s.articleSvc.articleRepo.FindByID(ctx, job.ContentID); err == nil {
			return article.ZhTitle, article.Slug
		}
	case model.ScheduledContentPage:
		if s.pageSvc != nil {
			if page, err := s.pageSvc.pageRepo.FindByID(ctx, job.ContentID); err == nil {
				return page.ZhTitle, page.Slug
			}
		}
	}
	return title, slug
}

// PublishOverdue preserves the old public method name while using the durable queue.
func (s *SchedulerService) PublishOverdue(ctx context.Context) (int, error) {
	return s.RunDue(ctx, time.Now())
}

func (s *SchedulerService) RunDue(ctx context.Context, now time.Time) (int, error) {
	jobs, err := s.jobRepo.ClaimDue(ctx, now, defaultClaimLimit, defaultLeaseDuration)
	if err != nil {
		return 0, err
	}
	succeeded := 0
	var errs []error
	for _, job := range jobs {
		if err := s.runJob(ctx, job, now); err != nil {
			errs = append(errs, fmt.Errorf("job %d: %w", job.ID, err))
			continue
		}
		succeeded++
	}
	return succeeded, errors.Join(errs...)
}

func (s *SchedulerService) runJob(ctx context.Context, job *model.ScheduledPublishJob, now time.Time) error {
	ctx = audit.WithMetadata(ctx, audit.Metadata{
		Actor:   "scheduler",
		ActorID: job.CreatedBy,
	})
	var err error
	switch job.ContentType {
	case model.ScheduledContentArticle:
		if s.articleSvc == nil {
			err = errors.New("article publication service is not configured")
		} else {
			_, err = s.articleSvc.Publish(
				ctx,
				job.ContentID,
				now,
				job.CreatedBy,
				job.ExpectedUpdatedAt,
				job.PublishPayload,
			)
		}
	case model.ScheduledContentPage:
		if s.pageSvc == nil {
			err = errors.New("page publication service is not configured")
		} else {
			err = s.publishPage(ctx, job)
		}
	default:
		err = fmt.Errorf("unsupported content type %q", job.ContentType)
	}
	if err != nil {
		retryAt := now.Add(time.Duration(job.Attempts) * time.Minute)
		if markErr := s.jobRepo.MarkFailed(ctx, job, err.Error(), retryAt, now); markErr != nil {
			return errors.Join(err, fmt.Errorf("mark failed: %w", markErr))
		}
		return err
	}
	if err := s.jobRepo.MarkSucceeded(ctx, job, now); err != nil {
		return fmt.Errorf("mark succeeded: %w", err)
	}
	return nil
}

func (s *SchedulerService) publishPage(ctx context.Context, job *model.ScheduledPublishJob) error {
	page, err := s.pageSvc.pageRepo.FindByID(ctx, job.ContentID)
	if err != nil {
		return err
	}
	if page.Status == "published" && page.PublishedAt != nil && page.ScheduledAt == nil {
		return nil
	}
	if job.ExpectedVersion == nil {
		return errors.New("scheduled page job has no expected version")
	}
	return s.pageSvc.PublishScheduled(ctx, job.ContentID, *job.ExpectedVersion, job.CreatedBy)
}

func (s *SchedulerService) prepareSchedule(
	ctx context.Context,
	contentType model.ScheduledContentType,
	contentID uint,
	expectedVersion *int,
	expectedUpdatedAt *time.Time,
	publishPayload model.JSONMap,
) (*int, *time.Time, error) {
	switch contentType {
	case model.ScheduledContentArticle:
		if s.articleSvc == nil {
			return nil, nil, errors.New("article publication service is not configured")
		}
		article, err := s.articleSvc.articleRepo.FindByID(ctx, contentID)
		if err != nil {
			return nil, nil, err
		}
		if err := s.articleSvc.ValidatePublishPayload(ctx, publishPayload); err != nil {
			return nil, nil, err
		}
		if expectedUpdatedAt != nil && !article.UpdatedAt.Equal(*expectedUpdatedAt) {
			return nil, nil, ErrArticleVersionConflict
		}
		resolvedUpdatedAt := article.UpdatedAt
		return nil, &resolvedUpdatedAt, nil
	case model.ScheduledContentPage:
		if s.pageSvc == nil {
			return nil, nil, errors.New("page publication service is not configured")
		}
		page, err := s.pageSvc.pageRepo.FindByID(ctx, contentID)
		if err != nil {
			return nil, nil, err
		}
		resolved := page.DraftVersion
		if expectedVersion != nil {
			resolved = *expectedVersion
		}
		if resolved != page.DraftVersion {
			return nil, nil, ErrPageVersionConflict
		}
		return &resolved, nil, nil
	default:
		return nil, nil, fmt.Errorf("unsupported content type %q", contentType)
	}
}

func (s *SchedulerService) Start() {
	s.startOnce.Do(func() {
		s.wg.Add(1)
		go func() {
			defer s.wg.Done()
			s.runTick(time.Now())
			ticker := time.NewTicker(1 * time.Minute)
			defer ticker.Stop()
			for {
				select {
				case <-s.done:
					return
				case tickAt := <-ticker.C:
					s.runTick(tickAt)
				}
			}
		}()
		s.logger.Info("Scheduler started", "interval", "1m")
	})
}

func (s *SchedulerService) Stop() {
	s.stopOnce.Do(func() {
		close(s.done)
		s.wg.Wait()
		s.logger.Info("Scheduler stopped")
	})
}

func (s *SchedulerService) runTick(tickAt time.Time) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if _, err := s.RunDue(ctx, tickAt); err != nil {
		s.logger.Error("Scheduler error", "error", err)
	}
}
