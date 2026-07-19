package repository

import (
	"context"
	"errors"
	"time"

	"github.com/yixian-huang/inkless/backend/internal/model"
	"gorm.io/gorm"
)

// GormRefreshTokenRepository implements RefreshTokenRepository using GORM
type GormRefreshTokenRepository struct {
	db *gorm.DB
}

// NewGormRefreshTokenRepository creates a new GormRefreshTokenRepository
func NewGormRefreshTokenRepository(db *gorm.DB) RefreshTokenRepository {
	return &GormRefreshTokenRepository{db: db}
}

// Create creates a new refresh token
func (r *GormRefreshTokenRepository) Create(ctx context.Context, token *model.RefreshToken) error {
	if err := token.Validate(); err != nil {
		return err
	}
	return r.db.WithContext(ctx).Create(token).Error
}

// FindByToken finds a refresh token by token string
func (r *GormRefreshTokenRepository) FindByToken(ctx context.Context, token string) (*model.RefreshToken, error) {
	var refreshToken model.RefreshToken
	err := r.db.WithContext(ctx).
		Preload("User").
		Where("token = ?", token).
		First(&refreshToken).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("refresh token not found")
		}
		return nil, err
	}
	return &refreshToken, nil
}

// FindByUserID finds all refresh tokens for a user
func (r *GormRefreshTokenRepository) FindByUserID(ctx context.Context, userID uint) ([]*model.RefreshToken, error) {
	var tokens []*model.RefreshToken
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&tokens).Error
	if err != nil {
		return nil, err
	}
	return tokens, nil
}

// Delete deletes a refresh token by ID
func (r *GormRefreshTokenRepository) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&model.RefreshToken{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("refresh token not found")
	}
	return nil
}

// DeleteByToken deletes a refresh token by token string
func (r *GormRefreshTokenRepository) DeleteByToken(ctx context.Context, token string) error {
	result := r.db.WithContext(ctx).Where("token = ?", token).Delete(&model.RefreshToken{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("refresh token not found")
	}
	return nil
}

// DeleteByUserID deletes all refresh tokens for a user
func (r *GormRefreshTokenRepository) DeleteByUserID(ctx context.Context, userID uint) error {
	return r.db.WithContext(ctx).Where("user_id = ?", userID).Delete(&model.RefreshToken{}).Error
}

// DeleteExpired deletes all expired tokens
func (r *GormRefreshTokenRepository) DeleteExpired(ctx context.Context, before time.Time) error {
	return r.db.WithContext(ctx).Where("expires_at < ?", before).Delete(&model.RefreshToken{}).Error
}
