package qa

import (
	"context"
	"errors"
	"testing"

	"blotting-consultancy/internal/provider"
	"blotting-consultancy/internal/service"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestChunkText_Short(t *testing.T) {
	chunks := ChunkText("Hello world", 2000, 200)
	assert.Len(t, chunks, 1)
	assert.Equal(t, "Hello world", chunks[0])
}

func TestChunkText_Empty(t *testing.T) {
	chunks := ChunkText("", 2000, 200)
	assert.Nil(t, chunks)

	chunks = ChunkText("   ", 2000, 200)
	assert.Nil(t, chunks)
}

func TestChunkText_LongText(t *testing.T) {
	// Create a text that's longer than chunk size
	text := ""
	for i := 0; i < 100; i++ {
		text += "This is sentence number. "
	}

	chunks := ChunkText(text, 100, 20)
	assert.True(t, len(chunks) > 1, "expected multiple chunks, got %d", len(chunks))

	// All chunks should be non-empty
	for i, c := range chunks {
		assert.NotEmpty(t, c, "chunk %d is empty", i)
	}
}

func TestChunkText_ChineseText(t *testing.T) {
	text := ""
	for i := 0; i < 50; i++ {
		text += "这是一个测试句子。"
	}

	chunks := ChunkText(text, 100, 20)
	assert.True(t, len(chunks) > 1)

	for _, c := range chunks {
		assert.NotEmpty(t, c)
	}
}

func TestChunkText_ParagraphBreaks(t *testing.T) {
	text := "First paragraph with enough content to fill the space.\n\n" +
		"Second paragraph with more content to test paragraph breaking.\n\n" +
		"Third paragraph that adds even more text to ensure chunking happens."

	chunks := ChunkText(text, 60, 10)
	assert.True(t, len(chunks) >= 2)
}

func TestEmbeddingService_IndexContentResolvesAIProviderFromRegistryAtCallTime(t *testing.T) {
	registry := provider.NewRegistry()
	registry.SetAI(&qaMockAI{name: "first", vector: []float64{1, 0}})
	vs := NewMemoryVectorStore()
	svc := NewEmbeddingServiceWithRegistry(registry, vs)

	registry.SetAI(&qaMockAI{name: "second", vector: []float64{0, 1}})

	count, err := svc.IndexContent(context.Background(), "article:1", "content to index", nil)
	require.NoError(t, err)
	assert.Equal(t, 1, count)

	results, err := vs.Search(context.Background(), []float64{0, 1}, 1)
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, 1.0, results[0].Score)
}

func TestEmbeddingService_IndexContentReturnsErrAINotConfiguredWhenRegistryHasNoAI(t *testing.T) {
	svc := NewEmbeddingServiceWithRegistry(provider.NewRegistry(), NewMemoryVectorStore())

	_, err := svc.IndexContent(context.Background(), "article:1", "content to index", nil)
	assert.True(t, errors.Is(err, service.ErrAINotConfigured))
}
