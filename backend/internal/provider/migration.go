package provider

import (
	"context"
	"io"
	"time"

	"github.com/yixian-huang/inkless/backend/internal/model"
)

// MigrationSource identifies the platform being migrated from.
type MigrationSource string

const (
	SourceWordPress MigrationSource = "wordpress"
	SourceHalo      MigrationSource = "halo"
	SourceMarkdown  MigrationSource = "markdown"
)

// MigrationArticle is the intermediate representation produced by a
// MigrationProvider. It carries everything needed to create an Article,
// its Category, its Tags, and any associated media URLs.
type MigrationArticle struct {
	// Core fields
	Slug   string
	Title  string
	Body   string
	Status model.ArticleStatus

	// Optional English translations (bilingual site)
	EnTitle string
	EnBody  string

	// Taxonomy
	CategoryName string   // will be resolved or created
	TagNames     []string // will be resolved or created

	// Media
	CoverImageURL string
	MediaURLs     []string // embedded media URLs to download

	// Metadata
	PublishedAt *time.Time
	CreatedAt   *time.Time

	// SEO
	SeoTitle        string
	MetaDescription string
}

// MigrationResult holds the outcome of parsing a single source.
type MigrationResult struct {
	Articles []*MigrationArticle
	Errors   []string // non-fatal parse warnings
}

// MigrationProgress reports ongoing import status.
type MigrationProgress struct {
	JobID      string          `json:"jobId"`
	Source     MigrationSource `json:"source"`
	Phase      string          `json:"phase"` // "parsing", "importing", "done", "failed"
	Total      int             `json:"total"`
	Processed  int             `json:"processed"`
	Succeeded  int             `json:"succeeded"`
	Failed     int             `json:"failed"`
	Errors     []string        `json:"errors,omitempty"`
	Attempt    int             `json:"attempt"`
	Retryable  bool            `json:"retryable"`
	StartedAt  time.Time       `json:"startedAt"`
	FinishedAt *time.Time      `json:"finishedAt,omitempty"`
}

// MigrationProvider is the interface that each source-specific importer
// must implement. It is responsible only for *parsing* the source data
// into MigrationArticle slices — the handler/service layer handles
// writing to the database.
type MigrationProvider interface {
	// Source returns which platform this provider handles.
	Source() MigrationSource

	// Parse reads source data from r and returns intermediate articles.
	// The context should be checked for cancellation on long-running parses.
	Parse(ctx context.Context, r io.Reader) (*MigrationResult, error)
}
