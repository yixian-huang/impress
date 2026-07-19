package db

import (
	"github.com/yixian-huang/inkless/backend/internal/model"

	"gorm.io/gorm"
)

// DataMigrations returns the list of data migrations to run after AutoMigrate
func DataMigrations() []Migration {
	return []Migration{
		{
			ID: "001_migrate_category_id_to_article_categories",
			Up: func(db *gorm.DB) error {
				// Copy existing category_id relationships to the article_categories join table
				// Only run if the join table is empty (first migration)
				var count int64
				if err := db.Table("article_categories").Count(&count).Error; err != nil {
					// Table might not exist yet, skip
					return nil
				}
				if count > 0 {
					return nil // Already has data, skip
				}

				return db.Exec(
					"INSERT INTO article_categories (article_id, category_id) SELECT id, category_id FROM articles WHERE category_id IS NOT NULL",
				).Error
			},
			Down: func(db *gorm.DB) error {
				return db.Exec("DELETE FROM article_categories").Error
			},
		},
		{
			ID: "002_set_admin_super_admin",
			Up: func(db *gorm.DB) error {
				return db.Exec("UPDATE users SET is_super_admin = ? WHERE username = ?", true, "admin").Error
			},
			Down: func(db *gorm.DB) error {
				return db.Exec("UPDATE users SET is_super_admin = ? WHERE username = ?", false, "admin").Error
			},
		},
		{
			ID: "003_create_ai_configs",
			Up: func(db *gorm.DB) error {
				return db.AutoMigrate(&model.AIConfig{})
			},
			Down: func(db *gorm.DB) error {
				return db.Migrator().DropTable(&model.AIConfig{})
			},
		},
		{
			ID: "004_preserve_legacy_meilisearch_index_prefix",
			Up: func(db *gorm.DB) error {
				if !db.Migrator().HasTable(&model.Plugin{}) {
					return nil
				}

				var plugins []model.Plugin
				if err := db.Where("plugin_id = ?", "mls-search").Find(&plugins).Error; err != nil {
					return err
				}
				for i := range plugins {
					settings := plugins[i].Settings
					if settings == nil {
						settings = make(model.JSONMap)
					}
					if _, configured := settings["index_prefix"]; configured {
						continue
					}
					settings["index_prefix"] = "impress_"
					if err := db.Model(&model.Plugin{}).
						Where("id = ?", plugins[i].ID).
						Update("settings", settings).Error; err != nil {
						return err
					}
				}
				return nil
			},
			// Removing the compatibility marker on rollback could make an existing
			// deployment silently switch to a different index namespace.
			Down: func(*gorm.DB) error { return nil },
		},
	}
}
