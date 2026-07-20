package contentexcerpt

import (
	"strings"
	"testing"

	"github.com/yixian-huang/inkless/backend/internal/model"
)

func TestPlainExcerpt_PrefersMeta(t *testing.T) {
	got := plainExcerpt("<p>long body text here</p>", "meta desc", 160)
	if got != "meta desc" {
		t.Fatalf("expected meta, got %q", got)
	}
}

func TestPlainExcerpt_StripsHTML(t *testing.T) {
	got := plainExcerpt("<p>Hello <strong>world</strong> &amp; friends</p>", "", 160)
	if got != "Hello world & friends" {
		t.Fatalf("got %q", got)
	}
}

func TestPlainExcerpt_Truncates(t *testing.T) {
	body := strings.Repeat("字", 200)
	got := plainExcerpt("<p>"+body+"</p>", "", 50)
	if !strings.HasSuffix(got, "...") {
		t.Fatalf("expected ellipsis, got %q", got)
	}
	// 50 runes + "..."
	runes := []rune(strings.TrimSuffix(got, "..."))
	if len(runes) > 50 {
		t.Fatalf("too long: %d runes", len(runes))
	}
}

func TestApplyListExcerpts(t *testing.T) {
	items := []*model.Article{
		{
			ZhBody:            "<p>中文正文内容用于列表摘要展示。</p>",
			ZhMetaDescription: "",
			EnBody:            "<p>English body for excerpt.</p>",
		},
	}
	ApplyListExcerpts(items)
	if items[0].ZhBody == "" || strings.Contains(items[0].ZhBody, "<p>") {
		t.Fatalf("zh body not excerpted: %q", items[0].ZhBody)
	}
	if items[0].EnBody != "English body for excerpt." {
		t.Fatalf("en body: %q", items[0].EnBody)
	}
}

func TestFillAndApplyStoredExcerpts(t *testing.T) {
	a := &model.Article{
		ZhBody:            "<p>发布时写入摘要字段的中文正文。</p>",
		EnBody:            "<p>English body stored as excerpt at publish.</p>",
		ZhMetaDescription: "",
	}
	FillStoredExcerpts(a)
	if a.ZhExcerpt == "" || strings.Contains(a.ZhExcerpt, "<p>") {
		t.Fatalf("zh excerpt: %q", a.ZhExcerpt)
	}
	// Simulate list row without bodies (only excerpts selected).
	listRow := &model.Article{ZhExcerpt: a.ZhExcerpt, EnExcerpt: a.EnExcerpt}
	ApplyListExcerpts([]*model.Article{listRow})
	if listRow.ZhBody != a.ZhExcerpt {
		t.Fatalf("list body should use stored excerpt, got %q", listRow.ZhBody)
	}
}
