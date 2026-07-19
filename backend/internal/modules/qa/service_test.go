package qa

import (
	"context"
	"errors"
	"testing"

	"github.com/yixian-huang/inkless/backend/internal/provider"
	"github.com/yixian-huang/inkless/backend/internal/service"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type qaMockAI struct {
	name   string
	answer string
	vector []float64
}

func (m *qaMockAI) ChatComplete(_ context.Context, _, _ string) (string, error) {
	return m.answer, nil
}

func (m *qaMockAI) Chat(_ context.Context, _ provider.ChatRequest) (*provider.ChatResponse, error) {
	return &provider.ChatResponse{Content: m.answer}, nil
}

func (m *qaMockAI) Complete(_ context.Context, _ provider.CompletionRequest) (*provider.CompletionResponse, error) {
	return &provider.CompletionResponse{Text: m.answer}, nil
}

func (m *qaMockAI) Summarize(_ context.Context, text string, _ int) (string, error) {
	return text, nil
}

func (m *qaMockAI) SuggestTitles(_ context.Context, _ string, count int) ([]string, error) {
	titles := make([]string, count)
	return titles, nil
}

func (m *qaMockAI) SuggestTags(_ context.Context, _ string, _ []string) ([]string, error) {
	return nil, nil
}

func (m *qaMockAI) StreamChat(_ context.Context, _ provider.ChatRequest) (<-chan provider.ChatChunk, error) {
	ch := make(chan provider.ChatChunk)
	close(ch)
	return ch, nil
}

func (m *qaMockAI) Embed(_ context.Context, _ string) ([]float64, error) {
	if m.vector != nil {
		return m.vector, nil
	}
	return []float64{1, 0}, nil
}

func (m *qaMockAI) Name() string { return m.name }

func TestQAService_AskEmpty(t *testing.T) {
	ai := service.NewStubAIProvider()
	vs := NewMemoryVectorStore()
	svc := NewQAService(ai, vs)

	_, err := svc.Ask(context.Background(), "", "zh")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "empty")
}

func TestQAService_AskWithNoContent(t *testing.T) {
	ai := service.NewStubAIProvider()
	vs := NewMemoryVectorStore()
	svc := NewQAService(ai, vs)

	result, err := svc.Ask(context.Background(), "What is this company?", "zh")
	require.NoError(t, err)
	assert.NotEmpty(t, result.Answer)
	assert.Empty(t, result.Sources) // no content indexed
}

func TestQAService_AskWithIndexedContent(t *testing.T) {
	ctx := context.Background()
	ai := service.NewStubAIProvider()
	vs := NewMemoryVectorStore()

	// Index some content
	embSvc := NewEmbeddingService(ai, vs)
	count, err := embSvc.IndexContent(ctx, "test:1", "We provide consulting services for businesses.", map[string]string{
		"type": "content",
	})
	require.NoError(t, err)
	assert.Equal(t, 1, count)

	// Now query
	svc := NewQAService(ai, vs)
	result, err := svc.Ask(ctx, "What services do you provide?", "en")
	require.NoError(t, err)
	assert.NotEmpty(t, result.Answer)
	// With the stub AI, the embedding similarity may or may not find relevant chunks
	// depending on the hash-based pseudo-embedding, so we just check structure
}

func TestQAService_AskResolvesAIProviderFromRegistryAtCallTime(t *testing.T) {
	registry := provider.NewRegistry()
	registry.SetAI(&qaMockAI{name: "first", answer: "first answer"})
	vs := NewMemoryVectorStore()
	svc := NewQAServiceWithRegistry(registry, vs)

	registry.SetAI(&qaMockAI{name: "second", answer: "second answer"})

	result, err := svc.Ask(context.Background(), "What is this company?", "en")
	require.NoError(t, err)
	assert.Equal(t, "second answer", result.Answer)
}

func TestQAService_AskReturnsErrAINotConfiguredWhenRegistryHasNoAI(t *testing.T) {
	svc := NewQAServiceWithRegistry(provider.NewRegistry(), NewMemoryVectorStore())

	_, err := svc.Ask(context.Background(), "What is this company?", "en")
	assert.True(t, errors.Is(err, service.ErrAINotConfigured))
}

func TestQAService_EndToEnd(t *testing.T) {
	ctx := context.Background()
	ai := service.NewStubAIProvider()
	vs := NewMemoryVectorStore()

	// Index multiple pieces of content
	embSvc := NewEmbeddingService(ai, vs)
	embSvc.IndexContent(ctx, "about:1", "Our company was founded in 2020. We are experts in digital transformation.", map[string]string{"type": "page"})
	embSvc.IndexContent(ctx, "services:1", "We offer cloud migration, AI consulting, and data analytics.", map[string]string{"type": "page"})

	assert.True(t, vs.Count() >= 2, "expected at least 2 vectors stored")

	svc := NewQAService(ai, vs)
	result, err := svc.Ask(ctx, "When was the company founded?", "en")
	require.NoError(t, err)
	assert.NotEmpty(t, result.Answer)
}

func TestBuildSystemPrompt(t *testing.T) {
	// No context
	prompt := buildSystemPrompt(nil, "zh")
	assert.Contains(t, prompt, "Chinese")
	assert.Contains(t, prompt, "helpful assistant")

	// With context, English locale
	prompt = buildSystemPrompt([]string{"chunk1", "chunk2"}, "en")
	assert.Contains(t, prompt, "English")
	assert.Contains(t, prompt, "chunk1")
	assert.Contains(t, prompt, "chunk2")
}
