package service

import (
	"context"
	"errors"

	"github.com/yixian-huang/inkless/backend/internal/provider"
)

// ErrAINotConfigured is returned when no AI provider is configured.
var ErrAINotConfigured = errors.New("AI provider not configured")

// NoopAIProvider is the default AI provider that returns errors
// indicating that no AI backend has been configured.
type NoopAIProvider struct{}

// NewNoopAIProvider creates a new NoopAIProvider.
func NewNoopAIProvider() *NoopAIProvider {
	return &NoopAIProvider{}
}

func (n *NoopAIProvider) Name() string {
	return "noop"
}

func (n *NoopAIProvider) Chat(_ context.Context, _ provider.ChatRequest) (*provider.ChatResponse, error) {
	return nil, ErrAINotConfigured
}

func (n *NoopAIProvider) Complete(_ context.Context, _ provider.CompletionRequest) (*provider.CompletionResponse, error) {
	return nil, ErrAINotConfigured
}

func (n *NoopAIProvider) Summarize(_ context.Context, _ string, _ int) (string, error) {
	return "", ErrAINotConfigured
}

func (n *NoopAIProvider) SuggestTitles(_ context.Context, _ string, _ int) ([]string, error) {
	return nil, ErrAINotConfigured
}

func (n *NoopAIProvider) SuggestTags(_ context.Context, _ string, _ []string) ([]string, error) {
	return nil, ErrAINotConfigured
}

func (n *NoopAIProvider) StreamChat(_ context.Context, _ provider.ChatRequest) (<-chan provider.ChatChunk, error) {
	return nil, ErrAINotConfigured
}

func (n *NoopAIProvider) Embed(_ context.Context, _ string) ([]float64, error) {
	return nil, ErrAINotConfigured
}

func (n *NoopAIProvider) ChatComplete(_ context.Context, _ string, _ string) (string, error) {
	return "", ErrAINotConfigured
}
