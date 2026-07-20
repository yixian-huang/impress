package media

import (
	"testing"
)

func TestSniffUploadMIME_PNGIgnoresSpoofedClientType(t *testing.T) {
	// Minimal PNG header magic
	png := []byte{0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a}
	got := SniffUploadMIME(png)
	if got != "image/png" {
		t.Fatalf("expected image/png, got %q", got)
	}
	if !IsAllowedUploadMIME(got) {
		t.Fatalf("png should be allowed")
	}
}

func TestSniffUploadMIME_HTMLNotAllowedAsImage(t *testing.T) {
	// Spoof path: client may send Content-Type: image/png but body is HTML.
	html := []byte("<!DOCTYPE html><html><script>alert(1)</script></html>")
	got := SniffUploadMIME(html)
	if stringsHasPrefixImage(got) {
		t.Fatalf("HTML must not sniff as image, got %q", got)
	}
	if IsAllowedUploadMIME(got) {
		t.Fatalf("HTML sniff %q must not be allowed upload MIME", got)
	}
}

func TestSniffUploadMIME_WOFF2(t *testing.T) {
	head := []byte("wOF2xxxxxxxx")
	got := SniffUploadMIME(head)
	if got != "font/woff2" {
		t.Fatalf("expected font/woff2, got %q", got)
	}
	if !IsAllowedUploadMIME(got) {
		t.Fatal("woff2 should be allowed")
	}
}

func TestIsAllowedUploadMIME(t *testing.T) {
	cases := map[string]bool{
		"image/jpeg":              true,
		"video/mp4":               true,
		"audio/mpeg":              true,
		"font/woff2":              true,
		"text/html":               false,
		"application/octet-stream": false,
		"":                        false,
	}
	for mime, want := range cases {
		if got := IsAllowedUploadMIME(mime); got != want {
			t.Errorf("IsAllowedUploadMIME(%q)=%v want %v", mime, got, want)
		}
	}
}

func stringsHasPrefixImage(s string) bool {
	return len(s) >= 6 && s[:6] == "image/"
}
