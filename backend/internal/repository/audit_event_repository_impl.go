package repository

import (
	"context"
	"time"

	"github.com/yixian-huang/inkless/backend/internal/model"

	"gorm.io/gorm"
)

// GormAuditEventRepository implements AuditEventRepository using GORM
type GormAuditEventRepository struct {
	db *gorm.DB
}

// NewGormAuditEventRepository creates a new GormAuditEventRepository
func NewGormAuditEventRepository(db *gorm.DB) AuditEventRepository {
	return &GormAuditEventRepository{db: db}
}

// Create creates a new audit event record
func (r *GormAuditEventRepository) Create(ctx context.Context, event *model.AuditEvent) error {
	return r.db.WithContext(ctx).Create(event).Error
}

// List returns a filtered, paginated list of audit events ordered by creation time (newest first)
func (r *GormAuditEventRepository) List(ctx context.Context, offset, limit int, action, actor string, from, to *time.Time) ([]model.AuditEvent, int64, error) {
	query := r.db.WithContext(ctx).Model(&model.AuditEvent{})

	if action != "" {
		query = query.Where("action = ?", action)
	}
	if actor != "" {
		query = query.Where("actor = ?", actor)
	}
	if from != nil {
		query = query.Where("created_at >= ?", *from)
	}
	if to != nil {
		query = query.Where("created_at <= ?", *to)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var items []model.AuditEvent
	if err := query.
		Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&items).Error; err != nil {
		return nil, 0, err
	}

	return items, total, nil
}
