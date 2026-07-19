package db_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/yixian-huang/inkless/backend/internal/db"
	"github.com/yixian-huang/inkless/backend/pkg/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm/logger"
)

// TestIntegrationWithConfig tests database initialization using config package
func TestIntegrationWithConfig(t *testing.T) {
	// Set up test environment variables
	os.Setenv("DB_DSN", ":memory:")
	os.Setenv("JWT_SECRET", "test-secret")
	os.Setenv("JWT_REFRESH_SECRET", "test-refresh-secret")
	os.Setenv("ENV", "test")
	defer func() {
		os.Unsetenv("DB_DSN")
		os.Unsetenv("JWT_SECRET")
		os.Unsetenv("JWT_REFRESH_SECRET")
		os.Unsetenv("ENV")
	}()

	// Load config
	cfg, err := config.Load()
	require.NoError(t, err)
	assert.Equal(t, ":memory:", cfg.DBDSN)

	// Initialize database with config
	database, err := db.Init(db.InitOptions{
		DSN:         cfg.DBDSN,
		MaxOpenConn: 10,
		MaxIdleConn: 5,
		MaxLifetime: 30 * time.Minute,
		LogLevel:    logger.Silent,
	})
	require.NoError(t, err)
	require.NotNil(t, database)
	defer database.Close()

	// Health check
	ctx := context.Background()
	err = database.HealthCheck(ctx)
	assert.NoError(t, err)
}

// TestDatabaseWorkflow tests a complete database workflow
func TestDatabaseWorkflow(t *testing.T) {
	// Initialize database
	database, err := db.Init(db.InitOptions{
		DSN:      ":memory:",
		LogLevel: logger.Silent,
	})
	require.NoError(t, err)
	defer database.Close()

	// Health check
	ctx := context.Background()
	err = database.HealthCheck(ctx)
	require.NoError(t, err)

	// Test model for workflow
	type Article struct {
		ID      uint   `gorm:"primaryKey"`
		Title   string `gorm:"not null"`
		Content string
	}

	// Create migrator and run auto-migration
	migrator := db.NewMigrator(database)
	err = migrator.AutoMigrate(&Article{})
	require.NoError(t, err)

	// Verify table exists
	hasTable := migrator.HasTable(&Article{})
	assert.True(t, hasTable)

	// Insert test data
	article := Article{
		Title:   "Test Article",
		Content: "This is test content",
	}
	err = database.Create(&article).Error
	require.NoError(t, err)
	assert.NotZero(t, article.ID)

	// Query data
	var retrieved Article
	err = database.First(&retrieved, article.ID).Error
	require.NoError(t, err)
	assert.Equal(t, "Test Article", retrieved.Title)
	assert.Equal(t, "This is test content", retrieved.Content)
}
