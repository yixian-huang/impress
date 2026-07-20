package media

import (
	"net/http"
	"strings"
)

// SniffUploadMIME detects MIME type from file content bytes.
// Client Content-Type headers are never trusted for authorization of the type.
// Fonts are recognized by magic bytes when http.DetectContentType falls back
// to application/octet-stream.
func SniffUploadMIME(head []byte) string {
	if len(head) == 0 {
		return "application/octet-stream"
	}
	if mt := sniffFontMIME(head); mt != "" {
		return mt
	}
	return http.DetectContentType(head)
}

func sniffFontMIME(head []byte) string {
	if len(head) < 4 {
		return ""
	}
	// wOFF / wOF2 signatures
	sig := string(head[:4])
	switch sig {
	case "wOFF":
		return "font/woff"
	case "wOF2":
		return "font/woff2"
	}
	// TrueType often starts with 0x00010000 or "true"
	if len(head) >= 4 && head[0] == 0x00 && head[1] == 0x01 && head[2] == 0x00 && head[3] == 0x00 {
		return "font/ttf"
	}
	if sig == "true" || sig == "OTTO" {
		return "font/otf"
	}
	return ""
}

// IsAllowedUploadMIME reports whether a sniffed MIME may be stored.
func IsAllowedUploadMIME(mimeType string) bool {
	mimeType = strings.TrimSpace(strings.ToLower(mimeType))
	if mimeType == "" {
		return false
	}
	return strings.HasPrefix(mimeType, "image/") ||
		strings.HasPrefix(mimeType, "video/") ||
		strings.HasPrefix(mimeType, "audio/") ||
		strings.HasPrefix(mimeType, "font/")
}
