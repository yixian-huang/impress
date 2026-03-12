package repository

import (
	"context"
	"errors"

	"blotting-consultancy/internal/model"

	"gorm.io/gorm"
)

// GormQALogRepository implements QALogRepository using GORM.
type GormQALogRepository struct {
	db *gorm.DB
}

// NewGormQALogRepository creates a new GormQALogRepository.
func NewGormQALogRepository(db *gorm.DB) QALogRepository {
	return &GormQALogRepository{db: db}
}

// Create creates a new Q&A log entry.
func (r *GormQALogRepository) Create(ctx context.Context, log *model.QALog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

// FindByID finds a Q&A log entry by ID.
func (r *GormQALogRepository) FindByID(ctx context.Context, id uint) (*model.QALog, error) {
	var log model.QALog
	err := r.db.WithContext(ctx).First(&log, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("qa log not found")
		}
		return nil, err
	}
	return &log, nil
}

// List returns paginated Q&A log entries, ordered by created_at DESC.
func (r *GormQALogRepository) List(ctx context.Context, offset, limit int) ([]*model.QALog, int64, error) {
	var logs []*model.QALog
	var total int64

	query := r.db.WithContext(ctx).Model(&model.QALog{})

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

// UpdateRating updates the rating of a Q&A log entry.
func (r *GormQALogRepository) UpdateRating(ctx context.Context, id uint, rating model.QAFeedback) error {
	result := r.db.WithContext(ctx).
		Model(&model.QALog{}).
		Where("id = ?", id).
		Update("rating", rating)

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("qa log not found")
	}
	return nil
}
