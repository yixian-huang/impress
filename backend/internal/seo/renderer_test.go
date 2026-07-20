package seo_test

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/yixian-huang/inkless/backend/internal/seo"
)

const testTemplate = `<!DOCTYPE html>
<html lang="{{.Locale}}">
<head>
  <title>{{.Title}}</title>
  <meta name="description" content="{{.Description}}">
  <meta property="og:title" content="{{.OgTitle}}">
  <meta property="og:description" content="{{.OgDescription}}">
  <meta property="og:image" content="{{.OgImage}}">
  <meta property="og:url" content="{{.OgURL}}">
  <meta property="og:type" content="{{.OgType}}">
  <link rel="canonical" href="{{.CanonicalURL}}">
</head>
<body><div id="root"></div></body>
</html>`

func TestRendererRender(t *testing.T) {
	r, err := seo.NewRendererFromString(testTemplate)
	if err != nil {
		t.Fatalf("new renderer: %v", err)
	}

	meta := seo.DefaultPageMeta()
	meta.Title = "About Us"
	meta.CanonicalURL = "https://example.com/about"

	result, err := r.Render(meta)
	if err != nil {
		t.Fatalf("render: %v", err)
	}

	if !strings.Contains(result, "<title>About Us</title>") {
		t.Error("expected custom title in output")
	}
	if !strings.Contains(result, `href="https://example.com/about"`) {
		t.Error("expected canonical URL in output")
	}
	if !strings.Contains(result, `lang="zh"`) {
		t.Error("expected locale in html lang attribute")
	}
}

func TestRendererDefaultMeta(t *testing.T) {
	r, err := seo.NewRendererFromString(testTemplate)
	if err != nil {
		t.Fatalf("new renderer: %v", err)
	}

	meta := seo.DefaultPageMeta()
	result, err := r.Render(meta)
	if err != nil {
		t.Fatalf("render: %v", err)
	}

	if !strings.Contains(result, "Site") {
		t.Error("expected default title in output")
	}
}

func TestRendererReloadsWhenIndexHTMLChanges(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/index.html"
	v1 := `<!DOCTYPE html><html><head><title>{{.Title}}</title><script src="/assets/v1.js"></script></head><body></body></html>`
	if err := os.WriteFile(path, []byte(v1), 0o644); err != nil {
		t.Fatal(err)
	}
	// Ensure distinct mtime for next write on coarse filesystems
	past := time.Now().Add(-2 * time.Second)
	_ = os.Chtimes(path, past, past)

	r, err := seo.NewRenderer(path)
	if err != nil {
		t.Fatalf("new renderer: %v", err)
	}
	out1, err := r.Render(seo.DefaultPageMeta())
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out1, "/assets/v1.js") {
		t.Fatalf("expected v1 asset, got %s", out1)
	}

	v2 := `<!DOCTYPE html><html><head><title>{{.Title}}</title><script src="/assets/v2.js"></script></head><body></body></html>`
	if err := os.WriteFile(path, []byte(v2), 0o644); err != nil {
		t.Fatal(err)
	}
	_ = os.Chtimes(path, time.Now(), time.Now())

	out2, err := r.Render(seo.DefaultPageMeta())
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out2, "/assets/v2.js") {
		t.Fatalf("expected reloaded v2 asset, got %s", out2)
	}
}
