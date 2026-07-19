package repository

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/yixian-huang/inkless/backend/internal/model"
)

func setupAIConfigRepo(t *testing.T) (*gorm.DB, AIConfigRepository) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.AIConfig{}))

	return db, NewGormAIConfigRepository(db)
}

func TestGormAIConfigRepository_GetDefaultsToDisabled(t *testing.T) {
	_, repo := setupAIConfigRepo(t)

	config, err := repo.Get(context.Background())
	require.NoError(t, err)
	require.Equal(t, model.AIConfigSingletonID, config.ID)
	require.Equal(t, model.AIProviderDisabled, config.Provider)
	require.False(t, config.HasAPIKey())
}

func TestGormAIConfigRepository_UpsertUsesSingletonRow(t *testing.T) {
	db, repo := setupAIConfigRepo(t)
	ctx := context.Background()

	require.NoError(t, repo.Upsert(ctx, &model.AIConfig{
		ID:               99,
		Provider:         model.AIProviderOpenAI,
		APIKeyCiphertext: "ciphertext",
		BaseURL:          "https://api.example.test",
		Model:            "model-a",
	}))
	require.NoError(t, repo.Upsert(ctx, &model.AIConfig{
		ID:               42,
		Provider:         model.AIProviderAnthropic,
		APIKeyCiphertext: "ciphertext-2",
		Model:            "model-b",
	}))

	config, err := repo.Get(ctx)
	require.NoError(t, err)
	require.Equal(t, model.AIConfigSingletonID, config.ID)
	require.Equal(t, model.AIProviderAnthropic, config.Provider)
	require.Equal(t, "ciphertext-2", config.APIKeyCiphertext)
	require.Equal(t, "model-b", config.Model)

	var count int64
	require.NoError(t, db.Model(&model.AIConfig{}).Count(&count).Error)
	require.Equal(t, int64(1), count)
}

func TestGormAIConfigRepository_RejectsInvalidProvider(t *testing.T) {
	_, repo := setupAIConfigRepo(t)

	err := repo.Upsert(context.Background(), &model.AIConfig{Provider: "bad"})
	require.Error(t, err)
}
