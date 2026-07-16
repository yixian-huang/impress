package repository

import (
	"context"
	"time"

	"blotting-consultancy/internal/model"
)

type ScheduledPublishJobRepository interface {
	Schedule(ctx context.Context, job *model.ScheduledPublishJob) error
	FindByID(ctx context.Context, id uint) (*model.ScheduledPublishJob, error)
	FindActiveByContent(ctx context.Context, contentType model.ScheduledContentType, contentID uint) (*model.ScheduledPublishJob, error)
	List(ctx context.Context, contentTypes []model.ScheduledContentType, contentID uint, status model.ScheduledJobStatus, offset, limit int) ([]*model.ScheduledPublishJob, int64, error)
	UpdateSchedule(ctx context.Context, id uint, scheduledAt time.Time, expectedVersion *int, expectedUpdatedAt *time.Time, publishPayload model.JSONMap, actorID uint) (*model.ScheduledPublishJob, error)
	Cancel(ctx context.Context, id uint, actorID uint, cancelledAt time.Time) (*model.ScheduledPublishJob, error)
	Retry(ctx context.Context, id uint, scheduledAt time.Time, expectedVersion *int, actorID uint) (*model.ScheduledPublishJob, error)
	ClaimDue(ctx context.Context, now time.Time, limit int, leaseDuration time.Duration) ([]*model.ScheduledPublishJob, error)
	MarkSucceeded(ctx context.Context, job *model.ScheduledPublishJob, finishedAt time.Time) error
	MarkFailed(ctx context.Context, job *model.ScheduledPublishJob, errText string, retryAt time.Time, failedAt time.Time) error
}
