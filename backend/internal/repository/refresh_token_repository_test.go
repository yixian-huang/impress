package repository

import (
	"context"
	"testing"
	"time"

	"github.com/yixian-huang/inkless/backend/internal/model"
)

func TestRefreshTokenRepository_CRUD(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	userRepo := NewGormUserRepository(database.DB)
	tokenRepo := NewGormRefreshTokenRepository(database.DB)
	ctx := context.Background()

	// Create user first
	user := &model.User{
		Username:     "testuser",
		PasswordHash: "hashedpassword",
		Role:         model.RoleAdmin,
	}
	_ = userRepo.Create(ctx, user)

	token := &model.RefreshToken{
		UserID:    user.ID,
		Token:     "test-token-123",
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	// Test Create
	err := tokenRepo.Create(ctx, token)
	if err != nil {
		t.Fatalf("Failed to create refresh token: %v", err)
	}

	// Test FindByToken
	found, err := tokenRepo.FindByToken(ctx, "test-token-123")
	if err != nil {
		t.Fatalf("Failed to find token: %v", err)
	}

	if found.UserID != user.ID {
		t.Errorf("Expected user ID %d, got %d", user.ID, found.UserID)
	}

	// Test FindByUserID
	tokens, _ := tokenRepo.FindByUserID(ctx, user.ID)
	if len(tokens) != 1 {
		t.Errorf("Expected 1 token, got %d", len(tokens))
	}

	// Test DeleteByToken
	err = tokenRepo.DeleteByToken(ctx, "test-token-123")
	if err != nil {
		t.Fatalf("Failed to delete token: %v", err)
	}

	_, err = tokenRepo.FindByToken(ctx, "test-token-123")
	if err == nil {
		t.Error("Expected error when finding deleted token")
	}
}

func TestRefreshTokenRepository_DeleteExpired(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	userRepo := NewGormUserRepository(database.DB)
	tokenRepo := NewGormRefreshTokenRepository(database.DB)
	ctx := context.Background()

	user := &model.User{
		Username:     "testuser",
		PasswordHash: "hashedpassword",
		Role:         model.RoleAdmin,
	}
	_ = userRepo.Create(ctx, user)

	// Create expired token
	expiredToken := &model.RefreshToken{
		UserID:    user.ID,
		Token:     "expired-token",
		ExpiresAt: time.Now().Add(-24 * time.Hour),
	}
	_ = tokenRepo.Create(ctx, expiredToken)

	// Create valid token
	validToken := &model.RefreshToken{
		UserID:    user.ID,
		Token:     "valid-token",
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	_ = tokenRepo.Create(ctx, validToken)

	// Delete expired tokens
	err := tokenRepo.DeleteExpired(ctx, time.Now())
	if err != nil {
		t.Fatalf("Failed to delete expired tokens: %v", err)
	}

	// Verify expired token is deleted
	_, err = tokenRepo.FindByToken(ctx, "expired-token")
	if err == nil {
		t.Error("Expected error when finding expired token")
	}

	// Verify valid token still exists
	_, err = tokenRepo.FindByToken(ctx, "valid-token")
	if err != nil {
		t.Error("Expected valid token to still exist")
	}
}

func TestRefreshTokenRepository_CascadeDelete(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	userRepo := NewGormUserRepository(database.DB)
	tokenRepo := NewGormRefreshTokenRepository(database.DB)
	ctx := context.Background()

	user := &model.User{
		Username:     "testuser",
		PasswordHash: "hashedpassword",
		Role:         model.RoleAdmin,
	}
	_ = userRepo.Create(ctx, user)

	token := &model.RefreshToken{
		UserID:    user.ID,
		Token:     "test-token-123",
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	_ = tokenRepo.Create(ctx, token)

	// Delete user (should cascade to tokens)
	err := userRepo.Delete(ctx, user.ID)
	if err != nil {
		t.Fatalf("Failed to delete user: %v", err)
	}

	// Verify token is also deleted
	_, err = tokenRepo.FindByToken(ctx, "test-token-123")
	if err == nil {
		t.Error("Expected token to be cascade deleted with user")
	}
}
