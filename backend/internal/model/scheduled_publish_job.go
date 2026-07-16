package model

import (
	"errors"
	"strconv"
	"time"

	"gorm.io/gorm"
)

type ScheduledContentType string

const (
	ScheduledContentArticle ScheduledContentType = "article"
	ScheduledContentPage    ScheduledContentType = "page"
)

type ScheduledJobStatus string

const (
	ScheduledJobPending   ScheduledJobStatus = "pending"
	ScheduledJobRunning   ScheduledJobStatus = "running"
	ScheduledJobSucceeded ScheduledJobStatus = "succeeded"
	ScheduledJobFailed    ScheduledJobStatus = "failed"
	ScheduledJobCancelled ScheduledJobStatus = "cancelled"
)

// ScheduledPublishJob is the durable queue entry for scheduled publishing.
type ScheduledPublishJob struct {
	ID                uint                 `gorm:"primaryKey" json:"id"`
	ContentType       ScheduledContentType `gorm:"size:20;not null;index:idx_sched_content" json:"resourceType"`
	ContentID         uint                 `gorm:"not null;index:idx_sched_content" json:"resourceId"`
	ActiveKey         *string              `gorm:"size:80;uniqueIndex" json:"-"`
	Status            ScheduledJobStatus   `gorm:"size:20;not null;default:'pending';index" json:"status"`
	ScheduledAt       time.Time            `gorm:"not null;index" json:"scheduledAt"`
	ExpectedVersion   *int                 `json:"expectedVersion,omitempty"`
	ExpectedUpdatedAt *time.Time           `json:"expectedUpdatedAt,omitempty"`
	PublishPayload    JSONMap              `gorm:"type:jsonb" json:"-"`
	Attempts          int                  `gorm:"not null;default:0" json:"attempts"`
	MaxAttempts       int                  `gorm:"not null;default:3" json:"maxAttempts"`
	ClaimedAt         *time.Time           `json:"claimedAt"`
	LeaseUntil        *time.Time           `gorm:"index" json:"leaseUntil"`
	LeaseToken        string               `gorm:"size:64;index" json:"-"`
	LastAttemptAt     *time.Time           `json:"lastAttemptAt"`
	LastError         string               `gorm:"type:text" json:"lastError"`
	LastErrorAt       *time.Time           `json:"lastErrorAt"`
	SucceededAt       *time.Time           `json:"succeededAt"`
	FailedAt          *time.Time           `json:"failedAt"`
	CancelledAt       *time.Time           `json:"cancelledAt"`
	CreatedBy         uint                 `gorm:"index" json:"createdBy"`
	UpdatedBy         uint                 `gorm:"index" json:"updatedBy"`
	CreatedAt         time.Time            `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt         time.Time            `gorm:"autoUpdateTime" json:"updatedAt"`
}

func (ScheduledPublishJob) TableName() string { return "scheduled_publish_jobs" }

func (j *ScheduledPublishJob) Validate() error {
	if j.ContentType != ScheduledContentArticle && j.ContentType != ScheduledContentPage {
		return errors.New("contentType must be article or page")
	}
	if j.ContentID == 0 {
		return errors.New("contentId is required")
	}
	if j.Status == "" {
		j.Status = ScheduledJobPending
	}
	switch j.Status {
	case ScheduledJobPending, ScheduledJobRunning, ScheduledJobSucceeded, ScheduledJobFailed, ScheduledJobCancelled:
	default:
		return errors.New("invalid scheduled job status")
	}
	if j.ScheduledAt.IsZero() {
		return errors.New("scheduledAt is required")
	}
	if j.MaxAttempts == 0 {
		j.MaxAttempts = 3
	}
	if j.Status == ScheduledJobPending || j.Status == ScheduledJobRunning {
		activeKey := string(j.ContentType) + ":" + strconv.FormatUint(uint64(j.ContentID), 10)
		j.ActiveKey = &activeKey
	}
	return nil
}

func (j *ScheduledPublishJob) BeforeCreate(tx *gorm.DB) error {
	return j.Validate()
}
