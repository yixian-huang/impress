package service

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/yixian-huang/inkless/backend/internal/provider"
)

const anthropicHTTPClientTimeout = 120 * time.Second

// AnthropicProvider implements AIProvider using the Anthropic Messages API.
type AnthropicProvider struct {
	apiKey  string
	baseURL string
	model   string
	client  *http.Client
}

// AnthropicConfig holds configuration for the Anthropic provider.
type AnthropicConfig struct {
	APIKey  string
	BaseURL string // defaults to "https://api.anthropic.com/v1"
	Model   string // defaults to "claude-sonnet-4-20250514"
	Client  *http.Client
}

// NewAnthropicProvider creates a new Anthropic AI provider.
func NewAnthropicProvider(cfg AnthropicConfig) *AnthropicProvider {
	baseURL := cfg.BaseURL
	if baseURL == "" {
		baseURL = "https://api.anthropic.com/v1"
	}
	model := cfg.Model
	if model == "" {
		model = "claude-sonnet-4-20250514"
	}
	client := cfg.Client
	if client == nil {
		client = &http.Client{Timeout: anthropicHTTPClientTimeout}
	}
	return &AnthropicProvider{
		apiKey:  cfg.APIKey,
		baseURL: strings.TrimRight(baseURL, "/"),
		model:   model,
		client:  client,
	}
}

func (a *AnthropicProvider) Name() string {
	return "anthropic"
}

// Anthropic API types

type anthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type anthropicRequest struct {
	Model       string             `json:"model"`
	MaxTokens   int                `json:"max_tokens"`
	System      string             `json:"system,omitempty"`
	Messages    []anthropicMessage `json:"messages"`
	Temperature float64            `json:"temperature,omitempty"`
	Stream      bool               `json:"stream,omitempty"`
}

type anthropicResponse struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Model      string `json:"model"`
	StopReason string `json:"stop_reason"`
	Usage      struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

type anthropicErrorResponse struct {
	Error struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error"`
}

// anthropicStreamEvent represents a server-sent event from the Anthropic streaming API.
type anthropicStreamEvent struct {
	Type  string          `json:"type"`
	Delta json.RawMessage `json:"delta,omitempty"`
}

type anthropicContentDelta struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

func (a *AnthropicProvider) Chat(ctx context.Context, req provider.ChatRequest) (*provider.ChatResponse, error) {
	model := req.Model
	if model == "" {
		model = a.model
	}
	maxTokens := req.MaxTokens
	if maxTokens == 0 {
		maxTokens = 1024
	}

	// Extract system message if present
	var system string
	var msgs []anthropicMessage
	for _, m := range req.Messages {
		if m.Role == "system" {
			system = m.Content
			continue
		}
		msgs = append(msgs, anthropicMessage{Role: m.Role, Content: m.Content})
	}

	body := anthropicRequest{
		Model:       model,
		MaxTokens:   maxTokens,
		System:      system,
		Messages:    msgs,
		Temperature: req.Temperature,
	}

	respBody, err := a.doRequest(ctx, "/messages", body)
	if err != nil {
		return nil, err
	}
	defer respBody.Close()

	var antResp anthropicResponse
	if err := json.NewDecoder(respBody).Decode(&antResp); err != nil {
		return nil, fmt.Errorf("anthropic: failed to decode response: %w", err)
	}

	content := ""
	for _, block := range antResp.Content {
		if block.Type == "text" {
			content += block.Text
		}
	}

	return &provider.ChatResponse{
		Content:      content,
		Model:        antResp.Model,
		FinishReason: antResp.StopReason,
		PromptTokens: antResp.Usage.InputTokens,
		OutputTokens: antResp.Usage.OutputTokens,
	}, nil
}

func (a *AnthropicProvider) Complete(ctx context.Context, req provider.CompletionRequest) (*provider.CompletionResponse, error) {
	chatReq := provider.ChatRequest{
		Messages:    []provider.ChatMessage{{Role: "user", Content: req.Prompt}},
		Model:       req.Model,
		MaxTokens:   req.MaxTokens,
		Temperature: req.Temperature,
	}
	resp, err := a.Chat(ctx, chatReq)
	if err != nil {
		return nil, err
	}
	return &provider.CompletionResponse{
		Text:         resp.Content,
		Model:        resp.Model,
		FinishReason: resp.FinishReason,
		PromptTokens: resp.PromptTokens,
		OutputTokens: resp.OutputTokens,
	}, nil
}

func (a *AnthropicProvider) Summarize(ctx context.Context, text string, maxLength int) (string, error) {
	prompt := fmt.Sprintf("Summarize the following text in at most %d characters. Return only the summary, no extra commentary.\n\n%s", maxLength, text)
	resp, err := a.Chat(ctx, provider.ChatRequest{
		Messages:  []provider.ChatMessage{{Role: "user", Content: prompt}},
		MaxTokens: maxLength / 2,
	})
	if err != nil {
		return "", err
	}
	return resp.Content, nil
}

func (a *AnthropicProvider) SuggestTitles(ctx context.Context, content string, count int) ([]string, error) {
	prompt := fmt.Sprintf("Suggest %d concise titles for the following content. Return as a JSON array of strings, nothing else.\n\n%s", count, content)
	resp, err := a.Chat(ctx, provider.ChatRequest{
		Messages:  []provider.ChatMessage{{Role: "user", Content: prompt}},
		MaxTokens: 500,
	})
	if err != nil {
		return nil, err
	}
	return parseJSONStringArray(resp.Content)
}

func (a *AnthropicProvider) SuggestTags(ctx context.Context, content string, existingTags []string) ([]string, error) {
	existingStr := "none"
	if len(existingTags) > 0 {
		existingStr = strings.Join(existingTags, ", ")
	}
	prompt := fmt.Sprintf("Suggest relevant tags for the following content. Existing tags: [%s]. Return new tag suggestions as a JSON array of strings, nothing else.\n\n%s", existingStr, content)
	resp, err := a.Chat(ctx, provider.ChatRequest{
		Messages:  []provider.ChatMessage{{Role: "user", Content: prompt}},
		MaxTokens: 300,
	})
	if err != nil {
		return nil, err
	}
	return parseJSONStringArray(resp.Content)
}

func (a *AnthropicProvider) StreamChat(ctx context.Context, req provider.ChatRequest) (<-chan provider.ChatChunk, error) {
	model := req.Model
	if model == "" {
		model = a.model
	}
	maxTokens := req.MaxTokens
	if maxTokens == 0 {
		maxTokens = 1024
	}

	var system string
	var msgs []anthropicMessage
	for _, m := range req.Messages {
		if m.Role == "system" {
			system = m.Content
			continue
		}
		msgs = append(msgs, anthropicMessage{Role: m.Role, Content: m.Content})
	}

	body := anthropicRequest{
		Model:       model,
		MaxTokens:   maxTokens,
		System:      system,
		Messages:    msgs,
		Temperature: req.Temperature,
		Stream:      true,
	}

	respBody, err := a.doRequest(ctx, "/messages", body)
	if err != nil {
		return nil, err
	}

	ch := make(chan provider.ChatChunk, 32)
	go func() {
		defer close(ch)
		defer respBody.Close()

		scanner := bufio.NewScanner(respBody)
		for scanner.Scan() {
			line := scanner.Text()
			if !strings.HasPrefix(line, "data: ") {
				continue
			}
			data := strings.TrimPrefix(line, "data: ")

			var event anthropicStreamEvent
			if err := json.Unmarshal([]byte(data), &event); err != nil {
				ch <- provider.ChatChunk{Err: fmt.Errorf("anthropic: stream parse error: %w", err)}
				return
			}

			switch event.Type {
			case "content_block_delta":
				var delta anthropicContentDelta
				if err := json.Unmarshal(event.Delta, &delta); err == nil {
					select {
					case ch <- provider.ChatChunk{Content: delta.Text}:
					case <-ctx.Done():
						return
					}
				}
			case "message_stop":
				ch <- provider.ChatChunk{FinishReason: "end_turn"}
				return
			case "error":
				ch <- provider.ChatChunk{Err: fmt.Errorf("anthropic: stream error event")}
				return
			}
		}
		if err := scanner.Err(); err != nil {
			ch <- provider.ChatChunk{Err: fmt.Errorf("anthropic: stream read error: %w", err)}
		}
	}()

	return ch, nil
}

// doRequest sends a POST request to the Anthropic API and returns the response body.
func (a *AnthropicProvider) doRequest(ctx context.Context, path string, body interface{}) (io.ReadCloser, error) {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("anthropic: failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", a.baseURL+path, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("anthropic: failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", a.apiKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	resp, err := a.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("anthropic: request failed: %w", err)
	}

	if resp.StatusCode >= 400 {
		defer resp.Body.Close()
		bodyBytes, _ := io.ReadAll(resp.Body)
		var errResp anthropicErrorResponse
		if json.Unmarshal(bodyBytes, &errResp) == nil && errResp.Error.Message != "" {
			return nil, fmt.Errorf("anthropic: API error (%d): %s", resp.StatusCode, errResp.Error.Message)
		}
		return nil, fmt.Errorf("anthropic: API error (%d): %s", resp.StatusCode, string(bodyBytes))
	}

	return resp.Body, nil
}

// Embed returns a vector embedding for the given text.
// Anthropic does not natively support embeddings, so this returns an error.
func (a *AnthropicProvider) Embed(_ context.Context, _ string) ([]float64, error) {
	return nil, fmt.Errorf("anthropic: embeddings not supported, use an OpenAI-compatible provider")
}

// ChatComplete sends a prompt with context and returns the LLM's response.
func (a *AnthropicProvider) ChatComplete(ctx context.Context, systemPrompt string, userMessage string) (string, error) {
	resp, err := a.Chat(ctx, provider.ChatRequest{
		Messages: []provider.ChatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userMessage},
		},
		MaxTokens: 2048,
	})
	if err != nil {
		return "", err
	}
	return resp.Content, nil
}
