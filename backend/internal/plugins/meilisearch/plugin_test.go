package meilisearch

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew_MissingHost(t *testing.T) {
	_, err := New(Config{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "host is required")
}

func TestNew_Valid(t *testing.T) {
	p, err := New(Config{Host: "http://localhost:7700"})
	require.NoError(t, err)
	assert.NotNil(t, p)
	assert.Equal(t, "inkless_", p.indexPrefix)
}

func TestNew_CustomPrefix(t *testing.T) {
	p, err := New(Config{Host: "http://localhost:7700", IndexPrefix: "cms_"})
	require.NoError(t, err)
	assert.Equal(t, "cms_", p.indexPrefix)
}

func TestNewFromSettings(t *testing.T) {
	settings := map[string]string{
		"host":         "http://meilisearch:7700",
		"api_key":      "masterKey",
		"index_prefix": "test_",
	}
	p, err := NewFromSettings(settings)
	require.NoError(t, err)
	assert.Equal(t, "http://meilisearch:7700", p.config.Host)
	assert.Equal(t, "masterKey", p.config.APIKey)
	assert.Equal(t, "test_", p.indexPrefix)
}

func TestIndexName(t *testing.T) {
	p, _ := New(Config{Host: "http://localhost:7700"})
	assert.Equal(t, "inkless_articles_zh", p.indexName("articles", "zh"))
	assert.Equal(t, "inkless_pages_en", p.indexName("pages", "en"))
}

func TestNewFromSettings_DefaultsToCanonicalPrefix(t *testing.T) {
	p, err := NewFromSettings(map[string]string{"host": "http://localhost:7700"})
	require.NoError(t, err)
	assert.Equal(t, "inkless_", p.indexPrefix)
}

func TestNewFromSettings_HonorsMigratedLegacyPrefix(t *testing.T) {
	p, err := NewFromSettings(map[string]string{
		"host":         "http://localhost:7700",
		"index_prefix": legacyIndexPrefix,
	})
	require.NoError(t, err)
	assert.Equal(t, "impress_", p.indexPrefix)
}

func TestNewFromSettings_customPrefix(t *testing.T) {
	p, err := NewFromSettings(map[string]string{
		"host":         "http://localhost:7700",
		"index_prefix": "cms_",
	})
	require.NoError(t, err)
	assert.Equal(t, "cms_", p.indexPrefix)
}

func TestSearch_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Contains(t, r.URL.Path, "/search")

		resp := searchResponse{
			Hits: []indexedDoc{
				{
					ID:        "article_1",
					NumericID: 1,
					Type:      "article",
					Locale:    "zh",
					Title:     "Test Article",
					Body:      "Some body content",
					Slug:      "test-article",
					Score:     0.95,
				},
			},
			EstimatedTotalHits: 1,
			Query:              "test",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	p, _ := New(Config{Host: srv.URL})
	p.httpClient = srv.Client()

	result, err := p.Search(context.Background(), "test", "zh", "article", 1, 10)
	require.NoError(t, err)
	assert.Len(t, result.Results, 1)
	assert.Equal(t, "Test Article", result.Results[0].Title)
	assert.Equal(t, "zh", result.Results[0].Locale)
	assert.Equal(t, "test", result.Query)
}

func TestSearch_IndexNotFound_ReturnsEmpty(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	p, _ := New(Config{Host: srv.URL})
	p.httpClient = srv.Client()

	result, err := p.Search(context.Background(), "test", "zh", "article", 1, 10)
	require.NoError(t, err)
	assert.Empty(t, result.Results)
	assert.Equal(t, int64(0), result.Total)
}

func TestSearch_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal error"))
	}))
	defer srv.Close()

	p, _ := New(Config{Host: srv.URL})
	p.httpClient = srv.Client()

	_, err := p.Search(context.Background(), "test", "zh", "article", 1, 10)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "500")
}

func TestSearch_SnippetTruncation(t *testing.T) {
	longBody := string(make([]byte, 300))
	for i := range longBody {
		longBody = longBody[:i] + "x" + longBody[i+1:]
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := searchResponse{
			Hits: []indexedDoc{
				{
					ID:        "article_1",
					NumericID: 1,
					Type:      "article",
					Locale:    "zh",
					Title:     "Long Article",
					Body:      longBody,
					Slug:      "long",
				},
			},
			EstimatedTotalHits: 1,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	p, _ := New(Config{Host: srv.URL})
	p.httpClient = srv.Client()

	result, err := p.Search(context.Background(), "x", "zh", "article", 1, 10)
	require.NoError(t, err)
	require.Len(t, result.Results, 1)
	assert.LessOrEqual(t, len(result.Results[0].Snippet), 203) // 200 + "..."
}

func TestIndexArticle_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/tasks/") {
			json.NewEncoder(w).Encode(taskStatusResponse{Status: "succeeded"})
			return
		}
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Contains(t, r.URL.Path, "/documents")
		w.WriteHeader(http.StatusAccepted)
		w.Write([]byte(`{"taskUid":1}`))
	}))
	defer srv.Close()

	p, _ := New(Config{Host: srv.URL})
	p.httpClient = srv.Client()

	err := p.IndexArticle(context.Background(), 42, "zh", "Title", "Body text", "my-article")
	require.NoError(t, err)
}

func TestIndexPage_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/tasks/") {
			json.NewEncoder(w).Encode(taskStatusResponse{Status: "succeeded"})
			return
		}
		w.WriteHeader(http.StatusAccepted)
		w.Write([]byte(`{"taskUid":2}`))
	}))
	defer srv.Close()

	p, _ := New(Config{Host: srv.URL})
	p.httpClient = srv.Client()

	err := p.IndexPage(context.Background(), 7, "en", "Page Title", "Page content", "my-page")
	require.NoError(t, err)
}

func TestRemoveFromIndex_Success(t *testing.T) {
	deleteCount := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			deleteCount++
			w.WriteHeader(http.StatusAccepted)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	p, _ := New(Config{Host: srv.URL})
	p.httpClient = srv.Client()

	err := p.RemoveFromIndex(context.Background(), "article", 42)
	require.NoError(t, err)
	// Should attempt deletion from both zh and en indexes
	assert.Equal(t, 2, deleteCount)
}

func TestRebuildIndex_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/tasks/") {
			json.NewEncoder(w).Encode(taskStatusResponse{Status: "succeeded"})
			return
		}
		switch r.Method {
		case http.MethodDelete:
			w.WriteHeader(http.StatusAccepted)
			w.Write([]byte(`{"taskUid":1}`))
		case http.MethodPost:
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(`{"uid":"test","primaryKey":"id"}`))
		default:
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer srv.Close()

	p, _ := New(Config{Host: srv.URL})
	p.httpClient = srv.Client()

	err := p.RebuildIndex(context.Background())
	require.NoError(t, err)
}

func TestReindexLegacyToInkless_CopiesAndDeletesAfterCountMatches(t *testing.T) {
	docs := []indexedDoc{
		{ID: "article_1", NumericID: 1, Type: "article", Locale: "zh", Title: "Hello", Body: "Body", Slug: "hello"},
	}
	created := false
	deletedLegacy := false
	targetDocs := 0
	taskPolls := map[string]int{}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/tasks/") {
			taskPolls[r.URL.Path]++
			if r.URL.Path == "/tasks/11" && taskPolls[r.URL.Path] == 1 {
				json.NewEncoder(w).Encode(taskStatusResponse{Status: "processing"})
				return
			}
			if r.URL.Path == "/tasks/11" {
				targetDocs = len(docs)
			}
			json.NewEncoder(w).Encode(taskStatusResponse{Status: "succeeded"})
			return
		}
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/indexes/impress_articles_zh/stats":
			json.NewEncoder(w).Encode(indexStatsResponse{NumberOfDocuments: len(docs)})
		case r.Method == http.MethodGet && r.URL.Path == "/indexes/inkless_articles_zh/stats":
			json.NewEncoder(w).Encode(indexStatsResponse{NumberOfDocuments: targetDocs})
		case r.Method == http.MethodGet && r.URL.Path == "/indexes/impress_articles_zh/documents":
			assert.Equal(t, "1", r.URL.Query().Get("limit"))
			json.NewEncoder(w).Encode(documentsResponse{Results: docs})
		case r.Method == http.MethodPost && r.URL.Path == "/indexes":
			created = true
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(`{"taskUid":10}`))
		case r.Method == http.MethodPost && r.URL.Path == "/indexes/inkless_articles_zh/documents":
			var posted []indexedDoc
			require.NoError(t, json.NewDecoder(r.Body).Decode(&posted))
			assert.Equal(t, len(docs), len(posted))
			w.WriteHeader(http.StatusAccepted)
			w.Write([]byte(`{"taskUid":11}`))
		case r.Method == http.MethodDelete && r.URL.Path == "/indexes/impress_articles_zh":
			assert.Equal(t, len(docs), targetDocs, "legacy index must not be deleted before target count is verified")
			deletedLegacy = true
			w.WriteHeader(http.StatusAccepted)
			w.Write([]byte(`{"taskUid":12}`))
		case r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/indexes/impress_"):
			w.WriteHeader(http.StatusNotFound)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	p, err := NewFromSettings(map[string]string{"host": srv.URL, "index_prefix": legacyIndexPrefix})
	require.NoError(t, err)
	p.httpClient = srv.Client()

	err = p.ReindexLegacyToInkless(context.Background(), ReindexOptions{Locales: []string{"zh"}, DeleteLegacy: true})
	require.NoError(t, err)
	assert.True(t, created)
	assert.True(t, deletedLegacy)
	assert.Equal(t, defaultIndexPrefix, p.indexPrefix)
}

func TestReindexLegacyToInkless_TaskFailureBlocksDelete(t *testing.T) {
	deletedLegacy := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/tasks/42" {
			resp := taskStatusResponse{Status: "failed"}
			resp.Error.Message = "import failed"
			json.NewEncoder(w).Encode(resp)
			return
		}
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/indexes/impress_articles_zh/stats":
			json.NewEncoder(w).Encode(indexStatsResponse{NumberOfDocuments: 1})
		case r.Method == http.MethodGet && r.URL.Path == "/indexes/inkless_articles_zh/stats":
			json.NewEncoder(w).Encode(indexStatsResponse{NumberOfDocuments: 0})
		case r.Method == http.MethodGet && r.URL.Path == "/indexes/impress_articles_zh/documents":
			json.NewEncoder(w).Encode(documentsResponse{Results: []indexedDoc{
				{ID: "article_1", NumericID: 1, Type: "article", Locale: "zh", Title: "Article", Slug: "article"},
			}})
		case r.Method == http.MethodPost && r.URL.Path == "/indexes":
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(`{"taskUid":42}`))
		case r.Method == http.MethodDelete && strings.HasPrefix(r.URL.Path, "/indexes/impress_"):
			deletedLegacy = true
			w.WriteHeader(http.StatusAccepted)
		case r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/indexes/impress_"):
			w.WriteHeader(http.StatusNotFound)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	p, err := NewFromSettings(map[string]string{"host": srv.URL, "index_prefix": legacyIndexPrefix})
	require.NoError(t, err)
	p.httpClient = srv.Client()

	err = p.ReindexLegacyToInkless(context.Background(), ReindexOptions{Locales: []string{"zh"}, DeleteLegacy: true})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "import failed")
	assert.False(t, deletedLegacy)
	assert.Equal(t, legacyIndexPrefix, p.indexPrefix)
}

func TestReindexLegacyToInkless_IsIdempotentWhenTargetCountMatches(t *testing.T) {
	deleteCount := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/indexes/impress_articles_en/stats":
			json.NewEncoder(w).Encode(indexStatsResponse{NumberOfDocuments: 2})
		case r.Method == http.MethodGet && r.URL.Path == "/indexes/inkless_articles_en/stats":
			json.NewEncoder(w).Encode(indexStatsResponse{NumberOfDocuments: 2})
		case r.Method == http.MethodDelete && r.URL.Path == "/indexes/impress_articles_en":
			deleteCount++
			w.WriteHeader(http.StatusAccepted)
		case r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/indexes/impress_"):
			w.WriteHeader(http.StatusNotFound)
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.String())
		}
	}))
	defer srv.Close()

	p, err := NewFromSettings(map[string]string{"host": srv.URL, "index_prefix": legacyIndexPrefix})
	require.NoError(t, err)
	p.httpClient = srv.Client()

	err = p.ReindexLegacyToInkless(context.Background(), ReindexOptions{Locales: []string{"en"}, DeleteLegacy: true})
	require.NoError(t, err)
	assert.Equal(t, 1, deleteCount)
	assert.Equal(t, defaultIndexPrefix, p.indexPrefix)
}

func TestReindexLegacyToInkless_CountMismatchBlocksCutover(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/indexes/impress_pages_zh/stats":
			json.NewEncoder(w).Encode(indexStatsResponse{NumberOfDocuments: 1})
		case r.Method == http.MethodGet && r.URL.Path == "/indexes/inkless_pages_zh/stats":
			json.NewEncoder(w).Encode(indexStatsResponse{NumberOfDocuments: 0})
		case r.Method == http.MethodGet && r.URL.Path == "/indexes/impress_pages_zh/documents":
			json.NewEncoder(w).Encode(documentsResponse{Results: []indexedDoc{
				{ID: "page_1", NumericID: 1, Type: "page", Locale: "zh", Title: "Page", Slug: "page"},
			}})
		case r.Method == http.MethodPost && r.URL.Path == "/indexes":
			w.WriteHeader(http.StatusCreated)
		case r.Method == http.MethodPost && r.URL.Path == "/indexes/inkless_pages_zh/documents":
			w.WriteHeader(http.StatusAccepted)
		case r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/indexes/impress_"):
			w.WriteHeader(http.StatusNotFound)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	p, err := NewFromSettings(map[string]string{"host": srv.URL, "index_prefix": legacyIndexPrefix})
	require.NoError(t, err)
	p.httpClient = srv.Client()

	err = p.ReindexLegacyToInkless(context.Background(), ReindexOptions{Locales: []string{"zh"}})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "count mismatch")
	assert.Equal(t, legacyIndexPrefix, p.indexPrefix)
}

func TestReindexLegacyToInkless_LaterMismatchKeepsEveryLegacyIndex(t *testing.T) {
	deleteCount := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/indexes/impress_articles_zh/stats":
			json.NewEncoder(w).Encode(indexStatsResponse{NumberOfDocuments: 1})
		case r.Method == http.MethodGet && r.URL.Path == "/indexes/inkless_articles_zh/stats":
			json.NewEncoder(w).Encode(indexStatsResponse{NumberOfDocuments: 1})
		case r.Method == http.MethodGet && r.URL.Path == "/indexes/impress_pages_zh/stats":
			json.NewEncoder(w).Encode(indexStatsResponse{NumberOfDocuments: 1})
		case r.Method == http.MethodGet && r.URL.Path == "/indexes/inkless_pages_zh/stats":
			json.NewEncoder(w).Encode(indexStatsResponse{NumberOfDocuments: 0})
		case r.Method == http.MethodGet && r.URL.Path == "/indexes/impress_pages_zh/documents":
			json.NewEncoder(w).Encode(documentsResponse{Results: []indexedDoc{{ID: "page_1", NumericID: 1}}})
		case r.Method == http.MethodPost && r.URL.Path == "/indexes":
			w.WriteHeader(http.StatusConflict)
		case r.Method == http.MethodPost && r.URL.Path == "/indexes/inkless_pages_zh/documents":
			w.WriteHeader(http.StatusAccepted)
			w.Write([]byte(`{}`))
		case r.Method == http.MethodDelete:
			deleteCount++
			w.WriteHeader(http.StatusAccepted)
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.String())
		}
	}))
	defer srv.Close()

	p, err := NewFromSettings(map[string]string{"host": srv.URL, "index_prefix": legacyIndexPrefix})
	require.NoError(t, err)
	p.httpClient = srv.Client()

	err = p.ReindexLegacyToInkless(context.Background(), ReindexOptions{Locales: []string{"zh"}, DeleteLegacy: true})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "count mismatch")
	assert.Zero(t, deleteCount)
	assert.Equal(t, legacyIndexPrefix, p.indexPrefix)
}

func TestSuggest_ReturnsTitles(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := searchResponse{
			Hits: []indexedDoc{
				{ID: "article_1", NumericID: 1, Type: "article", Locale: "zh", Title: "Alpha", Slug: "alpha"},
				{ID: "article_2", NumericID: 2, Type: "article", Locale: "zh", Title: "Beta", Slug: "beta"},
			},
			EstimatedTotalHits: 2,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	p, _ := New(Config{Host: srv.URL})
	p.httpClient = srv.Client()

	suggestions, err := p.Suggest(context.Background(), "al", "zh", 5)
	require.NoError(t, err)
	assert.Contains(t, suggestions, "Alpha")
}

func TestManifest(t *testing.T) {
	err := Manifest.Validate()
	require.NoError(t, err)
}
