package provider

import "context"

// ChatMessage represents a single message in a chat conversation.
type ChatMessage struct {
	Role    string `json:"role"`    // "system", "user", "assistant"
	Content string `json:"content"`
}

// ChatRequest is the input for a chat completion call.
type ChatRequest struct {
	Messages    []ChatMessage `json:"messages"`
	Model       string        `json:"model,omitempty"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
	Temperature float64       `json:"temperature,omitempty"`
}

// ChatResponse is the output of a chat completion call.
type ChatResponse struct {
	Content      string `json:"content"`
	Model        string `json:"model"`
	FinishReason string `json:"finish_reason"`
	PromptTokens int    `json:"prompt_tokens"`
	OutputTokens int    `json:"output_tokens"`
}

// ChatChunk is a single chunk from a streaming chat response.
type ChatChunk struct {
	Content      string `json:"content"`
	FinishReason string `json:"finish_reason,omitempty"`
	Err          error  `json:"-"`
}

// CompletionRequest is the input for a text completion call.
type CompletionRequest struct {
	Prompt      string  `json:"prompt"`
	Model       string  `json:"model,omitempty"`
	MaxTokens   int     `json:"max_tokens,omitempty"`
	Temperature float64 `json:"temperature,omitempty"`
}

// CompletionResponse is the output of a text completion call.
type CompletionResponse struct {
	Text         string `json:"text"`
	Model        string `json:"model"`
	FinishReason string `json:"finish_reason"`
	PromptTokens int    `json:"prompt_tokens"`
	OutputTokens int    `json:"output_tokens"`
}

// AIProvider defines the interface for AI/LLM integrations.
type AIProvider interface {
	// Chat sends a multi-turn conversation and returns a single response.
	Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error)

	// Complete generates a text completion from a prompt.
	Complete(ctx context.Context, req CompletionRequest) (*CompletionResponse, error)

	// Summarize produces a summary of the given text.
	Summarize(ctx context.Context, text string, maxLength int) (string, error)

	// SuggestTitles generates title suggestions for the given content.
	SuggestTitles(ctx context.Context, content string, count int) ([]string, error)

	// SuggestTags suggests tags for the given content, considering existing tags.
	SuggestTags(ctx context.Context, content string, existingTags []string) ([]string, error)

	// StreamChat sends a conversation and returns a channel of streamed chunks.
	StreamChat(ctx context.Context, req ChatRequest) (<-chan ChatChunk, error)

	// Embed returns a vector embedding for the given text.
	Embed(ctx context.Context, text string) ([]float64, error)

	// ChatComplete sends a prompt with context and returns the LLM's response.
	ChatComplete(ctx context.Context, systemPrompt string, userMessage string) (string, error)

	// Name returns the provider name (e.g., "openai", "anthropic", "noop").
	Name() string
}
