package service

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"unicode"

	"gorm.io/gorm"

	"github.com/yixian-huang/inkless/backend/internal/provider"
)

// SearchService implements provider.SearchProvider using SQLite FTS5
// (with a PostgreSQL path for future use).
type SearchService struct {
	db         *gorm.DB
	isPostgres bool
}

// NewSearchService creates a new SearchService.
// Set isPostgres to true when using PostgreSQL (tsvector); false for SQLite FTS5.
func NewSearchService(db *gorm.DB, isPostgres bool) *SearchService {
	return &SearchService{db: db, isPostgres: isPostgres}
}

// containsCJK returns true if the string contains any CJK characters.
func containsCJK(s string) bool {
	for _, r := range s {
		if unicode.Is(unicode.Han, r) ||
			unicode.Is(unicode.Hangul, r) ||
			unicode.Is(unicode.Katakana, r) ||
			unicode.Is(unicode.Hiragana, r) {
			return true
		}
	}
	return false
}

// truncateSnippet returns a rune-safe preview for search results.
func truncateSnippet(s string, maxRunes int) string {
	if maxRunes <= 0 || s == "" {
		return s
	}
	runes := []rune(s)
	if len(runes) <= maxRunes {
		return s
	}
	return string(runes[:maxRunes]) + "…"
}

// Search performs a full-text search across the index.
func (s *SearchService) Search(ctx context.Context, query, locale, contentType string, page, pageSize int) (*provider.SearchResponse, error) {
	if s.isPostgres {
		return s.searchPostgres(ctx, query, locale, contentType, page, pageSize)
	}
	// Use LIKE-based search for CJK queries since FTS5 unicode61 tokenizer
	// does not handle CJK characters well (each character becomes a separate
	// token and multi-character matches fail).
	if containsCJK(query) {
		return s.searchSQLiteLike(ctx, query, locale, contentType, page, pageSize)
	}
	return s.searchSQLiteFTS(ctx, query, locale, contentType, page, pageSize)
}

// escapeLike escapes SQL LIKE special characters.
func escapeLike(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "%", "\\%")
	s = strings.ReplaceAll(s, "_", "\\_")
	return s
}

// sanitizeFTS5Query makes user input safe for FTS5 MATCH by wrapping it in
// double quotes (phrase query) after escaping embedded double quotes.
func sanitizeFTS5Query(q string) string {
	q = strings.TrimSpace(q)
	if q == "" {
		return q
	}
	// Remove FTS5 special operators and characters
	q = strings.ReplaceAll(q, "\"", "")
	// Strip column filter syntax (e.g. "title:")
	q = regexp.MustCompile(`\w+\s*:`).ReplaceAllString(q, "")
	q = strings.TrimSpace(q)
	if q == "" {
		return q
	}
	// Wrap in double quotes to force phrase matching
	return "\"" + q + "\""
}

// searchSQLiteLike performs a LIKE-based search for CJK content.
func (s *SearchService) searchSQLiteLike(ctx context.Context, query, locale, contentType string, page, pageSize int) (*provider.SearchResponse, error) {
	offset := (page - 1) * pageSize
	likePattern := "%" + escapeLike(query) + "%"

	conditions := []string{"(title LIKE ? ESCAPE '\\' OR body LIKE ? ESCAPE '\\')"}
	args := []interface{}{likePattern, likePattern}

	if locale != "" {
		conditions = append(conditions, "locale = ?")
		args = append(args, locale)
	}
	if contentType != "" {
		conditions = append(conditions, "content_type = ?")
		args = append(args, contentType)
	}

	where := strings.Join(conditions, " AND ")

	var total int64
	countSQL := fmt.Sprintf("SELECT COUNT(*) FROM search_index_fts WHERE %s", where)
	if err := s.db.WithContext(ctx).Raw(countSQL, args...).Scan(&total).Error; err != nil {
		return nil, fmt.Errorf("search count: %w", err)
	}

	// Cap body payload in SQL so CJK LIKE searches do not ship full article
	// bodies over the wire (snippet truncation is also applied in Go).
	selectSQL := fmt.Sprintf(
		"SELECT content_type, content_id, locale, title, substr(body, 1, 280) as body, slug FROM search_index_fts WHERE %s LIMIT ? OFFSET ?",
		where,
	)
	fetchArgs := make([]interface{}, len(args), len(args)+2)
	copy(fetchArgs, args)
	fetchArgs = append(fetchArgs, pageSize, offset)

	var rows []struct {
		ContentType string
		ContentID   uint
		Locale      string
		Title       string
		Body        string
		Slug        string
	}
	if err := s.db.WithContext(ctx).Raw(selectSQL, fetchArgs...).Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("search query: %w", err)
	}

	results := make([]provider.SearchResult, len(rows))
	for i, row := range rows {
		url := "/" + row.Slug
		if row.ContentType == "article" {
			url = "/blog/" + row.Slug
		}
		results[i] = provider.SearchResult{
			ID:      row.ContentID,
			Type:    row.ContentType,
			Title:   row.Title,
			Snippet: truncateSnippet(row.Body, 200),
			URL:     url,
			Locale:  row.Locale,
			Score:   1.0,
		}
	}

	return &provider.SearchResponse{
		Results:  results,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
		Query:    query,
	}, nil
}

// searchSQLiteFTS performs FTS5 MATCH-based search for non-CJK queries.
func (s *SearchService) searchSQLiteFTS(ctx context.Context, query, locale, contentType string, page, pageSize int) (*provider.SearchResponse, error) {
	offset := (page - 1) * pageSize

	sanitized := sanitizeFTS5Query(query)
	if sanitized == "" {
		return &provider.SearchResponse{Results: []provider.SearchResult{}, Total: 0, Page: page, PageSize: pageSize, Query: query}, nil
	}

	conditions := []string{"search_index_fts MATCH ?"}
	args := []interface{}{sanitized}

	if locale != "" {
		conditions = append(conditions, "locale = ?")
		args = append(args, locale)
	}
	if contentType != "" {
		conditions = append(conditions, "content_type = ?")
		args = append(args, contentType)
	}

	where := strings.Join(conditions, " AND ")

	// Count total matches
	var total int64
	countSQL := fmt.Sprintf("SELECT COUNT(*) FROM search_index_fts WHERE %s", where)
	if err := s.db.WithContext(ctx).Raw(countSQL, args...).Scan(&total).Error; err != nil {
		return nil, fmt.Errorf("search count: %w", err)
	}

	// Fetch results with snippet highlighting
	selectSQL := fmt.Sprintf(
		"SELECT content_type, content_id, locale, title, snippet(search_index_fts, 4, '<mark>', '</mark>', '...', 32) as body, slug, rank FROM search_index_fts WHERE %s ORDER BY rank LIMIT ? OFFSET ?",
		where,
	)
	// Copy args to avoid mutating the original slice
	fetchArgs := make([]interface{}, len(args), len(args)+2)
	copy(fetchArgs, args)
	fetchArgs = append(fetchArgs, pageSize, offset)

	var rows []struct {
		ContentType string
		ContentID   uint
		Locale      string
		Title       string
		Body        string
		Slug        string
		Rank        float64
	}
	if err := s.db.WithContext(ctx).Raw(selectSQL, fetchArgs...).Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("search query: %w", err)
	}

	results := make([]provider.SearchResult, len(rows))
	for i, row := range rows {
		url := "/" + row.Slug
		if row.ContentType == "article" {
			url = "/blog/" + row.Slug
		}
		results[i] = provider.SearchResult{
			ID:      row.ContentID,
			Type:    row.ContentType,
			Title:   row.Title,
			Snippet: row.Body,
			URL:     url,
			Locale:  row.Locale,
			Score:   -row.Rank, // FTS5 rank is negative; negate for a positive score
		}
	}

	return &provider.SearchResponse{
		Results:  results,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
		Query:    query,
	}, nil
}

// searchPostgres is a placeholder that delegates to the SQLite path for now.
func (s *SearchService) searchPostgres(ctx context.Context, query, locale, contentType string, page, pageSize int) (*provider.SearchResponse, error) {
	// TODO: implement PostgreSQL tsvector-based search
	if containsCJK(query) {
		return s.searchSQLiteLike(ctx, query, locale, contentType, page, pageSize)
	}
	return s.searchSQLiteFTS(ctx, query, locale, contentType, page, pageSize)
}

// Suggest returns title completions matching a prefix.
func (s *SearchService) Suggest(ctx context.Context, prefix, locale string, limit int) ([]string, error) {
	if limit == 0 {
		limit = 5
	}
	var titles []string
	if containsCJK(prefix) {
		sql := "SELECT DISTINCT title FROM search_index_fts WHERE title LIKE ? ESCAPE '\\' AND locale = ? LIMIT ?"
		if err := s.db.WithContext(ctx).Raw(sql, escapeLike(prefix)+"%", locale, limit).Scan(&titles).Error; err != nil {
			return nil, fmt.Errorf("suggest: %w", err)
		}
	} else {
		safePrefix := strings.ReplaceAll(prefix, "\"", "")
		safePrefix = strings.TrimSpace(safePrefix)
		if safePrefix == "" {
			return titles, nil
		}
		sql := "SELECT DISTINCT title FROM search_index_fts WHERE title MATCH ? AND locale = ? LIMIT ?"
		if err := s.db.WithContext(ctx).Raw(sql, "\""+safePrefix+"\"*", locale, limit).Scan(&titles).Error; err != nil {
			return nil, fmt.Errorf("suggest: %w", err)
		}
	}
	return titles, nil
}

// IndexArticle adds or replaces an article in the search index.
func (s *SearchService) IndexArticle(ctx context.Context, id uint, locale, title, body, slug string) error {
	if err := s.RemoveFromIndex(ctx, "article", id); err != nil {
		return err
	}
	sql := "INSERT INTO search_index_fts(content_type, content_id, locale, title, body, slug) VALUES(?, ?, ?, ?, ?, ?)"
	return s.db.WithContext(ctx).Exec(sql, "article", id, locale, title, body, slug).Error
}

// IndexPage adds or replaces a page in the search index.
func (s *SearchService) IndexPage(ctx context.Context, id uint, locale, title, body, slug string) error {
	if err := s.RemoveFromIndex(ctx, "page", id); err != nil {
		return err
	}
	sql := "INSERT INTO search_index_fts(content_type, content_id, locale, title, body, slug) VALUES(?, ?, ?, ?, ?, ?)"
	return s.db.WithContext(ctx).Exec(sql, "page", id, locale, title, body, slug).Error
}

// RemoveFromIndex deletes entries for a given content type and ID.
func (s *SearchService) RemoveFromIndex(ctx context.Context, contentType string, id uint) error {
	sql := "DELETE FROM search_index_fts WHERE content_type = ? AND content_id = ?"
	return s.db.WithContext(ctx).Exec(sql, contentType, id).Error
}

// RebuildIndex drops and recreates the entire search index from published articles.
func (s *SearchService) RebuildIndex(ctx context.Context) error {
	if err := s.db.WithContext(ctx).Exec("DELETE FROM search_index_fts").Error; err != nil {
		return fmt.Errorf("clear index: %w", err)
	}

	var articles []struct {
		ID      uint
		ZhTitle string
		EnTitle string
		ZhBody  string
		EnBody  string
		Slug    string
	}
	if err := s.db.WithContext(ctx).Table("articles").Where("status = ?", "published").Find(&articles).Error; err != nil {
		return fmt.Errorf("fetch articles: %w", err)
	}

	insertSQL := "INSERT INTO search_index_fts(content_type, content_id, locale, title, body, slug) VALUES(?, ?, ?, ?, ?, ?)"
	for _, a := range articles {
		if a.ZhTitle != "" {
			if err := s.db.WithContext(ctx).Exec(insertSQL, "article", a.ID, "zh", a.ZhTitle, a.ZhBody, a.Slug).Error; err != nil {
				return fmt.Errorf("index article %d zh: %w", a.ID, err)
			}
		}
		if a.EnTitle != "" {
			if err := s.db.WithContext(ctx).Exec(insertSQL, "article", a.ID, "en", a.EnTitle, a.EnBody, a.Slug).Error; err != nil {
				return fmt.Errorf("index article %d en: %w", a.ID, err)
			}
		}
	}

	return nil
}

// Compile-time check that SearchService satisfies the SearchProvider interface.
var _ provider.SearchProvider = (*SearchService)(nil)
