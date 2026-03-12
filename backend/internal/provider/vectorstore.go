package provider

import "context"

// VectorResult represents a single result from a vector similarity search.
type VectorResult struct {
	ID       string
	Score    float64
	Metadata map[string]string
}

// VectorStoreProvider defines an interface for storing and searching vector embeddings.
type VectorStoreProvider interface {
	// Store saves an embedding with its ID and metadata.
	Store(ctx context.Context, id string, embedding []float64, metadata map[string]string) error
	// Search finds the top-K most similar vectors to the query embedding.
	Search(ctx context.Context, query []float64, topK int) ([]VectorResult, error)
	// Delete removes an embedding by ID.
	Delete(ctx context.Context, id string) error
}
