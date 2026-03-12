# Phase 5: AI Intelligence Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Differentiate from Halo by integrating AI capabilities: content generation assistance in the editor, one-click bilingual translation, RAG-powered knowledge base Q&A for site visitors, and an AI-driven site setup wizard.

**Architecture:** All AI features are built on provider interfaces registered in the existing `internal/provider/` Registry. Backend handlers follow the handler/service/repository layering. Frontend components integrate into the existing TipTap editor and admin UI. The AI features are optional — the system works without any AI provider configured.

**Tech Stack:** Go/Gin/GORM (backend), OpenAI/Claude/Ollama SDKs, TipTap (editor integration), i18next (bilingual), React (frontend widgets)

**Spec:** `docs/superpowers/specs/2026-03-11-open-source-evolution-design.md` — Phase 5

**Prerequisites:** Phase 0-4 complete. Key dependencies: Provider Registry (`internal/provider/registry.go`), EventBus (`internal/eventbus/`), Article model with bilingual fields (`ZhTitle`/`EnTitle`, `ZhBody`/`EnBody`, `ZhMetaDescription`/`EnMetaDescription`), TipTap editor, i18next, page config system.

**Convention note:** New handlers use `RegisterRoutes(public, admin)`. Provider interfaces go in `internal/provider/`. Default implementations go in `internal/service/`. Go module: `blotting-consultancy`.

---

## File Structure Overview

```
backend/
├── internal/provider/
│   ├── ai.go                              (create - AIProvider interface)
│   ├── translation.go                     (create - TranslationProvider interface)
│   └── vectorstore.go                     (create - VectorStore interface)
├── internal/service/
│   ├── ai_openai.go                       (create - OpenAI adapter)
│   ├── ai_claude.go                       (create - Claude adapter)
│   ├── ai_ollama.go                       (create - Ollama adapter)
│   ├── ai_service.go                      (create - AI orchestration service)
│   ├── ai_service_test.go                 (create)
│   ├── translation_ai.go                  (create - AI-based TranslationProvider)
│   ├── translation_service.go             (create - translation orchestration)
│   ├── translation_service_test.go        (create)
│   ├── vectorstore_sqlite.go             (create - SQLite vec impl)
│   ├── vectorstore_pgvector.go           (create - pgvector impl)
│   ├── rag_service.go                     (create - RAG pipeline)
│   ├── rag_service_test.go                (create)
│   └── wizard_service.go                  (create - AI site wizard)
├── internal/handler/
│   ├── ai/
│   │   ├── handler.go                     (create - AI content endpoints)
│   │   └── handler_test.go                (create)
│   ├── translation/
│   │   ├── handler.go                     (create - translation endpoints)
│   │   └── handler_test.go                (create)
│   ├── qa/
│   │   ├── handler.go                     (create - Q&A endpoints)
│   │   └── handler_test.go                (create)
│   └── wizard/
│       ├── handler.go                     (create - wizard endpoints)
│       └── handler_test.go                (create)
├── internal/model/
│   ├── ai_config.go                       (create - AI provider config model)
│   ├── glossary.go                        (create - translation glossary model)
│   ├── qa_log.go                          (create - Q&A log model)
│   └── wizard.go                          (create - wizard questionnaire model)
├── internal/repository/
│   ├── glossary_repository.go             (create)
│   ├── glossary_repository_impl.go        (create)
│   ├── qa_log_repository.go               (create)
│   └── qa_log_repository_impl.go          (create)
├── cmd/server/main.go                     (modify - wire AI providers)
├── go.mod                                 (modify - add AI SDK deps)
frontend/src/
├── api/
│   ├── ai.ts                              (create - AI API client)
│   ├── translation.ts                     (create - translation API client)
│   └── qa.ts                              (create - Q&A API client)
├── components/feature/
│   ├── AIPanel/
│   │   ├── AIPanel.tsx                    (create - editor AI sidebar)
│   │   ├── AISuggestions.tsx              (create - title/tag suggestions)
│   │   └── AIPanel.test.tsx               (create)
│   ├── TranslationPanel/
│   │   ├── TranslationPanel.tsx           (create - translate UI)
│   │   ├── TranslationDiff.tsx            (create - diff view)
│   │   ├── GlossaryManager.tsx            (create)
│   │   └── TranslationPanel.test.tsx      (create)
│   ├── QAWidget/
│   │   ├── QAWidget.tsx                   (create - floating Q&A)
│   │   ├── QAWidget.test.tsx              (create)
│   │   └── QABubble.tsx                   (create - chat bubble)
│   └── SiteWizard/
│       ├── WizardFlow.tsx                 (create - questionnaire)
│       ├── WizardPreview.tsx              (create - plan preview)
│       └── WizardFlow.test.tsx            (create)
├── pages/admin/
│   ├── ai-settings/page.tsx               (create - AI config page)
│   ├── glossary/page.tsx                  (create - glossary mgmt)
│   ├── qa-logs/page.tsx                   (create - Q&A logs)
│   └── wizard/page.tsx                    (create - wizard entry)
├── i18n/local/zh/ai.ts                    (create)
├── i18n/local/en/ai.ts                    (create)
```

---

## Chunk 1: AI Provider Foundation (Tasks 5.1.1 — 5.1.2)

### Task 1: Define AIProvider interface (5.1.1)

**Files:**
- Create: `backend/internal/provider/ai.go`
- Create: `backend/internal/model/ai_config.go`

- [ ] **Step 1: Create AIProvider interface**

```go
// backend/internal/provider/ai.go
package provider

import "context"

// AIMessage represents a chat message.
type AIMessage struct {
    Role    string `json:"role"`    // "system", "user", "assistant"
    Content string `json:"content"`
}

// AIRequest holds parameters for an AI completion.
type AIRequest struct {
    Messages    []AIMessage `json:"messages"`
    MaxTokens   int         `json:"maxTokens,omitempty"`
    Temperature float64     `json:"temperature,omitempty"`
    Stream      bool        `json:"stream,omitempty"`
}

// AIResponse holds the AI completion result.
type AIResponse struct {
    Content      string `json:"content"`
    FinishReason string `json:"finishReason"`
    TokensUsed   int    `json:"tokensUsed"`
}

// AIStreamChunk is a single chunk in a streaming response.
type AIStreamChunk struct {
    Content string `json:"content"`
    Done    bool   `json:"done"`
    Error   error  `json:"-"`
}

// AIProvider defines the contract for LLM backends.
type AIProvider interface {
    // Complete sends a chat completion request.
    Complete(ctx context.Context, req AIRequest) (*AIResponse, error)

    // CompleteStream sends a streaming completion request.
    CompleteStream(ctx context.Context, req AIRequest) (<-chan AIStreamChunk, error)

    // Embed generates vector embeddings for the given text.
    Embed(ctx context.Context, texts []string) ([][]float32, error)

    // Name returns the provider identifier.
    Name() string
}
```

- [ ] **Step 2: Create AI config model**

```go
// backend/internal/model/ai_config.go
package model

import "time"

// AIProviderConfig stores per-provider API configuration.
type AIProviderConfig struct {
    ID        uint      `gorm:"primaryKey" json:"id"`
    Provider  string    `gorm:"uniqueIndex;not null;size:50" json:"provider"` // "openai", "claude", "ollama"
    APIKey    string    `gorm:"size:500" json:"-"`                            // encrypted at rest
    BaseURL   string    `gorm:"size:500" json:"baseUrl"`
    Model     string    `gorm:"size:100" json:"model"`
    Enabled   bool      `gorm:"default:false" json:"enabled"`
    CreatedAt time.Time `gorm:"autoCreateTime" json:"createdAt"`
    UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updatedAt"`
}
```

- [ ] **Step 3: Add goose migration for ai_provider_configs table**

Create `backend/migrations/XXXXXX_create_ai_provider_configs.sql` with up/down.

### Task 2: Implement OpenAI and Claude adapters (5.1.2)

**Files:**
- Create: `backend/internal/service/ai_openai.go`
- Create: `backend/internal/service/ai_claude.go`
- Create: `backend/internal/service/ai_ollama.go`
- Modify: `backend/go.mod`

- [ ] **Step 1: Add dependencies**

```bash
cd /home/dev/impress/backend
go get github.com/sashabaranov/go-openai
go get github.com/anthropics/anthropic-sdk-go
```

- [ ] **Step 2: Implement OpenAI adapter**

`ai_openai.go` — struct `OpenAIProvider` implementing `provider.AIProvider`. Constructor takes `APIKey`, `Model`, `BaseURL`. Uses `go-openai` SDK for `Complete`, `CompleteStream` (SSE), and `Embed` (text-embedding-3-small).

- [ ] **Step 3: Implement Claude adapter**

`ai_claude.go` — struct `ClaudeProvider` implementing `provider.AIProvider`. Constructor takes `APIKey`, `Model`. Uses `anthropic-sdk-go` for `Complete` and `CompleteStream`. For `Embed`, delegates to a configured embedding provider or returns an error (Claude does not offer embeddings natively).

- [ ] **Step 4: Implement Ollama adapter**

`ai_ollama.go` — struct `OllamaProvider` implementing `provider.AIProvider`. Uses HTTP calls to Ollama's REST API (`/api/chat`, `/api/embeddings`). No API key required. `BaseURL` defaults to `http://localhost:11434`.

- [ ] **Step 5: Create AI orchestration service**

`ai_service.go` — `AIService` struct wrapping the active `AIProvider` (fetched from Registry). Provides high-level methods: `Summarize(ctx, text, locale)`, `SuggestTitles(ctx, body, locale, count)`, `RecommendTags(ctx, body, existingTags)`, `Rewrite(ctx, text, instruction)`. Each composes a system prompt + user content and calls `Complete`.

- [ ] **Step 6: Register AI provider on startup**

In `cmd/server/main.go`, read `AIProviderConfig` from DB. If enabled, instantiate the matching adapter and register as `registry.Register("ai", adapter)`.

- [ ] **Step 7: Write unit tests for AIService**

`ai_service_test.go` — mock `AIProvider`, test that `Summarize`/`SuggestTitles`/`RecommendTags` build correct prompts and parse responses. Test fallback when no provider is configured.

---

## Chunk 2: Editor AI Assistant (Tasks 5.1.3 — 5.1.6)

### Task 3: AI content handler (5.1.3 backend)

**Files:**
- Create: `backend/internal/handler/ai/handler.go`
- Modify: `backend/cmd/server/main.go` (register routes)

- [ ] **Step 1: Create AI handler with endpoints**

```
POST /api/v1/ai/complete        — general completion (rewrite/expand/shorten/tone)
POST /api/v1/ai/complete/stream — SSE streaming completion
POST /api/v1/ai/summarize       — generate summary for article body
POST /api/v1/ai/suggest-titles  — return N title candidates
POST /api/v1/ai/suggest-tags    — return tag recommendations
```

All admin-only (JWT required). Request bodies include `content`, `locale`, `instruction`. Handler delegates to `AIService`.

- [ ] **Step 2: Implement SSE streaming endpoint**

`/ai/complete/stream` sets `Content-Type: text/event-stream`, reads from the `AIStreamChunk` channel, and flushes each chunk as an SSE `data:` line. Closes on `Done` or context cancellation.

- [ ] **Step 3: Register AI routes in main.go**

Wire `aiHandler.RegisterRoutes(publicGroup, adminGroup)`.

### Task 4: Editor AI panel (5.1.3 frontend)

**Files:**
- Create: `frontend/src/api/ai.ts`
- Create: `frontend/src/components/feature/AIPanel/AIPanel.tsx`
- Create: `frontend/src/components/feature/AIPanel/AISuggestions.tsx`
- Modify: article editor page to include AIPanel

- [ ] **Step 1: Create AI API client**

```typescript
// frontend/src/api/ai.ts
import http from "./http";

export interface AICompleteReq {
  content: string;
  instruction: string;
  locale: string;
}

export const aiComplete = (data: AICompleteReq) => http.post("/ai/complete", data);
export const aiSummarize = (content: string, locale: string) =>
  http.post("/ai/summarize", { content, locale });
export const aiSuggestTitles = (content: string, locale: string, count?: number) =>
  http.post("/ai/suggest-titles", { content, locale, count: count ?? 5 });
export const aiSuggestTags = (content: string) =>
  http.post("/ai/suggest-tags", { content });

// SSE streaming helper
export function aiCompleteStream(
  data: AICompleteReq,
  onChunk: (text: string) => void,
  onDone: () => void
) {
  // Uses fetch + ReadableStream for SSE
}
```

- [ ] **Step 2: Build AIPanel component**

Sidebar panel that slides open from the right edge of the article editor. Contains:
- Action buttons: Continue writing, Rewrite, Expand, Shorten, Change tone (formal/casual)
- Text selection awareness: operates on TipTap selection or full body
- Streaming output area showing AI response in real-time
- "Apply" button to replace selection / append to editor
- Uses `useTranslation("ai")` for bilingual labels

- [ ] **Step 3: Build AISuggestions component**

Inline suggestion cards for:
- Title suggestions (shown below title input, click to apply)
- Tag recommendations (shown as pills, click to add)
- Auto-summary (one-click fills `ZhMetaDescription`/`EnMetaDescription`)

- [ ] **Step 4: Integrate AIPanel into article editor**

Add a toggle button (sparkle icon) in the editor toolbar. When active, renders `<AIPanel>` alongside the TipTap editor. Pass editor instance ref so AIPanel can read selection and insert text.

- [ ] **Step 5: Add i18n keys**

Add AI-related keys to `frontend/src/i18n/local/zh/ai.ts` and `en/ai.ts`: panel title, button labels, loading states, error messages.

- [ ] **Step 6: Write AIPanel render test**

`AIPanel.test.tsx` — renders without crash, shows action buttons, mock API call.

### Task 5: Auto-summary on save (5.1.4)

- [ ] **Step 1: Backend — add optional auto-summary trigger**

In `AIService.AutoSummarize(ctx, articleID)`: load article, generate summary for `ZhBody` (locale=zh) and `EnBody` (locale=en) if non-empty, update `ZhMetaDescription`/`EnMetaDescription`. Only runs if `Article.AutoSummary == true`.

- [ ] **Step 2: Hook into article save**

In article handler's create/update flow, after successful save, if `AutoSummary` is true and an AI provider is configured, call `AIService.AutoSummarize` asynchronously (goroutine). Log errors, do not block the save response.

- [ ] **Step 3: Frontend — auto-summary toggle**

Add a checkbox "Auto-generate summary" in the article editor SEO section. Maps to `Article.AutoSummary` field.

### Task 6: Title suggestions and tag recommendations (5.1.5, 5.1.6)

- [ ] **Step 1: Title suggestion endpoint already defined in Task 3**

Verify `POST /api/v1/ai/suggest-titles` returns `{ titles: string[] }`.

- [ ] **Step 2: Tag recommendation endpoint already defined in Task 3**

Verify `POST /api/v1/ai/suggest-tags` returns `{ tags: [{ name: string, confidence: float }] }`. Cross-references existing tags in the database to prefer reuse over creating new ones.

- [ ] **Step 3: Frontend integration in AISuggestions (already built in Task 4)**

Verify title suggestion cards and tag pills are functional.

---

## Chunk 3: Smart Translation (Tasks 5.2.1 — 5.2.5)

### Task 7: TranslationProvider interface (5.2.1)

**Files:**
- Create: `backend/internal/provider/translation.go`

- [ ] **Step 1: Define TranslationProvider interface**

```go
// backend/internal/provider/translation.go
package provider

import "context"

// TranslationRequest holds translation parameters.
type TranslationRequest struct {
    Text       string            `json:"text"`
    SourceLang string            `json:"sourceLang"` // "zh", "en"
    TargetLang string            `json:"targetLang"`
    Glossary   map[string]string `json:"glossary,omitempty"` // term -> translation
    Format     string            `json:"format,omitempty"`   // "plain", "html"
}

// TranslationResponse holds translation output.
type TranslationResponse struct {
    Text       string `json:"text"`
    TokensUsed int    `json:"tokensUsed"`
}

// TranslationProvider defines the contract for translation backends.
type TranslationProvider interface {
    Translate(ctx context.Context, req TranslationRequest) (*TranslationResponse, error)
    TranslateBatch(ctx context.Context, reqs []TranslationRequest) ([]TranslationResponse, error)
    Name() string
}
```

### Task 8: AI-based translation implementation (5.2.1 impl)

**Files:**
- Create: `backend/internal/service/translation_ai.go`
- Create: `backend/internal/service/translation_service.go`
- Create: `backend/internal/service/translation_service_test.go`

- [ ] **Step 1: Implement AITranslationProvider**

`translation_ai.go` — wraps `AIProvider` to implement `TranslationProvider`. Builds a system prompt with glossary terms injected as "always translate X as Y" rules. Handles HTML content by instructing the LLM to preserve tags.

- [ ] **Step 2: Create TranslationService**

`translation_service.go` — orchestrates translation for articles/pages:
- `TranslateArticle(ctx, articleID, sourceLang, targetLang)` — translates title, body, meta description
- `TranslateArticleBatch(ctx, articleIDs, sourceLang, targetLang)` — batch mode
- `TranslatePage(ctx, pageID, sourceLang, targetLang)` — translates page section content
- Loads glossary from DB before each translation call

- [ ] **Step 3: Write tests**

Mock AIProvider, verify prompt construction includes glossary terms, verify field mapping (ZhTitle -> EnTitle etc).

### Task 9: Translation handler and endpoints (5.2.2, 5.2.4)

**Files:**
- Create: `backend/internal/handler/translation/handler.go`

- [ ] **Step 1: Create translation endpoints**

```
POST /api/v1/translation/article/:id      — translate single article
POST /api/v1/translation/article/batch     — batch translate articles { ids: [] }
POST /api/v1/translation/page/:id          — translate single page
POST /api/v1/translation/preview           — preview translation without saving
```

All admin-only. Request body specifies `sourceLang` and `targetLang`. Returns translated fields.

- [ ] **Step 2: Register routes in main.go**

### Task 10: Glossary management (5.2.5)

**Files:**
- Create: `backend/internal/model/glossary.go`
- Create: `backend/internal/repository/glossary_repository.go`
- Create: `backend/internal/repository/glossary_repository_impl.go`
- Create: `frontend/src/pages/admin/glossary/page.tsx`

- [ ] **Step 1: Create Glossary model**

```go
type GlossaryEntry struct {
    ID        uint      `gorm:"primaryKey" json:"id"`
    ZhTerm    string    `gorm:"not null;size:200;uniqueIndex:idx_glossary_pair" json:"zhTerm"`
    EnTerm    string    `gorm:"not null;size:200;uniqueIndex:idx_glossary_pair" json:"enTerm"`
    Category  string    `gorm:"size:100;index" json:"category"` // e.g. "brand", "technical"
    Note      string    `gorm:"size:500" json:"note"`
    CreatedAt time.Time `gorm:"autoCreateTime" json:"createdAt"`
    UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updatedAt"`
}
```

- [ ] **Step 2: Create glossary repository (interface + impl)**

Standard CRUD: `List(category)`, `Create`, `Update`, `Delete`, `GetAll()` (for translation injection).

- [ ] **Step 3: Add CRUD endpoints to translation handler**

```
GET    /api/v1/translation/glossary
POST   /api/v1/translation/glossary
PUT    /api/v1/translation/glossary/:id
DELETE /api/v1/translation/glossary/:id
```

- [ ] **Step 4: Create glossary admin page**

Table view with inline edit. Columns: Chinese term, English term, category, note. Supports search/filter by category. Import/export CSV.

- [ ] **Step 5: Add migration for glossary_entries table**

### Task 11: Translation diff comparison (5.2.3)

**Files:**
- Create: `frontend/src/components/feature/TranslationPanel/TranslationDiff.tsx`
- Create: `frontend/src/components/feature/TranslationPanel/TranslationPanel.tsx`

- [ ] **Step 1: Backend — translation freshness check**

Add `ZhBodyHash` and `EnBodyHash` fields (sha256 of body at time of last translation) to Article model. When translating, store hash. `GET /api/v1/translation/article/:id/status` returns `{ stale: bool, changedSections: [] }`.

- [ ] **Step 2: Frontend — TranslationPanel**

Side-by-side view in article editor. Left: source language content. Right: target language content. "Translate" button for one-click fill. Shows "stale" badge when source has changed since last translation.

- [ ] **Step 3: Frontend — TranslationDiff**

When source changed, highlight paragraphs that differ from the last-translated version using a simple diff algorithm (e.g., `diff-match-patch`). Show which paragraphs need re-translation.

- [ ] **Step 4: Write render tests**

---

## Chunk 4: Knowledge Base Q&A (Tasks 5.3.1 — 5.3.5)

### Task 12: VectorStore interface (5.3.2)

**Files:**
- Create: `backend/internal/provider/vectorstore.go`

- [ ] **Step 1: Define VectorStore interface**

```go
// backend/internal/provider/vectorstore.go
package provider

import "context"

// VectorDocument represents a document stored with its embedding.
type VectorDocument struct {
    ID         string    `json:"id"`         // "article:42", "page:7"
    Content    string    `json:"content"`    // chunk text
    Embedding  []float32 `json:"-"`
    ContentType string   `json:"contentType"` // "article", "page"
    ContentID  uint      `json:"contentId"`
    Locale     string    `json:"locale"`
    Metadata   map[string]string `json:"metadata,omitempty"`
}

// VectorSearchResult is a document with similarity score.
type VectorSearchResult struct {
    Document   VectorDocument `json:"document"`
    Score      float64        `json:"score"`
}

// VectorStore defines the contract for vector storage backends.
type VectorStore interface {
    Upsert(ctx context.Context, docs []VectorDocument) error
    Search(ctx context.Context, embedding []float32, limit int, filter map[string]string) ([]VectorSearchResult, error)
    Delete(ctx context.Context, ids []string) error
    Count(ctx context.Context) (int64, error)
}
```

### Task 13: VectorStore implementations (5.3.2 impl)

**Files:**
- Create: `backend/internal/service/vectorstore_sqlite.go`
- Create: `backend/internal/service/vectorstore_pgvector.go`

- [ ] **Step 1: SQLite vec implementation**

Uses `sqlite-vec` extension. Table `vector_documents(id TEXT PK, content TEXT, embedding BLOB, content_type TEXT, content_id INT, locale TEXT, metadata JSON)`. Search uses `vec_distance_cosine`. Requires `sqlite-vec` extension loaded at DB init.

- [ ] **Step 2: pgvector implementation**

Uses `pgvector` extension. Table with `vector` column type. Search uses `<=>` cosine distance operator. Create migration with `CREATE EXTENSION IF NOT EXISTS vector`.

- [ ] **Step 3: Add migrations for vector tables**

Both SQLite and PostgreSQL migration files.

- [ ] **Step 4: Register VectorStore in startup**

Detect DB type, instantiate matching VectorStore, register as `registry.Register("vectorstore", store)`.

### Task 14: Content vectorization (5.3.1)

**Files:**
- Modify: `backend/internal/service/rag_service.go`

- [ ] **Step 1: Create RAGService with content indexing**

`RAGService` struct with deps: `AIProvider` (for embeddings), `VectorStore`, article/page repositories.

Method `IndexArticle(ctx, articleID)`:
1. Load article from DB
2. Chunk body text into ~500-token segments with overlap
3. Call `AIProvider.Embed()` for each chunk
4. Upsert into VectorStore with metadata (article ID, locale, title)

Method `IndexPage(ctx, pageID)`: same pattern for page content.

Method `ReindexAll(ctx)`: iterate all published articles/pages.

- [ ] **Step 2: Hook into EventBus**

Subscribe to `ContentPublished` event. On publish, call `IndexArticle`/`IndexPage` asynchronously. On `ContentDeleted`, call `VectorStore.Delete`.

- [ ] **Step 3: Text chunking utility**

`internal/service/chunker.go` — split text into chunks by paragraph boundaries, respecting max token limit. Overlap last 50 tokens of previous chunk for context continuity.

### Task 15: RAG pipeline (5.3.3)

**Files:**
- Create: `backend/internal/service/rag_service.go` (extend from Task 14)

- [ ] **Step 1: Implement RAG query method**

`RAGService.Answer(ctx, question, locale)`:
1. Embed the question via `AIProvider.Embed()`
2. Search VectorStore for top-K (default 5) relevant chunks, filtered by locale
3. Assemble context: concatenate chunk contents with source attribution
4. Build prompt: system instruction (answer based on provided context only, cite sources) + context + user question
5. Call `AIProvider.Complete()`
6. Return answer with source references

- [ ] **Step 2: Implement streaming RAG**

`RAGService.AnswerStream(ctx, question, locale)` — same retrieval, but uses `CompleteStream` for response. Returns channel.

- [ ] **Step 3: Write RAG tests**

Mock AIProvider and VectorStore. Verify context assembly, source attribution, and graceful handling of empty results.

### Task 16: Q&A handler (5.3.3 — 5.3.5 backend)

**Files:**
- Create: `backend/internal/handler/qa/handler.go`
- Create: `backend/internal/model/qa_log.go`
- Create: `backend/internal/repository/qa_log_repository.go`
- Create: `backend/internal/repository/qa_log_repository_impl.go`

- [ ] **Step 1: Create Q&A endpoints**

```
POST   /public/qa/ask            — public: ask a question (no auth)
POST   /public/qa/ask/stream     — public: streaming answer (SSE)
POST   /api/v1/qa/feedback       — public: mark answer helpful/unhelpful
GET    /api/v1/qa/logs            — admin: list Q&A logs
POST   /api/v1/qa/reindex        — admin: trigger full reindex
```

Public endpoints have rate limiting (e.g., 10 req/min per IP).

- [ ] **Step 2: QALog model**

```go
type QALog struct {
    ID        uint      `gorm:"primaryKey" json:"id"`
    Question  string    `gorm:"type:text;not null" json:"question"`
    Answer    string    `gorm:"type:text" json:"answer"`
    Sources   JSONMap   `gorm:"type:jsonb" json:"sources"`  // [{type, id, title}]
    Locale    string    `gorm:"size:10" json:"locale"`
    Helpful   *bool     `json:"helpful"`                     // null = no feedback
    IP        string    `gorm:"size:45" json:"-"`
    CreatedAt time.Time `gorm:"autoCreateTime" json:"createdAt"`
}
```

- [ ] **Step 3: Create repository (interface + impl) and migration**

- [ ] **Step 4: Admin Q&A logs page**

`frontend/src/pages/admin/qa-logs/page.tsx` — table of recent Q&A pairs with feedback status, filter by date/locale, view sources.

### Task 17: Frontend Q&A widget (5.3.4)

**Files:**
- Create: `frontend/src/components/feature/QAWidget/QAWidget.tsx`
- Create: `frontend/src/components/feature/QAWidget/QABubble.tsx`
- Create: `frontend/src/api/qa.ts`

- [ ] **Step 1: Create Q&A API client**

```typescript
export const askQuestion = (question: string, locale: string) =>
  http.post("/public/qa/ask", { question, locale });

export function askQuestionStream(
  question: string, locale: string,
  onChunk: (text: string) => void,
  onDone: (sources: any[]) => void
) { /* SSE via fetch */ }

export const submitFeedback = (logId: number, helpful: boolean) =>
  http.post("/qa/feedback", { logId, helpful });
```

- [ ] **Step 2: Build QABubble component**

Fixed-position floating button (bottom-right corner). Click to expand chat panel. Uses `useTranslation("ai")` for labels. Shows on public pages only (not admin).

- [ ] **Step 3: Build QAWidget component**

Chat-style interface:
- Input field at bottom, send button
- Messages list: user questions + AI answers with streaming render
- Source citations as clickable links below each answer
- Thumbs up/down feedback buttons on each answer
- Locale-aware: uses current i18next language

- [ ] **Step 4: Integrate into App layout**

Render `<QABubble />` in the public layout (not admin). Conditionally shown based on GlobalConfig flag `enableQA`.

- [ ] **Step 5: Write render tests**

---

## Chunk 5: AI Site Wizard (Tasks 5.4.1 — 5.4.5)

### Task 18: Wizard questionnaire and plan generation (5.4.1 — 5.4.2)

**Files:**
- Create: `backend/internal/handler/wizard/handler.go`
- Create: `backend/internal/service/wizard_service.go`
- Create: `backend/internal/model/wizard.go`

- [ ] **Step 1: Define wizard questionnaire model**

```go
type WizardInput struct {
    Industry    string   `json:"industry"`    // e.g. "technology", "restaurant", "consulting"
    SiteName    string   `json:"siteName"`
    Description string   `json:"description"` // brief business description
    Style       string   `json:"style"`       // "modern", "warm", "minimal", "bold"
    Features    []string `json:"features"`    // ["blog", "portfolio", "contact", "products"]
    Locale      string   `json:"locale"`      // primary locale
    BrandColor  string   `json:"brandColor,omitempty"` // optional hex color
}

type WizardPlan struct {
    Theme       string              `json:"theme"`       // recommended theme package
    ColorScheme map[string]string   `json:"colorScheme"` // primary, secondary, accent, bg
    Pages       []WizardPagePlan    `json:"pages"`
    Navigation  []WizardNavItem     `json:"navigation"`
    LogoPrompt  string              `json:"logoPrompt"`  // prompt for logo generation
}

type WizardPagePlan struct {
    Title    string   `json:"title"`
    Slug     string   `json:"slug"`
    Sections []string `json:"sections"` // section type names from theme system
}

type WizardNavItem struct {
    Label string `json:"label"`
    Path  string `json:"path"`
}
```

- [ ] **Step 2: Implement WizardService**

`WizardService.GeneratePlan(ctx, input WizardInput) (*WizardPlan, error)`:
- Builds a prompt describing available themes (`default`, `modern-dark`, `warm-earth`), available section types (`HeroSection`, `CardGridSection`, `ContactFormSection`, etc.), and asks AI to produce a JSON plan matching the user's industry/style
- Parses structured JSON response into `WizardPlan`

`WizardService.Scaffold(ctx, plan WizardPlan, input WizardInput) error`:
- Creates pages using page API/service
- Sets theme and color tokens
- Generates navigation menu entries

- [ ] **Step 3: Create wizard endpoints**

```
POST /api/v1/wizard/plan        — generate site plan from questionnaire
POST /api/v1/wizard/scaffold    — execute plan: create pages, apply theme
```

Admin-only.

### Task 19: AI content filling and color suggestions (5.4.3 — 5.4.5)

**Files:**
- Modify: `backend/internal/service/wizard_service.go`

- [ ] **Step 1: AI content generation for scaffold**

Extend `WizardService.Scaffold`:
- For each page in the plan, call `AIProvider.Complete()` with a prompt to generate bilingual page content (hero text, card descriptions, about-us copy) based on industry and description
- Fill both `zh` and `en` content fields
- Use glossary if available for consistent terminology

- [ ] **Step 2: Color scheme suggestion**

`WizardService.SuggestColors(ctx, industry, style, brandColor)`:
- If `brandColor` provided, generate complementary palette
- Otherwise, ask AI for industry-appropriate color palette
- Return primary, secondary, accent, background, text colors as hex values

- [ ] **Step 3: Logo prompt generation**

Generate a text prompt suitable for image-generation tools (DALL-E, Midjourney) based on site name, industry, and style. Return as `logoPrompt` string — actual image generation is out of scope (users use external tools with the prompt).

### Task 20: Frontend wizard flow (5.4.1 — 5.4.5 frontend)

**Files:**
- Create: `frontend/src/pages/admin/wizard/page.tsx`
- Create: `frontend/src/components/feature/SiteWizard/WizardFlow.tsx`
- Create: `frontend/src/components/feature/SiteWizard/WizardPreview.tsx`

- [ ] **Step 1: Build WizardFlow component**

Multi-step form:
1. **Industry & Name** — text inputs + industry dropdown
2. **Style & Features** — visual style selector (cards with previews) + feature checkboxes
3. **Brand Color** — optional color picker
4. **Review** — show generated plan (calls `POST /wizard/plan`)
5. **Confirm** — execute scaffold (calls `POST /wizard/scaffold`)

Progress indicator at top. Back/Next navigation.

- [ ] **Step 2: Build WizardPreview component**

Shows the generated plan: site map tree, recommended theme name, color swatches, page list with section types. Editable before confirming.

- [ ] **Step 3: Create wizard admin page**

`pages/admin/wizard/page.tsx` — entry point. Accessible from admin sidebar. Shows WizardFlow. After completion, redirects to admin dashboard.

- [ ] **Step 4: Add route and navigation**

Add `/admin/wizard` route to `src/router/config.tsx`. Add "Site Wizard" link to admin sidebar navigation.

- [ ] **Step 5: Write render tests**

---

## Chunk 6: Admin Settings & Wiring (Cross-cutting)

### Task 21: AI settings admin page

**Files:**
- Create: `frontend/src/pages/admin/ai-settings/page.tsx`

- [ ] **Step 1: Create AI provider configuration page**

Form with:
- Provider selector (OpenAI / Claude / Ollama)
- API key input (masked)
- Base URL (for Ollama or custom endpoints)
- Model name input
- Enable/disable toggle
- Test connection button (calls a lightweight completion to verify)

- [ ] **Step 2: Backend settings endpoints**

Add to AI handler:
```
GET    /api/v1/ai/config           — get current config (API key masked)
PUT    /api/v1/ai/config           — update config
POST   /api/v1/ai/config/test      — test connection
```

- [ ] **Step 3: Add admin sidebar navigation**

Add "AI Settings", "Glossary", "Q&A Logs" links under a new "AI" section in admin sidebar.

### Task 22: GlobalConfig integration

- [ ] **Step 1: Add AI feature flags to GlobalConfig**

Add fields: `enableAI` (bool), `enableQA` (bool), `enableTranslation` (bool), `enableWizard` (bool). These control frontend visibility of AI features.

- [ ] **Step 2: Conditional rendering**

Frontend checks `GlobalConfig.enableQA` before rendering QABubble. Editor AIPanel checks `GlobalConfig.enableAI`. TranslationPanel checks `GlobalConfig.enableTranslation`.

---

## Verification

After each chunk, run:

```bash
pnpm lint && pnpm type-check
cd backend && go vet ./... && go test -v -race ./...
```

## Dependency Graph

```
Chunk 1 (Provider + Adapters)
    ↓
Chunk 2 (Editor AI) ← depends on Chunk 1
    ↓
Chunk 3 (Translation) ← depends on Chunk 1
    ↓
Chunk 4 (Knowledge Q&A) ← depends on Chunk 1 (embeddings)
    ↓
Chunk 5 (Wizard) ← depends on Chunks 1, 3 (content gen + translation)
    ↓
Chunk 6 (Settings) ← depends on all above
```

Chunks 2, 3, and 4 can be developed in parallel after Chunk 1 is complete.

## Risk Mitigation

| Risk | Mitigation |
|------|-----------|
| AI provider API costs during dev | Use Ollama with local models for development; mock providers in tests |
| Embedding model mismatch across providers | Standardize on OpenAI `text-embedding-3-small` dimensions (1536); Ollama models may differ — store dimension in config |
| Vector extension availability | `sqlite-vec` and `pgvector` are optional; VectorStore returns clear error if not available; Q&A feature degrades gracefully |
| Streaming SSE through reverse proxy | Document nginx config: `proxy_buffering off; X-Accel-Buffering: no` |
| Large article chunking quality | Use paragraph-boundary splitting with configurable chunk size; allow admin to manually adjust |
| Glossary consistency | Inject glossary into every translation prompt; warn when glossary grows beyond prompt token budget |
