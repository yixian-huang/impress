package repository

import (
	"context"

	"blotting-consultancy/internal/model"
)

// QALogRepository defines the interface for Q&A log data access.
type QALogRepository interface {
	Create(ctx context.Context, log *model.QALog) error
	FindByID(ctx context.Context, id uint) (*model.QALog, error)
	List(ctx context.Context, offset, limit int) ([]*model.QALog, int64, error)
	UpdateRating(ctx context.Context, id uint, rating model.QAFeedback) error
}
