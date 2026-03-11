package repository

import (
	"context"
	"errors"

	"blotting-consultancy/internal/model"

	"gorm.io/gorm"
)

// GormCommentRepository implements CommentRepository using GORM
type GormCommentRepository struct {
	db *gorm.DB
}

// NewGormCommentRepository creates a new GormCommentRepository
func NewGormCommentRepository(db *gorm.DB) CommentRepository {
	return &GormCommentRepository{db: db}
}

// Create creates a new comment
func (r *GormCommentRepository) Create(ctx context.Context, comment *model.Comment) error {
	if err := comment.Validate(); err != nil {
		return err
	}
	return r.db.WithContext(ctx).Create(comment).Error
}

// FindByID finds a comment by ID with Children preloaded
func (r *GormCommentRepository) FindByID(ctx context.Context, id uint) (*model.Comment, error) {
	var comment model.Comment
	err := r.db.WithContext(ctx).Preload("Children").First(&comment, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("comment not found")
		}
		return nil, err
	}
	return &comment, nil
}

// Update updates a comment
func (r *GormCommentRepository) Update(ctx context.Context, comment *model.Comment) error {
	return r.db.WithContext(ctx).Save(comment).Error
}

// Delete deletes a comment by ID
func (r *GormCommentRepository) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&model.Comment{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("comment not found")
	}
	return nil
}

// ListByContent returns a paginated list of top-level comments for a given content target
func (r *GormCommentRepository) ListByContent(ctx context.Context, contentType string, contentID uint, status model.CommentStatus, page, pageSize int) ([]*model.Comment, int64, error) {
	var total int64
	var comments []*model.Comment

	query := r.db.WithContext(ctx).Model(&model.Comment{}).
		Where("content_type = ? AND content_id = ? AND parent_id IS NULL", contentType, contentID)
	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.
		Preload("Children", func(db *gorm.DB) *gorm.DB {
			return db.Where("status = ?", model.CommentStatusApproved).Order("created_at ASC")
		}).
		Order("pinned DESC, created_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&comments).Error
	if err != nil {
		return nil, 0, err
	}

	return comments, total, nil
}

// ListAll returns a paginated list of all comments with optional status filter
func (r *GormCommentRepository) ListAll(ctx context.Context, status string, page, pageSize int) ([]*model.Comment, int64, error) {
	var total int64
	var comments []*model.Comment

	query := r.db.WithContext(ctx).Model(&model.Comment{})
	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Order("created_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&comments).Error; err != nil {
		return nil, 0, err
	}

	return comments, total, nil
}

// CountByContent returns the count of approved comments for a given content target
func (r *GormCommentRepository) CountByContent(ctx context.Context, contentType string, contentID uint) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.Comment{}).
		Where("content_type = ? AND content_id = ? AND status = ?", contentType, contentID, model.CommentStatusApproved).
		Count(&count).Error
	return count, err
}

// UpdateStatus updates the moderation status of a comment
func (r *GormCommentRepository) UpdateStatus(ctx context.Context, id uint, status model.CommentStatus) error {
	return r.db.WithContext(ctx).Model(&model.Comment{}).Where("id = ?", id).Update("status", status).Error
}

// SetPinned sets the pinned flag on a comment
func (r *GormCommentRepository) SetPinned(ctx context.Context, id uint, pinned bool) error {
	return r.db.WithContext(ctx).Model(&model.Comment{}).Where("id = ?", id).Update("pinned", pinned).Error
}
