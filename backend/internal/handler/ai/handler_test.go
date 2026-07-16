package ai

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"blotting-consultancy/internal/model"
	"blotting-consultancy/internal/provider"
	"blotting-consultancy/internal/repository"
	"blotting-consultancy/internal/service"
	"blotting-consultancy/pkg/secretcipher"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func setupAIHandler(t *testing.T, opts ...service.AIConfigServiceOption) (*Handler, *provider.Registry) {
	t.Helper()

	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.AIConfig{}))

	registry := provider.NewRegistry()
	cipher, err := secretcipher.New("server-secret")
	require.NoError(t, err)
	configSvc := service.NewAIConfigService(
		repository.NewGormAIConfigRepository(db),
		cipher,
		registry,
		opts...,
	)
	return NewHandler(registry, configSvc), registry
}

func TestHandler_UpdateAndGetConfigUseService(t *testing.T) {
	handler, registry := setupAIHandler(t)
	router := gin.New()
	router.PUT("/config", handler.UpdateConfig)
	router.GET("/config", handler.GetConfig)

	req := httptest.NewRequest(http.MethodPut, "/config", strings.NewReader(`{
		"provider":"openai",
		"api_key":"sk-handler-1234",
		"base_url":"https://api.example.test",
		"model":"gpt-handler"
	}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, "openai", registry.AI().Name())

	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/config", nil))
	require.Equal(t, http.StatusOK, rec.Code)

	var body service.AIConfigResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	require.Equal(t, "openai", body.Provider)
	require.True(t, body.Enabled)
	require.True(t, body.HasAPIKey)
	require.Equal(t, "sk-h...1234", body.APIKeyMasked)
	require.Equal(t, "gpt-handler", body.Model)
}

func TestHandler_TestConfigUsesServiceHealthCheck(t *testing.T) {
	client := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		require.Equal(t, "Bearer sk-test", r.Header.Get("Authorization"))
		body, err := json.Marshal(map[string]interface{}{
			"choices": []map[string]interface{}{
				{"message": map[string]string{"content": "ok"}, "finish_reason": "stop"},
			},
			"model": "gpt-handler-health",
		})
		require.NoError(t, err)
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body:       io.NopCloser(strings.NewReader(string(body))),
		}, nil
	})}

	handler, _ := setupAIHandler(t, service.WithAIConfigHTTPClient(client))
	router := gin.New()
	router.POST("/config/test", handler.TestConfig)

	req := httptest.NewRequest(http.MethodPost, "/config/test", strings.NewReader(`{
		"provider":"openai",
		"api_key":"sk-test",
		"base_url":"https://api.example.test",
		"model":"gpt-handler-health"
	}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	var body service.AIHealthResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	require.True(t, body.Healthy)
	require.Equal(t, "openai", body.Provider)
	require.Equal(t, "gpt-handler-health", body.Model)
	require.Equal(t, "ok", body.Message)
}

type dynamicAIProvider struct {
	name string
	text string
}

func (p dynamicAIProvider) Chat(_ context.Context, _ provider.ChatRequest) (*provider.ChatResponse, error) {
	return &provider.ChatResponse{Content: p.text, Model: p.name}, nil
}
func (p dynamicAIProvider) Complete(_ context.Context, _ provider.CompletionRequest) (*provider.CompletionResponse, error) {
	return nil, service.ErrAINotConfigured
}
func (p dynamicAIProvider) Summarize(_ context.Context, _ string, _ int) (string, error) {
	return "", service.ErrAINotConfigured
}
func (p dynamicAIProvider) SuggestTitles(_ context.Context, _ string, _ int) ([]string, error) {
	return nil, service.ErrAINotConfigured
}
func (p dynamicAIProvider) SuggestTags(_ context.Context, _ string, _ []string) ([]string, error) {
	return nil, service.ErrAINotConfigured
}
func (p dynamicAIProvider) StreamChat(_ context.Context, _ provider.ChatRequest) (<-chan provider.ChatChunk, error) {
	return nil, service.ErrAINotConfigured
}
func (p dynamicAIProvider) Embed(_ context.Context, _ string) ([]float64, error) {
	return nil, service.ErrAINotConfigured
}
func (p dynamicAIProvider) ChatComplete(_ context.Context, _ string, _ string) (string, error) {
	return "", service.ErrAINotConfigured
}
func (p dynamicAIProvider) Name() string {
	return p.name
}

func TestHandler_ChatResolvesRegistryDynamically(t *testing.T) {
	handler, registry := setupAIHandler(t)
	router := gin.New()
	router.POST("/chat", handler.Chat)

	registry.SetAI(dynamicAIProvider{name: "first", text: "first response"})
	firstReq := httptest.NewRequest(http.MethodPost, "/chat", strings.NewReader(`{"messages":[{"role":"user","content":"hello"}]}`))
	firstReq.Header.Set("Content-Type", "application/json")
	firstRec := httptest.NewRecorder()
	router.ServeHTTP(firstRec, firstReq)
	require.Equal(t, http.StatusOK, firstRec.Code)
	require.Contains(t, firstRec.Body.String(), "first response")

	registry.SetAI(dynamicAIProvider{name: "second", text: "second response"})
	secondReq := httptest.NewRequest(http.MethodPost, "/chat", strings.NewReader(`{"messages":[{"role":"user","content":"hello"}]}`))
	secondReq.Header.Set("Content-Type", "application/json")
	secondRec := httptest.NewRecorder()
	router.ServeHTTP(secondRec, secondReq)
	require.Equal(t, http.StatusOK, secondRec.Code)
	require.Contains(t, secondRec.Body.String(), "second response")
}
