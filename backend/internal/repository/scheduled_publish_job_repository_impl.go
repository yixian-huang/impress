package repository

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/yixian-huang/inkless/backend/internal/model"
	"gorm.io/gorm"
)

type GormScheduledPublishJobRepository struct {
	db *gorm.DB
}

func NewGormScheduledPublishJobRepository(db *gorm.DB) ScheduledPublishJobRepository {
	return &GormScheduledPublishJobRepository{db: db}
}

func (r *GormScheduledPublishJobRepository) Schedule(ctx context.Context, job *model.ScheduledPublishJob) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.ScheduledPublishJob{}).
			Where("content_type = ? AND content_id = ? AND status = ?", job.ContentType, job.ContentID, model.ScheduledJobPending).
			Updates(map[string]interface{}{
				"status":       model.ScheduledJobCancelled,
				"cancelled_at": time.Now(),
				"active_key":   nil,
				"lease_until":  nil,
				"lease_token":  "",
				"claimed_at":   nil,
			}).Error; err != nil {
			return err
		}
		if err := tx.Create(job).Error; err != nil {
			return err
		}
		return updateScheduledContentProjection(tx, job.ContentType, job.ContentID, &job.ScheduledAt, true)
	})
}

func (r *GormScheduledPublishJobRepository) FindByID(ctx context.Context, id uint) (*model.ScheduledPublishJob, error) {
	var job model.ScheduledPublishJob
	if err := r.db.WithContext(ctx).First(&job, id).Error; err != nil {
		return nil, err
	}
	return &job, nil
}

func (r *GormScheduledPublishJobRepository) FindActiveByContent(
	ctx context.Context,
	contentType model.ScheduledContentType,
	contentID uint,
) (*model.ScheduledPublishJob, error) {
	var job model.ScheduledPublishJob
	err := r.db.WithContext(ctx).
		Where("content_type = ? AND content_id = ? AND status IN ?", contentType, contentID, []model.ScheduledJobStatus{
			model.ScheduledJobPending,
			model.ScheduledJobRunning,
		}).
		Order("scheduled_at ASC, id ASC").
		First(&job).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if err != nil {
		return nil, err
	}
	return &job, nil
}

func (r *GormScheduledPublishJobRepository) List(
	ctx context.Context,
	contentTypes []model.ScheduledContentType,
	contentID uint,
	status model.ScheduledJobStatus,
	offset,
	limit int,
) ([]*model.ScheduledPublishJob, int64, error) {
	var items []*model.ScheduledPublishJob
	var total int64
	q := r.db.WithContext(ctx).Model(&model.ScheduledPublishJob{})
	if len(contentTypes) > 0 {
		q = q.Where("content_type IN ?", contentTypes)
	}
	if contentID != 0 {
		q = q.Where("content_id = ?", contentID)
	}
	if status != "" {
		q = q.Where("status = ?", status)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if limit <= 0 {
		limit = 50
	}
	order := "scheduled_at ASC, id ASC"
	if status == model.ScheduledJobFailed ||
		status == model.ScheduledJobSucceeded ||
		status == model.ScheduledJobCancelled {
		order = "updated_at DESC, id DESC"
	}
	if err := q.Order(order).Offset(offset).Limit(limit).Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (r *GormScheduledPublishJobRepository) UpdateSchedule(
	ctx context.Context,
	id uint,
	scheduledAt time.Time,
	expectedVersion *int,
	expectedUpdatedAt *time.Time,
	publishPayload model.JSONMap,
	actorID uint,
) (*model.ScheduledPublishJob, error) {
	updates := map[string]interface{}{
		"scheduled_at":  scheduledAt,
		"updated_by":    actorID,
		"last_error":    "",
		"last_error_at": nil,
	}
	if expectedVersion != nil {
		updates["expected_version"] = expectedVersion
	}
	if expectedUpdatedAt != nil {
		updates["expected_updated_at"] = expectedUpdatedAt
	}
	if publishPayload != nil {
		updates["publish_payload"] = publishPayload
	}
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var job model.ScheduledPublishJob
		if err := tx.First(&job, id).Error; err != nil {
			return err
		}
		result := tx.Model(&model.ScheduledPublishJob{}).
			Where("id = ? AND status = ?", id, model.ScheduledJobPending).
			Updates(updates)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		return updateScheduledContentProjection(tx, job.ContentType, job.ContentID, &scheduledAt, true)
	})
	if err != nil {
		return nil, err
	}
	return r.FindByID(ctx, id)
}

func (r *GormScheduledPublishJobRepository) Cancel(
	ctx context.Context,
	id uint,
	actorID uint,
	cancelledAt time.Time,
) (*model.ScheduledPublishJob, error) {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var job model.ScheduledPublishJob
		if err := tx.First(&job, id).Error; err != nil {
			return err
		}
		result := tx.Model(&model.ScheduledPublishJob{}).
			Where("id = ? AND status = ?", id, model.ScheduledJobPending).
			Updates(map[string]interface{}{
				"status":       model.ScheduledJobCancelled,
				"updated_by":   actorID,
				"cancelled_at": cancelledAt,
				"active_key":   nil,
				"lease_until":  nil,
				"lease_token":  "",
				"claimed_at":   nil,
			})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		return updateScheduledContentProjection(tx, job.ContentType, job.ContentID, nil, true)
	})
	if err != nil {
		return nil, err
	}
	return r.FindByID(ctx, id)
}

func (r *GormScheduledPublishJobRepository) Retry(
	ctx context.Context,
	id uint,
	scheduledAt time.Time,
	expectedVersion *int,
	actorID uint,
) (*model.ScheduledPublishJob, error) {
	job, err := r.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	activeKey := string(job.ContentType) + ":" + strconv.FormatUint(uint64(job.ContentID), 10)
	updates := map[string]interface{}{
		"status":           model.ScheduledJobPending,
		"scheduled_at":     scheduledAt,
		"expected_version": expectedVersion,
		"active_key":       activeKey,
		"attempts":         0,
		"updated_by":       actorID,
		"claimed_at":       nil,
		"lease_until":      nil,
		"lease_token":      "",
		"last_attempt_at":  nil,
		"last_error":       "",
		"last_error_at":    nil,
		"failed_at":        nil,
	}
	err = r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.Model(&model.ScheduledPublishJob{}).
			Where("id = ? AND status = ?", id, model.ScheduledJobFailed).
			Updates(updates)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		return updateScheduledContentProjection(tx, job.ContentType, job.ContentID, &scheduledAt, true)
	})
	if err != nil {
		return nil, err
	}
	return r.FindByID(ctx, id)
}

func (r *GormScheduledPublishJobRepository) ClaimDue(
	ctx context.Context,
	now time.Time,
	limit int,
	leaseDuration time.Duration,
) ([]*model.ScheduledPublishJob, error) {
	if limit <= 0 {
		limit = 10
	}
	var candidates []*model.ScheduledPublishJob
	err := r.db.WithContext(ctx).
		Where("(status = ? AND attempts < max_attempts AND scheduled_at <= ?) OR (status = ? AND lease_until IS NOT NULL AND lease_until <= ?)",
			model.ScheduledJobPending, now, model.ScheduledJobRunning, now).
		Order("scheduled_at ASC, id ASC").
		Limit(limit).
		Find(&candidates).Error
	if err != nil {
		return nil, err
	}

	claimed := make([]*model.ScheduledPublishJob, 0, len(candidates))
	for _, candidate := range candidates {
		leaseUntil := now.Add(leaseDuration)
		leaseToken, err := newLeaseToken()
		if err != nil {
			return nil, err
		}
		result := r.db.WithContext(ctx).Model(&model.ScheduledPublishJob{}).
			Where("id = ?", candidate.ID).
			Where("(status = ? AND attempts < max_attempts AND scheduled_at <= ?) OR (status = ? AND lease_until IS NOT NULL AND lease_until <= ?)",
				model.ScheduledJobPending, now, model.ScheduledJobRunning, now).
			Updates(map[string]interface{}{
				"status":          model.ScheduledJobRunning,
				"attempts":        gorm.Expr("attempts + 1"),
				"claimed_at":      now,
				"lease_until":     leaseUntil,
				"lease_token":     leaseToken,
				"last_attempt_at": now,
				"last_error":      "",
				"last_error_at":   nil,
			})
		if result.Error != nil {
			return nil, result.Error
		}
		if result.RowsAffected == 0 {
			continue
		}
		job, err := r.FindByID(ctx, candidate.ID)
		if err != nil {
			return nil, err
		}
		claimed = append(claimed, job)
	}
	return claimed, nil
}

func (r *GormScheduledPublishJobRepository) MarkSucceeded(
	ctx context.Context,
	job *model.ScheduledPublishJob,
	finishedAt time.Time,
) error {
	result := r.db.WithContext(ctx).Model(&model.ScheduledPublishJob{}).
		Where("id = ? AND status = ? AND lease_token = ?", job.ID, model.ScheduledJobRunning, job.LeaseToken).
		Updates(map[string]interface{}{
			"status":        model.ScheduledJobSucceeded,
			"succeeded_at":  finishedAt,
			"active_key":    nil,
			"lease_until":   nil,
			"lease_token":   "",
			"claimed_at":    nil,
			"last_error":    "",
			"last_error_at": nil,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *GormScheduledPublishJobRepository) MarkFailed(
	ctx context.Context,
	job *model.ScheduledPublishJob,
	errText string,
	retryAt time.Time,
	failedAt time.Time,
) error {
	status := model.ScheduledJobPending
	updates := map[string]interface{}{
		"status":        status,
		"scheduled_at":  retryAt,
		"lease_until":   nil,
		"lease_token":   "",
		"claimed_at":    nil,
		"last_error":    errText,
		"last_error_at": failedAt,
	}
	if job.Attempts >= job.MaxAttempts {
		status = model.ScheduledJobFailed
		updates["status"] = status
		updates["failed_at"] = failedAt
		updates["active_key"] = nil
	}
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.Model(&model.ScheduledPublishJob{}).
			Where("id = ? AND status = ? AND lease_token = ?", job.ID, model.ScheduledJobRunning, job.LeaseToken).
			Updates(updates)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		if status == model.ScheduledJobFailed {
			return updateScheduledContentProjection(tx, job.ContentType, job.ContentID, nil, false)
		}
		return updateScheduledContentProjection(tx, job.ContentType, job.ContentID, &retryAt, false)
	})
}

func newLeaseToken() (string, error) {
	var token [16]byte
	if _, err := rand.Read(token[:]); err != nil {
		return "", fmt.Errorf("generate scheduler lease token: %w", err)
	}
	return hex.EncodeToString(token[:]), nil
}

func updateScheduledContentProjection(
	tx *gorm.DB,
	contentType model.ScheduledContentType,
	contentID uint,
	scheduledAt *time.Time,
	requireContent bool,
) error {
	var table string
	var publishedStatus interface{}
	var scheduledStatus interface{}
	var draftStatus interface{}
	switch contentType {
	case model.ScheduledContentArticle:
		table = "articles"
		publishedStatus = model.ArticleStatusPublished
		scheduledStatus = model.ArticleStatusScheduled
		draftStatus = model.ArticleStatusDraft
	case model.ScheduledContentPage:
		table = "unified_pages"
		publishedStatus = "published"
		scheduledStatus = "scheduled"
		draftStatus = "draft"
	default:
		return fmt.Errorf("unsupported content type %q", contentType)
	}

	updates := map[string]interface{}{"scheduled_at": scheduledAt}
	if scheduledAt != nil {
		updates["status"] = gorm.Expr(
			"CASE WHEN status = ? THEN status ELSE ? END",
			publishedStatus,
			scheduledStatus,
		)
	} else {
		updates["status"] = gorm.Expr(
			"CASE WHEN status = ? THEN ? ELSE status END",
			scheduledStatus,
			draftStatus,
		)
	}
	result := tx.Table(table).Where("id = ?", contentID).Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if requireContent && result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
