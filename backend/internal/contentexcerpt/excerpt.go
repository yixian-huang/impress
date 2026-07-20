package contentexcerpt

import (
	"html"
	"strings"
	"unicode/utf8"

	"github.com/yixian-huang/inkless/backend/internal/model"
)

const publicListExcerptMaxRunes = 160

// stripHTML removes tags for plain-text excerpts (list / feed style).
func stripHTML(s string) string {
	out := s
	for {
		i := strings.Index(out, "<")
		if i < 0 {
			break
		}
		j := strings.Index(out[i:], ">")
		if j < 0 {
			break
		}
		out = out[:i] + " " + out[i+j+1:]
	}
	// Collapse whitespace
	fields := strings.Fields(html.UnescapeString(out))
	return strings.Join(fields, " ")
}

// plainExcerpt prefers SEO meta description, else plain text from HTML body.
func plainExcerpt(bodyHTML, metaDescription string, maxRunes int) string {
	if maxRunes <= 0 {
		maxRunes = publicListExcerptMaxRunes
	}
	meta := strings.TrimSpace(metaDescription)
	if meta != "" {
		return truncateRunes(meta, maxRunes)
	}
	text := stripHTML(bodyHTML)
	if text == "" {
		return ""
	}
	return truncateRunes(text, maxRunes)
}

func truncateRunes(s string, maxRunes int) string {
	if maxRunes <= 0 || utf8.RuneCountInString(s) <= maxRunes {
		return s
	}
	runes := []rune(s)
	cut := maxRunes
	// Prefer cutting on a space near the end when possible.
	if cut > 20 {
		for i := cut; i > cut-20 && i > 0; i-- {
			if runes[i-1] == ' ' || runes[i-1] == '，' || runes[i-1] == '。' {
				cut = i
				break
			}
		}
	}
	return strings.TrimSpace(string(runes[:cut])) + "..."
}

// FillStoredExcerpts writes plain-text previews onto the article for list/feed.
// Call on create/update when status is published (and on scheduled publish).
func FillStoredExcerpts(a *model.Article) {
	if a == nil {
		return
	}
	a.ZhExcerpt = plainExcerpt(a.ZhBody, a.ZhMetaDescription, publicListExcerptMaxRunes)
	a.EnExcerpt = plainExcerpt(a.EnBody, a.EnMetaDescription, publicListExcerptMaxRunes)
}

// ApplyListExcerpts replaces full HTML bodies with short plain-text excerpts
// so public list payloads stay small while home/archive can show previews.
// Prefers stored ZhExcerpt/EnExcerpt when present (no body required on the row).
func ApplyListExcerpts(items []*model.Article) {
	for _, a := range items {
		if a == nil {
			continue
		}
		if a.ZhExcerpt != "" {
			a.ZhBody = a.ZhExcerpt
		} else {
			a.ZhBody = plainExcerpt(a.ZhBody, a.ZhMetaDescription, publicListExcerptMaxRunes)
		}
		if a.EnExcerpt != "" {
			a.EnBody = a.EnExcerpt
		} else {
			a.EnBody = plainExcerpt(a.EnBody, a.EnMetaDescription, publicListExcerptMaxRunes)
		}
	}
}

// Description prefers stored excerpt, then meta, then a truncated body.
func Description(a *model.Article, maxRunes int) string {
	if a == nil {
		return ""
	}
	if maxRunes <= 0 {
		maxRunes = publicListExcerptMaxRunes
	}
	if a.ZhExcerpt != "" {
		return truncateRunes(a.ZhExcerpt, maxRunes)
	}
	if a.EnExcerpt != "" {
		return truncateRunes(a.EnExcerpt, maxRunes)
	}
	if a.ZhMetaDescription != "" {
		return truncateRunes(a.ZhMetaDescription, maxRunes)
	}
	if a.EnMetaDescription != "" {
		return truncateRunes(a.EnMetaDescription, maxRunes)
	}
	body := a.ZhBody
	if body == "" {
		body = a.EnBody
	}
	return plainExcerpt(body, "", maxRunes)
}
