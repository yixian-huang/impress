package service

import (
	"context"
	"math"
	"sort"
	"sync"

	"blotting-consultancy/internal/provider"
)

// vectorEntry stores a single vector with its metadata.
type vectorEntry struct {
	ID        string
	Embedding []float64
	Metadata  map[string]string
}

// MemoryVectorStore is an in-memory implementation of VectorStoreProvider
// using brute-force cosine similarity. Suitable for development and testing.
type MemoryVectorStore struct {
	mu      sync.RWMutex
	entries map[string]*vectorEntry
}

// NewMemoryVectorStore creates a new in-memory vector store.
func NewMemoryVectorStore() *MemoryVectorStore {
	return &MemoryVectorStore{
		entries: make(map[string]*vectorEntry),
	}
}

// Store saves an embedding with its ID and metadata.
func (m *MemoryVectorStore) Store(_ context.Context, id string, embedding []float64, metadata map[string]string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.entries[id] = &vectorEntry{
		ID:        id,
		Embedding: embedding,
		Metadata:  metadata,
	}
	return nil
}

// Search finds the top-K most similar vectors to the query embedding using cosine similarity.
func (m *MemoryVectorStore) Search(_ context.Context, query []float64, topK int) ([]provider.VectorResult, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(m.entries) == 0 {
		return nil, nil
	}

	type scored struct {
		id       string
		score    float64
		metadata map[string]string
	}

	results := make([]scored, 0, len(m.entries))
	for _, entry := range m.entries {
		sim := cosineSimilarity(query, entry.Embedding)
		results = append(results, scored{
			id:       entry.ID,
			score:    sim,
			metadata: entry.Metadata,
		})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].score > results[j].score
	})

	if topK > len(results) {
		topK = len(results)
	}

	out := make([]provider.VectorResult, topK)
	for i := 0; i < topK; i++ {
		out[i] = provider.VectorResult{
			ID:       results[i].id,
			Score:    results[i].score,
			Metadata: results[i].metadata,
		}
	}

	return out, nil
}

// Delete removes an embedding by ID.
func (m *MemoryVectorStore) Delete(_ context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.entries, id)
	return nil
}

// Count returns the number of stored entries (useful for testing).
func (m *MemoryVectorStore) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.entries)
}

// cosineSimilarity computes the cosine similarity between two vectors.
// Returns 0 if either vector has zero magnitude.
func cosineSimilarity(a, b []float64) float64 {
	if len(a) != len(b) || len(a) == 0 {
		return 0
	}

	var dot, normA, normB float64
	for i := range a {
		dot += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	magA := math.Sqrt(normA)
	magB := math.Sqrt(normB)
	if magA == 0 || magB == 0 {
		return 0
	}

	return dot / (magA * magB)
}
