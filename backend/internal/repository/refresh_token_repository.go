package repository

import (
	"context"
	"time"

	"github.com/yixian-huang/inkless/backend/internal/model"
)

// RefreshTokenRepository defines the interface for refresh token data access
type RefreshTokenRepository interface {
	// Create creates a new refresh token
	Create(ctx context.Context, token *model.RefreshToken) error

	// FindByToken finds a refresh token by token string
	FindByToken(ctx context.Context, token string) (*model.RefreshToken, error)

	// FindByUserID finds all refresh tokens for a user
	FindByUserID(ctx context.Context, userID uint) ([]*model.RefreshToken, error)

	// Delete deletes a refresh token by ID
	Delete(ctx context.Context, id uint) error

	// DeleteByToken deletes a refresh token by token string
	DeleteByToken(ctx context.Context, token string) error

	// DeleteByUserID deletes all refresh tokens for a user
	DeleteByUserID(ctx context.Context, userID uint) error

	// DeleteExpired deletes all expired tokens
	DeleteExpired(ctx context.Context, before time.Time) error
}
