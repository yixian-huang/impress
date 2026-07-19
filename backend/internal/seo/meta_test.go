package seo_test

import (
	"testing"

	"github.com/yixian-huang/inkless/backend/internal/seo"
)

func TestDefaultPageMeta(t *testing.T) {
	meta := seo.DefaultPageMeta()
	if meta.Title == "" {
		t.Error("default title should not be empty")
	}
	if meta.OgType != "website" {
		t.Errorf("expected og:type 'website', got %q", meta.OgType)
	}
	if meta.Locale != "zh" {
		t.Errorf("expected default locale 'zh', got %q", meta.Locale)
	}
}

func TestPageMetaWithOverrides(t *testing.T) {
	meta := seo.DefaultPageMeta()
	meta.Title = "Custom Title"
	meta.Description = "Custom Desc"
	meta.OgImage = "https://example.com/img.png"
	meta.CanonicalURL = "https://example.com/about"

	if meta.Title != "Custom Title" {
		t.Errorf("expected custom title, got %q", meta.Title)
	}
}

func TestApplyGlobal_OverlaysTitleAndOG(t *testing.T) {
	pm := seo.DefaultPageMeta()
	pm.ApplyGlobal(map[string]any{
		"identity": map[string]any{
			"name":       map[string]any{"zh": "我的博客"},
			"localeMode": "mono-zh",
		},
		"seo":   map[string]any{},
		"brand": map[string]any{"ogImage": "https://x.test/og.png"},
	}, "zh")
	if pm.Title != "我的博客" {
		t.Errorf("Title: got %q want 我的博客", pm.Title)
	}
	if pm.OgImage != "https://x.test/og.png" {
		t.Errorf("OgImage: got %q", pm.OgImage)
	}
	if pm.Locale != "zh" {
		t.Errorf("Locale: got %q", pm.Locale)
	}
}
