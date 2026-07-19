package service

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/yixian-huang/inkless/backend/internal/model"
	"github.com/yixian-huang/inkless/backend/internal/provider"
	"github.com/yixian-huang/inkless/backend/internal/repository"
	"github.com/yixian-huang/inkless/backend/pkg/secretcipher"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func setupAIConfigService(t *testing.T) (*AIConfigService, repository.AIConfigRepository, *provider.Registry, *secretcipher.Cipher) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.AIConfig{}))

	repo := repository.NewGormAIConfigRepository(db)
	registry := provider.NewRegistry()
	cipher, err := secretcipher.New("server-secret")
	require.NoError(t, err)
	return NewAIConfigService(repo, cipher, registry), repo, registry, cipher
}

func setupAIConfigServiceWithClient(t *testing.T, client *http.Client) (*AIConfigService, repository.AIConfigRepository, *provider.Registry, *secretcipher.Cipher) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.AIConfig{}))

	repo := repository.NewGormAIConfigRepository(db)
	registry := provider.NewRegistry()
	cipher, err := secretcipher.New("server-secret")
	require.NoError(t, err)
	return NewAIConfigService(repo, cipher, registry, WithAIConfigHTTPClient(client)), repo, registry, cipher
}

func TestAIConfigService_UpdateEncryptsMasksAndSwitchesRegistry(t *testing.T) {
	svc, repo, registry, _ := setupAIConfigService(t)

	resp, err := svc.Update(context.Background(), AIConfigInput{
		Provider: "openai",
		APIKey:   "sk-secret-1234",
		BaseURL:  "https://api.example.test",
		Model:    "gpt-test",
	})
	require.NoError(t, err)
	require.Equal(t, model.AIProviderOpenAI, resp.Provider)
	require.True(t, resp.Enabled)
	require.True(t, resp.HasAPIKey)
	require.Equal(t, "sk-s...1234", resp.APIKeyMasked)
	require.Equal(t, "gpt-test", resp.Model)

	config, err := repo.Get(context.Background())
	require.NoError(t, err)
	require.NotContains(t, config.APIKeyCiphertext, "sk-secret-1234")
	require.NotEmpty(t, config.APIKeyCiphertext)
	require.Equal(t, "openai", registry.AI().Name())
}

func TestAIConfigService_UpdatePreservesExistingAPIKey(t *testing.T) {
	svc, repo, _, cipher := setupAIConfigService(t)
	ctx := context.Background()

	_, err := svc.Update(ctx, AIConfigInput{
		Provider: "openai",
		APIKey:   "sk-secret-1234",
		Model:    "gpt-a",
	})
	require.NoError(t, err)
	before, err := repo.Get(ctx)
	require.NoError(t, err)

	resp, err := svc.Update(ctx, AIConfigInput{
		Provider: "openai",
		Model:    "gpt-b",
	})
	require.NoError(t, err)
	require.Equal(t, "gpt-b", resp.Model)
	require.Equal(t, "sk-s...1234", resp.APIKeyMasked)

	after, err := repo.Get(ctx)
	require.NoError(t, err)
	require.Equal(t, before.APIKeyCiphertext, after.APIKeyCiphertext)
	decrypted, err := cipher.Decrypt(after.APIKeyCiphertext)
	require.NoError(t, err)
	require.Equal(t, "sk-secret-1234", decrypted)
}

func TestAIConfigService_UpdateDisabledClearsKeyAndRegistry(t *testing.T) {
	svc, repo, registry, _ := setupAIConfigService(t)
	ctx := context.Background()

	_, err := svc.Update(ctx, AIConfigInput{Provider: "openai", APIKey: "sk-secret-1234"})
	require.NoError(t, err)
	resp, err := svc.Update(ctx, AIConfigInput{Provider: "disabled"})
	require.NoError(t, err)

	require.Equal(t, model.AIProviderDisabled, resp.Provider)
	require.False(t, resp.Enabled)
	require.False(t, resp.HasAPIKey)
	require.Equal(t, "noop", registry.AI().Name())

	config, err := repo.Get(ctx)
	require.NoError(t, err)
	require.Empty(t, config.APIKeyCiphertext)
}

func TestAIConfigService_RestoreAppliesPersistedProvider(t *testing.T) {
	svc, repo, registry, cipher := setupAIConfigService(t)
	ciphertext, err := cipher.Encrypt("anthropic-key")
	require.NoError(t, err)
	require.NoError(t, repo.Upsert(context.Background(), &model.AIConfig{
		Provider:         model.AIProviderAnthropic,
		APIKeyCiphertext: ciphertext,
		Model:            "claude-test",
	}))

	require.NoError(t, svc.Restore(context.Background()))
	require.Equal(t, "anthropic", registry.AI().Name())
}

func TestAIConfigService_TestCurrentUsesProviderHealthCall(t *testing.T) {
	var gotAuth string
	client := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		gotAuth = r.Header.Get("Authorization")
		require.Equal(t, "/chat/completions", r.URL.Path)

		body, err := json.Marshal(map[string]interface{}{
			"choices": []map[string]interface{}{
				{
					"message":       map[string]string{"content": "ok"},
					"finish_reason": "stop",
				},
			},
			"model": "gpt-health",
			"usage": map[string]int{
				"prompt_tokens":     1,
				"completion_tokens": 1,
			},
		})
		require.NoError(t, err)
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body:       io.NopCloser(strings.NewReader(string(body))),
		}, nil
	})}

	svc, _, _, _ := setupAIConfigServiceWithClient(t, client)
	_, err := svc.Update(context.Background(), AIConfigInput{
		Provider: "openai",
		APIKey:   "sk-health",
		BaseURL:  "https://api.example.test",
		Model:    "gpt-health",
	})
	require.NoError(t, err)

	resp, err := svc.TestCurrent(context.Background())
	require.NoError(t, err)
	require.True(t, resp.Healthy)
	require.Equal(t, "openai", resp.Provider)
	require.Equal(t, "gpt-health", resp.Model)
	require.Equal(t, "ok", resp.Message)
	require.Equal(t, "Bearer sk-health", gotAuth)
}

func TestAIConfigService_TestUsesExistingKeyWithoutPersistingCandidate(t *testing.T) {
	client := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		require.Equal(t, "Bearer sk-existing", r.Header.Get("Authorization"))
		body, err := json.Marshal(map[string]interface{}{
			"choices": []map[string]interface{}{
				{"message": map[string]string{"content": "ok"}, "finish_reason": "stop"},
			},
			"model": "gpt-candidate",
		})
		require.NoError(t, err)
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body:       io.NopCloser(strings.NewReader(string(body))),
		}, nil
	})}

	svc, repo, _, _ := setupAIConfigServiceWithClient(t, client)
	ctx := context.Background()
	_, err := svc.Update(ctx, AIConfigInput{Provider: "openai", APIKey: "sk-existing", Model: "gpt-saved"})
	require.NoError(t, err)

	resp, err := svc.Test(ctx, AIConfigInput{Provider: "openai", BaseURL: "https://api.example.test", Model: "gpt-candidate"})
	require.NoError(t, err)
	require.True(t, resp.Healthy)
	require.Equal(t, "gpt-candidate", resp.Model)

	config, err := repo.Get(ctx)
	require.NoError(t, err)
	require.Equal(t, "gpt-saved", config.Model)
}

func TestAIConfigService_UpdateRejectsMissingKeyForEnabledProvider(t *testing.T) {
	svc, _, _, _ := setupAIConfigService(t)

	_, err := svc.Update(context.Background(), AIConfigInput{Provider: "openai"})
	require.ErrorIs(t, err, ErrAIAPIKeyRequired)
}

func TestAIConfigService_UpdateRejectsUnsupportedProvider(t *testing.T) {
	svc, _, _, _ := setupAIConfigService(t)

	_, err := svc.Update(context.Background(), AIConfigInput{Provider: "other", APIKey: "key"})
	require.ErrorIs(t, err, ErrAIUnsupportedConfig)
}
