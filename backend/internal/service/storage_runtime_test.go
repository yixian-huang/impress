package service

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yixian-huang/inkless/backend/internal/model"
	"github.com/yixian-huang/inkless/backend/internal/provider"
	"github.com/yixian-huang/inkless/backend/pkg/brandcompat"
	"github.com/yixian-huang/inkless/backend/pkg/secretcipher"
)

type fakeStorageConfigRepo struct {
	config      *model.StorageConfig
	upsertCount int
}

func (r *fakeStorageConfigRepo) Get(ctx context.Context) (*model.StorageConfig, error) {
	if r.config == nil {
		return &model.StorageConfig{ID: 1, Strategy: model.StorageLocal}, nil
	}
	cp := *r.config
	return &cp, nil
}

func (r *fakeStorageConfigRepo) Upsert(ctx context.Context, config *model.StorageConfig) error {
	cp := *config
	r.config = &cp
	r.upsertCount++
	return nil
}

type fakeStorageProvider struct {
	existsErr error
	existing  map[string]bool
	deleted   []string
	savedKey  string
	savedData []byte
}

func (p *fakeStorageProvider) Save(ctx context.Context, filename string, reader io.Reader, size int64) (string, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return "", err
	}
	p.savedData = data
	if p.savedKey != "" {
		return p.savedKey, nil
	}
	return filename, nil
}

func (p *fakeStorageProvider) Get(ctx context.Context, key string) (io.ReadCloser, error) {
	return io.NopCloser(bytes.NewReader(p.savedData)), nil
}

func (p *fakeStorageProvider) Delete(ctx context.Context, key string) error {
	p.deleted = append(p.deleted, key)
	return nil
}
func (p *fakeStorageProvider) URL(key string) string { return "https://cdn.test/" + key }

func (p *fakeStorageProvider) Exists(ctx context.Context, key string) (bool, error) {
	return p.existing[key], p.existsErr
}

func TestRestoreStartupConfigCleansOnlyLegacyProbe(t *testing.T) {
	repo := &fakeStorageConfigRepo{config: &model.StorageConfig{Strategy: model.StorageS3}}
	remote := &fakeStorageProvider{existing: map[string]bool{
		brandcompat.LegacyStorageProbeKey: true,
		"uploads/keep.jpg":                true,
	}}
	runtime := NewStorageRuntimeService(repo, provider.NewRegistry(), &fakeStorageProvider{}, nil)
	runtime.remoteProviderFactory = func(config *model.StorageConfig) (provider.StorageProvider, error) {
		return remote, nil
	}

	require.NoError(t, runtime.RestoreStartupConfig(context.Background()))
	assert.Equal(t, []string{brandcompat.LegacyStorageProbeKey}, remote.deleted)
}

func TestStorageRuntimeProbeFailureKeepsOldProvider(t *testing.T) {
	repo := &fakeStorageConfigRepo{}
	registry := provider.NewRegistry()
	local := &fakeStorageProvider{}
	runtime := NewStorageRuntimeService(repo, registry, local, nil)
	probeErr := errors.New("head failed")
	runtime.remoteProviderFactory = func(config *model.StorageConfig) (provider.StorageProvider, error) {
		return &fakeStorageProvider{existsErr: probeErr}, nil
	}

	_, err := runtime.UpdateConfig(context.Background(), StorageConfigRequest{
		Strategy:  "s3",
		Bucket:    "bucket",
		Region:    "us-east-1",
		Endpoint:  "https://s3.test",
		AccessKey: "access",
		SecretKey: "secret",
	})
	if err == nil {
		t.Fatal("expected probe failure")
	}
	if repo.upsertCount != 0 {
		t.Fatalf("config was persisted despite probe failure")
	}
	if runtime.ActiveProviderName() != "local" {
		t.Fatalf("expected active provider to remain local, got %q", runtime.ActiveProviderName())
	}
	if registry.Storage() != local {
		t.Fatal("old provider was replaced after probe failure")
	}
}

func TestStorageRuntimeEncryptsSecretAndRestoresPlaintext(t *testing.T) {
	t.Setenv("STORAGE_SECRET_ENCRYPTION_KEY", "test-key")
	repo := &fakeStorageConfigRepo{}
	runtime := NewStorageRuntimeService(repo, provider.NewRegistry(), &fakeStorageProvider{}, NewEnvSecretCipher())
	runtime.remoteProviderFactory = func(config *model.StorageConfig) (provider.StorageProvider, error) {
		return &fakeStorageProvider{}, nil
	}

	if _, err := runtime.UpdateConfig(context.Background(), StorageConfigRequest{
		Strategy:  "s3",
		Bucket:    "bucket",
		Region:    "us-east-1",
		Endpoint:  "https://s3.test",
		AccessKey: "access",
		SecretKey: "plain-secret",
	}); err != nil {
		t.Fatalf("UpdateConfig failed: %v", err)
	}
	if repo.config.SecretKey == "plain-secret" {
		t.Fatal("secret was persisted in plaintext")
	}
	if !bytes.HasPrefix([]byte(repo.config.SecretKey), []byte(secretcipher.Prefix)) {
		t.Fatalf("secret missing encrypted prefix: %q", repo.config.SecretKey)
	}

	var restoredSecret string
	restored := NewStorageRuntimeService(repo, provider.NewRegistry(), &fakeStorageProvider{}, NewEnvSecretCipher())
	restored.remoteProviderFactory = func(config *model.StorageConfig) (provider.StorageProvider, error) {
		restoredSecret = config.SecretKey
		return &fakeStorageProvider{}, nil
	}
	if err := restored.RestoreStartupConfig(context.Background()); err != nil {
		t.Fatalf("RestoreStartupConfig failed: %v", err)
	}
	if restoredSecret != "plain-secret" {
		t.Fatalf("restored secret = %q, want plaintext", restoredSecret)
	}
	if restored.ActiveProviderName() != "s3" {
		t.Fatalf("restored active provider = %q, want s3", restored.ActiveProviderName())
	}
}

func TestEnvSecretCipherPrefersCanonicalBrandKey(t *testing.T) {
	t.Setenv("STORAGE_SECRET_ENCRYPTION_KEY", "")
	t.Setenv(brandcompat.SecretKeyVariable, "canonical-secret")
	t.Setenv(brandcompat.LegacySecretVariable, "legacy-secret")

	ciphertext, err := NewEnvSecretCipher().Encrypt("payload")
	require.NoError(t, err)
	canonical, err := secretcipher.New("canonical-secret")
	require.NoError(t, err)
	plaintext, err := canonical.Decrypt(ciphertext)
	require.NoError(t, err)
	assert.Equal(t, "payload", plaintext)
}

func TestEnvSecretCipherReadsLegacyBrandKey(t *testing.T) {
	t.Setenv("STORAGE_SECRET_ENCRYPTION_KEY", "")
	t.Setenv(brandcompat.SecretKeyVariable, "")
	t.Setenv(brandcompat.LegacySecretVariable, "legacy-secret")

	ciphertext, err := NewEnvSecretCipher().Encrypt("payload")
	require.NoError(t, err)
	legacy, err := secretcipher.New("legacy-secret")
	require.NoError(t, err)
	plaintext, err := legacy.Decrypt(ciphertext)
	require.NoError(t, err)
	assert.Equal(t, "payload", plaintext)
}

func TestStorageRuntimePreservesExistingSecret(t *testing.T) {
	t.Setenv("STORAGE_SECRET_ENCRYPTION_KEY", "test-key")
	repo := &fakeStorageConfigRepo{}
	runtime := NewStorageRuntimeService(repo, provider.NewRegistry(), &fakeStorageProvider{}, NewEnvSecretCipher())

	var builtSecrets []string
	runtime.remoteProviderFactory = func(config *model.StorageConfig) (provider.StorageProvider, error) {
		builtSecrets = append(builtSecrets, config.SecretKey)
		return &fakeStorageProvider{}, nil
	}

	_, err := runtime.UpdateConfig(context.Background(), StorageConfigRequest{
		Strategy:  "s3",
		Bucket:    "bucket",
		Region:    "us-east-1",
		Endpoint:  "https://s3.test",
		AccessKey: "access",
		SecretKey: "kept-secret",
	})
	if err != nil {
		t.Fatalf("initial UpdateConfig failed: %v", err)
	}
	_, err = runtime.UpdateConfig(context.Background(), StorageConfigRequest{
		Strategy:  "s3",
		Bucket:    "bucket-2",
		Region:    "us-east-1",
		Endpoint:  "https://s3.test",
		AccessKey: "access-2",
	})
	if err != nil {
		t.Fatalf("second UpdateConfig failed: %v", err)
	}
	if got := builtSecrets[len(builtSecrets)-1]; got != "kept-secret" {
		t.Fatalf("preserved secret = %q, want kept-secret", got)
	}
}
