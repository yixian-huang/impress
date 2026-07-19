package auth

import (
	"github.com/yixian-huang/inkless/backend/internal/repository"
	"github.com/yixian-huang/inkless/backend/pkg/config"
)

// Handler handles auth-related HTTP requests
type Handler struct {
	userRepo         repository.UserRepository
	refreshTokenRepo repository.RefreshTokenRepository
	config           *config.Config
}

// NewHandler creates a new auth handler
func NewHandler(
	userRepo repository.UserRepository,
	refreshTokenRepo repository.RefreshTokenRepository,
	config *config.Config,
) *Handler {
	return &Handler{
		userRepo:         userRepo,
		refreshTokenRepo: refreshTokenRepo,
		config:           config,
	}
}
