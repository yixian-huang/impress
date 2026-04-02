package comment

import (
	"context"
	"errors"

	"gorm.io/gorm"
)

// Repository defines the interface for comment data access
type Repository interface {
	Create(ctx context.Context, comment *Comment) error
	FindByID(ctx context.Context, id uint) (*Comment, error)
	Update(ctx context.Context, comment *Comment) error
	Delete(ctx context.Context, id uint) error
	ListByContent(ctx context.Context, contentType string, contentID uint, status CommentStatus, page, pageSize int) ([]*Comment, int64, error)
	ListAll(ctx context.Context, status string, page, pageSize int) ([]*Comment, int64, error)
	CountByContent(ctx context.Context, contentType string, contentID uint) (int64, error)
	UpdateStatus(ctx context.Context, id uint, status CommentStatus) error
	SetPinned(ctx context.Context, id uint, pinned bool) error
}

// gormRepository implements Repository using GORM
type gormRepository struct {
	db *gorm.DB
}

func newGormRepository(db *gorm.DB) Repository {
	return &gormRepository{db: db}
}

func (r *gormRepository) Create(ctx context.Context, comment *Comment) error {
	if err := comment.Validate(); err != nil {
		return err
	}
	return r.db.WithContext(ctx).Create(comment).Error
}

func (r *gormRepository) FindByID(ctx context.Context, id uint) (*Comment, error) {
	var comment Comment
	err := r.db.WithContext(ctx).Preload("Children").First(&comment, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("comment not found")
		}
		return nil, err
	}
	return &comment, nil
}

func (r *gormRepository) Update(ctx context.Context, comment *Comment) error {
	return r.db.WithContext(ctx).Save(comment).Error
}

func (r *gormRepository) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&Comment{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("comment not found")
	}
	return nil
}

func (r *gormRepository) ListByContent(ctx context.Context, contentType string, contentID uint, status CommentStatus, page, pageSize int) ([]*Comment, int64, error) {
	var total int64
	var comments []*Comment

	query := r.db.WithContext(ctx).Model(&Comment{}).
		Where("content_type = ? AND content_id = ? AND parent_id IS NULL", contentType, contentID)
	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.
		Preload("Children", func(db *gorm.DB) *gorm.DB {
			return db.Where("status = ?", CommentStatusApproved).Order("created_at ASC")
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

func (r *gormRepository) ListAll(ctx context.Context, status string, page, pageSize int) ([]*Comment, int64, error) {
	var total int64
	var comments []*Comment

	query := r.db.WithContext(ctx).Model(&Comment{})
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

func (r *gormRepository) CountByContent(ctx context.Context, contentType string, contentID uint) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&Comment{}).
		Where("content_type = ? AND content_id = ? AND status = ?", contentType, contentID, CommentStatusApproved).
		Count(&count).Error
	return count, err
}

func (r *gormRepository) UpdateStatus(ctx context.Context, id uint, status CommentStatus) error {
	return r.db.WithContext(ctx).Model(&Comment{}).Where("id = ?", id).Update("status", status).Error
}

func (r *gormRepository) SetPinned(ctx context.Context, id uint, pinned bool) error {
	return r.db.WithContext(ctx).Model(&Comment{}).Where("id = ?", id).Update("pinned", pinned).Error
}
