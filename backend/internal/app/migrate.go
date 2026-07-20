package app

import (
	"fmt"

	"github.com/pressly/goose/v3"
	"gorm.io/gorm/logger"

	"github.com/yixian-huang/inkless/backend/internal/db"
	"github.com/yixian-huang/inkless/backend/internal/db/migrations"
	"github.com/yixian-huang/inkless/backend/internal/model"
	commentMod "github.com/yixian-huang/inkless/backend/internal/modules/comment"
	appLogger "github.com/yixian-huang/inkless/backend/pkg/logger"
)

// migrateSchema runs GORM AutoMigrate, data migrations, and goose SQL/Go migrations.
// Kept out of New so the composition root reads as wire → serve.
func migrateSchema(database *db.DB, log *appLogger.Logger) error {
	// One-shot legacy index cleanup (idempotent). Prefer a goose migration for new DDL.
	if err := database.DB.Exec("DROP INDEX IF EXISTS idx_page_version").Error; err != nil {
		log.Warn("legacy index drop failed", "index", "idx_page_version", "error", err)
	}

	migrator := db.NewMigrator(database)
	if err := migrator.AutoMigrate(autoMigrateModels()...); err != nil {
		return fmt.Errorf("auto migrate: %w", err)
	}

	if err := migrator.RunMigrations(db.DataMigrations()); err != nil {
		return fmt.Errorf("data migrations: %w", err)
	}

	sqlDB, err := database.DB.DB()
	if err != nil {
		return fmt.Errorf("sql.DB for goose: %w", err)
	}
	goose.SetBaseFS(migrations.EmbedMigrations)
	dialect := "sqlite3"
	if database.DB.Dialector.Name() == "postgres" {
		dialect = "postgres"
	}
	if err := goose.SetDialect(dialect); err != nil {
		return fmt.Errorf("goose dialect: %w", err)
	}
	migrations.Dialect = dialect
	if err := goose.Up(sqlDB, "."); err != nil {
		return fmt.Errorf("goose up: %w", err)
	}
	log.Info("Goose migrations applied successfully")
	log.Info("Database migrations completed")
	return nil
}

func autoMigrateModels() []interface{} {
	return []interface{}{
		&model.User{},
		&model.RefreshToken{},
		&model.APIKey{},
		&model.ContentDocument{},
		&model.Media{},
		&model.PageView{},
		&model.Category{},
		&model.Tag{},
		&model.Article{},
		&model.ArticleVersion{},
		&model.BackupRecord{},
		&model.AuditEvent{},
		&model.Page{},
		&model.InstalledTheme{},
		&model.MenuGroup{},
		&model.MenuItem{},
		&model.RBACRole{},
		&model.Permission{},
		&model.UserRole{},
		&model.MarketplaceItem{},
		&model.MarketplaceVersion{},
		&model.MediaFolder{},
		&model.ChunkedUpload{},
		&model.Glossary{},
		&model.StorageConfig{},
		&model.AIConfig{},
		&model.UnifiedPage{},
		&model.PageVersion{},
		&model.ScheduledPublishJob{},
		&model.PageTemplate{},
		&model.SiteConfig{},
		&model.Plugin{},
		&model.PluginSetting{},
		&commentMod.Comment{},
	}
}

// gormLogLevel maps app env to GORM logger verbosity.
func gormLogLevel(env string) logger.LogLevel {
	if env == "production" {
		return logger.Warn
	}
	return logger.Info
}
