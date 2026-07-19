package repository

import (
	"testing"

	"github.com/yixian-huang/inkless/backend/internal/db"
	"github.com/yixian-huang/inkless/backend/internal/model"
	"gorm.io/gorm/logger"
)

// setupTestDB creates a test database for repository tests
func setupTestDB(t *testing.T) *db.DB {
	t.Helper()

	opts := db.InitOptions{
		DSN:      ":memory:",
		LogLevel: logger.Silent,
	}

	database, err := db.Init(opts)
	if err != nil {
		t.Fatalf("Failed to initialize test database: %v", err)
	}

	// Enable foreign keys for SQLite
	database.Exec("PRAGMA foreign_keys = ON")

	// Run migrations
	migrator := db.NewMigrator(database)
	if err := migrator.AutoMigrate(
		&model.User{},
		&model.RefreshToken{},
		&model.ContentDocument{},
		&model.ContentVersion{},
	); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	return database
}
