package service

import (
	"context"
	"hash/fnv"
	"math"

	"github.com/yixian-huang/inkless/backend/internal/provider"
)

// StubAIProvider is a development/testing stub that implements the AIProvider interface.
// It generates deterministic pseudo-embeddings and simple responses.
type StubAIProvider struct {
	embeddingDim int
}

// NewStubAIProvider creates a new stub AI provider.
func NewStubAIProvider() *StubAIProvider {
	return &StubAIProvider{
		embeddingDim: 128,
	}
}

// Embed generates a deterministic pseudo-embedding based on the text hash.
// This is for development/testing only -- a real implementation would call
// an embedding API (e.g., OpenAI, Cohere, or a local model).
func (s *StubAIProvider) Embed(_ context.Context, text string) ([]float64, error) {
	h := fnv.New64a()
	h.Write([]byte(text))
	seed := h.Sum64()

	embedding := make([]float64, s.embeddingDim)
	for i := range embedding {
		// Simple deterministic pseudo-random based on seed and index
		v := float64(seed>>uint(i%64)&0xFF) / 255.0
		embedding[i] = v*2 - 1 // Normalize to [-1, 1]
	}

	// Normalize to unit vector
	var norm float64
	for _, v := range embedding {
		norm += v * v
	}
	norm = math.Sqrt(norm)
	if norm > 0 {
		for i := range embedding {
			embedding[i] /= norm
		}
	}

	return embedding, nil
}

// ChatComplete returns a stub response indicating that a real AI provider is needed.
func (s *StubAIProvider) ChatComplete(_ context.Context, _ string, userMessage string) (string, error) {
	return "This is a stub response. Please configure a real AI provider (e.g., OpenAI) to get actual answers. " +
		"Your question was: " + userMessage, nil
}

// The following methods satisfy the full provider.AIProvider interface for testing.

func (s *StubAIProvider) Chat(_ context.Context, _ provider.ChatRequest) (*provider.ChatResponse, error) {
	return &provider.ChatResponse{Content: "stub response"}, nil
}

func (s *StubAIProvider) Complete(_ context.Context, _ provider.CompletionRequest) (*provider.CompletionResponse, error) {
	return &provider.CompletionResponse{Text: "stub completion"}, nil
}

func (s *StubAIProvider) Summarize(_ context.Context, text string, _ int) (string, error) {
	if len(text) > 100 {
		return text[:100], nil
	}
	return text, nil
}

func (s *StubAIProvider) SuggestTitles(_ context.Context, _ string, count int) ([]string, error) {
	titles := make([]string, count)
	for i := range titles {
		titles[i] = "Stub Title"
	}
	return titles, nil
}

func (s *StubAIProvider) SuggestTags(_ context.Context, _ string, _ []string) ([]string, error) {
	return []string{"stub-tag"}, nil
}

func (s *StubAIProvider) StreamChat(_ context.Context, _ provider.ChatRequest) (<-chan provider.ChatChunk, error) {
	ch := make(chan provider.ChatChunk, 1)
	ch <- provider.ChatChunk{Content: "stub stream", FinishReason: "stop"}
	close(ch)
	return ch, nil
}

func (s *StubAIProvider) Name() string {
	return "stub"
}
