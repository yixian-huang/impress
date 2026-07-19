package db

import (
	"testing"

	"github.com/yixian-huang/inkless/backend/internal/model"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestPreserveLegacyMeilisearchIndexPrefixMigration(t *testing.T) {
	database, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, database.AutoMigrate(&model.Plugin{}))

	legacy := model.Plugin{
		PluginID: "mls-search",
		Name:     "Meilisearch",
		Version:  "1.0.0",
		Settings: model.JSONMap{"host": "http://legacy:7700"},
	}
	custom := model.Plugin{
		PluginID: "mls-search-custom",
		Name:     "Other search",
		Version:  "1.0.0",
		Settings: model.JSONMap{"host": "http://other:7700"},
	}
	require.NoError(t, database.Create(&legacy).Error)
	require.NoError(t, database.Create(&custom).Error)

	var migration Migration
	for _, candidate := range DataMigrations() {
		if candidate.ID == "004_preserve_legacy_meilisearch_index_prefix" {
			migration = candidate
			break
		}
	}
	require.NotNil(t, migration.Up)
	require.NoError(t, migration.Up(database))
	require.NoError(t, migration.Up(database), "migration must be idempotent")

	var migrated model.Plugin
	require.NoError(t, database.Where("plugin_id = ?", "mls-search").First(&migrated).Error)
	require.Equal(t, "http://legacy:7700", migrated.Settings["host"])
	require.Equal(t, "impress_", migrated.Settings["index_prefix"])

	var untouched model.Plugin
	require.NoError(t, database.Where("plugin_id = ?", "mls-search-custom").First(&untouched).Error)
	require.NotContains(t, untouched.Settings, "index_prefix")
}
