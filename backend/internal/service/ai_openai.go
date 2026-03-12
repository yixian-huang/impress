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

	"blotting-consultancy/internal/provider"
)

// OpenAIProvider implements AIProvider using the OpenAI-compatible API.
type OpenAIProvider struct {
	apiKey  string
	baseURL string
	model   string
	client  *http.Client
}

// OpenAIConfig holds configuration for the OpenAI provider.
type OpenAIConfig struct {
	APIKey  string
	BaseURL string // defaults to "https://api.openai.com/v1"
	Model   string // defaults to "gpt-4o-mini"
	Client  *http.Client
}

// NewOpenAIProvider creates a new OpenAI AI provider.
func NewOpenAIProvider(cfg OpenAIConfig) *OpenAIProvider {
	baseURL := cfg.BaseURL
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}
	model := cfg.Model
	if model == "" {
		model = "gpt-4o-mini"
	}
	client := cfg.Client
	if client == nil {
		client = http.DefaultClient
	}
	return &OpenAIProvider{
		apiKey:  cfg.APIKey,
		baseURL: strings.TrimRight(baseURL, "/"),
		model:   model,
		client:  client,
	}
}

func (o *OpenAIProvider) Name() string {
	return "openai"
}

// openAIMessage is the message format for the OpenAI API.
type openAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAIChatRequest struct {
	Model       string          `json:"model"`
	Messages    []openAIMessage `json:"messages"`
	MaxTokens   int             `json:"max_tokens,omitempty"`
	Temperature float64         `json:"temperature,omitempty"`
	Stream      bool            `json:"stream,omitempty"`
}

type openAIChatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Model string `json:"model"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
	} `json:"usage"`
}

type openAIStreamChunk struct {
	Choices []struct {
		Delta struct {
			Content string `json:"content"`
		} `json:"delta"`
		FinishReason *string `json:"finish_reason"`
	} `json:"choices"`
}

type openAIErrorResponse struct {
	Error struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error"`
}

func (o *OpenAIProvider) Chat(ctx context.Context, req provider.ChatRequest) (*provider.ChatResponse, error) {
	model := req.Model
	if model == "" {
		model = o.model
	}

	msgs := make([]openAIMessage, len(req.Messages))
	for i, m := range req.Messages {
		msgs[i] = openAIMessage{Role: m.Role, Content: m.Content}
	}

	body := openAIChatRequest{
		Model:       model,
		Messages:    msgs,
		MaxTokens:   req.MaxTokens,
		Temperature: req.Temperature,
	}

	respBody, err := o.doRequest(ctx, "/chat/completions", body)
	if err != nil {
		return nil, err
	}
	defer respBody.Close()

	var oaiResp openAIChatResponse
	if err := json.NewDecoder(respBody).Decode(&oaiResp); err != nil {
		return nil, fmt.Errorf("openai: failed to decode response: %w", err)
	}

	if len(oaiResp.Choices) == 0 {
		return nil, fmt.Errorf("openai: empty response")
	}

	return &provider.ChatResponse{
		Content:      oaiResp.Choices[0].Message.Content,
		Model:        oaiResp.Model,
		FinishReason: oaiResp.Choices[0].FinishReason,
		PromptTokens: oaiResp.Usage.PromptTokens,
		OutputTokens: oaiResp.Usage.CompletionTokens,
	}, nil
}

func (o *OpenAIProvider) Complete(ctx context.Context, req provider.CompletionRequest) (*provider.CompletionResponse, error) {
	chatReq := provider.ChatRequest{
		Messages:    []provider.ChatMessage{{Role: "user", Content: req.Prompt}},
		Model:       req.Model,
		MaxTokens:   req.MaxTokens,
		Temperature: req.Temperature,
	}
	resp, err := o.Chat(ctx, chatReq)
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

func (o *OpenAIProvider) Summarize(ctx context.Context, text string, maxLength int) (string, error) {
	prompt := fmt.Sprintf("Summarize the following text in at most %d characters. Return only the summary, no extra commentary.\n\n%s", maxLength, text)
	resp, err := o.Chat(ctx, provider.ChatRequest{
		Messages:  []provider.ChatMessage{{Role: "user", Content: prompt}},
		MaxTokens: maxLength / 2, // rough estimate
	})
	if err != nil {
		return "", err
	}
	return resp.Content, nil
}

func (o *OpenAIProvider) SuggestTitles(ctx context.Context, content string, count int) ([]string, error) {
	prompt := fmt.Sprintf("Suggest %d concise titles for the following content. Return as a JSON array of strings, nothing else.\n\n%s", count, content)
	resp, err := o.Chat(ctx, provider.ChatRequest{
		Messages:  []provider.ChatMessage{{Role: "user", Content: prompt}},
		MaxTokens: 500,
	})
	if err != nil {
		return nil, err
	}
	return parseJSONStringArray(resp.Content)
}

func (o *OpenAIProvider) SuggestTags(ctx context.Context, content string, existingTags []string) ([]string, error) {
	existingStr := "none"
	if len(existingTags) > 0 {
		existingStr = strings.Join(existingTags, ", ")
	}
	prompt := fmt.Sprintf("Suggest relevant tags for the following content. Existing tags: [%s]. Return new tag suggestions as a JSON array of strings, nothing else.\n\n%s", existingStr, content)
	resp, err := o.Chat(ctx, provider.ChatRequest{
		Messages:  []provider.ChatMessage{{Role: "user", Content: prompt}},
		MaxTokens: 300,
	})
	if err != nil {
		return nil, err
	}
	return parseJSONStringArray(resp.Content)
}

func (o *OpenAIProvider) StreamChat(ctx context.Context, req provider.ChatRequest) (<-chan provider.ChatChunk, error) {
	model := req.Model
	if model == "" {
		model = o.model
	}

	msgs := make([]openAIMessage, len(req.Messages))
	for i, m := range req.Messages {
		msgs[i] = openAIMessage{Role: m.Role, Content: m.Content}
	}

	body := openAIChatRequest{
		Model:       model,
		Messages:    msgs,
		MaxTokens:   req.MaxTokens,
		Temperature: req.Temperature,
		Stream:      true,
	}

	respBody, err := o.doRequest(ctx, "/chat/completions", body)
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
			if data == "[DONE]" {
				return
			}
			var chunk openAIStreamChunk
			if err := json.Unmarshal([]byte(data), &chunk); err != nil {
				ch <- provider.ChatChunk{Err: fmt.Errorf("openai: stream parse error: %w", err)}
				return
			}
			if len(chunk.Choices) > 0 {
				c := provider.ChatChunk{
					Content: chunk.Choices[0].Delta.Content,
				}
				if chunk.Choices[0].FinishReason != nil {
					c.FinishReason = *chunk.Choices[0].FinishReason
				}
				select {
				case ch <- c:
				case <-ctx.Done():
					return
				}
			}
		}
		if err := scanner.Err(); err != nil {
			ch <- provider.ChatChunk{Err: fmt.Errorf("openai: stream read error: %w", err)}
		}
	}()

	return ch, nil
}

// doRequest sends a POST request to the OpenAI API and returns the response body.
func (o *OpenAIProvider) doRequest(ctx context.Context, path string, body interface{}) (io.ReadCloser, error) {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("openai: failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", o.baseURL+path, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("openai: failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+o.apiKey)

	resp, err := o.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("openai: request failed: %w", err)
	}

	if resp.StatusCode >= 400 {
		defer resp.Body.Close()
		bodyBytes, _ := io.ReadAll(resp.Body)
		var errResp openAIErrorResponse
		if json.Unmarshal(bodyBytes, &errResp) == nil && errResp.Error.Message != "" {
			return nil, fmt.Errorf("openai: API error (%d): %s", resp.StatusCode, errResp.Error.Message)
		}
		return nil, fmt.Errorf("openai: API error (%d): %s", resp.StatusCode, string(bodyBytes))
	}

	return resp.Body, nil
}

// parseJSONStringArray extracts a JSON string array from LLM output.
// It handles the case where the LLM wraps the array in markdown code fences.
func parseJSONStringArray(text string) ([]string, error) {
	text = strings.TrimSpace(text)
	// Strip markdown code fences if present
	if strings.HasPrefix(text, "```") {
		lines := strings.Split(text, "\n")
		// Remove first and last lines (fences)
		if len(lines) >= 2 {
			lines = lines[1 : len(lines)-1]
			// Remove trailing fence if present
			if len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) == "```" {
				lines = lines[:len(lines)-1]
			}
		}
		text = strings.Join(lines, "\n")
	}

	var result []string
	if err := json.Unmarshal([]byte(text), &result); err != nil {
		return nil, fmt.Errorf("failed to parse AI response as string array: %w", err)
	}
	return result, nil
}

// Embed returns a vector embedding for the given text using the OpenAI embeddings API.
func (o *OpenAIProvider) Embed(ctx context.Context, text string) ([]float64, error) {
	reqBody := map[string]interface{}{
		"model": "text-embedding-3-small",
		"input": text,
	}
	body, _ := json.Marshal(reqBody)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, o.baseURL+"/embeddings", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+o.apiKey)

	resp, err := o.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("openai embed request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("openai embed error (%d): %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Data []struct {
			Embedding []float64 `json:"embedding"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode embedding response: %w", err)
	}
	if len(result.Data) == 0 {
		return nil, fmt.Errorf("openai: empty embedding response")
	}
	return result.Data[0].Embedding, nil
}

// ChatComplete sends a prompt with context and returns the LLM's response.
func (o *OpenAIProvider) ChatComplete(ctx context.Context, systemPrompt string, userMessage string) (string, error) {
	resp, err := o.Chat(ctx, provider.ChatRequest{
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
