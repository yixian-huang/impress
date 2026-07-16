package service

import (
	"context"

	"blotting-consultancy/internal/provider"
)

// NoopTranslationProvider is a no-op implementation of TranslationProvider.
// It returns ErrAINotConfigured and is used as the default when no AI provider is configured.
type NoopTranslationProvider struct{}

// NewNoopTranslationProvider creates a new NoopTranslationProvider
func NewNoopTranslationProvider() provider.TranslationProvider {
	return &NoopTranslationProvider{}
}

// Translate reports that AI translation is not configured.
func (n *NoopTranslationProvider) Translate(_ context.Context, req provider.TranslateRequest) (*provider.TranslateResponse, error) {
	return nil, ErrAINotConfigured
}

// BatchTranslate reports that AI translation is not configured.
func (n *NoopTranslationProvider) BatchTranslate(_ context.Context, items []provider.TranslateRequest) ([]provider.TranslateResponse, error) {
	return nil, ErrAINotConfigured
}

// DetectLanguage reports that AI translation is not configured.
func (n *NoopTranslationProvider) DetectLanguage(_ context.Context, _ string) (string, error) {
	return "", ErrAINotConfigured
}
