package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"blotting-consultancy/internal/model"
	"blotting-consultancy/internal/provider"
	"blotting-consultancy/internal/repository"
	"blotting-consultancy/pkg/secretcipher"
)

const aiHealthTimeout = 10 * time.Second

var (
	ErrAIAPIKeyRequired    = errors.New("AI API key is required")
	ErrAIUnsupportedConfig = errors.New("unsupported AI provider")
)

// AIConfigService owns persisted AI settings and runtime registry switching.
type AIConfigService struct {
	repo     repository.AIConfigRepository
	cipher   *secretcipher.Cipher
	registry *provider.Registry
	client   *http.Client
}

type AIConfigServiceOption func(*AIConfigService)

func WithAIConfigHTTPClient(client *http.Client) AIConfigServiceOption {
	return func(s *AIConfigService) {
		if client != nil {
			s.client = client
		}
	}
}

func NewAIConfigService(repo repository.AIConfigRepository, cipher *secretcipher.Cipher, registry *provider.Registry, opts ...AIConfigServiceOption) *AIConfigService {
	s := &AIConfigService{
		repo:     repo,
		cipher:   cipher,
		registry: registry,
		client:   http.DefaultClient,
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

type AIConfigInput struct {
	Provider string
	APIKey   string
	BaseURL  string
	Model    string
}

type AIConfigResponse struct {
	Provider     string `json:"provider"`
	Enabled      bool   `json:"enabled"`
	BaseURL      string `json:"base_url,omitempty"`
	Model        string `json:"model,omitempty"`
	HasAPIKey    bool   `json:"has_api_key"`
	APIKeyMasked string `json:"api_key_masked,omitempty"`
}

type AIHealthResponse struct {
	Provider string `json:"provider"`
	Healthy  bool   `json:"healthy"`
	Model    string `json:"model,omitempty"`
	Message  string `json:"message,omitempty"`
}

func (s *AIConfigService) Get(ctx context.Context) (*AIConfigResponse, error) {
	config, err := s.repo.Get(ctx)
	if err != nil {
		return nil, err
	}
	return s.responseForConfig(ctx, config)
}

func (s *AIConfigService) Update(ctx context.Context, input AIConfigInput) (*AIConfigResponse, error) {
	existing, err := s.repo.Get(ctx)
	if err != nil {
		return nil, err
	}

	providerName := normalizeAIProvider(input.Provider)
	config := &model.AIConfig{
		ID:       model.AIConfigSingletonID,
		Provider: providerName,
		BaseURL:  strings.TrimSpace(input.BaseURL),
		Model:    strings.TrimSpace(input.Model),
	}

	apiKey := strings.TrimSpace(input.APIKey)
	switch providerName {
	case model.AIProviderDisabled:
		config.APIKeyCiphertext = ""
	case model.AIProviderOpenAI, model.AIProviderAnthropic:
		if apiKey == "" {
			config.APIKeyCiphertext = existing.APIKeyCiphertext
		} else {
			ciphertext, err := s.cipher.Encrypt(apiKey)
			if err != nil {
				return nil, err
			}
			config.APIKeyCiphertext = ciphertext
		}
		if config.APIKeyCiphertext == "" {
			return nil, ErrAIAPIKeyRequired
		}
	default:
		return nil, fmt.Errorf("%w: %s", ErrAIUnsupportedConfig, input.Provider)
	}

	if err := s.repo.Upsert(ctx, config); err != nil {
		return nil, err
	}
	if err := s.applyRuntimeProvider(ctx, config); err != nil {
		return nil, err
	}

	return s.responseForConfig(ctx, config)
}

// Restore loads persisted AI settings and applies them to the runtime registry.
func (s *AIConfigService) Restore(ctx context.Context) error {
	config, err := s.repo.Get(ctx)
	if err != nil {
		return err
	}
	return s.applyRuntimeProvider(ctx, config)
}

// TestCurrent runs a bounded health check through the currently persisted provider.
func (s *AIConfigService) TestCurrent(ctx context.Context) (*AIHealthResponse, error) {
	config, err := s.repo.Get(ctx)
	if err != nil {
		return nil, err
	}
	return s.testConfig(ctx, config)
}

// Test runs a bounded health check through the supplied provider config without persisting it.
func (s *AIConfigService) Test(ctx context.Context, input AIConfigInput) (*AIHealthResponse, error) {
	existing, err := s.repo.Get(ctx)
	if err != nil {
		return nil, err
	}

	providerName := normalizeAIProvider(input.Provider)
	config := &model.AIConfig{
		ID:       model.AIConfigSingletonID,
		Provider: providerName,
		BaseURL:  strings.TrimSpace(input.BaseURL),
		Model:    strings.TrimSpace(input.Model),
	}
	apiKey := strings.TrimSpace(input.APIKey)
	if apiKey != "" {
		config.APIKeyCiphertext, err = s.cipher.Encrypt(apiKey)
		if err != nil {
			return nil, err
		}
	} else {
		config.APIKeyCiphertext = existing.APIKeyCiphertext
	}

	return s.testConfig(ctx, config)
}

func (s *AIConfigService) testConfig(ctx context.Context, config *model.AIConfig) (*AIHealthResponse, error) {
	aiProvider, err := s.providerForConfig(ctx, config)
	if err != nil {
		return nil, err
	}
	if aiProvider.Name() == "noop" {
		return &AIHealthResponse{
			Provider: model.AIProviderDisabled,
			Healthy:  false,
			Message:  "AI provider is disabled",
		}, nil
	}

	healthCtx, cancel := context.WithTimeout(ctx, aiHealthTimeout)
	defer cancel()
	resp, err := aiProvider.Chat(healthCtx, provider.ChatRequest{
		Messages: []provider.ChatMessage{
			{Role: "user", Content: "Reply with ok."},
		},
		MaxTokens:   4,
		Temperature: 0,
	})
	if err != nil {
		return nil, err
	}

	return &AIHealthResponse{
		Provider: aiProvider.Name(),
		Healthy:  true,
		Model:    resp.Model,
		Message:  strings.TrimSpace(resp.Content),
	}, nil
}

func (s *AIConfigService) applyRuntimeProvider(ctx context.Context, config *model.AIConfig) error {
	aiProvider, err := s.providerForConfig(ctx, config)
	if err != nil {
		return err
	}
	s.registry.SetAI(aiProvider)
	return nil
}

func (s *AIConfigService) providerForConfig(ctx context.Context, config *model.AIConfig) (provider.AIProvider, error) {
	providerName := normalizeAIProvider(config.Provider)
	switch providerName {
	case model.AIProviderDisabled:
		return NewNoopAIProvider(), nil
	case model.AIProviderOpenAI:
		apiKey, err := s.decryptAPIKey(ctx, config)
		if err != nil {
			return nil, err
		}
		return NewOpenAIProvider(OpenAIConfig{
			APIKey:  apiKey,
			BaseURL: config.BaseURL,
			Model:   config.Model,
			Client:  s.client,
		}), nil
	case model.AIProviderAnthropic:
		apiKey, err := s.decryptAPIKey(ctx, config)
		if err != nil {
			return nil, err
		}
		return NewAnthropicProvider(AnthropicConfig{
			APIKey:  apiKey,
			BaseURL: config.BaseURL,
			Model:   config.Model,
			Client:  s.client,
		}), nil
	default:
		return nil, fmt.Errorf("%w: %s", ErrAIUnsupportedConfig, config.Provider)
	}
}

func (s *AIConfigService) decryptAPIKey(_ context.Context, config *model.AIConfig) (string, error) {
	if strings.TrimSpace(config.APIKeyCiphertext) == "" {
		return "", ErrAIAPIKeyRequired
	}
	apiKey, err := s.cipher.Decrypt(config.APIKeyCiphertext)
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(apiKey) == "" {
		return "", ErrAIAPIKeyRequired
	}
	return apiKey, nil
}

func (s *AIConfigService) responseForConfig(ctx context.Context, config *model.AIConfig) (*AIConfigResponse, error) {
	resp := &AIConfigResponse{
		Provider:  normalizeAIProvider(config.Provider),
		Enabled:   normalizeAIProvider(config.Provider) != model.AIProviderDisabled,
		BaseURL:   config.BaseURL,
		Model:     config.Model,
		HasAPIKey: config.HasAPIKey(),
	}
	if config.HasAPIKey() {
		apiKey, err := s.decryptAPIKey(ctx, config)
		if err != nil {
			return nil, err
		}
		resp.APIKeyMasked = maskAPIKey(apiKey)
	}
	return resp, nil
}

func normalizeAIProvider(providerName string) string {
	switch strings.ToLower(strings.TrimSpace(providerName)) {
	case "", "disabled", "noop":
		return model.AIProviderDisabled
	case model.AIProviderOpenAI:
		return model.AIProviderOpenAI
	case model.AIProviderAnthropic:
		return model.AIProviderAnthropic
	default:
		return strings.ToLower(strings.TrimSpace(providerName))
	}
}

func maskAPIKey(apiKey string) string {
	apiKey = strings.TrimSpace(apiKey)
	if apiKey == "" {
		return ""
	}
	if len(apiKey) <= 8 {
		return "********"
	}
	return apiKey[:4] + "..." + apiKey[len(apiKey)-4:]
}
