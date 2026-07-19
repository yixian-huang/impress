package seo_test

import (
	"strings"
	"testing"

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
