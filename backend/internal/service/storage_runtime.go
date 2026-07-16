package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"blotting-consultancy/internal/model"
	"blotting-consultancy/internal/plugins/s3storage"
	"blotting-consultancy/internal/provider"
	"blotting-consultancy/internal/repository"
	"blotting-consultancy/pkg/secretcipher"
)

const (
	storageProviderLocal = "local"
	storageProviderS3    = "s3"
	storageProviderOSS   = "oss"

	storageProbeKey     = ".impress-storage-probe"
	defaultProbeTimeout = 5 * time.Second
)

var ErrStorageProviderUnavailable = errors.New("storage provider unavailable")

// SecretCipher is the integration seam for a shared secret cipher package.
type SecretCipher interface {
	Encrypt(plaintext string) (string, error)
	Decrypt(ciphertext string) (string, error)
}

// StorageConfigRequest is the validated camelCase API shape for storage updates.
type StorageConfigRequest struct {
	Strategy  string `json:"strategy"`
	Bucket    string `json:"bucket"`
	Region    string `json:"region"`
	Endpoint  string `json:"endpoint"`
	AccessKey string `json:"accessKey"`
	SecretKey string `json:"secretKey"`
	BasePath  string `json:"basePath"`
}

// StorageConfigResponse is safe to return from API handlers.
type StorageConfigResponse struct {
	Strategy     model.StorageStrategy `json:"strategy"`
	Bucket       string                `json:"bucket,omitempty"`
	Region       string                `json:"region,omitempty"`
	Endpoint     string                `json:"endpoint,omitempty"`
	AccessKey    string                `json:"accessKey,omitempty"`
	HasSecretKey bool                  `json:"hasSecretKey"`
	BasePath     string                `json:"basePath,omitempty"`
	UpdatedAt    time.Time             `json:"updatedAt"`
}

type StorageRuntimeService struct {
	mu                    sync.RWMutex
	repo                  repository.StorageConfigRepository
	registry              *provider.Registry
	cipher                SecretCipher
	localStorage          provider.StorageProvider
	remoteProviderFactory func(*model.StorageConfig) (provider.StorageProvider, error)
	probeTimeout          time.Duration
	activeName            string
}

func NewStorageRuntimeService(repo repository.StorageConfigRepository, registry *provider.Registry, localStorage provider.StorageProvider, cipher SecretCipher) *StorageRuntimeService {
	if registry == nil {
		registry = provider.NewRegistry()
	}
	if localStorage == nil {
		localStorage = NewLocalStorage("./uploads")
	}
	if cipher == nil {
		cipher = NewEnvSecretCipher()
	}
	s := &StorageRuntimeService{
		repo:         repo,
		registry:     registry,
		cipher:       cipher,
		localStorage: localStorage,
		remoteProviderFactory: func(config *model.StorageConfig) (provider.StorageProvider, error) {
			endpoint := strings.TrimSpace(config.Endpoint)
			if endpoint == "" && config.Strategy == model.StorageS3 {
				endpoint = "https://s3.amazonaws.com"
			}
			return s3storage.New(s3storage.Config{
				Endpoint:        endpoint,
				Region:          strings.TrimSpace(config.Region),
				Bucket:          strings.TrimSpace(config.Bucket),
				AccessKeyID:     strings.TrimSpace(config.AccessKey),
				SecretAccessKey: config.SecretKey,
				UsePathStyle:    true,
			})
		},
		probeTimeout: defaultProbeTimeout,
		activeName:   storageProviderLocal,
	}
	s.registry.SetStorage(localStorage)
	s.registry.SetStorageByName(storageProviderLocal, localStorage)
	return s
}

func (s *StorageRuntimeService) SetRepository(repo repository.StorageConfigRepository) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.repo = repo
}

func (s *StorageRuntimeService) RestoreStartupConfig(ctx context.Context) error {
	s.mu.RLock()
	repo := s.repo
	s.mu.RUnlock()
	if repo == nil {
		return nil
	}

	stored, err := repo.Get(ctx)
	if err != nil {
		return fmt.Errorf("load storage config: %w", err)
	}
	config, err := s.decryptConfig(stored)
	if err != nil {
		return fmt.Errorf("decrypt storage config: %w", err)
	}
	storage, name, err := s.buildProvider(config)
	if err != nil {
		return err
	}
	if err := s.probe(ctx, config.Strategy, storage); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.registry.SetStorageByName(name, storage)
	s.registry.SetStorage(storage)
	s.activeName = name
	return nil
}

func (s *StorageRuntimeService) GetConfig(ctx context.Context) (*StorageConfigResponse, error) {
	s.mu.RLock()
	repo := s.repo
	s.mu.RUnlock()
	if repo == nil {
		return &StorageConfigResponse{Strategy: model.StorageLocal}, nil
	}
	config, err := repo.Get(ctx)
	if err != nil {
		return nil, err
	}
	return configResponse(config), nil
}

func (s *StorageRuntimeService) UpdateConfig(ctx context.Context, req StorageConfigRequest) (*StorageConfigResponse, error) {
	s.mu.RLock()
	repo := s.repo
	s.mu.RUnlock()
	if repo == nil {
		return nil, errors.New("storage config repository is not configured")
	}

	existing, err := repo.Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("load existing storage config: %w", err)
	}
	existingPlain, err := s.decryptConfig(existing)
	if err != nil {
		return nil, fmt.Errorf("decrypt existing storage config: %w", err)
	}

	config := requestToConfig(req)
	if config.SecretKey == "" && config.Strategy == existingPlain.Strategy {
		config.SecretKey = existingPlain.SecretKey
	}
	if err := config.Validate(); err != nil {
		return nil, err
	}

	nextProvider, nextName, err := s.buildProvider(config)
	if err != nil {
		return nil, err
	}
	if err := s.probe(ctx, config.Strategy, nextProvider); err != nil {
		return nil, err
	}

	encryptedConfig, err := s.encryptConfig(config)
	if err != nil {
		return nil, fmt.Errorf("encrypt storage secret: %w", err)
	}
	if err := repo.Upsert(ctx, encryptedConfig); err != nil {
		return nil, fmt.Errorf("persist storage config: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.registry.SetStorageByName(nextName, nextProvider)
	s.registry.SetStorage(nextProvider)
	s.activeName = nextName
	return configResponse(encryptedConfig), nil
}

func (s *StorageRuntimeService) TestConnection(ctx context.Context) error {
	s.mu.RLock()
	repo := s.repo
	s.mu.RUnlock()
	if repo == nil {
		return nil
	}
	stored, err := repo.Get(ctx)
	if err != nil {
		return err
	}
	config, err := s.decryptConfig(stored)
	if err != nil {
		return err
	}
	storage, _, err := s.buildProvider(config)
	if err != nil {
		return err
	}
	return s.probe(ctx, config.Strategy, storage)
}

func (s *StorageRuntimeService) ActiveProviderName() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.activeName
}

func (s *StorageRuntimeService) Save(ctx context.Context, filename string, reader io.Reader, size int64) (key, providerName, publicURL string, err error) {
	s.mu.RLock()
	storage := s.registry.Storage()
	providerName = s.activeName
	s.mu.RUnlock()
	if storage == nil {
		return "", "", "", fmt.Errorf("%w: active storage provider is not registered", ErrStorageProviderUnavailable)
	}
	key, err = storage.Save(ctx, filename, reader, size)
	if err != nil {
		return "", "", "", err
	}
	return key, providerName, s.url(providerName, storage, key), nil
}

func (s *StorageRuntimeService) Delete(ctx context.Context, providerName, key string) error {
	if providerName == "" {
		providerName = storageProviderLocal
	}
	storage, ok := s.registry.StorageByName(providerName)
	if !ok {
		return fmt.Errorf("%w: storage provider %q is not registered", ErrStorageProviderUnavailable, providerName)
	}
	return storage.Delete(ctx, key)
}

func (s *StorageRuntimeService) URL(providerName, key string) string {
	if providerName == "" {
		s.mu.RLock()
		providerName = s.activeName
		s.mu.RUnlock()
	}
	storage, ok := s.registry.StorageByName(providerName)
	if !ok {
		return ""
	}
	return s.url(providerName, storage, key)
}

func (s *StorageRuntimeService) buildProvider(config *model.StorageConfig) (provider.StorageProvider, string, error) {
	switch config.Strategy {
	case model.StorageLocal:
		return s.localStorage, storageProviderLocal, nil
	case model.StorageS3, model.StorageOSS:
		p, err := s.remoteProviderFactory(config)
		if err != nil {
			return nil, "", err
		}
		var storage provider.StorageProvider = p
		if strings.TrimSpace(config.BasePath) != "" {
			storage = &prefixStorageProvider{
				next:   storage,
				prefix: strings.Trim(strings.TrimSpace(config.BasePath), "/"),
			}
		}
		return storage, string(config.Strategy), nil
	default:
		return nil, "", errors.New("strategy must be one of: local, s3, oss")
	}
}

func (s *StorageRuntimeService) probe(ctx context.Context, strategy model.StorageStrategy, storage provider.StorageProvider) error {
	if strategy == model.StorageLocal {
		return nil
	}
	probeCtx, cancel := context.WithTimeout(ctx, s.probeTimeout)
	defer cancel()
	if _, err := storage.Exists(probeCtx, storageProbeKey); err != nil {
		return fmt.Errorf("storage remote probe failed: %w", err)
	}
	return nil
}

func (s *StorageRuntimeService) decryptConfig(config *model.StorageConfig) (*model.StorageConfig, error) {
	cp := *config
	if cp.SecretKey == "" {
		return &cp, nil
	}
	plain, err := s.cipher.Decrypt(cp.SecretKey)
	if err != nil {
		return nil, err
	}
	cp.SecretKey = plain
	return &cp, nil
}

func (s *StorageRuntimeService) encryptConfig(config *model.StorageConfig) (*model.StorageConfig, error) {
	cp := *config
	if cp.SecretKey == "" {
		return &cp, nil
	}
	encrypted, err := s.cipher.Encrypt(cp.SecretKey)
	if err != nil {
		return nil, err
	}
	cp.SecretKey = encrypted
	return &cp, nil
}

func (s *StorageRuntimeService) url(providerName string, storage provider.StorageProvider, key string) string {
	if storage == nil || key == "" {
		return ""
	}
	if providerName == storageProviderLocal {
		return "/uploads/" + strings.TrimLeft(filepath.ToSlash(key), "/")
	}
	return storage.URL(key)
}

func requestToConfig(req StorageConfigRequest) *model.StorageConfig {
	return &model.StorageConfig{
		Strategy:  model.StorageStrategy(strings.TrimSpace(req.Strategy)),
		Bucket:    strings.TrimSpace(req.Bucket),
		Region:    strings.TrimSpace(req.Region),
		Endpoint:  strings.TrimSpace(req.Endpoint),
		AccessKey: strings.TrimSpace(req.AccessKey),
		SecretKey: req.SecretKey,
		BasePath:  strings.Trim(strings.TrimSpace(req.BasePath), "/"),
	}
}

func configResponse(config *model.StorageConfig) *StorageConfigResponse {
	return &StorageConfigResponse{
		Strategy:     config.Strategy,
		Bucket:       config.Bucket,
		Region:       config.Region,
		Endpoint:     config.Endpoint,
		AccessKey:    config.AccessKey,
		HasSecretKey: config.SecretKey != "",
		BasePath:     config.BasePath,
		UpdatedAt:    config.UpdatedAt,
	}
}

type prefixStorageProvider struct {
	next   provider.StorageProvider
	prefix string
}

func (p *prefixStorageProvider) Save(ctx context.Context, filename string, reader io.Reader, size int64) (string, error) {
	return p.next.Save(ctx, p.withPrefix(filename), reader, size)
}

func (p *prefixStorageProvider) Get(ctx context.Context, key string) (io.ReadCloser, error) {
	return p.next.Get(ctx, key)
}

func (p *prefixStorageProvider) Delete(ctx context.Context, key string) error {
	return p.next.Delete(ctx, key)
}

func (p *prefixStorageProvider) URL(key string) string {
	return p.next.URL(key)
}

func (p *prefixStorageProvider) Exists(ctx context.Context, key string) (bool, error) {
	return p.next.Exists(ctx, p.withPrefix(key))
}

func (p *prefixStorageProvider) withPrefix(key string) string {
	if p.prefix == "" {
		return strings.TrimLeft(path.Clean(key), "/")
	}
	return path.Join(p.prefix, strings.TrimLeft(path.Clean(key), "/"))
}

func NewEnvSecretCipher() SecretCipher {
	material := os.Getenv("STORAGE_SECRET_ENCRYPTION_KEY")
	if material == "" {
		material = os.Getenv("IMPRESS_SECRET_KEY")
	}
	if material == "" {
		material = os.Getenv("JWT_SECRET")
	}
	if material == "" {
		material = fallbackStorageSecret()
	}
	c, err := secretcipher.New(material)
	if err != nil {
		panic(err)
	}
	return secretCipherAdapter{cipher: c}
}

var (
	fallbackStorageSecretOnce  sync.Once
	fallbackStorageSecretValue string
)

func fallbackStorageSecret() string {
	fallbackStorageSecretOnce.Do(func() {
		buf := make([]byte, 32)
		if _, err := rand.Read(buf); err != nil {
			panic(fmt.Errorf("generate ephemeral storage secret: %w", err))
		}
		fallbackStorageSecretValue = "ephemeral-" + hex.EncodeToString(buf)
	})
	return fallbackStorageSecretValue
}

type secretCipherAdapter struct {
	cipher *secretcipher.Cipher
}

func (a secretCipherAdapter) Encrypt(plaintext string) (string, error) {
	if plaintext == "" || strings.HasPrefix(plaintext, secretcipher.Prefix) {
		return plaintext, nil
	}
	return a.cipher.Encrypt(plaintext)
}

func (a secretCipherAdapter) Decrypt(ciphertext string) (string, error) {
	if ciphertext == "" || !strings.HasPrefix(ciphertext, secretcipher.Prefix) {
		return ciphertext, nil
	}
	return a.cipher.Decrypt(ciphertext)
}

var (
	defaultStorageRuntimeMu sync.Mutex
	defaultStorageRuntime   *StorageRuntimeService
)

func ConfigureDefaultStorageRuntime(repo repository.StorageConfigRepository, uploadDir string) *StorageRuntimeService {
	defaultStorageRuntimeMu.Lock()
	defer defaultStorageRuntimeMu.Unlock()
	if uploadDir == "" {
		uploadDir = "./uploads"
	}
	if defaultStorageRuntime == nil {
		defaultStorageRuntime = NewStorageRuntimeService(repo, provider.NewRegistry(), NewLocalStorage(uploadDir), nil)
		return defaultStorageRuntime
	}
	if repo != nil {
		defaultStorageRuntime.SetRepository(repo)
	}
	return defaultStorageRuntime
}

func DefaultStorageRuntime() *StorageRuntimeService {
	return ConfigureDefaultStorageRuntime(nil, "")
}
