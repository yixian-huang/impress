package service

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/yixian-huang/inkless/backend/internal/provider"
)

type metaMockAI struct {
	content string
	err     error
	lastReq provider.ChatRequest
}

func (m *metaMockAI) Name() string { return "mock" }
func (m *metaMockAI) Chat(_ context.Context, req provider.ChatRequest) (*provider.ChatResponse, error) {
	m.lastReq = req
	if m.err != nil {
		return nil, m.err
	}
	return &provider.ChatResponse{
		Content:      m.content,
		Model:        "mock-model",
		PromptTokens: 10,
		OutputTokens: 20,
	}, nil
}
func (m *metaMockAI) Complete(context.Context, provider.CompletionRequest) (*provider.CompletionResponse, error) {
	return nil, ErrAINotConfigured
}
func (m *metaMockAI) Summarize(context.Context, string, int) (string, error) {
	return "", ErrAINotConfigured
}
func (m *metaMockAI) SuggestTitles(context.Context, string, int) ([]string, error) {
	return nil, ErrAINotConfigured
}
func (m *metaMockAI) SuggestTags(context.Context, string, []string) ([]string, error) {
	return nil, ErrAINotConfigured
}
func (m *metaMockAI) StreamChat(context.Context, provider.ChatRequest) (<-chan provider.ChatChunk, error) {
	return nil, ErrAINotConfigured
}
func (m *metaMockAI) Embed(context.Context, string) ([]float64, error) {
	return nil, ErrAINotConfigured
}
func (m *metaMockAI) ChatComplete(context.Context, string, string) (string, error) {
	return "", ErrAINotConfigured
}

func longBody() string {
	return strings.Repeat("这是一段用于测试 AI 元数据生成的中文正文。", 10) // > 80 runes
}

func sampleLLMJSON() string {
	p := articleMetaLLMPayload{
		ZhTitles:          []string{"第一标题", "第二标题", "第三标题"},
		EnTitles:          []string{"First Title", "Second Title", "Third Title"},
		Slug:              "First Title!",
		ZhSeoTitle:        "中文 SEO 标题示例足够短",
		EnSeoTitle:        "English SEO Title Example",
		ZhMetaDescription: "这是一段中文 meta 描述，用于搜索引擎结果展示，内容概括文章主题。",
		EnMetaDescription: "An English meta description summarizing the article for search results.",
		ZhExcerpt:         "中文摘要",
		EnExcerpt:         "English excerpt",
		Tags:              []string{"go", "cms"},
	}
	b, _ := json.Marshal(p)
	return string(b)
}

func TestGenerateArticleMeta_Rewrite(t *testing.T) {
	ai := &metaMockAI{content: sampleLLMJSON()}
	resp, err := GenerateArticleMeta(context.Background(), ai, ArticleMetaRequest{
		SourceLang: "zh",
		ZhBody:     longBody(),
		Mode:       "rewrite",
		Fields:     []string{"titles", "slug", "seo", "meta", "tags", "excerpts"},
		TitleCount: 3,
	})
	require.NoError(t, err)
	require.Equal(t, "第一标题", resp.Suggested.ZhTitle)
	require.Equal(t, "First Title", resp.Suggested.EnTitle)
	require.Equal(t, "first-title", resp.Suggested.Slug)
	require.Equal(t, "中文 SEO 标题示例足够短", resp.Suggested.ZhSeoTitle)
	require.Contains(t, resp.Suggested.Tags, "go")
	require.Equal(t, "mock-model", resp.Model)
	require.Empty(t, resp.Skipped)
	require.Len(t, resp.Candidates.ZhTitles, 3)
}

func TestGenerateArticleMeta_FillEmptySkipsExisting(t *testing.T) {
	ai := &metaMockAI{content: sampleLLMJSON()}
	resp, err := GenerateArticleMeta(context.Background(), ai, ArticleMetaRequest{
		SourceLang: "zh",
		ZhBody:     longBody(),
		ZhTitle:    "已有标题",
		Mode:       "fill_empty",
		Existing: ArticleMetaExisting{
			ZhTitle:           "已有标题",
			Slug:              "keep-me",
			ZhMetaDescription: "已有 meta 描述内容",
		},
		Fields: []string{"titles", "slug", "seo", "meta"},
	})
	require.NoError(t, err)
	require.Empty(t, resp.Suggested.ZhTitle)
	require.Empty(t, resp.Suggested.Slug)
	require.Empty(t, resp.Suggested.ZhMetaDescription)
	require.NotEmpty(t, resp.Suggested.EnTitle) // empty existing
	require.Contains(t, resp.Skipped, "zhTitle")
	require.Contains(t, resp.Skipped, "slug")
	require.Contains(t, resp.Skipped, "zhMetaDescription")
}

func TestGenerateArticleMeta_SlugLocked(t *testing.T) {
	ai := &metaMockAI{content: sampleLLMJSON()}
	resp, err := GenerateArticleMeta(context.Background(), ai, ArticleMetaRequest{
		SourceLang: "zh",
		ZhBody:     longBody(),
		Mode:       "rewrite",
		SlugLocked: true,
		Fields:     []string{"slug", "titles"},
	})
	require.NoError(t, err)
	require.Empty(t, resp.Suggested.Slug)
	require.Contains(t, resp.Skipped, "slug")
}

func TestGenerateArticleMeta_ContentTooShort(t *testing.T) {
	ai := &metaMockAI{content: sampleLLMJSON()}
	_, err := GenerateArticleMeta(context.Background(), ai, ArticleMetaRequest{
		SourceLang: "zh",
		ZhBody:     "太短",
		Mode:       "rewrite",
	})
	require.ErrorIs(t, err, ErrArticleMetaContentTooShort)
}

func TestGenerateArticleMeta_NotConfigured(t *testing.T) {
	_, err := GenerateArticleMeta(context.Background(), NewNoopAIProvider(), ArticleMetaRequest{
		SourceLang: "zh",
		ZhBody:     longBody(),
	})
	require.ErrorIs(t, err, ErrAINotConfigured)
}

func TestGenerateArticleMeta_FencedJSON(t *testing.T) {
	ai := &metaMockAI{content: "```json\n" + sampleLLMJSON() + "\n```"}
	resp, err := GenerateArticleMeta(context.Background(), ai, ArticleMetaRequest{
		SourceLang: "zh",
		ZhBody:     longBody(),
		Mode:       "rewrite",
		Fields:     []string{"titles"},
	})
	require.NoError(t, err)
	require.Equal(t, "第一标题", resp.Suggested.ZhTitle)
}

func TestNormalizeSlug(t *testing.T) {
	require.Equal(t, "hello-world", normalizeSlug("Hello World!"))
	require.Equal(t, "abc-123", normalizeSlug("  ABC  123  "))
	require.Equal(t, "", normalizeSlug("中文标题"))
}

func TestPlainTextFromHTML(t *testing.T) {
	got := plainTextFromHTML(`<p>Hello <b>world</b></p><script>x</script>`)
	require.Contains(t, got, "Hello")
	require.Contains(t, got, "world")
	require.NotContains(t, got, "script")
}
