package service

import (
	"context"

	"blotting-consultancy/internal/provider"
)

// NoopTranslationProvider is a no-op implementation of TranslationProvider.
// It returns empty translations and is used as the default when no AI provider is configured.
type NoopTranslationProvider struct{}

// NewNoopTranslationProvider creates a new NoopTranslationProvider
func NewNoopTranslationProvider() provider.TranslationProvider {
	return &NoopTranslationProvider{}
}

// Translate returns an empty translation response
func (n *NoopTranslationProvider) Translate(_ context.Context, req provider.TranslateRequest) (*provider.TranslateResponse, error) {
	return &provider.TranslateResponse{
		OriginalText:   req.Text,
		TranslatedText: "",
		SourceLang:     req.SourceLang,
		TargetLang:     req.TargetLang,
	}, nil
}

// BatchTranslate returns empty translation responses for each item
func (n *NoopTranslationProvider) BatchTranslate(_ context.Context, items []provider.TranslateRequest) ([]provider.TranslateResponse, error) {
	responses := make([]provider.TranslateResponse, len(items))
	for i, req := range items {
		responses[i] = provider.TranslateResponse{
			OriginalText:   req.Text,
			TranslatedText: "",
			SourceLang:     req.SourceLang,
			TargetLang:     req.TargetLang,
		}
	}
	return responses, nil
}

// DetectLanguage returns an empty string (unknown language)
func (n *NoopTranslationProvider) DetectLanguage(_ context.Context, _ string) (string, error) {
	return "", nil
}
